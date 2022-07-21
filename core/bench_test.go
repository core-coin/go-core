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
	"crypto/rand"
	eddsa "github.com/core-coin/go-goldilocks"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/math"
	"github.com/core-coin/go-core/consensus/cryptore"
	"github.com/core-coin/go-core/core/rawdb"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/core/vm"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/params"
	"github.com/core-coin/go-core/xcbdb"
)

func BenchmarkInsertChain_empty_memdb(b *testing.B) {
	benchInsertChain(b, false, nil)
}
func BenchmarkInsertChain_empty_diskdb(b *testing.B) {
	benchInsertChain(b, true, nil)
}
func BenchmarkInsertChain_valueTx_memdb(b *testing.B) {
	benchInsertChain(b, false, genValueTx(0))
}
func BenchmarkInsertChain_valueTx_diskdb(b *testing.B) {
	benchInsertChain(b, true, genValueTx(0))
}
func BenchmarkInsertChain_valueTx_100kB_memdb(b *testing.B) {
	benchInsertChain(b, false, genValueTx(100*1024))
}
func BenchmarkInsertChain_valueTx_100kB_diskdb(b *testing.B) {
	benchInsertChain(b, true, genValueTx(100*1024))
}
func BenchmarkInsertChain_uncles_memdb(b *testing.B) {
	benchInsertChain(b, false, genUncles)
}
func BenchmarkInsertChain_uncles_diskdb(b *testing.B) {
	benchInsertChain(b, true, genUncles)
}
func BenchmarkInsertChain_ring200_memdb(b *testing.B) {
	benchInsertChain(b, false, genTxRing(200))
}
func BenchmarkInsertChain_ring200_diskdb(b *testing.B) {
	benchInsertChain(b, true, genTxRing(200))
}
func BenchmarkInsertChain_ring1000_memdb(b *testing.B) {
	benchInsertChain(b, false, genTxRing(1000))
}
func BenchmarkInsertChain_ring1000_diskdb(b *testing.B) {
	benchInsertChain(b, true, genTxRing(1000))
}

var (
	// This is the content of the genesis block used by the benchmarks.
	benchRootKey, _ = crypto.HexToEDDSA("856a9af6b0b651dd2f43b5e12193652ec1701c4da6f1c0d2a366ac4b9dabc9433ef09e41ca129552bd2c029086d9b03604de872a3b3432041f")
	benchRootPub    = eddsa.Ed448DerivePublicKey(*benchRootKey)
	benchRootAddr   = crypto.PubkeyToAddress(benchRootPub)
	benchRootFunds  = math.BigPow(2, 100)
)

// genValueTx returns a block generator that includes a single
// value-transfer transaction with n bytes of extra data in each
// block.
func genValueTx(nbytes int) func(int, *BlockGen) {
	return func(i int, gen *BlockGen) {
		toaddr := common.Address{}
		data := make([]byte, nbytes)
		energy, _ := IntrinsicEnergy(data, false)
		tx, _ := types.SignTx(types.NewTransaction(gen.TxNonce(benchRootAddr), toaddr, big.NewInt(1), energy, nil, data), types.NewNucleusSigner(params.AllCryptoreProtocolChanges.NetworkID), benchRootKey)
		gen.AddTx(tx)
	}
}

var (
	ringKeys  = make([]*eddsa.PrivateKey, 1000)
	ringAddrs = make([]common.Address, len(ringKeys))
)

func init() {
	ringKeys[0] = benchRootKey
	ringAddrs[0] = benchRootAddr
	for i := 1; i < len(ringKeys); i++ {
		ringKeys[i], _ = crypto.GenerateKey(rand.Reader)
		pub := eddsa.Ed448DerivePublicKey(*ringKeys[i])
		ringAddrs[i] = crypto.PubkeyToAddress(pub)
	}
}

// genTxRing returns a block generator that sends core in a ring
// among n accounts. This is creates n entries in the state database
// and fills the blocks with many small transactions.
func genTxRing(naccounts int) func(int, *BlockGen) {
	from := 0
	return func(i int, gen *BlockGen) {
		block := gen.PrevBlock(i - 1)
		energy := CalcEnergyLimit(block, block.EnergyLimit(), block.EnergyLimit())
		for {
			energy -= params.TxEnergy
			if energy < params.TxEnergy {
				break
			}
			to := (from + 1) % naccounts
			tx := types.NewTransaction(
				gen.TxNonce(ringAddrs[from]),
				ringAddrs[to],
				benchRootFunds,
				params.TxEnergy,
				nil,
				nil,
			)
			tx, _ = types.SignTx(tx, types.NewNucleusSigner(params.AllCryptoreProtocolChanges.NetworkID), ringKeys[from])
			gen.AddTx(tx)
			from = to
		}
	}
}

// genUncles generates blocks with two uncle headers.
func genUncles(i int, gen *BlockGen) {
	if i >= 6 {
		b2 := gen.PrevBlock(i - 6).Header()
		b2.Extra = []byte("foo")
		gen.AddUncle(b2)
		b3 := gen.PrevBlock(i - 6).Header()
		b3.Extra = []byte("bar")
		gen.AddUncle(b3)
	}
}

func benchInsertChain(b *testing.B, disk bool, gen func(int, *BlockGen)) {
	// Create the database in memory or in a temporary directory.
	var db xcbdb.Database
	if !disk {
		db = rawdb.NewMemoryDatabase()
	} else {
		dir, err := ioutil.TempDir("", "xcb-core-bench")
		if err != nil {
			b.Fatalf("cannot create temporary directory: %v", err)
		}
		defer os.RemoveAll(dir)
		db, err = rawdb.NewLevelDBDatabase(dir, 128, 128, "")
		if err != nil {
			b.Fatalf("cannot create temporary database: %v", err)
		}
		defer db.Close()
	}

	// Generate a chain of b.N blocks using the supplied block
	// generator function.
	gspec := Genesis{
		Config: params.TestChainConfig,
		Alloc:  GenesisAlloc{benchRootAddr: {Balance: benchRootFunds}},
	}
	genesis := gspec.MustCommit(db)
	chain, _ := GenerateChain(gspec.Config, genesis, cryptore.NewFaker(), db, b.N, gen)

	// Time the insertion of the new chain.
	// State and blocks are stored in the same DB.
	chainman, _ := NewBlockChain(db, nil, gspec.Config, cryptore.NewFaker(), vm.Config{}, nil, nil)
	defer chainman.Stop()
	b.ReportAllocs()
	b.ResetTimer()
	if i, err := chainman.InsertChain(chain); err != nil {
		b.Fatalf("insert error (block %d): %v\n", i, err)
	}
}

func BenchmarkChainRead_header_10k(b *testing.B) {
	benchReadChain(b, false, 10000)
}
func BenchmarkChainRead_full_10k(b *testing.B) {
	benchReadChain(b, true, 10000)
}
func BenchmarkChainRead_header_100k(b *testing.B) {
	benchReadChain(b, false, 100000)
}
func BenchmarkChainRead_full_100k(b *testing.B) {
	benchReadChain(b, true, 100000)
}
func BenchmarkChainRead_header_500k(b *testing.B) {
	benchReadChain(b, false, 500000)
}
func BenchmarkChainRead_full_500k(b *testing.B) {
	benchReadChain(b, true, 500000)
}
func BenchmarkChainWrite_header_10k(b *testing.B) {
	benchWriteChain(b, false, 10000)
}
func BenchmarkChainWrite_full_10k(b *testing.B) {
	benchWriteChain(b, true, 10000)
}
func BenchmarkChainWrite_header_100k(b *testing.B) {
	benchWriteChain(b, false, 100000)
}
func BenchmarkChainWrite_full_100k(b *testing.B) {
	benchWriteChain(b, true, 100000)
}
func BenchmarkChainWrite_header_500k(b *testing.B) {
	benchWriteChain(b, false, 500000)
}
func BenchmarkChainWrite_full_500k(b *testing.B) {
	benchWriteChain(b, true, 500000)
}

// makeChainForBench writes a given number of headers or empty blocks/receipts
// into a database.
func makeChainForBench(db xcbdb.Database, full bool, count uint64) {
	var hash common.Hash
	for n := uint64(0); n < count; n++ {
		header := &types.Header{
			Coinbase:    common.Address{},
			Number:      big.NewInt(int64(n)),
			ParentHash:  hash,
			Difficulty:  big.NewInt(1),
			UncleHash:   types.EmptyUncleHash,
			TxHash:      types.EmptyRootHash,
			ReceiptHash: types.EmptyRootHash,
		}
		hash = header.Hash()

		rawdb.WriteHeader(db, header)
		rawdb.WriteCanonicalHash(db, hash, n)
		rawdb.WriteTd(db, hash, n, big.NewInt(int64(n+1)))

		if full || n == 0 {
			block := types.NewBlockWithHeader(header)
			rawdb.WriteBody(db, hash, n, block.Body())
			rawdb.WriteReceipts(db, hash, n, nil)
		}
	}
}

func benchWriteChain(b *testing.B, full bool, count uint64) {
	for i := 0; i < b.N; i++ {
		dir, err := ioutil.TempDir("", "xcb-chain-bench")
		if err != nil {
			b.Fatalf("cannot create temporary directory: %v", err)
		}
		db, err := rawdb.NewLevelDBDatabase(dir, 128, 1024, "")
		if err != nil {
			b.Fatalf("error opening database at %v: %v", dir, err)
		}
		makeChainForBench(db, full, count)
		db.Close()
		os.RemoveAll(dir)
	}
}

func benchReadChain(b *testing.B, full bool, count uint64) {
	dir, err := ioutil.TempDir("", "xcb-chain-bench")
	if err != nil {
		b.Fatalf("cannot create temporary directory: %v", err)
	}
	defer os.RemoveAll(dir)

	db, err := rawdb.NewLevelDBDatabase(dir, 128, 1024, "")
	if err != nil {
		b.Fatalf("error opening database at %v: %v", dir, err)
	}
	makeChainForBench(db, full, count)
	db.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		db, err := rawdb.NewLevelDBDatabase(dir, 128, 1024, "")
		if err != nil {
			b.Fatalf("error opening database at %v: %v", dir, err)
		}
		chain, err := NewBlockChain(db, nil, params.TestChainConfig, cryptore.NewFaker(), vm.Config{}, nil, nil)
		if err != nil {
			b.Fatalf("error creating chain: %v", err)
		}

		for n := uint64(0); n < count; n++ {
			header := chain.GetHeaderByNumber(n)
			if full {
				hash := header.Hash()
				rawdb.ReadBody(db, hash, n)
				rawdb.ReadReceipts(db, hash, n, chain.Config())
			}
		}
		chain.Stop()
		db.Close()
	}
}
