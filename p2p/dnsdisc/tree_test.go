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
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/p2p/enode"
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
			input: "enrtree-root:v1 e=7525E77SHV6G7TBPKTRFYQKNEI l=TCMJVN7523OUUZZ5MYGACYBBIU seq=1 sig=hbftK14lRXy81Y1WtqhWgJ8VVOT2bYo31mTDhouxMCA-ENY3JA0Du6imZIEd6voIgm1yn4_k2aKAWsfLkyKs6-q6rZhOWd81CuDYg2eEPUcK1FyWaj1hMMXj3TGVPgSqfHLX6AU1xYKcchdwCeiY0hAAhR3c_l0JEw3XUFpFrCjAQu-MCmR7bhcFQ_-KFwrG6r_1oQ2Q_-glxs9NYZz2ueAgQ8Lm0te4tZsA",
			e: rootEntry{
				eroot: "7525E77SHV6G7TBPKTRFYQKNEI",
				lroot: "TCMJVN7523OUUZZ5MYGACYBBIU",
				seq:   1,
				sig:   hexutil.MustDecode("0x85b7ed2b5e25457cbcd58d56b6a856809f1554e4f66d8a37d664c3868bb130203e10d637240d03bba8a664811deafa08826d729f8fe4d9a2805ac7cb9322acebeabaad984e59df350ae0d88367843d470ad45c966a3d6130c5e3dd31953e04aa7c72d7e80535c5829c72177009e898d21000851ddcfe5d09130dd7505a45ac28c042ef8c0a647b6e170543ff8a170ac6eabff5a10d90ffe825c6cf4d619cf6b9e02043c2e6d2d7b8b59b00"),
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
	pub := testkey.PublicKey()
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
			e:     &linkEntry{"SALGRL5CVC6ZXZSAJ5HJWZ5X5I2MURYEE7MRPZFFS3FRN4OJJDC2ISGFNZOKZPAUC3B2SOWTTAX4F3GD6VOR777KM2AA@nodes.example.org", "nodes.example.org", pub},
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
