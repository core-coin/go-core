// Copyright 2020 by the Authors
// This file is part of go-core.
//
// go-core is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-core is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-core. If not, see <http://www.gnu.org/licenses/>.

package t8ntool

import (
	"fmt"
	"math/big"
	"os"

	"golang.org/x/crypto/sha3"

	"github.com/core-coin/go-core/v2/xcbdb"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/math"
	"github.com/core-coin/go-core/v2/core"
	"github.com/core-coin/go-core/v2/core/rawdb"
	"github.com/core-coin/go-core/v2/core/state"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/core/vm"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/log"
	"github.com/core-coin/go-core/v2/params"
	"github.com/core-coin/go-core/v2/rlp"
	"github.com/core-coin/go-core/v2/trie"
)

type Prestate struct {
	Env stEnv             `json:"env"`
	Pre core.GenesisAlloc `json:"pre"`
}

// ExecutionResult contains the execution status after running a state test, any
// error that might have occurred and a dump of the final state if requested.
type ExecutionResult struct {
	StateRoot   common.Hash    `json:"stateRoot"`
	TxRoot      common.Hash    `json:"txRoot"`
	ReceiptRoot common.Hash    `json:"receiptRoot"`
	LogsHash    common.Hash    `json:"logsHash"`
	Bloom       types.Bloom    `json:"logsBloom"        gencodec:"required"`
	Receipts    types.Receipts `json:"receipts"`
	Rejected    []int          `json:"rejected,omitempty"`
}

type ommer struct {
	Delta   uint64         `json:"delta"`
	Address common.Address `json:"address"`
}

//go:generate gencodec -type stEnv -field-override stEnvMarshaling -out gen_stenv.go
type stEnv struct {
	Coinbase    common.Address                      `json:"currentCoinbase"   gencodec:"required"`
	Difficulty  *big.Int                            `json:"currentDifficulty" gencodec:"required"`
	EnergyLimit uint64                              `json:"currentEnergyLimit"   gencodec:"required"`
	Number      uint64                              `json:"currentNumber"     gencodec:"required"`
	Timestamp   uint64                              `json:"currentTimestamp"  gencodec:"required"`
	BlockHashes map[math.HexOrDecimal64]common.Hash `json:"blockHashes,omitempty"`
	Ommers      []ommer                             `json:"ommers,omitempty"`
}

type stEnvMarshaling struct {
	Coinbase    common.UnprefixedAddress
	Difficulty  *math.HexOrDecimal256
	EnergyLimit math.HexOrDecimal64
	Number      math.HexOrDecimal64
	Timestamp   math.HexOrDecimal64
}

// Apply applies a set of transactions to a pre-state
func (pre *Prestate) Apply(vmConfig vm.Config, chainConfig *params.ChainConfig,
	txs types.Transactions, miningReward int64,
	getTracerFn func(txIndex int, txHash common.Hash) (tracer vm.Tracer, err error)) (*state.StateDB, *ExecutionResult, error) {

	// Capture errors for BLOCKHASH operation, if we haven't been supplied the
	// required blockhashes
	var hashError error
	getHash := func(num uint64) common.Hash {
		if pre.Env.BlockHashes == nil {
			hashError = fmt.Errorf("getHash(%d) invoked, no blockhashes provided", num)
			return common.Hash{}
		}
		h, ok := pre.Env.BlockHashes[math.HexOrDecimal64(num)]
		if !ok {
			hashError = fmt.Errorf("getHash(%d) invoked, blockhash for that block not provided", num)
		}
		return h
	}
	var (
		statedb     = MakePreState(rawdb.NewMemoryDatabase(), pre.Pre)
		signer      = types.MakeSigner(chainConfig.NetworkID)
		energypool  = new(core.EnergyPool)
		blockHash   = common.Hash{0x13, 0x37}
		rejectedTxs []int
		includedTxs types.Transactions
		energyUsed  = uint64(0)
		receipts    = make(types.Receipts, 0)
		txIndex     = 0
	)
	energypool.AddEnergy(pre.Env.EnergyLimit)
	vmContext := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    pre.Env.Coinbase,
		BlockNumber: new(big.Int).SetUint64(pre.Env.Number),
		Time:        new(big.Int).SetUint64(pre.Env.Timestamp),
		Difficulty:  pre.Env.Difficulty,
		EnergyLimit: pre.Env.EnergyLimit,
		GetHash:     getHash,
	}

	for i, tx := range txs {
		msg, err := tx.AsMessage(signer)
		if err != nil {
			log.Info("rejected tx", "index", i, "hash", tx.Hash(), "error", err)
			rejectedTxs = append(rejectedTxs, i)
			continue
		}
		tracer, err := getTracerFn(txIndex, tx.Hash())
		if err != nil {
			return nil, nil, err
		}
		vmConfig.Tracer = tracer
		vmConfig.Debug = (tracer != nil)
		statedb.Prepare(tx.Hash(), blockHash, txIndex)
		txContext := core.NewCVMTxContext(msg)

		cvm := vm.NewCVM(vmContext, txContext, statedb, chainConfig, vmConfig)
		snapshot := statedb.Snapshot()
		// (ret []byte, usedEnergy uint64, failed bool, err error)
		msgResult, err := core.ApplyMessage(cvm, msg, energypool)
		if err != nil {
			statedb.RevertToSnapshot(snapshot)
			log.Info("rejected tx", "index", i, "hash", tx.Hash(), "from", msg.From(), "error", err)
			rejectedTxs = append(rejectedTxs, i)
			continue
		}
		includedTxs = append(includedTxs, tx)
		if hashError != nil {
			return nil, nil, NewError(ErrorMissingBlockhash, hashError)
		}
		energyUsed += msgResult.UsedEnergy
		// Create a new receipt for the transaction, storing the intermediate root and energy used by the tx
		{
			var root []byte
			statedb.Finalise(true)

			receipt := types.NewReceipt(root, msgResult.Failed(), energyUsed)
			receipt.TxHash = tx.Hash()
			receipt.EnergyUsed = msgResult.UsedEnergy
			// if the transaction created a contract, store the creation address in the receipt.
			if msg.To() == nil {
				receipt.ContractAddress = crypto.CreateAddress(cvm.TxContext.Origin, tx.Nonce())
			}
			// Set the receipt logs and create a bloom for filtering
			receipt.Logs = statedb.GetLogs(tx.Hash())
			receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
			// These three are non-consensus fields
			//receipt.BlockHash
			//receipt.BlockNumber =
			receipt.TransactionIndex = uint(txIndex)
			receipts = append(receipts, receipt)
		}
		txIndex++
	}
	statedb.IntermediateRoot(true)
	// Add mining reward?
	if miningReward > 0 {
		// Add mining reward. The mining reward may be `0`, which only makes a difference in the cases
		// where
		// - the coinbase suicided, or
		// - there are only 'bad' transactions, which aren't executed. In those cases,
		//   the coinbase gets no txfee, so isn't created, and thus needs to be touched
		var (
			blockReward = big.NewInt(miningReward)
			minerReward = new(big.Int).Set(blockReward)
			perOmmer    = new(big.Int).Div(blockReward, big.NewInt(32))
		)
		for _, ommer := range pre.Env.Ommers {
			// Add 1/32th for each ommer included
			minerReward.Add(minerReward, perOmmer)
			// Add (8-delta)/8
			reward := big.NewInt(8)
			reward.Sub(reward, big.NewInt(0).SetUint64(ommer.Delta))
			reward.Mul(reward, blockReward)
			reward.Div(reward, big.NewInt(8))
			statedb.AddBalance(ommer.Address, reward)
		}
		statedb.AddBalance(pre.Env.Coinbase, minerReward)
	}
	// Commit block
	root, err := statedb.Commit(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not commit state: %v", err)
		return nil, nil, NewError(ErrorCVM, fmt.Errorf("could not commit state: %v", err))
	}
	execRs := &ExecutionResult{
		StateRoot:   root,
		TxRoot:      types.DeriveSha(includedTxs, new(trie.Trie)),
		ReceiptRoot: types.DeriveSha(receipts, new(trie.Trie)),
		Bloom:       types.CreateBloom(receipts),
		LogsHash:    rlpHash(statedb.Logs()),
		Receipts:    receipts,
		Rejected:    rejectedTxs,
	}
	return statedb, execRs, nil
}

func MakePreState(db xcbdb.Database, accounts core.GenesisAlloc) *state.StateDB {
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
	statedb, _ = state.New(root, sdb, nil)
	return statedb
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.New256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
