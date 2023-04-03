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
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"hash"
	"io"

	"golang.org/x/crypto/sha3"

	"github.com/core-coin/go-core/v2/crypto"
)

var (
	ErrInvalidPublicKey = fmt.Errorf("ecies: invalid public key")
	ErrInvalidMessage   = fmt.Errorf("ecies: invalid message")
)

// NIST SP 800-56 Concatenation Key Derivation Function (see section 5.8.1).
func concatKDF(hash hash.Hash, z, s1 []byte, kdLen int) []byte {
	counterBytes := make([]byte, 4)
	k := make([]byte, 0, roundup(kdLen, hash.Size()))
	for counter := uint32(1); len(k) < kdLen; counter++ {
		binary.BigEndian.PutUint32(counterBytes, counter)
		hash.Reset()
		hash.Write(counterBytes)
		hash.Write(z)
		hash.Write(s1)
		k = hash.Sum(k)
	}
	return k[:kdLen]
}

// roundup rounds size up to the next multiple of blocksize.
func roundup(size, blocksize int) int {
	return size + blocksize - (size % blocksize)
}

// deriveKeys creates the encryption and MAC keys using concatKDF.
func deriveKeys(hash hash.Hash, z, s1 []byte, keyLen int) (Ke, Km []byte) {
	K := concatKDF(hash, z, s1, 2*keyLen)
	Ke = K[:keyLen]
	Km = K[keyLen:]
	hash.Reset()
	hash.Write(Km)
	Km = hash.Sum(Km[:0])
	return Ke, Km
}

// messageTag computes the MAC of a message (called the tag) as per
// SEC 1, 3.5.
func messageTag(hash func() hash.Hash, km, msg, shared []byte) []byte {
	mac := hmac.New(hash, km)
	mac.Write(msg)
	mac.Write(shared)
	tag := mac.Sum(nil)
	return tag
}

// Generate an initialisation vector for CTR mode.
func generateIV(rand io.Reader) (iv []byte, err error) {
	iv = make([]byte, aes.BlockSize)
	_, err = io.ReadFull(rand, iv)
	return
}

// symEncrypt carries out CTR encryption using the block cipher specified in the
func symEncrypt(rand io.Reader, key, m []byte) (ct []byte, err error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	iv, err := generateIV(rand)
	if err != nil {
		return
	}
	ctr := cipher.NewCTR(c, iv)

	ct = make([]byte, len(m)+aes.BlockSize)
	copy(ct, iv)
	ctr.XORKeyStream(ct[aes.BlockSize:], m)
	return
}

// symDecrypt carries out CTR decryption using the block cipher specified in
// the parameters
func symDecrypt(key, ct []byte) (m []byte, err error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	ctr := cipher.NewCTR(c, ct[:aes.BlockSize])

	m = make([]byte, len(ct)-aes.BlockSize)
	ctr.XORKeyStream(m, ct[aes.BlockSize:])
	return
}

// Encrypt encrypts a message using ECIES as specified in SEC 1, 5.1.
//
// s1 and s2 contain shared information that is not part of the resulting
// ciphertext. s1 is fed into key derivation, s2 is fed into the MAC. If the
// shared information parameters aren't being used, they should be nil.
func Encrypt(rand io.Reader, pub *crypto.PublicKey, m, s1, s2 []byte) (ct []byte, err error) {
	R, err := crypto.GenerateKey(rand)
	if err != nil {
		return nil, err
	}

	z := crypto.ComputeSecret(R, pub)

	hash := sha3.New256()
	Ke, Km := deriveKeys(hash, z, s1, 16)

	em, err := symEncrypt(rand, Ke, m)
	if err != nil || len(em) <= aes.BlockSize {
		return nil, err
	}

	d := messageTag(sha3.New256, Km, em, s2)

	ct = make([]byte, len(R.PublicKey())+len(em)+len(d))
	copy(ct, R.PublicKey()[:])
	copy(ct[len(R.PublicKey()):], em)
	copy(ct[len(R.PublicKey())+len(em):], d)
	return ct, nil
}

// Decrypt decrypts an ECIES ciphertext.
func Decrypt(prv *crypto.PrivateKey, c, s1, s2 []byte) (m []byte, err error) {
	if len(c) == 0 {
		return nil, ErrInvalidMessage
	}

	hash := sha3.New256()

	var (
		rLen   int = len(prv.PublicKey())
		hLen   int = hash.Size()
		mStart int = rLen
		mEnd   int = len(c) - hLen
	)

	R, err := crypto.UnmarshalPubKey(c[:rLen])
	if err != nil {
		return nil, err
	}
	if len(R) == 0 {
		return nil, ErrInvalidPublicKey
	}

	z := crypto.ComputeSecret(prv, R)

	Ke, Km := deriveKeys(hash, z, s1, 16)

	d := messageTag(sha3.New256, Km, c[mStart:mEnd], s2)
	if subtle.ConstantTimeCompare(c[mEnd:], d) != 1 {
		return nil, ErrInvalidMessage
	}

	return symDecrypt(Ke, c[mStart:mEnd])
}
