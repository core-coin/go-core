// Copyright 2017 The go-core Authors
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

package xcc

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"testing"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/core/rawdb"
	"github.com/core-coin/go-core/core/state"
	"github.com/core-coin/go-core/crypto"
	"github.com/davecgh/go-spew/spew"
)

var dumper = spew.ConfigState{Indent: "    "}

func accountRangeTest(t *testing.T, trie *state.Trie, statedb *state.StateDB, start common.Hash, requestedNum int, expectedNum int) state.IteratorDump {
	result := statedb.IteratorDump(true, true, false, start.Bytes(), requestedNum)

	if len(result.Accounts) != expectedNum {
		t.Fatalf("expected %d results, got %d", expectedNum, len(result.Accounts))
	}
	for address := range result.Accounts {
		if address == (common.Address{}) {
			t.Fatalf("empty address returned")
		}
		if !statedb.Exist(address) {
			t.Fatalf("account not found in state %s", address.Hex())
		}
	}
	return result
}

type resultHash []common.Hash

func (h resultHash) Len() int           { return len(h) }
func (h resultHash) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h resultHash) Less(i, j int) bool { return bytes.Compare(h[i].Bytes(), h[j].Bytes()) < 0 }

func TestAccountRange(t *testing.T) { // TODO: TEST
	var (
		statedb  = state.NewDatabase(rawdb.NewMemoryDatabase())
		state, _ = state.New(common.Hash{}, statedb, nil)
		addrs    = [AccountRangeMaxResults * 2]common.Address{}
		m        = map[common.Address]bool{}
	)

	for i := range addrs {
		hash := common.HexToHash(fmt.Sprintf("%x", i))
		addr := common.BytesToAddress(crypto.SHA3Hash(hash.Bytes()).Bytes())
		addrs[i] = addr
		state.SetBalance(addrs[i], big.NewInt(1))
		if _, ok := m[addr]; ok {
			t.Fatalf("bad")
		} else {
			m[addr] = true
		}
	}
	state.Commit(true)
	root := state.IntermediateRoot(true)

	trie, err := statedb.OpenTrie(root)
	if err != nil {
		t.Fatal(err)
	}

	accountRangeTest(t, &trie, state, common.Hash{}, AccountRangeMaxResults/2, AccountRangeMaxResults/2)
	// test pagination
	firstResult := accountRangeTest(t, &trie, state, common.Hash{}, AccountRangeMaxResults, AccountRangeMaxResults)
	secondResult := accountRangeTest(t, &trie, state, common.BytesToHash(firstResult.Next), AccountRangeMaxResults, AccountRangeMaxResults)

	hList := make(resultHash, 0)
	for addr1 := range firstResult.Accounts {
		// If address is empty, then it makes no sense to compare
		// them as they might be two different accounts.
		if addr1 == (common.Address{}) {
			continue
		}
		if _, duplicate := secondResult.Accounts[addr1]; duplicate {
			t.Fatalf("pagination test failed:  results should not overlap")
		}
		hList = append(hList, crypto.SHA3Hash(addr1.Bytes()))
	}
	// Test to see if it's possible to recover from the middle of the previous
	// set and get an even split between the first and second sets.
	sort.Sort(hList)
	middleH := hList[AccountRangeMaxResults/2]
	middleResult := accountRangeTest(t, &trie, state, middleH, AccountRangeMaxResults, AccountRangeMaxResults)
	missing, infirst, insecond := 0, 0, 0
	for h := range middleResult.Accounts {
		if _, ok := firstResult.Accounts[h]; ok {
			infirst++
		} else if _, ok := secondResult.Accounts[h]; ok {
			insecond++
		} else {
			missing++
		}
	}
	if missing != 0 {
		t.Fatalf("%d hashes in the 'middle' set were neither in the first not the second set", missing)
	}
	if infirst != AccountRangeMaxResults/2 {
		t.Fatalf("Imbalance in the number of first-test results: %d != %d", infirst, AccountRangeMaxResults/2)
	}
	if insecond != AccountRangeMaxResults/2 {
		t.Fatalf("Imbalance in the number of second-test results: %d != %d", insecond, AccountRangeMaxResults/2)
	}
}

func TestEmptyAccountRange(t *testing.T) {
	var (
		statedb  = state.NewDatabase(rawdb.NewMemoryDatabase())
		state, _ = state.New(common.Hash{}, statedb, nil)
	)
	state.Commit(true)
	state.IntermediateRoot(true)
	results := state.IteratorDump(true, true, true, (common.Hash{}).Bytes(), AccountRangeMaxResults)
	if bytes.Equal(results.Next, (common.Hash{}).Bytes()) {
		t.Fatalf("Empty results should not return a second page")
	}
	if len(results.Accounts) != 0 {
		t.Fatalf("Empty state should not return addresses: %v", results.Accounts)
	}
}

func TestStorageRangeAt(t *testing.T) {
	// Create a state where account 0x010000... has a few storage entries.
	var (
		state, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		addr     = common.Address{0x01}
		keys     = []common.Hash{ // hashes of Keys of storage
			common.HexToHash("04527935532e92b92643e934da84bf65e789e245f4ca0b085b900bbdb81578da"),
			common.HexToHash("17cd8acc6c4e438664ef675e23dd274fed89954bc8e1e5ad0003f99332212603"),
			common.HexToHash("d7027bc25d82ac8d2c6419fed9692fa1a3001c98b97baec850241f6119b746c2"),
			common.HexToHash("e1fb0112e1ea3e72c8828d3024821a29a8637b94469e2f767f49cec25f24f1e3"),
		}
		storage = storageMap{
			keys[0]: {Key: &common.Hash{0x03}, Value: common.Hash{0x04}},
			keys[1]: {Key: &common.Hash{0x01}, Value: common.Hash{0x03}},
			keys[2]: {Key: &common.Hash{0x04}, Value: common.Hash{0x02}},
			keys[3]: {Key: &common.Hash{0x02}, Value: common.Hash{0x01}},
		}
	)
	for _, entry := range storage {
		state.SetState(addr, *entry.Key, entry.Value)
	}

	// Check a few combinations of limit and start/end.
	tests := []struct {
		start []byte
		limit int
		want  StorageRangeResult
	}{
		{
			start: []byte{}, limit: 0,
			want: StorageRangeResult{storageMap{}, &keys[0]},
		},
		{
			start: []byte{}, limit: 100,
			want: StorageRangeResult{storage, nil},
		},
		{
			start: []byte{}, limit: 2,
			want: StorageRangeResult{storageMap{keys[0]: storage[keys[0]], keys[1]: storage[keys[1]]}, &keys[2]},
		},
		{
			start: []byte{0x00}, limit: 4,
			want: StorageRangeResult{storage, nil},
		},
		{
			start: []byte{0x10}, limit: 2,
			want: StorageRangeResult{storageMap{keys[1]: storage[keys[1]], keys[2]: storage[keys[2]]}, &keys[3]},
		},
	}
	for _, test := range tests {
		result, err := storageRangeAt(state.StorageTrie(addr), test.start, test.limit)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(result, test.want) {
			t.Fatalf("wrong result for range 0x%x.., limit %d:\ngot %s\nwant %s",
				test.start, test.limit, dumper.Sdump(result), dumper.Sdump(&test.want))
		}
	}
}
