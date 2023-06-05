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

//go:build !nacl && !js && cgo && !android && !ios && !mobile

package crypto

import (
	"bytes"

	"github.com/core-coin/go-goldilocks"
)

// SigToPub returns the public key that created the given signature.
func SigToPub(hash, sig []byte) (*PublicKey, error) {
	if len(sig) != ExtendedSignatureLength {
		return nil, errInvalidSignature
	}
	pub := sig[SignatureLength:]
	ok := VerifySignature(pub, hash, sig)
	if !ok {
		return nil, errInvalidSignature
	}
	return UnmarshalPubKey(pub)
}

// Sign calculates an EDDSA signature.
//
// This function is susceptible to chosen plaintext attacks that can leak
// information about the private key that is used for signing. Callers must
// be aware that the given digest cannot be chosen by an adversery. Common
// solution is to hash any input before calculating the signature.
func Sign(hash []byte, prv *PrivateKey) ([]byte, error) {
	if prv == nil || len(prv.PrivateKey()) == 0 || bytes.Equal(prv.PrivateKey(), []byte{}) || bytes.Equal(prv.PrivateKey(), make([]byte, 57)) {
		return []byte{}, errInvalidPrivkey
	}
	sig := goldilocks.Ed448Sign(goldilocks.BytesToPrivateKey(prv.PrivateKey()), goldilocks.BytesToPublicKey(prv.PublicKey()[:]), hash, []byte{}, false)
	if len(sig) == ExtendedSignatureLength {
		return sig[:], nil
	}
	return append(sig[:], prv.PublicKey()[:]...), nil
}

// VerifySignature checks that the given public key created signature over hash.
func VerifySignature(pub, hash, signature []byte) bool {
	if len(signature) != ExtendedSignatureLength {
		return false
	}
	return goldilocks.Ed448Verify(goldilocks.BytesToPublicKey(pub), signature[:SignatureLength], hash, []byte{}, false)
}
