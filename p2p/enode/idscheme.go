// Copyright 2018 by the Authors
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

package enode

import (
	"fmt"
	"io"

	"golang.org/x/crypto/sha3"

	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/p2p/enr"
	"github.com/core-coin/go-core/v2/rlp"
)

// List of known secure identity schemes.
var ValidSchemes = enr.SchemeMap{
	"v4": V4ID{},
}

var ValidSchemesForTesting = enr.SchemeMap{
	"v4":   V4ID{},
	"null": NullID{},
}

// v4ID is the "v4" identity scheme.
type V4ID struct{}

// SignV4 signs a record using the v4 scheme.
func SignV4(r *enr.Record, privkey *crypto.PrivateKey) error {
	// Copy r to avoid modifying it if signing fails.
	cpy := *r
	cpy.Set(enr.ID("v4"))
	cpy.Set(Ed448(*privkey.PublicKey()))

	h := sha3.New256()
	err := rlp.Encode(h, cpy.AppendElements(nil))
	if err != nil {
		return err
	}

	sig, err := crypto.Sign(h.Sum(nil), privkey)
	if err != nil {
		return err
	}
	if err = cpy.SetSig(V4ID{}, sig); err == nil {
		*r = cpy
	}
	return err
}

func (V4ID) Verify(r *enr.Record, sig []byte) error {
	var entry ed448raw
	if err := r.Load(&entry); err != nil {
		return err
	} else if len(entry) != 57 {
		return fmt.Errorf("invalid public key")
	}

	h := sha3.New256()
	err := rlp.Encode(h, r.AppendElements(nil))
	if err != nil {
		return err
	}

	if !crypto.VerifySignature(entry, h.Sum(nil), sig) {
		return enr.ErrInvalidSig
	}
	return nil
}

func (V4ID) NodeAddr(r *enr.Record) []byte {
	var pubkey Ed448
	err := r.Load(&pubkey)
	if err != nil {
		return nil
	}
	return crypto.SHA3(pubkey[:])
}

// Ed448 is the "ed448" key, which holds a public key.
type Ed448 crypto.PublicKey

func (v Ed448) ENRKey() string { return "secp256k1" }

// EncodeRLP implements rlp.Encoder.
func (v Ed448) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, (*crypto.PublicKey)(&v))
}

// DecodeRLP implements rlp.Decoder.
func (v *Ed448) DecodeRLP(s *rlp.Stream) error {
	buf, err := s.Bytes()
	if err != nil {
		return err
	}
	pk, err := crypto.UnmarshalPubKey(buf)
	if err != nil {
		return err
	}
	*v = (Ed448)(*pk)
	return nil
}

// ed448raw is an unparsed ed448 public key entry.
type ed448raw []byte

func (ed448raw) ENRKey() string { return "secp256k1" }

// v4CompatID is a weaker and insecure version of the "v4" scheme which only checks for the
// presence of a ed448 public key, but doesn't verify the signature.
type v4CompatID struct {
	V4ID
}

func (v4CompatID) Verify(r *enr.Record, sig []byte) error {
	var pubkey Ed448
	return r.Load(&pubkey)
}

func signV4Compat(r *enr.Record, pubkey *crypto.PublicKey) {
	r.Set((*Ed448)(pubkey))
	if err := r.SetSig(v4CompatID{}, []byte{}); err != nil {
		panic(err)
	}
}

// NullID is the "null" ENR identity scheme. This scheme stores the node
// ID in the record without any signature.
type NullID struct{}

func (NullID) Verify(r *enr.Record, sig []byte) error {
	return nil
}

func (NullID) NodeAddr(r *enr.Record) []byte {
	var id ID
	r.Load(enr.WithEntry("nulladdr", &id))
	return id[:]
}

func SignNull(r *enr.Record, id ID) *Node {
	r.Set(enr.ID("null"))
	r.Set(enr.WithEntry("nulladdr", id))
	if err := r.SetSig(NullID{}, []byte{}); err != nil {
		panic(err)
	}
	return &Node{r: *r, id: id}
}
