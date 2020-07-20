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

package crypto

import (
	"bytes"
	"encoding/hex"
	"github.com/core-coin/eddsa"
	"io/ioutil"
	"os"
	"testing"

	"github.com/core-coin/go-core/common"
)

var testAddrHex = "cb26b2a6f9a2c6925d2407314b99518f8109156d0b09"
var testPrivHex = "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0ec01df931bb7405b5db26f6b98e136fa736df081c42698e425b493891f6195cc71b5cc76fac19461468d22d1359f0ad87e22dbdd5a202a32683dcaabd9c5cf3034fe44c155c1b06c59f7d6fc14b7e6172c18c6b0076d9a4"

// These tests are sanity checks.
// They should ensure that we don't e.g. use Sha3-224 instead of Sha3-256
// and that the sha3 library uses keccak-f permutation.
func TestSHA3Hash(t *testing.T) {
	msg := []byte("abc")
	exp, _ := hex.DecodeString("3a985da74fe225b2045c172d6bd390bd855f086e3e9d525b46bfe24511431532")
	checkhash(t, "Sha3-256-array", func(in []byte) []byte { h := SHA3Hash(in); return h[:] }, msg, exp)
}

func TestToEDDSAErrors(t *testing.T) {
	if _, err := HexToEDDSA("0000000000000000000000000000000000000000000000000000000000000000"); err == nil {
		t.Fatal("HexToEDDSA should've returned error")
	}
	if _, err := HexToEDDSA("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"); err == nil {
		t.Fatal("HexToEDDSA should've returned error")
	}
}

func BenchmarkSha3(b *testing.B) { //TODO: TEST
	a := []byte("hello world")
	for i := 0; i < b.N; i++ {
		SHA3(a)
	}
}

func TestUnmarshalPubkey(t *testing.T) {
	key, err := UnmarshalPubkey(nil)
	if err != errInvalidPubkey || key != nil {
		t.Fatalf("expected error, got %v, %v", err, key)
	}
	key, err = UnmarshalPubkey([]byte{1, 2, 3})
	if err != errInvalidPubkey || key != nil {
		t.Fatalf("expected error, got %v, %v", err, key)
	}

	var (
		enc, _ = hex.DecodeString("ee47e4f7afb3a0dfd813320278e8ce0c9b1f94bded9a7e0ad9f9250c3360e16cbb3d90484ccc59805be6398b6ca774959d37a8a4cdc81faf")
	)
	key, err = UnmarshalPubkey(enc)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSign(t *testing.T) {
	key, _ := HexToEDDSA(testPrivHex)
	addr, err := common.HexToAddress(testAddrHex)
	if err != nil {
		t.Error(err)
	}
	msg := SHA3([]byte("foo"))
	sig, err := Sign(msg, key)
	if err != nil {
		t.Errorf("Sign error: %s", err)
	}
	recoveredPub, err := Ecrecover(msg, sig)
	if err != nil {
		t.Errorf("ECRecover error: %s", err)
	}
	pubKey, _ := UnmarshalPubkey(recoveredPub)
	recoveredAddr := PubkeyToAddress(*pubKey)
	if addr != recoveredAddr {
		t.Errorf("Address mismatch: want: %x have: %x", addr, recoveredAddr)
	}

	// should be equal to SigToPub
	recoveredPub2, err := SigToPub(msg, sig)
	if err != nil {
		t.Errorf("ECRecover error: %s", err)
	}
	recoveredAddr2 := PubkeyToAddress(*recoveredPub2)
	if addr != recoveredAddr2 {
		t.Errorf("Address mismatch: want: %x have: %x", addr, recoveredAddr2)
	}
}

func TestInvalidSign(t *testing.T) {
	if _, err := Sign(make([]byte, 1), nil); err == nil {
		t.Errorf("expected sign with hash 1 byte to error")
	}
	if _, err := Sign(make([]byte, 33), nil); err == nil {
		t.Errorf("expected sign with hash 33 byte to error")
	}
}

func TestNewContractAddress(t *testing.T) {
	key, _ := HexToEDDSA(testPrivHex)
	addr, err := common.HexToAddress(testAddrHex)
	if err != nil {
		t.Error(err)
	}
	genAddr := PubkeyToAddress(key.PublicKey)
	// sanity check before using addr to create contract address
	checkAddr(t, genAddr, addr)
	caddr0 := CreateAddress(addr, 0)
	caddr1 := CreateAddress(addr, 1)
	caddr2 := CreateAddress(addr, 2)

	addr0, err := common.HexToAddress("cb045175867858103f849c7ab4b8a5f9284295ed33fb")
	if err != nil {
		t.Error(err)
	}
	addr1, err := common.HexToAddress("cb83ef5c322194131ae16cdcb7dae75f0f51bded4f55")
	if err != nil {
		t.Error(err)
	}
	addr2, err := common.HexToAddress("cb85ed1b013357fc96403540a4eddc322c0df29f3d63")
	if err != nil {
		t.Error(err)
	}
	checkAddr(t, addr0, caddr0)
	checkAddr(t, addr1, caddr1)
	checkAddr(t, addr2, caddr2)
}

func TestLoadEDDSAFile(t *testing.T) {
	keyBytes := common.FromHex(testPrivHex)
	fileName0 := "test_key0"
	fileName1 := "test_key1"
	addr, _ := common.HexToAddress(testAddrHex)
	checkKey := func(k *eddsa.PrivateKey) {
		checkAddr(t, PubkeyToAddress(k.PublicKey), addr)
		loadedKeyBytes := FromEDDSA(k)
		if !bytes.Equal(loadedKeyBytes, keyBytes) {
			t.Fatalf("private key mismatch: want: %x have: %x", keyBytes, loadedKeyBytes)
		}
	}

	ioutil.WriteFile(fileName0, []byte(testPrivHex), 0600)
	defer os.Remove(fileName0)

	key0, err := LoadEDDSA(fileName0)
	if err != nil {
		t.Fatal(err)
	}
	checkKey(key0)

	// again, this time with SaveEDDSA instead of manual save:
	err = SaveEDDSA(fileName1, key0)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fileName1)

	key1, err := LoadEDDSA(fileName1)
	if err != nil {
		t.Fatal(err)
	}
	checkKey(key1)
}

func checkhash(t *testing.T, name string, f func([]byte) []byte, msg, exp []byte) {
	sum := f(msg)
	if !bytes.Equal(exp, sum) {
		t.Fatalf("hash %s mismatch: want: %x have: %x", name, exp, sum)
	}
}

func checkAddr(t *testing.T, addr0, addr1 common.Address) {
	if addr0 != addr1 {
		t.Fatalf("address mismatch: want: %x have: %x", addr0, addr1)
	}
}

// test to help Python team with integration of libsecp256k1
// skip but keep it after they are done
func TestPythonIntegration(t *testing.T) {
	kh := "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0ec01df931bb7405b5db26f6b98e136fa736df081c42698e425b493891f6195cc71b5cc76fac19461468d22d1359f0ad87e22dbdd5a202a32683dcaabd9c5cf3034fe44c155c1b06c59f7d6fc14b7e6172c18c6b0076d9a4"
	k0, _ := HexToEDDSA(kh)

	msg0 := SHA3([]byte("foo"))
	sig0, _ := Sign(msg0, k0)

	msg1 := common.FromHex("00000000000000000000000000000000")
	sig1, _ := Sign(msg0, k0)

	t.Logf("msg: %x, privkey: %s sig: %x\n", msg0, kh, sig0)
	t.Logf("msg: %x, privkey: %s sig: %x\n", msg1, kh, sig1)
}
