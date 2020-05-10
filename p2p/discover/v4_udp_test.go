// Copyright 2015 The go-core Authors
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
	"bytes"
	eddsa "github.com/core-coin/eddsa"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"math/rand"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/internal/testlog"
	"github.com/core-coin/go-core/log"
	"github.com/core-coin/go-core/p2p/enode"
	"github.com/core-coin/go-core/p2p/enr"
	"github.com/core-coin/go-core/rlp"
)

func init() {
	spew.Config.DisableMethods = true
}

// shared test variables
var (
	futureExp          = uint64(time.Now().Add(10 * time.Hour).Unix())
	testTarget         = encPubkey{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1}
	testRemote         = rpcEndpoint{IP: net.ParseIP("1.1.1.1").To4(), UDP: 1, TCP: 2}
	testLocalAnnounced = rpcEndpoint{IP: net.ParseIP("2.2.2.2").To4(), UDP: 3, TCP: 4}
	testLocal          = rpcEndpoint{IP: net.ParseIP("3.3.3.3").To4(), UDP: 5, TCP: 6}
)

type udpTest struct {
	t                   *testing.T
	pipe                *dgramPipe
	table               *Table
	db                  *enode.DB
	udp                 *UDPv4
	sent                [][]byte
	localkey, remotekey *eddsa.PrivateKey
	remoteaddr          *net.UDPAddr
}

func newUDPTest(t *testing.T) *udpTest {
	test := &udpTest{
		t:          t,
		pipe:       newpipe(),
		localkey:   newkey(),
		remotekey:  newkey(),
		remoteaddr: &net.UDPAddr{IP: net.IP{10, 0, 1, 99}, Port: 30300},
	}

	test.db, _ = enode.OpenDB("")
	ln := enode.NewLocalNode(test.db, test.localkey)
	test.udp, _ = ListenV4(test.pipe, ln, Config{
		PrivateKey: test.localkey,
		Log:        testlog.Logger(t, log.LvlTrace),
	})
	test.table = test.udp.tab
	// Wait for initial refresh so the table doesn't send unexpected findnode.
	<-test.table.initDone
	return test
}

func (test *udpTest) close() {
	test.udp.Close()
	test.db.Close()
}

// handles a packet as if it had been sent to the transport.
func (test *udpTest) packetIn(wantError error, data packetV4) {
	test.t.Helper()

	test.packetInFrom(wantError, test.remotekey, test.remoteaddr, data)
}

// handles a packet as if it had been sent to the transport by the key/endpoint.
func (test *udpTest) packetInFrom(wantError error, key *eddsa.PrivateKey, addr *net.UDPAddr, data packetV4) {
	test.t.Helper()

	enc, _, err := test.udp.encode(key, data)
	if err != nil {
		test.t.Errorf("%s encode error: %v", data.name(), err)
	}
	test.sent = append(test.sent, enc)
	if err = test.udp.handlePacket(addr, enc); err != wantError {
		test.t.Errorf("error mismatch: got %q, want %q", err, wantError)
	}
}

// waits for a packet to be sent by the transport.
// validate should have type func(X, *net.UDPAddr, []byte), where X is a packet type.
func (test *udpTest) waitPacketOut(validate interface{}) (closed bool) {
	test.t.Helper()

	dgram, ok := test.pipe.receive()
	if !ok {
		return true
	}
	p, _, hash, err := decodeV4(dgram.data)
	if err != nil {
		test.t.Errorf("sent packet decode error: %v", err)
		return false
	}
	fn := reflect.ValueOf(validate)
	exptype := fn.Type().In(0)
	if !reflect.TypeOf(p).AssignableTo(exptype) {
		test.t.Errorf("sent packet type mismatch, got: %v, want: %v", reflect.TypeOf(p), exptype)
		return false
	}
	fn.Call([]reflect.Value{reflect.ValueOf(p), reflect.ValueOf(&dgram.to), reflect.ValueOf(hash)})
	return false
}

func TestUDPv4_packetErrors(t *testing.T) {
	test := newUDPTest(t)
	defer test.close()

	test.packetIn(errExpired, &pingV4{From: testRemote, To: testLocalAnnounced, Version: 4})
	test.packetIn(errUnsolicitedReply, &pongV4{ReplyTok: []byte{}, Expiration: futureExp})
	test.packetIn(errUnknownNode, &findnodeV4{Expiration: futureExp})
	test.packetIn(errUnsolicitedReply, &neighborsV4{Expiration: futureExp})
}

func TestUDPv4_pingTimeout(t *testing.T) {
	t.Parallel()
	test := newUDPTest(t)
	defer test.close()

	key := newkey()
	toaddr := &net.UDPAddr{IP: net.ParseIP("1.2.3.4"), Port: 2222}
	node := enode.NewV4(&key.PublicKey, toaddr.IP, 0, toaddr.Port)
	if _, err := test.udp.ping(node); err != errTimeout {
		t.Error("expected timeout error, got", err)
	}
}

type testPacket byte

func (req testPacket) kind() byte   { return byte(req) }
func (req testPacket) name() string { return "" }
func (req testPacket) preverify(*UDPv4, *net.UDPAddr, enode.ID, encPubkey) error {
	return nil
}
func (req testPacket) handle(*UDPv4, *net.UDPAddr, enode.ID, []byte) {
}

func TestUDPv4_responseTimeouts(t *testing.T) {
	t.Parallel()
	test := newUDPTest(t)
	defer test.close()

	rand.Seed(time.Now().UnixNano())
	randomDuration := func(max time.Duration) time.Duration {
		return time.Duration(rand.Int63n(int64(max)))
	}

	var (
		nReqs      = 200
		nTimeouts  = 0                       // number of requests with ptype > 128
		nilErr     = make(chan error, nReqs) // for requests that get a reply
		timeoutErr = make(chan error, nReqs) // for requests that time out
	)
	for i := 0; i < nReqs; i++ {
		// Create a matcher for a random request in udp.loop. Requests
		// with ptype <= 128 will not get a reply and should time out.
		// For all other requests, a reply is scheduled to arrive
		// within the timeout window.
		p := &replyMatcher{
			ptype:    byte(rand.Intn(255)),
			callback: func(interface{}) (bool, bool) { return true, true },
		}
		binary.BigEndian.PutUint64(p.from[:], uint64(i))
		if p.ptype <= 128 {
			p.errc = timeoutErr
			test.udp.addReplyMatcher <- p
			nTimeouts++
		} else {
			p.errc = nilErr
			test.udp.addReplyMatcher <- p
			time.AfterFunc(randomDuration(60*time.Millisecond), func() {
				if !test.udp.handleReply(p.from, p.ip, testPacket(p.ptype)) {
					t.Logf("not matched: %v", p)
				}
			})
		}
		time.Sleep(randomDuration(30 * time.Millisecond))
	}

	// Check that all timeouts were delivered and that the rest got nil errors.
	// The replies must be delivered.
	var (
		recvDeadline        = time.After(20 * time.Second)
		nTimeoutsRecv, nNil = 0, 0
	)
	for i := 0; i < nReqs; i++ {
		select {
		case err := <-timeoutErr:
			if err != errTimeout {
				t.Fatalf("got non-timeout error on timeoutErr %d: %v", i, err)
			}
			nTimeoutsRecv++
		case err := <-nilErr:
			if err != nil {
				t.Fatalf("got non-nil error on nilErr %d: %v", i, err)
			}
			nNil++
		case <-recvDeadline:
			t.Fatalf("exceeded recv deadline")
		}
	}
	if nTimeoutsRecv != nTimeouts {
		t.Errorf("wrong number of timeout errors received: got %d, want %d", nTimeoutsRecv, nTimeouts)
	}
	if nNil != nReqs-nTimeouts {
		t.Errorf("wrong number of successful replies: got %d, want %d", nNil, nReqs-nTimeouts)
	}
}

func TestUDPv4_findnodeTimeout(t *testing.T) {
	t.Parallel()
	test := newUDPTest(t)
	defer test.close()

	toaddr := &net.UDPAddr{IP: net.ParseIP("1.2.3.4"), Port: 2222}
	toid := enode.ID{1, 2, 3, 4}
	target := encPubkey{4, 5, 6, 7}
	result, err := test.udp.findnode(toid, toaddr, target)
	if err != errTimeout {
		t.Error("expected timeout error, got", err)
	}
	if len(result) > 0 {
		t.Error("expected empty result, got", result)
	}
}

func TestUDPv4_findnode(t *testing.T) {
	test := newUDPTest(t)
	defer test.close()

	// put a few nodes into the table. their exact
	// distribution shouldn't matter much, although we need to
	// take care not to overflow any bucket.
	nodes := &nodesByDistance{target: testTarget.id()}
	live := make(map[enode.ID]bool)
	numCandidates := 2 * bucketSize
	for i := 0; i < numCandidates; i++ {
		key := newkey()
		ip := net.IP{10, 13, 0, byte(i)}
		n := wrapNode(enode.NewV4(&key.PublicKey, ip, 0, 2000))
		// Ensure half of table content isn't verified live yet.
		if i > numCandidates/2 {
			n.livenessChecks = 1
			live[n.ID()] = true
		}
		nodes.push(n, numCandidates)
	}
	fillTable(test.table, nodes.entries)

	// ensure there's a bond with the test node,
	// findnode won't be accepted otherwise.
	remoteID := encodePubkey(&test.remotekey.PublicKey).id()
	test.table.db.UpdateLastPongReceived(remoteID, test.remoteaddr.IP, time.Now())

	// check that closest neighbors are returned.
	expected := test.table.closest(testTarget.id(), bucketSize, true)
	test.packetIn(nil, &findnodeV4{Target: testTarget, Expiration: futureExp})
	waitNeighbors := func(want []*node) {
		test.waitPacketOut(func(p *neighborsV4, to *net.UDPAddr, hash []byte) {
			if len(p.Nodes) != len(want) {
				t.Errorf("wrong number of results: got %d, want %d", len(p.Nodes), bucketSize)
			}
			for i, n := range p.Nodes {
				if n.ID.id() != want[i].ID() {
					t.Errorf("result mismatch at %d:\n  got:  %v\n  want: %v", i, n, expected.entries[i])
				}
				if !live[n.ID.id()] {
					t.Errorf("result includes dead node %v", n.ID.id())
				}
			}
		})
	}
	// Receive replies.
	want := expected.entries
	if len(want) > maxNeighbors {
		waitNeighbors(want[:maxNeighbors])
		want = want[maxNeighbors:]
	}
	waitNeighbors(want)
}

func TestUDPv4_findnodeMultiReply(t *testing.T) {
	test := newUDPTest(t)
	defer test.close()

	rid := enode.PubkeyToIDV4(&test.remotekey.PublicKey)
	test.table.db.UpdateLastPingReceived(rid, test.remoteaddr.IP, time.Now())

	// queue a pending findnode request
	resultc, errc := make(chan []*node), make(chan error)
	go func() {
		rid := encodePubkey(&test.remotekey.PublicKey).id()
		ns, err := test.udp.findnode(rid, test.remoteaddr, testTarget)
		if err != nil && len(ns) == 0 {
			errc <- err
		} else {
			resultc <- ns
		}
	}()

	// wait for the findnode to be sent.
	// after it is sent, the transport is waiting for a reply
	test.waitPacketOut(func(p *findnodeV4, to *net.UDPAddr, hash []byte) {
		if p.Target != testTarget {
			t.Errorf("wrong target: got %v, want %v", p.Target, testTarget)
		}
	})

	// send the reply as two packets.
	list := []*node{
		wrapNode(enode.MustParse("enode://ba85011c70bcc5c04d8607d3a0ed29aa6179c092cbdda10d5d32684fb33ed01bd94f588ca8f91ac48318087dcb02eaf36773a7a453f0eedd@10.0.1.16:30303?discport=30304")),
		wrapNode(enode.MustParse("enode://81fa361d25f157cd421c60dcc28d8dac5ef6a89476633339c5df30287474520caca09627da18543d9079b5b288698b542d56167aa5c09111@10.0.1.16:30303")),
		wrapNode(enode.MustParse("enode://9bffefd833d53fac8e652415f4973bee289e8b1a5c6c4cbe70abf817ce8a64cee11b823b66a987f51aaa9fba0d6a91b3e6bf0d5a5d1042de@10.0.1.36:30301?discport=17")),
		wrapNode(enode.MustParse("enode://1b5b4aa662d7cb44a7221bfba67302590b643028197a7d5214790f3bac7aaa4a3241be9e83c09cf1f6c69d007c634faae3dc1b1221793e84@10.0.1.16:30303")),
	}
	rpclist := make([]rpcNode, len(list))
	for i := range list {
		rpclist[i] = nodeToRPC(list[i])
	}
	test.packetIn(nil, &neighborsV4{Expiration: futureExp, Nodes: rpclist[:2]})
	test.packetIn(nil, &neighborsV4{Expiration: futureExp, Nodes: rpclist[2:]})

	// check that the sent neighbors are all returned by findnode
	select {
	case result := <-resultc:
		want := append(list[:2], list[3:]...)
		if !reflect.DeepEqual(result, want) {
			t.Errorf("neighbors mismatch:\n  got:  %v\n  want: %v", result, want)
		}
	case err := <-errc:
		t.Errorf("findnode error: %v", err)
	case <-time.After(5 * time.Second):
		t.Error("findnode did not return within 5 seconds")
	}
}

// This test checks that reply matching of pong verifies the ping hash.
func TestUDPv4_pingMatch(t *testing.T) {
	test := newUDPTest(t)
	defer test.close()

	randToken := make([]byte, 32)
	crand.Read(randToken)

	test.packetIn(nil, &pingV4{From: testRemote, To: testLocalAnnounced, Version: 4, Expiration: futureExp})
	test.waitPacketOut(func(*pongV4, *net.UDPAddr, []byte) {})
	test.waitPacketOut(func(*pingV4, *net.UDPAddr, []byte) {})
	test.packetIn(errUnsolicitedReply, &pongV4{ReplyTok: randToken, To: testLocalAnnounced, Expiration: futureExp})
}

// This test checks that reply matching of pong verifies the sender IP address.
func TestUDPv4_pingMatchIP(t *testing.T) {
	test := newUDPTest(t)
	defer test.close()

	test.packetIn(nil, &pingV4{From: testRemote, To: testLocalAnnounced, Version: 4, Expiration: futureExp})
	test.waitPacketOut(func(*pongV4, *net.UDPAddr, []byte) {})

	test.waitPacketOut(func(p *pingV4, to *net.UDPAddr, hash []byte) {
		wrongAddr := &net.UDPAddr{IP: net.IP{33, 44, 1, 2}, Port: 30000}
		test.packetInFrom(errUnsolicitedReply, test.remotekey, wrongAddr, &pongV4{
			ReplyTok:   hash,
			To:         testLocalAnnounced,
			Expiration: futureExp,
		})
	})
}

func TestUDPv4_successfulPing(t *testing.T) {
	test := newUDPTest(t)
	added := make(chan *node, 1)
	test.table.nodeAddedHook = func(n *node) { added <- n }
	defer test.close()

	// The remote side sends a ping packet to initiate the exchange.
	go test.packetIn(nil, &pingV4{From: testRemote, To: testLocalAnnounced, Version: 4, Expiration: futureExp})

	// The ping is replied to.
	test.waitPacketOut(func(p *pongV4, to *net.UDPAddr, hash []byte) {
		pinghash := test.sent[0][:macSize]
		if !bytes.Equal(p.ReplyTok, pinghash) {
			t.Errorf("got pong.ReplyTok %x, want %x", p.ReplyTok, pinghash)
		}
		wantTo := rpcEndpoint{
			// The mirrored UDP address is the UDP packet sender
			IP: test.remoteaddr.IP, UDP: uint16(test.remoteaddr.Port),
			// The mirrored TCP port is the one from the ping packet
			TCP: testRemote.TCP,
		}
		if !reflect.DeepEqual(p.To, wantTo) {
			t.Errorf("got pong.To %v, want %v", p.To, wantTo)
		}
	})

	// Remote is unknown, the table pings back.
	test.waitPacketOut(func(p *pingV4, to *net.UDPAddr, hash []byte) {
		if !reflect.DeepEqual(p.From, test.udp.ourEndpoint()) {
			t.Errorf("got ping.From %#v, want %#v", p.From, test.udp.ourEndpoint())
		}
		wantTo := rpcEndpoint{
			// The mirrored UDP address is the UDP packet sender.
			IP:  test.remoteaddr.IP,
			UDP: uint16(test.remoteaddr.Port),
			TCP: 0,
		}
		if !reflect.DeepEqual(p.To, wantTo) {
			t.Errorf("got ping.To %v, want %v", p.To, wantTo)
		}
		test.packetIn(nil, &pongV4{ReplyTok: hash, Expiration: futureExp})
	})

	// The node should be added to the table shortly after getting the
	// pong packet.
	select {
	case n := <-added:
		rid := encodePubkey(&test.remotekey.PublicKey).id()
		if n.ID() != rid {
			t.Errorf("node has wrong ID: got %v, want %v", n.ID(), rid)
		}
		if !n.IP().Equal(test.remoteaddr.IP) {
			t.Errorf("node has wrong IP: got %v, want: %v", n.IP(), test.remoteaddr.IP)
		}
		if n.UDP() != test.remoteaddr.Port {
			t.Errorf("node has wrong UDP port: got %v, want: %v", n.UDP(), test.remoteaddr.Port)
		}
		if n.TCP() != int(testRemote.TCP) {
			t.Errorf("node has wrong TCP port: got %v, want: %v", n.TCP(), testRemote.TCP)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("node was not added within 2 seconds")
	}
}

// This test checks that EIP-868 requests work.
func TestUDPv4_EIP868(t *testing.T) {
	test := newUDPTest(t)
	defer test.close()

	test.udp.localNode.Set(enr.WithEntry("foo", "bar"))
	wantNode := test.udp.localNode.Node()

	// ENR requests aren't allowed before endpoint proof.
	test.packetIn(errUnknownNode, &enrRequestV4{Expiration: futureExp})

	// Perform endpoint proof and check for sequence number in packet tail.
	test.packetIn(nil, &pingV4{Expiration: futureExp})
	test.waitPacketOut(func(p *pongV4, addr *net.UDPAddr, hash []byte) {
		if seq := seqFromTail(p.Rest); seq != wantNode.Seq() {
			t.Errorf("wrong sequence number in pong: %d, want %d", seq, wantNode.Seq())
		}
	})
	test.waitPacketOut(func(p *pingV4, addr *net.UDPAddr, hash []byte) {
		if seq := seqFromTail(p.Rest); seq != wantNode.Seq() {
			t.Errorf("wrong sequence number in ping: %d, want %d", seq, wantNode.Seq())
		}
		test.packetIn(nil, &pongV4{Expiration: futureExp, ReplyTok: hash})
	})

	// Request should work now.
	test.packetIn(nil, &enrRequestV4{Expiration: futureExp})
	test.waitPacketOut(func(p *enrResponseV4, addr *net.UDPAddr, hash []byte) {
		n, err := enode.New(enode.ValidSchemes, &p.Record)
		if err != nil {
			t.Fatalf("invalid record: %v", err)
		}
		if !reflect.DeepEqual(n, wantNode) {
			t.Fatalf("wrong node in enrResponse: %v", n)
		}
	})
}

// EIP-8 test vectors.
var testPackets = []struct {
	input      string
	wantPacket interface{}
}{
	{
		input: "3275085fc5a72393567533599a3425fb458e10432baa71ea276bc6cb8a45693538d9c11498a90d451a30b876dd6ee9394deb99a56f34c8c6c5b08cd104e2f6df9b29c884af231c49395fa3bd53444202e7e8c4b55396478cb4f2028e93e83832339cce3376f715252530a9a13e86ec6d0cb94601333bfef3c9fbf8023043ac772958c3cb535f55518146489b4c1fd50e46cdbc7b6001d869d622e93baa4004f9895c274b0199262afc2311fdf6c302e288fe75965253c0df585c19ca1cc3adbf4130c667a9d658f401ea04cb847f000001820cfa8215a8d790000000000000000000000000000000018208ae820d058443b9a355",
		wantPacket: &pingV4{
			Version:    4,
			From:       rpcEndpoint{net.ParseIP("127.0.0.1").To4(), 3322, 5544},
			To:         rpcEndpoint{net.ParseIP("::1"), 2222, 3333},
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{},
		},
	},
	{
		input: "636b69c28fb9456875e3dc449fda0c92f0aae8ea67bcd89db8b0c338a3e0fe8f266f0f792bd6c427bb54511f3b0cef7e92ce8ad222cf4c4f766218051e90b20229235bc0cc9bf36ed529ef120abc4027820ae29ff0cea81696c48d4893ad2abca5fa2dd19e7b6ea5a9389cfcf9f87609492d58b87b8ebd8d5fd0eff70f3b6e79a18ff1a6d09b4ff0a75533d3af36cf2d46cdbc7b6001d869d622e93baa4004f9895c274b0199262afc2311fdf6c302e288fe75965253c0df585c19ca1cc3adbf4130c667a9d658f401ec04cb847f000001820cfa8215a8d790000000000000000000000000000000018208ae820d058443b9a3550102",
		wantPacket: &pingV4{
			Version:    4,
			From:       rpcEndpoint{net.ParseIP("127.0.0.1").To4(), 3322, 5544},
			To:         rpcEndpoint{net.ParseIP("::1"), 2222, 3333},
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0x01}, {0x02}},
		},
	},
	{
		input: "03862f53ca094d040f2be6b9cc17fbf32a85868f166a5227d2ed8cc7a5ad96d849b685ed3fadd5c60d0e8f0efeb62f1f383fe046a38fcfce91ded6cf80399a5a9168e4d2bc274cc802ae197ed54399ce63855a858d1fed4c8eeb2e17d46d666bde2f07d873517f530f444ae62c3f8ba5b2c15985ecccaabd4228f67ea84330db8dff4cbf7a3f9f10567ef24e429bef1b46cdbc7b6001d869d622e93baa4004f9895c274b0199262afc2311fdf6c302e288fe75965253c0df585c19ca1cc3adbf4130c667a9d658f401f83e82022bd79020010db83c4d001500000000abcdef12820cfa8215a8d79020010db885a308d313198a2e037073488208ae82823a8443b9a355c50102030405",
		wantPacket: &pingV4{
			Version:    555,
			From:       rpcEndpoint{net.ParseIP("2001:db8:3c4d:15::abcd:ef12"), 3322, 5544},
			To:         rpcEndpoint{net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348"), 2222, 33338},
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0xC5, 0x01, 0x02, 0x03, 0x04, 0x05}},
		},
	},
	{
		input: "3d36cff7a6ab60172858e522c2c31211b8691d06885e3aeec4a0197677c5d565f4633bdaa5454670c68cb92a48d7cc7608ff6fa9ab17af4f2f5e2ff01e224d19b3ea7186a450b2b9385a86fb1415141a70ce0fd68ea0aefb82319104746cfd57aa6f36a856e03e023815e0a7d14e0d92f51b7446d665af350914e17b38f7fe71b52002b02c890beef4b25e2444d4202246cdbc7b6001d869d622e93baa4004f9895c274b0199262afc2311fdf6c302e288fe75965253c0df585c19ca1cc3adbf4130c667a9d658f402f846d79020010db885a308d313198a2e037073488208ae82823aa0fbc914b16819237dcd8801d7e53f69e9719adecb3cc0e790c57e91ca4461c9548443b9a355c6010203c2040506",
		wantPacket: &pongV4{
			To:         rpcEndpoint{net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348"), 2222, 33338},
			ReplyTok:   common.Hex2Bytes("fbc914b16819237dcd8801d7e53f69e9719adecb3cc0e790c57e91ca4461c954"),
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0xC6, 0x01, 0x02, 0x03, 0xC2, 0x04, 0x05}, {0x06}},
		},
	},
	{
		input: "b4ad39f74f172352bd57f38bf8e389b6ce9b652b383459f0f572b8f568c3e517bafd1ecf5baf1ba421bfd83a25a54c2c2e5c434e44aff59ba572e3d48d0e513854eae1bf1628a8f49941c0bc5c57f844f9bdd82cc49b76e695fc7ab1d28eaefab3d9c86c0033e3c977529366542052aa562890949b39428288a0aead47002ab74635f77aa148bb7f97271b98eb567f3c46cdbc7b6001d869d622e93baa4004f9895c274b0199262afc2311fdf6c302e288fe75965253c0df585c19ca1cc3adbf4130c667a9d658f403f846b8384aa8946ebf664270106abd7aea3a27e66124c04deb391de0997be90f973d5369fd287442b79e7553b642aaca22159567b503c36e128918478443b9a35582999983999999",
		wantPacket: &findnodeV4{
			Target:     hexEncPubkey("4aa8946ebf664270106abd7aea3a27e66124c04deb391de0997be90f973d5369fd287442b79e7553b642aaca22159567b503c36e12891847"),
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0x82, 0x99, 0x99}, {0x83, 0x99, 0x99, 0x99}},
		},
	},
	{
		input: "c1f75777a854d39fd0cf54e9c807141d78c7bb824aabd68323f86bd76448e801a0702817b3c5e013d04292df77d69d4b7ce3de7c60da848a2d837396ca072caebc0f4039be552371dc700cafa9537b228fc56ea3d0d8435c0e6564e4dfbb345803122d922f3c77e4f7e56e8d123c55f230e5bc84a51a607a573c96099dd3ddb8832078abd78a0c205d4f87008513252946cdbc7b6001d869d622e93baa4004f9895c274b0199262afc2311fdf6c302e288fe75965253c0df585c19ca1cc3adbf4130c667a9d658f404f9013bf90130f845846321163782115c82115db8384aa8946ebf664270106abd7aea3a27e66124c04deb391de0997be90f973d5369fd287442b79e7553b642aaca22159567b503c36e12891847f84184010203040101b8385acf8e211e3d3e2ba310afa91edc15389a9cb2e59525774646dc46030d18d880b5f0ef616f842231a355e725589dd45a2f2677d028b6fe46f8519020010db83c4d001500000000abcdef12820d05820d05b8384b9fcfcbeb73288841583b311b5ac405e8cba059953db4f7cfa3008a7b41099f8713ba6195ed0168607342675f77d26d5a6088e540401ac2f8519020010db885a308d313198a2e037073488203e78203e8b838d3c777805908f662ce69981006a68da37b3e2105ab9a96b28b79832f7f40715e77d1c589f14bc914d7915b8976317c78603b732a7c7f54538443b9a355010203",
		wantPacket: &neighborsV4{
			Nodes: []rpcNode{
				{
					ID:  hexEncPubkey("4aa8946ebf664270106abd7aea3a27e66124c04deb391de0997be90f973d5369fd287442b79e7553b642aaca22159567b503c36e12891847"),
					IP:  net.ParseIP("99.33.22.55").To4(),
					UDP: 4444,
					TCP: 4445,
				},
				{
					ID:  hexEncPubkey("5acf8e211e3d3e2ba310afa91edc15389a9cb2e59525774646dc46030d18d880b5f0ef616f842231a355e725589dd45a2f2677d028b6fe46"),
					IP:  net.ParseIP("1.2.3.4").To4(),
					UDP: 1,
					TCP: 1,
				},
				{
					ID:  hexEncPubkey("4b9fcfcbeb73288841583b311b5ac405e8cba059953db4f7cfa3008a7b41099f8713ba6195ed0168607342675f77d26d5a6088e540401ac2"),
					IP:  net.ParseIP("2001:db8:3c4d:15::abcd:ef12"),
					UDP: 3333,
					TCP: 3333,
				},
				{
					ID:  hexEncPubkey("d3c777805908f662ce69981006a68da37b3e2105ab9a96b28b79832f7f40715e77d1c589f14bc914d7915b8976317c78603b732a7c7f5453"),
					IP:  net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348"),
					UDP: 999,
					TCP: 1000,
				},
			},
			Expiration: 1136239445,
			Rest:       []rlp.RawValue{{0x01}, {0x02}, {0x03}},
		},
	},
}

func TestUDPv4_forwardCompatibility(t *testing.T) {
	testkey, _ := crypto.HexToEDDSA("835bbff17efac2c97895784041c507959cdb9e45c599cc205e453a962c11c09ac8834f6524d0842cc469db2afcc0424ca4afc42968d3441846cdbc7b6001d869d622e93baa4004f9895c274b0199262afc2311fdf6c302e288fe75965253c0df585c19ca1cc3adbf4130c667a9d658f459d242e82ba3981fb7004b02568a6750a604e54bd592b5b95aabc553bfde613f")
	wantNodeKey := encodePubkey(&testkey.PublicKey)

	for _, test := range testPackets {
		input, err := hex.DecodeString(test.input)
		if err != nil {
			t.Fatalf("invalid hex: %s", test.input)
		}
		packet, nodekey, _, err := decodeV4(input)
		if err != nil {
			t.Errorf("did not accept packet %s\n%v", test.input, err)
			continue
		}
		if !reflect.DeepEqual(packet, test.wantPacket) {
			t.Errorf("got %s\nwant %s", spew.Sdump(packet), spew.Sdump(test.wantPacket))
		}
		if nodekey != wantNodeKey {
			t.Errorf("got id %v\nwant id %v", nodekey, wantNodeKey)
		}
	}
}

// dgramPipe is a fake UDP socket. It queues all sent datagrams.
type dgramPipe struct {
	mu      *sync.Mutex
	cond    *sync.Cond
	closing chan struct{}
	closed  bool
	queue   []dgram
}

type dgram struct {
	to   net.UDPAddr
	data []byte
}

func newpipe() *dgramPipe {
	mu := new(sync.Mutex)
	return &dgramPipe{
		closing: make(chan struct{}),
		cond:    &sync.Cond{L: mu},
		mu:      mu,
	}
}

// WriteToUDP queues a datagram.
func (c *dgramPipe) WriteToUDP(b []byte, to *net.UDPAddr) (n int, err error) {
	msg := make([]byte, len(b))
	copy(msg, b)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, errors.New("closed")
	}
	c.queue = append(c.queue, dgram{*to, b})
	c.cond.Signal()
	return len(b), nil
}

// ReadFromUDP just hangs until the pipe is closed.
func (c *dgramPipe) ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error) {
	<-c.closing
	return 0, nil, io.EOF
}

func (c *dgramPipe) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		close(c.closing)
		c.closed = true
	}
	c.cond.Broadcast()
	return nil
}

func (c *dgramPipe) LocalAddr() net.Addr {
	return &net.UDPAddr{IP: testLocal.IP, Port: int(testLocal.UDP)}
}

func (c *dgramPipe) receive() (dgram, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for len(c.queue) == 0 && !c.closed {
		c.cond.Wait()
	}
	if c.closed {
		return dgram{}, false
	}
	p := c.queue[0]
	copy(c.queue, c.queue[1:])
	c.queue = c.queue[:len(c.queue)-1]
	return p, true
}
