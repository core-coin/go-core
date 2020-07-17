// Copyright 2015 The go-core Authors
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

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/hexutil"
	"github.com/core-coin/go-core/common/math"
	"github.com/core-coin/go-core/core"
	"github.com/core-coin/go-core/core/rawdb"
	"github.com/core-coin/go-core/core/state"
	"github.com/core-coin/go-core/core/vm"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/params"
)

// VMTest checks CVM execution without block or transaction context.
// See https://github.com/core-coin/tests/wiki/VM-Tests for the test format specification.
type VMTest struct {
	json vmJSON
}

func (t *VMTest) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &t.json)
}

type vmJSON struct {
	Env             stEnv                 `json:"env"`
	Exec            vmExec                `json:"exec"`
	Logs            common.UnprefixedHash `json:"logs"`
	EnergyRemaining *math.HexOrDecimal64  `json:"energy"`
	Out             hexutil.Bytes         `json:"out"`
	Pre             core.GenesisAlloc     `json:"pre"`
	Post            core.GenesisAlloc     `json:"post"`
	PostStateRoot   common.Hash           `json:"postStateRoot"`
}

//go:generate gencodec -type vmExec -field-override vmExecMarshaling -out gen_vmexec.go

type vmExec struct {
	Address     common.Address `json:"address"  gencodec:"required"`
	Caller      common.Address `json:"caller"   gencodec:"required"`
	Origin      common.Address `json:"origin"   gencodec:"required"`
	Code        []byte         `json:"code"     gencodec:"required"`
	Data        []byte         `json:"data"     gencodec:"required"`
	Value       *big.Int       `json:"value"    gencodec:"required"`
	EnergyLimit uint64         `json:"energy"      gencodec:"required"`
	EnergyPrice *big.Int       `json:"energyPrice" gencodec:"required"`
}

type vmExecMarshaling struct {
	Address     common.UnprefixedAddress
	Caller      common.UnprefixedAddress
	Origin      common.UnprefixedAddress
	Code        hexutil.Bytes
	Data        hexutil.Bytes
	Value       *math.HexOrDecimal256
	EnergyLimit math.HexOrDecimal64
	EnergyPrice *math.HexOrDecimal256
}

func (t *VMTest) Run(vmconfig vm.Config, snapshotter bool) error {
	statedb := MakePreState(rawdb.NewMemoryDatabase(), t.json.Pre, snapshotter)
	ret, energyRemaining, err := t.exec(statedb, vmconfig)

	if t.json.EnergyRemaining == nil {
		if err == nil {
			return fmt.Errorf("energy unspecified (indicating an error), but VM returned no error")
		}
		if energyRemaining > 0 {
			return fmt.Errorf("energy unspecified (indicating an error), but VM returned energy remaining > 0")
		}
		return nil
	}
	// Test declares energy, expecting outputs to match.
	if !bytes.Equal(ret, t.json.Out) {
		return fmt.Errorf("return data mismatch: got %x, want %x", ret, t.json.Out)
	}
	if energyRemaining != uint64(*t.json.EnergyRemaining) {
		return fmt.Errorf("remaining energy %v, want %v", energyRemaining, *t.json.EnergyRemaining)
	}
	for addr, account := range t.json.Post {
		for k, wantV := range account.Storage {
			if haveV := statedb.GetState(addr, k); haveV != wantV {
				return fmt.Errorf("wrong storage value at %x:\n  got  %x\n  want %x", k, haveV, wantV)
			}
		}
	}
	// if root := statedb.IntermediateRoot(false); root != t.json.PostStateRoot {
	// 	return fmt.Errorf("post state root mismatch, got %x, want %x", root, t.json.PostStateRoot)
	// }
	if logs := rlpHash(statedb.Logs()); logs != common.Hash(t.json.Logs) {
		return fmt.Errorf("post state logs hash mismatch: got %x, want %x", logs, t.json.Logs)
	}
	return nil
}

func (t *VMTest) exec(statedb *state.StateDB, vmconfig vm.Config) ([]byte, uint64, error) {
	cvm := t.newCVM(statedb, vmconfig)
	e := t.json.Exec
	return cvm.Call(vm.AccountRef(e.Caller), e.Address, e.Data, e.EnergyLimit, e.Value)
}

func (t *VMTest) newCVM(statedb *state.StateDB, vmconfig vm.Config) *vm.CVM {
	initialCall := true
	canTransfer := func(db vm.StateDB, address common.Address, amount *big.Int) bool {
		if initialCall {
			initialCall = false
			return true
		}
		return core.CanTransfer(db, address, amount)
	}
	transfer := func(db vm.StateDB, sender, recipient common.Address, amount *big.Int) {}
	context := vm.Context{
		CanTransfer: canTransfer,
		Transfer:    transfer,
		GetHash:     vmTestBlockHash,
		Origin:      t.json.Exec.Origin,
		Coinbase:    t.json.Env.Coinbase,
		BlockNumber: new(big.Int).SetUint64(t.json.Env.Number),
		Time:        new(big.Int).SetUint64(t.json.Env.Timestamp),
		EnergyLimit: t.json.Env.EnergyLimit,
		Difficulty:  t.json.Env.Difficulty,
		EnergyPrice: t.json.Exec.EnergyPrice,
	}
	vmconfig.NoRecursion = true
	return vm.NewCVM(context, statedb, params.MainnetChainConfig, vmconfig)
}

func vmTestBlockHash(n uint64) common.Hash {
	return common.BytesToHash(crypto.Keccak256([]byte(big.NewInt(int64(n)).String())))
}
