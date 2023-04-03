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

package crypto

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/core-coin/go-core/v2/common"
)

var testAddrHex = "cb82a5fd22b9bee8b8ab877c86e0a2c21765e1d5bfc5"
var testPrivHex = "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0e"

// These tests are sanity checks.
// They should ensure that we don't e.g. use Sha3-224 instead of Sha3-256
// and that the sha3 library uses keccak-f permutation.
func TestSHA3Hash(t *testing.T) {
	msg := []byte("abc")
	exp, _ := hex.DecodeString("3a985da74fe225b2045c172d6bd390bd855f086e3e9d525b46bfe24511431532")
	checkhash(t, "Sha3-256-array", func(in []byte) []byte { h := SHA3Hash(in); return h[:] }, msg, exp)
}

func TestToEDDSAErrors(t *testing.T) {
	if _, err := UnmarshalPrivateKeyHex("0000000000000000000000000000000000000000000000000000000000000000"); err == nil {
		t.Fatal("UnmarshalPrivateKeyHex should've returned error")
	}
	if _, err := UnmarshalPrivateKeyHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"); err == nil {
		t.Fatal("UnmarshalPrivateKeyHex should've returned error")
	}
}

func BenchmarkSha3(b *testing.B) {
	a := []byte("hello world")
	for i := 0; i < b.N; i++ {
		SHA3(a)
	}
}

func TestUnmarshalPubkey(t *testing.T) {
	key, err := UnmarshalPubKey(nil)
	if err != errInvalidPubkey || key != nil {
		t.Fatalf("expected error, got %v, %v", err, key)
	}
	key, err = UnmarshalPubKey([]byte{1, 2, 3})
	if err != errInvalidPubkey || key != nil {
		t.Fatalf("expected error, got %v, %v", err, key)
	}

	var (
		enc, _ = hex.DecodeString("aaee47e4f7afb3a0dfd813320278e8ce0c9b1f94bded9a7e0ad9f9250c3360e16cbb3d90484ccc59805be6398b6ca774959d37a8a4cdc81faf")
	)
	key, err = UnmarshalPubKey(enc)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSign(t *testing.T) {
	key, _ := UnmarshalPrivateKeyHex(testPrivHex)
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
	pubKey, _ := UnmarshalPubKey(recoveredPub)
	recoveredAddr := PubkeyToAddress(pubKey)
	if addr != recoveredAddr {
		t.Errorf("Address mismatch: want: %x have: %x", addr, recoveredAddr)
	}

	// should be equal to SigToPub
	recoveredPub2, err := SigToPub(msg, sig)
	if err != nil {
		t.Errorf("ECRecover error: %s", err)
	}
	recoveredAddr2 := PubkeyToAddress(recoveredPub2)
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
	key, _ := UnmarshalPrivateKeyHex(testPrivHex)
	addr, err := common.HexToAddress(testAddrHex)
	if err != nil {
		t.Error(err)
	}
	pub := DerivePublicKey(key)
	genAddr := PubkeyToAddress(pub)
	// sanity check before using addr to create contract address
	checkAddr(t, genAddr, addr)
	caddr0 := CreateAddress(addr, 0)
	caddr1 := CreateAddress(addr, 1)
	caddr2 := CreateAddress(addr, 2)

	addr0, err := common.HexToAddress("cb57718e2b338b99d2587a6dd6c01fc2b97a4296449f")
	if err != nil {
		t.Error(err)
	}
	addr1, err := common.HexToAddress("cb812bae2e00797890802e8aa6c162aac5cac4d8990c")
	if err != nil {
		t.Error(err)
	}
	addr2, err := common.HexToAddress("cb98c562c98ac1be66b1302be0cac7f8da9694900b09")
	if err != nil {
		t.Error(err)
	}
	checkAddr(t, addr0, caddr0)
	checkAddr(t, addr1, caddr1)
	checkAddr(t, addr2, caddr2)
}

func TestLoadEDDSA(t *testing.T) {
	tests := []struct {
		input string
		err   string
	}{
		// good
		{input: "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0e"},
		{input: "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0e\n"},
		{input: "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0e\n\r"},
		{input: "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0e\r\n"},
		{input: "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0e\n\n"},
		{input: "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0e\n\r"},
		// bad
		{
			input: "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0",
			err:   "key file too short, want 57 hex characters",
		},
		{
			input: "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0\n",
			err:   "key file too short, want 57 hex characters",
		},
		{
			input: "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0X",
			err:   "invalid hex character 'X' in private key",
		},
		{
			input: "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0eX",
			err:   "invalid character 'X' at end of key file",
		},
		{
			input: "69bb68c3a00a0cd9cbf2cab316476228c758329bbfe0b1759e8634694a9497afea05bcbf24e2aa0627eac4240484bb71de646a9296872a3c0e\n\n\n",
			err:   "key file too long, want 57 hex characters",
		},
	}

	for _, test := range tests {
		f, err := ioutil.TempFile("", "loaded448_test.*.txt")
		if err != nil {
			t.Fatal(err)
		}
		filename := f.Name()
		f.WriteString(test.input)
		f.Close()

		_, err = LoadEDDSA(filename)
		switch {
		case err != nil && test.err == "":
			t.Fatalf("unexpected error for input %q:\n  %v", test.input, err)
		case err != nil && err.Error() != test.err:
			t.Fatalf("wrong error for input %q:\n  %v", test.input, err)
		case err == nil && test.err != "":
			t.Fatalf("LoadEDDSA did not return error for input %q", test.input)
		}
	}
}

func TestSaveEDDSA(t *testing.T) {
	f, err := ioutil.TempFile("", "saveed448_test.*.txt")
	if err != nil {
		t.Fatal(err)
	}
	file := f.Name()
	f.Close()
	defer os.Remove(file)

	key, _ := UnmarshalPrivateKeyHex(testPrivHex)
	if err := SaveEDDSA(file, key); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadEDDSA(file)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(key, loaded) {
		t.Fatal("loaded key not equal to saved key")
	}
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
