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
	eddsa "github.com/core-coin/go-goldilocks"
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
	"f973213b8dc2debd115d06c8b9478780d8881d6b7abd77787abea2c542a22a12043e0973b4c9a29ee87fccbec3b01e440da96c356e7d80202b",
	"e771190f75acf998879b399b18449cf0d56bb3041ec7681e40bd3a1c45d9aab1e73c3d3a123d64d08c207b24f267a8087e55831259e62716a1",
	"d6cff2205c0967a1ebfdd1fa0ebe4e732276b6d5be2eab8f3b68578b65a61ab60deb31a401814fa5e9cf1dde4e9ec71f7182de0be69529354c",
	"d3ad61749ba2d27b3df90eac1e6a0d92b9d92b320512dc603bd696a5413d3f1c53aa5ad9ab7ef255ba435a16001fb5d45793a89650d4782508",
	"f7edd052dc053feec54b96b94f690d7495b7f6e90bb1714340a833d9adf0ab776b1b55576d050e08ae71e4d1f0467e3599e64d49a7af8e190a",
	"4ca72453d4b218ee7cad26ffe11e3b87200a9dbaa0558fce755d5b0dfa0417fad526df1bbfb9ef0e8721cf49e4e0ec26aedd513e705bc22cc5",
	"3ebd6f12b41d4faecc091d98c0128e99fec0425ce2471306ba623bb3615018bce195bba4fb0a2497332874bd2040aa112d7a48ef9a604f36d0",
	"ee3f6057feb9c0f384ccafbf7aa07157d11c543d4e0b50fcde41d27365c142d05c85d72d9d79bcbe5b94b7c3b2d38986066774caacb90a37a7",
	"61eda6235dcf77eb7b206b97378cb70367524abad3ada6b801eb8b2413df22cdea0b332594639be27ce7fe3e71c165c25c994391f4661e15f6",
	"6e8e4928db08dfd4f4e804616636bf30a3f5a857b3a89a2f60489304185511b7034ec66e41df1fb94d11f59f4a33667e72510cacae78410fd8",
	"61917ceeb798fd7a5aad99878232723fe464ee0e6329f805ead8e34c9828f3b30a898185653c8cd18484dd8e2974a601f15e59d6e75f1915f4",
	"27aaea6c15a2207906fc4e1e4a60f5dcfe99a47bab98c9accb08885ac3c110694a8141c2bbc8b8eaf4bee0f47518795383009564fa1161085f",
	"0b4f9143d093dd2ed7179d33cdfb41844e5e27bc8f469b7713f6c87963b23d881c3dcbe7adcbb53d6ed382aed916cf688f2cd56c6e26e52564",
	"07b45ecd00056509fa8d340a882e80aae90da8dee2649933a7b55d6bf5ba8fba67932679a37bc56ce079f28d3be0504b4b67888052a70e2ea6",
	"eea4d9ee2d38b926e9d5df37b9cec37224c672e4440b41817b72c2d92aa8210f42050869717a8331177891cc9d515b6a7042312eff74890f4a",
	"0f10b73b0ed5727a12b95b5d87ac0860cabe6ffa7017345b02b49fbf5799c9aef9ab7b5753238500acc2ab932de91b12fe842f79157190397c",
	"cb6bf2e737b53faf069f025a77400d63ce20daa88d0f290e7b8c39ba6aa8d76b22f505f249cb3503ac60fae358e7d6606bbaacc1368b6a1a6d",
	"ddbe05178eb13896ee38f0d7148101d490552ae89459a02530b03e5a5c6f7121e2d9c98287e1d04e338336bd1553c045d37e73092f380a18d4",
	"2e7717af726631bbacb3e7a90c9f25fc6ef14eeae8facd054876073826774bf56296959fbf58f6507399af435949b0d53353696e5f161f0ef4",
	"ce5f7add74034a841a6586b5e0a2290c81a210bc4e7e3064248117746b4072e66f24bce47f383528626c832ef107492f47cc22c6b0fdc10138",
	"def11377abb23dedbf3675c6c8ec3c7233436fe55510b0b285ec681e88e7550e9cb4bdc7c0c966070fc7f69ace519c72efc9d5ac2ffa9b0d14",
	"f3e957aa084a030d7e6101e2331b4abe7f0d2fb432e56db8e1e7c9402e11dc163aa1cc89b654b0e9be53e2dc75cbff16377a1e7b9e4d9e0937",
	"ba6d89ae28bd45e47175a98d27b334e7fc59c62fd8b21a38be1dff681e09f9c4b50e7b216ec5774fcc30b345b3ded1d9db75085effe75f31e7",
	"3d53796e7570d829556b334f772fb21389eea38d93d1651e5a0b24b0b23e892c41f5e82f93a4d1f7ba1ff49ec18ff2a1d3d304fb53fc5d1aa3",
	"635e74f0ed57a4d466d08f1815c7c98157be0a99f8c8522dc75b19709297f95ee95c40051ed4be0a2abba4ddd249f35dab0a4f73788fdc2984",
	"8fa5ab9916d884c99f928dbdb554d7add48d5a289caa7c3b87d23279619ef2c0a331442aff0b9cd3b7fdb623f6274c3a1292f049fa266b2e45",
	"5ffb670e50db6d4b60bc743c097ea3eb812615f1ff24cc726bc575a0db245f9a83de7f1e97f51541fe0ed01e0c3f1c01cf5f9fb00a3963364b",
	"b0c354a5b155cb640168e8c0995e1cdb92fc9eae02d3c4f4c85da44ca73cddcddd46c4512ab5c1964d99ac87d71721d107adec9e72a82c30de",
	"fdc3e27de9920693b333e3cc911d8ad25d8f1395689e0f9c8730ab3dfecbd66a2957f30ab257d28dab1d1b3d7997ff1d2c42a695fddeb02692",
	"4c9a00e6699d32910c0e9cf6f0afea214f749526fca86fdbdef0964d53a5adbea15edd803cea6ee67a178df35f0f66fa76b55b5237c8e2375d",
	"3fb9d032b1a65199c738535ffa0156f088d5711f6643f347afcbea4a6022d34a688a2e85d50c867bd74cda0b0c4963c52d072db6474b6f3165",
	"084e44c699d643f0314461e3f4189d0951fbea92954cb021863f85f95caed5955522a5e132b654b01d3bfdbd0d06438028f88e61687a6214b2",
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
			pub := eddsa.Ed448DerivePublicKey(*peerNodeId)
			peer := enode.NewV4(&pub, address.IP, address.Port, address.Port)
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
