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

package accounts

import (
	"testing"
)

func TestURLParsing(t *testing.T) {
	url, err := parseURL("https://coreblockchain.cc")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if url.Scheme != "https" {
		t.Errorf("expected: %v, got: %v", "https", url.Scheme)
	}
	if url.Path != "coreblockchain.cc" {
		t.Errorf("expected: %v, got: %v", "coreblockchain.cc", url.Path)
	}

	_, err = parseURL("coreblockchain.cc")
	if err == nil {
		t.Error("expected err, got: nil")
	}
}

func TestURLString(t *testing.T) {
	url := URL{Scheme: "https", Path: "coreblockchain.cc"}
	if url.String() != "https://coreblockchain.cc" {
		t.Errorf("expected: %v, got: %v", "https://coreblockchain.cc", url.String())
	}

	url = URL{Scheme: "", Path: "coreblockchain.cc"}
	if url.String() != "coreblockchain.cc" {
		t.Errorf("expected: %v, got: %v", "coreblockchain.cc", url.String())
	}
}

func TestURLMarshalJSON(t *testing.T) {
	url := URL{Scheme: "https", Path: "coreblockchain.cc"}
	json, err := url.MarshalJSON()
	if err != nil {
		t.Errorf("unexpcted error: %v", err)
	}
	if string(json) != "\"https://coreblockchain.cc\"" {
		t.Errorf("expected: %v, got: %v", "\"https://coreblockchain.cc\"", string(json))
	}
}

func TestURLUnmarshalJSON(t *testing.T) {
	url := &URL{}
	err := url.UnmarshalJSON([]byte("\"https://coreblockchain.cc\""))
	if err != nil {
		t.Errorf("unexpcted error: %v", err)
	}
	if url.Scheme != "https" {
		t.Errorf("expected: %v, got: %v", "https", url.Scheme)
	}
	if url.Path != "coreblockchain.cc" {
		t.Errorf("expected: %v, got: %v", "https", url.Path)
	}
}

func TestURLComparison(t *testing.T) {
	tests := []struct {
		urlA   URL
		urlB   URL
		expect int
	}{
		{URL{"https", "coreblockchain.cc"}, URL{"https", "coreblockchain.cc"}, 0},
		{URL{"http", "coreblockchain.cc"}, URL{"https", "coreblockchain.cc"}, -1},
		{URL{"https", "coreblockchain.cc/a"}, URL{"https", "coreblockchain.cc"}, 1},
		{URL{"https", "abc.org"}, URL{"https", "coreblockchain.cc"}, -1},
	}

	for i, tt := range tests {
		result := tt.urlA.Cmp(tt.urlB)
		if result != tt.expect {
			t.Errorf("test %d: cmp mismatch: expected: %d, got: %d", i, tt.expect, result)
		}
	}
}
