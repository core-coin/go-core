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
	"github.com/core-coin/ed448"
)

// Ecrecover returns the uncompressed public key that created the given signature.
func Ecrecover(hash, sig []byte) (ed448.PublicKey, error) {
	pubkey, err := SigToPub(hash, sig)
	if err != nil {
		return ed448.PublicKey{}, err
	}
	return pubkey, nil
}

// SigToPub returns the public key that created the given signature.
func SigToPub(hash, sig []byte) (ed448.PublicKey, error) {
	if len(sig) != ExtendedSignatureLength {
		return ed448.PublicKey{}, errInvalidSignature
	}
	pub := sig[SignatureLength:]
	ok := VerifySignature(pub, hash, sig)
	if !ok {
		return ed448.PublicKey{}, errInvalidSignature
	}
	return UnmarshalPubkey(pub)
}

// Sign calculates an EDDSA signature.
//
// This function is susceptible to chosen plaintext attacks that can leak
// information about the private key that is used for signing. Callers must
// be aware that the given digest cannot be chosen by an adversery. Common
// solution is to hash any input before calculating the signature.
//
// The produced signature is in the [R || S || V] format where V is 0 or 1.
func Sign(hash []byte, prv ed448.PrivateKey) ([171]byte, error) {
	if (prv == ed448.PrivateKey{}) {
		return [171]byte{}, errInvalidPrivkey
	}
	pub := ed448.Ed448DerivePublicKey(prv)

	sig := ed448.Ed448Sign(prv, pub, hash, []byte{}, false)

	var sigWithPub [171]byte
	copy(sigWithPub[:], append(sig[:], pub[:]...))

	return sigWithPub, nil
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
	return ed448.Ed448Verify(pubkey, signature[:SignatureLength], hash, []byte{}, false)
}

// DecompressPubkey parses a public key in the 33-byte compressed format.
func DecompressPubkey(pubkey []byte) (ed448.PublicKey, error) {
	return UnmarshalPubkey(pubkey)
}

// CompressPubkey encodes a public key to the 33-byte compressed format.
func CompressPubkey(pubkey ed448.PublicKey) []byte {
	return pubkey[:]
}
