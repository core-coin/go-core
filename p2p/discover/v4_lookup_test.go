// Copyright 2019 The go-core Authors
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
	ecdsa "github.com/core-coin/eddsa"
	"fmt"
	"net"
	"sort"
	"testing"

	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/p2p/enode"
)

func TestUDPv4_Lookup(t *testing.T) {
	t.Parallel()
	test := newUDPTest(t)

	// Lookup on empty table returns no nodes.
	targetKey, _ := decodePubkey(lookupTestnet.target)
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
	if hasDuplicates(wrapNodes(results)) {
		t.Errorf("result set contains duplicate entries")
	}
	if !sortedByDistanceTo(lookupTestnet.target.id(), wrapNodes(results)) {
		t.Errorf("result set not sorted by distance to target")
	}
	if err := checkNodesEqual(results, lookupTestnet.closest(bucketSize)); err != nil {
		t.Errorf("results aren't the closest %d nodes\n%v", bucketSize, err)
	}
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
		done = test.waitPacketOut(func(p packetV4, to *net.UDPAddr, hash []byte) {
			n, key := testnet.nodeByAddr(to)
			switch p.(type) {
			case *pingV4:
				test.packetInFrom(nil, key, to, &pongV4{Expiration: futureExp, ReplyTok: hash})
			case *findnodeV4:
				dist := enode.LogDist(n.ID(), testnet.target.id())
				nodes := testnet.nodesAtDistance(dist - 1)
				test.packetInFrom(nil, key, to, &neighborsV4{Expiration: futureExp, Nodes: nodes})
			}
		})
	}
}

// This is the test network for the Lookup test.
// The nodes were obtained by running lookupTestnet.mine with a random NodeID as target.
var lookupTestnet = &preminedTestnet{
	target: hexEncPubkey("5d485bdcbe9bc89314a10ae9231e429d33853e3a8fa2af39f5f827370a2e4185e344ace5d16237491dad41f278f1d3785210d29ace76cd62"),
	dists: [257][]*ecdsa.PrivateKey{
		251: {
			hexEncPrivkey("4ec11de85d5eae2185956f08e214dad13b77643cc68fdc78cd91d6f7a641ab8f4e3ec37d0a58aaafb77fd37723e021b2303b45f42f848b17e82c8cc056266762a202371c972318cb1afe3772a6fb9640a01ea2aba1156b503fabf83aa7792c06585ada8fbfcdeee15f2b83a9bbe57a1eea13a7beccb024bd57ad6d0cdf2930b519d3d2b3941e8ff05e33cf374612692c"),
			hexEncPrivkey("8572729f472bd32f4e4cfdca7d8d9c1d1869cd24c4bfee25f3d80033aba409bfddc37a2a7d41b5df7af7acbf8efe64bbbbd320f268faf929f6265e47b3a5b3b2c9a89a8a877f510f6942400eb18b12301c7a2796a1fd2f38407b07bff41dacb401d18f757e76a5ba5d9401d71f0532fe177dac7ac10ba9f846ac052e2ffa7e72f92b5b859633a31542485db293c76968"),
			hexEncPrivkey("4e0fa28ffcc38dde53d62cae67296a409dbac36abc5b77242a1e36f9569cf5073fc4dbb8cbf656b2219e173071bc46eb52c3788b92635430abeab62310d615c5c9ad43e967e70b2b6237ec6f9b28e07d735ccacf86e7e03fb022cb5b4a006740d381172e8d3200ce2e8c317ea8eaa97db1e98922280871239f820cfb449198402d980948eca8704ee2e8a8462d6b056b"),
			hexEncPrivkey("d1f02c0d56ce58bb7b086a1da55cabb14ca29368d46c84d91a047dd8e16b97ea37ee8937e139d2fcfbfb6fef08f2a2f824e3aa700e820b214bf366faf881006d9d64b4b29ff119b06e21b4e22449f384f15634854cd3dc20518ba95b591c3b68b47db85aed4badd98ee5d1b64a9fdab8430e5b5c85fdca19ac80e576a9d99eb8044a798942b0d6407b35f850a7fd6b09"),
			hexEncPrivkey("838a6a651b88046811642af151a11c92bf7bcea94f3501b2dc640464a0826594e00cb263d048d336f486b7377ac6a1cc316596412a35922d4b9fcfcbeb73288841583b311b5ac405e8cba059953db4f7cfa3008a7b41099f8713ba6195ed0168607342675f77d26d5a6088e540401ac2afa3a2acca107e6dfe43d1cf167e50754973889e710d9faf13f316c16746b0ac"),
			hexEncPrivkey("01f201e25338f6527680b108383719167522e2f47c327e024fd4fc07d297160540ab6f390729dc344ddb439fb089d4a144a9abc2c047fb128a891d5d6a3f6614080090c39d4d29884093d66ad53ea3a698873d913c7a42f436be2f47ace698ac7554cdc5e549ad2defcc39a8d1a9324d2b9774af814c47af621e9b00b1f72b855752ac81ec9194ca63067900df3cc640"),
			hexEncPrivkey("1a13fc973e88ef8cfebb7504e8e9ecc025ad39c3be9e93b6e6997a426aababa32c1bba38e1c51231a12917e6b04b4826117ffddd6ffc2500bcec3c02f4da7b3a3ae0afa1d3a56972b6513138ed2039597ebc79f28bcd07b3323cfd41e98eb2fa7d43e9fc63adce146e1b5ef602f565a40c5f44416c7354bfdc4e3fba549d51b8a9ed8465498d99d308cac2c24787003d"),
			hexEncPrivkey("dc7f83393ee14d76fc172916a97b2fae99c1e519254e22e5496b3fc1e141d294655f659af388aa39185bfc5edbbe477d2fca276b742d400ac65146b7bf64a2ff8e664876dbafdc43982fe2b094de423d203f395ea4873efbb8504d5b8a179da69293a42192f35f61f71a877ac4054d1f2d457ec55cba1a55a80167f3de20b2b51dd3714f7dbe94b11de2c513f7c75972"),
		},
		252: {
			hexEncPrivkey("db25870ff65a850f906b651ef5269134173a83780c6208efd32da1163545472fba7b1ca36941d14a7bf35ad513d54483ef1cd90eda70c631f3c6c009c07790762f9377d1bb68df43b111b520a77b074073c991b603a7c8f1948826c041d23aeb3fa347df19cd0d4eb047f693b2e2a94f51add8873c28b1385f878c3d34ffdb298af73258ea7053e1ae8cdd8670eea93e"),
			hexEncPrivkey("e8ddff19e703d8fc127f84804368f3206d07cbeb9b52329e12facafc57b33c45806513d1200976749955d460d4ffcaf1e89932f54fd08929a9b7928bd6eb059deaba3160636ebbd514aa00b177a0f65ac675221bf71a91f4eadeab68a2c8cab94e830ffef532877b205ca5d6f5526d00af2ad814f43d06e245ccd6bf6552a342edda3c0fe3a1fcc5460bf33b0fa843fa"),
			hexEncPrivkey("c9ac39684fa2a736a3a2d3dcecf381e49d051fe1af2491324790474d0968d7e8d17d85cc7ddb99443e67b8bca53447851c047340407d2705c6ea0ea2063b3502b1743051940c52a5a0100f61093cd404a3eb751bdd950d6f7beb75d709254066850666d2ebcde2b1188bb8d2e67babeecba9c7a6d875005ea9441ea75a09f6b6565aef58670bd4506af0243469b1e241"),
			hexEncPrivkey("cd68374c9e3563e783a64d1c4e8e10374ee683d02db3b8abda8d95344e2309defb2b6b8d6beea7534c662a5b015f62ef52b84f2a79024a3cc41d90b3566a908c8bd31afcf070ec07041e2f2b5d9a74fc91353c5fa42f747f893321e4072ca3b2fcc0bf42cbf3d384b12f298783aa53c1cf7f59556ddadca75c8255ac701ddb6ae510d8a2f4d69bed751c683144c9551e"),
			hexEncPrivkey("7be4205cc11c06f7cd5229b97c5cfd1520857b5324ea94d081af3d0f5d45cb2a9b18af6c0347f4b93df8d0f3fa2abc100bd324dd3649fb305f2a97274ff15a0a3594fde911205103865f1a39dcb44fec1158b85b7bb5b48a7184cf4897da0bfd6b193fc14aa293ebe6c8c8adb11345fdc356653ed47282d4d69a5b860e7b02bb9662c89c94eb97eff1287a1241fd5e11"),
		},
		253: {
			hexEncPrivkey("de76ffa137c42a6907d7ec018c7c396edbbc7e7381d0694bc0735ecb4c2a678d9ba1557d7b7cd55fb5e15caa36effba76cbb092e78bfae2aacf6701342f8ef1e063f31e9eff7928993ffdcb0d13750c07985a15c1e9f4888fcd42baa5f1cb91dd4a13323d39fad11a019e4dd3d2e989631223aa92034e5b37f6e19906f9d39a122a451e62101d7d883d5d6be1b870642"),
			hexEncPrivkey("c89f9b6bbd5243ca9d42619b72fc0b65b3f1f913af1f23d71f3efab4e56ec69019a3a71bd8ebf0f68316c2e88fb826208bd432186967ba3dd3c777805908f662ce69981006a68da37b3e2105ab9a96b28b79832f7f40715e77d1c589f14bc914d7915b8976317c78603b732a7c7f5453a7e3ceb4484623c31c98cd05ae2cb77bb00252b42b9f5251bcfaee738d5c7a7c"),
			hexEncPrivkey("8f262dcbadd8e8473874491ebf27603f76e899a58934701f656ab98de1e704a84e31dbf72168026456d85a58aec626e593946d23b750723c5fd6d1d5bc7ed5a7e73beace21d6568206e698797be64a314a669f918dbff28decd845502fc79f3eb79ac37f9cecbea424f9231f5bec1a159ae7942163e55044cf39213e3c08dd92fdd7f8bc2b50324f65900396fab66d49"),
			hexEncPrivkey("6a8bd432989c7b9da111ba0ee260884b196b802a3afccccf7faaeb76974d830e9095f3b8f9f30ed961b23be06cd5013ce7a60ecc1831c9105acf8e211e3d3e2ba310afa91edc15389a9cb2e59525774646dc46030d18d880b5f0ef616f842231a355e725589dd45a2f2677d028b6fe46192c2f08eb5141a7e3537318291d39fb1f7028537b56416fcc97bbaa63422d15"),
			hexEncPrivkey("8dcde5e7e89f70a14e512f0f8044689739d983de63b26b887a454c227db43202a17afb37cb4d0176470e319ab0ad0d74e38ba10335e8c0184aa8946ebf664270106abd7aea3a27e66124c04deb391de0997be90f973d5369fd287442b79e7553b642aaca22159567b503c36e128918479809262bc78d681491f510769dd76b33c89f3f4de416d4180c9af611ee809ae7"),
			hexEncPrivkey("45252d43d8ac130d31abf0c3b7f461b72d21e8749f811533ca9bfac2cde74e914ee480e9e293913df6a8cf13e99c1eabafbd92cca9fffb0041edd55ee57e76cf7e7150631e1dd42d8ca20a00b678689f8b863fa0c4f995de939796abee99e7709dc8cfcd313852d72a8f7a0afafa3e4d3b6b78a38e7d0f4b2a366634a909a7931b69c1296beb5c65a1d55c5d0f5b838a"),
			hexEncPrivkey("5355491fedbb2faee68a855a4b562ad19e187e84df8ec52a827c71c7f8cf30aee16a9511c339d1f330c397caa184db5d7d80ab7b29e0313cbc3c28ae4ef15d1343b9178442e28c6214682dbf1b819d326dce0a171d8fd85f04cc1c9646f0452d4da85e207a21dd10f7a56fe6e92d8999e411f507ab12eb30e313966909d586c41384a2da63d96ffa627c1151fdbc2bc9"),
			hexEncPrivkey("e7ccb1c830a74568e02512f9e2fdc8c09bb93ab4302781c040f1b4b0ac3b4bb0ab73c898a6ba30bec3d936f48493fc6a4fc2c6a6b7c04734ab2caa6c9c5c5fc3f2b03b626be895abc279d78a4c34c7b6c2e33a9455f8d25233a001ea9f22eef45a6498f0ecada2daa25185676318c00364930e9b5fa762537ab47f69db873350bbde1d58793cfd21bc00c1b9fe05ae15"),
			hexEncPrivkey("6d2fbdfaf77ce9faf47349e8c67f786036a9cb30ca44853567c474d35d6a12a28fdcb909e2f08793e40060ce84524496913b43c7bfcb0502bf934014e88ba516c6b8ab849c70a6a28a32df9247c5bf4b964812861199ef982ef811e73dae09fdfb39db470bff1190c3e7f3df7b74cb1c8981a40bcbb96d056a35637440d6f91e3be093c527b30c1415db12f0b1293552"),
			hexEncPrivkey("703d57c414280bd8a37c9b1006bbc9688029d381d2f7923670af9b6b5b75d5cffbb41fc48e9d4dc1b2dacfa236d90de006e1cb23cbbd710d7c0eab3bb1f14cca70fb5742cb3b1d971479588f6f1c311bc514a9c3f1aac36d3a1437f8f755e16cdd7d9f255cd2dd99ba75350b758aaf17ed22c7748f29ca340373e4c20105673660dc56e422f4483b4e370ef613b9a6c8"),
		},
		254: {
			hexEncPrivkey("02476279c874e251a436da90658741aab2abc70c13dce1fb1f760fc15a33134c6632dec8b3efaf7d104fbed699cd0a6c0b4c18a6cbb509148a1099d5d965c3755c42b918b31226c707f41c0b4c05ca5e3d6a317028a41497d648fd7ddcd4c816ba0750e3ccd06675e7d09d6cf82d9ba247f2c4819bff240fe8ef5dc0a1dd49c3d2335e553f9a9d0bad3f5bda280f04be"),
			hexEncPrivkey("86554af46fb8bef12c7c2b692de76d40faf3763c777d73c2a9e8d75fc45762e4486df4367223d5b21dd8a16f6e82eb93dc2cba22ab984c316334948c5d94762da39fa358e4b9260b035da86d231dd64a3a90f4b45f4aed02a99c7b55e20b163601f3482240faa06fc8be1a33ac4a279115c833ad25dfa94e7eedcb755dd7832611e51b8d8e344e99bb4d07be60b825a5"),
			hexEncPrivkey("2e0b546595b30da4686cb0de2a297f784790cf9e464e5148ef5fb17073daab62fc49a0db9fb85ceaa5b57694e7595fd4b1e871c49f21f53fb2b87e556be889442dac0c35eba643c25afe954f06f34793c051209c94e981a7bfac06f9df56f30139a507a1d67b0597ad5dfe450754b58b5de3fea3d540918163d73187224559899b74e0f670d8b9456aa14e38c43baa82"),
			hexEncPrivkey("f77933ae3d48cb6dff8d102db4b96258cba5d2695f9f9f23384fa08f666a50558ed58ec402742d39da5507b9700a0a12758bda6530290a2e1d3213bb14e5cccc5ea5e5e5e910f285cce71f1ca917c5f4a10c3609312ec278b229af2a7a08730cf92a6a24f7357843e3dbb234d6ed26c0182d90eaf6e116e9366b97642a3330699058b921c30496ffcf5723910d0bc31e"),
			hexEncPrivkey("c58a2e28d4f6b636073ba2f7743e6b339580d673a3bd7912e2d3fbcbef916bc87f8a7135f9385035c169b1b5254d6f46f9fec5a578e0be06162a015e5f1e193ed0842b6e649184a25313cbb76378d6f73d63bb0af287daa0f1a73f78178c07a2faf8e9e27626eb14529344da1c0cc347fcf89199f108259ccececf10b0d72037031a29954b906c9446948da3fb8582e1"),
			hexEncPrivkey("efb928b925f75f788d8a7608673cff2491a545a229f8dd3317cef12dd1e01a1513b4f32548ef2857709a0dfaf1e7608e5d2405bd459b4d2c3fca0eee867995e486e7a5c1cf93fb98bfc9e1722bd52ad4e8d9e0d03ae3ecdbb4bb8f53c2b0f46be63b623f95a94f7b85dc2a854c36df31e621ac1219a921022fe6ae9ce140440985512eda7f01647a9da7faba01871a3f"),
			hexEncPrivkey("f01db2e896243ebeae26224f1e7cf431f9b7ed42cef4891786a56bce00913058402a957a55170677d6bacdfd3ed9d477b7a475d1f01ae70a36232445c01ba35149062216ec94c8617bd96c09105f454383520ae4fd4eb33ea353a2bb823a3be09e23897ae084aeaf7cbfde739c26f6f1b61acb829f1b6837784b0d164b11ae63ab0810f71447dfcc592965d729ec134b"),
			hexEncPrivkey("a5c81bd0cb63b0e7c8115661736fda9bfd00a1dbedbc992d7e29672420410f3a68f6bfd9e3198411d1646984a2c71a2664010bf98d486f042935e665a6d38b2afee272dbd4c1de6938a74ec52587bb1a9ffdb9fc24eeda6c08c55076fd3f7a0b4e1fe8cc86f03ba905fed8140c5079f2f2827713e30f1a0c5f31591a35226129ec0d521b81fdf03aec150f4411c9ce50"),
			hexEncPrivkey("62b9528842ef9c6718174d08fe2885c5986be762df7f9b9b79a6197541af9d254bff702993f0d8838fb73fef9f65d8d4973696eccb25d210203cd15e39da3c9ce72c3855cd0e58d079db16249d037eb6b47a0a919728438c5c29746f10322782f521847dde7f79266df373d9691dea840b0038a323703a6109851870d5b0ffb22bca9fcfca2a6f9ba8d30f12b52ab8ce"),
		},
		255: {
			hexEncPrivkey("3182b70f6a9005a84d20ea90b3c4c4e349dd1af41cbdd616e8c6d5bf19bf5b913d4882ca1afac2a1ff5e3bf243d95603f26a8322f5930e35a60fa7c358ab0b974595530af9ef598bb2d3ee059d51cd45e577461c7109964f892ed945fa7e6f042bae9ddbff38737d6d3862c5c817117d1a18c9bd9d0916a035459c23269625d279be11829600f5a9fa8e996e6f70222e"),
			hexEncPrivkey("d57bc206235847a6054e7094a6bdce137fdd8c8d991a312ddec7ebbb10ca9378dfaca0f2cb44e05143d868829add5c44c09d4ab6d224563e9e4230f46324216e351324857b6a3197920c2ecf4d74b7c5ae1ed77a28a5c3b055dbfe2069ad99d3478b3be877862d9bdcdd4422339d8992ffaa18647afaf29041b8d30c9240ac5281c13887bd7bbf3c2808e89a50e9745b"),
			hexEncPrivkey("b3e39098d5b41d7d9d306068213e7312540335ef1cfd3bbaab5457790c093739aa488185129892f315a5b998700f5c87b3b43fcee00c1118810f71c02c1b3378b9d061086140cd452665ed4450288f282db512c8db5cd50470bc215780356e896974cabe1351154a1f034a87a6877acfb45b3bb1a36e371a9b53a102a1b8cf70fae15e0be182996056eedf16fcf1fb21"),
			hexEncPrivkey("4c58b93a9273b11b0ea9d2918c0d5c43b34ea26e35492816d13b981a4a6e6e57588b5449dc00487b34eb2687cf6715933ad6704b3821093886c0f2f609e0abd2aa3be5eefc432b0337ddb010eb1617a551c7a560b0e5e76269fdf833513681ff73c11a6a9c2e4052aad9191fc86ca9bc06bc6d754f50351d53af31b863b49775ec67e7cf1e911e9b8ba95910f794b954"),
			hexEncPrivkey("823a8b9e05d8ae96d270e1a45d9f65ec5e339ae34bb4d6d4a72fd65878a79a71575b443288a967e4c05edde1d12bd68794961f74129b1b359e35cda5cc01a9cfafd34840851d3b72aa61ed5a7b3fce9afed0ef40bc70493565dc253cb07ce7a39613499de84a4b12979a21de3b65bf83a5afe238229129cbd275a4a0dea7315ba05c471aff1da890519dc057e6ca2bfd"),
		},
		256: {
			hexEncPrivkey("29018630edb797f47a0bd266e2b7d20acb87563ee299e0a9fc93ef9d74823655c2bb06495901e584cb97e6b5c0c5640f1525ff063e2f6422e6b916250e60a6ceece44a96b82257abbdcfc3b94f84ee1ab5cee42bcf810ce88fe164f221efd10eb3208d2e4c7997724f3b727bfbab29cf2a42b9956a7cfb2020bef38700e9d92ab7ce258461ac72d4b561c0939a6490e9"),
			hexEncPrivkey("3d3c1ecc8d3d24dfe57e87667752304a98e54aa9ae930bb8fcc54a72db74779633233fca7cbf3accf660ee9797252de3744ea10e713d8216e9fbf85dabcc4257f8fd10bfbfbc25232b554b3eb11f14ea2e8fad730385fb46414e185feddfa34b7c9e06f0e71093ba98591d4299e27e0059c8ab865e61ac2550e493adc0ea74e8c29129ee9948888a5a82b6c4724cd69c"),
			hexEncPrivkey("5a60fa5213ae61a87286d64483fffc850691437014e478841f17538d8cc612a4d1664e3d1e8ef3bde5e75b7edb1fc137370f2630f7211734cb5f1d28c3e95c894d3a630edb2c44f1552984a0ea6eb5bcc1453d9f11e992acc7aa2cf840fa172d251a1c76e0f891da0d4948b3805520e87d86b046dda90e26cfd024d2d321ffcbc3ea70d38e1ebe96fb81013945058b32"),
		},
	},
}

type preminedTestnet struct {
	target encPubkey
	dists  [hashBits + 1][]*ecdsa.PrivateKey
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
	ip := net.IP{127, byte(dist >> 8), byte(dist), byte(index)}
	return enode.NewV4(&key.PublicKey, ip, 0, 5000)
}

func (tn *preminedTestnet) nodeByAddr(addr *net.UDPAddr) (*enode.Node, *ecdsa.PrivateKey) {
	dist := int(addr.IP[1])<<8 + int(addr.IP[2])
	index := int(addr.IP[3])
	key := tn.dists[dist][index]
	return tn.node(dist, index), key
}

func (tn *preminedTestnet) nodesAtDistance(dist int) []rpcNode {
	result := make([]rpcNode, len(tn.dists[dist]))
	for i := range result {
		result[i] = nodeToRPC(wrapNode(tn.node(dist, i)))
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
		ld := enode.LogDist(targetSha, encodePubkey(&k.PublicKey).id())
		if len(tn.dists[ld]) < 8 {
			tn.dists[ld] = append(tn.dists[ld], k)
			found++
			fmt.Printf("found ID with ld %d (%d/%d)\n", ld, found, need)
		}
	}
	fmt.Printf("&preminedTestnet{\n")
	fmt.Printf("	target: hexEncPubkey(\"%x\"),\n", tn.target[:])
	fmt.Printf("	dists: [%d][]*ecdsa.PrivateKey{\n", len(tn.dists))
	for ld, ns := range tn.dists {
		if len(ns) == 0 {
			continue
		}
		fmt.Printf("		%d: {\n", ld)
		for _, key := range ns {
			fmt.Printf("			hexEncPrivkey(\"%x\"),\n", crypto.FromECDSA(key))
		}
		fmt.Printf("		},\n")
	}
	fmt.Printf("	},\n")
	fmt.Printf("}\n")
}
