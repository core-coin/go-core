// Copyright 2019 The go-core Authors
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

package core

import (
	"sync/atomic"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/consensus"
	"github.com/core-coin/go-core/core/state"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/core/vm"
	"github.com/core-coin/go-core/params"
)

// statePrefetcher is a basic Prefetcher, which blindly executes a block on top
// of an arbitrary state with the goal of prefetching potentially useful state
// data from disk before the main block processor start executing.
type statePrefetcher struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
}

// newStatePrefetcher initialises a new statePrefetcher.
func newStatePrefetcher(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *statePrefetcher {
	return &statePrefetcher{
		config: config,
		bc:     bc,
		engine: engine,
	}
}

// Prefetch processes the state changes according to the Core rules by running
// the transaction messages using the statedb, but any changes are discarded. The
// only goal is to pre-cache transaction signatures and state trie nodes.
func (p *statePrefetcher) Prefetch(block *types.Block, statedb *state.StateDB, cfg vm.Config, interrupt *uint32) {
	var (
		header  = block.Header()
		energypool = new(EnergyPool).AddEnergy(block.EnergyLimit())
	)
	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions() {
		// If block precaching was interrupted, abort
		if interrupt != nil && atomic.LoadUint32(interrupt) == 1 {
			return
		}
		// Block precaching permitted to continue, execute the transaction
		statedb.Prepare(tx.Hash(), block.Hash(), i)
		if err := precacheTransaction(p.config, p.bc, nil, energypool, statedb, header, tx, cfg); err != nil {
			return // Ugh, something went horribly wrong, bail out
		}
	}
}

// precacheTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. The goal is not to execute
// the transaction successfully, rather to warm up touched data slots.
func precacheTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, energypool *EnergyPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, cfg vm.Config) error {
	// Convert the transaction into an executable message and pre-cache its sender
	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return err
	}
	// Create the CVM and execute the transaction
	context := NewCVMContext(msg, header, bc, author)
	vm := vm.NewCVM(context, statedb, config, cfg)

	_, _, _, err = ApplyMessage(vm, msg, energypool)
	return err
}
