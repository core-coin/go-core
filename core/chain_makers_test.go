// Copyright 2015 by the Authors
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
	"fmt"
	"math/big"

	"github.com/core-coin/go-core/v2/consensus/cryptore"

	"github.com/core-coin/go-core/v2/core/rawdb"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/core/vm"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/params"
)

func ExampleGenerateChain() {
	var (
		key1, _ = crypto.UnmarshalPrivateKeyHex("89bdfaa2b6f9c30b94ee98fec96c58ff8507fabf49d36a6267e6cb5516eaa2a9e854eccc041f9f67e109d0eb4f653586855355c5b2b87bb313")
		key2, _ = crypto.UnmarshalPrivateKeyHex("ab856a9af6b0b651dd2f43b5e12193652ec1701c4da6f1c0d2a366ac4b9dabc9433ef09e41ca129552bd2c029086d9b03604de872a3b343204")
		key3, _ = crypto.UnmarshalPrivateKeyHex("c0b711eea422df26d5ffdcaae35fe0527cf647c5ce62d3efb5e09a0e14fc8afe57fac1a5daa330bc10bfa1d3db11e172a822dcfffb86a0b26d")
		db      = rawdb.NewMemoryDatabase()
	)

	// Ensure that key1 has some funds in the genesis block.
	gspec := &Genesis{
		Config: &params.ChainConfig{NetworkID: big.NewInt(1)},
		Alloc:  GenesisAlloc{key1.Address(): {Balance: big.NewInt(1000000)}},
	}
	genesis := gspec.MustCommit(db)

	// This call generates a chain of 5 blocks. The function runs for
	// each block and adds different features to gen based on the
	// block index.
	signer := types.NewNucleusSigner(big.NewInt(1))
	chain, _ := GenerateChain(gspec.Config, genesis, cryptore.NewFaker(), db, 5, func(i int, gen *BlockGen) {
		switch i {
		case 0:
			// In block 1, key1.Address() sends key2.Address() some core.
			tx, _ := types.SignTx(types.NewTransaction(gen.TxNonce(key1.Address()), key2.Address(), big.NewInt(10000), params.TxEnergy, nil, nil), signer, key1)
			gen.AddTx(tx)
		case 1:
			// In block 2, key1.Address() sends some more core to key2.Address().
			// key2.Address() passes it on to key3.Address().
			tx1, _ := types.SignTx(types.NewTransaction(gen.TxNonce(key1.Address()), key2.Address(), big.NewInt(1000), params.TxEnergy, nil, nil), signer, key1)
			tx2, _ := types.SignTx(types.NewTransaction(gen.TxNonce(key2.Address()), key3.Address(), big.NewInt(1000), params.TxEnergy, nil, nil), signer, key2)
			gen.AddTx(tx1)
			gen.AddTx(tx2)
		case 2:
			// Block 3 is empty but was mined by key3.Address().
			gen.SetCoinbase(key3.Address())
			gen.SetExtra([]byte("yeehaw"))
		case 3:
			// Block 4 includes blocks 2 and 3 as uncle headers (with modified extra data).
			b2 := gen.PrevBlock(1).Header()
			b2.Extra = []byte("foo")
			gen.AddUncle(b2)
			b3 := gen.PrevBlock(2).Header()
			b3.Extra = []byte("foo")
			gen.AddUncle(b3)
		}
	})

	// Import the chain. This runs all block validation rules.
	blockchain, _ := NewBlockChain(db, nil, gspec.Config, cryptore.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(chain); err != nil {
		fmt.Printf("insert error (block %d): %v\n", chain[i].NumberU64(), err)
		return
	}

	state, _ := blockchain.State()
	fmt.Printf("last block: #%d\n", blockchain.CurrentBlock().Number())
	fmt.Println("balance of key1.Address():", state.GetBalance(key1.Address()))
	fmt.Println("balance of key2.Address():", state.GetBalance(key2.Address()))
	fmt.Println("balance of key3.Address():", state.GetBalance(key3.Address()))
	// Output:
	// last block: #5
	// balance of key1.Address(): 989000
	// balance of key2.Address(): 10000
	// balance of key3.Address(): 19687500000000001000
}
