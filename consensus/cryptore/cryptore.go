// Copyright 2020 by the Authors
// This file is part of the go-core library.
//
// The go-core library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-core library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-core library. If not, see <http://www.gnu.org/licenses/>.

// Package cryptore implements the cryptore proof-of-work consensus engine.
package cryptore

import (
	"github.com/core-coin/go-core/consensus"
	"github.com/core-coin/go-core/log"
	"github.com/core-coin/go-core/metrics"
	"github.com/core-coin/go-core/rpc"
	"github.com/core-coin/go-randomx"
	"math/big"
	"math/rand"
	"sync"
	"time"
)

var (
	// two256 is a big integer representing 2^256
	two256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))

	// sharedCryptore is a full instance that can be shared between multiple users.
	sharedCryptore = New(Config{ModeNormal, nil}, nil, false)
)
var (
	vm    *randomx.RandxVm
	mutex *sync.Mutex
)

func init() {
	vm, mutex = randomx.NewRandomXVMWithKeyAndMutex()
}

// Mode defines the type and amount of PoW verification an cryptore engine makes.
type Mode uint

const (
	ModeNormal Mode = iota
	ModeShared
	ModeTest
	ModeFake
	ModeFullFake
)

// Config are the configuration parameters of the cryptore.
type Config struct {
	PowMode Mode

	Log log.Logger `toml:"-"`
}

// Cryptore is a consensus engine based on proof-of-work implementing the cryptore
// algorithm.
type Cryptore struct {
	config Config

	randomXVM *randomx.RandxVm
	vmMutex   *sync.Mutex

	// Mining related fields
	rand     *rand.Rand    // Properly seeded random source for nonces
	threads  int           // Number of threads to mine on if mining
	update   chan struct{} // Notification channel to update mining parameters
	hashrate metrics.Meter // Meter tracking the average hashrate
	remote   *remoteSealer

	// The fields below are hooks for testing
	shared    *Cryptore     // Shared PoW verifier to avoid cache regeneration
	fakeFail  uint64        // Block number which fails PoW check even in fake mode
	fakeDelay time.Duration // Time delay to sleep for before returning from verify

	lock      sync.Mutex // Ensures thread safety for the in-memory caches and mining fields
	closeOnce sync.Once  // Ensures exit channel will not be closed twice.
}

// New creates a full sized cryptore PoW scheme and starts a background thread for
// remote mining, also optionally notifying a batch of remote services of new work
// packages.
func New(config Config, notify []string, noverify bool) *Cryptore {
	if config.Log == nil {
		config.Log = log.Root()
	}
	vm, mutex := randomx.NewRandomXVMWithKeyAndMutex()
	cryptore := &Cryptore{
		config:    config,
		update:    make(chan struct{}),
		hashrate:  metrics.NewMeterForced(),
		randomXVM: vm,
		vmMutex:   mutex,
	}
	cryptore.remote = startRemoteSealer(cryptore, notify, noverify)
	return cryptore
}

// NewTester creates a small sized cryptore PoW scheme useful only for testing
// purposes.
func NewTester(notify []string, noverify bool) *Cryptore {
	vm, mutex := randomx.NewRandomXVMWithKeyAndMutex()
	cryptore := &Cryptore{
		config:    Config{PowMode: ModeTest, Log: log.Root()},
		update:    make(chan struct{}),
		hashrate:  metrics.NewMeterForced(),
		randomXVM: vm,
		vmMutex:   mutex,
	}
	cryptore.remote = startRemoteSealer(cryptore, notify, noverify)
	return cryptore
}

// NewFaker creates a cryptore consensus engine with a fake PoW scheme that accepts
// all blocks' seal as valid, though they still have to conform to the Core
// consensus rules.
func NewFaker() *Cryptore {
	return &Cryptore{
		config: Config{
			PowMode: ModeFake,
			Log:     log.Root(),
		},
	}
}

// NewFakeFailer creates a cryptore consensus engine with a fake PoW scheme that
// accepts all blocks as valid apart from the single one specified, though they
// still have to conform to the Core consensus rules.
func NewFakeFailer(fail uint64) *Cryptore {
	return &Cryptore{
		config: Config{
			PowMode: ModeFake,
			Log:     log.Root(),
		},
		fakeFail: fail,
	}
}

// NewFakeDelayer creates a cryptore consensus engine with a fake PoW scheme that
// accepts all blocks as valid, but delays verifications by some time, though
// they still have to conform to the Core consensus rules.
func NewFakeDelayer(delay time.Duration) *Cryptore {
	return &Cryptore{
		config: Config{
			PowMode: ModeFake,
			Log:     log.Root(),
		},
		fakeDelay: delay,
	}
}

// NewFullFaker creates an cryptore consensus engine with a full fake scheme that
// accepts all blocks as valid, without checking any consensus rules whatsoever.
func NewFullFaker() *Cryptore {
	return &Cryptore{
		config: Config{
			PowMode: ModeFullFake,
			Log:     log.Root(),
		},
	}
}

// NewShared creates a full sized cryptore PoW shared between all requesters running
// in the same process.
func NewShared() *Cryptore {
	return &Cryptore{shared: sharedCryptore}
}

// Close closes the exit channel to notify all backend threads exiting.
func (cryptore *Cryptore) Close() error {
	var err error
	cryptore.closeOnce.Do(func() {
		// Short circuit if the exit channel is not allocated.
		if cryptore.remote == nil {
			return
		}
		close(cryptore.remote.requestExit)
		<-cryptore.remote.exitCh
	})
	return err
}

// Threads returns the number of mining threads currently enabled. This doesn't
// necessarily mean that mining is running!
func (cryptore *Cryptore) Threads() int {
	cryptore.lock.Lock()
	defer cryptore.lock.Unlock()

	return cryptore.threads
}

// SetThreads updates the number of mining threads currently enabled. Calling
// this method does not start mining, only sets the thread count. If zero is
// specified, the miner will use all cores of the machine. Setting a thread
// count below zero is allowed and will cause the miner to idle, without any
// work being done.
func (cryptore *Cryptore) SetThreads(threads int) {
	cryptore.lock.Lock()
	defer cryptore.lock.Unlock()

	// If we're running a shared PoW, set the thread count on that instead
	if cryptore.shared != nil {
		cryptore.shared.SetThreads(threads)
		return
	}
	// Update the threads and ping any running seal to pull in any changes
	cryptore.threads = threads
	select {
	case cryptore.update <- struct{}{}:
	default:
	}
}

// Hashrate implements PoW, returning the measured rate of the search invocations
// per second over the last minute.
// Note the returned hashrate includes local hashrate, but also includes the total
// hashrate of all remote miner.
func (cryptore *Cryptore) Hashrate() float64 {
	// Short circuit if we are run the cryptore in normal/test mode.
	if cryptore.config.PowMode != ModeNormal && cryptore.config.PowMode != ModeTest {
		return cryptore.hashrate.Rate1()
	}
	var res = make(chan uint64, 1)

	select {
	case cryptore.remote.fetchRateCh <- res:
	case <-cryptore.remote.exitCh:
		// Return local hashrate only if cryptore is stopped.
		return cryptore.hashrate.Rate1()
	}

	// Gather total submitted hash rate of remote sealers.
	return cryptore.hashrate.Rate1() + float64(<-res)
}

// APIs implements consensus.Engine, returning the user facing RPC APIs.
func (cryptore *Cryptore) APIs(chain consensus.ChainReader) []rpc.API {
	// In order to ensure backward compatibility, we exposes cryptore RPC APIs
	// to both eth and cryptore namespaces.
	return []rpc.API{
		{
			Namespace: "xcb",
			Version:   "1.0",
			Service:   &API{cryptore},
			Public:    true,
		},
		{
			Namespace: "cryptore",
			Version:   "1.0",
			Service:   &API{cryptore},
			Public:    true,
		},
	}
}

// SeedHash is the seed to use for generating a verification cache and the mining
// dataset.
func SeedHash(block uint64) []byte {
	return seedHash(block)
}
