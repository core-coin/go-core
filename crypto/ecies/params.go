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

// This file contains parameters for ECIES encryption, specifying the
// symmetric encryption and HMAC parameters.

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"hash"
)

var (
	ErrUnsupportedECDHAlgorithm   = fmt.Errorf("ecies: unsupported ECDH algorithm")
	ErrUnsupportedECIESParameters = fmt.Errorf("ecies: unsupported ECIES parameters")
	ErrInvalidKeyLen              = fmt.Errorf("ecies: invalid key size (> %d) in ECIESParams", maxKeyLen)
)

// KeyLen is limited to prevent overflow of the counter
// in concatKDF. While the theoretical limit is much higher,
// no known cipher uses keys larger than 512 bytes.
const maxKeyLen = 512

type ECIESParams struct {
	Hash      func() hash.Hash // hash function
	hashAlgo  crypto.Hash
	Cipher    func([]byte) (cipher.Block, error) // symmetric cipher
	BlockSize int                                // block size of symmetric cipher
	KeyLen    int                                // length of symmetric key
}

var ECIES_AES128_SHA256 = &ECIESParams{
	Hash:      sha256.New,
	hashAlgo:  crypto.SHA256,
	Cipher:    aes.NewCipher,
	BlockSize: aes.BlockSize,
	KeyLen:    16,
}

func ParamsFromCurve() (params *ECIESParams) {
	return ECIES_AES128_SHA256
}

func pubkeyParams(key *PublicKey) (*ECIESParams, error) {
	params := key.Params
	if params == nil {
		if params = ParamsFromCurve(); params == nil {
			return nil, ErrUnsupportedECIESParameters
		}
	}
	if params.KeyLen > maxKeyLen {
		return nil, ErrInvalidKeyLen
	}
	return params, nil
}
