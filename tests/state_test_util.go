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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/hexutil"
	"github.com/core-coin/go-core/common/math"
	"github.com/core-coin/go-core/core"
	"github.com/core-coin/go-core/core/rawdb"
	"github.com/core-coin/go-core/core/state"
	"github.com/core-coin/go-core/core/state/snapshot"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/core/vm"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/params"
	"github.com/core-coin/go-core/rlp"
	"github.com/core-coin/go-core/xcbdb"
	"golang.org/x/crypto/sha3"
)

// StateTest checks transaction processing without block context.
// See https://github.com/core/CIPs/issues/176 for the test format specification.
type StateTest struct {
	json stJSON
}

// StateSubtest selects a specific configuration of a General State Test.
type StateSubtest struct {
	Fork  string
	Index int
}

func (t *StateTest) UnmarshalJSON(in []byte) error {
	return json.Unmarshal(in, &t.json)
}

type stJSON struct {
	Env  stEnv                    `json:"env"`
	Pre  core.GenesisAlloc        `json:"pre"`
	Tx   stTransaction            `json:"transaction"`
	Out  hexutil.Bytes            `json:"out"`
	Post map[string][]stPostState `json:"post"`
}

type stPostState struct {
	Root    common.UnprefixedHash `json:"hash"`
	Logs    common.UnprefixedHash `json:"logs"`
	Indexes struct {
		Data   int `json:"data"`
		Energy int `json:"energy"`
		Value  int `json:"value"`
	}
}

//go:generate gencodec -type stEnv -field-override stEnvMarshaling -out gen_stenv.go

type stEnv struct {
	Coinbase    common.Address `json:"currentCoinbase"   gencodec:"required"`
	Difficulty  *big.Int       `json:"currentDifficulty" gencodec:"required"`
	EnergyLimit uint64         `json:"currentEnergyLimit"   gencodec:"required"`
	Number      uint64         `json:"currentNumber"     gencodec:"required"`
	Timestamp   uint64         `json:"currentTimestamp"  gencodec:"required"`
}

type stEnvMarshaling struct {
	Coinbase    common.UnprefixedAddress
	Difficulty  *math.HexOrDecimal256
	EnergyLimit math.HexOrDecimal64
	Number      math.HexOrDecimal64
	Timestamp   math.HexOrDecimal64
}

//go:generate gencodec -type stTransaction -field-override stTransactionMarshaling -out gen_sttransaction.go

type stTransaction struct {
	EnergyPrice *big.Int `json:"energyPrice"`
	Nonce       uint64   `json:"nonce"`
	To          string   `json:"to"`
	Data        []string `json:"data"`
	EnergyLimit []uint64 `json:"energyLimit"`
	Value       []string `json:"value"`
	PrivateKey  []byte   `json:"secretKey"`
}

type stTransactionMarshaling struct {
	EnergyPrice *math.HexOrDecimal256
	Nonce       math.HexOrDecimal64
	EnergyLimit []math.HexOrDecimal64
	PrivateKey  hexutil.Bytes
}

// getVMConfig takes a fork definition and returns a chain config.
// The fork definition can be
// - a plain forkname, e.g. `Nucleus`,
// - a fork basename, and a list of CIPs to enable; e.g. `Nucleus+1884+1283`.
func getVMConfig(forkString string) (baseConfig *params.ChainConfig, cips []int, err error) {
	var (
		splitForks            = strings.Split(forkString, "+")
		ok                    bool
		baseName, cipsStrings = splitForks[0], splitForks[1:]
	)
	if baseConfig, ok = Forks[baseName]; !ok {
		return nil, nil, UnsupportedForkError{baseName}
	}
	for _, cip := range cipsStrings {
		if cipNum, err := strconv.Atoi(cip); err != nil {
			return nil, nil, fmt.Errorf("syntax error, invalid cip number %v", cipNum)
		} else {
			cips = append(cips, cipNum)
		}
	}
	return baseConfig, cips, nil
}

// Subtests returns all valid subtests of the test.
func (t *StateTest) Subtests() []StateSubtest {
	var sub []StateSubtest
	for fork, pss := range t.json.Post {
		for i := range pss {
			sub = append(sub, StateSubtest{fork, i})
		}
	}
	return sub
}

// Run executes a specific subtest and verifies the post-state and logs
func (t *StateTest) Run(subtest StateSubtest, vmconfig vm.Config, snapshotter bool) (*state.StateDB, error) {
	statedb, root, err := t.RunNoVerify(subtest, vmconfig, snapshotter)
	if err != nil {
		return statedb, err
	}
	post := t.json.Post[subtest.Fork][subtest.Index]
	// N.B: We need to do this in a two-step process, because the first Commit takes care
	// of suicides, and we need to touch the coinbase _after_ it has potentially suicided.
	if root != common.Hash(post.Root) {
		return statedb, fmt.Errorf("post state root mismatch: got %x, want %x", root, post.Root)
	}
	if logs := rlpHash(statedb.Logs()); logs != common.Hash(post.Logs) {
		return statedb, fmt.Errorf("post state logs hash mismatch: got %x, want %x", logs, post.Logs)
	}
	return statedb, nil
}

// RunNoVerify runs a specific subtest and returns the statedb and post-state root
func (t *StateTest) RunNoVerify(subtest StateSubtest, vmconfig vm.Config, snapshotter bool) (*state.StateDB, common.Hash, error) {
	config, cips, err := getVMConfig(subtest.Fork)
	if err != nil {
		return nil, common.Hash{}, UnsupportedForkError{subtest.Fork}
	}
	vmconfig.ExtraCips = cips
	block := t.genesis(config).ToBlock(nil)
	statedb := MakePreState(rawdb.NewMemoryDatabase(), t.json.Pre, snapshotter)

	post := t.json.Post[subtest.Fork][subtest.Index]
	msg, err := t.json.Tx.toMessage(post)
	if err != nil {
		return nil, common.Hash{}, err
	}
	context := core.NewCVMContext(msg, block.Header(), nil, &t.json.Env.Coinbase)
	context.GetHash = vmTestBlockHash
	cvm := vm.NewCVM(context, statedb, config, vmconfig)

	energypool := new(core.EnergyPool)
	energypool.AddEnergy(block.EnergyLimit())
	snapshot := statedb.Snapshot()
	if _, _, _, err := core.ApplyMessage(cvm, msg, energypool); err != nil {
		statedb.RevertToSnapshot(snapshot)
	}
	// Commit block
	statedb.Commit(true)
	// Add 0-value mining reward. This only makes a difference in the cases
	// where
	// - the coinbase suicided, or
	// - there are only 'bad' transactions, which aren't executed. In those cases,
	//   the coinbase gets no txfee, so isn't created, and thus needs to be touched
	statedb.AddBalance(block.Coinbase(), new(big.Int))
	// And _now_ get the state root
	root := statedb.IntermediateRoot(true)
	return statedb, root, nil
}

func (t *StateTest) energyLimit(subtest StateSubtest) uint64 {
	return t.json.Tx.EnergyLimit[t.json.Post[subtest.Fork][subtest.Index].Indexes.Energy]
}

func MakePreState(db xcbdb.Database, accounts core.GenesisAlloc, snapshotter bool) *state.StateDB {
	sdb := state.NewDatabase(db)
	statedb, _ := state.New(common.Hash{}, sdb, nil)
	for addr, a := range accounts {
		statedb.SetCode(addr, a.Code)
		statedb.SetNonce(addr, a.Nonce)
		statedb.SetBalance(addr, a.Balance)
		for k, v := range a.Storage {
			statedb.SetState(addr, k, v)
		}
	}
	// Commit and re-open to start with a clean state.
	root, _ := statedb.Commit(false)

	var snaps *snapshot.Tree
	if snapshotter {
		snaps = snapshot.New(db, sdb.TrieDB(), 1, root, false)
	}
	statedb, _ = state.New(root, sdb, snaps)
	return statedb
}

func (t *StateTest) genesis(config *params.ChainConfig) *core.Genesis {
	return &core.Genesis{
		Config:      config,
		Coinbase:    t.json.Env.Coinbase,
		Difficulty:  t.json.Env.Difficulty,
		EnergyLimit: t.json.Env.EnergyLimit,
		Number:      t.json.Env.Number,
		Timestamp:   t.json.Env.Timestamp,
		Alloc:       t.json.Pre,
	}
}

func (tx *stTransaction) toMessage(ps stPostState) (core.Message, error) {
	// Derive sender from private key if present.
	var from common.Address
	if len(tx.PrivateKey) > 0 {
		key, err := crypto.ToEDDSA(tx.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("invalid private key: %v", err)
		}
		from = crypto.PubkeyToAddress(key.PublicKey)
	}
	// Parse recipient if present.
	var to *common.Address
	if tx.To != "" {
		to = new(common.Address)
		if err := to.UnmarshalText([]byte(tx.To)); err != nil {
			return nil, fmt.Errorf("invalid to address: %v", err)
		}
	}

	// Get values specific to this post state.
	if ps.Indexes.Data > len(tx.Data) {
		return nil, fmt.Errorf("tx data index %d out of bounds", ps.Indexes.Data)
	}
	if ps.Indexes.Value > len(tx.Value) {
		return nil, fmt.Errorf("tx value index %d out of bounds", ps.Indexes.Value)
	}
	if ps.Indexes.Energy > len(tx.EnergyLimit) {
		return nil, fmt.Errorf("tx energy limit index %d out of bounds", ps.Indexes.Energy)
	}
	dataHex := tx.Data[ps.Indexes.Data]
	valueHex := tx.Value[ps.Indexes.Value]
	energyLimit := tx.EnergyLimit[ps.Indexes.Energy]
	// Value, Data hex encoding is messy: https://github.com/core-coin/tests/issues/203
	value := new(big.Int)
	if valueHex != "0x" {
		v, ok := math.ParseBig256(valueHex)
		if !ok {
			return nil, fmt.Errorf("invalid tx value %q", valueHex)
		}
		value = v
	}
	data, err := hex.DecodeString(strings.TrimPrefix(dataHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid tx data %q", dataHex)
	}

	msg := types.NewMessage(from, to, tx.Nonce, value, energyLimit, tx.EnergyPrice, data, true)
	return msg, nil
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.New256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
