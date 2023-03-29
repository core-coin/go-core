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

package crypto

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/sha3"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/rlp"
)

type PublicKey [57]byte

type PrivateKey struct {
	privateKey [57]byte
	publicKey  *PublicKey
	address    common.Address
}

func (p *PrivateKey) PrivateKey() []byte {
	return p.privateKey[:]
}

func (p *PrivateKey) PublicKey() *PublicKey {
	if p.publicKey == nil {
		p.publicKey = DerivePublicKey(p)
	}
	return p.publicKey
}

func (p *PrivateKey) Address() common.Address {
	if p.publicKey == nil {
		p.publicKey = DerivePublicKey(p)
	}
	if (common.Address{} == p.address) {
		p.address = PubkeyToAddress(p.publicKey)
	}
	return p.address
}

const SignatureLength = 114
const PubkeyLength = 57
const PrivkeyLength = 57
const ExtendedSignatureLength = SignatureLength + PubkeyLength

var errInvalidPubkey = errors.New("invalid public key")
var errInvalidPrivkey = errors.New("invalid private key")
var errInvalidSignature = errors.New("invalid signature")

// SHA3State wraps sha3.state. In addition to the usual hash methods, it also supports
// Read to get a variable amount of data from the hash state. Read is faster than Sum
// because it doesn't copy the internal state, but also modifies the internal state.
type SHA3State interface {
	hash.Hash
	Read([]byte) (int, error)
}

// SHA3 calculates and returns the SHA3 hash of the input data.
func SHA3(data ...[]byte) []byte {
	b := make([]byte, 32)
	d := sha3.New256().(SHA3State)
	for _, b := range data {
		d.Write(b)
	}
	d.Read(b)
	return b
}

// SHA3Hash calculates and returns the SHA3 hash of the input data,
// converting it to an internal Hash data structure.
func SHA3Hash(data ...[]byte) (h common.Hash) {
	d := sha3.New256().(SHA3State)
	for _, b := range data {
		d.Write(b)
	}
	d.Read(h[:])
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

func PubkeyToAddress(p *PublicKey) common.Address {
	addr := SHA3(p[:])[12:]
	prefix := common.DefaultNetworkID.Bytes()
	checksum := common.Hex2Bytes(common.CalculateChecksum(addr, prefix))
	return common.BytesToAddress(append(append(prefix, checksum...), addr...))
}

// UnmarshalPrivateKey creates a private key with the given D value.
func UnmarshalPrivateKey(d []byte) (*PrivateKey, error) {
	if len(d) != PrivkeyLength {
		return nil, errInvalidPrivkey
	}
	priv := PrivateKey{}
	copy(priv.privateKey[:], d[:])
	return &priv, nil
}

// UnmarshalPrivateKeyHex parses a private key.
func UnmarshalPrivateKeyHex(hexkey string) (*PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if byteErr, ok := err.(hex.InvalidByteError); ok {
		return nil, fmt.Errorf("invalid hex character %q in private key", byte(byteErr))
	} else if err != nil {
		return nil, errors.New("invalid hex data for private key")
	}
	return UnmarshalPrivateKey(b)
}

// MarshalPrivateKey exports a private key into a binary dump.
func MarshalPrivateKey(priv *PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	return priv.privateKey[:]
}

// UnmarshalPubKey converts bytes to a public key.
func UnmarshalPubKey(pub []byte) (*PublicKey, error) {
	if len(pub) != PubkeyLength {
		return nil, errInvalidPubkey
	}
	p := PublicKey{}
	copy(p[:], pub[:])
	return &p, nil
}

// LoadEDDSA loads a ed448 private key from the given file.
func LoadEDDSA(file string) (*PrivateKey, error) {
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	r := bufio.NewReader(fd)
	buf := make([]byte, PrivkeyLength*2)
	n, err := readASCII(buf, r)
	if err != nil {
		return nil, err
	} else if n != len(buf) {
		return nil, fmt.Errorf("key file too short, want 57 hex characters")
	}
	if err := checkKeyFileEnd(r); err != nil {
		return nil, err
	}

	return UnmarshalPrivateKeyHex(string(buf))
}

// Ecrecover returns the public key that created the given signature.
func Ecrecover(hash, sig []byte) ([]byte, error) {
	pubkey, err := SigToPub(hash, sig)
	if err != nil {
		return nil, err
	}
	return pubkey[:], nil
}

// SaveEDDSA saves a private key to the given file with
// restrictive permissions. The key data is saved hex-encoded.
func SaveEDDSA(file string, key *PrivateKey) error {
	k := hex.EncodeToString(key.privateKey[:])
	return ioutil.WriteFile(file, []byte(k), 0600)
}

// readASCII reads into 'buf', stopping when the buffer is full or
// when a non-printable control character is encountered.
func readASCII(buf []byte, r *bufio.Reader) (n int, err error) {
	for ; n < len(buf); n++ {
		buf[n], err = r.ReadByte()
		switch {
		case err == io.EOF || buf[n] < '!':
			return n, nil
		case err != nil:
			return n, err
		}
	}
	return n, nil
}

// checkKeyFileEnd skips over additional newlines at the end of a key file.
func checkKeyFileEnd(r *bufio.Reader) error {
	for i := 0; ; i++ {
		b, err := r.ReadByte()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case b != '\n' && b != '\r':
			return fmt.Errorf("invalid character %q at end of key file", b)
		case i >= 2:
			return errors.New("key file too long, want 57 hex characters")
		}
	}
}

func zeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}
