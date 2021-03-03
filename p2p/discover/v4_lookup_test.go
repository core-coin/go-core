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
	"github.com/core-coin/go-core/common"
	eddsa "github.com/core-coin/go-goldilocks"
	"net"
	"sort"
	"testing"

	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/p2p/enode"
	"github.com/core-coin/go-core/p2p/enr"
)

func TestUDPv4_Lookup(t *testing.T) {
	t.Parallel()
	test := newUDPTest(t)

	// Lookup on empty table returns no nodes.
	targetKey, _ := decodePubkey(lookupDevin.target)
	if results := test.udp.LookupPubkey(targetKey); len(results) > 0 {
		t.Fatalf("lookup on empty table returned %d results: %#v", len(results), results)
	}

	// Seed table with initial node.
	fillTable(test.table, []*node{wrapNode(lookupDevin.node(256, 0))})

	// Start the lookup.
	resultC := make(chan []*enode.Node, 1)
	go func() {
		resultC <- test.udp.LookupPubkey(targetKey)
		test.close()
	}()

	// Answer lookup packets.
	serveDevin(test, lookupDevin)

	// Verify result nodes.
	results := <-resultC
	t.Logf("results:")
	for _, e := range lookupDevin.dists {
		for _, ee := range e {
			pub := eddsa.Ed448DerivePublicKey(*ee)
			fmt.Println(common.Bytes2Hex(ee[:]), common.Bytes2Hex(pub[:]))
		}
	}
	for _, e := range results {
		t.Logf("  ld=%d, %x, %v", enode.LogDist(lookupDevin.target.id(), e.ID()), e.ID().Bytes(), common.Bytes2Hex(e.Pubkey()[:]))
	}
	if len(results) != bucketSize {
		t.Errorf("wrong number of results: got %d, want %d", len(results), bucketSize)
	}
	checkLookupResults(t, lookupDevin, results)
}

func TestUDPv4_LookupIterator(t *testing.T) {
	t.Parallel()
	test := newUDPTest(t)
	defer test.close()

	// Seed table with initial nodes.
	bootnodes := make([]*node, len(lookupDevin.dists[256]))
	for i := range lookupDevin.dists[256] {
		bootnodes[i] = wrapNode(lookupDevin.node(256, i))
	}
	fillTable(test.table, bootnodes)
	go serveDevin(test, lookupDevin)

	// Create the iterator and collect the nodes it yields.
	iter := test.udp.RandomNodes()
	seen := make(map[enode.ID]*enode.Node)
	for limit := lookupDevin.len(); iter.Next() && len(seen) < limit; {
		seen[iter.Node().ID()] = iter.Node()
	}
	iter.Close()

	// Check that all nodes in lookupDevin were seen by the iterator.
	results := make([]*enode.Node, 0, len(seen))
	for _, n := range seen {
		results = append(results, n)
	}
	sortByID(results)
	want := lookupDevin.nodes()
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
	bootnodes := make([]*node, len(lookupDevin.dists[256]))
	for i := range lookupDevin.dists[256] {
		bootnodes[i] = wrapNode(lookupDevin.node(256, i))
	}
	fillTable(test.table, bootnodes)
	go serveDevin(test, lookupDevin)

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

func serveDevin(test *udpTest, devin *preminedDevin) {
	for done := false; !done; {
		done = test.waitPacketOut(func(p packetV4, to *net.UDPAddr, hash []byte) {
			n, key := devin.nodeByAddr(to)
			switch p.(type) {
			case *pingV4:
				test.packetInFrom(nil, key, to, &pongV4{Expiration: futureExp, ReplyTok: hash})
			case *findnodeV4:
				dist := enode.LogDist(n.ID(), devin.target.id())
				nodes := devin.nodesAtDistance(dist - 1)
				test.packetInFrom(nil, key, to, &neighborsV4{Expiration: futureExp, Nodes: nodes})
			}
		})
	}
}

// checkLookupResults verifies that the results of a lookup are the closest nodes to
// the testnet's target.
func checkLookupResults(t *testing.T, tn *preminedDevin, results []*enode.Node) {
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
// The nodes were obtained by running lookupDevin.mine with a random NodeID as target.
var lookupDevin = &preminedDevin{
	target: hexEncPubkey("5d485bdcbe9bc89314a10ae9231e429d33853e3a8fa2af39f5f827370a2e4185e344ace5d16237491dad41f278f1d3785210d29ace76cd62aa"),
	dists: [257][]*eddsa.PrivateKey{
		250: {
			hexEncPrivkey("4e0fa28ffcc38dde53d62cae67296a409dbac36abc5b77242a1e36f9569cf5073fc4dbb8cbf656b2219e173071bc46eb52c3788b92635430ab"),
			hexEncPrivkey("e8ddff19e703d8fc127f84804368f3206d07cbeb9b52329e12facafc57b33c45806513d1200976749955d460d4ffcaf1e89932f54fd08929a9"),
			hexEncPrivkey("c9ac39684fa2a736a3a2d3dcecf381e49d051fe1af2491324790474d0968d7e8d17d85cc7ddb99443e67b8bca53447851c047340407d2705c6"),
			hexEncPrivkey("b3e39098d5b41d7d9d306068213e7312540335ef1cfd3bbaab5457790c093739aa488185129892f315a5b998700f5c87b3b43fcee00c111881"),
			hexEncPrivkey("01f201e25338f6527680b108383719167522e2f47c327e024fd4fc07d297160540ab6f390729dc344ddb439fb089d4a144a9abc2c047fb128a"),
		},
		251: {
			hexEncPrivkey("4ec11de85d5eae2185956f08e214dad13b77643cc68fdc78cd91d6f7a641ab8f4e3ec37d0a58aaafb77fd37723e021b2303b45f42f848b17e8"),
			hexEncPrivkey("8572729f472bd32f4e4cfdca7d8d9c1d1869cd24c4bfee25f3d80033aba409bfddc37a2a7d41b5df7af7acbf8efe64bbbbd320f268faf929f6"),
			hexEncPrivkey("02476279c874e251a436da90658741aab2abc70c13dce1fb1f760fc15a33134c6632dec8b3efaf7d104fbed699cd0a6c0b4c18a6cbb509148a"),
			hexEncPrivkey("2e0b546595b30da4686cb0de2a297f784790cf9e464e5148ef5fb17073daab62fc49a0db9fb85ceaa5b57694e7595fd4b1e871c49f21f53fb2"),
			hexEncPrivkey("3182b70f6a9005a84d20ea90b3c4c4e349dd1af41cbdd616e8c6d5bf19bf5b913d4882ca1afac2a1ff5e3bf243d95603f26a8322f5930e35a6"),
			hexEncPrivkey("29018630edb797f47a0bd266e2b7d20acb87563ee299e0a9fc93ef9d74823655c2bb06495901e584cb97e6b5c0c5640f1525ff063e2f6422e6"),
			hexEncPrivkey("1a13fc973e88ef8cfebb7504e8e9ecc025ad39c3be9e93b6e6997a426aababa32c1bba38e1c51231a12917e6b04b4826117ffddd6ffc2500bc"),
			hexEncPrivkey("c58a2e28d4f6b636073ba2f7743e6b339580d673a3bd7912e2d3fbcbef916bc87f8a7135f9385035c169b1b5254d6f46f9fec5a578e0be0616"),
		},
		253: {
			hexEncPrivkey("de76ffa137c42a6907d7ec018c7c396edbbc7e7381d0694bc0735ecb4c2a678d9ba1557d7b7cd55fb5e15caa36effba76cbb092e78bfae2aac"),
			hexEncPrivkey("a5c81bd0cb63b0e7c8115661736fda9bfd00a1dbedbc992d7e29672420410f3a68f6bfd9e3198411d1646984a2c71a2664010bf98d486f0429"),
			hexEncPrivkey("8f262dcbadd8e8473874491ebf27603f76e899a58934701f656ab98de1e704a84e31dbf72168026456d85a58aec626e593946d23b750723c5f"),
			hexEncPrivkey("6a8bd432989c7b9da111ba0ee260884b196b802a3afccccf7faaeb76974d830e9095f3b8f9f30ed961b23be06cd5013ce7a60ecc1831c9105a"),
			hexEncPrivkey("8dcde5e7e89f70a14e512f0f8044689739d983de63b26b887a454c227db43202a17afb37cb4d0176470e319ab0ad0d74e38ba10335e8c0184a"),
			hexEncPrivkey("45252d43d8ac130d31abf0c3b7f461b72d21e8749f811533ca9bfac2cde74e914ee480e9e293913df6a8cf13e99c1eabafbd92cca9fffb0041"),
			hexEncPrivkey("5355491fedbb2faee68a855a4b562ad19e187e84df8ec52a827c71c7f8cf30aee16a9511c339d1f330c397caa184db5d7d80ab7b29e0313cbc"),
			hexEncPrivkey("e7ccb1c830a74568e02512f9e2fdc8c09bb93ab4302781c040f1b4b0ac3b4bb0ab73c898a6ba30bec3d936f48493fc6a4fc2c6a6b7c04734ab"),
			hexEncPrivkey("6d2fbdfaf77ce9faf47349e8c67f786036a9cb30ca44853567c474d35d6a12a28fdcb909e2f08793e40060ce84524496913b43c7bfcb0502bf"),
			hexEncPrivkey("823a8b9e05d8ae96d270e1a45d9f65ec5e339ae34bb4d6d4a72fd65878a79a71575b443288a967e4c05edde1d12bd68794961f74129b1b359e"),
		},
		254: {
			hexEncPrivkey("dc7f83393ee14d76fc172916a97b2fae99c1e519254e22e5496b3fc1e141d294655f659af388aa39185bfc5edbbe477d2fca276b742d400ac6"),
			hexEncPrivkey("86554af46fb8bef12c7c2b692de76d40faf3763c777d73c2a9e8d75fc45762e4486df4367223d5b21dd8a16f6e82eb93dc2cba22ab984c3163"),
			hexEncPrivkey("3d3c1ecc8d3d24dfe57e87667752304a98e54aa9ae930bb8fcc54a72db74779633233fca7cbf3accf660ee9797252de3744ea10e713d8216e9"),
			hexEncPrivkey("838a6a651b88046811642af151a11c92bf7bcea94f3501b2dc640464a0826594e00cb263d048d336f486b7377ac6a1cc316596412a35922d4b"),
			hexEncPrivkey("db25870ff65a850f906b651ef5269134173a83780c6208efd32da1163545472fba7b1ca36941d14a7bf35ad513d54483ef1cd90eda70c631f3"),
			hexEncPrivkey("efb928b925f75f788d8a7608673cff2491a545a229f8dd3317cef12dd1e01a1513b4f32548ef2857709a0dfaf1e7608e5d2405bd459b4d2c3f"),
			hexEncPrivkey("f01db2e896243ebeae26224f1e7cf431f9b7ed42cef4891786a56bce00913058402a957a55170677d6bacdfd3ed9d477b7a475d1f01ae70a36"),
			hexEncPrivkey("cd68374c9e3563e783a64d1c4e8e10374ee683d02db3b8abda8d95344e2309defb2b6b8d6beea7534c662a5b015f62ef52b84f2a79024a3cc4"),
			hexEncPrivkey("62b9528842ef9c6718174d08fe2885c5986be762df7f9b9b79a6197541af9d254bff702993f0d8838fb73fef9f65d8d4973696eccb25d21020"),
		},
		255: {
			hexEncPrivkey("703d57c414280bd8a37c9b1006bbc9688029d381d2f7923670af9b6b5b75d5cffbb41fc48e9d4dc1b2dacfa236d90de006e1cb23cbbd710d7c"),
			hexEncPrivkey("d57bc206235847a6054e7094a6bdce137fdd8c8d991a312ddec7ebbb10ca9378dfaca0f2cb44e05143d868829add5c44c09d4ab6d224563e9e"),
			hexEncPrivkey("c89f9b6bbd5243ca9d42619b72fc0b65b3f1f913af1f23d71f3efab4e56ec69019a3a71bd8ebf0f68316c2e88fb826208bd432186967ba3dd3"),
			hexEncPrivkey("4c58b93a9273b11b0ea9d2918c0d5c43b34ea26e35492816d13b981a4a6e6e57588b5449dc00487b34eb2687cf6715933ad6704b3821093886"),
			hexEncPrivkey("d1f02c0d56ce58bb7b086a1da55cabb14ca29368d46c84d91a047dd8e16b97ea37ee8937e139d2fcfbfb6fef08f2a2f824e3aa700e820b214b"),
		},
		256: {
			hexEncPrivkey("7be4205cc11c06f7cd5229b97c5cfd1520857b5324ea94d081af3d0f5d45cb2a9b18af6c0347f4b93df8d0f3fa2abc100bd324dd3649fb305f"),
			hexEncPrivkey("f77933ae3d48cb6dff8d102db4b96258cba5d2695f9f9f23384fa08f666a50558ed58ec402742d39da5507b9700a0a12758bda6530290a2e1d"),
			hexEncPrivkey("5a60fa5213ae61a87286d64483fffc850691437014e478841f17538d8cc612a4d1664e3d1e8ef3bde5e75b7edb1fc137370f2630f7211734cb"),
		},
	},
}

type preminedDevin struct {
	target encPubkey
	dists  [hashBits + 1][]*eddsa.PrivateKey
}

func (tn *preminedDevin) len() int {
	n := 0
	for _, keys := range tn.dists {
		n += len(keys)
	}
	return n
}

func (tn *preminedDevin) nodes() []*enode.Node {
	result := make([]*enode.Node, 0, tn.len())
	for dist, keys := range tn.dists {
		for index := range keys {
			result = append(result, tn.node(dist, index))
		}
	}
	sortByID(result)
	return result
}

func (tn *preminedDevin) node(dist, index int) *enode.Node {
	key := tn.dists[dist][index]
	rec := new(enr.Record)
	rec.Set(enr.IP{127, byte(dist >> 8), byte(dist), byte(index)})
	rec.Set(enr.UDP(5000))
	enode.SignV4(rec, key)
	n, _ := enode.New(enode.ValidSchemes, rec)
	return n
}

func (tn *preminedDevin) nodeByAddr(addr *net.UDPAddr) (*enode.Node, *eddsa.PrivateKey) {
	dist := int(addr.IP[1])<<8 + int(addr.IP[2])
	index := int(addr.IP[3])
	key := tn.dists[dist][index]
	return tn.node(dist, index), key
}

func (tn *preminedDevin) nodesAtDistance(dist int) []rpcNode {
	result := make([]rpcNode, len(tn.dists[dist]))
	for i := range result {
		result[i] = nodeToRPC(wrapNode(tn.node(dist, i)))
	}
	return result
}

func (tn *preminedDevin) neighborsAtDistance(base *enode.Node, distance uint, elems int) []*enode.Node {
	nodes := nodesByDistance{target: base.ID()}
	for d := range lookupDevin.dists {
		for i := range lookupDevin.dists[d] {
			n := lookupDevin.node(d, i)
			if uint(enode.LogDist(n.ID(), base.ID())) == distance {
				nodes.push(wrapNode(n), elems)
			}
		}
	}
	return unwrapNodes(nodes.entries)
}

func (tn *preminedDevin) closest(n int) (nodes []*enode.Node) {
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

var _ = (*preminedDevin).mine // avoid linter warning about mine being dead code.

// mine generates a devin struct literal with nodes at
// various distances to the network's target.
func (tn *preminedDevin) mine() {
	// Clear existing slices first (useful when re-mining).
	for i := range tn.dists {
		tn.dists[i] = nil
	}

	targetSha := tn.target.id()
	found, need := 0, 40
	for found < need {
		k := newkey()
		pub := eddsa.Ed448DerivePublicKey(*k)
		ld := enode.LogDist(targetSha, encodePubkey(&pub).id())
		if len(tn.dists[ld]) < 8 {
			tn.dists[ld] = append(tn.dists[ld], k)
			found++
			fmt.Printf("found ID with ld %d (%d/%d)\n", ld, found, need)
		}
	}
	fmt.Printf("&preminedDevin{\n")
	fmt.Printf("	target: hexEncPubkey(\"%x\"),\n", tn.target[:])
	fmt.Printf("	dists: [%d][]*eddsa.PrivateKey{\n", len(tn.dists))
	for ld, ns := range tn.dists {
		if len(ns) == 0 {
			continue
		}
		fmt.Printf("		%d: {\n", ld)
		for _, key := range ns {
			fmt.Printf("			hexEncPrivkey(\"%x\"),\n", crypto.FromEDDSA(key))
		}
		fmt.Printf("		},\n")
	}
	fmt.Printf("	},\n")
	fmt.Printf("}\n")
}
