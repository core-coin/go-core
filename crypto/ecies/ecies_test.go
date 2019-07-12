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
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"testing"

	"github.com/core-coin/go-core/crypto"
)

var dumpEnc bool

func init() {
	flDump := flag.Bool("dump", false, "write encrypted test message to file")
	flag.Parse()
	dumpEnc = *flDump
}

// Ensure the KDF generates appropriately sized keys.
func TestKDF(t *testing.T) {
	msg := []byte("Hello, world")
	h := sha256.New()

	k, err := concatKDF(h, msg, nil, 64)
	if err != nil {
		t.Fatal(err)
	}
	if len(k) != 64 {
		t.Fatalf("KDF: generated key is the wrong size (%d instead of 64\n", len(k))
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
	prv1, err := GenerateKey(rand.Reader, DefaultCurve, nil)
	if err != nil {
		t.Fatal(err)
	}
	skLen := 16

	prv2, err := GenerateKey(rand.Reader, DefaultCurve, nil)
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
	prv0 := hexKey("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e783639919d879fb971fc1857f8744ddbd489a668527eaedf4941b8fb5b1252e8431a5072695b65912e99d12c45e2d207f115a1c2d930bce2272bd1d2aadf161392088ca860e461536cb3729a5852f002d7ad6b3ffcdfa95999f3a9")
	prv1 := hexKey("fdf02153a9d5e3e0f3a958bbe9ee7e79eaf77a22703aee462354998ab0178f06566707c297df3510a3b071ccedac6b3154531aa51d10401868f3c1ffadea540d3f1277c439825929abc05f113a32e71ddb8c8e2f65e8677a052101e85b62ed46ba249d433a40262eb8ae3d9def99a13bf2fc20ac3e0077b6a0413efbed5d21e6a488b68d8b9b7f1381ff1e1b066b69ec")
	pub0 := hexPub("919d879fb971fc1857f8744ddbd489a668527eaedf4941b8fb5b1252e8431a5072695b65912e99d12c45e2d207f115a1c2d930bce2272bd1")
	pub1 := hexPub("b086d547bc4bd08d7808da56399bc1ec182c5cb26a21be11a8cb89546094fe95bacee4cae7d4f233fabb3407c4f3e56cfb22b0517dcd2211")

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

// Verify that the key generation code fails when too much key data is
// requested.
func TestTooBigSharedKey(t *testing.T) {
	prv1, err := GenerateKey(rand.Reader, DefaultCurve, nil)
	if err != nil {
		t.Fatal(err)
	}

	prv2, err := GenerateKey(rand.Reader, DefaultCurve, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = prv1.GenerateShared(&prv2.PublicKey, 32, 32)
	if err != ErrSharedKeyTooBig {
		t.Fatal("ecdh: shared key should be too large for curve")
	}

	_, err = prv2.GenerateShared(&prv1.PublicKey, 32, 32)
	if err != ErrSharedKeyTooBig {
		t.Fatal("ecdh: shared key should be too large for curve")
	}
}

// Benchmark the generation of P256 keys.
func BenchmarkGenerateKeyP256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := GenerateKey(rand.Reader, elliptic.P256(), nil); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark the generation of P256 shared keys.
func BenchmarkGenSharedKeyP256(b *testing.B) {
	prv, err := GenerateKey(rand.Reader, elliptic.P256(), nil)
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

// Benchmark the generation of S256 shared keys.
func BenchmarkGenSharedKeyS256(b *testing.B) {
	prv, err := GenerateKey(rand.Reader, crypto.S256(), nil)
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
	prv1, err := GenerateKey(rand.Reader, DefaultCurve, nil)
	if err != nil {
		t.Fatal(err)
	}

	prv2, err := GenerateKey(rand.Reader, DefaultCurve, nil)
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
	prv, err := GenerateKey(rand.Reader, DefaultCurve, nil)
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
	Curve    elliptic.Curve
	Name     string
	Expected *ECIESParams
}

var testCases = []testCase{
	{
		Curve:    elliptic.P256(),
		Name:     "P256",
		Expected: ECIES_AES128_SHA256,
	},
	{
		Curve:    elliptic.P384(),
		Name:     "P384",
		Expected: ECIES_AES256_SHA384,
	},
	{
		Curve:    elliptic.P521(),
		Name:     "P521",
		Expected: ECIES_AES256_SHA512,
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
	params := ParamsFromCurve(c.Curve)
	if params == nil && c.Expected != nil {
		t.Fatalf("%s (%s)\n", ErrInvalidParams.Error(), c.Name)
	} else if params != nil && !cmpParams(params, c.Expected) {
		t.Fatalf("ecies: parameters should be invalid (%s)\n", c.Name)
	}

	prv1, err := GenerateKey(rand.Reader, DefaultCurve, nil)
	if err != nil {
		t.Fatalf("%s (%s)\n", err.Error(), c.Name)
	}

	prv2, err := GenerateKey(rand.Reader, DefaultCurve, nil)
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

// Ensure that the basic public key validation in the decryption operation
// works.
func TestBasicKeyValidation(t *testing.T) {
	badBytes := []byte{0, 1, 5, 6, 7, 8, 9}

	prv, err := GenerateKey(rand.Reader, DefaultCurve, nil)
	if err != nil {
		t.Fatal(err)
	}

	message := []byte("Hello, world.")
	ct, err := Encrypt(rand.Reader, &prv.PublicKey, message, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, b := range badBytes {
		ct[0] = b
		_, err := prv.Decrypt(ct, nil, nil)
		if err != ErrInvalidPublicKey {
			t.Fatal("ecies: validated an invalid key")
		}
	}
}

func TestBox(t *testing.T) {
	prv1 := hexKey("4b50fa71f5c3eeb8fdc452224b2395af2fcc3d125e06c32c82e048c0559db03f")
	prv2 := hexKey("d0b043b4c5d657670778242d82d68a29d25d7d711127d17b8e299f156dad361a")
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
	prv1 := hexKey("7ebbc6a8358bc76dd73ebc557056702c8cfc34e5cfcd90eb83af0347575fd2ad")
	prv2 := hexKey("6a3d6396903245bba5837752b9e0348874e72db0c4e11e9c485a81b4ea4353b9")

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

	sk, _ := hex.DecodeString("167ccc13ac5e8a26b131c3446030c60fbfac6aa8e31149d0869f93626a4cdf62")
	if !bytes.Equal(sk1, sk) {
		t.Fatalf("shared secret mismatch: want: %x have: %x", sk, sk1)
	}
}

func hexKey(prv string) *PrivateKey {
	key, err := crypto.HexToECDSA(prv)
	if err != nil {
		panic(err)
	}
	return ImportECDSA(key)
}

func hexPub(key string) []byte {
	b, err := hex.DecodeString(key)
	if err != nil {
		panic(err)
	}

	pub, err := crypto.UnmarshalPubkey(b)
	if err != nil {
		panic(err)
	}
	return pub.X
}

