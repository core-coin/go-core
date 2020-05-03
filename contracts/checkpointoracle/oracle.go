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

// Package checkpointoracle is a an on-chain light client checkpoint oracle.
package checkpointoracle

//go:generate abigen --sol contract/oracle.sol --pkg contract --out contract/oracle.go

import (
	"errors"
	"math/big"

	"github.com/core-coin/go-core/accounts/abi/bind"
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/contracts/checkpointoracle/contract"
	"github.com/core-coin/go-core/core/types"
)

// CheckpointOracle is a Go wrapper around an on-chain checkpoint oracle contract.
type CheckpointOracle struct {
	address  common.Address
	contract *contract.CheckpointOracle
}

// NewCheckpointOracle binds checkpoint contract and returns a registrar instance.
func NewCheckpointOracle(contractAddr common.Address, backend bind.ContractBackend) (*CheckpointOracle, error) {
	c, err := contract.NewCheckpointOracle(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	return &CheckpointOracle{address: contractAddr, contract: c}, nil
}

// ContractAddr returns the address of contract.
func (oracle *CheckpointOracle) ContractAddr() common.Address {
	return oracle.address
}

// Contract returns the underlying contract instance.
func (oracle *CheckpointOracle) Contract() *contract.CheckpointOracle {
	return oracle.contract
}

// LookupCheckpointEvents searches checkpoint event for specific section in the
// given log batches.
func (oracle *CheckpointOracle) LookupCheckpointEvents(blockLogs [][]*types.Log, section uint64, hash common.Hash) []*contract.CheckpointOracleNewCheckpointVote {
	var votes []*contract.CheckpointOracleNewCheckpointVote

	for _, logs := range blockLogs {
		for _, log := range logs {
			event, err := oracle.contract.ParseNewCheckpointVote(*log)
			if err != nil {
				continue
			}
			if event.Index == section && common.Hash(event.CheckpointHash) == hash {
				votes = append(votes, event)
			}
		}
	}
	return votes
}

// RegisterCheckpoint registers the checkpoint with a batch of associated signatures
// that are collected off-chain and sorted by lexicographical order.
//
// Notably all signatures given should be transformed to "ethereum style" which transforms
func (oracle *CheckpointOracle) RegisterCheckpoint(opts *bind.TransactOpts, index uint64, hash []byte, rnum *big.Int, rhash [32]byte, sigs [][]byte) (*types.Transaction, error) {
	for i := 0; i < len(sigs); i++ {
		if len(sigs[i]) != crypto.SignatureLength {
			return nil, errors.New("invalid signature")
		}
	}
	//return oracle.contract.SetCheckpoint(opts, rnum, rhash, common.BytesToHash(hash), index, r, s, v)
	return nil, errors.New("reimplement CheckpointOralce with ed448")
}
