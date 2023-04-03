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

// Package miner implements Core block creation and mining.
package miner

import (
	"testing"
	"time"

	"github.com/core-coin/go-core/v2/xcbdb/memorydb"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/consensus/clique"
	"github.com/core-coin/go-core/v2/core"
	"github.com/core-coin/go-core/v2/core/rawdb"
	"github.com/core-coin/go-core/v2/core/state"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/core/vm"
	"github.com/core-coin/go-core/v2/event"
	"github.com/core-coin/go-core/v2/trie"
	"github.com/core-coin/go-core/v2/xcb/downloader"
)

type mockBackend struct {
	bc     *core.BlockChain
	txPool *core.TxPool
}

func NewMockBackend(bc *core.BlockChain, txPool *core.TxPool) *mockBackend {
	return &mockBackend{
		bc:     bc,
		txPool: txPool,
	}
}

func (m *mockBackend) BlockChain() *core.BlockChain {
	return m.bc
}

func (m *mockBackend) TxPool() *core.TxPool {
	return m.txPool
}

type testBlockChain struct {
	statedb       *state.StateDB
	energyLimit   uint64
	chainHeadFeed *event.Feed
}

func (bc *testBlockChain) CurrentBlock() *types.Block {
	return types.NewBlock(&types.Header{
		EnergyLimit: bc.energyLimit,
	}, nil, nil, nil, new(trie.Trie))
}

func (bc *testBlockChain) GetBlock(hash common.Hash, number uint64) *types.Block {
	return bc.CurrentBlock()
}

func (bc *testBlockChain) StateAt(common.Hash) (*state.StateDB, error) {
	return bc.statedb, nil
}

func (bc *testBlockChain) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return bc.chainHeadFeed.Subscribe(ch)
}

func TestMiner(t *testing.T) {
	miner, mux := createMiner(t)
	addr, err := common.HexToAddress("cb375a538daf54f2e568bb4237357b1cee1aa3cb7eba")
	if err != nil {
		t.Error(err)
	}
	miner.Start(addr)
	waitForMiningState(t, miner, true)
	// Start the downloader
	mux.Post(downloader.StartEvent{})
	waitForMiningState(t, miner, false)
	// Stop the downloader and wait for the update loop to run
	mux.Post(downloader.DoneEvent{})
	waitForMiningState(t, miner, true)

	// Subsequent downloader events after a successful DoneEvent should not cause the
	// miner to start or stop. This prevents a security vulnerability
	// that would allow entities to present fake high blocks that would
	// stop mining operations by causing a downloader sync
	// until it was discovered they were invalid, whereon mining would resume.
	mux.Post(downloader.StartEvent{})
	waitForMiningState(t, miner, true)

	mux.Post(downloader.FailedEvent{})
	waitForMiningState(t, miner, true)
}

// TestMinerDownloaderFirstFails tests that mining is only
// permitted to run indefinitely once the downloader sees a DoneEvent (success).
// An initial FailedEvent should allow mining to stop on a subsequent
// downloader StartEvent.
func TestMinerDownloaderFirstFails(t *testing.T) {
	addr, err := common.HexToAddress("cb375a538daf54f2e568bb4237357b1cee1aa3cb7eba")
	if err != nil {
		t.Error(err)
	}
	miner, mux := createMiner(t)
	miner.Start(addr)
	waitForMiningState(t, miner, true)
	// Start the downloader
	mux.Post(downloader.StartEvent{})
	waitForMiningState(t, miner, false)

	// Stop the downloader and wait for the update loop to run
	mux.Post(downloader.FailedEvent{})
	waitForMiningState(t, miner, true)

	// Since the downloader hasn't yet emitted a successful DoneEvent,
	// we expect the miner to stop on next StartEvent.
	mux.Post(downloader.StartEvent{})
	waitForMiningState(t, miner, false)

	// Downloader finally succeeds.
	mux.Post(downloader.DoneEvent{})
	waitForMiningState(t, miner, true)

	// Downloader starts again.
	// Since it has achieved a DoneEvent once, we expect miner
	// state to be unchanged.
	mux.Post(downloader.StartEvent{})
	waitForMiningState(t, miner, true)

	mux.Post(downloader.FailedEvent{})
	waitForMiningState(t, miner, true)
}

func TestMinerStartStopAfterDownloaderEvents(t *testing.T) {
	addr, err := common.HexToAddress("cb375a538daf54f2e568bb4237357b1cee1aa3cb7eba")
	if err != nil {
		t.Error(err)
	}
	addr2, err := common.HexToAddress("cb76a631db606f1452ddc2432931d611f1d5b126f848")
	if err != nil {
		t.Error(err)
	}

	miner, mux := createMiner(t)

	miner.Start(addr)
	waitForMiningState(t, miner, true)
	// Start the downloader
	mux.Post(downloader.StartEvent{})
	waitForMiningState(t, miner, false)

	// Downloader finally succeeds.
	mux.Post(downloader.DoneEvent{})
	waitForMiningState(t, miner, true)

	miner.Stop()
	waitForMiningState(t, miner, false)

	miner.Start(addr2)
	waitForMiningState(t, miner, true)

	miner.Stop()
	waitForMiningState(t, miner, false)
}

func TestStartWhileDownload(t *testing.T) {
	miner, mux := createMiner(t)
	addr, err := common.HexToAddress("cb375a538daf54f2e568bb4237357b1cee1aa3cb7eba")
	if err != nil {
		t.Error(err)
	}
	waitForMiningState(t, miner, false)
	miner.Start(addr)
	waitForMiningState(t, miner, true)
	// Stop the downloader and wait for the update loop to run
	mux.Post(downloader.StartEvent{})
	waitForMiningState(t, miner, false)
	// Starting the miner after the downloader should not work
	miner.Start(addr)
	waitForMiningState(t, miner, false)
}

func TestStartStopMiner(t *testing.T) {
	miner, _ := createMiner(t)
	addr, err := common.HexToAddress("cb375a538daf54f2e568bb4237357b1cee1aa3cb7eba")
	if err != nil {
		t.Error(err)
	}
	waitForMiningState(t, miner, false)
	miner.Start(addr)
	waitForMiningState(t, miner, true)
	miner.Stop()
	waitForMiningState(t, miner, false)
}

func TestCloseMiner(t *testing.T) {
	miner, _ := createMiner(t)
	addr, err := common.HexToAddress("cb375a538daf54f2e568bb4237357b1cee1aa3cb7eba")
	if err != nil {
		t.Error(err)
	}
	waitForMiningState(t, miner, false)
	miner.Start(addr)
	waitForMiningState(t, miner, true)
	// Terminate the miner and wait for the update loop to run
	miner.Close()
	waitForMiningState(t, miner, false)
}

// TestMinerSetCorebase checks that corebase becomes set even if mining isn't
// possible at the moment
func TestMinerSetCorebase(t *testing.T) {
	miner, mux := createMiner(t)
	addr, err := common.HexToAddress("cb375a538daf54f2e568bb4237357b1cee1aa3cb7eba")
	if err != nil {
		t.Error(err)
	}
	addr2, err := common.HexToAddress("cb27de521e43741cf785cbad450d5649187b9612018f")
	if err != nil {
		t.Error(err)
	}
	// Start with a 'bad' mining address
	miner.Start(addr2)
	waitForMiningState(t, miner, true)
	// Start the downloader
	mux.Post(downloader.StartEvent{})
	waitForMiningState(t, miner, false)
	// Now user tries to configure proper mining address
	miner.Start(addr)
	// Stop the downloader and wait for the update loop to run
	mux.Post(downloader.DoneEvent{})

	waitForMiningState(t, miner, true)
	// The miner should now be using the good address
	if got, exp := miner.coinbase, addr; got != exp {
		t.Fatalf("Wrong coinbase, got %x expected %x", got, exp)
	}
}

// waitForMiningState waits until either
// * the desired mining state was reached
// * a timeout was reached which fails the test
func waitForMiningState(t *testing.T, m *Miner, mining bool) {
	t.Helper()

	var state bool
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		if state = m.Mining(); state == mining {
			return
		}
	}
	t.Fatalf("Mining() == %t, want %t", state, mining)
}

func createMiner(t *testing.T) (*Miner, *event.TypeMux) {
	// Create Cryptore config
	addr, err := common.HexToAddress("cb375a538daf54f2e568bb4237357b1cee1aa3cb7eba")
	if err != nil {
		t.Error(err)
	}
	addr2, err := common.HexToAddress("cb270000000000000000000000000000000000000001")
	if err != nil {
		t.Error(err)
	}
	config := Config{
		Corebase: addr2,
	}
	// Create chainConfig
	memdb := memorydb.New()
	chainDB := rawdb.NewDatabase(memdb)
	genesis := core.DeveloperGenesisBlock(15, addr)
	chainConfig, _, err := core.SetupGenesisBlock(chainDB, genesis)
	if err != nil {
		t.Fatalf("can't create new chain config: %v", err)
	}
	// Create consensus engine
	engine := clique.New(chainConfig.Clique, chainDB)
	// Create Core backend
	bc, err := core.NewBlockChain(chainDB, nil, chainConfig, engine, vm.Config{}, nil, nil)
	if err != nil {
		t.Fatalf("can't create new chain %v", err)
	}
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(chainDB), nil)
	blockchain := &testBlockChain{statedb, 10000000, new(event.Feed)}

	pool := core.NewTxPool(testTxPoolConfig, chainConfig, blockchain)
	backend := NewMockBackend(bc, pool)
	// Create event Mux
	mux := new(event.TypeMux)
	// Create Miner
	return New(backend, &config, chainConfig, mux, engine, nil), mux
}
