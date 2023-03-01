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

package core

import (
	"math/big"
	"testing"

	"golang.org/x/crypto/sha3"

	"github.com/core-coin/go-core/v2/consensus/cryptore"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/consensus"
	"github.com/core-coin/go-core/v2/core/rawdb"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/core/vm"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/params"
	"github.com/core-coin/go-core/v2/trie"
)

// TestStateProcessorErrors tests the output from the 'core' errors
// as defined in core/error.go. These errors are generated when the
// blockchain imports bad blocks, meaning blocks which have valid headers but
// contain invalid transactions
func TestStateProcessorErrors(t *testing.T) {
	var (
		signer     = types.NewNucleusSigner(big.NewInt(1337))
		testKey, _ = crypto.UnmarshalPrivateKeyHex("89bdfaa2b6f9c30b94ee98fec96c58ff8507fabf49d36a6267e6cb5516eaa2a9e854eccc041f9f67e109d0eb4f653586855355c5b2b87bb313")
		db         = rawdb.NewMemoryDatabase()
		gspec      = &Genesis{
			Config: params.TestChainConfig,
		}
		genesis         = gspec.MustCommit(db)
		blockchain, err = NewBlockChain(db, nil, gspec.Config, cryptore.NewFaker(), vm.Config{}, nil, nil)
	)
	if err != nil {
		t.Error(err)
	}
	defer blockchain.Stop()
	var makeTx = func(nonce uint64, to common.Address, amount *big.Int, energyLimit uint64, energyPrice *big.Int, data []byte) *types.Transaction {
		tx, err := types.SignTx(types.NewTransaction(nonce, to, amount, energyLimit, energyPrice, data), signer, testKey)
		if err != nil {
			t.Error(err)
		}
		return tx
	}
	for i, tt := range []struct {
		txs  []*types.Transaction
		want string
	}{
		{
			txs: []*types.Transaction{
				makeTx(0, common.Address{}, big.NewInt(0), params.TxEnergy, nil, nil),
				makeTx(0, common.Address{}, big.NewInt(0), params.TxEnergy, nil, nil),
			},
			want: "could not apply tx 1 [0x6525177df11954a61e2f7538b58c145fc32fd406e0403d560c8a4e7bc68fd936]: nonce too low: address cb53c378bf81ade6f8e505ac7c298c84f7709f9b5a4e, tx: 0 state: 1",
		},
		{
			txs: []*types.Transaction{
				makeTx(100, common.Address{}, big.NewInt(0), params.TxEnergy, nil, nil),
			},
			want: "could not apply tx 0 [0x51be2588d67d8916e5eea2f02706cdfe9cc8400d78683ea121369ff2382fb4b5]: nonce too high: address cb53c378bf81ade6f8e505ac7c298c84f7709f9b5a4e, tx: 100 state: 0",
		},
		{
			txs: []*types.Transaction{
				makeTx(0, common.Address{}, big.NewInt(0), 21000000, nil, nil),
			},
			want: "could not apply tx 0 [0x4c4f73f64c4ed52fbc996c318443ed9853093be4b3467515ce4276c8db12d257]: energy limit reached",
		},
		{
			txs: []*types.Transaction{
				makeTx(0, common.Address{}, big.NewInt(1), params.TxEnergy, nil, nil),
			},
			want: "could not apply tx 0 [0x3325751de5ba3eb5e13a749b8f8d26c865bbd433eddf881fbb33b12ec884dbe5]: insufficient funds for transfer: address cb53c378bf81ade6f8e505ac7c298c84f7709f9b5a4e",
		},
		{
			txs: []*types.Transaction{
				makeTx(0, common.Address{}, big.NewInt(0), params.TxEnergy, big.NewInt(0xffffff), nil),
			},
			want: "could not apply tx 0 [0x2fcbc6ea0f0f35dd730f6ca030290064d442adc779ffd89ab8b26bef7ac77225]: insufficient funds for energy * price + value: address cb53c378bf81ade6f8e505ac7c298c84f7709f9b5a4e have 0 want 352321515000",
		},
		{
			txs: []*types.Transaction{
				makeTx(0, common.Address{}, big.NewInt(0), params.TxEnergy, nil, nil),
				makeTx(1, common.Address{}, big.NewInt(0), params.TxEnergy, nil, nil),
				makeTx(2, common.Address{}, big.NewInt(0), params.TxEnergy, nil, nil),
				makeTx(3, common.Address{}, big.NewInt(0), params.TxEnergy-1000, big.NewInt(0), nil),
			},
			want: "could not apply tx 3 [0x9adbf986d47b21a49e2fac59e20481bf50e66ffdffad2739dc09499b0a3b5b62]: intrinsic energy too low: have 20000, want 21000",
		},
		// The last 'core' error is ErrEnergyUintOverflow: "energy uint64 overflow", but in order to
		// trigger that one, we'd have to allocate a _huge_ chunk of data, such that the
		// multiplication len(data) +energy_per_byte overflows uint64. Not testable at the moment
	} {
		block := GenerateBadBlock(genesis, cryptore.NewFaker(), tt.txs)
		_, err := blockchain.InsertChain(types.Blocks{block})
		if err == nil {
			t.Fatal("block imported without errors")
		}
		if have, want := err.Error(), tt.want; have != want {
			t.Errorf("test %d:\nhave \"%v\"\nwant \"%v\"\n", i, have, want)
		}
	}
}

// GenerateBadBlock constructs a "block" which contains the transactions. The transactions are not expected to be
// valid, and no proper post-state can be made. But from the perspective of the blockchain, the block is sufficiently
// valid to be considered for import:
// - valid pow (fake), ancestry, difficulty, energylimit etc
func GenerateBadBlock(parent *types.Block, engine consensus.Engine, txs types.Transactions) *types.Block {
	header := &types.Header{
		ParentHash: parent.Hash(),
		Coinbase:   parent.Coinbase(),
		Difficulty: engine.CalcDifficulty(&fakeChainReader{params.TestChainConfig}, parent.Time()+10, &types.Header{
			Number:     parent.Number(),
			Time:       parent.Time(),
			Difficulty: parent.Difficulty(),
			UncleHash:  parent.UncleHash(),
		}),
		EnergyLimit: CalcEnergyLimit(parent, parent.EnergyLimit(), parent.EnergyLimit()),
		Number:      new(big.Int).Add(parent.Number(), common.Big1),
		Time:        parent.Time() + 10,
		UncleHash:   types.EmptyUncleHash,
	}
	var receipts []*types.Receipt

	// The post-state result doesn't need to be correct (this is a bad block), but we do need something there
	// Preferably something unique. So let's use a combo of blocknum + txhash
	hasher := sha3.New256()
	hasher.Write(header.Number.Bytes())
	var cumulativeEnergy uint64
	for _, tx := range txs {
		txh := tx.Hash()
		hasher.Write(txh[:])
		receipt := types.NewReceipt(nil, false, cumulativeEnergy+tx.Energy())
		receipt.TxHash = tx.Hash()
		receipt.EnergyUsed = tx.Energy()
		receipts = append(receipts, receipt)
		cumulativeEnergy += tx.Energy()
	}
	header.Root = common.BytesToHash(hasher.Sum(nil))
	// Assemble and return the final block for sealing
	return types.NewBlock(header, txs, nil, receipts, new(trie.Trie))
}
