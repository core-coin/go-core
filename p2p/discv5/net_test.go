// Copyright 2016 by the Authors
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

package discv5

import (
	crand "crypto/rand"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/crypto"
)

func TestNetwork_Lookup(t *testing.T) {

	key, _ := crypto.GenerateKey(crand.Reader)
	network, err := newNetwork(lookupTestnet, *key.PublicKey(), "", nil)
	if err != nil {
		t.Fatal(err)
	}
	lookupTestnet.net = network
	defer network.Close()

	// lookup on empty table returns no nodes
	// if results := network.Lookup(lookupTestnet.target, false); len(results) > 0 {
	// 	t.Fatalf("lookup on empty table returned %d results: %#v", len(results), results)
	// }
	// seed table with initial node (otherwise lookup will terminate immediately)
	seeds := []*Node{NewNode(lookupTestnet.dists[256][0], net.IP{10, 0, 2, 99}, lowPort+256, 999)}
	if err := network.SetFallbackNodes(seeds); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)

	results := network.Lookup(lookupTestnet.target)
	t.Logf("results:")
	for _, e := range results {
		t.Logf("  ld=%d, %x", logdist(lookupTestnet.targetSha, e.sha), e.sha[:])
	}
	if len(results) != bucketSize {
		t.Errorf("wrong number of results: got %d, want %d", len(results), bucketSize)
	}
	if hasDuplicates(results) {
		t.Errorf("result set contains duplicate entries")
	}
	if !sortedByDistanceTo(lookupTestnet.targetSha, results) {
		t.Errorf("result set not sorted by distance to target")
	}
	// TODO: check result nodes are actually closest
}

// This is the test network for the Lookup test.
// The nodes were obtained by running testnet.mine with a random NodeID as target.
var lookupTestnet = &preminedTestnet{
	target:    MustHexID("1033b1bac4c731e800b6399a357e51cf1b20eec942aac608c90b89553003e2ed3f94bd80613ee9006b1e62b6bb45109d0db9a4833e78363991"),
	targetSha: common.HexToHash("0x836e8561dd02b1cd880efd2eab25756f5c7adc3c253c257b89cec4c1e7d3690b"),
	dists: [257][]NodeID{
		243: []NodeID{
			MustHexID("38885618e2d8086858f4815eb7c4696daa4122083bcb866a95f9126ef8efe40085317f61f015f811308cc450d71030b9b1dfb9a7197168fd00"),
		},
		244: []NodeID{
			MustHexID("87230d983b1d7dc9fbdef4c79d2a85467b465fe51a8bbcdfbd61cbae4df1bd05943162051db8a944e52a856b752ae437bc8217b9e9e9976b00"),
		},
		245: []NodeID{
			MustHexID("5488fcaa6cde1f167aa36a5e5b80382c53dd4ac27793aaf4ccfb4a8ccc9fcda6128290f4734d8ab9ea3fb735ebca8080cd38d6ec0c2ec6a000"),
			MustHexID("4cbd55aca598750a34669fba4cf7a48b07807f4e61fae21458b121343f6182bf27fbb7f794bf2a65c49c7d2fba53d09c4477302c9a4e0d2380"),
		},
		246: []NodeID{
			MustHexID("437c3b5eabf721a5a241cddaefa0e9b6d229ee20c83116e45cf1983595e653258763ea4bf926456b4701c110ed3cb94f1f7fc0e3456478e100"),
			MustHexID("23c03b51d98f6351928a2d2720922d665853b5edbb4118fc1e2d460a0ab9cff5b4949115a11d03b676088992bc418ae3adcd7bd432725cd480"),
			MustHexID("94dc4b05d71c031a7f789d774e78603af4f2d61f299acdc00b212e09ce74f57c8b1a3277a0de7353c65802e88e02723e139eb87f56b33ed900"),
			MustHexID("aa5db34e9345ed3063be1e33766a85ca634f1fb35b889c6cfba40b8066da325a187b3cf3eea9b5b248a8ded7f10893aaedcba7f4f5aacff380"),
			MustHexID("d132668dda298dc1d0cad75fe3a1eb569a5a215f0abf3ad5c0ace0312e4f1e690c747fd60f4837f6ed827a5cb1035d8622df27cf32df0d7900"),
		},
		247: []NodeID{
			MustHexID("07f85ef84be314ff879ce86b6c1bf04e7dbf5994aa062461353fcefc0c4e24ca2ce8d74b3fc77e0db274b7cbea50d97cc7b01fe9f4a8ddc680"),
			MustHexID("1a28a3a9985fe9e17ae2cbc7275304824aad53cb5d6415e38d04ff85d33fd9c6fa9e142c241d7d3fb85b8a84681eaa2245ac933e1749a28300"),
			MustHexID("c063cadb1e1e177cf9a888ed9c7efdf3b96d3b521399ff6b6db82c1556cdc2ebf9fa57d92cddb39b6aaa90548da0b79c373586fc279c677200"),
			MustHexID("1d07416339a15ab8d096eb8f7625778035931deb1c344af117c2fcf4fe1e3f1017e60388efdbc1d6672efac73497911eb0a576ad7f46262b80"),
			MustHexID("15708f2a52df92d9edddf13fddcaf67c73338c9790baf4ccdb596cb6d897103aa273dff7781bebce665e0df59355329e8ef51471a14d073200"),
			MustHexID("f94a85cf3cab501ef3a194b3c163e61a6f6e39d43bae6a17ae6ae4ec6c636e6a08d9651fbac7a83a017179e226a71718eff65f568d3b536700"),
			MustHexID("36e248deb5bb3e313d2a15b56f6b787b25d1e4702ad63d1e8a2723e316c09d4b26e3df7a12037c18e743e3a32427cc041397590fe16a344400"),
		},
		248: []NodeID{
			MustHexID("8a415b5dc4c325dab1a005586ed771fc8f5fbf6f0fbd171b1b2510880a0f24b3c2b7f9318fe6ea55dd781982501f26b067ea4a29740de9af80"),
			MustHexID("2bace5abbb92d0aea00b2bcec933c7d1c9016c571ea4366b6f482e318643576f40c5cd90bd315eb655857672babf5ea203fb7b1a3e520e3280"),
			MustHexID("4056e028e03536fe6b0e062138f416604d0164d302e30244e3990c9f4206e8f2a82d09bcb7baaed2693570aa72a21ae49d15298068b9a08700"),
			MustHexID("5038eca49c7446c33809950dae7e5f57defe7e6d10fe6d6da6994862b7990c18656a904402785709f0add7783b60129c88fdf25a5d9b42c700"),
			MustHexID("cdd5753f19bbe687227e74d6ff7f1383f278d42ad4258905c653d36c09b68c0f799beaee2380b10e7cd7141db4ed550681846e3edd18c40900"),
			MustHexID("9edd72992a4c65a65cf40e3ce37609eebeeeff9c68a4e2df983345d8036041c3de47e809ac6e19ddff0a1d1fca3305fc8032b21f7ff8c5d200"),
			MustHexID("c5195940abb9d35a44390fd182f17388bfb7d4e5358f1bfc8bb8cca82b431d28691abebce09515b1272e45fb572ed5cb513ba5e727f0b45600"),
			MustHexID("6993ab7eb663afb68dd22fb3f8ce817adc1d24ba214e25bafa79daf92c8b2fee316d209d2f32b8a6e66ff7d8785722721364bf628e1a382680"),
			MustHexID("a47334e3d024032165390f0d3d89b2af41978327c4ccb3fcb1297ee4f1173dfe188023849697e21fe6e16fda4e6b8d7a7b584e68d94ca0f700"),
			MustHexID("3a46d40f853cfac1ceeb55078d86ded68e7e2a68077851fc41927056cf8c14cd9747cee7526a83d5599b103359fe83a8c6b40cc6c776f8e980"),
			MustHexID("7b0a972a717a2261e8b4b028dd87263588b8b2bd830115bd5176b0ba41b4a57a13225c3e153c698dd49cb496598b23f1e922d93c5536cba500"),
			MustHexID("66eeba3a9c13f8174d2c6a38b2c33f7a19559ebb0de984e4f6d32e3f4c0645fad86220baf94dd71d5cc0654c09de5bbdf7740c345fe8877900"),
			MustHexID("84e12cc5efcde7c0ca5d65ffd9e4445df051490b36438428ceb1af3db4ec2fab335f72022c8d8d254f2d8f8f229a87488ad63df68867b8f980"),
			MustHexID("a4b9962c63cba51fd0a74118c080ebc43750efc05c84c12700a9e90afca95c863e8823b29cd63dc7942f99af6c53218e1297fc310c06024680"),
			MustHexID("a530a17d341ce88728eb26809bf950b811d6fc043f6bf00ec21c5a6fced1fcb6fb572d0738df0fe64da2b9e03cec6194e68f23b192a483ab00"),
			MustHexID("dd4bf75025b7e613c77c31e0d397ba3bb40885a0158153746d2f65a2fb14b345a9d206c5d74d951e40ca15ae51b2dff7412aaaeb137835a100"),
		},
		249: []NodeID{
			MustHexID("e4b662cedb6f29d3336409283d5f02d8dc6fa07348506243ec826fafae86c64a8601dd39fd51aaace8e899fc3fe514f233fc972223254c7900"),
			MustHexID("428d681f3d13312ad84eeb7f03a764ba01457e828e5f5fc2629893fba93af8024dcb76db6285606a5a3819c1c3e59344709df2b6ecbc453300"),
			MustHexID("8ee1b28c83f7040abbc78829c240f78ba1326c440a1c16c77c60848313267268718f877289427e3fad1b8c87419e229ce46f8b46ca4a34a900"),
			MustHexID("e5268b4fdcad2f8511134ff8159c1c6e2c44a6f474c396caef7541c36b90bcbee85f4bde4a36b0815317aae25884fa673c146e27a12db50f80"),
			MustHexID("e7f958a1e805f2ad7b3d874038c33aa48a19c01102682b6d8428dbaf13dfff6c58d6fc0df4d80d507fcd5f0dedcf67e018ebd0ad2c18522d80"),
			MustHexID("3ae2e985ebdc66e70de560fbefdeb1a310290ee97b1d78755148d26c7636d4402560396f49f3b0427ff98cbd5ed3f09c2dbb920f69e215b200"),
			MustHexID("c86987959ded894113539f8a5dfa499e0df64cc9a85e67f4068e1d5285d0a7b2ba54956e133247d9b6cb017c40375ff8b3e2e9083a11727600"),
			MustHexID("0d97e508e3f7381fb5475ea38900ac4f9e05a60b1c106f4bfd77ca8393611fa005eda697c4126638c5acaf0a7a2e586563ef35ef8c8d4b7600"),
			MustHexID("0ed93300433af4f7d51b9cc81123985a6032ec9999af08b9fe4733dc438cfbf10f565cdb7a9e805e3b9c1f36b05481821801bbd1c745140f80"),
			MustHexID("2a9b39b6b7a952b76e4312a6098c13e44582c8b0febea1a079fd420a34f3d281ea545cba9452f95b92cd850ed0b57e0b4f6f918709ec163500"),
			MustHexID("ce1824e36726aaeb037d6388152450f1a9eada6333a3b8e8514eb7f70e430b0ea147346caa25c238eceb59839efd3987a7602f90da989de480"),
			MustHexID("07a6e79b4dd5ef730a3f760eeb2afe8019d01da420ee8d4bf5e57a6242f4913c9b62d758df7751a8ab72cf3ee3def1c44e903a2e1eceb99d80"),
			MustHexID("0404a4c27078c53ae2a44c7757db383c044cff12c746fe97f8449feb0099657eabd3a702f95738ff017d22983634d4ae5b249e678cd84c7100"),
			MustHexID("c8514efd73f9d025195576c4a28cae434ff6c5bdba54d55c0b3af998fc8b795af0092b77134877125a625797599d681f91a7854480fb5f0800"),
			MustHexID("16f502c3083e7cf6243f16a9aebb74514b053b0b781b86ef7349c78904daec87e00df3c6a2c2eef941b314fa4a33df752222e83def13b39600"),
			MustHexID("4401f807669daa850219dd617fe3642d9d638bdf0f916fec57243e9258121bc8ed3a682f34595c83ea9d1f2291a82b64cb02e30eac115e6300"),
		},
		250: []NodeID{
			MustHexID("5af5075e639707cf2040d3092203d929288d89f851f792aec21ad6c8866ee31ceecf7985b41324f0dda851b544ae2ac5084a460276c4469380"),
			MustHexID("365adcb9e0ac1aaf3f28ccb46b8d1f0859d30c8603e6fdb93aaf49b1eed2748e968e9186cc2b9c9aaacbc84c89235334ff91234d6523973580"),
			MustHexID("9443232f2275eb94c063fe8fe45136bd3b90d23da9a28c8e9100e783c4f987bdaabd3485a817fda20aab86bca845bd9d39f2248a2b6dee0c80"),
			MustHexID("cfad1533b7530186cc6069de19ca810f226e6416d68a7584b5ad1c1714c308942ce1e6f66e268bb2d351b8bd39be067ad5f23c3d5459ded100"),
			MustHexID("34ae619e31e9686b0afcc65976d71629523a7eb15dec24b61ce62c4ced0fd0bca6531bfa5924c08ba916211a612dc9108c8e12d1b28cbe5f00"),
			MustHexID("91438f7cc639c2dcfd95182660d1d90039d01badea90005068919c369503275733b96b607e2da42c030010ef16bcdbd3c1a6e6cdc5ec0a3880"),
			MustHexID("510d4cb11f6ba6ab7a0897d042e87159c5e078499b407af0158ff0d509253f1f0e757ce3889ff36cd227334b8e38f4c63d4ae1115024c74800"),
			MustHexID("e22e2a86f065712099c53d3276c2590fc9fe500490dcde669e3ef502fc537c6f4de482aedf725265b8e0ede20e9646d9d99ec3e8765df55c00"),
			MustHexID("c47dfe1f6f060aa86bec4fce30027b229af13e10e976b3cb459824dc6a78d33249fd2a9c9026fbe2ef9ff9a578a4e3eede3e1e1afd88925800"),
			MustHexID("6a6bb2bb2c4e2ccb55be77e93e7638d9777b5a3a81b5ed23bf493d0910e5c9493dbbf8d028f65d967a20a8ffcfac27253df5b78cafe31c4300"),
			MustHexID("24aef291f910e918372ba1bf6da55b75bcefc983e150b148ea9781e5e25b3273df9a5d7bea06ef112d9c5aad6ea66f2ec5121ad402ad3f8300"),
			MustHexID("cad4ef5bf3f4472a728d07e8264637c983ff5ab81ceaff3597d64e84981d3642b623f4548578f53a0389d679649d0a6b9c054c5000cfbe3780"),
			MustHexID("60cabadcf1041d96bab0315723e9efab633552e5cb3ccb583dc2f9186111286020d820861ae0108153aea889e6bfdb94248bd70a326e306580"),
			MustHexID("2ad3786ddd56ea2bcabd364389e60c357348c51e714d296bb1e7e299c0d599521f8a0a670c1f24a21a8929783510ae91878727763923cbbb00"),
			MustHexID("50b2b65d7bd7363217fe52491841f4cc8dec39bea51df3e0f7724a40f066290769226acf2b1bc4bea6441b5f53dac85abd7c0eaf2b3eb1ae80"),
			MustHexID("ca275e82f4cba98b6ce88f5421debd2ba2ecfe218d4dd3e7f1ebf8cf148b7dfe328f982eca46d1a6f030497117897a43cbbe2a1e5d4338b080"),
		},
		251: []NodeID{
			MustHexID("7f83d513b028638ae005f0db16945b0523c96a7a02a223bd04a5d372d499ecf7c7d3ae954f809145fbd33adfdef044ed02a7515d9223cb1180"),
			MustHexID("b7b396697102d30b92acc607167a67114c059fb78e94e76798818bad57a831b1e07bc1b848411b332d9dc6613264d50784f845d99dcdbe0280"),
			MustHexID("e2cb1ba295ad7dc584c34233d95d96652d64ae064834be140dbb7fc22bf1d3c95016bc576ab0107988b407f76bbb1d8615678dd8a014aae980"),
			MustHexID("4619a2a33a77dbe167651d1a9ed9e70e5a87e22af89d881942cdd57463d21bf5125de5efb5b4f29856382087ab652c8618fddbba8f33419c80"),
			MustHexID("ec0595444f37322e760d3bcb26c6f187cae4b0817568584c744332b8d65bda546db91a27a0178893e7513e3379653fec8ef6ccf7f07f50ab00"),
			MustHexID("391c5e14f9dd84d1d43b9707afdd8d8619686b965800bb0a89faf782215b2788a30d636ef34a0efe41237ca78e0e8090810b1d8e5cd609b100"),
			MustHexID("b3bb4651fa69cd36c2aa48e16e1d07977a10a8e76a5f73a52de2d109db595ccbdb29a8d76bab1d87c1ddb1eb48e635a01e0c15d310c7499d00"),
			MustHexID("0320e68c42288db0966c1811ccf10a2d2a14c6b5e1afd4cc11829a78503e8012113390db3761720ec620147bcad2d4a4baa1e85db8ee5bbb80"),
			MustHexID("fae8cf1fbec62a4e4189a43b735775e35e1aeef3b4fda24e50f51cc5d4ced22fb7c6599b8f9d7b451110f010887448642b0472e49dd4782d00"),
			MustHexID("b796c916ad516f9319ae98d480d34e23ad5c0aecde989208f25f135da03e86ad66380810def35827abf5f2760cf9ca9ffc47f527dd90458e80"),
			MustHexID("ca24e25e235aee3893fa79df23563847697b852d6ee8f26c31d5d176bdc16eac113573c61891c06df962e16d50d0492446db10ff4c52d7b700"),
			MustHexID("99fefda30b099f9ae37b51c79768aa3b6b0b57c2de258b4e2a4b8034aa4ae991fd1c62bed1134123f2c5cf9c6aff7108724adb0dbd69d33880"),
			MustHexID("9b4ca296ed5e7db1b2d4a6880bedaa40cab7400db3391e02d0e959917c3dd5c91ad5c99b94c238499fd54157e486aee7f30c3cf1cd006c1100"),
			MustHexID("5dca7bd6d94dbd173e42009405fa182b21a00688d850d6944ee71c3e44e3b67f6b272f3075eb3e95fab6f2f34cb9696bfa688c9ba2cbfef500"),
			MustHexID("123c31b4b3a0368a8b7f55b6e13500c244f4f62258d3c17f11715d41f57aab917ab8d6131902c7132061bc9abf575c0dc3a4b154da63d52c80"),
			MustHexID("10a3094ce1fc37fc39aa0afa14db0b14d9a1e3e2a51c05da393c5b555d7cc0af627e32a02f52c38988cdfd55f16c0f45a35f3c04f788f96d80"),
		},
		252: []NodeID{
			MustHexID("80e041ac61ce9f8f60b39e9c2a44e3ebe6bd6ed378bacb2e5b7e2b2b2a16a28cc9aeeb5413ab14dd70779d65d901491dbdbdc884e3f7e39880"),
			MustHexID("7b64aa84e0e827194833805ec7820405a8159cc04a673e74552a37ae2a6e9c96cc331f222e169335a80bb0a27c937f1264f784aee4c6367f80"),
			MustHexID("18bbf9b46208a0c7b965a0fc2ea70fdb75c58d2c93a671bcd704cfb173abe5742319d590aba8d35a9601fcc7eab802a6a2f36c81d23478c080"),
			MustHexID("3aa9aeb600e721474b3302405eab83a0b100259bb2bab307cb64387c606ed1c771d351642dfc213b9dcb6ad5e9a80449cc6792a6d76d704280"),
			MustHexID("b53320680c0775e19a5f3ad09dbd818d91c75d6f4500ceec1b0a0e2cf9fc6b3e8f0ffe15bc2205d6af1650ca0e849e9b0d2e8543d79b6d3600"),
			MustHexID("211003ed3de0fdd5999c63d32dcbced81cd9263e813587fcd6abdc76272083c1feb9565dae8274d06de9dfa40783605c89bfeea1e4b85a6c00"),
			MustHexID("23be80192ac579f8a80ff2a2248aef767d193d737217a2167978cc9d5ff95d27a91163c1b26c3a65a95294a888eb208358f9d5a087f4d11d00"),
			MustHexID("75a85ee1eacbd737b290420c629bce8ce013038133fb3e7985df0497a1d5b5e09d364babb536eb4b5bde55ba23b2451917cdbcbc60ea300100"),
			MustHexID("c8166bfb126d420c669d21cfc8b1fa24621bb7846c157e8c9ac584d0c1e19ec3a95e0be92da55e6820e7f51c8494032f60daeb68c28f390200"),
			MustHexID("e34c78b5f875bf83171d182c4c642475c28a2b28f932a91381e3da04f3f59628476085959fe61297455de9efd087bf9b23640d0e13a26c2680"),
			MustHexID("96949a021099b402aebf87e13f270cbbdb1ae2d378f5bd8d40702cfd7a45b229c32c7eb447569dbb3bd88f654b7cca60cf3ccb51d79e7d0b80"),
			MustHexID("9230a904ecf52ecf6dc80792f477565fa76b22b45c3df2bc8a0eb5888837f33c387b34dc933a36313aa691657cc3b1bb0cf7713b7627cf2c80"),
			MustHexID("0890cca4750d420f470ee6d8f64fb7a5d7a8f7a48c32ece8b0c01909f4c0decb3e6502c8ba6e508fc87c79ce0ec0629e1ca0bd746a43da2a00"),
			MustHexID("fcce67fff53ec684e4abebd8c57e873f856d4d6efdf120ff4d59f215c957c79db417bcc2a34d12b414b5cc9be246ee67839d373707daa06f80"),
			MustHexID("2d9a653c39340c2acafad188b8079c7eb929bea9c2fe6718d6a0f1e910b04bf6eaa04b95a2340f1819c9521575d9cb3a26fd0e211dc968ac00"),
			MustHexID("4d73598f40a7b616423ab971914b7eca399fe58941c1057cb87d284a776e083f3fe8032ded5c6e966545293cfc92a5b1b69866b9203dcd4080"),
		},
		253: []NodeID{
			MustHexID("82295855a0eac84c092e5e18cd8c722729bb2af1830c5b53a3b954c62fa5067cd046896e6d6d12e3883f7ea15459f96090f91447028bce1200"),
			MustHexID("386f361a0db2b5da60ca36ffee0f8c11ea419773f8927691c73ffffef2917efd5a6681cc8d22c469102d3319ef3e6bb62e190181c1bef3c100"),
			MustHexID("740d3c3ad22860cd1d0b205583bd99090bf500d8fa7b4454e181980b688a4a1ff821b50159861dd625ea634b3346ab52a8b527d0cfed4e1700"),
			MustHexID("a15a65ac76a6e5a5bd5ad0ed08fe1b5fa9d4e3d4a0da3cfbff361173f50738a766a2ea7cfd54a2a907b8ddaa7cf2e6b1a13643c00e1d62a500"),
			MustHexID("9f7ab3544989d17abddbdafbd92f39e94a6933bbf0f70fbed7ee21f11c463caf7e35a26b7f24c6a97b8e797e1d848bb19b1a40e191bbb32100"),
			MustHexID("e4c605d57ab5e3c7b425e4082b09d6fa6c2c878dcc4bfd9e65bacb186d391339cebbd2aa4235f2d6316610788c54c0e4b0e554899700e3bb00"),
			MustHexID("707535ceba86582c88daf48ece1aaf807ececcd0762a891587a5189aed2d9155060b9ad285fc106660b67ba8b5a67c848b09dba830876a7b00"),
			MustHexID("59870e6a3a40e8a8bc251deb9c679a1f25355d803e2facb520f0ed5dd00f869712b546d0a52f76dfa7c786e0842e0b858a8b90055616375780"),
			MustHexID("3fbba6299952bf02e2f02f09bc907882e3bc2013a4535203a538d09d04c29d7110f7ac5a0f55a813485c11149146ca2dd62efce3d290045400"),
			MustHexID("4b6efd38678c29dcac5815b718a5290db8f951e8749f26c9219d022a2fc4cf93405037310238c9394ad0860eb7ddea813c5ce8e306e8941780"),
			MustHexID("40371e3d4213cd9145ec4e96032142fd73c4c02532109b8791b1800e86012e68315798a33acaa6bdce0a301454f1ccb6cd17a2dce2b3faec00"),
			MustHexID("0337c6c4f234d6f7b664ee14e167172e4404610e4f68fca7243f90b8a67cf1ac765664ed792948a5ff87dced9694d863e39c504bf513c1f380"),
			MustHexID("517e726844b8b972699ee2544e5b8703f2fd92091d81088ae97eb3199d3ead1a77b2c862b54ad89f1130b792465f1a8a497a3d40bfae150400"),
			MustHexID("6d1d23fafd4065c8a7deafc64ae63b4b4b3417367d5a3db096113bfc6d82b80e677bcdbb427ac4c9d8a175806833b6859041e62390068fb500"),
			MustHexID("da0f16f147ec0f1e4b4f28c82bf563937b14330704a0fcdfface773e5b729814a99d7d9fda2544919cbe629e68c2d46cf8cc56658547096e00"),
			MustHexID("36bb38dff5126d370b048d097def744bd8e7c3f3157c2f0b617991b2f4cc2b3f6b5d29e6b867e80623f6e28f9ef1f468e3eff00b4695059080"),
		},
		254: []NodeID{
			MustHexID("4ba6ab8ceb04e4b4c35bba7cabe63b8859b35b0ae7f59940cf66b67f69ecd917d0811dbbca0d6a16c4d263f0b7caaa92702003a321168db980"),
			MustHexID("456d5667e2bd29ecc4b1f3aec7ba46bc2115624c24766aaf710fc8308cf03fdacf399918db20f347ecc17aab23c1999daf8f38494d85bc1500"),
			MustHexID("77a9253fbf2dd516f07b577aefeab49634be45250e3c65255b87823b02a815f46db473dbee4cf3e7ef9bcfd9a5fde21a1ac70bc73584c3c200"),
			MustHexID("f7312556f3ac6642596192dcca2a99327703b2784a6c4c3cfe0285a2755ef8dc87d3bd7828f6bc7cfc8e1d8e31d7d8e785227d81c1dac54800"),
			MustHexID("d9d1d1e8c065251fe23ccd5894574b8cdcd02b4c28c6f05c65e08fb781225ef386119dd917a4761087ce493d7b27f31b0f176e20b3923b9880"),
			MustHexID("7d3cd46c1fd840ea7e91913c1c6303e6eee9172b2750ac763830875f759918f951529857be785b8882dee8285a03bab55e2fe0e6d2a79a8c00"),
			MustHexID("3d6d0abc0cec9988a59a4682062dbf78482a7abf995c6312637f6d4f6855f062ce5b6f03f101c144d31afe837f27a5b9822f299b2e2adfc100"),
			MustHexID("493c46be93d215f8457911db968cdc79ede31d0c3522466a2167a7495562a5ce8dfe5de1230b9f5852dee94738e0bb633020efc3c57eb33d00"),
			MustHexID("1f18f4d73ee84b62841ccb3f936b56387ee20d0f3b3b174c73c3a5bfe7096926cfd7b6b7db41e1778ca52ac38495db6d5f16cec5a5e60f1600"),
			MustHexID("debb2da75a8f919be477e9902948e57bfb61385249d972bdaed9de37fbd92ffea109a0b73933c1fb5710a21bbbb30b0fd7912a3e6499a37d00"),
			MustHexID("e955693e07e0cd7061c28a2546047701adbd440ad725b85183617bdbb77078c588aa67d1e232804b1f2b206113c4195998f8bdc225761c1d00"),
			MustHexID("3af2159c64b0237b4fada2f1da72f6a86626ac8028977fa27903d6db39522e80b348681e356e120db061647fe91f9042eddd9cf6740044b280"),
			MustHexID("3cad0bade65d942a67ea4d10d416bdd2a5f8e49783e576d05730706d757982dafa97606519bfcad3354716318edb6de1edd72646cae7b08880"),
			MustHexID("49e0a2de00080398df5f7086a43610191f733d031637c4118ba0c54a82cf1ba75ba49dc35cd98f38aefca80018626a0d917cc2ddf8cab3fe80"),
			MustHexID("6437f2e04a732e3ba76b43c5c8d66735fdf07facaffc76e9b717788b2c4c963451e1b34c229d34cb68174079d8c6de60f719bc2b1a715d0080"),
			MustHexID("02b2380337e16c80acb078c87a6f6e8c2adc21b03fbe7b0ac6f0e3bb29c26879009e633361c9c0e9eb0094b53e05d8c45da595e77f47d52580"),
		},
		255: []NodeID{
			MustHexID("8a6392fc33f4739ef446e63607eb3179f68a6cea194bf846d8aa56162ea1cdc4dcaea610c333fd7a415919d9c02d41d8175fb2f47c4bb36e00"),
			MustHexID("5d06a202e8e8420582f54dc392bbdb23a1b9e77434c57b8f06e3144e42678cabaf2b443c591a50a67f1c9528659c1d2a6767495d0567f95a80"),
			MustHexID("76ac2388ec40047777d54d8095cae3612ad3f465e3d15c3ad3ce18e518ba03fae5efb9a20d5205cc074a7bb517c01a06eac474b7b21aa72e80"),
			MustHexID("a0d524a0e1c73e69580a6cb0260e828fe99090925b35b073f47842db0bfd00e1ac88ac320d776acc4559a24b3f670ed0317afb2eb3991aa200"),
			MustHexID("950b5f9654d5790c0e1f56c431f195e2d04b4e30fe3a9e9c18f4e9cfb5e96f08783bfd11642eba2e8a080da454e9dbc86d5756a3d2e8c49b80"),
			MustHexID("444919079160090d4a50bd3aae4041ec32f36fa13ba538505da604dca37e0410f5b0177866c6fa401ce5adbc559777b393a8902601311f4b00"),
			MustHexID("4223d62d77b1d73326b4b6b93f1745aabef88bd7ede81073910e343a01868ad215303254b9c1cd8f126f7312d09e608cf732b334e6bd4a8d00"),
			MustHexID("cf06370c223d5512b98d54a3c7b8055fcdbcbadf25815e35a40c29801e47a51c99afe9bad11cb1155d1d84362407b111b665fe3bac0ed57280"),
			MustHexID("5a6d0ba71c6ca7babf02f7ec69c3142b49fbe86d1e07591f1639a0bb98e5da731ab1b47c23cc8bc372f0ed17a37324c8784c2cc4de01d2cb80"),
			MustHexID("44eebf0d462b5c6bdcab5d6d779b35fbb97400698472fd9cb2f25bf5736c48cb91357a38f7296c9a05f5249e81ea34dc87dc74268147ff2f80"),
			MustHexID("6f4b028cc7b5a161f35c6b2c96bc34a596a8d8cafbe6f2da8c12ad81c0eb17f8ef32124093ed4cebd141579c3daf94a0c47471d23c0ae37800"),
			MustHexID("8b8900955b88a9146087b17a54b226c47c4dc0f1a5150ce27cc0f6e049bc55922cb746b62b8a1166a122862993876d7da7fd19a8904e539780"),
			MustHexID("25a169fc0fabcffc0de61fdfad885757d96c83cf7f6d5126c2f54c4c6faa4a010ba72b23ca02b15725bc8d9b53fcdb2c65fcc32fd51b8db080"),
			MustHexID("18e27093f0c817b0649aab53d219dbd38b6ce23bf2a87cc57a600ca42c74f324cc2bbcbac66b14bab39fce48798bf1f69e4cc0c7fa397ee780"),
			MustHexID("1f9d77cdc252ee43f14c04af25cefa7706ea0af96f36a12f8953c182386011d3f77e87ce466d44b9d1ca0c8fcd988e8174b0e15de7f59afe00"),
			MustHexID("33187bbdc92af317f66bdd31b558130f1c070132a8326a3b3203d5d1326d91bdd1a7b3816429f544c7641c69f84b580c5edb6f16a6ca708d00"),
		},
		256: []NodeID{
			MustHexID("44b807263a87c731ee7827ccf49b9317c1542240a273d828e8b7ae645ac08a1b2e3f626d541efb4d7517674fb1728bda9b2ec36e1f381e6c00"),
			MustHexID("2ebb7ab548a3d35a3344f99c567bdb72062a2bf80ccb4354e4330b0702e4a8a64d355a22889d260b7bc0c08e83f409c0a84340792410011280"),
			MustHexID("75a00897d96084b34cfa66f6a6e98a64ebe3eab38b00cb084ffc592050031aea10929447cadce43a59158ea2351b7ec3ef8bdea847ec65d480"),
			MustHexID("a76ec284683d672923eec1c4c63f1e7206b893347ab50047d43cdb885dd302a52b38b813944b777ed109e75e520ea4a8b13248c59cd3d45a80"),
			MustHexID("b726778307ed459472f61559c1aa1041b58cecb8426c8bb7c8d1a0e05f0f41b0b3490ca9c20444a8800ac063a8fde01d8bd18777ed71f04280"),
			MustHexID("04df13fc466efa1fada1ce5f8998c8c5572014774876181af4f0ab5ebf1bbafd363bdd04a010ab485daa52f59e0b7b8bed7a848ab3570dee00"),
			MustHexID("9f2d7e02c66de6921e8cc51c6f5ebf3bfdeb925a1d47101e1713534fc296266a6840903b24a9b70d7a943ba3aa1efd12efeb4fe9823a232d80"),
			MustHexID("dfed1ac84875d34cd82ebfca936256cf9fb08b2bcf38df640a33596c003e5f1d04820bf5dc8be681c134dcd88bf0784b78cd0c009a9b48e080"),
			MustHexID("4320cb39c3d5f0e2d68874114175a76a5f731b75b7592795ae25c72fb768e74c301a62542aa8636d78425c16360999bf96b4263c7e4e5ebc00"),
			MustHexID("0ced62661eec316b47ad282b0dab33e6ce8f9700e79372af12798526f8e841d6c61a9203dd88d2c3e694c772852f04b6d76f9b807dce7e1c00"),
			MustHexID("a9c66d51937ef1f62ed0592d76f170e8079d19e4086c1cc54dcfed13580f9f2c62556a013a626d49da84812b2558dc6544db604670bc608780"),
			MustHexID("aa54e2657213be6ad995c4528a7b22ac6f76235f713851de12f236a4ebbb24647f220341f9d9c8c641d48c999b94e812755ed32ecf8deb3b80"),
			MustHexID("fb1f701d7ecfe04947804a265eaf292040e0fc0521c7e7fffc13f8a9df3b6d581f4bdb3b61f28f95f82e22f6a863983b197a82dcb6a3954480"),
			MustHexID("f15d4c8d03e17933f49713da4ed3c06aad1889726b7e113745bec8969d0399d5c831dd3315f09f0ff445bb1890e442eb7f88a76f51be614180"),
			MustHexID("831d945e354bdd0677738134e3195ab77ae0a095a4d19fa24b6d8ec74539cbbd3e005db1fa09988a074624c152351235a68d7bb65f89aa0f00"),
			MustHexID("c46abf83d9bb77840fad64b66acceaa16676a59b46749209aad9e8d5ed2b3d5868a80b8ca194e26cb7f2b61bb1eeb1aaf9bfcdbf1d85fe6280"),
		},
	},
}

type preminedTestnet struct {
	target    NodeID
	targetSha common.Hash // sha3(target)
	dists     [hashBits + 1][]NodeID
	net       *Network
}

func (tn *preminedTestnet) sendFindnodeHash(to *Node, target common.Hash) {
	// current log distance is encoded in port number
	// fmt.Println("findnode query at dist", toaddr.Port)
	if to.UDP <= lowPort {
		panic("query to node at or below distance 0")
	}
	next := to.UDP - 1
	var result []rpcNode
	for i, id := range tn.dists[to.UDP-lowPort] {
		result = append(result, nodeToRPC(NewNode(id, net.ParseIP("10.0.2.99"), next, uint16(i)+1+lowPort)))
	}
	injectResponse(tn.net, to, neighborsPacket, &neighbors{Nodes: result})
}

func (tn *preminedTestnet) sendPing(to *Node, addr *net.UDPAddr, topics []Topic) []byte {
	injectResponse(tn.net, to, pongPacket, &pong{ReplyTok: []byte{1}})
	return []byte{1}
}

func (tn *preminedTestnet) send(to *Node, ptype nodeEvent, data interface{}) (hash []byte) {
	switch ptype {
	case pingPacket:
		injectResponse(tn.net, to, pongPacket, &pong{ReplyTok: []byte{1}})
	case pongPacket:
		// ignored
	case findnodeHashPacket:
		// current log distance is encoded in port number
		// fmt.Println("findnode query at dist", toaddr.Port-lowPort)
		if to.UDP <= lowPort {
			panic("query to node at or below  distance 0")
		}
		next := to.UDP - 1
		var result []rpcNode
		for i, id := range tn.dists[to.UDP-lowPort] {
			result = append(result, nodeToRPC(NewNode(id, net.ParseIP("10.0.2.99"), next, uint16(i)+1+lowPort)))
		}
		injectResponse(tn.net, to, neighborsPacket, &neighbors{Nodes: result})
	default:
		panic("send(" + ptype.String() + ")")
	}
	return []byte{2}
}

func (tn *preminedTestnet) sendNeighbours(to *Node, nodes []*Node) {
	panic("sendNeighbours called")
}

func (tn *preminedTestnet) sendTopicNodes(to *Node, queryHash common.Hash, nodes []*Node) {
	panic("sendTopicNodes called")
}

func (tn *preminedTestnet) sendTopicRegister(to *Node, topics []Topic, idx int, pong []byte) {
	panic("sendTopicRegister called")
}

func (*preminedTestnet) Close() {}

func (*preminedTestnet) localAddr() *net.UDPAddr {
	return &net.UDPAddr{IP: net.ParseIP("10.0.1.1"), Port: 40000}
}

func injectResponse(net *Network, from *Node, ev nodeEvent, packet interface{}) {
	go net.reqReadPacket(ingressPacket{remoteID: from.ID, remoteAddr: from.addr(), ev: ev, data: packet})
}

// mine generates a testnet struct literal with nodes at
// various distances to the given target.
func (tn *preminedTestnet) mine(target NodeID) {
	// Clear existing slices first (useful when re-mining).
	for i := range tn.dists {
		tn.dists[i] = nil
	}

	tn.target = target
	tn.targetSha = crypto.SHA3Hash(tn.target[:])
	found := 0
	for found < bucketSize*10 {
		k := newkey()
		id := PubkeyID(k.PublicKey())
		sha := crypto.SHA3Hash(id[:])
		ld := logdist(tn.targetSha, sha)
		if len(tn.dists[ld]) < bucketSize {
			tn.dists[ld] = append(tn.dists[ld], id)
			fmt.Println("found ID with ld", ld)
			found++
		}
	}
	fmt.Println("&preminedTestnet{")
	fmt.Printf("	target: %#v,\n", tn.target)
	fmt.Printf("	targetSha: %#v,\n", tn.targetSha)
	fmt.Printf("	dists: [%d][]NodeID{\n", len(tn.dists))
	for ld, ns := range &tn.dists {
		if len(ns) == 0 {
			continue
		}
		fmt.Printf("		%d: []NodeID{\n", ld)
		for _, n := range ns {
			fmt.Printf("			MustHexID(\"%x\"),\n", n[:])
		}
		fmt.Println("		},")
	}
	fmt.Println("	},")
	fmt.Println("}")
}
