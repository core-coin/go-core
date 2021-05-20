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
	eddsa "github.com/core-coin/go-goldilocks"
	"math/big"

	"github.com/core-coin/go-core/consensus/cryptore"
	"github.com/core-coin/go-core/core/rawdb"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/core/vm"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/params"
)

func ExampleGenerateChain() {
	var (
		key1, _ = crypto.HexToEDDSA("856a9af6b0b651dd2f43b5e12193652ec1701c4da6f1c0d2a366ac4b9dabc9433ef09e41ca129552bd2c029086d9b03604de872a3b3432041f")
		key2, _ = crypto.HexToEDDSA("c7b3545db244c1ea1c720086c2c4c9f5eff2f0f31263101f0e8486201e6605414c240fe851d5fd0b4122b764e4cb7ef02695bfd9aed9d00cc5")
		key3, _ = crypto.HexToEDDSA("ec4f51f2db12a88c2675cb1241e83b83dbe13df604a4c3d4d4482099273e2b07e2e812ed9d035938d5c0a5ee1c4be5602a3fb82cfe6a9b2383")
		pub1    = eddsa.Ed448DerivePublicKey(*key1)
		pub2    = eddsa.Ed448DerivePublicKey(*key2)
		pub3    = eddsa.Ed448DerivePublicKey(*key3)
		addr1   = crypto.PubkeyToAddress(pub1)
		addr2   = crypto.PubkeyToAddress(pub2)
		addr3   = crypto.PubkeyToAddress(pub3)
		db      = rawdb.NewMemoryDatabase()
	)

	// Ensure that key1 has some funds in the genesis block.
	gspec := &Genesis{
		Config: &params.ChainConfig{},
		Alloc:  GenesisAlloc{addr1: {Balance: big.NewInt(1000000)}},
	}
	genesis := gspec.MustCommit(db)

	// This call generates a chain of 5 blocks. The function runs for
	// each block and adds different features to gen based on the
	// block index.
	signer := types.NewNucleusSigner(gspec.Config.NetworkID)
	chain, _ := GenerateChain(gspec.Config, genesis, cryptore.NewFaker(), db, 5, func(i int, gen *BlockGen) {
		switch i {
		case 0:
			// In block 1, addr1 sends addr2 some core.
			tx, _ := types.SignTx(types.NewTransaction(gen.TxNonce(addr1), addr2, big.NewInt(10000), params.TxEnergy, nil, nil), signer, key1)
			gen.AddTx(tx)
		case 1:
			// In block 2, addr1 sends some more core to addr2.
			// addr2 passes it on to addr3.
			tx1, _ := types.SignTx(types.NewTransaction(gen.TxNonce(addr1), addr2, big.NewInt(1000), params.TxEnergy, nil, nil), signer, key1)
			tx2, _ := types.SignTx(types.NewTransaction(gen.TxNonce(addr2), addr3, big.NewInt(1000), params.TxEnergy, nil, nil), signer, key2)
			gen.AddTx(tx1)
			gen.AddTx(tx2)
		case 2:
			// Block 3 is empty but was mined by addr3.
			gen.SetCoinbase(addr3)
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
	blockchain, _ := NewBlockChain(db, nil, gspec.Config, cryptore.NewFaker(), vm.Config{}, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(chain); err != nil {
		fmt.Printf("insert error (block %d): %v\n", chain[i].NumberU64(), err)
		return
	}

	state, _ := blockchain.State()
	fmt.Printf("last block: #%d\n", blockchain.CurrentBlock().Number())
	fmt.Println("balance of addr1:", state.GetBalance(addr1))
	fmt.Println("balance of addr2:", state.GetBalance(addr2))
	fmt.Println("balance of addr3:", state.GetBalance(addr3))
	// Output:
	// last block: #5
	// balance of addr1: 989000
	// balance of addr2: 10000
	// balance of addr3: 19687500000000001000
}
