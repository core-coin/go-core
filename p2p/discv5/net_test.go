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
	"crypto/rand"
	eddsa "github.com/core-coin/go-goldilocks"
	"net"
	"testing"
	"time"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
)

func TestNetwork_Lookup(t *testing.T) {
	key, _ := crypto.GenerateKey(rand.Reader)
	pub := eddsa.Ed448DerivePublicKey(*key)
	network, err := newNetwork(lookupDevin, pub, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	lookupDevin.net = network
	defer network.Close()

	// lookup on empty table returns no nodes
	// if results := network.Lookup(lookupDevin.target, false); len(results) > 0 {
	// 	t.Fatalf("lookup on empty table returned %d results: %#v", len(results), results)
	// }
	// seed table with initial node (otherwise lookup will terminate immediately)
	seeds := []*Node{NewNode(lookupDevin.dists[256][0], net.IP{10, 0, 2, 99}, lowPort+256, 999)}
	if err := network.SetFallbackNodes(seeds); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)

	results := network.Lookup(lookupDevin.target)
	t.Logf("results:")
	for _, e := range results {
		t.Logf("  ld=%d, %x", logdist(lookupDevin.targetSha, e.sha), e.sha[:])
	}
	if len(results) != bucketSize {
		t.Errorf("wrong number of results: got %d, want %d", len(results), bucketSize)
	}
	if hasDuplicates(results) {
		t.Errorf("result set contains duplicate entries")
	}
	if !sortedByDistanceTo(lookupDevin.targetSha, results) {
		t.Errorf("result set not sorted by distance to target")
	}
	// TODO: check result nodes are actually closest
}

// This is the test network for the Lookup test.
// The nodes were obtained by running devin.mine with a random NodeID as target.
var lookupDevin = &preminedDevin{
	target:    MustHexID("166aea4f556532c6d34e8b740e5d314af7e9ac0ca79833bd751d6b665f12dfd38ec563c363b32f02aef4a80b44fd3def94612d497b99cb5faa"),
	targetSha: crypto.SHA3Hash(common.Hex2Bytes("166aea4f556532c6d34e8b740e5d314af7e9ac0ca79833bd751d6b665f12dfd38ec563c363b32f02aef4a80b44fd3def94612d497b99cb5faa")),
	dists: [257][]NodeID{
		240: {
			MustHexID("2001ad5e3e80c71b952161bc0186731cf5ffe942d24a79230a0555802296238e57ea7a32f5b6f18564eadc1c65389448481f8c9338df0a3daa"),
			MustHexID("6ba3f4f57d084b6bf94cc4555b8c657e4a8ac7b7baf23c6874efc21dd1e4f56b7eb2721e07f5242d2f1d8381fc8cae535e860197c6923679aa"),
		},
		244: {
			MustHexID("696ba1f0a9d55c59246f776600542a9e6432490f0cd78f8bb55a196918df2081a9b521c3c3ba48e465a75c10768807717f8f689b0b4adce0aa"),
		},
		246: {
			MustHexID("d6d32178bdc38416f46ffb8b3ec9e4cb2cfff8d04dd7e4311a70e403cb62b10be1b447311b60b4f9ee221a8131fc2cbd45b96dd80deba68aaa"),
			MustHexID("3ea3d04a43a3dfb5ac11cffc2319248cf41b6279659393c2f55b8a0a5fc9d12581a9d97ef5d8ff9b5abf3321a290e8f63a4f785f450dc8a6aa"),
			MustHexID("2fc897f05ae585553e5c014effd3078f84f37f9333afacffb109f00ca8e7a3373de810a3946be971cbccdfd40249f9fe7f322118ea459ac7aa"),
		},
		247: {
			MustHexID("3155e1427f85f10a5c9a7755877748041af1bcd8d474ec065eb33df57a97babf54bfd2103575fa829115d224c523596b401065a97f740106aa"),
			MustHexID("312c55512422cf9b8a4097e9a6ad79402e87a15ae909a4bfefa22398f03d20951933beea1e4dfa6f968212385e829f04c2d314fc2d4e255eaa"),
			MustHexID("38643200b172dcfef857492156971f0e6aa2c538d8b74010f8e140811d53b98c765dd2d96126051913f44582e8c199ad7c6d6819e9a56483aa"),
			MustHexID("8dcab8618c3253b558d459da53bd8fa68935a719aff8b811197101a4b2b47dd2d47295286fc00cc081bb542d760717d1bdd6bec2c37cd72eaa"),
			MustHexID("8b58c6073dd98bbad4e310b97186c8f822d3a5c7d57af40e2136e88e315afd115edb27d2d0685a908cfe5aa49d0debdda6e6e63972691d6baa"),
			MustHexID("2cbb718b7dc682da19652e7d9eb4fefaf7b7147d82c1c2b6805edf77b85e29fde9f6da195741467ff2638dc62c8d3e014ea5686693c15ed0aa"),
			MustHexID("e84027696d3f12f2de30a9311afea8fbd313c2360daff52bb5fc8c7094d5295758bec3134e4eef24e4cdf377b40da344993284628a7a346eaa"),
			MustHexID("f1357a4f04f9d33753a57c0b65ba20a5d8777abbffd04e906014491c9103fb08590e45548d37aa4bd70965e2e81ddba94f31860348df0146aa"),
			MustHexID("4ab0a75941b12892369b4490a1928c8ca52a9ad6d3dffbd1d8c0b907bc200fe74c022d011ec39b64808a39c0ca41f1d3254386c3e7733e70aa"),
			MustHexID("d45150a72dc74388773e68e03133a3b5f51447fe91837d566706b3c035ee4b56f160c878c6273394daee7f56cc398985269052f22f75a805aa"),
		},
		248: {
			MustHexID("6aadfce366a189bab08ac84721567483202c86590642ea6d6a14f37ca78d82bdb6509eb7b8b2f6f63c78ae3ae1d8837c89509e41497d719baa"),
			MustHexID("a605ecfd6069a4cf4cf7f5840e5bc0ce10d23a3ac59e2aaa70c6afd5637359d2519b4524f56fc2ca180cdbebe54262f720ccaae8c1b28fd5aa"),
			MustHexID("29701451cb9448ca33fc33680b44b840d815be90146eb521641efbffed0859c154e8892d3906eae9934bfacee72cd1d2fa9dd050fd18888eaa"),
			MustHexID("3ed426322dee7572b08592e1e079f8b6c6b30e10e6243edd144a6a48fdbdb83df73a6e41b1143722cb82604f2203a32758610b5d9544f44aaa"),
			MustHexID("b2e2a2b7fdd363572a3256e75435fab1da3b16f7891a8bd2015f30995dae665d7eabfd194d87d99d5df628b4bbc7b04e5b492c596422dd82aa"),
			MustHexID("0c69c9756162c593e85615b814ce57a2a8ca2df6c690b9c4e4602731b61e1531a3bbe3f7114271554427ffabea80ad8f36fa95a49fa77b67aa"),
			MustHexID("8d28be21d5a97b0876442fa4f5e5387f5bf3faad0b6f13b8607b64d6e448c0991ca28dd7fe2f64eb8eadd7150bff5d5666aa6ed868b84c71aa"),
			MustHexID("2c677e1c64b9c9df6359348a7f5f33dc79e22f0177042486d125f8b6ca7f0dc756b1f672aceee5f1746bcff80aaf6f92a8dc0c9fbeb259b3aa"),
			MustHexID("3994880f94a8678f0cd247a43f474a8af375d2a072128da1ad6cae84a244105ff85e94fc7d8496f639468de7ee998908a91c7e33ef7585ffaa"),
			MustHexID("b45a9153c08d002a48090d15d61a7c7dad8c2af85d4ff5bd36ce23a9a11e0709bf8d56614c7b193bc028c16cbf7f20dfbcc751328b64a924aa"),
			MustHexID("057ab3a9e53c7a84b0f3fc586117a525cdd18e313f52a67bf31798d48078e325abe5cfee3f6c2533230cb37d0549289d692a29dd400e899baa"),
			MustHexID("0ddf663d308791eb92e6bd88a2f8cb45e4f4f35bb16708a0e6ff7f1362aa6a73fedd0a1b1557fb3365e38e1b79d6918e2fae2788728b70c9aa"),
			MustHexID("f637e07ff50cc1e3731735841c4798411059f2023abcf3885674f3e8032531b0edca50fd715df6feb489b6177c345374d64f4b07d257a774aa"),
			MustHexID("e24ec7c6eec094f63c7b3239f56d311ec5a3e45bc4e622a1095a65b95eea6fe13e29f3b6b7a2cbfe40906e3989f17ac834c3102dd0cadaaaaa"),
			MustHexID("b76ea1a6fd6506ef6e3506a4f1f60ed6287fff8114af6141b2ff13e61242331b54082b023cfea5b3083354a4fb3f9eb8be01fb4a518f579eaa"),
			MustHexID("9b53a37950ca8890ee349b325032d7b672cab7eced178d3060137b24ef6b92a43977922d5bdfb4a3409a2d80128e02f795f9dae6d7d99973aa"),
		},
		249: {
			MustHexID("675ae65567c3c72c50c73bc0fd4f61f202ea5f93346ca57b551de3411ccc614fad61cb9035493af47615311b9d44ee7a161972ee4d77c28faa"),
			MustHexID("8eb81408389da88536ae5800392b16ef5109d7ea132c18e9a82928047ecdb502693f6e4a4cdd18b54296caf561db937185731456c456c98baa"),
			MustHexID("2adba8b1612a541771cb93a726a38a4b88e97b18eced2593eb7daf82f05a5321ca94a72cc780c306ff21e551a932fc2c6d791e4681907b5caa"),
			MustHexID("b1b4bfbda514d9b8f35b1c28961da5d5216fe50548f4066f69af3b7666a3b2e06eac646735e963e5c8f8138a2fb95af15b13b23ff00c6986aa"),
			MustHexID("d2139281b289ad0e4d7b4243c4364f5c51aac8b60f4806135de06b12b5b369c9e43a6eb494eab860d115c15c6fbb8c5a1b0e382972e0e460aa"),
			MustHexID("4a693df4b8fc5bdc7cec342c3ed2e228d7c5b4ab7321ddaa6cccbeb45b05a9f1d95766b4002e6d4791c2deacb8a667aadea6a700da28a3eeaa"),
			MustHexID("ab41611195ec3c62bb8cd762ee19fb182d194fd141f4a66780efbef4b07ce916246c022b841237a3a6b512a93431157edd221e854ed2a259aa"),
			MustHexID("68e8e26099030d10c3c703ae7045c0a48061fb88058d853b3e67880014c449d4311014da99d617d3150a20f1a3da5e34bf0f14f1c51fe4ddaa"),
			MustHexID("3fbcacf546fb129cd70fc48de3b593ba99d3c473798bc309292aca280320e0eacc04442c914cad5c4cf6950345ba79b0d51302df88285d4eaa"),
			MustHexID("1d4a623659f7c8f80b6c3939596afdf42e78f892f682c768ad36eb7bfba402dbf97aea3a268f3badd8fe7636be216edf3d67ee1e08789ebbaa"),
			MustHexID("a283c474ab09da02bbc96b16317241d0627646fcc427d1fe790b76a7bf1989ced90f92101a973047ae9940c92720dffbac8eff21df8cae46aa"),
			MustHexID("dbf7e5ad7f87c3dfecae65d87c3039e14ed0bdc56caf00ce81931073e2e16719d746295512ff7937a15c3b03603e7c41a4f9df94fcd37bb2aa"),
			MustHexID("caaa070a26692f64fc77f30d7b5ae980d419b4393a0f442b1c821ef58c0862898b0d22f74a4f8c5d83069493e3ec0b92f17dc1fe6e4cd437aa"),
			MustHexID("874cc8d1213beb65c4e0e1de38ef5d8165235893ac74ab5ea937c885eaab25c8d79dad0456e9fd3e9450626cac7e107b004478fb59842f06aa"),
			MustHexID("d94193f236105010972f5df1b7818b55846592a0445b9cdc4eaed811b8c4c0f7c27dc8cc9837a4774656d6b34682d6d329d42b6ebb55da1daa"),
			MustHexID("edd9af6aded4094e9785637c28fccbd3980cbe28e2eb9a411048a23c2ace4bd6b0b7088a7817997b49a3dd05fc6929ca6c7abbb69438dbdaaa"),
		},
		250: {
			MustHexID("53a5bd1215d4ab709ae8fdc2ced50bba320bced78bd9c5dc92947fb402250c914891786db0978c898c058493f86fc68b1c5de8a5cb363361aa"),
			MustHexID("b7f79e3ab59f79262623c9ccefc8f01d682323aee56ffbe295437487e9d5acaf556a9c92e1f1c6a9601f2b9eb6b027ae1aeaebac71d61b9baa"),
			MustHexID("d374bf7e8d7ffff69cc00bebff38ef5bc1dcb0a8d51c1a3d70e61ac6b2e2d6617109254b0ac224354dfbf79009fe4239e09020c483cc60c0aa"),
			MustHexID("1e1eac1c9add703eb252eb991594f8f5a173255d526a855fab24ae57dc277e055bc3c7a7ae0b45d437c4f47a72d97eb7b126f2ba344ba6c0aa"),
			MustHexID("ae28953f63d4bc4e706712a59319c111f5ff8f312584f65d7436b4cd3d14b217b958f8486bad666b4481fe879019fb1f767cf15b3e3e2711aa"),
			MustHexID("934bb1edf9c7a318b82306aca67feb3d6b434421fa275d694f0b4927afd8b1d3935b727fd4ff6e3d012e0c82f1824385174e8c6450ade59caa"),
			MustHexID("9eef3f28f70ce19637519a0916555bf76d26de31312ac656cf9d3e379899ea44e4dd7ffcce923b4f3563f8a00489a34bd6936db0cbb4c959aa"),
			MustHexID("82200872e8f871c48f1fad13daec6478298099b591bb3dbc4ef6890aa28ebee5860d07d70be62f4c0af85085a90ae8179ee8f937cf37915caa"),
			MustHexID("6c75a5834a08476b7fc37ff3dc2011dc3ea3b36524bad7a6d319b18878fad813c0ba76d1f4555cacd3890c865438c21f0e0aed1f80e0a157aa"),
			MustHexID("995b873742206cb02b736e73a88580c2aacb0bd4a3c97a647b647bcab3f5e03c0e0736520a8b3600da09edf4248991fb01091ec7ff3ec7cdaa"),
			MustHexID("c773a056594b5cdef2e850d30891ff0e927c3b1b9c35cd8e8d53a1017001e237468e1ece3ae33d612ca3e6abb0a9169aa352e9dcda358e5aaa"),
			MustHexID("2b46a5f6923f475c6be99ec6d134437a6d11f6bb4b4ac6bcd94572fa1092639d1c08aeefcb51f0912f0a060f71d4f38ee4da70ecc16010b0aa"),
			MustHexID("af6ab501366debbaa0d22e20e9688f32ef6b3b644440580fd78de4fe0e99e2a16eb5636bbae0d1c259df8ddda77b35b9a35cbc36137473e9aa"),
			MustHexID("c9f6f2dd1a941926f03f770695bda289859e85fabaf94baaae20b93e5015dc014ba41150176a36a1884adb52f405194693e63b0c464a6891aa"),
			MustHexID("5b116f0751526868a909b61a30b0c5282c37df6925cc03ddea556ef0d0602a9595fd6c14d371f8ed7d45d89918a032dcd22be4342a8793d8aa"),
			MustHexID("50f3222fb6b82481c7c813b2172e1daea43e2710a443b9c2a57a12bd160dd37e20f87aa968c82ad639af6972185609d47036c0d93b4b7269aa"),
		},
		251: {
			MustHexID("9b8f702a62d1bee67bedfeb102eca7f37fa1713e310f0d6651cc0c33ea7c5477575289ccd463e5a2574a00a676a1fdce05658ba447bb9d28aa"),
			MustHexID("b97532eb83054ed054b4abdf413bb30c00e4205545c93521554dbe77faa3cfaa5bd31ef466a107b0b34a71ec97214c0c83919720142cddacaa"),
			MustHexID("2f7a5e952bfb67f2f90b8441b5fadc9ee13b1dcde3afeeb3dd64bf937f86663cc5c55d1fa83952b5422763c7df1b7f2794b751c6be316ebcaa"),
			MustHexID("42c7483781727051a0b3660f14faf39e0d33de5e643702ae933837d036508ab856ce7eec8ec89c4929a4901256e5233a3d847d5d4893f91baa"),
			MustHexID("873bae27bf1dc854408fba94046a53ab0c965cebe1e4e12290806fc62b88deb1f4a47f9e18f78fc0e7913a0c6e42ac4d0fc3a20cea6bc65faa"),
			MustHexID("a7e3a370bbd761d413f8d209e85886f68bf73d5c3089b2dc6fa42aab1ecb5162635497eed95dee2417f3c9c74a3e76319625c48ead2e963caa"),
			MustHexID("528597534776a40df2addaaea15b6ff832ce36b9748a265768368f657e76d58569d9f30dbb91e91cf0ae7efe8f402f17aa0ae15f5c55051baa"),
			MustHexID("461d8bd4f13c3c09031fdb84f104ed737a52f630261463ce0bdb5704259bab4b737dda688285b8444dbecaecad7f50f835190b38684ced5eaa"),
			MustHexID("6ec50c0be3fd232737090fc0111caaf0bb6b18f72be453428087a11a97fd6b52db0344acbf789a689bd4f5f50f79017ea784f8fd6fe723adaa"),
			MustHexID("12fc5e2f77a83fdcc727b79d8ae7fe6a516881138d3011847ee136b400fed7cfba1f53fd7a9730253c7aa4f39abeacd04f138417ba7fcb0faa"),
			MustHexID("4fdbe75914ccd0bce02101606a1ccf3657ec963e3b3c20239d5fec87673fe446d649b4f15f1fe1a40e6cfbd446dda2d31d40bb602b1093b8aa"),
			MustHexID("3753668a0f6281e425ea69b52cb2d17ab97afbe6eb84cf5d25425bc5e53009388857640668fadd7c110721e6047c9697803bd8a6487b43bbaa"),
			MustHexID("2e81b16346637dec4410fd88e527346145b9c0a849dbf2628049ac7dae016c8f4305649d5659ec77f1e8a0fac0db457b6080547226f06283aa"),
			MustHexID("802c3cc27f91c89213223d758f8d2ecd41135b357b6d698f24d811cdf113033a81c38e0bdff574a5c005b00a8c193dc2531f8c1fa05fa60aaa"),
			MustHexID("fcc9a2e1ac3667026ff16192876d1813bb75abdbf39b929a92863012fe8b1d890badea7a0de36274d5c1eb1e8f975785532c50d80fd44b1aaa"),
			MustHexID("6d8b3efb461151dd4f6de809b62726f5b89e9b38e9ba1391967f61cde844f7528fecf821b74049207cee5a527096b31f3ad623928cd3ce51aa"),
		},
		252: {
			MustHexID("f1ae93157cc48c2075dd5868fbf523e79e06caf4b8198f352f6e526680b78ff4227263de92612f7d63472bd09367bb92a636fff16fe46ccfaa"),
			MustHexID("587f482d111b239c27c0cb89b51dd5d574db8efd8de14a2e6a1400c54d4567e77c65f89c1da52841212080b91604104768350276b6682f2faa"),
			MustHexID("e3f88274d35cefdaabdf205afe0e80e936cc982b8e3e47a84ce664c413b29016a4fb4f3a3ebae0a2f79671f8323661ed462bf4390af94c42aa"),
			MustHexID("0ddc736077da9a12ba410dc5ea63cbcbe7659dd08596485b2bff3435221f82c10d263efd9af938e128464be64a178b7cd22e19f400d5802faa"),
			MustHexID("784aa34d833c6ce63fcc1279630113c3272e82c4ae8c126c5a52a88ac461b6baeed4244e607b05dc14e5b2f41c70a273c3804dea237f14f7aa"),
			MustHexID("f253a2c354ee0e27cfcae786d726753d4ad24be6516b279a936195a487de4a59dbc296accf20463749ff55293263ed8c1b6365eecb248d44aa"),
			MustHexID("a1910b80357b3ad9b4593e0628922939614dc9056a5fbf477279c8b2c1d0b4b31d89a0c09d0d41f795271d14d3360ef08a3f821e65e7e1f5aa"),
			MustHexID("f1168552c2efe541160f0909b0b4a9d6aeedcf595cdf0e9b165c97e3e197471a1ee6320e93389edfba28af6eaf10de98597ad56e7ab1b504aa"),
			MustHexID("b0c8e5d2c8634a7930e1a6fd082e448c6cf9d2d8b7293558b59238815a4df926c286bf297d2049f14e8296a6eb3256af614ec1812c4f2bbeaa"),
			MustHexID("0fb346076396a38badc342df3679b55bd7f40a609ab103411fe45082c01f12ea016729e95914b2b5540e987ff5c9b133e85862648e7f36abaa"),
			MustHexID("f736e0cc83417feaa280d9483f5d4d72d1b036cd0c6d9cbdeb8ac35ceb2604780de46dddaa32a378474e1d5ccdf79b373331c30c7911ade2aa"),
			MustHexID("8b02991457602f42b38b342d3f2259ae4100c354b3843885f7e4e07bd644f64dab94bb7f38a3915f8b7f11d8e3f81c28e07a0078cf79d739aa"),
			MustHexID("9221d9f04a8a184993d12baa91116692bb685f887671302999d69300ad103eb2d2c75a09d8979404c6dd28f12362f58a1a43619c493d9108aa"),
			MustHexID("652797801744dada833fff207d67484742eea6835d695925f3e618d71b68ec3c65bdd85b4302b2cdcb835ad3f94fd00d8da07e570b41bc0daa"),
			MustHexID("d84f06fe64debc4cd0625e36d19b99014b6218375262cc2209202bdbafd7dffcc4e34ce6398e182e02fd8faeed622c3e175545864902dfd3aa"),
			MustHexID("d0ed87b294f38f1d741eb601020eeec30ac16331d05880fe27868f1e454446de367d7457b41c79e202eaf9525b029e4f1d7e17d85a55f83aaa"),
		},
		253: {
			MustHexID("ad4485e386e3cc7c7310366a7c38fb810b8896c0d52e55944bfd320ca294e7912d6c53c0a0cf85e7ce226e92491d60430e86f8f15cda0161aa"),
			MustHexID("36d0e7e5b7734f98c6183eeeb8ac5130a85e910a925311a19c4941b1290f945d4fc3996b12ef4966960b6fa0fb29b1604f83a0f81bd5fd63aa"),
			MustHexID("7d307d8acb4a561afa23bdf0bd945d35c90245e26345ec3a1f9f7df354222a7cdcb81339c9ed6744526c27a1a0c8d10857e98df942fa4336aa"),
			MustHexID("d97bf55f88c83fae36232661af115d66ca600fc4bd6d1fb35ff9bb4dad674c02cf8c8d05f317525b5522250db58bb1ecafb7157392bf5aa6aa"),
			MustHexID("7045d678f1f9eb7a4613764d17bd5698796494d0bf977b16f2dbc272b8a0f7858a60805c022fc3d1fe4f31c37e63cdaca0416c0d053ef48aaa"),
			MustHexID("14e1f21418d445748de2a95cd9a8c3b15b506f86a0acabd8af44bb968ce39885b19c8822af61b3dd58a34d1f265baec30e3ae56149dc7d2aaa"),
			MustHexID("b9453d78281b66a4eac95a1546017111eaaa5f92a65d0de10b1122940e92b319728a24edf4dec6acc412321b1c95266d39c7b3a5d265c629aa"),
			MustHexID("e8a49248419e3824a00d86af422f22f7366e2d4922b304b7169937616a01d9d6fa5abf5cc01061a352dc866f48e1fa2240dbb453d872b1d7aa"),
			MustHexID("bebcff24b52362f30e0589ee573ce2d86f073d58d18e6852a592fa86ceb1a6c9b96d7fb9ec7ed1ed98a51b6743039e780279f6bb49d0a043aa"),
			MustHexID("d0835e5a4291db249b8d2fca9f503049988180c7d247bedaa2cf3a1bad0a76709360a85d4f9a1423b2cbc82bb4d94b47c0cde20afc430224aa"),
			MustHexID("6b087fe2a2da5e4f0b0f4777598a4a7fb66bf77dbd5bfc44e8a7eaa432ab585a6e226891f56a7d4f5ed11a7c57b90f1661bba1059590ca42aa"),
			MustHexID("d901e5bde52d1a0f4ddf010a686a53974cdae4ebe5c6551b3c37d6b6d635d38d5b0e5f80bc0186a2c7809dbf3a42870dd09643e68d32db89aa"),
			MustHexID("96419fb80efae4b674402bb969ebaab86c1274f29a83a311e24516d36cdf148fe21754d46c97688cdd7468f24c08b13e4727c29263393638aa"),
			MustHexID("7b9c1889ae916a5d5abcdfb0aaedcc9c6f9eb1c1a4f68d0c2d034fe79ac610ce917c3abc670744150fa891bfcd8ab14fed6983fca964de92aa"),
			MustHexID("7a369b2b8962cc4c65900be046482fbf7c14f98a135bbbae25152c82ad168fb2097b3d1429197cf46d3ce9fdeb64808f908a489cc6019725aa"),
			MustHexID("47bcae48288da5ecc7f5058dfa07cf14d89d06d6e449cb946e237aa6652ea050d9f5a24a65efdc0013ccf232bf88670979eddef249b054f6aa"),
		},
		254: {
			MustHexID("099739d7abc8abd38ecc7a816c521a1168a4dbd359fa7212a5123ab583ffa1cf485a5fed219575d6475dbcdd541638b2d3631a6c7fce7474aa"),
			MustHexID("c2b01603b088a7182d0cf7ef29fb2b04c70acb320fccf78526bf9472e10c74ee70b3fcfa6f4b11d167bd7d3bc4d936b660f2c9bff934793daa"),
			MustHexID("20e4d8f45f2f863e94b45548c1ef22a11f7d36f263e4f8623761e05a64c4572379b000a52211751e2561b0f14f4fc92dd4130410c8ccc71eaa"),
			MustHexID("27f4a16cc085e72d86e25c98bd2eca173eaaee7565c78ec5a52e9e12b2211f35de81b5b45e9195de2ebfe29106742c59112b951a04eb7ae4aa"),
			MustHexID("55db5ee7d98e7f0b1c3b9d5be6f2bc619a1b86c3cdd513160ad4dcf267037a5fffad527ac15d50aeb32c59c13d1d4c1e567ebbf4de0d2523aa"),
			MustHexID("883df308b0130fc928a8559fe50667a0fff80493bc09685d18213b2db241a3ad11310ed86b0ef662b3ce21fc3d9aa7f3fc24b8d9afe17c74aa"),
			MustHexID("c7af968cc9bc8200c3ee1a387405f7563be1dce6710a3439f42ea40657d0eae9d2b3c16c42d779605351fcdece4da637b9804e60ca08cfb8aa"),
			MustHexID("3e66f2b788e3ff1d04106b80597915cd7afa06c405a7ae026556b6e583dca8e05cfbab5039bb9a1b5d06083ffe8de5780b1775550e7218f5aa"),
			MustHexID("4fc7f53764de3337fdaec0a711d35d3a923e72fa65025444d12230b3552ed43d9b2d1ad08ccb11f2d50c58809e6dd74dde910e195294fca3aa"),
			MustHexID("bafdfdcf6ccaa989436752fa97c77477b6baa7deb374b16c095492c529eb133e8e2f99e1977012b64767b9d34b2cf6d2048ed489bd822b51aa"),
			MustHexID("7f5d78008a4312fe059104ce80202c82b8915c2eb4411c6b812b16f7642e57c00f2c9425121f5cbac4257fe0b3e81ef5dea97ea2dbaa98f6aa"),
			MustHexID("598c37fe78f922751a052f463aeb0cb0bc7f52b7c2a4cf2da72ec0931c7c32175d4165d0f8998f7320e87324ac3311c03f9382a5385c55f0aa"),
			MustHexID("f758c4136e1c148777a7f3275a76e2db0b2b04066fd738554ec398c1c6cc9fb47e14a3b4c87bd47deaeab3ffd2110514c3855685a374794daa"),
			MustHexID("0307bb9e4fd865a49dcf1fe4333d1b944547db650ab580af0b33e53c4fef6c789531110fac801bbcbce21fc4d6f61b6d5b24abdf5b22e303aa"),
			MustHexID("82504b6eb49bb2c0f91a7006ce9cefdbaf6df38706198502c2e06601091fc9dc91e4f15db3410d45c6af355bc270b0f268d3dff560f95698aa"),
			MustHexID("b39b5b677b45944ceebe76e76d1f051de2f2a0ec7b0d650da52135743e66a9a5dba45f638258f9a7545d9a790c7fe6d3fdf82c25425c7887aa"),
		},
		255: {
			MustHexID("5c4d58d46e055dd1f093f81ee60a675e1f02f54da6206720adee4dccef9b67a31efc5c2a2949c31a04ee31beadc79aba10da31440a1f9ff2aa"),
			MustHexID("ea72161ffdd4b1e124c7b93b0684805f4c4b58d617ed498b37a145c670dbc2e04976f8785583d9c805ffbf343c31d492d79f841652bbbd01aa"),
			MustHexID("51caa1d93352d47a8e531692a3612adac1e8ac68d0a200d086c1c57ae1e1a91aa285ab242e8c52ef9d7afe374c9485b122ae815f1707b875aa"),
			MustHexID("c08397d5751b47bd3da044b908be0fb0e510d3149574dff7aeab33749b023bb171b5769990fe17469dbebc100bc150e798aeda426a2dcc76aa"),
			MustHexID("0222c1c194b749736e593f937fad67ee348ac57287a15c7e42877aa38a9b87732a408bca370f812efd0eedbff13e6d5b854bf3ba1dec431aaa"),
			MustHexID("03d859cd46ef02d9bfad5268461a6955426845eef4126de6be0fa4e8d7e0727ba2385b78f1a883a8239e95ebb814f2af8379632c7d5b1006aa"),
			MustHexID("64d5004b7e043c39ff0bd10cb20094c287721d5251715884c280a612b494b3e9e1c64ba6f67614994c7d969a0d0c0295d107d53fc225d47caa"),
			MustHexID("b0a5eefb2dab6f786670f35bf9641eefe6dd87fd3f1362bcab4aaa792903500ab23d88fae68411372e0813b057535a601d46e454323745a9aa"),
			MustHexID("0cc6df0a3433d448b5684d2a3ffa9d1a825388177a18f44ad0008c7bd7702f1ec0fc38b83506f7de689c3b6ecb552599927e29699eed6bb8aa"),
			MustHexID("50772f7b8c03a4e153355fbbf79c8a80cf32af656ff0c7873c99911099d04a0dae0674706c357e0145ad017a0ade65e6052cb1b0d574fcd6aa"),
			MustHexID("1ae37829c9ef41f8b508b82259ebac76b1ed900d7a45c08b7970f25d2d48ddd1829e2f11423a18749940b6dab8598c6e416cef0efd47e46eaa"),
			MustHexID("ba973cab31c2af091fc1644a93527d62b2394999e2b6ccbf158dd5ab9796a43d408786f1803ef4e29debfeb62fce2b6caa5ab2b24d1549c8aa"),
			MustHexID("bc413ad270dd6ea25bddba78f3298b03b8ba6f8608ac03d06007d4116fa78ef5a0cfe8c80155089382fc7a193243ee5500082660cb5d7793aa"),
			MustHexID("5a6a9ef07634d9eec3baa87c997b529b92652afa11473dfee41ef7037d5c06e0ddb9fe842364462d79dd31cff8a59a1b8d5bc2b810dea1d4aa"),
			MustHexID("f492c6ee2696d5f682f7f537757e52744c2ae560f1090a07024609e903d334e9e174fc01609c5a229ddbcac36c9d21adaf6457dab38a25bfaa"),
			MustHexID("459e4db99298cb0467a90acee6888b08bb857450deac11015cced5104853be5adce5b69c740968bc7f931495d671a70cad9f48546d7cd203aa"),
		},
		256: {
			MustHexID("a8593af8a4aef7b806b5197612017951bac8845a1917ca9a6a15dd6086d608505144990b245785c4cd2d67a295701c7aac2aa18823fb0033aa"),
			MustHexID("d2eebef914928c3aad77fc1b2a495f52d2294acf5edaa7d8a530b540f094b861a68fe8348a46a7c302f08ab609d85912a4968eacfea07408aa"),
			MustHexID("b14bfcb31495f32b650b63cf7d08492e3e29071fdc73cf2da0da48d4b191a70ba1a65f42ad8c343206101f00f8a48e8db4b08bf3f622c085aa"),
			MustHexID("7feaee0d818c03eb30e4e0bf03ade0f3c21ca38e938a761aa1781cf70bda8cc5cd631a6cc53dd44f1d4a6d3e2dae6513c6c66ee50cb2f0e9aa"),
			MustHexID("4ca3b657b139311db8d583c25dd5963005e46689e1317620496cc64129c7f3e52870820e0ec7941d28809311df6db8a2867bbd4f235b4248aa"),
			MustHexID("1181defb1d16851d42dd951d84424d6bd1479137f587fa184d5a8152be6b6b16ed08bcdb2c2ed8539bcde98c80c432875f9f724737c316a2aa"),
			MustHexID("d9dd818769fa0c3ec9f553c759b92476f082817252a04a47dc1777740b1731d280058c66f982812f173a294acf4944a85ba08346e2de153baa"),
			MustHexID("bd7c4f8a9e770aa915c771b15e107ca123d838762da0d3ffc53aa6b53e9cd076cffc534ec4d2e4c334c683f1f5ea72e0e123f6c261915ed5aa"),
			MustHexID("3dd5739c73649d510456a70e9d6b46a855864a4a3f744e088fd8c8da11b18e4c9b5f2d7da50b1c147b2bae5ca9609ae01f7a3cdea9dce34faa"),
			MustHexID("f0d7df1efc439b4bcc0b762118c1cfa99b2a6143a9f4b10e3c9465125f4c9fca4ab88a2504169bbcad65492cf2f50da9dd5d077c39574a94aa"),
			MustHexID("dd598b9ba441448e5fb1a6ec6c5f5aa9605bad6e223297c729b1705d11d05f6bfd3d41988b694681ae69bb03b9a08bff4beab5596503d12aaa"),
			MustHexID("3fce284ac97e567aebae681b15b7a2b6df9d873945536335883e4bbc26460c064370537f323fd1ada828ea43154992d14ac0cec0940a2bd2aa"),
			MustHexID("7c8dfa8c1311cb14fb29a8ac11bca23ecc115e56d9fcf7b7ac1db9066aa4eb39f8b1dabf46e192a65be95ebfb4e839b5ab4533fef4149218aa"),
			MustHexID("cafa6934f82120456620573d7f801390ed5e16ed619613a37e409e44ab355ef755e83565a913b48a9466db786f8d4fbd590bfec474c2524daa"),
			MustHexID("9d16600d0dd310d77045769fed2cb427f32db88cd57d86e49390c2ba8a9698cfa856f775be2013237226e7bf47b248871cf865d23015937daa"),
			MustHexID("17be6b6ba54199b1d80eff866d348ea11d8a4b341d63ad9a6681d3ef8a43853ac564d153eb2a8737f0afc9ab320f6f95c55aa11aaa13bbb1aa"),
		},
	},
}

type preminedDevin struct {
	target    NodeID
	targetSha common.Hash // sha3(target)
	dists     [hashBits + 1][]NodeID
	net       *Network
}

func (tn *preminedDevin) sendFindnodeHash(to *Node, target common.Hash) {
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

func (tn *preminedDevin) sendPing(to *Node, addr *net.UDPAddr, topics []Topic) []byte {
	injectResponse(tn.net, to, pongPacket, &pong{ReplyTok: []byte{1}})
	return []byte{1}
}

func (tn *preminedDevin) send(to *Node, ptype nodeEvent, data interface{}) (hash []byte) {
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

func (tn *preminedDevin) sendNeighbours(to *Node, nodes []*Node) {
	panic("sendNeighbours called")
}

func (tn *preminedDevin) sendTopicNodes(to *Node, queryHash common.Hash, nodes []*Node) {
	panic("sendTopicNodes called")
}

func (tn *preminedDevin) sendTopicRegister(to *Node, topics []Topic, idx int, pong []byte) {
	panic("sendTopicRegister called")
}

func (*preminedDevin) Close() {}

func (*preminedDevin) localAddr() *net.UDPAddr {
	return &net.UDPAddr{IP: net.ParseIP("10.0.1.1"), Port: 40000}
}

func injectResponse(net *Network, from *Node, ev nodeEvent, packet interface{}) {
	go net.reqReadPacket(ingressPacket{remoteID: from.ID, remoteAddr: from.addr(), ev: ev, data: packet})
}
