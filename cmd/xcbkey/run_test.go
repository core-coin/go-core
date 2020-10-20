// Copyright 2020 by the Authors
// This file is part of go-core.
//
// go-core is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-core is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-core. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/core-coin/go-core/v2/internal/cmdtest"
	"github.com/docker/docker/pkg/reexec"
)

type testXcbkey struct {
	*cmdtest.TestCmd
}

// spawns xcbkey with the given command line args.
func runXcbkey(t *testing.T, args ...string) *testXcbkey {
	tt := new(testXcbkey)
	tt.TestCmd = cmdtest.NewTestCmd(t, tt)
	tt.Run("xcbkey-test", args...)
	return tt
}

func TestMain(m *testing.M) {
	// Run the app if we've been exec'd as "xcbkey-test" in runXcbkey.
	reexec.Register("xcbkey-test", func() {
		if err := app.Run(os.Args); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	})
	// check if we have been reexec'd
	if reexec.Init() {
		return
	}
	os.Exit(m.Run())
}
