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

package v4wire

import (
	"encoding/hex"
	"net"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/rlp"
)

// CIP-8 test vectors.
var testPackets = []struct {
	input      string
	wantPacket Packet
}{
	{
		input: "802afdb4f6c046ec7fc290954c52e9cbe25c224e81e5699d0c5c94ef02e91bbdea4735b2b6ca5359e51a027f9fa23e6d681cc53d0142b79924a1e3b9f1112ae50a8a1aabf1640357e7aa91d56abf216c20d208be15c2867180744f8f87cde0c8ea7cd176901a35debb45607e75e6acc3cb5996592efd1838e89cace74e5ca3111dfdbb53aeb46fd00f0bbc4483a95aa332000ed96186788a1416cc4381045349c228365b85948392a08b783d9b776a056ab5f45b36f957e452f2b174ab2c94bb701a27de1f160c0111b90001ea04cb847f000001820cfa8215a8d790000000000000000000000000000000018208ae820d058443b9a355",
		wantPacket: &Ping{
			Version:    4,
			From:       Endpoint{net.ParseIP("127.0.0.1").To4(), 3322, 5544},
			To:         Endpoint{net.ParseIP("::1"), 2222, 3333},
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{},
		},
	},
	{
		input: "8eb8c1cddec2b7934b87511a284d5655e59d94df4460b1f2410462ebee55e34f6792316f6a12cbbf4ed9c8f8d9cbe3d6c196579d641d03792adbec82ab1b112e26cc535603fafaa11675f7b02e7536af478b81e3f603859a80d9a0e44fd81ecbf0faa5edb62c3cc3eca26931ab8f849baa7a55e5c3ebca06a42daae7fb98bba0db211e6775d3b63d6036723d1c2cab1805000ed96186788a1416cc4381045349c228365b85948392a08b783d9b776a056ab5f45b36f957e452f2b174ab2c94bb701a27de1f160c0111b90001ec04cb847f000001820cfa8215a8d790000000000000000000000000000000018208ae820d058443b9a3550102",
		wantPacket: &Ping{
			Version:    4,
			From:       Endpoint{net.ParseIP("127.0.0.1").To4(), 3322, 5544},
			To:         Endpoint{net.ParseIP("::1"), 2222, 3333},
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0x01}, {0x02}},
		},
	},
	{
		input: "417494d8abac5f08bb678d17d77da9e6456f4090c3e77082ad99d0d9e22dec22d2bc646835057b7500f2a0e6ca0a15ac5354fff280a8090235af817d5956cc16328a44dd1116a864ffdf8c8ffe94f8e047c91deba79f4527803ed1a9e0e3e23626ec2d02d5d2ce2898fb6cf42920ba66a2699197085cc644a1241cec871a77438b4aeff41b9f5dc646ef840183efa42c39000ed96186788a1416cc4381045349c228365b85948392a08b783d9b776a056ab5f45b36f957e452f2b174ab2c94bb701a27de1f160c0111b90001f83e82022bd79020010db83c4d001500000000abcdef12820cfa8215a8d79020010db885a308d313198a2e037073488208ae82823a8443b9a355c50102030405",
		wantPacket: &Ping{
			Version:    555,
			From:       Endpoint{net.ParseIP("2001:db8:3c4d:15::abcd:ef12"), 3322, 5544},
			To:         Endpoint{net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348"), 2222, 33338},
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0xC5, 0x01, 0x02, 0x03, 0x04, 0x05}},
		},
	},
	{
		input: "be643d04b4c84b381991b0aeb62885767fb4832cc0a63fd113080e39aae7fde9852fee56e016a083bf813bfcc3d8542e8c40838ef8803eba7d945a18e40eb79fa8386f1e95c09bd4ef43a2bff343986abfb4654e02199abd00a58975515338f3f14d7b78ae7773f9280e8ea299f50df4bbe282022c411f3bd645f1602d20729ed9778b5b50333fac97a57e58539570c21e000ed96186788a1416cc4381045349c228365b85948392a08b783d9b776a056ab5f45b36f957e452f2b174ab2c94bb701a27de1f160c0111b90002f846d79020010db885a308d313198a2e037073488208ae82823aa0fbc914b16819237dcd8801d7e53f69e9719adecb3cc0e790c57e91ca4461c9548443b9a355c6010203c2040506",
		wantPacket: &Pong{
			To:         Endpoint{net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348"), 2222, 33338},
			ReplyTok:   common.Hex2Bytes("fbc914b16819237dcd8801d7e53f69e9719adecb3cc0e790c57e91ca4461c954"),
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0xC6, 0x01, 0x02, 0x03, 0xC2, 0x04, 0x05}, {0x06}},
		},
	},
	{
		input: "fbe5616915d08b2a685b350a47767441303d112acde59b8c63d6e38b479549696cdf719226288484f7a33cb1a8711fae3277e15537e1b40904c3def1dcf2c292d80c9939933c25efd4305a3bf35135db43a90692399d251100af2e1c0e8e759e9e89c9a623c84594b8e803e68a644ecfbce670a7545ac7168c8588505023ccd5fc4758b62e6536e6d45edb0e0ebcd5623c000ed96186788a1416cc4381045349c228365b85948392a08b783d9b776a056ab5f45b36f957e452f2b174ab2c94bb701a27de1f160c0111b90003f847b8391033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e783639918443b9a35582999983999999",
		wantPacket: &Findnode{
			Target:     hexPubkey("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363991"),
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0x82, 0x99, 0x99}, {0x83, 0x99, 0x99, 0x99}},
		},
	},
	{
		input: "6aefbdad79ef5cf4e40f33a867bb6851dcdb22cfbb000f527ce78f8534a0cb1494e9c5165e02d658fb7f3488dfed50398972817cc86144d358da89058b7c9f6e5bda4ce053ebf4c18e6268e4a57e43822a9831355654ef6780f8a9e11030440081f2c75b8da706997250eadbeb94470ebb069f0e8a5bcd0aeca3f0d1b368af77020b47b2bdf569377334fe08df7290d627000ed96186788a1416cc4381045349c228365b85948392a08b783d9b776a056ab5f45b36f957e452f2b174ab2c94bb701a27de1f160c0111b90004f9013ff90134f846846321163782115c82115db8391033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363992f84284010203040101b8391033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363993f8529020010db83c4d001500000000abcdef12820d05820d05b8391033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363994f8529020010db885a308d313198a2e037073488203e78203e8b8391033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e783639958443b9a355010203",
		wantPacket: &Neighbors{
			Nodes: []Node{
				{
					ID:  hexPubkey("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363992"),
					IP:  net.ParseIP("99.33.22.55").To4(),
					UDP: 4444,
					TCP: 4445,
				},
				{
					ID:  hexPubkey("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363993"),
					IP:  net.ParseIP("1.2.3.4").To4(),
					UDP: 1,
					TCP: 1,
				},
				{
					ID:  hexPubkey("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363994"),
					IP:  net.ParseIP("2001:db8:3c4d:15::abcd:ef12"),
					UDP: 3333,
					TCP: 3333,
				},
				{
					ID:  hexPubkey("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363995"),
					IP:  net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348"),
					UDP: 999,
					TCP: 1000,
				},
			},
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0x01}, {0x02}, {0x03}},
		},
	},
}

// This test checks that the decoder accepts packets according to CIP-8.
func TestForwardCompatibility(t *testing.T) {
	testkey, _ := crypto.UnmarshalPrivateKeyHex("89bdfaa2b6f9c30b94ee98fec96c58ff8507fabf49d36a6267e6cb5516eaa2a9e854eccc041f9f67e109d0eb4f653586855355c5b2b87bb313")
	wantNodeKey := EncodePubkey(testkey.PublicKey())

	for _, test := range testPackets {
		input, err := hex.DecodeString(test.input)
		if err != nil {
			t.Fatalf("invalid hex: %s", test.input)
		}
		packet, nodekey, _, err := Decode(input)
		if err != nil {
			t.Errorf("did not accept packet %s\n%v", test.input, err)
			continue
		}
		if !reflect.DeepEqual(packet, test.wantPacket) {
			t.Errorf("got %s\nwant %s", spew.Sdump(packet), spew.Sdump(test.wantPacket))
		}
		if nodekey != wantNodeKey {
			t.Errorf("got id %v\nwant id %v", nodekey, wantNodeKey)
		}
	}
}

func hexPubkey(h string) (ret Pubkey) {
	b, err := hex.DecodeString(h)
	if err != nil {
		panic(err)
	}
	if len(b) != len(ret) {
		panic("invalid length")
	}
	copy(ret[:], b)
	return ret
}
