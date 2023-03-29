// Copyright 2020 by the Authors
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

package v5wire

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/crypto/sha3"

	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/p2p/enode"
)

func TestVector_ECDH(t *testing.T) {
	var (
		staticKey = hexPrivkey("0x1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363991")
		publicKey = hexPubkey("0x1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363992")
		want      = hexutil.MustDecode("0x1f361690a5508a7d725bce70d87e284cf1f7cad2b0a1bc1518dac3668c401d4a32944f26f18136203125ceb06266c3c53f766b80beea2659")
	)
	result := ecdh(staticKey, publicKey)
	check(t, "shared-secret", result, want)
}

func TestVector_KDF(t *testing.T) {
	var (
		ephKey = hexPrivkey("0x1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363991")
		cdata  = hexutil.MustDecode("0x000000000000000000000000000000006469736376350001010102030405060708090a0b0c00180102030405060708090a0b0c0d0e0f100000000000000000")
		net    = newHandshakeTest()
	)
	defer net.close()

	s := deriveKeys(sha3.New256, ephKey, testKeyB.PublicKey(), net.nodeA.id(), net.nodeB.id(), cdata)
	t.Logf("ephemeral-key = %#x", ephKey.PrivateKey())
	t.Logf("dest-pubkey = %#x", EncodePubkey(testKeyB.PublicKey()))
	t.Logf("node-id-a = %#x", net.nodeA.id().Bytes())
	t.Logf("node-id-b = %#x", net.nodeB.id().Bytes())
	t.Logf("challenge-data = %#x", cdata)
	check(t, "initiator-key", s.writeKey, hexutil.MustDecode("0x32f7236576f4f16d00850a0bc1849d65"))
	check(t, "recipient-key", s.readKey, hexutil.MustDecode("0xcb0f6a26538541a796211d6b074237d2"))
}

func TestVector_IDSignature(t *testing.T) {
	var (
		key    = hexPrivkey("0x1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363991")
		destID = enode.HexID("0xbbbb9d047f0488c0b5a93c1c3f2d8bafc7c8ff337024a55434a0d0555de64db9")
		ephkey = hexutil.MustDecode("0x1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363992")
		cdata  = hexutil.MustDecode("0x000000000000000000000000000000006469736376350001010102030405060708090a0b0c00180102030405060708090a0b0c0d0e0f100000000000000000")
	)

	sig, err := makeIDSignature(sha3.New256(), key, cdata, ephkey, destID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("static-key = %#x", key.PrivateKey())
	t.Logf("challenge-data = %#x", cdata)
	t.Logf("ephemeral-pubkey = %#x", ephkey)
	t.Logf("node-id-B = %#x", destID.Bytes())
	expected := "0x602c459e5ec3401426132f09c525bcf31404b6bb7ff2b0c079bc261a18ab0ce19c54784baed55c5d9b12a0a43c41be3aca422588c546e53d80b4140b4eb4fec32ab6c5e0f0e310f6ad1a3ed8097f98f39958b4699231b209e6d936854a37d5f63cdded03a9a9a75508d6d985a881fe3f32002f65ab658f3b0bc9fbdea48703b9c5c0dc2151c5ae8c4b77b1e5cdaee9fa20748e01960ab51ddb118d1209f73d186f0444921ad72c7c757480"
	check(t, "id-signature", sig, hexutil.MustDecode(expected))
}

func TestDeriveKeys(t *testing.T) {
	t.Parallel()

	var (
		n1    = enode.ID{1}
		n2    = enode.ID{2}
		cdata = []byte{1, 2, 3, 4}
	)
	sec1 := deriveKeys(sha3.New256, testKeyA, testKeyB.PublicKey(), n1, n2, cdata)
	sec2 := deriveKeys(sha3.New256, testKeyB, testKeyA.PublicKey(), n1, n2, cdata)
	if sec1 == nil || sec2 == nil {
		t.Fatal("key agreement failed")
	}
	if !reflect.DeepEqual(sec1, sec2) {
		t.Fatalf("keys not equal:\n  %+v\n  %+v", sec1, sec2)
	}
}

func check(t *testing.T, what string, x, y []byte) {
	t.Helper()

	if !bytes.Equal(x, y) {
		t.Errorf("wrong %s: %#x != %#x", what, x, y)
	} else {
		t.Logf("%s = %#x", what, x)
	}
}

func hexPrivkey(input string) *crypto.PrivateKey {
	key, err := crypto.UnmarshalPrivateKeyHex(strings.TrimPrefix(input, "0x"))
	if err != nil {
		panic(err)
	}
	return key
}

func hexPubkey(input string) *crypto.PublicKey {
	key, err := DecodePubkey(hexutil.MustDecode(input))
	if err != nil {
		panic(err)
	}
	return key
}
