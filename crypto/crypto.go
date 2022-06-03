// Copyright 2014 by the Authors
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

package crypto

import (
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"

	eddsa "github.com/core-coin/ed448"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/rlp"
	"golang.org/x/crypto/sha3"
)

const SignatureLength = 114
const PubkeyLength = 57
const PrivkeyLength = 57
const ExtendedSignatureLength = SignatureLength + PubkeyLength

var errInvalidPubkey = errors.New("invalid public key")
var errInvalidPrivkey = errors.New("invalid private key")
var errInvalidSignature = errors.New("invalid signature")

// Keccak256 calculates and returns the Keccak256 hash of the input data. //Substituted in CoreCoin by NIPS SHA3-256
func Keccak256(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

//NIPS implementation of SHA3-256
func SHA3(data ...[]byte) []byte {
	d := sha3.New256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

//NIPS implementation of SHA3-512
func SHA3_512(data ...[]byte) []byte {
	d := sha3.New512()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// Keccak256Hash calculates and returns the Keccak256 hash of the input data,
// converting it to an internal Hash data structure.
func Keccak256Hash(data ...[]byte) (h common.Hash) {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}

// SHA3Hash calculates and returns the NIPS SHA3-256 hash of the input data,
// converting it to an internal Hash data structure.
func SHA3Hash(data ...[]byte) (h common.Hash) {
	d := sha3.New256()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}

// Keccak512 calculates and returns the Keccak512 hash of the input data.
func Keccak512(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak512()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// CreateAddress creates an core address given the bytes and the nonce
func CreateAddress(b common.Address, nonce uint64) common.Address {
	data, _ := rlp.EncodeToBytes([]interface{}{b, nonce})
	addr := SHA3(data)[12:]
	prefix := common.DefaultNetworkID.Bytes()
	checksum := common.Hex2Bytes(common.CalculateChecksum(addr, prefix))
	return common.BytesToAddress(append(append(prefix, checksum...), addr...))
}

// CreateAddress2 creates an core address given the address bytes, initial
// contract code hash and a salt.
func CreateAddress2(b common.Address, salt [32]byte, inithash []byte) common.Address {
	addr := SHA3([]byte{0xff}, b.Bytes(), salt[:], inithash)[12:]
	prefix := common.DefaultNetworkID.Bytes()
	checksum := common.Hex2Bytes(common.CalculateChecksum(addr, prefix))
	return common.BytesToAddress(append(append(prefix, checksum...), addr...))
}

// ToEDDSA creates a private key with the given D value.
func ToEDDSA(d []byte) (*eddsa.PrivateKey, error) {
	return toEDDSA(d, true)
}

// ToEDDSAUnsafe blindly converts a binary blob to a private key. It should almost
// never be used unless you are sure the input is valid and want to avoid hitting
// errors due to bad origin encoding (0 prefixes cut off).
func ToEDDSAUnsafe(d []byte) *eddsa.PrivateKey {
	priv, err := toEDDSA(d, false)
	if err != nil {
		panic(err)
	}
	return priv
}

// toEDDSA creates a private key with the given D value. The strict parameter
// controls whether the key's length should be enforced at the curve size or
// it can also accept legacy encodings (0 prefixes).
func toEDDSA(d []byte, strict bool) (*eddsa.PrivateKey, error) {
	_ = strict
	if len(d) != PrivkeyLength {
		return nil, errInvalidPrivkey
	}
	p := eddsa.PrivateKey{}
	copy(p[:], d[:])
	return &p, nil
}

// FromEDDSA exports a private key into a binary dump.
func FromEDDSA(priv *eddsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	return priv[:]
}

// UnmarshalPubkey converts bytes to a secp256k1 public key.
func UnmarshalPubkey(pub []byte) (*eddsa.PublicKey, error) {
	if len(pub) != PubkeyLength {
		return nil, errInvalidPubkey
	}
	p := eddsa.PublicKey{}
	copy(p[:], pub[:])
	return &p, nil
}

func FromEDDSAPub(pub *eddsa.PublicKey) []byte {
	if pub == nil {
		return nil
	}
	return pub[:]
}

// HexToEDDSA parses a secp256k1 private key.
func HexToEDDSA(hexkey string) (*eddsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	return ToEDDSA(b)
}

// LoadEDDSA loads a secp256k1 private key from the given file.
func LoadEDDSA(file string) (*eddsa.PrivateKey, error) {
	buf := make([]byte, PrivkeyLength*2)
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer fd.Close()
	if _, err := io.ReadFull(fd, buf); err != nil {
		return nil, err
	}
	key, err := hex.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}
	return toEDDSA(key, true)
}

// SaveEDDSA saves a secp256k1 private key to the given file with
// restrictive permissions. The key data is saved hex-encoded.
func SaveEDDSA(file string, key *eddsa.PrivateKey) error {
	k := hex.EncodeToString(key[:])
	return ioutil.WriteFile(file, []byte(k), 0600)
}

func GenerateKey(read io.Reader) (*eddsa.PrivateKey, error) {
	key, err := eddsa.Ed448GenerateKey(read)
	return &key, err
}

// ValidateSignatureValues verifies whether the signature values are valid with
// the given chain rules. The v value is assumed to be either 0 or 1.
func ValidateSignatureValues(v byte) bool {
	return v == 0 || v == 1
}

func ComputeSecret(privkey *eddsa.PrivateKey, pubkey *eddsa.PublicKey) []byte {
	secret := eddsa.Ed448DeriveSecret(*pubkey, *privkey)
	return secret[:]
}

func PubkeyToAddress(p eddsa.PublicKey) common.Address {
	pubBytes := FromEDDSAPub(&p)
	if pubBytes == nil {
		return common.Address{}
	}
	addr := SHA3(pubBytes)[12:]
	prefix := common.DefaultNetworkID.Bytes()
	checksum := common.Hex2Bytes(common.CalculateChecksum(addr, prefix))
	return common.BytesToAddress(append(append(prefix, checksum...), addr...))
}

func zeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}
