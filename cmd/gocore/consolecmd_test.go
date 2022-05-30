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
	ipcAPIs  = "admin:1.0 cryptore:1.0 debug:1.0 miner:1.0 net:1.0 personal:1.0 rpc:1.0 txpool:1.0 web3:1.0 xcb:1.0"
	httpAPIs = "net:1.0 rpc:1.0 web3:1.0 xcb:1.0"
)

// Tests that a node embedded within a console can be started up properly and
// then terminated by closing the input stream.
func TestConsoleWelcome(t *testing.T) {
	coinbase := "cb348605cdbbdb6d264aa742e77020dcbc58fcdce182"

	// Start a gocore console, make sure it's cleaned up and terminate the console
	gocore := runGocore(t,
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--corebase", coinbase,
		"console")

	// Gather all the infos the welcome message needs to contain
	gocore.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	gocore.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	gocore.SetTemplateFunc("gover", runtime.Version)
	gocore.SetTemplateFunc("gocorever", func() string { return params.VersionWithTag("", "", "") })
	gocore.SetTemplateFunc("niltime", func() string {
		return time.Unix(1651833293, 0).Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
	})
	gocore.SetTemplateFunc("apis", func() string { return ipcAPIs })

	// Verify the actual welcome message to the required template
	gocore.Expect(`
Welcome to the Gocore JavaScript console!

instance: Gocore/{{goos}}-{{goarch}}/{{gover}}
coinbase: {{.Corebase}}
at block: 0 ({{niltime}})
 datadir: {{.Datadir}}
 modules: {{apis}}

> {{.InputLine "exit"}}
`)
	gocore.ExpectExit()
}

// Tests that a console can be attached to a running node via various means.
func TestIPCAttachWelcome(t *testing.T) {
	// Configure the instance for IPC attachement
	coinbase := "cb348605cdbbdb6d264aa742e77020dcbc58fcdce182"
	timeout := 15 * time.Second
	var ipc string
	if runtime.GOOS == "windows" {
		ipc = `\\.\pipe\gocore` + strconv.Itoa(trulyRandInt(100000, 999999))
	} else {
		ws := tmpdir(t)
		defer os.RemoveAll(ws)
		ipc = filepath.Join(ws, "gocore.ipc")
	}
	gocore := runGocore(t,
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--corebase", coinbase, "--ipcpath", ipc)

	defer func() {
		gocore.Interrupt()
		gocore.ExpectExit()
	}()

	waitForEndpoint(t, ipc, timeout)
	testAttachWelcome(t, gocore, "ipc:"+ipc, ipcAPIs)

}

func TestHTTPAttachWelcome(t *testing.T) {
	coinbase := "cb348605cdbbdb6d264aa742e77020dcbc58fcdce182"
	timeout := 15 * time.Second
	port := strconv.Itoa(trulyRandInt(1024, 65536)) // Yeah, sometimes this will fail, sorry :P
	gocore := runGocore(t,
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--corebase", coinbase, "--rpc", "--rpcport", port)
	defer func() {
		gocore.Interrupt()
		gocore.ExpectExit()
	}()

	endpoint := "http://127.0.0.1:" + port
	waitForEndpoint(t, endpoint, timeout)
	testAttachWelcome(t, gocore, endpoint, httpAPIs)
}

func TestWSAttachWelcome(t *testing.T) {
	coinbase := "cb348605cdbbdb6d264aa742e77020dcbc58fcdce182"
	timeout := 15 * time.Second
	port := strconv.Itoa(trulyRandInt(1024, 65536)) // Yeah, sometimes this will fail, sorry :P

	gocore := runGocore(t,
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--corebase", coinbase, "--ws", "--wsport", port)
	defer func() {
		gocore.Interrupt()
		gocore.ExpectExit()
	}()

	endpoint := "ws://127.0.0.1:" + port
	waitForEndpoint(t, endpoint, timeout)
	testAttachWelcome(t, gocore, endpoint, httpAPIs)
}

func testAttachWelcome(t *testing.T, gocore *testgocore, endpoint, apis string) {
	// Attach to a running gocore note and terminate immediately
	attach := runGocore(t, "attach", endpoint)
	defer attach.ExpectExit()
	attach.CloseStdin()

	// Gather all the infos the welcome message needs to contain
	attach.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	attach.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	attach.SetTemplateFunc("gover", runtime.Version)
	attach.SetTemplateFunc("gocorever", func() string { return params.VersionWithTag("", "", "") })
	attach.SetTemplateFunc("corebase", func() string { return gocore.Corebase })
	attach.SetTemplateFunc("niltime", func() string {
		return time.Unix(1651833293, 0).Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
	})
	attach.SetTemplateFunc("ipc", func() bool { return strings.HasPrefix(endpoint, "ipc") })
	attach.SetTemplateFunc("datadir", func() string { return gocore.Datadir })
	attach.SetTemplateFunc("apis", func() string { return apis })

	// Verify the actual welcome message to the required template
	attach.Expect(`
Welcome to the Gocore JavaScript console!

instance: Gocore/{{goos}}-{{goarch}}/{{gover}}
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
