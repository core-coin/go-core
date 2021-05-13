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

// +build !nacl,!js,cgo

package crypto

import (
	eddsa "github.com/core-coin/go-goldilocks"
)

// Ecrecover returns the uncompressed public key that created the given signature.
func Ecrecover(hash, sig []byte) ([]byte, error) {
	pubkey, err := SigToPub(hash, sig)
	if err != nil {
		return nil, err
	}
	return pubkey[:], nil
}

// SigToPub returns the public key that created the given signature.
func SigToPub(hash, sig []byte) (*eddsa.PublicKey, error) {
	if len(sig) != ExtendedSignatureLength {
		return nil, errInvalidSignature
	}

	pub, err := UnmarshalPubkey(sig[SignatureLength:])
	if err != nil {
		return nil, err
	}
	ok := VerifySignature(pub[:], hash, sig)
	if !ok {
		return nil, errInvalidSignature
	}
	return pub, nil
}

// Sign calculates an EDDSA signature.
//
// This function is susceptible to chosen plaintext attacks that can leak
// information about the private key that is used for signing. Callers must
// be aware that the given digest cannot be chosen by an adversery. Common
// solution is to hash any input before calculating the signature.
//
// The produced signature is in the [R || S || V] format where V is 0 or 1.
func Sign(hash []byte, prv *eddsa.PrivateKey) ([]byte, error) {
	if prv == nil || len(prv) == 0 {
		return []byte{}, errInvalidPrivkey
	}
	pub := eddsa.Ed448DerivePublicKey(*prv)

	sig := eddsa.Ed448Sign(*prv, eddsa.Ed448DerivePublicKey(*prv), hash, []byte{}, false)
	if len(sig) == 171 {
		return sig[:], nil
	}
	return append(sig[:], pub[:]...), nil
}

// VerifySignature checks that the given public key created signature over hash.
func VerifySignature(pub, hash, signature []byte) bool {
	if len(signature) != ExtendedSignatureLength {
		return false
	}
	pubkey, err := UnmarshalPubkey(pub)
	if err != nil {
		return false
	}
	return eddsa.Ed448Verify(*pubkey, signature[:SignatureLength], hash, []byte{}, false)
}

// DecompressPubkey parses a public key in the 33-byte compressed format.
func DecompressPubkey(pubkey []byte) (*eddsa.PublicKey, error) {
	return UnmarshalPubkey(pubkey)
}

// CompressPubkey encodes a public key to the 33-byte compressed format.
func CompressPubkey(pubkey *eddsa.PublicKey) []byte {
	return pubkey[:]
}
