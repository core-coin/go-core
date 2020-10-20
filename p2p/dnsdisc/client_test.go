// Copyright 2019 by the Authors
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
	"errors"
	"github.com/core-coin/eddsa"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/core-coin/go-core/v2/common/mclock"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/internal/testlog"
	"github.com/core-coin/go-core/v2/log"
	"github.com/core-coin/go-core/v2/p2p/enode"
	"github.com/core-coin/go-core/v2/p2p/enr"
	"github.com/davecgh/go-spew/spew"
)

const (
	signingKeySeed = 0x111111
	nodesSeed1     = 0x2945237
	nodesSeed2     = 0x4567299
)

func TestClientSyncTree(t *testing.T) {
	r := mapResolver{
		"n": "enrtree-root:v1 e=COL2I4YN3X3OQQOASJ4HCDP57JOC3RT55XXWTQ6QBLI3TQILUOFA l=MQXVTT7AWMPCQZLLCF6OI4PWOBYEBODNT324O2DO2FVD7TWKGSMA seq=1 sig=n15FzHN6dts6sXL1S9CYb2iRWs2L4eWXqIJRIateo0IBtDX5FLOoEqdBNVeBUyohFOFwyhlFBBh9VrOArOVubbYYcKYHeSWTH5RqvQPvNF2WkC_TV3I1G7WsDDYsI0NLXDfk22ZdKOoSi7Ap_5GbGGBg4gCAhk-5QLDFuJX-jqweDuRq7etFByjtGPycI39kBHufkjBW46v0tB3qWsI07Cs3my_or6Fs",
		"MQXVTT7AWMPCQZLLCF6OI4PWOBYEBODNT324O2DO2FVD7TWKGSMA.n": "enrtree://MJKQMAHXTDNOWVILXACK25P54GVTLUJHRXUEPWAYXCFOITYOMYUO7746ZHHWZYPEFXKNFBV4QDCC4S7POQMCBGVLJM@morenodes.example.org",
		"COL2I4YN3X3OQQOASJ4HCDP57JOC3RT55XXWTQ6QBLI3TQILUOFA.n": "enrtree-branch:KC4PLFMOX2LHMRJFOXFGW2HQ6DMKPEUBRE4WCFPWDNLST54TZAKA,7MY4MKNM3NJRQAB4FMQ3KYLPWHUL7NGPOMN7J4RAY3NZ2PSTMDBQ,VW3Q4GPAEDH6DLOYX4GQYUSFHKFHBUC4TRH2PZTH5SBAAQNE7P2A",
		"KC4PLFMOX2LHMRJFOXFGW2HQ6DMKPEUBRE4WCFPWDNLST54TZAKA.n": "enr:-PW4qDy6Kc3wPRjcwiNsIHmqnTzlOWs0gpOp0W2ekx6Q0pbaMrZP6KRKQkdMhXqgtj5YL_Tw3qMcB4r_Iwdl5PAVGrbn7UR0uTLvUttJh-80kg4h1zNv7ItUg1bKat2eMtYcsWWxc9lZRHTdPsoek6_1TgaaamYHL_OyITW9e9ZTdgkHXtR2veUQlvyitbRQlz5vHtZm8aS_yDrq_ogFuoq7SIRVFYs2S1u9woCCaWSCdjSJc2VjcDI1NmsxuDiaamYHL_OyITW9e9ZTdgkHXtR2veUQlvyitbRQlz5vHtZm8aS_yDrq_ogFuoq7SIRVFYs2S1u9wg",
		"7MY4MKNM3NJRQAB4FMQ3KYLPWHUL7NGPOMN7J4RAY3NZ2PSTMDBQ.n": "enr:-PW4qKQpUaY7UrlvoVpCOUZj8lptTPY9QUWueGKOjN5pYnJVdDloYZfsUtmszTRPXQtVkUG6GbZH1kFEAt0fDSizgzL4OziqPtJWWfoQwj4QSYSDNjjwsV0uxeYr4DiViWD6zaToGw6sUZEDNJxAXXHFcy4ks9qDJBkuBBOLuNImIkAsjXsEjtqHBIhEGGbaplqRMXQooVX3XHw2R3AYic0QJn_nQHrwWxSLGAGCaWSCdjSJc2VjcDI1NmsxuDgks9qDJBkuBBOLuNImIkAsjXsEjtqHBIhEGGbaplqRMXQooVX3XHw2R3AYic0QJn_nQHrwWxSLGA",
		"VW3Q4GPAEDH6DLOYX4GQYUSFHKFHBUC4TRH2PZTH5SBAAQNE7P2A.n": "enr:-PW4qMw2cm6MCY8ze_z9V6LP3DolIEiD08XAppFD3jf-GfSOKvs8lo1Hk0SQ0zomoRuhXMBLxPqWXhsGCWgSdG-8cD9mMqjRxmKtuW7LwqmOa_BhzQEBHlfX6sIPLHEQ6i4c15sn2n5NWQZ2yHYf9geCMx5TllePNBLFhpJeOasP8Uv90XspWoW40v4X3u_sxtFNZQDTbUlXIIMO5_jkgtyzpHU69uHGttB6EAKCaWSCdjSJc2VjcDI1NmsxuDhTllePNBLFhpJeOasP8Uv90XspWoW40v4X3u_sxtFNZQDTbUlXIIMO5_jkgtyzpHU69uHGttB6EA",
	}
	var (
		wantNodes = testNodes(0x29452, 3)
		wantLinks = []string{"enrtree://MJKQMAHXTDNOWVILXACK25P54GVTLUJHRXUEPWAYXCFOITYOMYUO7746ZHHWZYPEFXKNFBV4QDCC4S7POQMCBGVLJM@morenodes.example.org"}
		wantSeq   = uint(1)
	)

	c := NewClient(Config{Resolver: r, Logger: testlog.Logger(t, log.LvlTrace)})
	stree, err := c.SyncTree("enrtree://MBQOEAEAQZH3SQFQYW4JL7UOVQPA5ZDK5XVUKBZI5UMPZHBDP5SAI647SIYFNY5L6S2B32S2YI2OYKZXTMX6RL5BNQ@n")
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
		"n": "enrtree-root:v1 e=JZ7RET254NLCICBLV5AZ3WVUHCXBPW5S7VM7ZWYRIAHGVFIWOVM l=MQXVTT7AWMPCQZLLCF6OI4PWOBYEBODNT324O2DO2FVD7TWKGSMA seq=3 sig=XEC_y0aPgNujlkCCwMGJhXeNZuN2f3t4NYSJWN_ETtvSfJ1IMxbsos60bviaI2Xln0Gl3laXvWHfM8Bcn1sYgAFTYVlBRXvJlwh7TEcBDOZnfd_bmlAU6mAq83fKYbLZoDZ6btqpVXwr3Re-IwNBLGBg4gCAhk-5QLDFuJX-jqweDuRq7etFByjtGPycI39kBHufkjBW46v0tB3qWsI07Cs3my_or6Fs",
		"MQXVTT7AWMPCQZLLCF6OI4PWOBYEBODNT324O2DO2FVD7TWKGSMA.n": "enrtree://MJKQMAHXTDNOWVILXACK25P54GVTLUJHRXUEPWAYXCFOITYOMYUO7746ZHHWZYPEFXKNFBV4QDCC4S7POQMCBGVLJM@morenodes.example.org",
		"JZ7RET254NLCICBLV5AZ3WVUHCXBPW5S7VM7ZWYRIAHGVFIWOVM.n":  "enr:-----",
	}
	c := NewClient(Config{Resolver: r, Logger: testlog.Logger(t, log.LvlTrace)})
	_, err := c.SyncTree("enrtree://MBQOEAEAQZH3SQFQYW4JL7UOVQPA5ZDK5XVUKBZI5UMPZHBDP5SAI647SIYFNY5L6S2B32S2YI2OYKZXTMX6RL5BNQ@n")
	wantErr := nameError{name: "JZ7RET254NLCICBLV5AZ3WVUHCXBPW5S7VM7ZWYRIAHGVFIWOVM.n", err: entryError{typ: "enr", err: errInvalidENR}}
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
func testKeys(seed int64, n int) []*eddsa.PrivateKey {
	randSeed := rand.New(rand.NewSource(seed))
	keys := make([]*eddsa.PrivateKey, n)
	for i := 0; i < n; i++ {
		key, err := crypto.GenerateKey(randSeed)
		if err != nil {
			panic("can't generate key: " + err.Error())
		}
		keys[i] = key
	}
	return keys
}

func testKey(seed int64) *eddsa.PrivateKey {
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
