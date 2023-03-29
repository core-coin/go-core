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
	"fmt"
	"math/big"
	"testing"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/crypto"
)

func TestBloom(t *testing.T) {
	positive := []string{
		"testtest",
		"test",
		"hallo",
		"other",
	}
	negative := []string{
		"tes",
		"lo",
	}

	var bloom Bloom
	for _, data := range positive {
		bloom.Add([]byte(data))
	}

	for _, data := range positive {
		if !bloom.Test([]byte(data)) {
			t.Error("expected", data, "to test true")
		}
	}
	for _, data := range negative {
		if bloom.Test([]byte(data)) {
			t.Error("did not expect", data, "to test true")
		}
	}
}

// TestBloomExtensively does some more thorough tests
func TestBloomExtensively(t *testing.T) {
	var exp = common.HexToHash("9b550e165549865e51115278075c198fd1c38d5a782d583f67c4f758963b86ac")
	var b Bloom
	// Add 100 "random" things
	for i := 0; i < 100; i++ {
		data := fmt.Sprintf("xxxxxxxxxx data %d yyyyyyyyyyyyyy", i)
		b.Add([]byte(data))
		//b.Add(new(big.Int).SetBytes([]byte(data)))
	}
	got := crypto.SHA3Hash(b.Bytes())
	if got != exp {
		t.Errorf("Got %x, exp %x", got, exp)
	}
	var b2 Bloom
	b2.SetBytes(b.Bytes())
	got2 := crypto.SHA3Hash(b2.Bytes())
	if got != got2 {
		t.Errorf("Got %x, exp %x", got, got2)
	}
}

func BenchmarkBloom9(b *testing.B) {
	test := []byte("testestestest")
	for i := 0; i < b.N; i++ {
		Bloom9(test)
	}
}

func BenchmarkBloom9Lookup(b *testing.B) {
	toTest := []byte("testtest")
	bloom := new(Bloom)
	for i := 0; i < b.N; i++ {
		bloom.Test(toTest)
	}
}

func BenchmarkCreateBloom(b *testing.B) {
	addr, err := common.HexToAddress("cb970000000000000000000000000000000000000002")
	if err != nil {
		b.Fatal(err)
	}
	var txs = Transactions{
		NewContractCreation(1, big.NewInt(1), 1, big.NewInt(1), nil),
		NewTransaction(2, addr, big.NewInt(2), 2, big.NewInt(2), nil),
	}
	var rSmall = Receipts{
		&Receipt{
			Status:               ReceiptStatusFailed,
			CumulativeEnergyUsed: 1,
			Logs: []*Log{
				{Address: common.BytesToAddress([]byte{0x11})},
				{Address: common.BytesToAddress([]byte{0x01, 0x11})},
			},
			TxHash:          txs[0].Hash(),
			ContractAddress: common.BytesToAddress([]byte{0x01, 0x11, 0x11}),
			EnergyUsed:      1,
		},
		&Receipt{
			PostState:            common.Hash{2}.Bytes(),
			CumulativeEnergyUsed: 3,
			Logs: []*Log{
				{Address: common.BytesToAddress([]byte{0x22})},
				{Address: common.BytesToAddress([]byte{0x02, 0x22})},
			},
			TxHash:          txs[1].Hash(),
			ContractAddress: common.BytesToAddress([]byte{0x02, 0x22, 0x22}),
			EnergyUsed:      2,
		},
	}

	var rLarge = make(Receipts, 200)
	// Fill it with 200 receipts x 2 logs
	for i := 0; i < 200; i += 2 {
		copy(rLarge[i:], rSmall)
	}
	b.Run("small", func(b *testing.B) {
		b.ReportAllocs()
		var bl Bloom
		for i := 0; i < b.N; i++ {
			bl = CreateBloom(rSmall)
		}
		b.StopTimer()
		var exp = common.HexToHash("c384c56ece49458a427c67b90fefe979ebf7104795be65dc398b280f24104949")
		got := crypto.SHA3Hash(bl.Bytes())
		if got != exp {
			b.Errorf("Got %x, exp %x", got, exp)
		}
	})
	b.Run("large", func(b *testing.B) {
		b.ReportAllocs()
		var bl Bloom
		for i := 0; i < b.N; i++ {
			bl = CreateBloom(rLarge)
		}
		b.StopTimer()
		var exp = common.HexToHash("c384c56ece49458a427c67b90fefe979ebf7104795be65dc398b280f24104949")
		got := crypto.SHA3Hash(bl.Bytes())
		if got != exp {
			b.Errorf("Got %x, exp %x", got, exp)
		}
	})
}
