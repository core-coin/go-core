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
	"golang.org/x/crypto/sha3"
	"testing"

	"github.com/core-coin/go-core/crypto"
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

// cmpParams compares a set of ECIES parameters. We assume, as per the
// docs, that AES is the only supported symmetric encryption algorithm.
func cmpParams(p1, p2 *ECIESParams) bool {
	return p1.hashAlgo == p2.hashAlgo &&
		p1.KeyLen == p2.KeyLen &&
		p1.BlockSize == p2.BlockSize
}

// Validate the ECDH component.
func TestSharedKey(t *testing.T) {
	prv1, err := GenerateKey(rand.Reader, nil)
	if err != nil {
		t.Fatal(err)
	}
	skLen := 16

	prv2, err := GenerateKey(rand.Reader, nil)
	if err != nil {
		t.Fatal(err)
	}

	sk1, err := prv1.GenerateShared(&prv2.PublicKey, skLen, skLen)
	if err != nil {
		t.Fatal(err)
	}

	sk2, err := prv2.GenerateShared(&prv1.PublicKey, skLen, skLen)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(sk1, sk2) {
		t.Fatal(ErrBadSharedKeys)
	}
}

func TestSharedKeyPadding(t *testing.T) {
	// sanity checks
	prv0 := hexKey("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363991")
	prv1 := hexKey("fdf02153a9d5e3e0f3a958bbe9ee7e79eaf77a22703aee462354998ab0178f06566707c297df3510a3b071ccedac6b3154531aa51d10401868")
	pub0 := decode("d0ac61cb8a3712468f5264def724a8e569d0973a88da1a14c3c3fd6db8cb79c947f36ac1b5a975790ffe46a019ed7c28500aa2f5a8b9562200")
	pub1 := decode("77b1d24670fee6dd811f4f06573ce5f19844eb50cb6ce960d12bdbc8bf77be2221111cf755371d9e896e544ea2a4ebf206b775df55f5e74580")
	if !bytes.Equal(prv0.PublicKey.X, pub0) {
		t.Errorf("mismatched prv0.X:\nhave: %x\nwant: %x\n", prv0.PublicKey.X, pub0)
	}
	if !bytes.Equal(prv1.PublicKey.X, pub1) {
		t.Errorf("mismatched prv1.X:\nhave: %x\nwant: %x\n", prv1.PublicKey.X, pub1)
	}

	// test shared secret generation
	sk1, err := prv0.GenerateShared(&prv1.PublicKey, 16, 16)
	if err != nil {
		t.Log(err.Error())
	}

	sk2, err := prv1.GenerateShared(&prv0.PublicKey, 16, 16)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !bytes.Equal(sk1, sk2) {
		t.Fatal(ErrBadSharedKeys.Error())
	}
}

// Benchmark the generation of S256 shared keys.
func BenchmarkGenSharedKeyS256(b *testing.B) {
	prv, err := GenerateKey(rand.Reader, nil)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := prv.GenerateShared(&prv.PublicKey, 16, 16)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Verify that an encrypted message can be successfully decrypted.
func TestEncryptDecrypt(t *testing.T) {
	prv1, err := GenerateKey(rand.Reader, nil)
	if err != nil {
		t.Fatal(err)
	}

	prv2, err := GenerateKey(rand.Reader, nil)
	if err != nil {
		t.Fatal(err)
	}

	message := []byte("Hello, world.")
	ct, err := Encrypt(rand.Reader, &prv2.PublicKey, message, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	pt, err := prv2.Decrypt(ct, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(pt, message) {
		t.Fatal("ecies: plaintext doesn't match message")
	}

	_, err = prv1.Decrypt(ct, nil, nil)
	if err == nil {
		t.Fatal("ecies: encryption should not have succeeded")
	}
}

func TestDecryptShared2(t *testing.T) {
	prv, err := GenerateKey(rand.Reader, nil)
	if err != nil {
		t.Fatal(err)
	}
	message := []byte("Hello, world.")
	shared2 := []byte("shared data 2")
	ct, err := Encrypt(rand.Reader, &prv.PublicKey, message, nil, shared2)
	if err != nil {
		t.Fatal(err)
	}

	// Check that decrypting with correct shared data works.
	pt, err := prv.Decrypt(ct, nil, shared2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pt, message) {
		t.Fatal("ecies: plaintext doesn't match message")
	}

	// Decrypting without shared data or incorrect shared data fails.
	if _, err = prv.Decrypt(ct, nil, nil); err == nil {
		t.Fatal("ecies: decrypting without shared data didn't fail")
	}
	if _, err = prv.Decrypt(ct, nil, []byte("garbage")); err == nil {
		t.Fatal("ecies: decrypting with incorrect shared data didn't fail")
	}
}

type testCase struct {
	Name     string
	Expected *ECIESParams
}

var testCases = []testCase{
	{
		Name:     "S256",
		Expected: ECIES_AES128_SHA256,
	},
}

// Test parameter selection for each curve, and that P224 fails automatic
// parameter selection (see README for a discussion of P224). Ensures that
// selecting a set of parameters automatically for the given curve works.
func TestParamSelection(t *testing.T) {
	for _, c := range testCases {
		testParamSelection(t, c)
	}
}

func testParamSelection(t *testing.T, c testCase) {
	params := ParamsFromCurve()
	if params == nil {
		t.Fatal("ParamsFromCurve returned nil")
	} else if params != nil && !cmpParams(params, c.Expected) {
		t.Fatalf("ecies: parameters should be invalid (%s)\n", c.Name)
	}

	prv1, err := GenerateKey(rand.Reader, nil)
	if err != nil {
		t.Fatalf("%s (%s)\n", err.Error(), c.Name)
	}

	prv2, err := GenerateKey(rand.Reader, nil)
	if err != nil {
		t.Fatalf("%s (%s)\n", err.Error(), c.Name)
	}

	message := []byte("Hello, world.")
	ct, err := Encrypt(rand.Reader, &prv2.PublicKey, message, nil, nil)
	if err != nil {
		t.Fatalf("%s (%s)\n", err.Error(), c.Name)
	}

	pt, err := prv2.Decrypt(ct, nil, nil)
	if err != nil {
		t.Fatalf("%s (%s)\n", err.Error(), c.Name)
	}

	if !bytes.Equal(pt, message) {
		t.Fatalf("ecies: plaintext doesn't match message (%s)\n", c.Name)
	}

	_, err = prv1.Decrypt(ct, nil, nil)
	if err == nil {
		t.Fatalf("ecies: encryption should not have succeeded (%s)\n", c.Name)
	}

}

func TestBox(t *testing.T) {
	prv1 := hexKey("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363991")
	prv2 := hexKey("fdf02153a9d5e3e0f3a958bbe9ee7e79eaf77a22703aee462354998ab0178f06566707c297df3510a3b071ccedac6b3154531aa51d10401868")
	pub2 := &prv2.PublicKey

	message := []byte("Hello, world.")
	ct, err := Encrypt(rand.Reader, pub2, message, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	pt, err := prv2.Decrypt(ct, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pt, message) {
		t.Fatal("ecies: plaintext doesn't match message")
	}
	if _, err = prv1.Decrypt(ct, nil, nil); err == nil {
		t.Fatal("ecies: encryption should not have succeeded")
	}
}

// Verify GenerateShared against static values - useful when
// debugging changes in underlying libs
func TestSharedKeyStatic(t *testing.T) {
	prv1 := hexKey("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363991")
	prv2 := hexKey("fdf02153a9d5e3e0f3a958bbe9ee7e79eaf77a22703aee462354998ab0178f06566707c297df3510a3b071ccedac6b3154531aa51d10401868")

	skLen := 16

	sk1, err := prv1.GenerateShared(&prv2.PublicKey, skLen, skLen)
	if err != nil {
		t.Fatal(err)
	}

	sk2, err := prv2.GenerateShared(&prv1.PublicKey, skLen, skLen)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(sk1, sk2) {
		t.Fatal(ErrBadSharedKeys)
	}

	sk := decode("3db819aa0ebae1996860cf44941e57b4df43b50908e46717e3a23b1eee114ce0a4535683eb44b9e533075d0a4e3333d8c6752736a6faa327")
	if !bytes.Equal(sk1, sk) {
		t.Fatalf("shared secret mismatch: want: %x have: %x", sk, sk1)
	}
}

func hexKey(prv string) *PrivateKey {
	key, err := crypto.HexToEDDSA(prv)
	if err != nil {
		panic(err)
	}
	return ImportEDDSA(key)
}

func decode(s string) []byte {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bytes
}
