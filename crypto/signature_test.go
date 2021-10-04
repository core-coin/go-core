// Copyright 2017 by the Authors
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
	eddsa "github.com/core-coin/go-goldilocks"
	"reflect"
	"testing"

	"github.com/core-coin/go-core/common/hexutil"
)

var (
	testmsg    = common.Hex2Bytes("ce0677bb30baa8cf067c88db9811f4333d131bf8bcf12fe7065d211dce971008")
	testsig    = common.Hex2Bytes("ea535a535ff0dbfda0b2c1394bad87311789c1c6eafe6eef48fd509c2e7ba0e67c4774fab8c45abf1c7e22532bb816115bf1da8438fdb81e00e13ca01494adc201c9c35bc32cdd7c1922a0b1121f1d8ed72b37786dfd6e5583b06ad172bdb4f1d2afd41b4444abd2b5901c851fcb3d641200fadc64a37e95ad1bcbaf19625bf95826e6a8cbab42b57fc91b72da98d26bae8bda2d1fc52c508a03724aded17b8cef8253f2116307bbbf7580")
	testpubkey = common.Hex2Bytes("fadc64a37e95ad1bcbaf19625bf95826e6a8cbab42b57fc91b72da98d26bae8bda2d1fc52c508a03724aded17b8cef8253f2116307bbbf7580")
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
		pub := eddsa.Ed448DerivePublicKey(*key)
		pubkey2, err := DecompressPubkey(CompressPubkey(&pub))
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if !reflect.DeepEqual(pub, *pubkey2) {
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
