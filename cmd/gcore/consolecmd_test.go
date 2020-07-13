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
	"crypto/rand"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/core-coin/go-core/params"
)

const (
	ipcAPIs  = "admin:1.0 cryptore:1.0 debug:1.0 miner:1.0 net:1.0 personal:1.0 rpc:1.0 shh:1.0 txpool:1.0 web3:1.0 xcc:1.0"
	httpAPIs = "net:1.0 rpc:1.0 web3:1.0 xcc:1.0"
)

// Tests that a node embedded within a console can be started up properly and
// then terminated by closing the input stream.
func TestConsoleWelcome(t *testing.T) {
	coinbase := "398605cdbbdb6d264aa742e77020dcbc58fcdce182"

	// Start a gcore console, make sure it's cleaned up and terminate the console
	gcore := runGcore(t,
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--corebase", coinbase, "--shh",
		"console")

	// Gather all the infos the welcome message needs to contain
	gcore.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	gcore.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	gcore.SetTemplateFunc("gover", runtime.Version)
	gcore.SetTemplateFunc("gcorever", func() string { return params.VersionWithCommit("", "") })
	gcore.SetTemplateFunc("niltime", func() string {
		return time.Unix(0, 0).Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
	})
	gcore.SetTemplateFunc("apis", func() string { return ipcAPIs })

	// Verify the actual welcome message to the required template
	gcore.Expect(`
Welcome to the Gcore JavaScript console!

instance: Gcore/v{{gcorever}}/{{goos}}-{{goarch}}/{{gover}}
coinbase: {{.Corebase}}
at block: 0 ({{niltime}})
 datadir: {{.Datadir}}
 modules: {{apis}}

> {{.InputLine "exit"}}
`)
	gcore.ExpectExit()
}

// Tests that a console can be attached to a running node via various means.
func TestIPCAttachWelcome(t *testing.T) {
	// Configure the instance for IPC attachement
	coinbase := "398605cdbbdb6d264aa742e77020dcbc58fcdce182"
	var ipc string
	if runtime.GOOS == "windows" {
		ipc = `\\.\pipe\gcore` + strconv.Itoa(trulyRandInt(100000, 999999))
	} else {
		ws := tmpdir(t)
		defer os.RemoveAll(ws)
		ipc = filepath.Join(ws, "gcore.ipc")
	}
	// Note: we need --shh because testAttachWelcome checks for default
	// list of ipc modules and shh is included there.
	gcore := runGcore(t,
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--corebase", coinbase, "--shh", "--ipcpath", ipc)

	defer func() {
		gcore.Interrupt()
		gcore.ExpectExit()
	}()

	waitForEndpoint(t, ipc, 3*time.Second)
	testAttachWelcome(t, gcore, "ipc:"+ipc, ipcAPIs)

}

func TestHTTPAttachWelcome(t *testing.T) {
	coinbase := "398605cdbbdb6d264aa742e77020dcbc58fcdce182"
	port := strconv.Itoa(trulyRandInt(1024, 65536)) // Yeah, sometimes this will fail, sorry :P
	gcore := runGcore(t,
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--corebase", coinbase, "--rpc", "--rpcport", port)
	defer func() {
		gcore.Interrupt()
		gcore.ExpectExit()
	}()

	endpoint := "http://127.0.0.1:" + port
	waitForEndpoint(t, endpoint, 3*time.Second)
	testAttachWelcome(t, gcore, endpoint, httpAPIs)
}

func TestWSAttachWelcome(t *testing.T) {
	coinbase := "398605cdbbdb6d264aa742e77020dcbc58fcdce182"
	port := strconv.Itoa(trulyRandInt(1024, 65536)) // Yeah, sometimes this will fail, sorry :P

	gcore := runGcore(t,
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--corebase", coinbase, "--ws", "--wsport", port)
	defer func() {
		gcore.Interrupt()
		gcore.ExpectExit()
	}()

	endpoint := "ws://127.0.0.1:" + port
	waitForEndpoint(t, endpoint, 3*time.Second)
	testAttachWelcome(t, gcore, endpoint, httpAPIs)
}

func testAttachWelcome(t *testing.T, gcore *testgcore, endpoint, apis string) {
	// Attach to a running gcore note and terminate immediately
	attach := runGcore(t, "attach", endpoint)
	defer attach.ExpectExit()
	attach.CloseStdin()

	// Gather all the infos the welcome message needs to contain
	attach.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	attach.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	attach.SetTemplateFunc("gover", runtime.Version)
	attach.SetTemplateFunc("gcorever", func() string { return params.VersionWithCommit("", "") })
	attach.SetTemplateFunc("corebase", func() string { return gcore.Corebase })
	attach.SetTemplateFunc("niltime", func() string {
		return time.Unix(0, 0).Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
	})
	attach.SetTemplateFunc("ipc", func() bool { return strings.HasPrefix(endpoint, "ipc") })
	attach.SetTemplateFunc("datadir", func() string { return gcore.Datadir })
	attach.SetTemplateFunc("apis", func() string { return apis })

	// Verify the actual welcome message to the required template
	attach.Expect(`
Welcome to the Gcore JavaScript console!

instance: Gcore/v{{gcorever}}/{{goos}}-{{goarch}}/{{gover}}
coinbase: {{corebase}}
at block: 0 ({{niltime}}){{if ipc}}
 datadir: {{datadir}}{{end}}
 modules: {{apis}}

> {{.InputLine "exit" }}
`)
	attach.ExpectExit()
}

// trulyRandInt generates a crypto random integer used by the console tests to
// not clash network ports with other tests running cocurrently.
func trulyRandInt(lo, hi int) int {
	num, _ := rand.Int(rand.Reader, big.NewInt(int64(hi-lo)))
	return int(num.Int64()) + lo
}
