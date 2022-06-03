// Copyright 2019 by the Authors
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
	eddsa "github.com/core-coin/ed448"
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
			input: "enrtree-root:v1 e=QAO2PMLNIZU6HVN4OGDWWQ4UCQ l=UXCFO6UUOO6E4BTJNZPMHAH2HY seq=1 sig=dKx74xYd64nBzlEXzYdK2UiCiKT05oqi3rL3C75nQ4dWxPh2RyY6R7hcXbDPomvdd7LuAbzTPhYAQWgNfReglAJL6gm6baXw46oA-5KpvZIt6qmUwFE8W2BUQWl34Xr4m3hKmUlU4AJO9OR8rNci5h8ASnsKymLWBgRmVCKoPAQ5OAZITJYZuTbCaLUSZjp91KylRphiCcVVLJpwqIOG5K34ZndqiPAvZ0QA",
			e: rootEntry{
				eroot: "QAO2PMLNIZU6HVN4OGDWWQ4UCQ",
				lroot: "UXCFO6UUOO6E4BTJNZPMHAH2HY",
				seq:   1,
				sig:   hexutil.MustDecode("0x74ac7be3161deb89c1ce5117cd874ad9488288a4f4e68aa2deb2f70bbe67438756c4f87647263a47b85c5db0cfa26bdd77b2ee01bcd33e160041680d7d17a094024bea09ba6da5f0e3aa00fb92a9bd922deaa994c0513c5b6054416977e17af89b784a994954e0024ef4e47cacd722e61f004a7b0aca62d60604665422a83c04393806484c9619b936c268b512663a7dd4aca546986209c5552c9a70a88386e4adf866776a88f02f674400"),
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
	testkey := testKey(nodesSeed1)
	pub := eddsa.Ed448DerivePublicKey(*testkey)
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
			input: "enrtree://SALGRL5CVC6ZXZSAJ5HJWZ5X5I2MURYEE7MRPZFFS3FRN4OJJDC2ISGFNZOKZPAUC3B2SOWTTAX4F3GD6VOR777KM2AA@nodes.example.org",
			e:     &linkEntry{"SALGRL5CVC6ZXZSAJ5HJWZ5X5I2MURYEE7MRPZFFS3FRN4OJJDC2ISGFNZOKZPAUC3B2SOWTTAX4F3GD6VOR777KM2AA@nodes.example.org", "nodes.example.org", &pub},
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
			input: "enr:-Pm4qzzDbjdzIJFOmZh6d1MP3dEFux7I5rEVg5jJ9kJHHli36IDCVM5L0bm1_MmD-KdxejveeMGCJ05_AHeBVyk5UqI1s7z1Tl1fabjz2RKwFI_eO8kEYZTY__1uzGDKL7GmeVSaw1Yzdc13QGgRYQfks3UpAJAWaK-iqL2b5kBPTptnt-o0ykcEJ9kX5KWWyxbxyUjFpEjFblysvBQWw6k605gvwuzD9V0f_-pmgICCaWSCdjSJc2VjcDI1NmsxuDmQFmivoqi9m-ZAT06bZ7fqNMpHBCfZF-SllssW8clIxaRIxW5crLwUFsOpOtOYL8Lsw_VdH__qZoA",
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
