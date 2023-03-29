// Copyright (c) 2013 Kyle Isom <kyle@tyrfingr.is>
// Copyright (c) 2012 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package ecies

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"golang.org/x/crypto/sha3"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/crypto"
)

func TestKDF(t *testing.T) {
	tests := []struct {
		length int
		output []byte
	}{
		{6, decode("7dd366b373fd")},
		{32, decode("7dd366b373fdc9b1626148e981a057b4913b0e00dec183907d22872a6b1f9db7")},
		{48, decode("7dd366b373fdc9b1626148e981a057b4913b0e00dec183907d22872a6b1f9db7f489ffb85e14eb26fb5443f374a985b2")},
		{64, decode("7dd366b373fdc9b1626148e981a057b4913b0e00dec183907d22872a6b1f9db7f489ffb85e14eb26fb5443f374a985b2a13189abab09aa0b497f2bcb257e6356")},
	}

	for _, test := range tests {
		h := sha3.New256()
		k := concatKDF(h, []byte("input"), nil, test.length)
		if !bytes.Equal(k, test.output) {
			t.Fatalf("KDF: generated key %x does not match expected output %x", k, test.output)
		}
	}
}

var ErrBadSharedKeys = fmt.Errorf("ecies: shared keys don't match")

// Validate the ECDH component.
func TestSharedKey(t *testing.T) {
	prv1, err := crypto.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	prv2, err := crypto.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	pub1 := crypto.DerivePublicKey(prv1)
	pub2 := crypto.DerivePublicKey(prv2)

	sk1 := crypto.ComputeSecret(prv1, pub2)

	sk2 := crypto.ComputeSecret(prv2, pub1)

	if !bytes.Equal(sk1, sk2) {
		t.Fatal(ErrBadSharedKeys)
	}
}

func TestSharedKeyPadding(t *testing.T) {
	// sanity checks
	prv0, _ := crypto.UnmarshalPrivateKey(common.Hex2Bytes("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363991"))
	prv1, _ := crypto.UnmarshalPrivateKey(common.Hex2Bytes("fdf02153a9d5e3e0f3a958bbe9ee7e79eaf77a22703aee462354998ab0178f06566707c297df3510a3b071ccedac6b3154531aa51d10401868"))
	pub0 := decode("2f65ab658f3b0bc9fbdea48703b9c5c0dc2151c5ae8c4b77b1e5cdaee9fa20748e01960ab51ddb118d1209f73d186f0444921ad72c7c757480")
	pub1 := decode("77b1d24670fee6dd811f4f06573ce5f19844eb50cb6ce960d12bdbc8bf77be2221111cf755371d9e896e544ea2a4ebf206b775df55f5e74580")

	prv0Pub := crypto.DerivePublicKey(prv0)
	prv1Pub := crypto.DerivePublicKey(prv1)

	if !bytes.Equal(prv0Pub[:], pub0) {
		t.Errorf("mismatched prv0.X:\nhave: %x\nwant: %x\n", prv0Pub, pub0)
	}
	if !bytes.Equal(prv1Pub[:], pub1) {
		t.Errorf("mismatched prv1.X:\nhave: %x\nwant: %x\n", prv1Pub, pub1)
	}

	// test shared secret generation
	sk1 := crypto.ComputeSecret(prv0, prv1Pub)

	sk2 := crypto.ComputeSecret(prv1, prv0Pub)

	if !bytes.Equal(sk1, sk2) {
		t.Fatal(ErrBadSharedKeys.Error())
	}
}

// Benchmark the generation of P256 keys.
func BenchmarkGenerateKeyP256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := crypto.GenerateKey(rand.Reader); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark the generation of P256 shared keys.
func BenchmarkGenSharedKeyP256(b *testing.B) {
	prv, err := crypto.GenerateKey(rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := crypto.ComputeSecret(prv, crypto.DerivePublicKey(prv))
		if len(k) != 0 {
			b.Fatal("zero key len")
		}
	}
}

// Verify that an encrypted message can be successfully decrypted.
func TestEncryptDecrypt(t *testing.T) {
	prv1, err := crypto.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	prv2, err := crypto.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	message := []byte("Hello, world.")
	ct, err := Encrypt(rand.Reader, crypto.DerivePublicKey(prv2), message, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	pt, err := Decrypt(prv2, ct, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(pt, message) {
		t.Fatal("ecies: plaintext doesn't match message")
	}

	_, err = Decrypt(prv1, ct, nil, nil)
	if err == nil {
		t.Fatal("ecies: encryption should not have succeeded")
	}
}

func TestDecryptShared2(t *testing.T) {
	prv, err := crypto.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	message := []byte("Hello, world.")
	shared2 := []byte("shared data 2")
	ct, err := Encrypt(rand.Reader, crypto.DerivePublicKey(prv), message, nil, shared2)
	if err != nil {
		t.Fatal(err)
	}

	// Check that decrypting with correct shared data works.
	pt, err := Decrypt(prv, ct, nil, shared2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pt, message) {
		t.Fatal("ecies: plaintext doesn't match message")
	}

	// Decrypting without shared data or incorrect shared data fails.
	if _, err = Decrypt(prv, ct, nil, nil); err == nil {
		t.Fatal("ecies: decrypting without shared data didn't fail")
	}
	if _, err = Decrypt(prv, ct, nil, []byte("garbage")); err == nil {
		t.Fatal("ecies: decrypting with incorrect shared data didn't fail")
	}
}

func TestBox(t *testing.T) {
	prv1, _ := crypto.UnmarshalPrivateKey(common.Hex2Bytes("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363991"))
	prv2, _ := crypto.UnmarshalPrivateKey(common.Hex2Bytes("fdf02153a9d5e3e0f3a958bbe9ee7e79eaf77a22703aee462354998ab0178f06566707c297df3510a3b071ccedac6b3154531aa51d10401868"))

	message := []byte("Hello, world.")
	ct, err := Encrypt(rand.Reader, crypto.DerivePublicKey(prv2), message, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	pt, err := Decrypt(prv2, ct, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pt, message) {
		t.Fatal("ecies: plaintext doesn't match message")
	}
	if _, err = Decrypt(prv1, ct, nil, nil); err == nil {
		t.Fatal("ecies: encryption should not have succeeded")
	}
}

// Verify GenerateShared against static values - useful when
// debugging changes in underlying libs
func TestSharedKeyStatic(t *testing.T) {
	prv1, _ := crypto.UnmarshalPrivateKey(common.Hex2Bytes("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363991"))
	prv2, _ := crypto.UnmarshalPrivateKey(common.Hex2Bytes("fdf02153a9d5e3e0f3a958bbe9ee7e79eaf77a22703aee462354998ab0178f06566707c297df3510a3b071ccedac6b3154531aa51d10401868"))

	pub1 := crypto.DerivePublicKey(prv1)
	pub2 := crypto.DerivePublicKey(prv2)

	sk1 := crypto.ComputeSecret(prv1, pub2)

	sk2 := crypto.ComputeSecret(prv2, pub1)

	if !bytes.Equal(sk1, sk2) {
		t.Fatal(ErrBadSharedKeys)
	}

	sk := decode("55a3895b1c32b1d9b2160d81da1f56d2f60641fc6b997adffc53b5f473e2b62a0dc65fb4aed8fddbae912864b683a29885df4bd86c1d4760")
	if !bytes.Equal(sk1, sk) {
		t.Fatalf("shared secret mismatch: want: %x have: %x", sk, sk1)
	}
}

func decode(s string) []byte {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bytes
}
