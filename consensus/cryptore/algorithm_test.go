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
)

// Tests whether the cryptonight lookup works for both light as well as the full
// datasets.
func TestCryptonight(t *testing.T) {
	// Create a block to verify
	hash := hexutil.MustDecode("0xc9149cc0386e689d789a1c2f3d5d169a61a6218ed30e74414dc736e442ef3d1f")
	nonce := uint64(0)

	wantDigest := hexutil.MustDecode("0x7496850e31f0c8b44aae2d57704312657496850e31f0c8b44aae2d5770431265")
	wantResult := hexutil.MustDecode("0x1f375cf3374bfa1bbda32674ef9077e7823b95a1a92ed07b5cea4d6004abf012")

	digest, result := hashcryptonight(hash, nonce)
	if !bytes.Equal(digest, wantDigest) {
		t.Errorf("cryptonight digest mismatch: have %x, want %x", digest, wantDigest)
	}
	if !bytes.Equal(result, wantResult) {
		t.Errorf("cryptonight result mismatch: have %x, want %x", result, wantResult)
	}
}
