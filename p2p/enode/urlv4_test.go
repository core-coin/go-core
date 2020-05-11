// Copyright 2018 The go-core Authors
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
	"github.com/core-coin/eddsa"
	"errors"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/p2p/enr"
)

func init() {
	lookupIPFunc = func(name string) ([]net.IP, error) {
		if name == "node.example.org" {
			return []net.IP{{33, 44, 55, 66}}, nil
		}
		return nil, errors.New("no such host")
	}
}

var parseNodeTests = []struct {
	input      string
	wantError  string
	wantResult *Node
}{
	// Records
	{
		input: "enr:-QEEuKg-xhCAH4rxbeh_FAWSBkFQJg2uYqutz-8XM8SuYhFeOhX3tM1ZjGb6jPtvDYCpWULW2w5MiH_eHNBZcae1eKJBenm9qvrqyLEGqOppUceywwdFsGztcvKbBgHSuR8YmVH2417PFz6Fiz6WPd42wwYf8rzvYCNcq3aPLD-PMB6M_uQeBPJtjQHEahCwjdPSfGHKSN_HQz0tXjv2C4Ys7dMZfoLAtwnHXEdjgmlkgnY0gmlwhH8AAAGJc2VjcDI1NmsxuDjyvO9gI1yrdo8sP48wHoz-5B4E8m2NAcRqELCN09J8YcpI38dDPS1eO_YLhizt0xl-gsC3CcdcR4N1ZHCCdlw",
		wantResult: func() *Node {
			testKey, _ := crypto.HexToEDDSA("07e988804055546babfb00e34d015314a21a76a1cb049cad4adeb3d931af355f2393ba45bfda9aeb7ca40c1e0a4e63ba4639e43957a54109f2bcef60235cab768f2c3f8f301e8cfee41e04f26d8d01c46a10b08dd3d27c61ca48dfc7433d2d5e3bf60b862cedd3197e82c0b709c75c47ced2896631075043550b8d6b0cfb0ec165d178df945ff8038f30c9ada2e7a69e")
			var r enr.Record
			r.Set(enr.IP{127, 0, 0, 1})
			r.Set(enr.UDP(30300))
			r.SetSeq(99)
			SignV4(&r, testKey)
			n, _ := New(ValidSchemes, &r)
			return n
		}(),
	},
	// Invalid Records
	{
		input:     "enr:",
		wantError: "EOF", // could be nicer
	},
	{
		input:     "enr:x",
		wantError: "illegal base64 data at input byte 0",
	},
	{
		input:     "enr:-EOAY4JpZIRudWxsgmlwhH8AAAGIbnVsbGFkZHKgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACDdWRwgnZf",
		wantError: enr.ErrInvalidSig.Error(),
	},
	// Complete node URLs with IP address and ports
	{
		input:     "enode://767b2d3eb5828a9c2e11d1e9aa515f0435ee7ce80f3749ed33f462921587bc9bb55a7231b2e79ae6ce86e8ff4f83e9a151e855d6bbe87838@invalid.:3",
		wantError: `no such host`,
	},
	{
		input:     "enode://1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f632@127.0.0.1:foo",
		wantError: `invalid port`,
	},
	{
		input:     "enode://1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f632@127.0.0.1:3?discport=foo",
		wantError: `invalid discport in query`,
	},
	{
		input: "enode://1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f632@127.0.0.1:52150",
		wantResult: NewV4(
			hexPubkey("1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f632"),
			net.IP{127, 0, 0, 1},
			52150,
			52150,
		),
	},
	{
		input: "enode://1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f632@[::]:52150",
		wantResult: NewV4(
			hexPubkey("1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f632"),
			net.ParseIP("::"),
			52150,
			52150,
		),
	},
	{
		input: "enode://1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f632@[2001:db8:3c4d:15::abcd:ef12]:52150",
		wantResult: NewV4(
			hexPubkey("1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f632"),
			net.ParseIP("2001:db8:3c4d:15::abcd:ef12"),
			52150,
			52150,
		),
	},
	{
		input: "enode://1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f632@127.0.0.1:52150?discport=22334",
		wantResult: NewV4(
			hexPubkey("1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f632"),
			net.IP{0x7f, 0x0, 0x0, 0x1},
			52150,
			22334,
		),
	},
	// Incomplete node URLs with no address
	{
		input: "enode://1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f632",
		wantResult: NewV4(
			hexPubkey("1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f632"),
			nil, 0, 0,
		),
	},
	// Invalid URLs
	{
		input:     "",
		wantError: errMissingPrefix.Error(),
	},
	{
		input:     "1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f632",
		wantError: errMissingPrefix.Error(),
	},
	{
		input:     "01010101",
		wantError: errMissingPrefix.Error(),
	},
	{
		input:     "enode://01010101@123.124.125.126:3",
		wantError: `invalid public key (wrong length, want 112 hex chars)`,
	},
	{
		input:     "enode://01010101",
		wantError: `invalid public key (wrong length, want 112 hex chars)`,
	},
	{
		input:     "http://foobar",
		wantError: errMissingPrefix.Error(),
	},
	{
		input:     "://foo",
		wantError: errMissingPrefix.Error(),
	},
}

func hexPubkey(h string) *eddsa.PublicKey {
	k, err := parsePubkey(h)
	if err != nil {
		panic(err)
	}
	return k
}

func TestParseNode(t *testing.T) {
	for _, test := range parseNodeTests {
		n, err := Parse(ValidSchemes, test.input)
		if test.wantError != "" {
			if err == nil {
				t.Errorf("test %q:\n  got nil error, expected %#q", test.input, test.wantError)
				continue
			} else if !strings.Contains(err.Error(), test.wantError) {
				t.Errorf("test %q:\n  got error %#q, expected %#q", test.input, err.Error(), test.wantError)
				continue
			}
		} else {
			if err != nil {
				t.Errorf("test %q:\n  unexpected error: %v", test.input, err)
				continue
			}
			if !reflect.DeepEqual(n, test.wantResult) {
				t.Errorf("test %q:\n  result mismatch:\ngot:  %#v\nwant: %#v", test.input, n, test.wantResult)
			}
		}
	}
}

func TestNodeString(t *testing.T) {
	for i, test := range parseNodeTests {
		if test.wantError == "" && strings.HasPrefix(test.input, "enode://") {
			str := test.wantResult.String()
			if str != test.input {
				t.Errorf("test %d: Node.String() mismatch:\ngot:  %s\nwant: %s", i, str, test.input)
			}
		}
	}
}
