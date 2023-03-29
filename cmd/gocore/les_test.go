// Copyright 2023 by the Authors
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
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/core-coin/go-core/v2/p2p"
	"github.com/core-coin/go-core/v2/rpc"
)

type gocorerpc struct {
	name     string
	rpc      *rpc.Client
	gocore   *testgocore
	nodeInfo *p2p.NodeInfo
}

func (g *gocorerpc) killAndWait() {
	g.gocore.Kill()
	g.gocore.WaitExit()
}

func (g *gocorerpc) callRPC(result interface{}, method string, args ...interface{}) {
	if err := g.rpc.Call(&result, method, args...); err != nil {
		g.gocore.Fatalf("callRPC %v: %v", method, err)
	}
}

func (g *gocorerpc) addPeer(peer *gocorerpc) {
	g.gocore.Logf("%v.addPeer(%v)", g.name, peer.name)
	enode := peer.getNodeInfo().Enode
	peerCh := make(chan *p2p.PeerEvent)
	sub, err := g.rpc.Subscribe(context.Background(), "admin", peerCh, "peerEvents")
	if err != nil {
		g.gocore.Fatalf("subscribe %v: %v", g.name, err)
	}
	defer sub.Unsubscribe()
	g.callRPC(nil, "admin_addPeer", enode)
	dur := 30 * time.Second
	timeout := time.After(dur)
	select {
	case ev := <-peerCh:
		g.gocore.Logf("%v received event: type=%v, peer=%v", g.name, ev.Type, ev.Peer)
	case err := <-sub.Err():
		g.gocore.Fatalf("%v sub error: %v", g.name, err)
	case <-timeout:
		g.gocore.Error("timeout adding peer after", dur)
	}
}

// Use this function instead of `g.nodeInfo` directly
func (g *gocorerpc) getNodeInfo() *p2p.NodeInfo {
	if g.nodeInfo != nil {
		return g.nodeInfo
	}
	g.nodeInfo = &p2p.NodeInfo{}
	g.callRPC(&g.nodeInfo, "admin_nodeInfo")
	return g.nodeInfo
}

func (g *gocorerpc) waitSynced() {
	// Check if it's synced now
	var result interface{}
	g.callRPC(&result, "xcb_syncing")
	syncing, ok := result.(bool)
	if ok && !syncing {
		g.gocore.Logf("%v already synced", g.name)
		return
	}

	// Actually wait, subscribe to the event
	ch := make(chan interface{})
	sub, err := g.rpc.Subscribe(context.Background(), "xcb", ch, "syncing")
	if err != nil {
		g.gocore.Fatalf("%v syncing: %v", g.name, err)
	}
	defer sub.Unsubscribe()
	timeout := time.After(10 * time.Second)
	select {
	case ev := <-ch:
		g.gocore.Log("'syncing' event", ev)
		syncing, ok := ev.(bool)
		if ok && !syncing {
			break
		}
		g.gocore.Log("Other 'syncing' event", ev)
	case err := <-sub.Err():
		g.gocore.Fatalf("%v notification: %v", g.name, err)
		break
	case <-timeout:
		g.gocore.Fatalf("%v timeout syncing", g.name)
		break
	}
}

// ipcEndpoint resolves an IPC endpoint based on a configured value, taking into
// account the set data folders as well as the designated platform we're currently
// running on.
func ipcEndpoint(ipcPath, datadir string) string {
	// On windows we can only use plain top-level pipes
	if runtime.GOOS == "windows" {
		if strings.HasPrefix(ipcPath, `\\.\pipe\`) {
			return ipcPath
		}
		return `\\.\pipe\` + ipcPath
	}
	// Resolve names into the data directory full paths otherwise
	if filepath.Base(ipcPath) == ipcPath {
		if datadir == "" {
			return filepath.Join(os.TempDir(), ipcPath)
		}
		return filepath.Join(datadir, ipcPath)
	}
	return ipcPath
}

// nextIPC ensures that each ipc pipe gets a unique name.
// On linux, it works well to use ipc pipes all over the filesystem (in datadirs),
// but windows require pipes to sit in "\\.\pipe\". Therefore, to run several
// nodes simultaneously, we need to distinguish between them, which we do by
// the pipe filename instead of folder.
var nextIPC = uint32(0)

func startGocoreWithIpc(t *testing.T, name string, args ...string) *gocorerpc {
	ipcName := fmt.Sprintf("gocore-%d.ipc", atomic.AddUint32(&nextIPC, 1))
	args = append([]string{"--networkid=42", "--port=0", "--ipcpath", ipcName}, args...)
	t.Logf("Starting %v with rpc: %v", name, args)

	g := &gocorerpc{
		name:   name,
		gocore: runGocore(t, args...),
	}
	// wait before we can attach to it. TODO: probe for it properly
	time.Sleep(10 * time.Second)
	var err error
	ipcpath := ipcEndpoint(ipcName, g.gocore.Datadir)
	if g.rpc, err = rpc.Dial(ipcpath); err != nil {
		t.Fatalf("%v rpc connect to %v: %v", name, ipcpath, err)
	}
	return g
}

func initGocore(t *testing.T) string {
	args := []string{"--networkid=42", "init", "./testdata/clique.json"}
	t.Logf("Initializing gocore: %v ", args)
	g := runGocore(t, args...)
	datadir := g.Datadir
	g.WaitExit()
	return datadir
}

func startLightServer(t *testing.T) *gocorerpc {
	datadir := initGocore(t)
	t.Logf("Importing keys to gocore")
	runGocore(t, "--datadir", datadir, "--networkid=42", "--password", "./testdata/password.txt", "account", "import", "./testdata/key.prv", "--lightkdf").WaitExit()
	account := "ce4404ad003b70526a4a1ac8ba121f2f96d95b6e11e0" // 9c0c5806dc0b007ed080762cadde92b3e4fefb9034f0a0e3e279a1d4139e38fd20364df61bc0eb34104e01764ae028470a73a167c6e624d443
	server := startGocoreWithIpc(t, "lightserver", "--networkid=42", "--allow-insecure-unlock", "--datadir", datadir, "--password", "./testdata/password.txt", "--unlock", account, "--mine", "--light.serve=100", "--light.maxpeers=1", "--nodiscover", "--nat=extip:127.0.0.1", "--verbosity=5")
	return server
}

func startClient(t *testing.T, name string) *gocorerpc {
	datadir := initGocore(t)
	return startGocoreWithIpc(t, name, "--networkid=42", "--datadir", datadir, "--nodiscover", "--syncmode=light", "--nat=extip:127.0.0.1", "--verbosity=5")
}

func TestPriorityClient(t *testing.T) {
	t.Skip()
	lightServer := startLightServer(t)
	defer lightServer.killAndWait()

	// Start client and add lightServer as peer
	freeCli := startClient(t, "freeCli")
	defer freeCli.killAndWait()
	freeCli.addPeer(lightServer)

	var peers []*p2p.PeerInfo
	freeCli.callRPC(&peers, "admin_peers")
	if len(peers) != 1 {
		t.Errorf("Expected: # of client peers == 1, actual: %v", len(peers))
		return
	}

	// Set up priority client, get its nodeID, increase its balance on the lightServer
	prioCli := startClient(t, "prioCli")
	defer prioCli.killAndWait()
	// 3_000_000_000 once we move to Go 1.13
	tokens := uint64(3000000000)
	lightServer.callRPC(nil, "les_addBalance", prioCli.getNodeInfo().ID, tokens)
	prioCli.addPeer(lightServer)

	// Check if priority client is actually syncing and the regular client got kicked out
	prioCli.callRPC(&peers, "admin_peers")
	if len(peers) != 1 {
		t.Errorf("Expected: # of prio peers == 1, actual: %v", len(peers))
	}

	nodes := map[string]*gocorerpc{
		lightServer.getNodeInfo().ID: lightServer,
		freeCli.getNodeInfo().ID:     freeCli,
		prioCli.getNodeInfo().ID:     prioCli,
	}
	time.Sleep(8 * time.Second)
	lightServer.callRPC(&peers, "admin_peers")
	peersWithNames := make(map[string]string)
	for _, p := range peers {
		peersWithNames[nodes[p.ID].name] = p.ID
	}
	if _, freeClientFound := peersWithNames[freeCli.name]; freeClientFound {
		t.Error("client is still a peer of lightServer", peersWithNames)
	}
	if _, prioClientFound := peersWithNames[prioCli.name]; !prioClientFound {
		t.Error("prio client is not among lightServer peers", peersWithNames)
	}
}
