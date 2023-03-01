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
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/p2p/enr"
	"github.com/core-coin/go-core/v2/rlp"
)

var (
	privkey, _ = crypto.UnmarshalPrivateKeyHex("856a9af6b0b651dd2f43b5e12193652ec1701c4da6f1c0d2a366ac4b9dabc9433ef09e41ca129552bd2c029086d9b03604de872a3b3432041f")
)

func TestEmptyNodeID(t *testing.T) {
	var r enr.Record
	if addr := ValidSchemes.NodeAddr(&r); addr != nil {
		t.Errorf("wrong address on empty record: got %v, want %v", addr, nil)
	}

	require.NoError(t, SignV4(&r, privkey))
	expected := "22384f8cdc73aa2005df6581c0b7afce8c1aa50eb259d87fdf72f6220f173ab6"
	assert.Equal(t, expected, hex.EncodeToString(ValidSchemes.NodeAddr(&r)))
}

// Checks that failure to sign leaves the record unmodified.
func TestSignError(t *testing.T) {
	invalidKey := &crypto.PrivateKey{}

	var r enr.Record
	emptyEnc, _ := rlp.EncodeToBytes(&r)
	if err := SignV4(&r, invalidKey); err == nil {
		t.Fatal("expected error from SignV4")
	}
	newEnc, _ := rlp.EncodeToBytes(&r)
	if !bytes.Equal(newEnc, emptyEnc) {
		t.Fatal("record modified even though signing failed")
	}
}

// TestGetSetEd448 tests encoding/decoding and setting/getting of the Ed448 key.
func TestGetSetEd448(t *testing.T) {
	var r enr.Record
	if err := SignV4(&r, privkey); err != nil {
		t.Fatal(err)
	}

	var pk Ed448
	require.NoError(t, r.Load(&pk))
	assert.EqualValues(t, privkey.PublicKey(), &pk)
}
