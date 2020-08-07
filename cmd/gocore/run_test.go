// Copyright 2016 The go-core Authors
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
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/core-coin/go-core/internal/cmdtest"
	"github.com/core-coin/go-core/rpc"
	"github.com/docker/docker/pkg/reexec"
)

func tmpdir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "gocore-test")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

type testgocore struct {
	*cmdtest.TestCmd

	// template variables for expect
	Datadir  string
	Corebase string
}

func init() {
	// Run the app if we've been exec'd as "gocore-test" in runGocore.
	reexec.Register("gocore-test", func() {
		if err := app.Run(os.Args); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func TestMain(m *testing.M) {
	// check if we have been reexec'd
	if reexec.Init() {
		return
	}
	os.Exit(m.Run())
}

// spawns gocore with the given command line args. If the args don't set --datadir, the
// child g gets a temporary data directory.
func runGocore(t *testing.T, args ...string) *testgocore {
	tt := &testgocore{}
	tt.TestCmd = cmdtest.NewTestCmd(t, tt)
	for i, arg := range args {
		switch {
		case arg == "-datadir" || arg == "--datadir":
			if i < len(args)-1 {
				tt.Datadir = args[i+1]
			}
		case arg == "-corebase" || arg == "--corebase":
			if i < len(args)-1 {
				tt.Corebase = args[i+1]
			}
		}
	}
	if tt.Datadir == "" {
		tt.Datadir = tmpdir(t)
		tt.Cleanup = func() { os.RemoveAll(tt.Datadir) }
		args = append([]string{"-datadir", tt.Datadir}, args...)
		// Remove the temporary datadir if something fails below.
		defer func() {
			if t.Failed() {
				tt.Cleanup()
			}
		}()
	}

	// Boot "gocore". This actually runs the test binary but the TestMain
	// function will prevent any tests from running.
	tt.Run("gocore-test", args...)

	return tt
}

// waitForEndpoint attempts to connect to an RPC endpoint until it succeeds.
func waitForEndpoint(t *testing.T, endpoint string, timeout time.Duration) {
	probe := func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		c, err := rpc.DialContext(ctx, endpoint)
		if c != nil {
			_, err = c.SupportedModules()
			c.Close()
		}
		return err == nil
	}

	start := time.Now()
	for {
		if probe() {
			return
		}
		if time.Since(start) > timeout {
			t.Fatal("endpoint", endpoint, "did not open within", timeout)
		}
		time.Sleep(200 * time.Millisecond)
	}
}
