// Copyright 2014 by the Authors
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

package types

import (
	"bytes"
	crand "crypto/rand"
	"hash"
	"math/big"
	"reflect"
	"testing"

	"golang.org/x/crypto/sha3"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/math"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/params"
	"github.com/core-coin/go-core/v2/rlp"
)

// from bcValidBlockTest.json, "SimpleTx"
func TestBlockEncoding(t *testing.T) {
	blockEnc := common.FromHex("f904b4f901daa00000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000096cb848888f1f195afa192cfee860698584c030f4c9db1a0ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008302000080832fefd8825208845506eb078088a13a5a8c8f2bb1c4f902d3f902d0800a82c3500196cb8238748ee459bc0c1d86eab1d3f6d83bb433cdad9c0a80b902aef902abf901daa00000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000096cb848888f1f195afa192cfee860698584c030f4c9db1a0ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008302000080832fefd8825208845506eb078088a13a5a8c8f2bb1c4f8cbf8c9800a82c3500196cb723ce833507cdfb66b76332523f034b38d8b077f560a80b8a8ac4f81b3c3fc94b5deccc435a1dda8408dc2d4f36cb8cb5d88d80ab0d93a9129cd5ffc28d591c101ed4725809f3635f5cf554599ea8f751af59b43497b6f956eb346946ae69648f3f9f6a1df3e5828486219a3c64fce0fd4b1de04696d7abfb8fabf0a5e3c655c02d839eb4f09308722dbcc413b5871d83426cd5b3a0083e6f589f60c1177b287b8f4f764acfcba7dfceadc51ef37b40d6182e3fe6bce148c8a48e07379754ebbf8c0c0")
	var block Block
	if err := rlp.DecodeBytes(blockEnc, &block); err != nil {
		t.Fatal("decode error: ", err)
	}

	check := func(f string, got, want interface{}) {
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s mismatch: got %v, want %v", f, got, want)
		}
	}
	coinbase, _ := common.HexToAddress("cb848888f1f195afa192cfee860698584c030f4c9db1")
	check("Difficulty", block.Difficulty(), big.NewInt(131072))
	check("EnergyLimit", block.EnergyLimit(), uint64(3141592))
	check("EnergyUsed", block.EnergyUsed(), uint64(21000))
	check("Coinbase", block.Coinbase(), coinbase)
	check("Root", block.Root(), common.HexToHash("ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017"))
	check("Hash", block.Hash(), common.HexToHash("cfc53a2640132939c41833cf8f194cfc5f7a929c4a265cda567e63f3e0dd5307"))
	check("Nonce", block.Nonce(), uint64(0xa13a5a8c8f2bb1c4))
	check("Time", block.Time(), uint64(1426516743))
	check("Size", block.Size(), common.StorageSize(len(blockEnc)))

	tx1 := NewTransaction(0, key.Address(), big.NewInt(10), 50000, big.NewInt(10), nil)
	tx1, _ = tx1.WithSignature(NewNucleusSigner(params.MainnetChainConfig.NetworkID), common.Hex2Bytes("f902abf901daa00000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000096cb848888f1f195afa192cfee860698584c030f4c9db1a0ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008302000080832fefd8825208845506eb078088a13a5a8c8f2bb1c4f8cbf8c9800a82c3500196cb723ce833507cdfb66b76332523f034b38d8b077f560a80b8a8ac4f81b3c3fc94b5deccc435a1dda8408dc2d4f36cb8cb5d88d80ab0d93a9129cd5ffc28d591c101ed4725809f3635f5cf554599ea8f751af59b43497b6f956eb346946ae69648f3f9f6a1df3e5828486219a3c64fce0fd4b1de04696d7abfb8fabf0a5e3c655c02d839eb4f09308722dbcc413b5871d83426cd5b3a0083e6f589f60c1177b287b8f4f764acfcba7dfceadc51ef37b40d6182e3fe6bce148c8a48e07379754ebbf8c0"))
	check("len(Transactions)", len(block.Transactions()), 1)
	check("Transactions[0].Hash", block.Transactions()[0].Hash(), tx1.Hash())

	ourBlockEnc, err := rlp.EncodeToBytes(&block)
	if err != nil {
		t.Fatal("encode error: ", err)
	}
	if !bytes.Equal(ourBlockEnc, blockEnc) {
		t.Errorf("encoded block mismatch:\ngot:  %x\nwant: %x", ourBlockEnc, blockEnc)
	}
}

func TestUncleHash(t *testing.T) {
	uncles := make([]*Header, 0)
	h := CalcUncleHash(uncles)
	exp := common.HexToHash("f18f47848fb293468f641c33863dca9e5278fa8e9690f77f7dc96e954ef9221b")
	if h != exp {
		t.Fatalf("empty uncle hash is wrong, got %x != %x", h, exp)
	}
}

var benchBuffer = bytes.NewBuffer(make([]byte, 0, 32000))

func BenchmarkEncodeBlock(b *testing.B) {
	block := makeBenchBlock()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchBuffer.Reset()
		if err := rlp.Encode(benchBuffer, block); err != nil {
			b.Fatal(err)
		}
	}
}

// testHasher is the helper tool for transaction/receipt list hashing.
// The original hasher is trie, in order to get rid of import cycle,
// use the testing hasher instead.
type testHasher struct {
	hasher hash.Hash
}

func newHasher() *testHasher {
	return &testHasher{hasher: sha3.New256()}
}

func (h *testHasher) Reset() {
	h.hasher.Reset()
}

func (h *testHasher) Update(key, val []byte) {
	h.hasher.Write(key)
	h.hasher.Write(val)
}

func (h *testHasher) Hash() common.Hash {
	return common.BytesToHash(h.hasher.Sum(nil))
}

func makeBenchBlock() *Block {
	var (
		key, _   = crypto.GenerateKey(crand.Reader)
		txs      = make([]*Transaction, 70)
		receipts = make([]*Receipt, len(txs))
		signer   = NewNucleusSigner(params.MainnetChainConfig.NetworkID)
		uncles   = make([]*Header, 3)
	)
	header := &Header{
		Difficulty:  math.BigPow(11, 11),
		Number:      math.BigPow(2, 9),
		EnergyLimit: 12345678,
		EnergyUsed:  1476322,
		Time:        9876543,
		Extra:       []byte("coolest block on chain"),
	}
	for i := range txs {
		amount := math.BigPow(2, int64(i))
		price := big.NewInt(300000)
		data := make([]byte, 100)
		tx := NewTransaction(uint64(i), common.Address{}, amount, 123457, price, data)
		signedTx, err := SignTx(tx, signer, key)
		if err != nil {
			panic(err)
		}
		txs[i] = signedTx
		receipts[i] = NewReceipt(make([]byte, 32), false, tx.Energy())
	}
	for i := range uncles {
		uncles[i] = &Header{
			Difficulty:  math.BigPow(11, 11),
			Number:      math.BigPow(2, 9),
			EnergyLimit: 12345678,
			EnergyUsed:  1476322,
			Time:        9876543,
			Extra:       []byte("benchmark uncle"),
		}
	}
	return NewBlock(header, txs, uncles, receipts, newHasher())
}
