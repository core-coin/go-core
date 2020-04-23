// Copyright 2018 The go-core Authors
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

package dnsdisc

import (
	"context"
	ecdsa "github.com/core-coin/eddsa"
	"errors"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/core-coin/go-core/common/mclock"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/internal/testlog"
	"github.com/core-coin/go-core/log"
	"github.com/core-coin/go-core/p2p/enode"
	"github.com/core-coin/go-core/p2p/enr"
)

const (
	signingKeySeed = 0x111111
	nodesSeed1     = 0x2945237
	nodesSeed2     = 0x4567299
)

func TestClientSyncTree(t *testing.T) {
	r := mapResolver{
		"n": "enrtree-root:v1 e=HFQW3XGWDR2BR4JIOHK5TTAUALCZBOS2OCM7JAUGRGOIHDMQBUQA l=OPRDFOL6SMXNEZ6G5SIFSBY7VGQEXW4ARSGHPPGC2F2HTRWGUHIA seq=1 sig=ab4ORUpvvYww0a4tQjlPFzDGsYHpe-KCDtPd90UpFepgkbKePE7GVl-apanTF5epCknYPmWcJkm2C_uQwKuqbv0IXcOo1VJSOap3lxvtqyOpRce37HazGZSaQpmbEk7OpbMjGxMJeirfCy6jS5CiPWHZ_At5y_o8wmjIVjusNrbwkvfjZhTlkpaOzxNv_3D3JcY2hQaVz5N5YELlJDGH2hWzPCh7NMsI",
		"OPRDFOL6SMXNEZ6G5SIFSBY7VGQEXW4ARSGHPPGC2F2HTRWGUHIA.n": "enrtree://MJKQMAHXTDNOWVILXACK25P54GVTLUJHRXUEPWAYXCFOITYOMYUO7746ZHHWZYPEFXKNFBV4QDCC4S7POQMCBGVLJM@morenodes.example.org",
		"HFQW3XGWDR2BR4JIOHK5TTAUALCZBOS2OCM7JAUGRGOIHDMQBUQA.n": "enrtree-branch:QMJ7JFKGC5REHM2CD6VIWQHE2XM5VEH5NXP35VNP7BM6UIALTVJA,E7PBDEEWGHDQWTPEYJQGHG2IZMMLB2QGZWRQ74F5TCJQAIWOP55A,L7DSSOSZVTZX6NPZFM7HF6MEAKR7GTKYJPCEMBCSBA46YRJCRJBA",
		"QMJ7JFKGC5REHM2CD6VIWQHE2XM5VEH5NXP35VNP7BM6UIALTVJA.n": "enr:-PW4qLeDADEGnzGAHXRrsdk5QQBIGAZenfc6HgdlgdS2GyIhRfv5qeh6101CmYewviF_5lfKDbFQcS6BJpwBoEJXFKhJLZ-SSZldEMXf2yAsPOpi84joYs0EMmS41UfexN7en4rVhNwILq2k9exvAKZkEDuSRl1Q0FfSDeyQuB8cxR9ziMO_Hk5RvghjksNgIG8TksIaMYYRMmMgAmnHcoZMAcFVB_U3IoaWLwKCaWSCdjSJc2VjcDI1NmsxuDiSRl1Q0FfSDeyQuB8cxR9ziMO_Hk5RvghjksNgIG8TksIaMYYRMmMgAmnHcoZMAcFVB_U3IoaWLw",
		"E7PBDEEWGHDQWTPEYJQGHG2IZMMLB2QGZWRQ74F5TCJQAIWOP55A.n": "enr:-PW4qOXkhYyKDKlP75YQaygFo9H8nhQlUK94lgAq95sGUQcf5wI1gpq59A-c3duqO1IsTkub-moxM7z7TrKpuT_P-a1O9kUGZJhteLO2CcQlUIW6_p1hWqEGl5F__TT0szQRJW9SEK9qLTMgvZbgwj3YNwFPQlXjESdidAEy1c5JZyieIIX_MgMNwBmg4WHf_2cfl-ghWI9rdRj046RIoenZUGXQMv5J6VaA-ICCaWSCdjSJc2VjcDI1NmsxuDhPQlXjESdidAEy1c5JZyieIIX_MgMNwBmg4WHf_2cfl-ghWI9rdRj046RIoenZUGXQMv5J6VaA-A",
		"L7DSSOSZVTZX6NPZFM7HF6MEAKR7GTKYJPCEMBCSBA46YRJCRJBA.n": "enr:-PW4qGycmnYGPAoGx9AQFbWeroLx46HCSGjU4MJiwEeOH6PZb8GmS3qGKPvKr4W0dEElS7RmGv--ocky5L7REhEI6V4o3_zQ9Gr_l3vUmKcTNBMVmVHGsTVjzvTfJJqHlDAEeOWuV6pOMXviVK7brFRt9hyKz9eoXeUr-A16r771L4lr4x_SBDDTu-sXBYFzqzcB8Ef0exKieR2Li7FAqN_mysAR-6QqbNkwugGCaWSCdjSJc2VjcDI1NmsxuDiKz9eoXeUr-A16r771L4lr4x_SBDDTu-sXBYFzqzcB8Ef0exKieR2Li7FAqN_mysAR-6QqbNkwug",
	}
	var (
		wantNodes = make([]*enode.Node, 3)
		wantLinks = []string{"enrtree://MJKQMAHXTDNOWVILXACK25P54GVTLUJHRXUEPWAYXCFOITYOMYUO7746ZHHWZYPEFXKNFBV4QDCC4S7POQMCBGVLJM@morenodes.example.org"}
		wantSeq   = uint(1)
	)

	c := NewClient(Config{Resolver: r, Logger: testlog.Logger(t, log.LvlTrace)})
	stree, err := c.SyncTree("enrtree://MHM7YC3ZZP5DZQTIZBLDXLBWW3YJF57DMYKOLEUWR3HRG377OD3SLRRWQUDJLT4TPFQEFZJEGGD5UFNTHQUHWNGLBA@n")
	if err != nil {
		t.Fatal("sync error:", err)
	}
	if !reflect.DeepEqual(sortByID(stree.Nodes()), sortByID(wantNodes)) {
		t.Errorf("wrong nodes in synced tree:\nhave %v\nwant %v", spew.Sdump(stree.Nodes()), spew.Sdump(wantNodes))
	}
	if !reflect.DeepEqual(stree.Links(), wantLinks) {
		t.Errorf("wrong links in synced tree: %v", stree.Links())
	}
	if stree.Seq() != wantSeq {
		t.Errorf("synced tree has wrong seq: %d", stree.Seq())
	}
}

// In this test, syncing the tree fails because it contains an invalid ENR entry.
func TestClientSyncTreeBadNode(t *testing.T) {
	// var b strings.Builder
	// b.WriteString(enrPrefix)
	// b.WriteString("-----")
	// badHash := subdomain(&b)
	// tree, _ := MakeTree(3, nil, []string{"enrtree://AM5FCQLWIZX2QFPNJAP7VUERCCRNGRHWZG3YYHIUV7BVDQ5FDPRT2@morenodes.example.org"})
	// tree.entries[badHash] = &b
	// tree.root.eroot = badHash
	// url, _ := tree.Sign(testKey(signingKeySeed), "n")
	// fmt.Println(url)
	// fmt.Printf("%#v\n", tree.ToTXT("n"))

	r := mapResolver{
		"n":                            "enrtree-root:v1 e=INDMVBZEEQ4ESVYAKGIYU74EAA l=OPRDFOL6SMXNEZ6G5SIFSBY7VGQEXW4ARSGHPPGC2F2HTRWGUHIA seq=3 sig=VZdbTTZmkWqwk7uHbiCg1sv_-e8SEP_p7rZv19nm72nFbhsGEXckSRIe1ABarwarZ6cE6P9JAbJaOB96M4hlOkcGtkLsiMBtfC7kPXRAlIyVSTeyzosyKmAhI0QFX3WFGskoHecfqsPRhg40nmS6BGHZ_At5y_o8wmjIVjusNrbwkvfjZhTlkpaOzxNv_3D3JcY2hQaVz5N5YELlJDGH2hWzPCh7NMsI",
		"OPRDFOL6SMXNEZ6G5SIFSBY7VGQEXW4ARSGHPPGC2F2HTRWGUHIA.n": "enrtree://MJKQMAHXTDNOWVILXACK25P54GVTLUJHRXUEPWAYXCFOITYOMYUO7746ZHHWZYPEFXKNFBV4QDCC4S7POQMCBGVLJM@morenodes.example.org",
		"INDMVBZEEQ4ESVYAKGIYU74EAA.n": "enr:-----",
	}
	c := NewClient(Config{Resolver: r, Logger: testlog.Logger(t, log.LvlTrace)})
	_, err := c.SyncTree("enrtree://MHM7YC3ZZP5DZQTIZBLDXLBWW3YJF57DMYKOLEUWR3HRG377OD3SLRRWQUDJLT4TPFQEFZJEGGD5UFNTHQUHWNGLBA@n")
	wantErr := nameError{name: "INDMVBZEEQ4ESVYAKGIYU74EAA.n", err: entryError{typ: "enr", err: errInvalidENR}}
	if err != wantErr {
		t.Fatalf("expected sync error %q, got %q", wantErr, err)
	}
}

// This test checks that randomIterator finds all entries.
func TestIterator(t *testing.T) {
	nodes := testNodes(nodesSeed1, 30)
	tree, url := makeTestTree("n", nodes, nil)
	r := mapResolver(tree.ToTXT("n"))
	c := NewClient(Config{
		Resolver:  r,
		Logger:    testlog.Logger(t, log.LvlTrace),
		RateLimit: 500,
	})
	it, err := c.NewIterator(url)
	if err != nil {
		t.Fatal(err)
	}

	checkIterator(t, it, nodes)
}

// This test checks if closing randomIterator races.
func TestIteratorClose(t *testing.T) {
	nodes := testNodes(nodesSeed1, 500)
	tree1, url1 := makeTestTree("t1", nodes, nil)
	c := NewClient(Config{Resolver: newMapResolver(tree1.ToTXT("t1"))})
	it, err := c.NewIterator(url1)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		for it.Next() {
			_ = it.Node()
		}
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	it.Close()
	<-done
}

// This test checks that randomIterator traverses linked trees as well as explicitly added trees.
func TestIteratorLinks(t *testing.T) {
	nodes := testNodes(nodesSeed1, 40)
	tree1, url1 := makeTestTree("t1", nodes[:10], nil)
	tree2, url2 := makeTestTree("t2", nodes[10:], []string{url1})
	c := NewClient(Config{
		Resolver:  newMapResolver(tree1.ToTXT("t1"), tree2.ToTXT("t2")),
		Logger:    testlog.Logger(t, log.LvlTrace),
		RateLimit: 5,
	})
	it, err := c.NewIterator(url2)
	if err != nil {
		t.Fatal(err)
	}

	checkIterator(t, it, nodes)
}

// This test verifies that randomIterator re-checks the root of the tree to catch
// updates to nodes.
func TestIteratorNodeUpdates(t *testing.T) {
	var (
		clock    = new(mclock.Simulated)
		nodes    = testNodes(nodesSeed1, 30)
		resolver = newMapResolver()
		c        = NewClient(Config{
			Resolver:        resolver,
			Logger:          testlog.Logger(t, log.LvlTrace),
			RecheckInterval: 1 * time.Minute,
			RateLimit:       5,
		})
	)
	c.clock = clock
	tree1, url := makeTestTree("n", nodes[:25], nil)
	it, err := c.NewIterator(url)
	if err != nil {
		t.Fatal(err)
	}

	// Sync the original tree.
	resolver.add(tree1.ToTXT("n"))
	checkIterator(t, it, nodes[:25])

	// Ensure RandomNode returns the new nodes after the tree is updated.
	updateSomeNodes(nodesSeed1, nodes)
	tree2, _ := makeTestTree("n", nodes, nil)
	resolver.clear()
	resolver.add(tree2.ToTXT("n"))
	t.Log("tree updated")

	clock.Run(c.cfg.RecheckInterval + 1*time.Second)
	checkIterator(t, it, nodes)
}

// This test checks that the tree root is rechecked when a couple of leaf
// requests have failed. The test is just like TestIteratorNodeUpdates, but
// without advancing the clock by recheckInterval after the tree update.
func TestIteratorRootRecheckOnFail(t *testing.T) {
	var (
		clock    = new(mclock.Simulated)
		nodes    = testNodes(nodesSeed1, 30)
		resolver = newMapResolver()
		c        = NewClient(Config{
			Resolver:        resolver,
			Logger:          testlog.Logger(t, log.LvlTrace),
			RecheckInterval: 20 * time.Minute,
			RateLimit:       500,
			// Disabling the cache is required for this test because the client doesn't
			// notice leaf failures if all records are cached.
			CacheLimit: 1,
		})
	)
	c.clock = clock
	tree1, url := makeTestTree("n", nodes[:25], nil)
	it, err := c.NewIterator(url)
	if err != nil {
		t.Fatal(err)
	}

	// Sync the original tree.
	resolver.add(tree1.ToTXT("n"))
	checkIterator(t, it, nodes[:25])

	// Ensure RandomNode returns the new nodes after the tree is updated.
	updateSomeNodes(nodesSeed1, nodes)
	tree2, _ := makeTestTree("n", nodes, nil)
	resolver.clear()
	resolver.add(tree2.ToTXT("n"))
	t.Log("tree updated")

	checkIterator(t, it, nodes)
}

// updateSomeNodes applies ENR updates to some of the given nodes.
func updateSomeNodes(keySeed int64, nodes []*enode.Node) {
	keys := testKeys(nodesSeed1, len(nodes))
	for i, n := range nodes[:len(nodes)/2] {
		r := n.Record()
		r.Set(enr.IP{127, 0, 0, 1})
		r.SetSeq(55)
		enode.SignV4(r, keys[i])
		n2, _ := enode.New(enode.ValidSchemes, r)
		nodes[i] = n2
	}
}

// This test verifies that randomIterator re-checks the root of the tree to catch
// updates to links.
func TestIteratorLinkUpdates(t *testing.T) {
	var (
		clock    = new(mclock.Simulated)
		nodes    = testNodes(nodesSeed1, 30)
		resolver = newMapResolver()
		c        = NewClient(Config{
			Resolver:        resolver,
			Logger:          testlog.Logger(t, log.LvlTrace),
			RecheckInterval: 20 * time.Minute,
			RateLimit:       500,
		})
	)
	c.clock = clock
	tree3, url3 := makeTestTree("t3", nodes[20:30], nil)
	tree2, url2 := makeTestTree("t2", nodes[10:20], nil)
	tree1, url1 := makeTestTree("t1", nodes[0:10], []string{url2})
	resolver.add(tree1.ToTXT("t1"))
	resolver.add(tree2.ToTXT("t2"))
	resolver.add(tree3.ToTXT("t3"))

	it, err := c.NewIterator(url1)
	if err != nil {
		t.Fatal(err)
	}

	// Sync tree1 using RandomNode.
	checkIterator(t, it, nodes[:20])

	// Add link to tree3, remove link to tree2.
	tree1, _ = makeTestTree("t1", nodes[:10], []string{url3})
	resolver.add(tree1.ToTXT("t1"))
	t.Log("tree1 updated")

	clock.Run(c.cfg.RecheckInterval + 1*time.Second)

	var wantNodes []*enode.Node
	wantNodes = append(wantNodes, tree1.Nodes()...)
	wantNodes = append(wantNodes, tree3.Nodes()...)
	checkIterator(t, it, wantNodes)

	// Check that linked trees are GCed when they're no longer referenced.
	knownTrees := it.(*randomIterator).trees
	if len(knownTrees) != 2 {
		t.Errorf("client knows %d trees, want 2", len(knownTrees))
	}
}

func checkIterator(t *testing.T, it enode.Iterator, wantNodes []*enode.Node) {
	t.Helper()

	var (
		want     = make(map[enode.ID]*enode.Node)
		maxCalls = len(wantNodes) * 3
		calls    = 0
	)
	for _, n := range wantNodes {
		want[n.ID()] = n
	}
	for ; len(want) > 0 && calls < maxCalls; calls++ {
		if !it.Next() {
			t.Fatalf("Next returned false (call %d)", calls)
		}
		n := it.Node()
		delete(want, n.ID())
	}
	t.Logf("checkIterator called Next %d times to find %d nodes", calls, len(wantNodes))
	for _, n := range want {
		t.Errorf("iterator didn't discover node %v", n.ID())
	}
}

func makeTestTree(domain string, nodes []*enode.Node, links []string) (*Tree, string) {
	tree, err := MakeTree(1, nodes, links)
	if err != nil {
		panic(err)
	}
	url, err := tree.Sign(testKey(signingKeySeed), domain)
	if err != nil {
		panic(err)
	}
	return tree, url
}

// testKeys creates deterministic private keys for testing.
func testKeys(seed int64, n int) []*ecdsa.PrivateKey {
	randSeed := rand.New(rand.NewSource(seed))
	keys := make([]*ecdsa.PrivateKey, n)
	for i := 0; i < n; i++ {
		key, err := crypto.GenerateKey(randSeed)
		if err != nil {
			panic("can't generate key: " + err.Error())
		}
		keys[i] = key
	}
	return keys
}

func testKey(seed int64) *ecdsa.PrivateKey {
	return testKeys(seed, 1)[0]
}

func testNodes(seed int64, n int) []*enode.Node {
	keys := testKeys(seed, n)
	nodes := make([]*enode.Node, n)
	for i, key := range keys {
		record := new(enr.Record)
		record.SetSeq(uint64(i))
		enode.SignV4(record, key)
		n, err := enode.New(enode.ValidSchemes, record)
		if err != nil {
			panic(err)
		}
		nodes[i] = n
	}
	return nodes
}

func testNode(seed int64) *enode.Node {
	return testNodes(seed, 1)[0]
}

type mapResolver map[string]string

func newMapResolver(maps ...map[string]string) mapResolver {
	mr := make(mapResolver)
	for _, m := range maps {
		mr.add(m)
	}
	return mr
}

func (mr mapResolver) clear() {
	for k := range mr {
		delete(mr, k)
	}
}

func (mr mapResolver) add(m map[string]string) {
	for k, v := range m {
		mr[k] = v
	}
}

func (mr mapResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	if record, ok := mr[name]; ok {
		return []string{record}, nil
	}
	return nil, errors.New("not found")
}
