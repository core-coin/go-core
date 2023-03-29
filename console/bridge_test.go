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

package console

import (
	"testing"

	"github.com/dop251/goja"

	"github.com/core-coin/go-core/v2/internal/jsre"
)

// TestUndefinedAsParam ensures that personal functions can receive
// `undefined` as a parameter.
func TestUndefinedAsParam(t *testing.T) {
	b := bridge{}
	call := jsre.Call{}
	call.Arguments = []goja.Value{goja.Undefined()}

	b.UnlockAccount(call)
	b.Sign(call)
	b.Sleep(call)
}

// TestNullAsParam ensures that personal functions can receive
// `null` as a parameter.
func TestNullAsParam(t *testing.T) {
	b := bridge{}
	call := jsre.Call{}
	call.Arguments = []goja.Value{goja.Null()}

	b.UnlockAccount(call)
	b.Sign(call)
	b.Sleep(call)
}
