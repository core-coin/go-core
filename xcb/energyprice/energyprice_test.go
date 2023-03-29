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

package energyprice

import (
	"context"
	"math"
	"math/big"
	"testing"

	"github.com/core-coin/go-core/v2/consensus/cryptore"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/core"
	"github.com/core-coin/go-core/v2/core/rawdb"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/core/vm"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/params"
	"github.com/core-coin/go-core/v2/rpc"
)

type testBackend struct {
	chain *core.BlockChain
}

func (b *testBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error) {
	if number == rpc.LatestBlockNumber {
		return b.chain.CurrentBlock().Header(), nil
	}
	return b.chain.GetHeaderByNumber(uint64(number)), nil
}

func (b *testBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error) {
	if number == rpc.LatestBlockNumber {
		return b.chain.CurrentBlock(), nil
	}
	return b.chain.GetBlockByNumber(uint64(number)), nil
}

func (b *testBackend) ChainConfig() *params.ChainConfig {
	return b.chain.Config()
}

func newTestBackend(t *testing.T) *testBackend {
	var (
		key, _ = crypto.UnmarshalPrivateKeyHex("89bdfaa2b6f9c30b94ee98fec96c58ff8507fabf49d36a6267e6cb5516eaa2a9e854eccc041f9f67e109d0eb4f653586855355c5b2b87bb313")
		gspec  = &core.Genesis{
			Config: params.TestChainConfig,
			Alloc:  core.GenesisAlloc{key.Address(): {Balance: big.NewInt(math.MaxInt64)}},
		}
		signer = types.NewNucleusSigner(gspec.Config.NetworkID)
	)
	engine := cryptore.NewFaker()
	db := rawdb.NewMemoryDatabase()
	genesis, _ := gspec.Commit(db)

	address, err := common.HexToAddress("cb4900000133700000deadbeef000000000000000000")
	if err != nil {
		t.Error(err)
	}
	// Generate testing blocks
	blocks, _ := core.GenerateChain(params.TestChainConfig, genesis, engine, db, 32, func(i int, b *core.BlockGen) {
		b.SetCoinbase(common.Address{1})
		tx, err := types.SignTx(types.NewTransaction(b.TxNonce(key.Address()), address, big.NewInt(100), 21000, big.NewInt(int64(i+1)*params.Nucle), nil), signer, key)
		if err != nil {
			t.Fatalf("failed to create tx: %v", err)
		}
		b.AddTx(tx)
	})
	// Construct testing chain
	diskdb := rawdb.NewMemoryDatabase()
	gspec.Commit(diskdb)
	chain, err := core.NewBlockChain(diskdb, nil, params.TestChainConfig, engine, vm.Config{}, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create local chain, %v", err)
	}
	chain.InsertChain(blocks)
	return &testBackend{chain: chain}
}

func (b *testBackend) CurrentHeader() *types.Header {
	return b.chain.CurrentHeader()
}

func (b *testBackend) GetBlockByNumber(number uint64) *types.Block {
	return b.chain.GetBlockByNumber(number)
}

func TestSuggestPrice(t *testing.T) {
	config := Config{
		Blocks:     3,
		Percentile: 60,
		Default:    big.NewInt(params.Nucle),
	}
	backend := newTestBackend(t)
	oracle := NewOracle(backend, config)

	// The energy price sampled is: 32G, 31G, 30G, 29G, 28G, 27G
	got, err := oracle.SuggestPrice(context.Background())
	if err != nil {
		t.Fatalf("Failed to retrieve recommended energy price: %v", err)
	}
	expect := big.NewInt(params.Nucle * int64(30))
	if got.Cmp(expect) != 0 {
		t.Fatalf("Energy price mismatch, want %d, got %d", expect, got)
	}
}
