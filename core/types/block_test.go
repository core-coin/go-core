// Copyright 2014 The go-core Authors
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
	"math/big"
	"reflect"
	"testing"

	"github.com/core-coin/go-core/params"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/rlp"
)

// from bcValidBlockTest.json, "SimpleTx"
func TestBlockEncoding(t *testing.T) {
	key, _ := crypto.HexToEDDSA("2da94fd47e8369ffe88850654de266727ff284c3f78d61b04153cb9a908ed3b61248ac5172d3caabbc3493807c0297645ae328e10eb9543bdbcc413b5871d83426cd5b3a0083e6f589f60c1177b287b8f4f764acfcba7dfceadc51ef37b40d6182e3fe6bce148c8a48e07379754ebbf83236643837663566326266393833356639613437656566616535373162633039")

	blockEnc := common.FromHex("f902ccf901fba00000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000096cb278888f1f195afa192cfee860698584c030f4c9db1a0ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008302000080832fefd8825208845506eb0780a0bd4472abb6659ebe3ee06ee4d7b72a00a9f4d001caca51342001075469aff49888a13a5a8c8f2bb1c4f8cbf8c9800a82c3500196cb84f8989a12c39d55e40ca9466fbc3963d9914a44880a80b8a84564a6d2b24573ce91b4d364f90eb56daac2ad03421ae1a265240951beb8826e4c54eadb83ea3bf01f1ccb8052fc06f407caea802c383a6cb73bb86fe2c0a7a3b440db7107a9dce5090bc53d54a8163b95b7318b20fc0840826eae593aeece09af8001c6fb0b092335b01656edb46b15dbcc413b5871d83426cd5b3a0083e6f589f60c1177b287b8f4f764acfcba7dfceadc51ef37b40d6182e3fe6bce148c8a48e07379754ebbf8c0")
	var block Block
	if err := rlp.DecodeBytes(blockEnc, &block); err != nil {
		t.Fatal("decode error: ", err)
	}

	check := func(f string, got, want interface{}) {
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s mismatch: got %v, want %v", f, got, want)
		}
	}
	coinbase, _ := common.HexToAddress("cb278888f1f195afa192cfee860698584c030f4c9db1")
	check("Difficulty", block.Difficulty(), big.NewInt(131072))
	check("EnergyLimit", block.EnergyLimit(), uint64(3141592))
	check("EnergyUsed", block.EnergyUsed(), uint64(21000))
	check("Coinbase", block.Coinbase(), coinbase)
	check("MixDigest", block.MixDigest(), common.HexToHash("bd4472abb6659ebe3ee06ee4d7b72a00a9f4d001caca51342001075469aff498"))
	check("Root", block.Root(), common.HexToHash("ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017"))
	check("Hash", block.Hash(), common.HexToHash("73dfd9a034149c1174ea0216b84c006f3331441560d2b4ed6e9add53031afcc9"))
	check("Nonce", block.Nonce(), uint64(0xa13a5a8c8f2bb1c4))
	check("Time", block.Time(), uint64(1426516743))
	check("Size", block.Size(), common.StorageSize(len(blockEnc)))

	tx1 := NewTransaction(0, crypto.PubkeyToAddress(key.PublicKey), big.NewInt(10), 50000, big.NewInt(10), nil)
	tx1, _ = tx1.WithSignature(NewNucleusSigner(params.AllCryptoreProtocolChanges.ChainID), common.Hex2Bytes("4564a6d2b24573ce91b4d364f90eb56daac2ad03421ae1a265240951beb8826e4c54eadb83ea3bf01f1ccb8052fc06f407caea802c383a6cb73bb86fe2c0a7a3b440db7107a9dce5090bc53d54a8163b95b7318b20fc0840826eae593aeece09af8001c6fb0b092335b01656edb46b15dbcc413b5871d83426cd5b3a0083e6f589f60c1177b287b8f4f764acfcba7dfceadc51ef37b40d6182e3fe6bce148c8a48e07379754ebbf8"))
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
	exp := common.HexToHash("1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347")
	if h != exp {
		t.Fatalf("empty uncle hash is wrong, got %x != %x", h, exp)
	}
}
func BenchmarkUncleHash(b *testing.B) {
	uncles := make([]*Header, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalcUncleHash(uncles)
	}
}
