// Copyright 2023 by the Authors
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

package cryptore

import (
	"errors"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/core/types"
)

var errCryptoreStopped = errors.New("cryptore stopped")

// API exposes cryptore related methods for the RPC interface.
type API struct {
	cryptore *Cryptore
}

// GetWork returns a work package for external miner.
//
// The work package consists of 3 strings:
//
//	result[0] - 32 bytes hex encoded current block header pow-hash
//	result[2] - 32 bytes hex encoded boundary condition ("target"), 2^256/difficulty
//	result[3] - hex encoded block number
func (api *API) GetWork() ([4]string, error) {
	if api.cryptore.remote == nil {
		return [4]string{}, errors.New("not supported")
	}

	var (
		workCh = make(chan [4]string, 1)
		errc   = make(chan error, 1)
	)
	select {
	case api.cryptore.remote.fetchWorkCh <- &sealWork{errc: errc, res: workCh}:
	case <-api.cryptore.remote.exitCh:
		return [4]string{}, errCryptoreStopped
	}
	select {
	case work := <-workCh:
		return work, nil
	case err := <-errc:
		return [4]string{}, err
	}
}

// SubmitWork can be used by external miner to submit their POW solution.
// It returns an indication if the work was accepted.
// Note either an invalid solution, a stale work a non-existent work will return false.
func (api *API) SubmitWork(nonce types.BlockNonce, hash common.Hash) bool {
	if api.cryptore.remote == nil {
		return false
	}

	var errc = make(chan error, 1)
	select {
	case api.cryptore.remote.submitWorkCh <- &mineResult{
		nonce: nonce,
		hash:  hash,
		errc:  errc,
	}:
	case <-api.cryptore.remote.exitCh:
		return false
	}
	err := <-errc
	return err == nil
}

// SubmitHashrate can be used for remote miners to submit their hash rate.
// This enables the node to report the combined hash rate of all miners
// which submit work through this node.
//
// It accepts the miner hash rate and an identifier which must be unique
// between nodes.
func (api *API) SubmitHashRate(rate hexutil.Uint64, id common.Hash) bool {
	if api.cryptore.remote == nil {
		return false
	}

	var done = make(chan struct{}, 1)
	select {
	case api.cryptore.remote.submitRateCh <- &hashrate{done: done, rate: uint64(rate), id: id}:
	case <-api.cryptore.remote.exitCh:
		return false
	}

	// Block until hash rate submitted successfully.
	<-done
	return true
}

// GetHashrate returns the current hashrate for local CPU miner and remote miner.
func (api *API) GetHashrate() uint64 {
	return uint64(api.cryptore.Hashrate())
}
