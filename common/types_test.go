// Copyright 2015 by the Authors
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

package common

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func TestBytesConversion(t *testing.T) {
	bytes := []byte{5}
	hash := BytesToHash(bytes)

	var exp Hash
	exp[31] = 5

	if hash != exp {
		t.Errorf("expected %x got %x", exp, hash)
	}
}

func TestIsHexAddress(t *testing.T) {
	tests := []struct {
		str string
		exp bool
	}{
		{"cc115aaeb6053f3e94c9b9a09f33669435e7ef1beaed", true},
		{"ca115aaeb6053f3e94c9b9a09f33669435e7ef1beaed", true},
		{"cf11AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", true},
		{"cd11AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", true},
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed1", false},
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beae", false},
		{"0xxaaeb6053f3e94c9b9a09f33669435e7ef1beaed", false},
	}

	for _, test := range tests {
		if result := IsHexAddress(test.str); result != test.exp {
			t.Errorf("IsHexAddress(%s) == %v; expected %v",
				test.str, result, test.exp)
		}
	}
}

func TestHashJsonValidation(t *testing.T) {
	var tests = []struct {
		Prefix string
		Size   int
		Error  string
	}{
		{"", 62, "json: cannot unmarshal hex string without 0x prefix into Go value of type common.Hash"},
		{"0x", 66, "hex string has length 66, want 64 for common.Hash"},
		{"0x", 63, "json: cannot unmarshal hex string of odd length into Go value of type common.Hash"},
		{"0x", 0, "hex string has length 0, want 64 for common.Hash"},
		{"0x", 64, ""},
		{"0X", 64, ""},
	}
	for _, test := range tests {
		input := `"` + test.Prefix + strings.Repeat("0", test.Size) + `"`
		var v Hash
		err := json.Unmarshal([]byte(input), &v)
		if err == nil {
			if test.Error != "" {
				t.Errorf("%s: error mismatch: have nil, want %q", input, test.Error)
			}
		} else {
			if err.Error() != test.Error {
				t.Errorf("%s: error mismatch: have %q, want %q", input, err, test.Error)
			}
		}
	}
}

func TestAddressUnmarshalJSON(t *testing.T) {
	firstValue, _ := new(big.Int).SetString("76119257383693214163347625799180543036016898731827897", 10)
	secondValue, _ := new(big.Int).SetString("76029076697074162734871399364090302227946215290884425", 10)
	var tests = []struct {
		Input     string
		ShouldErr bool
		Output    *big.Int
	}{
		{`"0x99"`, true, nil},
		{`"0x0099"`, true, nil},
		{`"0xG000000000000000000000600000000000000000"`, true, nil},
		{`"cd41e8cf4629acb360350399b6cff367a97cf36e62b9"`, true, nil}, // wrong prefix (network id = 1)
		{`"41e8cf4629acb360350399b6cff367a97cf36e62b9"`, true, nil},   // without prefix
		{`"e8cf4629acb360350399b6cff367a97cf36e62b9"`, true, nil},     // without prefix and checksum
		{`"cb35348d6db8bfe52ab1199deeacbc4c1ffa0686d149"`, false, secondValue},
		{`"0xcb35348d6db8bfe52ab1199deeacbc4c1ffa0686d149"`, false, secondValue},
		{`"cb72e8cf4629acb360350399b6cff367a97cf36e62b9"`, false, firstValue},
	}
	for i, test := range tests {
		var v Address
		err := json.Unmarshal([]byte(test.Input), &v)
		if err != nil && !test.ShouldErr {
			t.Errorf("test #%d: unexpected error: %v", i, err)
		}
		if err == nil {
			if test.ShouldErr {
				t.Errorf("test #%d: expected error, got none", i)
			}
			fmt.Println(test.Input)
			if got := new(big.Int).SetBytes(v.Bytes()); got.Cmp(test.Output) != 0 {
				t.Errorf("test #%d: address mismatch: have %v, want %v", i, got, test.Output)
			}
		}
	}
}

func BenchmarkAddressHex(b *testing.B) {
	testAddr, err := HexToAddress("cb885aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	if err != nil {
		b.Error(err)
	}
	for n := 0; n < b.N; n++ {
		testAddr.Hex()
	}
}

func TestHash_Scan(t *testing.T) {
	type args struct {
		src interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "working scan",
			args: args{src: []byte{
				0xb2, 0x6f, 0x2b, 0x34, 0x2a, 0xab, 0x24, 0xbc, 0xf6, 0x3e,
				0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a, 0x15,
				0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a, 0x15,
				0x10, 0x00,
			}},
			wantErr: false,
		},
		{
			name:    "non working scan",
			args:    args{src: int64(1234567890)},
			wantErr: true,
		},
		{
			name: "invalid length scan",
			args: args{src: []byte{
				0xb2, 0x6f, 0x2b, 0x34, 0x2a, 0xab, 0x24, 0xbc, 0xf6, 0x3e,
				0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a, 0x15,
				0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a, 0x15,
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Hash{}
			if err := h.Scan(tt.args.src); (err != nil) != tt.wantErr {
				t.Errorf("Hash.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				for i := range h {
					if h[i] != tt.args.src.([]byte)[i] {
						t.Errorf(
							"Hash.Scan() didn't scan the %d src correctly (have %X, want %X)",
							i, h[i], tt.args.src.([]byte)[i],
						)
					}
				}
			}
		})
	}
}

func TestHash_Value(t *testing.T) {
	b := []byte{
		0xb2, 0x6f, 0x2b, 0x34, 0x2a, 0xab, 0x24, 0xbc, 0xf6, 0x3e,
		0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a, 0x15,
		0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a, 0x15,
		0x10, 0x00,
	}
	var usedH Hash
	usedH.SetBytes(b)
	tests := []struct {
		name    string
		h       Hash
		want    driver.Value
		wantErr bool
	}{
		{
			name:    "Working value",
			h:       usedH,
			want:    b,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hash.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_Scan(t *testing.T) {
	type args struct {
		src interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "working scan",
			args: args{src: []byte{
				0xcb, 0x01,
				0xb2, 0x6f, 0x2b, 0x34, 0x2a, 0xab, 0x24, 0xbc, 0xf6, 0x3e,
				0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a, 0x15,
			}},
			wantErr: false,
		},
		{
			name:    "non working scan",
			args:    args{src: int64(1234567890)},
			wantErr: true,
		},
		{
			name: "invalid length scan",
			args: args{src: []byte{
				0xb2, 0x6f, 0x2b, 0x34, 0x2a, 0xab, 0x24, 0xbc, 0xf6,
				0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a,
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Address{}
			if err := a.Scan(tt.args.src); (err != nil) != tt.wantErr {
				t.Errorf("Address.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				for i := range a {
					if a[i] != tt.args.src.([]byte)[i] {
						t.Errorf(
							"Address.Scan() didn't scan the %d src correctly (have %X, want %X)",
							i, a[i], tt.args.src.([]byte)[i],
						)
					}
				}
			}
		})
	}
}

func TestAddress_Value(t *testing.T) {
	b := []byte{
		0xcb, 0x01,
		0xb2, 0x6f, 0x2b, 0x34, 0x2a, 0xab, 0x24, 0xbc, 0xf6, 0x3e,
		0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a, 0x15,
	}
	var usedA Address
	usedA.SetBytes(b)
	tests := []struct {
		name    string
		a       Address
		want    driver.Value
		wantErr bool
	}{
		{
			name:    "Working value",
			a:       usedA,
			want:    b,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("Address.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Address.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_Format(t *testing.T) {
	b := []byte{
		0xcb, 0x01,
		0xb2, 0x6f, 0x2b, 0x34, 0x2a, 0xab, 0x24, 0xbc, 0xf6, 0x3e,
		0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a, 0x15,
	}
	var addr Address
	addr.SetBytes(b)

	tests := []struct {
		name string
		out  string
		want string
	}{
		{
			name: "println",
			out:  fmt.Sprintln(addr),
			want: "cb01b26f2b342aab24bcf63ea218c6a9274d30ab9a15\n",
		},
		{
			name: "print",
			out:  fmt.Sprint(addr),
			want: "cb01b26f2b342aab24bcf63ea218c6a9274d30ab9a15",
		},
		{
			name: "printf-s",
			out: func() string {
				buf := new(bytes.Buffer)
				fmt.Fprintf(buf, "%s", addr)
				return buf.String()
			}(),
			want: "cb01b26f2b342aab24bcf63ea218c6a9274d30ab9a15",
		},
		{
			name: "printf-q",
			out:  fmt.Sprintf("%q", addr),
			want: `"cb01b26f2b342aab24bcf63ea218c6a9274d30ab9a15"`,
		},
		{
			name: "printf-x",
			out:  fmt.Sprintf("%x", addr),
			want: "cb01b26f2b342aab24bcf63ea218c6a9274d30ab9a15",
		},
		{
			name: "printf-X",
			out:  fmt.Sprintf("%X", addr),
			want: "cb01b26f2b342aab24bcf63ea218c6a9274d30ab9a15",
		},
		{
			name: "printf-#x",
			out:  fmt.Sprintf("%#x", addr),
			want: "cb01b26f2b342aab24bcf63ea218c6a9274d30ab9a15",
		},
		{
			name: "printf-v",
			out:  fmt.Sprintf("%v", addr),
			want: "cb01b26f2b342aab24bcf63ea218c6a9274d30ab9a15",
		},
		// The original default formatter for byte slice
		{
			name: "printf-d",
			out:  fmt.Sprintf("%d", addr),
			want: "[203 1 178 111 43 52 42 171 36 188 246 62 162 24 198 169 39 77 48 171 154 21]",
		},
		// Invalid format char.
		{
			name: "printf-t",
			out:  fmt.Sprintf("%t", addr),
			want: "%!t(address=cb01b26f2b342aab24bcf63ea218c6a9274d30ab9a15)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.out != tt.want {
				t.Errorf("%s does not render as expected:\n got %s\nwant %s", tt.name, tt.out, tt.want)
			}
		})
	}
}

func TestHash_Format(t *testing.T) {
	var hash Hash
	hash.SetBytes([]byte{
		0xb2, 0x6f, 0x2b, 0x34, 0x2a, 0xab, 0x24, 0xbc, 0xf6, 0x3e,
		0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a, 0x15,
		0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a, 0x15,
		0x10, 0x00,
	})

	tests := []struct {
		name string
		out  string
		want string
	}{
		{
			name: "println",
			out:  fmt.Sprintln(hash),
			want: "0xb26f2b342aab24bcf63ea218c6a9274d30ab9a15a218c6a9274d30ab9a151000\n",
		},
		{
			name: "print",
			out:  fmt.Sprint(hash),
			want: "0xb26f2b342aab24bcf63ea218c6a9274d30ab9a15a218c6a9274d30ab9a151000",
		},
		{
			name: "printf-s",
			out: func() string {
				buf := new(bytes.Buffer)
				fmt.Fprintf(buf, "%s", hash)
				return buf.String()
			}(),
			want: "0xb26f2b342aab24bcf63ea218c6a9274d30ab9a15a218c6a9274d30ab9a151000",
		},
		{
			name: "printf-q",
			out:  fmt.Sprintf("%q", hash),
			want: `"0xb26f2b342aab24bcf63ea218c6a9274d30ab9a15a218c6a9274d30ab9a151000"`,
		},
		{
			name: "printf-x",
			out:  fmt.Sprintf("%x", hash),
			want: "b26f2b342aab24bcf63ea218c6a9274d30ab9a15a218c6a9274d30ab9a151000",
		},
		{
			name: "printf-X",
			out:  fmt.Sprintf("%X", hash),
			want: "B26F2B342AAB24BCF63EA218C6A9274D30AB9A15A218C6A9274D30AB9A151000",
		},
		{
			name: "printf-#x",
			out:  fmt.Sprintf("%#x", hash),
			want: "0xb26f2b342aab24bcf63ea218c6a9274d30ab9a15a218c6a9274d30ab9a151000",
		},
		{
			name: "printf-#X",
			out:  fmt.Sprintf("%#X", hash),
			want: "0XB26F2B342AAB24BCF63EA218C6A9274D30AB9A15A218C6A9274D30AB9A151000",
		},
		{
			name: "printf-v",
			out:  fmt.Sprintf("%v", hash),
			want: "0xb26f2b342aab24bcf63ea218c6a9274d30ab9a15a218c6a9274d30ab9a151000",
		},
		// The original default formatter for byte slice
		{
			name: "printf-d",
			out:  fmt.Sprintf("%d", hash),
			want: "[178 111 43 52 42 171 36 188 246 62 162 24 198 169 39 77 48 171 154 21 162 24 198 169 39 77 48 171 154 21 16 0]",
		},
		// Invalid format char.
		{
			name: "printf-t",
			out:  fmt.Sprintf("%t", hash),
			want: "%!t(hash=b26f2b342aab24bcf63ea218c6a9274d30ab9a15a218c6a9274d30ab9a151000)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.out != tt.want {
				t.Errorf("%s does not render as expected:\n got %s\nwant %s", tt.name, tt.out, tt.want)
			}
		})
	}
}
