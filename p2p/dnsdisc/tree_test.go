// Copyright 2020 The CORE FOUNDATION, nadacia
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

package dnsdisc

import (
	"reflect"
	"testing"

	"github.com/core-coin/go-core/common/hexutil"
	"github.com/core-coin/go-core/p2p/enode"
	"github.com/davecgh/go-spew/spew"
)

func TestParseRoot(t *testing.T) {
	tests := []struct {
		input string
		e     rootEntry
		err   error
	}{
		{
			input: "enrtree-root:v1 e=TO4Q75OQ2N7DX4EOOR7X66A6OM seq=3 sig=N-YY6UB9xD0hFx1Gmnt7v0RfSxch5tKyry2SRDoLx7B4GfPXagwLxQqyf7gAMvApFn_ORwZQekMWa_pXrcGCtw",
			err:   entryError{"root", errSyntax},
		},
		{
			input: "enrtree-root:v1 e=TO4Q75OQ2N7DX4EOOR7X66A6OM l=TO4Q75OQ2N7DX4EOOR7X66A6OM seq=3 sig=N-YY6UB9xD0hFx1Gmnt7v0RfSxch5tKyry2SRDoLx7B4GfPXagwLxQqyf7gAMvApFn_ORwZQekMWa_pXrcGCtw",
			err:   entryError{"root", errInvalidSig},
		},
		{
			input: "enrtree-root:v1 e=QFT4PBCRX4XQCV3VUYJ6BTCEPU l=JGUFMSAGI7KZYB3P7IZW4S5Y3A seq=3 sig=ab4ORUpvvYww0a4tQjlPFzDGsYHpe-KCDtPd90UpFepgkbKePE7GVl-apanTF5epCknYPmWcJkm2C_uQwKuqbv0IXcOo1VJSOap3lxvtqyOpRce37HazGZSaQpmbEk7OpbMjGxMJeirfCy6jS5CiPWHZ_At5y_o8wmjIVjusNrbwkvfjZhTlkpaOzxNv_3D3JcY2hQaVz5N5YELlJDGH2hWzPCh7NMsI",
			e: rootEntry{
				eroot: "QFT4PBCRX4XQCV3VUYJ6BTCEPU",
				lroot: "JGUFMSAGI7KZYB3P7IZW4S5Y3A",
				seq:   3,
				sig:   hexutil.MustDecode("0x69be0e454a6fbd8c30d1ae2d42394f1730c6b181e97be2820ed3ddf7452915ea6091b29e3c4ec6565f9aa5a9d31797a90a49d83e659c2649b60bfb90c0abaa6efd085dc3a8d5525239aa77971bedab23a945c7b7ec76b319949a42999b124ecea5b3231b13097a2adf0b2ea34b90a23d61d9fc0b79cbfa3cc268c8563bac36b6f092f7e36614e592968ecf136fff70f725c636850695cf93796042e5243187da15b33c287b34cb08"),
			},
		},
	}
	for i, test := range tests {
		e, err := parseRoot(test.input)
		if !reflect.DeepEqual(e, test.e) {
			t.Errorf("test %d: wrong entry %s, want %s", i, spew.Sdump(e), spew.Sdump(test.e))
		}
		if err != test.err {
			t.Errorf("test %d: wrong error %q, want %q", i, err, test.err)
		}
	}
}

func TestParseEntry(t *testing.T) {
	testkey := testKey(signingKeySeed)
	tests := []struct {
		input string
		e     entry
		err   error
	}{
		// Subtrees:
		{
			input: "enrtree-branch:1,2",
			err:   entryError{"branch", errInvalidChild},
		},
		{
			input: "enrtree-branch:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
			err:   entryError{"branch", errInvalidChild},
		},
		{
			input: "enrtree-branch:",
			e:     &branchEntry{},
		},
		{
			input: "enrtree-branch:AAAAAAAAAAAAAAAAAAAA",
			e:     &branchEntry{[]string{"AAAAAAAAAAAAAAAAAAAA"}},
		},
		{
			input: "enrtree-branch:AAAAAAAAAAAAAAAAAAAA,BBBBBBBBBBBBBBBBBBBB",
			e:     &branchEntry{[]string{"AAAAAAAAAAAAAAAAAAAA", "BBBBBBBBBBBBBBBBBBBB"}},
		},
		// Links
		{
			input: "enrtree://ZFB62B25A272YUGLWYBUY4XJI2IVOP5ONZUV2OHVFDLC6RLWFCDOQUHKM3SKPBMBFHQ6CA7Q3OEEO2APB5H5XYZ3JQ@nodes.example.org",
			e:     &linkEntry{"ZFB62B25A272YUGLWYBUY4XJI2IVOP5ONZUV2OHVFDLC6RLWFCDOQUHKM3SKPBMBFHQ6CA7Q3OEEO2APB5H5XYZ3JQ@nodes.example.org", "nodes.example.org", &testkey.PublicKey},
		},
		{
			input: "enrtree://nodes.example.org",
			err:   entryError{"link", errNoPubkey},
		},
		{
			input: "enrtree://AP62DT7WOTEQZGQZOU474PP3KMEGVTTE7A7NPRXKX3DUD57@nodes.example.org",
			err:   entryError{"link", errBadPubkey},
		},
		{
			input: "enrtree://AP62DT7WONEQZGQZOU474PP3KMEGVTTE7A7NPRXKX3DUD57TQHGIA@nodes.example.org",
			err:   entryError{"link", errBadPubkey},
		},
		// ENRs
		{
			input: "enr:-PW4qH7cG4dllE36VCDWgjhj5Rv8pchCnzMpXhbMMGNOnqHv02POGtkUewfj1XX9Z3DhggbsmiGYn8qmMC0iBF3d1fyxlhBlX8s4QqrwXoMbs6V08iTaM7_k70Xvtz_Gvk-DLd9HCKsBakSJ_N97_Ny5LQ5vGZVeIhUqYfABAeD0g-l58Suhbo8ZGu8Dv6HUT33kYlbXcXKSeUavooPsGySJV3YUDMra7UuAGICCaWSCdjSJc2VjcDI1NmsxuDhvGZVeIhUqYfABAeD0g-l58Suhbo8ZGu8Dv6HUT33kYlbXcXKSeUavooPsGySJV3YUDMra7UuAGA",
			e:     &enrEntry{node: testNode(nodesSeed1)},
		},
		{
			input: "enr:-HW4QLZHjM4vZXkbp-5xJoHsKSbE7W39FPC8283X-y8oHcHPTnDDlIlzL5ArvDUlHZVDPgmFASrh7cWgLOLxj4wprRkHgmlkgnY0iXNlY3AyNTZrMaEC3t2jLMhDpCDX5mbSEwDn4L3iUfyXzoO8G28XvjGRkrAg=",
			err:   entryError{"enr", errInvalidENR},
		},
		// Invalid:
		{input: "", err: errUnknownEntry},
		{input: "foo", err: errUnknownEntry},
		{input: "enrtree", err: errUnknownEntry},
		{input: "enrtree-x=", err: errUnknownEntry},
	}
	for i, test := range tests {
		e, err := parseEntry(test.input, enode.ValidSchemes)
		if !reflect.DeepEqual(e, test.e) {
			t.Errorf("test %d: wrong entry %s, want %s", i, spew.Sdump(e), spew.Sdump(test.e))
		}
		if err != test.err {
			t.Errorf("test %d: wrong error %q, want %q", i, err, test.err)
		}
	}
}

func TestMakeTree(t *testing.T) {
	nodes := testNodes(nodesSeed2, 50)
	tree, err := MakeTree(2, nodes, nil)
	if err != nil {
		t.Fatal(err)
	}
	txt := tree.ToTXT("")
	if len(txt) < len(nodes)+1 {
		t.Fatal("too few TXT records in output")
	}
}
