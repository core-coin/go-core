// Copyright 2023 by the Authors
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
	"io"

	"github.com/core-coin/go-goldilocks"
)

func GenerateKey(read io.Reader) (*PrivateKey, error) {
	key, err := goldilocks.Ed448GenerateKey(read)
	if err != nil {
		return nil, err
	}
	return UnmarshalPrivateKey(key[:])

}

func DerivePublicKey(priv *PrivateKey) *PublicKey {
	key := goldilocks.Ed448DerivePublicKey(goldilocks.BytesToPrivateKey(priv.privateKey[:]))
	pub := PublicKey{}
	copy(pub[:], key[:])
	return &pub
}

func ComputeSecret(privkey *PrivateKey, pubkey *PublicKey) []byte {
	priv := goldilocks.PrivateKey{}
	pub := goldilocks.PublicKey{}

	copy(priv[:], privkey.privateKey[:])
	copy(pub[:], pubkey[:])

	secret := goldilocks.Ed448DeriveSecret(pub, priv)
	return secret[:]
}
