// Copyright 2022 by the Authors
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
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/rlp"
	eddsa "github.com/core-coin/go-goldilocks"
	"github.com/davecgh/go-spew/spew"
	"net"
	"reflect"
	"testing"
)

// CIP-8 test vectors.
var testPackets = []struct {
	input      string
	wantPacket interface{}
}{
	{
		input: "3c8ead3bd8396db64aaccebf7c40aa194307a14bb42519967b01a8869b0db7566b8163632a0312ec5808fe0f9dec2cf85698a6ce4c6579217abd426f82ecf68e32a57789721cefa6162acf509c6f2280f673e22afe10d142800c53e7d03fac7a2c93b16a4c5d7e35e3b01ed8139259f9e24026ea396261ef674efccb7f7e95f96dff4d63ed5a8fc4090c5dae66f1c7a40900c0d4a39c59cc720a122fdb9e28d57979c19bf64988e42c2b1bf916d009a26fcc1feacd90b610f7a9ff614320035138a4739ed0e4e877cb1f8001ea04cb847f000001820cfa8215a8d790000000000000000000000000000000018208ae820d058443b9a355",
		wantPacket: &Ping{
			Version:    4,
			From:       Endpoint{net.ParseIP("127.0.0.1").To4(), 3322, 5544},
			To:         Endpoint{net.ParseIP("::1"), 2222, 3333},
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{},
		},
	},
	{
		input: "6242b5a46b9ba8837f3708a59777a4b8640a6cacc6fc1b8ffcbe4dde7b49e09463596abad6d91e706cb20797e88406055c2abc2aa32cfa347b3b8cf7ced2e04c5219c357f78b0a1fc903e6e99c192274186a5211add7845680cf21813043bb7d2d3366a508c5c7f30aad679c85197b58ee0563f4624deaf64be27d94110307df17e074c9812b9c12e22b68c855e68af90500c0d4a39c59cc720a122fdb9e28d57979c19bf64988e42c2b1bf916d009a26fcc1feacd90b610f7a9ff614320035138a4739ed0e4e877cb1f8001ec04cb847f000001820cfa8215a8d790000000000000000000000000000000018208ae820d058443b9a3550102",
		wantPacket: &Ping{
			Version:    4,
			From:       Endpoint{net.ParseIP("127.0.0.1").To4(), 3322, 5544},
			To:         Endpoint{net.ParseIP("::1"), 2222, 3333},
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0x01}, {0x02}},
		},
	},
	{
		input: "a86cccff39536829aa3b26c496da3b12ad595cef598e0184a22f75300ae9c5ff0a5e1b6ed5ce199d9cc84bd12db0a176151c85731956fb2c0486fe2471a79db42fde99edec4e0c80f89910850429cc7a1e1c10c64c5e15dc0062b890d7ede928da241587560a49ded0e089aebe059b70f8f214373df2cb4ac08c4baa9c6a4905d6e71961153cc168aae86d20992d88cc3100c0d4a39c59cc720a122fdb9e28d57979c19bf64988e42c2b1bf916d009a26fcc1feacd90b610f7a9ff614320035138a4739ed0e4e877cb1f8001f83e82022bd79020010db83c4d001500000000abcdef12820cfa8215a8d79020010db885a308d313198a2e037073488208ae82823a8443b9a355c50102030405",
		wantPacket: &Ping{
			Version:    555,
			From:       Endpoint{net.ParseIP("2001:db8:3c4d:15::abcd:ef12"), 3322, 5544},
			To:         Endpoint{net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348"), 2222, 33338},
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0xC5, 0x01, 0x02, 0x03, 0x04, 0x05}},
		},
	},
	{
		input: "da0b91e002acf3b850697491b6962eba12f5f89712532a72150ae4cff18845d75a325b440d96fef78cc4a30420bc60121608df35cc0fcfd604a5bf6ecfe337b403556140cc68faac23674f49864d51e662f98ea224c83dbb00ec5152af3b9694809d0b7344c861f89a712d492f4d64ffd9b9a081cfbf86e1c61522e3277e4b31da4eaed733df66b73565e4ca4d22cdc42f00c0d4a39c59cc720a122fdb9e28d57979c19bf64988e42c2b1bf916d009a26fcc1feacd90b610f7a9ff614320035138a4739ed0e4e877cb1f8002f846d79020010db885a308d313198a2e037073488208ae82823aa0fbc914b16819237dcd8801d7e53f69e9719adecb3cc0e790c57e91ca4461c9548443b9a355c6010203c2040506",
		wantPacket: &Pong{
			To:         Endpoint{net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348"), 2222, 33338},
			ReplyTok:   common.Hex2Bytes("fbc914b16819237dcd8801d7e53f69e9719adecb3cc0e790c57e91ca4461c954"),
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0xC6, 0x01, 0x02, 0x03, 0xC2, 0x04, 0x05}, {0x06}},
		},
	},
	{
		input: "fa6a7f74b9ce9b1d848b1bd9a2e650b39e2123056f7ad0d3afbe85b3cb081ad597227a42a4f0b9434a337163d3ddcd3f393f08f2d8c567fbbb8c1e4df568436bc732df994d6d68ff227844cda9767cebfecaa3a7c12af1ec00635fe52f5835a5734ad8e75bc4b5db3c28146cd51ac07a2c0941b6ad6646ea2e950fb0ba3dcb43ed48ebd4acf24a998ccad34d0b74017b1400c0d4a39c59cc720a122fdb9e28d57979c19bf64988e42c2b1bf916d009a26fcc1feacd90b610f7a9ff614320035138a4739ed0e4e877cb1f8003f847b8394aa8946ebf664270106abd7aea3a27e66124c04deb391de0997be90f973d5369fd287442b79e7553b642aaca22159567b503c36e12891847aa8443b9a35582999983999999",
		wantPacket: &Findnode{
			Target:     hexPubkey("4aa8946ebf664270106abd7aea3a27e66124c04deb391de0997be90f973d5369fd287442b79e7553b642aaca22159567b503c36e12891847aa"),
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0x82, 0x99, 0x99}, {0x83, 0x99, 0x99, 0x99}},
		},
	},
	{
		input: "4ecd2ff52df0f82632b9000d152c561a0437e5819b9d78c31dcf40f7a93c2560f3deaedc5462c71712377c08ec657ffadd3a5de8a45079833563cbf925ed3e2da9f6a2a186c860b75ada6ef1b6147cab49548ccf35dd6c6d80a1c3ac5053c7be1a8d979bf6c1368dcdead23859f6533881b41bb72c7c85b913cf96cecc458be99f5a13253bbd1b9807a035a89e98b5bb1000c0d4a39c59cc720a122fdb9e28d57979c19bf64988e42c2b1bf916d009a26fcc1feacd90b610f7a9ff614320035138a4739ed0e4e877cb1f8004f9013ff90134f846846321163782115c82115db8394aa8946ebf664270106abd7aea3a27e66124c04deb391de0997be90f973d5369fd287442b79e7553b642aaca22159567b503c36e12891847aaf84284010203040101b8395acf8e211e3d3e2ba310afa91edc15389a9cb2e59525774646dc46030d18d880b5f0ef616f842231a355e725589dd45a2f2677d028b6fe46aaf8529020010db83c4d001500000000abcdef12820d05820d05b8394b9fcfcbeb73288841583b311b5ac405e8cba059953db4f7cfa3008a7b41099f8713ba6195ed0168607342675f77d26d5a6088e540401ac2aaf8529020010db885a308d313198a2e037073488203e78203e8b839d3c777805908f662ce69981006a68da37b3e2105ab9a96b28b79832f7f40715e77d1c589f14bc914d7915b8976317c78603b732a7c7f5453aa8443b9a355010203",
		wantPacket: &Neighbors{
			Nodes: []Node{
				{
					ID:  hexPubkey("4aa8946ebf664270106abd7aea3a27e66124c04deb391de0997be90f973d5369fd287442b79e7553b642aaca22159567b503c36e12891847aa"),
					IP:  net.ParseIP("99.33.22.55").To4(),
					UDP: 4444,
					TCP: 4445,
				},
				{
					ID:  hexPubkey("5acf8e211e3d3e2ba310afa91edc15389a9cb2e59525774646dc46030d18d880b5f0ef616f842231a355e725589dd45a2f2677d028b6fe46aa"),
					IP:  net.ParseIP("1.2.3.4").To4(),
					UDP: 1,
					TCP: 1,
				},
				{
					ID:  hexPubkey("4b9fcfcbeb73288841583b311b5ac405e8cba059953db4f7cfa3008a7b41099f8713ba6195ed0168607342675f77d26d5a6088e540401ac2aa"),
					IP:  net.ParseIP("2001:db8:3c4d:15::abcd:ef12"),
					UDP: 3333,
					TCP: 3333,
				},
				{
					ID:  hexPubkey("d3c777805908f662ce69981006a68da37b3e2105ab9a96b28b79832f7f40715e77d1c589f14bc914d7915b8976317c78603b732a7c7f5453aa"),
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

func TestUDPv4_forwardCompatibility(t *testing.T) {
	testkey, _ := crypto.HexToEDDSA("835bbff17efac2c97895784041c507959cdb9e45c599cc205e453a962c11c09ac8834f6524d0842cc469db2afcc0424ca4afc42968d3441846")
	pub := eddsa.Ed448DerivePublicKey(*testkey)
	wantNodeKey := EncodePubkey(&pub)
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
