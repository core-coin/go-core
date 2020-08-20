// Copyright 2020 The CORE FOUNDATION, nadacia
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
	"crypto/rand"
	"github.com/core-coin/go-core/common"
	"reflect"
	"testing"

	"github.com/core-coin/go-core/common/hexutil"
)

var (
	testmsg     = common.Hex2Bytes("ce0677bb30baa8cf067c88db9811f4333d131bf8bcf12fe7065d211dce971008")
	testsig     = common.Hex2Bytes("d89625e1348f5444e4aa07c65b6810d4984d2d75d28de35e4bf79cf2bf6068d6922e2628a7111ff91d6c676f0be5c9d464ea935045ad8cb0f0e29c15cd4473a474e3c950214d6303633c4abee7d4d1582530d1bc9f75d37754f6ed82bfafb4a70b3283ca6dc996815b0042eedd7c873671583e2e906397e2833cb0c4321941cd91592ca47cf10ccccc5a7df7442d750616d943f01ef6841c2e248148264c66bdc0649ddd8d62e3b6")
	testpubkey  = common.Hex2Bytes("71583e2e906397e2833cb0c4321941cd91592ca47cf10ccccc5a7df7442d750616d943f01ef6841c2e248148264c66bdc0649ddd8d62e3b6")
)

func TestEcrecover(t *testing.T) {
	pubkey, err := Ecrecover(testmsg, testsig)
	if err != nil {
		t.Fatalf("recover error: %s", err)
	}
	if !bytes.Equal(pubkey, testpubkey) {
		t.Errorf("pubkey mismatch: want: %x have: %x", testpubkey, pubkey)
	}
}

func TestVerifySignature(t *testing.T) {
	sig := testsig
	if !VerifySignature(testpubkey, testmsg, sig) {
		t.Errorf("can't verify signature with uncompressed key")
	}

	if VerifySignature(nil, testmsg, sig) {
		t.Errorf("signature valid with no key")
	}
	if VerifySignature(testpubkey, nil, testsig) {
		t.Errorf("signature valid with no message")
	}
	if VerifySignature(testpubkey, testmsg, nil) {
		t.Errorf("nil signature valid")
	}
	if VerifySignature(testpubkey, testmsg, append(common.CopyBytes(testsig), 1, 2, 3)) {
		t.Errorf("signature valid with extra bytes at the end")
	}
	if VerifySignature(testpubkey, testmsg, testsig[:len(testsig)-2]) {
		t.Errorf("signature valid even though it's incomplete")
	}
	wrongkey := common.CopyBytes(testpubkey)
	wrongkey[10]++
	if VerifySignature(wrongkey, testmsg, testsig) {
		t.Errorf("signature valid with with wrong public key")
	}
}

// This test checks that VerifySignature rejects malleable signatures with s > N/2.
func TestVerifySignatureMalleable(t *testing.T) {
	sig := hexutil.MustDecode("0x638a54215d80a6713c8d523a6adc4e6e73652d859103a36b700851cb0e61b66b8ebfc1a610c57d732ec6e0a8f06a9a7a28df5051ece514702ff9cdff0b11f454")
	key := hexutil.MustDecode("0x03ca634cae0d49acb401d8a4c6b6fe8c55b70d115bf400769cc1400f3258cd3138")
	msg := hexutil.MustDecode("0xd301ce462d3e639518f482c7f03821fec1e602018630ce621e1e7851c12343a6")
	if VerifySignature(key, msg, sig) {
		t.Error("VerifySignature returned true for malleable signature")
	}
}

func TestPubkeyRandom(t *testing.T) {
	const runs = 200

	for i := 0; i < runs; i++ {
		key, err := GenerateKey(rand.Reader)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		pubkey2, err := DecompressPubkey(CompressPubkey(&key.PublicKey))
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if !reflect.DeepEqual(key.PublicKey, *pubkey2) {
			t.Fatalf("iteration %d: keys not equal", i)
		}
	}
}

func BenchmarkEcrecoverSignature(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := Ecrecover(testmsg, testsig); err != nil {
			b.Fatal("ecrecover error", err)
		}
	}
}

func BenchmarkVerifySignature(b *testing.B) {
	sig := testsig[:len(testsig)-1] // remove recovery id
	for i := 0; i < b.N; i++ {
		if !VerifySignature(testpubkey, testmsg, sig) {
			b.Fatal("verify error")
		}
	}
}
