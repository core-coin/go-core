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

package cryptore

import (
	"bytes"
	"testing"

	"github.com/core-coin/go-core/common/hexutil"

	"github.com/core-coin/go-randomx"
)

// Tests whether the randomx lookup works for both light as well as the full
// datasets.
func TestRandomX(t *testing.T) {
	// Create a block to verify
	hash := hexutil.MustDecode("0xc9149cc0386e689d789a1c2f3d5d169a61a6218ed30e74414dc736e442ef3d1f")
	nonce := uint64(0)

	wantResult := hexutil.MustDecode("0xb620364373923b57353c668dcedcfc636d456e1c0d7da8733586c0e54ada6aa4")
	vm, mutex := randomx.NewRandomXVMWithKeyAndMutex()
	result, err := randomx.RandomX(vm, mutex, hash, nonce)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(result, wantResult) {
		t.Errorf("cryptonight result mismatch: have %x, want %x", result, wantResult)
	}
}
