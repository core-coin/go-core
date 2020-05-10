// Copyright 2017 The go-core Authors
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

// +build !nacl,!js,cgo

package crypto

import (
	eddsa "github.com/core-coin/eddsa"
	"crypto/elliptic"
	"github.com/core-coin/go-core/crypto/secp256k1"
)

// Ecrecover returns the uncompressed public key that created the given signature.
func Ecrecover(hash, sig []byte) ([]byte, error) {
	pubkey, err := SigToPub(hash, sig)
	if err != nil {
		return nil, err
	}
	return pubkey.X, nil
}

// SigToPub returns the public key that created the given signature.
func SigToPub(hash, sig []byte) (*eddsa.PublicKey, error) {
	_ = hash

	if len(sig) != SignatureLength {
		return nil, errInvalidSignature
	}

	pubkey, err := eddsa.Ed448().SigToPub(sig)
	if err != nil {
		return nil, err
	}
	return eddsa.Ed448().UnmarshalPub(pubkey)
}

// Sign calculates an EDDSA signature.
//
// This function is susceptible to chosen plaintext attacks that can leak
// information about the private key that is used for signing. Callers must
// be aware that the given digest cannot be chosen by an adversery. Common
// solution is to hash any input before calculating the signature.
//
// The produced signature is in the [R || S || V] format where V is 0 or 1.
func Sign(hash []byte, prv *eddsa.PrivateKey) (sig []byte, err error) {
	if prv == nil || len(prv.D) == 0 {
		return []byte{}, errInvalidPrivkey
	}
	return eddsa.Ed448().Sign(prv, hash)
}

// VerifySignature checks that the given public key created signature over hash.
// The public key should be in compressed (33 bytes) or uncompressed (112 + 56 bytes) format.
// The signature should have the 64 byte [R || S] format.
func VerifySignature(pub, hash, signature []byte) bool {
	pubkey, err := eddsa.Ed448().UnmarshalPub(pub)
	if err != nil {
		return false
	}
	return eddsa.Ed448().Verify(pubkey, hash, signature)
}

// DecompressPubkey parses a public key in the 33-byte compressed format.
func DecompressPubkey(pubkey []byte) (*eddsa.PublicKey, error) {
	return eddsa.Ed448().UnmarshalPub(pubkey)
}

// CompressPubkey encodes a public key to the 33-byte compressed format.
func CompressPubkey(pubkey *eddsa.PublicKey) []byte {
	return pubkey.X
}

// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	return secp256k1.S256()
}
