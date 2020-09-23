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

package whisperv6

import (
	"bytes"
	"fmt"
	"github.com/core-coin/eddsa"
	mrand "math/rand"
	"sync"
	"testing"
	"time"

	"net"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/hexutil"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/p2p"
	"github.com/core-coin/go-core/p2p/enode"
	"github.com/core-coin/go-core/p2p/nat"
	"github.com/core-coin/go-core/rlp"
)

var keys = []string{
	"f973213b8dc2debd115d06c8b9478780d8881d6b7abd77787abea2c542a22a12043e0973b4c9a29ee87fccbec3b01e440da96c356e7d80202b8099ed102d987ba4e1ce0d7751309c31167f150054e3c718da2103c0c05679495befaabf6d94bbc491d0e1421efc1ed3ee433183efb46c4066c200198894deac70ec8bfb6c9589883a2cafaa7e19f2978172d3a6676dc0",
	"405cf7102e157be61e84ed4125479f982e4b29eaaaf434c8fd26bdbbb65281c8dc6f97aaabdc6c71d691f9b0ab493a96f1430905ce96d315dde65964ce54c4aee1c4125c00b1a02da246366edac8f6b039ac83897376af2c9c0a87795cedfe174be8c707dbffbaff2259bf6089bb37a19ffce3c0c294933840e5fe1d63165ce140fc3cc092484c2a29ef963bd4ea5caf",
	"e771190f75acf998879b399b18449cf0d56bb3041ec7681e40bd3a1c45d9aab1e73c3d3a123d64d08c207b24f267a8087e55831259e62716a15808fccda9444ff44a60ac56e295b3b53e7e5846ac323c52c4bb1a2026c06fe9344602b01d182196aa254ca927698e9e3c96b46fc79bb2129c3245c5feb161a0f282d4efc83cf647eb06cc8c39785867493b336dd0cdbc",
	"d6cff2205c0967a1ebfdd1fa0ebe4e732276b6d5be2eab8f3b68578b65a61ab60deb31a401814fa5e9cf1dde4e9ec71f7182de0be69529354c1d5fa3ca5df7b7d93d67fa1fcae55dd9b7f178f9228b9be24caa8a7a1f043c698a9bf3603b2917a200deb6ad6fc3da6d8d17c6d8868fcafa465336358d598dfa04e46294a45ba7f8fc7048ff79600d446d5f0c1c8cf8b1",
	"d3ad61749ba2d27b3df90eac1e6a0d92b9d92b320512dc603bd696a5413d3f1c53aa5ad9ab7ef255ba435a16001fb5d45793a89650d4782508fa0c6ecabf0773d1288f0bd98da492eece0263f294d783161eca2cecaa492434946f82fcfd67ac0af7388ba1c0167ce5ba66100009a345240108e38f6c7218eac034648d9bae6ed79fa5ea5df26c21f42e88493a4cce49",
	"f7edd052dc053feec54b96b94f690d7495b7f6e90bb1714340a833d9adf0ab776b1b55576d050e08ae71e4d1f0467e3599e64d49a7af8e190a4fd982f6ee9541a741d3cf0c507e53a69d50e5d8c4a6bf8088675505a566166538da2bbd64109725fe5cf5fa5fdaa3f4ecf5e94d290212e4319f5c22dba135e381a8cbf17d00297d8f85a1f710613519d8b4d260ba61dc",
	"4ca72453d4b218ee7cad26ffe11e3b87200a9dbaa0558fce755d5b0dfa0417fad526df1bbfb9ef0e8721cf49e4e0ec26aedd513e705bc22cc5bec9fa5a498b9998a0ed65adb0a5769bc9329c71e5db073507137253973075f85f5c1664283fe1de81962af8e8d7824f746d3357a78ff5703e4283fb0ff8ca4e56e1c83462e68af4cc545482cc88c11cfbb7ba680b411a",
	"3ebd6f12b41d4faecc091d98c0128e99fec0425ce2471306ba623bb3615018bce195bba4fb0a2497332874bd2040aa112d7a48ef9a604f36d0f39d182cbc3dcca2021c6e97651002f1f0e789a371fa5f52f1369aa369d2c91e2739f2450d4d51dd0396f8856a7a8fd99c42c0cb5f56aa99c2a47b9736f27c99c585f596e8a8b465ec1385baf24df92bbb46306465b31e",
	"ee3f6057feb9c0f384ccafbf7aa07157d11c543d4e0b50fcde41d27365c142d05c85d72d9d79bcbe5b94b7c3b2d38986066774caacb90a37a74e9cf8cf91356292373bcf0a7b465a255ac4608d95e2f9688dc73999fc2aa5d490e2b4803a63fde003f5a9a1cb2c847c4ac7ac8813964a2504e148e21590b96f207cb882fa680ff239005f4a0e76fb0708ec2f12487297",
	"61eda6235dcf77eb7b206b97378cb70367524abad3ada6b801eb8b2413df22cdea0b332594639be27ce7fe3e71c165c25c994391f4661e15f68127492f77ae1babdd6cdefef4e190e9cb7a72e5d4eff4193d120f8ee629468251bf1eaf0c023c0642e3f747e88d8c1719d78d9aab8a9105db1197c55ca06887971596e1d2a0d91a4f3344b6a5480940345fa31a87d041",
	"6e8e4928db08dfd4f4e804616636bf30a3f5a857b3a89a2f60489304185511b7034ec66e41df1fb94d11f59f4a33667e72510cacae78410fd8b8d84e3591da513c537bc3762d4abb0529d552828fc1efb4cd5f081c4d299dc12b02a084efd793536c4d30ed61533ab4e34a88fd350e6970396b893b2e946be5b9fbb0522962d6638ff206e1cb1b44593574b3f76f0935",
	"61917ceeb798fd7a5aad99878232723fe464ee0e6329f805ead8e34c9828f3b30a898185653c8cd18484dd8e2974a601f15e59d6e75f1915f47b3494f169eac040cb476dd94032a8685d6da3b9727ba49bbef3ea3276cedcab2ae0c3b204edce9ab06d3d8c2f235d59d012755cd23fdd56825e583608687e5eae64be98f1b56e9b1ddcf9f0c71ad817268fba85b4b270",
	"27aaea6c15a2207906fc4e1e4a60f5dcfe99a47bab98c9accb08885ac3c110694a8141c2bbc8b8eaf4bee0f47518795383009564fa1161085ffece28d2c7b70a277753f4b17df59528fe1a7648bbdaba21b2db108fc010299154c95a7530c5b3ca32fcf46ceba56850a6baaeb464e8c02271c40047a38c1d221d7cc117f3cf2b85897481a8a98e844b41909817e9796f",
	"0b4f9143d093dd2ed7179d33cdfb41844e5e27bc8f469b7713f6c87963b23d881c3dcbe7adcbb53d6ed382aed916cf688f2cd56c6e26e52564f9ae94831c77847f488664dde03cec48710a02b820afd97da919e959515324e6063c061fdc918de9db9756fb1c908439ddcfd74c682dc29903b507ebb7d21e9625ae9f3a10154d719216d9398ae47d3525ed835b5e739d",
	"07b45ecd00056509fa8d340a882e80aae90da8dee2649933a7b55d6bf5ba8fba67932679a37bc56ce079f28d3be0504b4b67888052a70e2ea6f692eb11f507aeeab5deb970cfa0649dabe150e8de0c68e38c3fcf47534fde23a0faa05c7ff88707f48a6fecf59795affb7d6fd6502d0e42f65b9bf525d2491fb519db502a7809c7128cdbcbeb73eda12883b0a67e5f12",
	"eea4d9ee2d38b926e9d5df37b9cec37224c672e4440b41817b72c2d92aa8210f42050869717a8331177891cc9d515b6a7042312eff74890f4ab021a8c870828947dca76438d80ebf5e912bf6f49fe58d95621f1471647a277dd05fef742587135b9e7b0a2459a00ced8650e07b95bf3dcb6182e5caf6ee683e8ff4d4913ada31c3c04004c710233e4a272eafe00f402d",
	"0f10b73b0ed5727a12b95b5d87ac0860cabe6ffa7017345b02b49fbf5799c9aef9ab7b5753238500acc2ab932de91b12fe842f79157190397c93aa0477462564443150bd9461184727c6b82e08608aefc181608f96218919ab59c3465e1ce54b0bf298bc218a9d8f4b32c0755853f7b7bd8f00e0a587c6b67668dd40f9fcb5fd4170c4d37adeae2d5baa0b185e36f2b3",
	"cb6bf2e737b53faf069f025a77400d63ce20daa88d0f290e7b8c39ba6aa8d76b22f505f249cb3503ac60fae358e7d6606bbaacc1368b6a1a6dfc2f5347bdfe3641b972d4258be66218abab624029b50851fb56f4f14f57243e2584cda19e508b4303cd14d7163abd6ac29794d4214795844514f48de7d5f9c07f4d18074402cfadcd5cdc00439359605a90b83beef2c0",
	"ddbe05178eb13896ee38f0d7148101d490552ae89459a02530b03e5a5c6f7121e2d9c98287e1d04e338336bd1553c045d37e73092f380a18d4c5997945b264bdf681e6fd769694e926cdd917ac4ea3818cd3fde780d8536bee5b149152f30bf9787a3fb1834abe85a05c469d9f17f7b7244c6b365a12ef612f08be03943facf5ba12435b9e58f7edcf70430717b6a124",
	"2e7717af726631bbacb3e7a90c9f25fc6ef14eeae8facd054876073826774bf56296959fbf58f6507399af435949b0d53353696e5f161f0ef4da80eb6903f4fb5dddd653ef92c19cc3944842fd0fff0720952f46c48761b869e1e9380447ebcfa10851eecfc5c53f44c36f96a180b70cc6e6d2211d195a7322cd0148a52a29398046fa003407bd181d913f4eedd4f643",
	"ce5f7add74034a841a6586b5e0a2290c81a210bc4e7e3064248117746b4072e66f24bce47f383528626c832ef107492f47cc22c6b0fdc10138965678289cfda01d1150d210cdd71ec060ddd9cc1b7da54d0076d64c762277a885dbeea9363fa4758cb44af21375115c22078cde2b4def2c003d95593b13e78e2760e5184e2050ffae5cfcc28298a2774983378676cd5d",
	"def11377abb23dedbf3675c6c8ec3c7233436fe55510b0b285ec681e88e7550e9cb4bdc7c0c966070fc7f69ace519c72efc9d5ac2ffa9b0d149d7773b52224d5decbe3a76234c36fd49803d320339aabcd7aefa53440385eadfed9cd224ede48d55486be16e78001f3dfa0de2e02fdb02efd806179f97f3d80dc018c4d5fd1760daddadaa1ea9e8f31da06f2089238de",
	"f3e957aa084a030d7e6101e2331b4abe7f0d2fb432e56db8e1e7c9402e11dc163aa1cc89b654b0e9be53e2dc75cbff16377a1e7b9e4d9e093792a73efb56feeaf9414da5ee3baffc68e6ca3770c3e83b2d1c20cbedc4bb82ca7ce448e6d4b0b997812bca4f341f7bf286d13fd02d9c545d3929de7bd02cc67ad053fb0ba0880681a65b240fa22b7699ba6d2cd277b53a",
	"ba6d89ae28bd45e47175a98d27b334e7fc59c62fd8b21a38be1dff681e09f9c4b50e7b216ec5774fcc30b345b3ded1d9db75085effe75f31e764b3dd0c9ad1923a0894c88035509f9789d5958f8bd35a178e5b874a5452de1fa5fba16bd335193e27d139d84b3dd3f1a9498e618b38ea315f8d466fe089334b0ec0bd0e641cbac0aa3580e3e90776d61b9919e07426f7",
	"3d53796e7570d829556b334f772fb21389eea38d93d1651e5a0b24b0b23e892c41f5e82f93a4d1f7ba1ff49ec18ff2a1d3d304fb53fc5d1aa3251f1663fc7c13cbda96b69cacf370317a9407bd3560ccc49536e4dd2ff3009815de1b4caac568a19234cd1a12a7717a1e47a3d1caa97288d9b0ec66522ce8ee97ed0579646c08fbaad4c60e99be819ed9b24b4a9f965c",
	"635e74f0ed57a4d466d08f1815c7c98157be0a99f8c8522dc75b19709297f95ee95c40051ed4be0a2abba4ddd249f35dab0a4f73788fdc2984d6139941b189e115ba3866c1e97d8402e78613fedd7163fcf7e1a9a25a092cd30a040f6e0c7adeae59d9f092df2bdca7f1069b4a2300bab27a57a6f798f385650ec0d12e7fcbdb1b851625a8e010ad49d0962f3fda06bd",
	"8fa5ab9916d884c99f928dbdb554d7add48d5a289caa7c3b87d23279619ef2c0a331442aff0b9cd3b7fdb623f6274c3a1292f049fa266b2e45778441a17dd00b3f5504d70e074197e908af86ebf42abd181ea8292aae9baf2984e6325c043f6dde9766beb7695eabdad146cd4578ccce68b64f85284233f25eefb74e7d69938b42775e859559ecf40921856ea688fc26",
	"5ffb670e50db6d4b60bc743c097ea3eb812615f1ff24cc726bc575a0db245f9a83de7f1e97f51541fe0ed01e0c3f1c01cf5f9fb00a3963364b257ae8f869394d4381be007413bf53ada90a57fbb919be6e7400a495bf879beac7eb0b2be919e61485c4842af1e6b0baa7b66aaa4acbc6547a56a47a88a17cbcba08f7ba60ad11ebd1ece0c7d95691c6f60f4d8e987cdd",
	"b0c354a5b155cb640168e8c0995e1cdb92fc9eae02d3c4f4c85da44ca73cddcddd46c4512ab5c1964d99ac87d71721d107adec9e72a82c30deaddcecc09058d2c1881c7aa733f421a985e03d69a51128410a65a1a47992dfc2fa66d3b352a07044393fb2a45ba8eb9b0d3a59877303ed488d7ef0342cc9c5f6332ef8a4cc987014ef9227da868a5b40d71d36e9bd38bb",
	"fdc3e27de9920693b333e3cc911d8ad25d8f1395689e0f9c8730ab3dfecbd66a2957f30ab257d28dab1d1b3d7997ff1d2c42a695fddeb02692e97cc320b7e57b160a305ad0ef02ebba7541a27263020fcba447da1bd4ee3f676adc81fb55ea197b9a24729eb68cdab7cf2cfac083926ee8d25028a01effd7430a1d58dbc6bb0d8d5ec82c9c8ac588fafb418ba89fa5b9",
	"4c9a00e6699d32910c0e9cf6f0afea214f749526fca86fdbdef0964d53a5adbea15edd803cea6ee67a178df35f0f66fa76b55b5237c8e2375dfdc30173b7265f7c7aa7effd343aa86ede3d7addbade76a7c1012ab52bf73a4f0449fb74cb6d35cbb9a90717995a8737190fcd0ed004aa4f114ea0f43b9cccb80cc9d6698fd1476e5d7dc4292360ff9924befa51d7ab4d",
	"3fb9d032b1a65199c738535ffa0156f088d5711f6643f347afcbea4a6022d34a688a2e85d50c867bd74cda0b0c4963c52d072db6474b6f316571de66b6665515d588be6d18bd66c711726f3f11dc2f24d4e30a9a46d28a967de0e326055c365d832e453c6b380472e8050bcae8b31983265d731354fd9344a5e0fdeda8e05e3cb1864873c845145644c5b7e67e2e7537",
	"084e44c699d643f0314461e3f4189d0951fbea92954cb021863f85f95caed5955522a5e132b654b01d3bfdbd0d06438028f88e61687a6214b2e4efb53082ee766bbc35580781898309da11aa1890d9468e45ee20fadcc63820ec7b8aa58c6017d5339730ca675cbe2233ce9d01dcdf33000606302b56f84baf07e972c10bc05a316fe5280f237dd720d630628f7e9dce",
}

type TestData struct {
	counter [NumNodes]int
	mutex   sync.RWMutex
}

type TestNode struct {
	shh     *Whisper
	id      *eddsa.PrivateKey
	server  *p2p.Server
	filerID string
}

const NumNodes = 8 // must not exceed the number of keys (32)

var result TestData
var nodes [NumNodes]*TestNode
var sharedKey = hexutil.MustDecode("0x03ca634cae0d49acb401d8a4c6b6fe8c55b70d115bf400769cc1400f3258cd31")
var wrongKey = hexutil.MustDecode("0xf91156714d7ec88d3edc1c652c2181dbb3044e8771c683f3b30d33c12b986b11")
var sharedTopic = TopicType{0xF, 0x1, 0x2, 0}
var wrongTopic = TopicType{0, 0, 0, 0}
var expectedMessage = []byte("per aspera ad astra")
var unexpectedMessage = []byte("per rectum ad astra")
var masterBloomFilter []byte
var masterPow = 0.00000001
var round = 1
var debugMode = false
var prevTime time.Time
var cntPrev int

func TestSimulation(t *testing.T) {
	// create a chain of whisper nodes,
	// installs the filters with shared (predefined) parameters
	initialize(t)

	// each node sends one random (not decryptable) message
	for i := 0; i < NumNodes; i++ {
		sendMsg(t, false, i)
	}

	// node #0 sends one expected (decryptable) message
	sendMsg(t, true, 0)

	// check if each node have received and decrypted exactly one message
	checkPropagation(t, true)

	// check if Status message was correctly decoded
	checkBloomFilterExchange(t)
	checkPowExchange(t)

	// send new pow and bloom exchange messages
	resetParams(t)

	// node #1 sends one expected (decryptable) message
	sendMsg(t, true, 1)

	// check if each node (except node #0) have received and decrypted exactly one message
	checkPropagation(t, false)

	// check if corresponding protocol-level messages were correctly decoded
	checkPowExchangeForNodeZero(t)
	checkBloomFilterExchange(t)

	stopServers()
}

func resetParams(t *testing.T) {
	// change pow only for node zero
	masterPow = 7777777.0
	nodes[0].shh.SetMinimumPoW(masterPow)

	// change bloom for all nodes
	masterBloomFilter = TopicToBloom(sharedTopic)
	for i := 0; i < NumNodes; i++ {
		nodes[i].shh.SetBloomFilter(masterBloomFilter)
	}

	round++
}

func initBloom(t *testing.T) {
	masterBloomFilter = make([]byte, BloomFilterSize)
	_, err := mrand.Read(masterBloomFilter)
	if err != nil {
		t.Fatalf("rand failed: %s.", err)
	}

	msgBloom := TopicToBloom(sharedTopic)
	masterBloomFilter = addBloom(masterBloomFilter, msgBloom)
	for i := 0; i < 32; i++ {
		masterBloomFilter[i] = 0xFF
	}

	if !BloomFilterMatch(masterBloomFilter, msgBloom) {
		t.Fatalf("bloom mismatch on initBloom.")
	}
}

func initialize(t *testing.T) {
	initBloom(t)

	var err error

	for i := 0; i < NumNodes; i++ {
		var node TestNode
		b := make([]byte, BloomFilterSize)
		copy(b, masterBloomFilter)
		node.shh = New(&DefaultConfig)
		node.shh.SetMinimumPoW(masterPow)
		node.shh.SetBloomFilter(b)
		if !bytes.Equal(node.shh.BloomFilter(), masterBloomFilter) {
			t.Fatalf("bloom mismatch on init.")
		}
		node.shh.Start(nil)
		topics := make([]TopicType, 0)
		topics = append(topics, sharedTopic)
		f := Filter{KeySym: sharedKey}
		f.Topics = [][]byte{topics[0][:]}
		node.filerID, err = node.shh.Subscribe(&f)
		if err != nil {
			t.Fatalf("failed to install the filter: %s.", err)
		}
		node.id, err = crypto.HexToEDDSA(keys[i])
		if err != nil {
			t.Fatalf("failed convert the key: %s.", keys[i])
		}
		name := common.MakeName("whisper-go", "2.0")

		node.server = &p2p.Server{
			Config: p2p.Config{
				PrivateKey: node.id,
				MaxPeers:   NumNodes/2 + 1,
				Name:       name,
				Protocols:  node.shh.Protocols(),
				ListenAddr: "127.0.0.1:0",
				NAT:        nat.Any(),
			},
		}

		startServer(t, node.server)
		nodes[i] = &node
	}

	for i := 0; i < NumNodes; i++ {
		for j := 0; j < i; j++ {
			peerNodeId := nodes[j].id
			address, _ := net.ResolveTCPAddr("tcp", nodes[j].server.ListenAddr)
			peer := enode.NewV4(&peerNodeId.PublicKey, address.IP, address.Port, address.Port)
			nodes[i].server.AddPeer(peer)
		}
	}
}

func startServer(t *testing.T, s *p2p.Server) {
	err := s.Start()
	if err != nil {
		t.Fatalf("failed to start the first server. err: %v", err)
	}
}

func stopServers() {
	for i := 0; i < NumNodes; i++ {
		n := nodes[i]
		if n != nil {
			n.shh.Unsubscribe(n.filerID)
			n.shh.Stop()
			n.server.Stop()
		}
	}
}

func checkPropagation(t *testing.T, includingNodeZero bool) {
	if t.Failed() {
		return
	}

	prevTime = time.Now()
	// (cycle * iterations) should not exceed 50 seconds, since TTL=50
	const cycle = 200 // time in milliseconds
	const iterations = 250

	first := 0
	if !includingNodeZero {
		first = 1
	}

	for j := 0; j < iterations; j++ {
		for i := first; i < NumNodes; i++ {
			f := nodes[i].shh.GetFilter(nodes[i].filerID)
			if f == nil {
				t.Fatalf("failed to get filterId %s from node %d, round %d.", nodes[i].filerID, i, round)
			}

			mail := f.Retrieve()
			validateMail(t, i, mail)

			if isTestComplete() {
				checkTestStatus()
				return
			}
		}

		checkTestStatus()
		time.Sleep(cycle * time.Millisecond)
	}

	if !includingNodeZero {
		f := nodes[0].shh.GetFilter(nodes[0].filerID)
		if f != nil {
			t.Fatalf("node zero received a message with low PoW.")
		}
	}

	t.Fatalf("Test was not complete (%d round): timeout %d seconds. nodes=%v", round, iterations*cycle/1000, nodes)
}

func validateMail(t *testing.T, index int, mail []*ReceivedMessage) {
	var cnt int
	for _, m := range mail {
		if bytes.Equal(m.Payload, expectedMessage) {
			cnt++
		}
	}

	if cnt == 0 {
		// no messages received yet: nothing is wrong
		return
	}
	if cnt > 1 {
		t.Fatalf("node %d received %d.", index, cnt)
	}

	if cnt == 1 {
		result.mutex.Lock()
		defer result.mutex.Unlock()
		result.counter[index] += cnt
		if result.counter[index] > 1 {
			t.Fatalf("node %d accumulated %d.", index, result.counter[index])
		}
	}
}

func checkTestStatus() {
	var cnt int
	var arr [NumNodes]int

	for i := 0; i < NumNodes; i++ {
		arr[i] = nodes[i].server.PeerCount()
		envelopes := nodes[i].shh.Envelopes()
		if len(envelopes) >= NumNodes {
			cnt++
		}
	}

	if debugMode {
		if cntPrev != cnt {
			fmt.Printf(" %v \t number of nodes that have received all msgs: %d, number of peers per node: %v \n",
				time.Since(prevTime), cnt, arr)
			prevTime = time.Now()
			cntPrev = cnt
		}
	}
}

func isTestComplete() bool {
	result.mutex.RLock()
	defer result.mutex.RUnlock()

	for i := 0; i < NumNodes; i++ {
		if result.counter[i] < 1 {
			return false
		}
	}

	for i := 0; i < NumNodes; i++ {
		envelopes := nodes[i].shh.Envelopes()
		if len(envelopes) < NumNodes+1 {
			return false
		}
	}

	return true
}

func sendMsg(t *testing.T, expected bool, id int) {
	if t.Failed() {
		return
	}

	opt := MessageParams{KeySym: sharedKey, Topic: sharedTopic, Payload: expectedMessage, PoW: 0.00000001, WorkTime: 1}
	if !expected {
		opt.KeySym = wrongKey
		opt.Topic = wrongTopic
		opt.Payload = unexpectedMessage
		opt.Payload[0] = byte(id)
	}

	msg, err := NewSentMessage(&opt)
	if err != nil {
		t.Fatalf("failed to create new message with seed %d: %s.", seed, err)
	}
	envelope, err := msg.Wrap(&opt)
	if err != nil {
		t.Fatalf("failed to seal message: %s", err)
	}

	err = nodes[id].shh.Send(envelope)
	if err != nil {
		t.Fatalf("failed to send message: %s", err)
	}
}

func TestPeerBasic(t *testing.T) {
	InitSingleTest()

	params, err := generateMessageParams()
	if err != nil {
		t.Fatalf("failed generateMessageParams with seed %d.", seed)
	}

	params.PoW = 0.001
	msg, err := NewSentMessage(params)
	if err != nil {
		t.Fatalf("failed to create new message with seed %d: %s.", seed, err)
	}
	env, err := msg.Wrap(params)
	if err != nil {
		t.Fatalf("failed Wrap with seed %d.", seed)
	}

	p := newPeer(nil, nil, nil)
	p.mark(env)
	if !p.marked(env) {
		t.Fatalf("failed mark with seed %d.", seed)
	}
}

func checkPowExchangeForNodeZero(t *testing.T) {
	const iterations = 200
	for j := 0; j < iterations; j++ {
		lastCycle := (j == iterations-1)
		ok := checkPowExchangeForNodeZeroOnce(t, lastCycle)
		if ok {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func checkPowExchangeForNodeZeroOnce(t *testing.T, mustPass bool) bool {
	cnt := 0
	for i, node := range nodes {
		for peer := range node.shh.peers {
			if peer.peer.ID() == nodes[0].server.Self().ID() {
				cnt++
				if peer.powRequirement != masterPow {
					if mustPass {
						t.Fatalf("node %d: failed to set the new pow requirement for node zero.", i)
					} else {
						return false
					}
				}
			}
		}
	}
	if cnt == 0 {
		t.Fatalf("looking for node zero: no matching peers found.")
	}
	return true
}

func checkPowExchange(t *testing.T) {
	for i, node := range nodes {
		for peer := range node.shh.peers {
			if peer.peer.ID() != nodes[0].server.Self().ID() {
				if peer.powRequirement != masterPow {
					t.Fatalf("node %d: failed to exchange pow requirement in round %d; expected %f, got %f",
						i, round, masterPow, peer.powRequirement)
				}
			}
		}
	}
}

func checkBloomFilterExchangeOnce(t *testing.T, mustPass bool) bool {
	for i, node := range nodes {
		for peer := range node.shh.peers {
			peer.bloomMu.Lock()
			equals := bytes.Equal(peer.bloomFilter, masterBloomFilter)
			peer.bloomMu.Unlock()
			if !equals {
				if mustPass {
					t.Fatalf("node %d: failed to exchange bloom filter requirement in round %d. \n%x expected \n%x got",
						i, round, masterBloomFilter, peer.bloomFilter)
				} else {
					return false
				}
			}
		}
	}

	return true
}

func checkBloomFilterExchange(t *testing.T) {
	const iterations = 200
	for j := 0; j < iterations; j++ {
		lastCycle := (j == iterations-1)
		ok := checkBloomFilterExchangeOnce(t, lastCycle)
		if ok {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
}

//two generic whisper node handshake
func TestPeerHandshakeWithTwoFullNode(t *testing.T) {
	w1 := Whisper{}
	p1 := newPeer(&w1, p2p.NewPeer(enode.ID{}, "test", []p2p.Cap{}), &rwStub{[]interface{}{ProtocolVersion, uint64(123), make([]byte, BloomFilterSize), false}})
	err := p1.handshake()
	if err != nil {
		t.Fatal()
	}
}

//two generic whisper node handshake. one don't send light flag
func TestHandshakeWithOldVersionWithoutLightModeFlag(t *testing.T) {
	w1 := Whisper{}
	p1 := newPeer(&w1, p2p.NewPeer(enode.ID{}, "test", []p2p.Cap{}), &rwStub{[]interface{}{ProtocolVersion, uint64(123), make([]byte, BloomFilterSize)}})
	err := p1.handshake()
	if err != nil {
		t.Fatal()
	}
}

//two light nodes handshake. restriction disabled
func TestTwoLightPeerHandshakeRestrictionOff(t *testing.T) {
	w1 := Whisper{}
	w1.settings.Store(restrictConnectionBetweenLightClientsIdx, false)
	w1.SetLightClientMode(true)
	p1 := newPeer(&w1, p2p.NewPeer(enode.ID{}, "test", []p2p.Cap{}), &rwStub{[]interface{}{ProtocolVersion, uint64(123), make([]byte, BloomFilterSize), true}})
	err := p1.handshake()
	if err != nil {
		t.FailNow()
	}
}

//two light nodes handshake. restriction enabled
func TestTwoLightPeerHandshakeError(t *testing.T) {
	w1 := Whisper{}
	w1.settings.Store(restrictConnectionBetweenLightClientsIdx, true)
	w1.SetLightClientMode(true)
	p1 := newPeer(&w1, p2p.NewPeer(enode.ID{}, "test", []p2p.Cap{}), &rwStub{[]interface{}{ProtocolVersion, uint64(123), make([]byte, BloomFilterSize), true}})
	err := p1.handshake()
	if err == nil {
		t.FailNow()
	}
}

type rwStub struct {
	payload []interface{}
}

func (stub *rwStub) ReadMsg() (p2p.Msg, error) {
	size, r, err := rlp.EncodeToReader(stub.payload)
	if err != nil {
		return p2p.Msg{}, err
	}
	return p2p.Msg{Code: statusCode, Size: uint32(size), Payload: r}, nil
}

func (stub *rwStub) WriteMsg(m p2p.Msg) error {
	return nil
}
