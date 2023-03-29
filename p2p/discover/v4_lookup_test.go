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

package discover

import (
	"fmt"
	"net"
	"sort"
	"testing"

	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/p2p/discover/v4wire"
	"github.com/core-coin/go-core/v2/p2p/enode"
	"github.com/core-coin/go-core/v2/p2p/enr"
)

func TestUDPv4_Lookup(t *testing.T) {
	t.Parallel()
	test := newUDPTest(t)

	// Lookup on empty table returns no nodes.
	targetKey, _ := decodePubkey(lookupTestnet.target[:])
	if results := test.udp.LookupPubkey(targetKey); len(results) > 0 {
		t.Fatalf("lookup on empty table returned %d results: %#v", len(results), results)
	}

	// Seed table with initial node.
	fillTable(test.table, []*node{wrapNode(lookupTestnet.node(256, 0))})

	// Start the lookup.
	resultC := make(chan []*enode.Node, 1)
	go func() {
		resultC <- test.udp.LookupPubkey(targetKey)
		test.close()
	}()

	// Answer lookup packets.
	serveTestnet(test, lookupTestnet)

	// Verify result nodes.
	results := <-resultC
	t.Logf("results:")
	for _, e := range results {
		t.Logf("  ld=%d, %x", enode.LogDist(lookupTestnet.target.id(), e.ID()), e.ID().Bytes())
	}
	if len(results) != bucketSize {
		t.Errorf("wrong number of results: got %d, want %d", len(results), bucketSize)
	}
	checkLookupResults(t, lookupTestnet, results)
}

func TestUDPv4_LookupIterator(t *testing.T) {
	t.Parallel()
	test := newUDPTest(t)
	defer test.close()

	// Seed table with initial nodes.
	bootnodes := make([]*node, len(lookupTestnet.dists[256]))
	for i := range lookupTestnet.dists[256] {
		bootnodes[i] = wrapNode(lookupTestnet.node(256, i))
	}
	fillTable(test.table, bootnodes)
	go serveTestnet(test, lookupTestnet)

	// Create the iterator and collect the nodes it yields.
	iter := test.udp.RandomNodes()
	seen := make(map[enode.ID]*enode.Node)
	for limit := lookupTestnet.len(); iter.Next() && len(seen) < limit; {
		seen[iter.Node().ID()] = iter.Node()
	}
	iter.Close()

	// Check that all nodes in lookupTestnet were seen by the iterator.
	results := make([]*enode.Node, 0, len(seen))
	for _, n := range seen {
		results = append(results, n)
	}
	sortByID(results)
	want := lookupTestnet.nodes()
	if err := checkNodesEqual(results, want); err != nil {
		t.Fatal(err)
	}
}

// TestUDPv4_LookupIteratorClose checks that lookupIterator ends when its Close
// method is called.
func TestUDPv4_LookupIteratorClose(t *testing.T) {
	t.Parallel()
	test := newUDPTest(t)
	defer test.close()

	// Seed table with initial nodes.
	bootnodes := make([]*node, len(lookupTestnet.dists[256]))
	for i := range lookupTestnet.dists[256] {
		bootnodes[i] = wrapNode(lookupTestnet.node(256, i))
	}
	fillTable(test.table, bootnodes)
	go serveTestnet(test, lookupTestnet)

	it := test.udp.RandomNodes()
	if ok := it.Next(); !ok || it.Node() == nil {
		t.Fatalf("iterator didn't return any node")
	}

	it.Close()

	ncalls := 0
	for ; ncalls < 100 && it.Next(); ncalls++ {
		if it.Node() == nil {
			t.Error("iterator returned Node() == nil node after Next() == true")
		}
	}
	t.Logf("iterator returned %d nodes after close", ncalls)
	if it.Next() {
		t.Errorf("Next() == true after close and %d more calls", ncalls)
	}
	if n := it.Node(); n != nil {
		t.Errorf("iterator returned non-nil node after close and %d more calls", ncalls)
	}
}

func serveTestnet(test *udpTest, testnet *preminedTestnet) {
	for done := false; !done; {
		done = test.waitPacketOut(func(p v4wire.Packet, to *net.UDPAddr, hash []byte) {
			n, key := testnet.nodeByAddr(to)
			switch p.(type) {
			case *v4wire.Ping:
				test.packetInFrom(nil, key, to, &v4wire.Pong{Expiration: futureExp, ReplyTok: hash})
			case *v4wire.Findnode:
				dist := enode.LogDist(n.ID(), testnet.target.id())
				nodes := testnet.nodesAtDistance(dist - 1)
				test.packetInFrom(nil, key, to, &v4wire.Neighbors{Expiration: futureExp, Nodes: nodes})
			}
		})
	}
}

// checkLookupResults verifies that the results of a lookup are the closest nodes to
// the testnet's target.
func checkLookupResults(t *testing.T, tn *preminedTestnet, results []*enode.Node) {
	t.Helper()
	t.Logf("results:")
	for _, e := range results {
		t.Logf("  ld=%d, %x", enode.LogDist(tn.target.id(), e.ID()), e.ID().Bytes())
	}
	if hasDuplicates(wrapNodes(results)) {
		t.Errorf("result set contains duplicate entries")
	}
	if !sortedByDistanceTo(tn.target.id(), wrapNodes(results)) {
		t.Errorf("result set not sorted by distance to target")
	}
	wantNodes := tn.closest(len(results))
	if err := checkNodesEqual(results, wantNodes); err != nil {
		t.Error(err)
	}
}

// This is the test network for the Lookup test.
// The nodes were obtained by running lookupTestnet.mine with a random NodeID as target.
var lookupTestnet = &preminedTestnet{
	target: hexEncPubkey("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363991"),
	dists: [257][]*crypto.PrivateKey{
		252: {
			hexEncPrivkey("bebf794fd362a6792e3d7067f3de2617a1d5e88476665c446ec8e935411c2af0a118f9daa577d29ca9354d40929d074368347d13bba4260011"),
			hexEncPrivkey("d66e4f8a6710394d176f7d8c5c7d2d670f46604f08dd087c7c3209826f7632f2150555c66263c5a10468696b754d3197a12f45be589bce8f47"),
			hexEncPrivkey("fd3a176f84fea182cd8f26a0266320a078dfb4ed301bf934968cfedbd381046b01c985c11c406da95fbdbd6935cff506c44dc9c946ee994922"),
			hexEncPrivkey("6c6f07be1b0a5f5c9f7ccc0baafa2a7bec1613edd65ca626e1ca9e3c083bf5be9a4250385cae574b48bd639328751ee4edb63bb8f34c6c4260"),
			hexEncPrivkey("51e312adda9f0e79b10230ebe2291f883516cf58ae1484fd5114622aaf55df51da4e4a0436679a99175ee9bbe9a10d74bcc181faaa97aa663f"),
			hexEncPrivkey("827cc2fad638bff047eac8fe72a8014772ddb1afbf7ead026c9bba79ae37fcf435d299885656bf8a7188c9237f0760cb20480b4274945f8c38"),
			hexEncPrivkey("a6b70ebf683daf2a45d35b8f11b3ff4b862a6ddbcb0abd8e84a143223c30fc8f189a29e89c1879797773f36a27f80b1d1dde1789c9228af67c"),
			hexEncPrivkey("27cf27280654c003d01cd9bef5ce6229076b47b3d8cdc8961fc51aa01e81930c2c3c874f3f41b4e2075f09227ef5ea885bf74d6a23b18b485d"),
		},
		253: {
			hexEncPrivkey("4081b41fcdcd6b37516600587e988f4163fdc62b3be481fb59484be0f64d7f2b6c61ac0c1a3a08276d1aeff4d730fca53f880865868b44070d"),
			hexEncPrivkey("a9ffa5a151b0c2c3322a87a996ca53aee1913585b6f8da1536a8159c6ff8181d6151a0b3a94bd91c98d2bcc6e2e955495696be3c0ac677bd63"),
			hexEncPrivkey("4e17d8213d936d55ecf74dac615cdb1c8c027b8ed6c17d1be8d218f9e6a537c7198a80db4d5cf1bbdedb2db390859d7506a9285b5f027ae87c"),
			hexEncPrivkey("13a04805c5520d94a2ba0d1ef1a569d07dcc4c755581facb78b9bb978e1cd1bdc1063c0b21ac60cab6468a4cede008fe68db06397daafd4709"),
			hexEncPrivkey("9aa7545a4a02d674a75e7bd28cd05c40fb741b6afe8a3f413ba9e0fcf45244baa9e86ef66e49056a2bdbda78cc833a105f5d37c77b79072426"),
			hexEncPrivkey("1fe288abdde49133b1b326248364cce118e5c65315497232f2dde6fe7e699d362b31128667de573038285eb9db1af8132fa0805a08083f5d0a"),
			hexEncPrivkey("8644b2e1bc1156ddd45cc1baba9277177ea395f12255280e1ca17e8023704a1aed46e65c166e190f608ba36bb8a4e0b98536d72b0f075af902"),
			hexEncPrivkey("56ebbea0e8e6359ec7346c8b4788305be98a52d3f1a51718a426616df4d9706b579db36287d1a5c7dba9d06aee91f0dd46f401afb7760bf15e"),
		},
		254: {
			hexEncPrivkey("1f41987fb42a7f2bae3ce3f8e307978d284bb97187055e7188399b0a8ae11f94b77a5996b0f0cdcfb23e2cb73ea3304e0b42d659fb5c0e467e"),
			hexEncPrivkey("3db577f0081c659d29a50a9dcbb0fbeafeb7c6802f6c2133dd46b2509e7815609586d12150290d32e9c0d0d0281f1f3a82187dca72826e510a"),
			hexEncPrivkey("99a1e743ecec28c7f6113260d9c7b1caba3f7af1cbef356c022b407d4bda91b37c83fdc27ba3b0754355dc868dcb6d764346830e5250737f7e"),
			hexEncPrivkey("54e2df17fc6f6f587e205a694424b773fac19423ea80d7d8c89a0dbb3e5a007fea913e93cc6cd348e2e301aaa3ea32b55b9bd0e95d44065669"),
			hexEncPrivkey("e95b80812a230beb081880619c1b87e3d066b57594308edc33b26cb70a433f6195e236808ddb7999af61c401d912557a63495868e60397717e"),
			hexEncPrivkey("c28a0283131c5e86f67089ce609f67981105a149425670898019c0cb62af39bdf7a29a1f64dace8858b304c4560e612718d50d1d8ed523b21a"),
			hexEncPrivkey("0428d8bfbfa899c9ac4a200cadc557965dbf339c632582ef32a77621d8313316cf377203bc61793f2ce557ff3c23db77b80d44b26774807b0b"),
			hexEncPrivkey("4f12886917556baa4d946a4957a64860c9a7411a69fd092c4a7f2e5f24edd5e94f2775565506dd6c6f9fc10be41980d0e8b78380ac7d086120"),
		},
		255: {
			hexEncPrivkey("2fc8bd43bfd3135ad641a9cd210051e99e02cefd36035b7aa1b5719ff0a2d08dccf3b3eebfc9da9637f6cf6d9adb1abff30e47137f249f6527"),
			hexEncPrivkey("24c32a500364ebc9d788d2f835dcc3eb5a366618252211a63ac4b34c79f37c575d9f4af0c0be39b1d6da2f129b4be3fba2f2bea40a1b718046"),
			hexEncPrivkey("8a49050a9a2bbf2770bb81dc8d24d441c872a2b3d705c36ec9e790bd9ba901e466c0fc928a8aeb5c6a5855413a4d4353c23111a599de5daf16"),
			hexEncPrivkey("eec97595492c8dd546d547ec5e89f88995f88de7502c3e8bab9f12c69beeabe4a7118d71bede4d0585ae626dabad34a0a8b3ef250cf952d311"),
			hexEncPrivkey("5dfab09f5711258141f0911162e9b6e45058444dbf14fb30b3b7cf8008060ff9af4c8a204a2b1775328db0a6893ff22691da8875b5ca600b01"),
			hexEncPrivkey("c0c59cd75ef165fe9efeb18d1174a501d20be2807dad3f78b154cc6da674c3ccb0f7eb86e2478eefb1c47de93edc39ad714acc58a6bfad8272"),
			hexEncPrivkey("e8c0bbbff285d0c52174a7d3369b322bc32fc3c9c4f1f4076dd4ff7f8cbab3578654fb2cad1b4b73df7e274be32ceff8e364e6c55542bc2550"),
			hexEncPrivkey("dc1415311d36f9d679d9420e47573a60b32378743f95a103f9601bc2009eaa937384cd759338fa69875043c1b98d4c0cedc7de63872e3c4803"),
		},
		256: {
			hexEncPrivkey("e8be3ec0aeaead80a4083359c8a701b6763a057c1d404e23c9fe126a2e1a82c91164b2d1c4212109e0d3e282f07ffbc55f0681006ac68dfd23"),
			hexEncPrivkey("b8c107152f99567ef59a79172812b98dbd18f4f0ea2742461dcabf5a6b68e2248f72186510edb72e456456ed7d581287ddcb49cec944f19e16"),
			hexEncPrivkey("a6755841a42a0335e548ef0669f660aae85795883671f6cc6aedbda7b99e13c0da91a73d04aedf735870a016eafac8d47c32c6fdc20db0f12c"),
			hexEncPrivkey("738dafc9e041a91195a06bbc6605f6feb7d885c0e1637d3bc84144ea5c03e2a0397fe3b7b7d1487751b4e4f861bfed8cbee1100c2dc4ee5b2f"),
			hexEncPrivkey("4835941961ecbe18e3409cd8321ddb4c8af298311ad4264e48d8a792e1bc3732569544fa190893fcff5a3706a18334f64e9e15474f92f0dc30"),
			hexEncPrivkey("556e3483f534b5f142522bcbf7678a3897be6b1d40127938c043e6eb692580cc9152e5f6cae0ab1c5a9f762beccb557165e164df62ced6e94d"),
			hexEncPrivkey("477f69b000be223c5405a042345b5e95565dd4363d854392a4e27f394d2ba903ba5ef0e3b758640be5144f137929d204d953f798b85a38d955"),
			hexEncPrivkey("76bcdf6cba61484498cc797fb051c867dc32fb12f80fbc1cacc439d810ce0bdd84372a72675e36202ba9eca3dd8e4ee7d7c66d51392b63eb1b"),
		},
	},
}

type preminedTestnet struct {
	target encPubkey
	dists  [hashBits + 1][]*crypto.PrivateKey
}

func (tn *preminedTestnet) len() int {
	n := 0
	for _, keys := range tn.dists {
		n += len(keys)
	}
	return n
}

func (tn *preminedTestnet) nodes() []*enode.Node {
	result := make([]*enode.Node, 0, tn.len())
	for dist, keys := range tn.dists {
		for index := range keys {
			result = append(result, tn.node(dist, index))
		}
	}
	sortByID(result)
	return result
}

func (tn *preminedTestnet) node(dist, index int) *enode.Node {
	key := tn.dists[dist][index]
	rec := new(enr.Record)
	rec.Set(enr.IP{127, byte(dist >> 8), byte(dist), byte(index)})
	rec.Set(enr.UDP(5000))
	err := enode.SignV4(rec, key)
	if err != nil {
		panic(err)
	}
	n, err := enode.New(enode.ValidSchemes, rec)
	if err != nil {
		panic(err)
	}
	return n
}

func (tn *preminedTestnet) nodeByAddr(addr *net.UDPAddr) (*enode.Node, *crypto.PrivateKey) {
	dist := int(addr.IP[1])<<8 + int(addr.IP[2])
	index := int(addr.IP[3])
	key := tn.dists[dist][index]
	return tn.node(dist, index), key
}

func (tn *preminedTestnet) nodesAtDistance(dist int) []v4wire.Node {
	result := make([]v4wire.Node, len(tn.dists[dist]))
	for i := range result {
		result[i] = nodeToRPC(wrapNode(tn.node(dist, i)))
	}
	return result
}

func (tn *preminedTestnet) neighborsAtDistances(base *enode.Node, distances []uint, elems int) []*enode.Node {
	var result []*enode.Node
	for d := range lookupTestnet.dists {
		for i := range lookupTestnet.dists[d] {
			n := lookupTestnet.node(d, i)
			d := enode.LogDist(base.ID(), n.ID())
			if containsUint(uint(d), distances) {
				result = append(result, n)
				if len(result) >= elems {
					return result
				}
			}
		}
	}
	return result
}

func (tn *preminedTestnet) closest(n int) (nodes []*enode.Node) {
	for d := range tn.dists {
		for i := range tn.dists[d] {
			nodes = append(nodes, tn.node(d, i))
		}
	}
	sort.Slice(nodes, func(i, j int) bool {
		return enode.DistCmp(tn.target.id(), nodes[i].ID(), nodes[j].ID()) < 0
	})
	return nodes[:n]
}

var _ = (*preminedTestnet).mine // avoid linter warning about mine being dead code.

// mine generates a testnet struct literal with nodes at
// various distances to the network's target.
func (tn *preminedTestnet) mine() {
	// Clear existing slices first (useful when re-mining).
	for i := range tn.dists {
		tn.dists[i] = nil
	}

	targetSha := tn.target.id()
	found, need := 0, 40
	for found < need {
		k := newkey()
		ld := enode.LogDist(targetSha, encodePubkey(k.PublicKey()).id())
		if ld > 251 {
			if len(tn.dists[ld]) < 8 {
				tn.dists[ld] = append(tn.dists[ld], k)
				found++
				fmt.Printf("found ID with ld %d (%d/%d)\n", ld, found, need)
			}
		}
	}
	fmt.Printf("&preminedTestnet{\n")
	fmt.Printf("	target: hexEncPubkey(\"%x\"),\n", tn.target[:])
	fmt.Printf("	dists: [%d][]*crypto.PrivateKey{\n", len(tn.dists))
	for ld, ns := range tn.dists {
		if len(ns) == 0 {
			continue
		}
		fmt.Printf("		%d: {\n", ld)
		for _, key := range ns {
			fmt.Printf("			hexEncPrivkey(\"%x\"),\n", key.PrivateKey())
		}
		fmt.Printf("		},\n")
	}
	fmt.Printf("	},\n")
	fmt.Printf("}\n")
}
