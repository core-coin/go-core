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

package fourbyte

import (
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/core-coin/go-core/v2/accounts/abi"
	"github.com/core-coin/go-core/v2/common"
)

func verify(t *testing.T, jsondata, calldata string, exp []interface{}) {
	abispec, err := abi.JSON(strings.NewReader(jsondata))
	if err != nil {
		t.Fatal(err)
	}
	cd := common.Hex2Bytes(calldata)
	sigdata, argdata := cd[:4], cd[4:]
	method, err := abispec.MethodById(sigdata)
	if err != nil {
		t.Fatal(err)
	}
	data, err := method.Inputs.UnpackValues(argdata)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(exp) {
		t.Fatalf("Mismatched length, expected %d, got %d", len(exp), len(data))
	}
	for i, elem := range data {
		if !reflect.DeepEqual(elem, exp[i]) {
			t.Fatalf("Unpack error, arg %d, got %v, want %v", i, elem, exp[i])
		}
	}
}

func TestNewUnpacker(t *testing.T) {
	type unpackTest struct {
		jsondata string
		calldata string
		exp      []interface{}
	}
	addr, err := common.HexToAddress("cb4900000133700000deadbeef000000000000000000")
	if err != nil {
		t.Error(err)
	}
	testcases := []unpackTest{
		{ // https://ylem.readthedocs.io/en/develop/abi-spec.html#use-of-dynamic-types
			`[{"type":"function","name":"f", "inputs":[{"type":"uint256"},{"type":"uint32[]"},{"type":"bytes10"},{"type":"bytes"}]}]`,
			// 0x123, [0x456, 0x789], "1234567890", "Hello, world!"
			"41cc790900000000000000000000000000000000000000000000000000000000000001230000000000000000000000000000000000000000000000000000000000000080313233343536373839300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000004560000000000000000000000000000000000000000000000000000000000000789000000000000000000000000000000000000000000000000000000000000000d48656c6c6f2c20776f726c642100000000000000000000000000000000000000",
			[]interface{}{
				big.NewInt(0x123),
				[]uint32{0x456, 0x789},
				[10]byte{49, 50, 51, 52, 53, 54, 55, 56, 57, 48},
				common.Hex2Bytes("48656c6c6f2c20776f726c6421"),
			},
		},
		{ // https://github.com/core/wiki/wiki/Core-Contract-ABI#examples
			`[{"type":"function","name":"sam","inputs":[{"type":"bytes"},{"type":"bool"},{"type":"uint256[]"}]}]`,
			//  "dave", true and [1,2,3]
			"cf3d9ff50000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000464617665000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003",
			[]interface{}{
				[]byte{0x64, 0x61, 0x76, 0x65},
				true,
				[]*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
			},
		},
		{
			`[{"type":"function","name":"send","inputs":[{"type":"uint256"}]}]`,
			"e10c92510000000000000000000000000000000000000000000000000000000000000012",
			[]interface{}{big.NewInt(0x12)},
		},
		{
			`[{"type":"function","name":"compareAndApprove","inputs":[{"type":"address"},{"type":"uint256"},{"type":"uint256"}]}]`,
			"db33183a00000000000000000000cb4900000133700000deadbeef00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001",
			[]interface{}{
				addr,
				new(big.Int).SetBytes([]byte{0x00}),
				big.NewInt(0x1),
			},
		},
	}
	for _, c := range testcases {
		verify(t, c.jsondata, c.calldata, c.exp)
	}
}

func TestCalldataDecoding(t *testing.T) {
	// send(uint256)                              : e10c9251
	// compareAndApprove(address,uint256,uint256) : db33183a
	// issue(address[],uint256)                   : 649cfa14
	jsondata := `
[
	{"type":"function","name":"send","inputs":[{"name":"a","type":"uint256"}]},
	{"type":"function","name":"compareAndApprove","inputs":[{"name":"a","type":"address"},{"name":"a","type":"uint256"},{"name":"a","type":"uint256"}]},
	{"type":"function","name":"issue","inputs":[{"name":"a","type":"address[]"},{"name":"a","type":"uint256"}]},
	{"type":"function","name":"sam","inputs":[{"name":"a","type":"bytes"},{"name":"a","type":"bool"},{"name":"a","type":"uint256[]"}]}
]`
	// Expected failures
	for i, hexdata := range []string{
		"e10c925100000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000042",
		"e10c9251000000000000000000000000000000000000000000000000000000000000001200",
		"e10c925100000000000000000000000000000000000000000000000000000000000000",
		"e10c9251",
		"a52c10",
		"",
		// Too short
		"db33183a0000000000000000000000000000000000000000000000000000000000000012",
		"db33183aFFffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		// Not valid multiple of 32
		"deadbeef00000000000000000000000000000000000000000000000000000000000000",
		// Too short 'issue'
		"649cfa1400000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000042",
		// Too short compareAndApprove
		"e10c925100ff0000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000042",
	} {
		_, err := parseCallData(common.Hex2Bytes(hexdata), jsondata)
		if err == nil {
			t.Errorf("test %d: expected decoding to fail: %s", i, hexdata)
		}
	}
	// Expected success
	for i, hexdata := range []string{
		// From https://github.com/core/wiki/wiki/Core-Contract-ABI
		"e10c92510000000000000000000000000000000000000000000000000000000000000012",
		"e10c9251FFffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"db33183a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"649cfa14" +
			// start of dynamic type
			"0000000000000000000000000000000000000000000000000000000000000040" +
			// uint256
			"0000000000000000000000000000000000000000000000000000000000000001" +
			// length of  array
			"0000000000000000000000000000000000000000000000000000000000000002" +
			// array values
			"000000000000000000000000000000000000000000000000000000000000dead" +
			"000000000000000000000000000000000000000000000000000000000000beef",
	} {
		_, err := parseCallData(common.Hex2Bytes(hexdata), jsondata)
		if err != nil {
			t.Errorf("test %d: unexpected failure on input %s:\n %v (%d bytes) ", i, hexdata, err, len(common.Hex2Bytes(hexdata)))
		}
	}
}

func TestMaliciousABIStrings(t *testing.T) {
	tests := []string{
		"func(uint256,uint256,[]uint256)",
		"func(uint256,uint256,uint256,)",
		"func(,uint256,uint256,uint256)",
	}
	data := common.Hex2Bytes("4401a6e40000000000000000000000000000000000000000000000000000000000000012")
	for i, tt := range tests {
		_, err := verifySelector(tt, data)
		if err == nil {
			t.Errorf("test %d: expected error for selector '%v'", i, tt)
		}
	}
}
