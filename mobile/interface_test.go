// Copyright 2019 The go-core Authors
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

package gocore

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/core-coin/go-core/common"
)

func TestInterfaceGetSet(t *testing.T) {
	deadbeef, err := common.HexToAddress("cb4812345678123456781234567812345678deadbeef")
	if err != nil {
		t.Error(err)
	}
	cafebabe, err := common.HexToAddress("cb9512345678123456781234567812345678cafebabe")
	if err != nil {
		t.Error(err)
	}
	var tests = []struct {
		method string
		input  interface{}
		expect interface{}
	}{
		{"Bool", true, true},
		{"Bool", false, false},
		{"Bools", &Bools{[]bool{false, true}}, &Bools{[]bool{false, true}}},
		{"String", "go-core", "go-core"},
		{"Strings", &Strings{strs: []string{"hello", "world"}}, &Strings{strs: []string{"hello", "world"}}},
		{"Binary", []byte{0x01, 0x02}, []byte{0x01, 0x02}},
		{"Binaries", &Binaries{[][]byte{{0x01, 0x02}, {0x03, 0x04}}}, &Binaries{[][]byte{{0x01, 0x02}, {0x03, 0x04}}}},
		{"Address", &Address{deadbeef}, &Address{deadbeef}},
		{"Addresses", &Addresses{[]common.Address{deadbeef, cafebabe}}, &Addresses{[]common.Address{deadbeef, cafebabe}}},
		{"Hash", &Hash{common.HexToHash("deadbeef")}, &Hash{common.HexToHash("deadbeef")}},
		{"Hashes", &Hashes{[]common.Hash{common.HexToHash("deadbeef"), common.HexToHash("cafebabe")}}, &Hashes{[]common.Hash{common.HexToHash("deadbeef"), common.HexToHash("cafebabe")}}},
		{"Int8", int8(1), int8(1)},
		{"Int16", int16(1), int16(1)},
		{"Int32", int32(1), int32(1)},
		{"Int64", int64(1), int64(1)},
		{"Int8s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Int16s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Int32s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Int64s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Uint8", NewBigInt(1), NewBigInt(1)},
		{"Uint16", NewBigInt(1), NewBigInt(1)},
		{"Uint32", NewBigInt(1), NewBigInt(1)},
		{"Uint64", NewBigInt(1), NewBigInt(1)},
		{"Uint8s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Uint16s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Uint32s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"Uint64s", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
		{"BigInt", NewBigInt(1), NewBigInt(1)},
		{"BigInts", &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}, &BigInts{[]*big.Int{big.NewInt(1), big.NewInt(2)}}},
	}

	args := NewInterfaces(len(tests))

	callFn := func(receiver interface{}, method string, arg interface{}) interface{} {
		rval := reflect.ValueOf(receiver)
		rval.MethodByName(fmt.Sprintf("Set%s", method)).Call([]reflect.Value{reflect.ValueOf(arg)})
		res := rval.MethodByName(fmt.Sprintf("Get%s", method)).Call(nil)
		if len(res) > 0 {
			return res[0].Interface()
		}
		return nil
	}

	for index, c := range tests {
		// In theory the change of iface shouldn't effect the args value
		iface, _ := args.Get(index)
		result := callFn(iface, c.method, c.input)
		if !reflect.DeepEqual(result, c.expect) {
			t.Errorf("Interface get/set mismatch, want %v, got %v", c.expect, result)
		}
		// Check whether the underlying value in args is still zero
		iface, _ = args.Get(index)
		if iface.object != nil {
			t.Error("Get operation is not write safe")
		}
	}
}
