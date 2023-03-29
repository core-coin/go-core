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

package trie

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/crypto/sha3"

	"github.com/core-coin/go-core/v2/xcbdb"
	"github.com/core-coin/go-core/v2/xcbdb/leveldb"
	"github.com/core-coin/go-core/v2/xcbdb/memorydb"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/rlp"
)

func init() {
	spew.Config.Indent = "    "
	spew.Config.DisableMethods = false
}

// Used for testing
func newEmpty() *Trie {
	trie, _ := New(common.Hash{}, NewDatabase(memorydb.New()))
	return trie
}

func TestEmptyTrie(t *testing.T) {
	var trie Trie
	res := trie.Hash()
	exp := emptyRoot
	if res != exp {
		t.Errorf("expected %x got %x", exp, res)
	}
}

func TestNull(t *testing.T) {
	var trie Trie
	key := make([]byte, 32)
	value := []byte("test")
	trie.Update(key, value)
	if !bytes.Equal(trie.Get(key), value) {
		t.Fatal("wrong value")
	}
}

func TestMissingRoot(t *testing.T) {
	trie, err := New(common.HexToHash("0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33"), NewDatabase(memorydb.New()))
	if trie != nil {
		t.Error("New returned non-nil trie for invalid root")
	}
	if _, ok := err.(*MissingNodeError); !ok {
		t.Errorf("New returned wrong error: %v", err)
	}
}

func TestMissingNodeDisk(t *testing.T)    { testMissingNode(t, false) }
func TestMissingNodeMemonly(t *testing.T) { testMissingNode(t, true) }

func testMissingNode(t *testing.T, memonly bool) {
	diskdb := memorydb.New()
	triedb := NewDatabase(diskdb)

	trie, _ := New(common.Hash{}, triedb)
	updateString(trie, "120000", "qwerqwerqwerqwerqwerqwerqwerqwer")
	updateString(trie, "123456", "asdfasdfasdfasdfasdfasdfasdfasdf")
	root, _ := trie.Commit(nil)
	if !memonly {
		triedb.Commit(root, true, nil)
	}

	trie, _ = New(root, triedb)
	_, err := trie.TryGet([]byte("120000"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	trie, _ = New(root, triedb)
	_, err = trie.TryGet([]byte("120099"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	trie, _ = New(root, triedb)
	_, err = trie.TryGet([]byte("123456"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	trie, _ = New(root, triedb)
	err = trie.TryUpdate([]byte("120099"), []byte("zxcvzxcvzxcvzxcvzxcvzxcvzxcvzxcv"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	trie, _ = New(root, triedb)
	err = trie.TryDelete([]byte("123456"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	hash := common.HexToHash("0x441104088924b2e5c7cad1f84c1489cef750baefc107201334de5aa75f05541a")
	if memonly {
		delete(triedb.dirties, hash)
	} else {
		diskdb.Delete(hash[:])
	}

	trie, _ = New(root, triedb)
	_, err = trie.TryGet([]byte("120000"))
	if _, ok := err.(*MissingNodeError); !ok {
		t.Errorf("Wrong error: %v", err)
	}
	trie, _ = New(root, triedb)
	_, err = trie.TryGet([]byte("120099"))
	if _, ok := err.(*MissingNodeError); !ok {
		t.Errorf("Wrong error: %v", err)
	}
	trie, _ = New(root, triedb)
	_, err = trie.TryGet([]byte("123456"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	trie, _ = New(root, triedb)
	err = trie.TryUpdate([]byte("120099"), []byte("zxcv"))
	if _, ok := err.(*MissingNodeError); !ok {
		t.Errorf("Wrong error: %v", err)
	}
	trie, _ = New(root, triedb)
	err = trie.TryDelete([]byte("123456"))
	if _, ok := err.(*MissingNodeError); !ok {
		t.Errorf("Wrong error: %v", err)
	}
}

func TestInsert(t *testing.T) {
	trie := newEmpty()

	updateString(trie, "doe", "reindeer")
	updateString(trie, "dog", "puppy")
	updateString(trie, "dogglesworth", "cat")

	exp := common.HexToHash("c33dc4124235214a96518fc8bfdef30a6c1462a08e423c29f7b08513829c551f")
	root := trie.Hash()
	if root != exp {
		t.Errorf("case 1: exp %x got %x", exp, root)
	}

	trie = newEmpty()
	updateString(trie, "A", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	exp = common.HexToHash("84ec0052caf94dc2df953a1a77ed86c20b401cff8e9b85c1c68511bc3d9259a6")
	root, err := trie.Commit(nil)
	if err != nil {
		t.Fatalf("commit error: %v", err)
	}
	if root != exp {
		t.Errorf("case 2: exp %x got %x", exp, root)
	}
}

func TestGet(t *testing.T) {
	trie := newEmpty()
	updateString(trie, "doe", "reindeer")
	updateString(trie, "dog", "puppy")
	updateString(trie, "dogglesworth", "cat")

	for i := 0; i < 2; i++ {
		res := getString(trie, "dog")
		if !bytes.Equal(res, []byte("puppy")) {
			t.Errorf("expected puppy got %x", res)
		}

		unknown := getString(trie, "unknown")
		if unknown != nil {
			t.Errorf("expected nil got %x", unknown)
		}

		if i == 1 {
			return
		}
		trie.Commit(nil)
	}
}

func TestDelete(t *testing.T) {
	trie := newEmpty()
	vals := []struct{ k, v string }{
		{"do", "verb"},
		{"core", "wookiedoo"},
		{"horse", "stallion"},
		{"shaman", "horse"},
		{"doge", "coin"},
		{"core", ""},
		{"dog", "puppy"},
		{"shaman", ""},
	}
	for _, val := range vals {
		if val.v != "" {
			updateString(trie, val.k, val.v)
		} else {
			deleteString(trie, val.k)
		}
	}

	hash := trie.Hash()
	exp := common.HexToHash("422a17a872ce2e062b0998b85bee742dfa98f429c49bb22b81495d23fd3ce841")
	if hash != exp {
		t.Errorf("expected %x got %x", exp, hash)
	}
}

func TestEmptyValues(t *testing.T) {
	trie := newEmpty()

	vals := []struct{ k, v string }{
		{"do", "verb"},
		{"core", "wookiedoo"},
		{"horse", "stallion"},
		{"shaman", "horse"},
		{"doge", "coin"},
		{"core", ""},
		{"dog", "puppy"},
		{"shaman", ""},
	}
	for _, val := range vals {
		updateString(trie, val.k, val.v)
	}

	hash := trie.Hash()
	exp := common.HexToHash("422a17a872ce2e062b0998b85bee742dfa98f429c49bb22b81495d23fd3ce841")
	if hash != exp {
		t.Errorf("expected %x got %x", exp, hash)
	}
}

func TestReplication(t *testing.T) {
	trie := newEmpty()
	vals := []struct{ k, v string }{
		{"do", "verb"},
		{"core", "wookiedoo"},
		{"horse", "stallion"},
		{"shaman", "horse"},
		{"doge", "coin"},
		{"dog", "puppy"},
		{"somethingveryoddindeedthis is", "myothernodedata"},
	}
	for _, val := range vals {
		updateString(trie, val.k, val.v)
	}
	exp, err := trie.Commit(nil)
	if err != nil {
		t.Fatalf("commit error: %v", err)
	}

	// create a new trie on top of the database and check that lookups work.
	trie2, err := New(exp, trie.db)
	if err != nil {
		t.Fatalf("can't recreate trie at %x: %v", exp, err)
	}
	for _, kv := range vals {
		if string(getString(trie2, kv.k)) != kv.v {
			t.Errorf("trie2 doesn't have %q => %q", kv.k, kv.v)
		}
	}
	hash, err := trie2.Commit(nil)
	if err != nil {
		t.Fatalf("commit error: %v", err)
	}
	if hash != exp {
		t.Errorf("root failure. expected %x got %x", exp, hash)
	}

	// perform some insertions on the new trie.
	vals2 := []struct{ k, v string }{
		{"do", "verb"},
		{"core", "wookiedoo"},
		{"horse", "stallion"},
		// {"shaman", "horse"},
		// {"doge", "coin"},
		// {"core", ""},
		// {"dog", "puppy"},
		// {"somethingveryoddindeedthis is", "myothernodedata"},
		// {"shaman", ""},
	}
	for _, val := range vals2 {
		updateString(trie2, val.k, val.v)
	}
	if hash := trie2.Hash(); hash != exp {
		t.Errorf("root failure. expected %x got %x", exp, hash)
	}
}

func TestLargeValue(t *testing.T) {
	trie := newEmpty()
	trie.Update([]byte("key1"), []byte{99, 99, 99, 99})
	trie.Update([]byte("key2"), bytes.Repeat([]byte{1}, 32))
	trie.Hash()
}

// TestRandomCases tests som cases that were found via random fuzzing
func TestRandomCases(t *testing.T) {
	var rt []randTestStep = []randTestStep{
		{op: 6, key: common.Hex2Bytes(""), value: common.Hex2Bytes("")},                                                                                                 // step 0
		{op: 6, key: common.Hex2Bytes(""), value: common.Hex2Bytes("")},                                                                                                 // step 1
		{op: 0, key: common.Hex2Bytes("d51b182b95d677e5f1c82508c0228de96b73092d78ce78b2230cd948674f66fd1483bd"), value: common.Hex2Bytes("0000000000000002")},           // step 2
		{op: 2, key: common.Hex2Bytes("c2a38512b83107d665c65235b0250002882ac2022eb00711552354832c5f1d030d0e408e"), value: common.Hex2Bytes("")},                         // step 3
		{op: 3, key: common.Hex2Bytes(""), value: common.Hex2Bytes("")},                                                                                                 // step 4
		{op: 3, key: common.Hex2Bytes(""), value: common.Hex2Bytes("")},                                                                                                 // step 5
		{op: 6, key: common.Hex2Bytes(""), value: common.Hex2Bytes("")},                                                                                                 // step 6
		{op: 3, key: common.Hex2Bytes(""), value: common.Hex2Bytes("")},                                                                                                 // step 7
		{op: 0, key: common.Hex2Bytes("c2a38512b83107d665c65235b0250002882ac2022eb00711552354832c5f1d030d0e408e"), value: common.Hex2Bytes("0000000000000008")},         // step 8
		{op: 0, key: common.Hex2Bytes("d51b182b95d677e5f1c82508c0228de96b73092d78ce78b2230cd948674f66fd1483bd"), value: common.Hex2Bytes("0000000000000009")},           // step 9
		{op: 2, key: common.Hex2Bytes("fd"), value: common.Hex2Bytes("")},                                                                                               // step 10
		{op: 6, key: common.Hex2Bytes(""), value: common.Hex2Bytes("")},                                                                                                 // step 11
		{op: 6, key: common.Hex2Bytes(""), value: common.Hex2Bytes("")},                                                                                                 // step 12
		{op: 0, key: common.Hex2Bytes("fd"), value: common.Hex2Bytes("000000000000000d")},                                                                               // step 13
		{op: 6, key: common.Hex2Bytes(""), value: common.Hex2Bytes("")},                                                                                                 // step 14
		{op: 1, key: common.Hex2Bytes("c2a38512b83107d665c65235b0250002882ac2022eb00711552354832c5f1d030d0e408e"), value: common.Hex2Bytes("")},                         // step 15
		{op: 3, key: common.Hex2Bytes(""), value: common.Hex2Bytes("")},                                                                                                 // step 16
		{op: 0, key: common.Hex2Bytes("c2a38512b83107d665c65235b0250002882ac2022eb00711552354832c5f1d030d0e408e"), value: common.Hex2Bytes("0000000000000011")},         // step 17
		{op: 5, key: common.Hex2Bytes(""), value: common.Hex2Bytes("")},                                                                                                 // step 18
		{op: 3, key: common.Hex2Bytes(""), value: common.Hex2Bytes("")},                                                                                                 // step 19
		{op: 0, key: common.Hex2Bytes("d51b182b95d677e5f1c82508c0228de96b73092d78ce78b2230cd948674f66fd1483bd"), value: common.Hex2Bytes("0000000000000014")},           // step 20
		{op: 0, key: common.Hex2Bytes("d51b182b95d677e5f1c82508c0228de96b73092d78ce78b2230cd948674f66fd1483bd"), value: common.Hex2Bytes("0000000000000015")},           // step 21
		{op: 0, key: common.Hex2Bytes("c2a38512b83107d665c65235b0250002882ac2022eb00711552354832c5f1d030d0e408e"), value: common.Hex2Bytes("0000000000000016")},         // step 22
		{op: 5, key: common.Hex2Bytes(""), value: common.Hex2Bytes("")},                                                                                                 // step 23
		{op: 1, key: common.Hex2Bytes("980c393656413a15c8da01978ed9f89feb80b502f58f2d640e3a2f5f7a99a7018f1b573befd92053ac6f78fca4a87268"), value: common.Hex2Bytes("")}, // step 24
		{op: 1, key: common.Hex2Bytes("fd"), value: common.Hex2Bytes("")},                                                                                               // step 25
	}
	runRandTest(rt)

}

// randTest performs random trie operations.
// Instances of this test are created by Generate.
type randTest []randTestStep

type randTestStep struct {
	op    int
	key   []byte // for opUpdate, opDelete, opGet
	value []byte // for opUpdate
	err   error  // for debugging
}

const (
	opUpdate = iota
	opDelete
	opGet
	opCommit
	opHash
	opReset
	opItercheckhash
	opMax // boundary value, not an actual op
)

func (randTest) Generate(r *rand.Rand, size int) reflect.Value {
	var allKeys [][]byte
	genKey := func() []byte {
		if len(allKeys) < 2 || r.Intn(100) < 10 {
			// new key
			key := make([]byte, r.Intn(50))
			r.Read(key)
			allKeys = append(allKeys, key)
			return key
		}
		// use existing key
		return allKeys[r.Intn(len(allKeys))]
	}

	var steps randTest
	for i := 0; i < size; i++ {
		step := randTestStep{op: r.Intn(opMax)}
		switch step.op {
		case opUpdate:
			step.key = genKey()
			step.value = make([]byte, 8)
			binary.BigEndian.PutUint64(step.value, uint64(i))
		case opGet, opDelete:
			step.key = genKey()
		}
		steps = append(steps, step)
	}
	return reflect.ValueOf(steps)
}

func runRandTest(rt randTest) bool {
	triedb := NewDatabase(memorydb.New())

	tr, _ := New(common.Hash{}, triedb)
	values := make(map[string]string) // tracks content of the trie

	for i, step := range rt {
		fmt.Printf("{op: %d, key: common.Hex2Bytes(\"%x\"), value: common.Hex2Bytes(\"%x\")}, // step %d\n",
			step.op, step.key, step.value, i)
		switch step.op {
		case opUpdate:
			tr.Update(step.key, step.value)
			values[string(step.key)] = string(step.value)
		case opDelete:
			tr.Delete(step.key)
			delete(values, string(step.key))
		case opGet:
			v := tr.Get(step.key)
			want := values[string(step.key)]
			if string(v) != want {
				rt[i].err = fmt.Errorf("mismatch for key 0x%x, got 0x%x want 0x%x", step.key, v, want)
			}
		case opCommit:
			_, rt[i].err = tr.Commit(nil)
		case opHash:
			tr.Hash()
		case opReset:
			hash, err := tr.Commit(nil)
			if err != nil {
				rt[i].err = err
				return false
			}
			newtr, err := New(hash, triedb)
			if err != nil {
				rt[i].err = err
				return false
			}
			tr = newtr
		case opItercheckhash:
			checktr, _ := New(common.Hash{}, triedb)
			it := NewIterator(tr.NodeIterator(nil))
			for it.Next() {
				checktr.Update(it.Key, it.Value)
			}
			if tr.Hash() != checktr.Hash() {
				rt[i].err = fmt.Errorf("hash mismatch in opItercheckhash")
			}
		}
		// Abort the test on error.
		if rt[i].err != nil {
			return false
		}
	}
	return true
}

func TestRandom(t *testing.T) {
	if err := quick.Check(runRandTest, nil); err != nil {
		if cerr, ok := err.(*quick.CheckError); ok {
			t.Fatalf("random test iteration %d failed: %s", cerr.Count, spew.Sdump(cerr.In))
		}
		t.Fatal(err)
	}
}

func BenchmarkGet(b *testing.B)      { benchGet(b, false) }
func BenchmarkGetDB(b *testing.B)    { benchGet(b, true) }
func BenchmarkUpdateBE(b *testing.B) { benchUpdate(b, binary.BigEndian) }
func BenchmarkUpdateLE(b *testing.B) { benchUpdate(b, binary.LittleEndian) }

const benchElemCount = 20000

func benchGet(b *testing.B, commit bool) {
	trie := new(Trie)
	if commit {
		_, tmpdb := tempDB()
		trie, _ = New(common.Hash{}, tmpdb)
	}
	k := make([]byte, 32)
	for i := 0; i < benchElemCount; i++ {
		binary.LittleEndian.PutUint64(k, uint64(i))
		trie.Update(k, k)
	}
	binary.LittleEndian.PutUint64(k, benchElemCount/2)
	if commit {
		trie.Commit(nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie.Get(k)
	}
	b.StopTimer()

	if commit {
		ldb := trie.db.diskdb.(*leveldb.Database)
		ldb.Close()
		os.RemoveAll(ldb.Path())
	}
}

func benchUpdate(b *testing.B, e binary.ByteOrder) *Trie {
	trie := newEmpty()
	k := make([]byte, 32)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		e.PutUint64(k, uint64(i))
		trie.Update(k, k)
	}
	return trie
}

// Benchmarks the trie hashing. Since the trie caches the result of any operation,
// we cannot use b.N as the number of hashing rouns, since all rounds apart from
// the first one will be NOOP. As such, we'll use b.N as the number of account to
// insert into the trie before measuring the hashing.
// BenchmarkHash-6   	  288680	      4561 ns/op	     682 B/op	       9 allocs/op
// BenchmarkHash-6   	  275095	      4800 ns/op	     685 B/op	       9 allocs/op
// pure hasher:
// BenchmarkHash-6   	  319362	      4230 ns/op	     675 B/op	       9 allocs/op
// BenchmarkHash-6   	  257460	      4674 ns/op	     689 B/op	       9 allocs/op
// With hashing in-between and pure hasher:
// BenchmarkHash-6   	  225417	      7150 ns/op	     982 B/op	      12 allocs/op
// BenchmarkHash-6   	  220378	      6197 ns/op	     983 B/op	      12 allocs/op
// same with old hasher
// BenchmarkHash-6   	  229758	      6437 ns/op	     981 B/op	      12 allocs/op
// BenchmarkHash-6   	  212610	      7137 ns/op	     986 B/op	      12 allocs/op
func BenchmarkHash(b *testing.B) {
	// Create a realistic account trie to hash. We're first adding and hashing N
	// entries, then adding N more.
	addresses, accounts := makeAccounts(2 * b.N)
	// Insert the accounts into the trie and hash it
	trie := newEmpty()
	i := 0
	for ; i < len(addresses)/2; i++ {
		trie.Update(crypto.SHA3(addresses[i][:]), accounts[i])
	}
	trie.Hash()
	for ; i < len(addresses); i++ {
		trie.Update(crypto.SHA3(addresses[i][:]), accounts[i])
	}
	b.ResetTimer()
	b.ReportAllocs()
	//trie.hashRoot(nil, nil)
	trie.Hash()
}

type account struct {
	Nonce   uint64
	Balance *big.Int
	Root    common.Hash
	Code    []byte
}

// Benchmarks the trie Commit following a Hash. Since the trie caches the result of any operation,
// we cannot use b.N as the number of hashing rouns, since all rounds apart from
// the first one will be NOOP. As such, we'll use b.N as the number of account to
// insert into the trie before measuring the hashing.
func BenchmarkCommitAfterHash(b *testing.B) {
	b.Run("no-onleaf", func(b *testing.B) {
		benchmarkCommitAfterHash(b, nil)
	})
	var a account
	onleaf := func(path []byte, leaf []byte, parent common.Hash) error {
		rlp.DecodeBytes(leaf, &a)
		return nil
	}
	b.Run("with-onleaf", func(b *testing.B) {
		benchmarkCommitAfterHash(b, onleaf)
	})
}

func benchmarkCommitAfterHash(b *testing.B, onleaf LeafCallback) {
	// Make the random benchmark deterministic
	addresses, accounts := makeAccounts(b.N)
	trie := newEmpty()
	for i := 0; i < len(addresses); i++ {
		trie.Update(crypto.SHA3(addresses[i][:]), accounts[i])
	}
	// Insert the accounts into the trie and hash it
	trie.Hash()
	b.ResetTimer()
	b.ReportAllocs()
	trie.Commit(onleaf)
}

func TestTinyTrie(t *testing.T) {
	// Create a realistic account trie to hash
	_, accounts := makeAccounts(5)
	trie := newEmpty()
	trie.Update(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000001337"), accounts[3])
	if exp, root := common.HexToHash("b98009397ee653ae93cbc4fecc9633b056a3f3edf787afab85e4e17788db0263"), trie.Hash(); exp != root {
		t.Errorf("1: got %x, exp %x", root, exp)
	}
	trie.Update(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000001338"), accounts[4])
	if exp, root := common.HexToHash("8c396b385e8d12e8719f98f05355bec192de8a6a26ac20da00fe017dc686ee53"), trie.Hash(); exp != root {
		t.Errorf("2: got %x, exp %x", root, exp)
	}
	trie.Update(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000001339"), accounts[4])
	if exp, root := common.HexToHash("68138b3fd0ce4d8bda74a36cd136e1172e59956769bb8a5782b4ad904bac533c"), trie.Hash(); exp != root {
		t.Errorf("3: got %x, exp %x", root, exp)
	}
	checktr, _ := New(common.Hash{}, trie.db)
	it := NewIterator(trie.NodeIterator(nil))
	for it.Next() {
		checktr.Update(it.Key, it.Value)
	}
	if troot, itroot := trie.Hash(), checktr.Hash(); troot != itroot {
		t.Fatalf("hash mismatch in opItercheckhash, trie: %x, check: %x", troot, itroot)
	}
}

func TestCommitAfterHash(t *testing.T) {
	// Create a realistic account trie to hash
	addresses, accounts := makeAccounts(1000)
	trie := newEmpty()
	for i := 0; i < len(addresses); i++ {
		trie.Update(crypto.SHA3(addresses[i][:]), accounts[i])
	}
	// Insert the accounts into the trie and hash it
	trie.Hash()
	trie.Commit(nil)
	root := trie.Hash()
	exp := common.HexToHash("281e4ca371ab2f3b9d6a0741bf9deea69a2af773ae172d8294e5492e8fa81499")
	if exp != root {
		t.Errorf("got %x, exp %x", root, exp)
	}
	root, _ = trie.Commit(nil)
	if exp != root {
		t.Errorf("got %x, exp %x", root, exp)
	}
}

func makeAccounts(size int) (addresses [][22]byte, accounts [][]byte) {
	// Make the random benchmark deterministic
	random := rand.New(rand.NewSource(0))
	// Create a realistic account trie to hash
	addresses = make([][22]byte, size)
	for i := 0; i < len(addresses); i++ {
		for j := 0; j < len(addresses[i]); j++ {
			addresses[i][j] = byte(random.Intn(256))
		}
		checksum := common.CalculateChecksum(addresses[i][1:], common.DefaultNetworkID.Bytes())
		addresses[i][1] = checksum[0]
		addresses[i][0] = common.DefaultNetworkID.Bytes()[0]
	}
	accounts = make([][]byte, len(addresses))
	for i := 0; i < len(accounts); i++ {
		var (
			nonce = uint64(random.Int63())
			root  = emptyRoot
			code  = crypto.SHA3(nil)
		)
		// The big.Rand function is not deterministic with regards to 64 vs 32 bit systems,
		// and will consume different amount of data from the rand source.
		//balance = new(big.Int).Rand(random, new(big.Int).Exp(common.Big2, common.Big256, nil))
		// Therefore, we instead just read via byte buffer
		numBytes := random.Uint32() % 33 // [0, 32] bytes
		balanceBytes := make([]byte, numBytes)
		random.Read(balanceBytes)
		balance := new(big.Int).SetBytes(balanceBytes)
		data, _ := rlp.EncodeToBytes(&account{nonce, balance, root, code})
		accounts[i] = data
	}
	return addresses, accounts
}

// spongeDb is a dummy db backend which accumulates writes in a sponge
type spongeDb struct {
	sponge  hash.Hash
	id      string
	journal []string
}

func (s *spongeDb) Has(key []byte) (bool, error)             { panic("implement me") }
func (s *spongeDb) Get(key []byte) ([]byte, error)           { return nil, errors.New("no such elem") }
func (s *spongeDb) Delete(key []byte) error                  { panic("implement me") }
func (s *spongeDb) NewBatch() xcbdb.Batch                    { return &spongeBatch{s} }
func (s *spongeDb) Stat(property string) (string, error)     { panic("implement me") }
func (s *spongeDb) Compact(start []byte, limit []byte) error { panic("implement me") }
func (s *spongeDb) Close() error                             { return nil }
func (s *spongeDb) Put(key []byte, value []byte) error {
	valbrief := value
	if len(valbrief) > 8 {
		valbrief = valbrief[:8]
	}
	s.journal = append(s.journal, fmt.Sprintf("%v: PUT([%x...], [%d bytes] %x...)\n", s.id, key[:8], len(value), valbrief))
	s.sponge.Write(key)
	s.sponge.Write(value)
	return nil
}
func (s *spongeDb) NewIterator(prefix []byte, start []byte) xcbdb.Iterator { panic("implement me") }

// spongeBatch is a dummy batch which immediately writes to the underlying spongedb
type spongeBatch struct {
	db *spongeDb
}

func (b *spongeBatch) Put(key, value []byte) error {
	b.db.Put(key, value)
	return nil
}
func (b *spongeBatch) Delete(key []byte) error             { panic("implement me") }
func (b *spongeBatch) ValueSize() int                      { return 100 }
func (b *spongeBatch) Write() error                        { return nil }
func (b *spongeBatch) Reset()                              {}
func (b *spongeBatch) Replay(w xcbdb.KeyValueWriter) error { return nil }

// TestCommitSequence tests that the trie.Commit operation writes the elements of the trie
// in the expected order, and calls the callbacks in the expected order.
// The test data was based on the 'master' code, and is basically random. It can be used
// to check whether changes to the trie modifies the write order or data in any way.
func TestCommitSequence(t *testing.T) {
	for i, tc := range []struct {
		count              int
		expWriteSeqHash    []byte
		expCallbackSeqHash []byte
	}{
		{20, common.FromHex("46423bd007c78e3483682810ef28b57548733055afdd3c3286181cf58a4e81dc"),
			common.FromHex("feb27c258a451ad67fa76f1a0da6b2a870480ceeafb45865293c0eae5337a11a")},
		{200, common.FromHex("698510e1addaa686563f5ee0fb8ee4d0f4fb3c5ed3fc79e357113671d16cf88b"),
			common.FromHex("30aba6fb3f727eb2ef942c18e264e436b775a4892a8701b7d8a2e357930b99fd")},
		{2000, common.FromHex("ac9e2300aa79f4a0c9715e240f11b92c6a3605207113a840ee66f3908415cd37"),
			common.FromHex("f9c5eee0136c9cd9b2d5f4836a2ec8c80f019ddd061653d45da4b5e812a50c1a")},
	} {
		addresses, accounts := makeAccounts(tc.count)
		// This spongeDb is used to check the sequence of disk-db-writes
		s := &spongeDb{sponge: sha3.New256()}
		db := NewDatabase(s)
		trie, _ := New(common.Hash{}, db)
		// Another sponge is used to check the callback-sequence
		callbackSponge := sha3.New256()
		// Fill the trie with elements
		for i := 0; i < tc.count; i++ {
			trie.Update(crypto.SHA3(addresses[i][:]), accounts[i])
		}
		// Flush trie -> database
		root, _ := trie.Commit(nil)
		// Flush memdb -> disk (sponge)
		db.Commit(root, false, func(c common.Hash) {
			// And spongify the callback-order
			callbackSponge.Write(c[:])
		})
		if got, exp := s.sponge.Sum(nil), tc.expWriteSeqHash; !bytes.Equal(got, exp) {
			t.Errorf("test %d, disk write sequence wrong:\ngot %x exp %x\n", i, got, exp)
		}
		if got, exp := callbackSponge.Sum(nil), tc.expCallbackSeqHash; !bytes.Equal(got, exp) {
			t.Errorf("test %d, call back sequence wrong:\ngot: %x exp %x\n", i, got, exp)
		}
	}
}

// TestCommitSequenceRandomBlobs is identical to TestCommitSequence
// but uses random blobs instead of 'accounts'
func TestCommitSequenceRandomBlobs(t *testing.T) {
	for i, tc := range []struct {
		count              int
		expWriteSeqHash    []byte
		expCallbackSeqHash []byte
	}{
		{20, common.FromHex("37f57697901e5e3e0d8479dbe8e6b69ca36e7d21a0280605fb6dddfe13a392eb"),
			common.FromHex("99ece3f0594aed23351ac60e2beb547dfda21e6188099a9009f165d4a1409e0a")},
		{200, common.FromHex("c6b50dda20cdd03885cf3fed28b01e773d2df98e5947ddb27fbfe058f5947d2c"),
			common.FromHex("16451bd32a69c0d31c16e93628302a805b99269211c8e0e1c1d5642877fe2a72")},
		{2000, common.FromHex("2e6d9e0b15ec296a2540190b5bb22cace6127e69c6dcc3763cf88625da68cf53"),
			common.FromHex("0af251355a1685c383fe5444a604efecac97ae1e93b05a526308c27ffdbd684d")},
	} {
		prng := rand.New(rand.NewSource(int64(i)))
		// This spongeDb is used to check the sequence of disk-db-writes
		s := &spongeDb{sponge: sha3.New256()}
		db := NewDatabase(s)
		trie, _ := New(common.Hash{}, db)
		// Another sponge is used to check the callback-sequence
		callbackSponge := sha3.New256()
		// Fill the trie with elements
		for i := 0; i < tc.count; i++ {
			key := make([]byte, 32)
			var val []byte
			// 50% short elements, 50% large elements
			if prng.Intn(2) == 0 {
				val = make([]byte, 1+prng.Intn(32))
			} else {
				val = make([]byte, 1+prng.Intn(4096))
			}
			prng.Read(key)
			prng.Read(val)
			trie.Update(key, val)
		}
		// Flush trie -> database
		root, _ := trie.Commit(nil)
		// Flush memdb -> disk (sponge)
		db.Commit(root, false, func(c common.Hash) {
			// And spongify the callback-order
			callbackSponge.Write(c[:])
		})
		if got, exp := s.sponge.Sum(nil), tc.expWriteSeqHash; !bytes.Equal(got, exp) {
			t.Fatalf("test %d, disk write sequence wrong:\ngot %x exp %x\n", i, got, exp)
		}
		if got, exp := callbackSponge.Sum(nil), tc.expCallbackSeqHash; !bytes.Equal(got, exp) {
			t.Fatalf("test %d, call back sequence wrong:\ngot: %x exp %x\n", i, got, exp)
		}
	}
}

func TestCommitSequenceStackTrie(t *testing.T) {
	for count := 1; count < 200; count++ {
		prng := rand.New(rand.NewSource(int64(count)))
		// This spongeDb is used to check the sequence of disk-db-writes
		s := &spongeDb{sponge: sha3.New256(), id: "a"}
		db := NewDatabase(s)
		trie, _ := New(common.Hash{}, db)
		// Another sponge is used for the stacktrie commits
		stackTrieSponge := &spongeDb{sponge: sha3.New256(), id: "b"}
		stTrie := NewStackTrie(stackTrieSponge)
		// Fill the trie with elements
		for i := 1; i < count; i++ {
			// For the stack trie, we need to do inserts in proper order
			key := make([]byte, 32)
			binary.BigEndian.PutUint64(key, uint64(i))
			var val []byte
			// 50% short elements, 50% large elements
			if prng.Intn(2) == 0 {
				val = make([]byte, 1+prng.Intn(32))
			} else {
				val = make([]byte, 1+prng.Intn(1024))
			}
			prng.Read(val)
			trie.TryUpdate(key, common.CopyBytes(val))
			stTrie.TryUpdate(key, common.CopyBytes(val))
		}
		// Flush trie -> database
		root, _ := trie.Commit(nil)
		// Flush memdb -> disk (sponge)
		db.Commit(root, false, nil)
		// And flush stacktrie -> disk
		stRoot, err := stTrie.Commit()
		if err != nil {
			t.Fatalf("Failed to commit stack trie %v", err)
		}
		if stRoot != root {
			t.Fatalf("root wrong, got %x exp %x", stRoot, root)
		}
		if got, exp := stackTrieSponge.sponge.Sum(nil), s.sponge.Sum(nil); !bytes.Equal(got, exp) {
			// Show the journal
			t.Logf("Expected:")
			for i, v := range s.journal {
				t.Logf("op %d: %v", i, v)
			}
			t.Logf("Stacktrie:")
			for i, v := range stackTrieSponge.journal {
				t.Logf("op %d: %v", i, v)
			}
			t.Fatalf("test %d, disk write sequence wrong:\ngot %x exp %x\n", count, got, exp)
		}
	}
}

// TestCommitSequenceSmallRoot tests that a trie which is essentially only a
// small (<32 byte) shortnode with an included value is properly committed to a
// database.
// This case might not matter, since in practice, all keys are 32 bytes, which means
// that even a small trie which contains a leaf will have an extension making it
// not fit into 32 bytes, rlp-encoded. However, it's still the correct thing to do.
func TestCommitSequenceSmallRoot(t *testing.T) {
	s := &spongeDb{sponge: sha3.New256(), id: "a"}
	db := NewDatabase(s)
	trie, _ := New(common.Hash{}, db)
	// Another sponge is used for the stacktrie commits
	stackTrieSponge := &spongeDb{sponge: sha3.New256(), id: "b"}
	stTrie := NewStackTrie(stackTrieSponge)
	// Add a single small-element to the trie(s)
	key := make([]byte, 5)
	key[0] = 1
	trie.TryUpdate(key, []byte{0x1})
	stTrie.TryUpdate(key, []byte{0x1})
	// Flush trie -> database
	root, _ := trie.Commit(nil)
	// Flush memdb -> disk (sponge)
	db.Commit(root, false, nil)
	// And flush stacktrie -> disk
	stRoot, err := stTrie.Commit()
	if err != nil {
		t.Fatalf("Failed to commit stack trie %v", err)
	}
	if stRoot != root {
		t.Fatalf("root wrong, got %x exp %x", stRoot, root)
	}
	fmt.Printf("root: %x\n", stRoot)
	if got, exp := stackTrieSponge.sponge.Sum(nil), s.sponge.Sum(nil); !bytes.Equal(got, exp) {
		t.Fatalf("test, disk write sequence wrong:\ngot %x exp %x\n", got, exp)
	}
}

// BenchmarkCommitAfterHashFixedSize benchmarks the Commit (after Hash) of a fixed number of updates to a trie.
// This benchmark is meant to capture the difference on efficiency of small versus large changes. Typically,
// storage tries are small (a couple of entries), whereas the full post-block account trie update is large (a couple
// of thousand entries)
func BenchmarkHashFixedSize(b *testing.B) {
	b.Run("10", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(20)
		for i := 0; i < b.N; i++ {
			benchmarkHashFixedSize(b, acc, add)
		}
	})
	b.Run("100", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(100)
		for i := 0; i < b.N; i++ {
			benchmarkHashFixedSize(b, acc, add)
		}
	})

	b.Run("1K", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(1000)
		for i := 0; i < b.N; i++ {
			benchmarkHashFixedSize(b, acc, add)
		}
	})
	b.Run("10K", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(10000)
		for i := 0; i < b.N; i++ {
			benchmarkHashFixedSize(b, acc, add)
		}
	})
	b.Run("100K", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(100000)
		for i := 0; i < b.N; i++ {
			benchmarkHashFixedSize(b, acc, add)
		}
	})
}

func benchmarkHashFixedSize(b *testing.B, addresses [][22]byte, accounts [][]byte) {
	b.ReportAllocs()
	trie := newEmpty()
	for i := 0; i < len(addresses); i++ {
		trie.Update(crypto.SHA3(addresses[i][:]), accounts[i])
	}
	// Insert the accounts into the trie and hash it
	b.StartTimer()
	trie.Hash()
	b.StopTimer()
}

func BenchmarkCommitAfterHashFixedSize(b *testing.B) {
	b.Run("10", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(20)
		for i := 0; i < b.N; i++ {
			benchmarkCommitAfterHashFixedSize(b, acc, add)
		}
	})
	b.Run("100", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(100)
		for i := 0; i < b.N; i++ {
			benchmarkCommitAfterHashFixedSize(b, acc, add)
		}
	})

	b.Run("1K", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(1000)
		for i := 0; i < b.N; i++ {
			benchmarkCommitAfterHashFixedSize(b, acc, add)
		}
	})
	b.Run("10K", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(10000)
		for i := 0; i < b.N; i++ {
			benchmarkCommitAfterHashFixedSize(b, acc, add)
		}
	})
	b.Run("100K", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(100000)
		for i := 0; i < b.N; i++ {
			benchmarkCommitAfterHashFixedSize(b, acc, add)
		}
	})
}

func benchmarkCommitAfterHashFixedSize(b *testing.B, addresses [][22]byte, accounts [][]byte) {
	b.ReportAllocs()
	trie := newEmpty()
	for i := 0; i < len(addresses); i++ {
		trie.Update(crypto.SHA3(addresses[i][:]), accounts[i])
	}
	// Insert the accounts into the trie and hash it
	trie.Hash()
	b.StartTimer()
	trie.Commit(nil)
	b.StopTimer()
}

func BenchmarkDerefRootFixedSize(b *testing.B) {
	b.Run("10", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(20)
		for i := 0; i < b.N; i++ {
			benchmarkDerefRootFixedSize(b, acc, add)
		}
	})
	b.Run("100", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(100)
		for i := 0; i < b.N; i++ {
			benchmarkDerefRootFixedSize(b, acc, add)
		}
	})

	b.Run("1K", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(1000)
		for i := 0; i < b.N; i++ {
			benchmarkDerefRootFixedSize(b, acc, add)
		}
	})
	b.Run("10K", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(10000)
		for i := 0; i < b.N; i++ {
			benchmarkDerefRootFixedSize(b, acc, add)
		}
	})
	b.Run("100K", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(100000)
		for i := 0; i < b.N; i++ {
			benchmarkDerefRootFixedSize(b, acc, add)
		}
	})
}

func benchmarkDerefRootFixedSize(b *testing.B, addresses [][22]byte, accounts [][]byte) {
	b.ReportAllocs()
	trie := newEmpty()
	for i := 0; i < len(addresses); i++ {
		trie.Update(crypto.SHA3(addresses[i][:]), accounts[i])
	}
	h := trie.Hash()
	trie.Commit(nil)
	b.StartTimer()
	trie.db.Dereference(h)
	b.StopTimer()
}

func tempDB() (string, *Database) {
	dir, err := ioutil.TempDir("", "trie-bench")
	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory: %v", err))
	}
	diskdb, err := leveldb.New(dir, 256, 0, "")
	if err != nil {
		panic(fmt.Sprintf("can't create temporary database: %v", err))
	}
	return dir, NewDatabase(diskdb)
}

func getString(trie *Trie, k string) []byte {
	return trie.Get([]byte(k))
}

func updateString(trie *Trie, k, v string) {
	trie.Update([]byte(k), []byte(v))
}

func deleteString(trie *Trie, k string) {
	trie.Delete([]byte(k))
}

func TestDecodeNode(t *testing.T) {
	t.Parallel()
	var (
		hash  = make([]byte, 20)
		elems = make([]byte, 20)
	)
	for i := 0; i < 5000000; i++ {
		rand.Read(hash)
		rand.Read(elems)
		decodeNode(hash, elems)
	}
}
