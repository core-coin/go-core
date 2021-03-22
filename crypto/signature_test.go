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
	testsig    = common.Hex2Bytes("59bcf2f083adb0df2d45adf500effdec7bf50396c41ad6bef98bbe9f05b127d4ee0557197cc489cb2446d0824bbcdedadf7e649f4d04e3c18012d49157fe68aea683779db4e579bc4b17d815a924f8f9cfe0c1e301bbee423aaba58952cebb60d43135d0a70237d691c178f36136140926009c9023964b1c45772f748867b8f3dbf02732537c21bceb3b7b0cc69d1b5c275fb3a76bfc482dfff2336d3e18da9fdb3139459b02b51a8d8f00")
	testpubkey = common.Hex2Bytes("9c9023964b1c45772f748867b8f3dbf02732537c21bceb3b7b0cc69d1b5c275fb3a76bfc482dfff2336d3e18da9fdb3139459b02b51a8d8f00")
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
