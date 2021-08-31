// Copyright 2015 by the Authors
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

package p2p

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/core-coin/ed448"
	"io"
	"io/ioutil"
	"net"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/p2p/simulations/pipes"
	"github.com/core-coin/go-core/rlp"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/crypto/sha3"
)

func TestSharedSecret(t *testing.T) {
	prv0, _ := crypto.GenerateKey(rand.Reader) // = eddsa.GenerateKey(crypto.S256(), rand.Reader)
	pub0 := ed448.Ed448DerivePublicKey(prv0)
	prv1, _ := crypto.GenerateKey(rand.Reader)
	pub1 := ed448.Ed448DerivePublicKey(prv1)

	ss0 := crypto.ComputeSecret(prv0, pub1)
	ss1 := crypto.ComputeSecret(prv1, pub0)
	t.Logf("Secret:\n%v %x\n%v %x", len(ss0), ss0, len(ss0), ss1)
	if !bytes.Equal(ss0, ss1) {
		t.Errorf("dont match :(")
	}
}

func TestEncHandshake(t *testing.T) {
	for i := 0; i < 10; i++ {
		start := time.Now()
		if err := testEncHandshake(nil); err != nil {
			t.Fatalf("i=%d %v", i, err)
		}
		t.Logf("(without token) %d %v\n", i+1, time.Since(start))
	}
	for i := 0; i < 10; i++ {
		tok := make([]byte, shaLen)
		rand.Reader.Read(tok)
		start := time.Now()
		if err := testEncHandshake(tok); err != nil {
			t.Fatalf("i=%d %v", i, err)
		}
		t.Logf("(with token) %d %v\n", i+1, time.Since(start))
	}
}

func testEncHandshake(token []byte) error {
	type result struct {
		side   string
		pubkey ed448.PublicKey
		err    error
	}
	var (
		prv0, _  = crypto.GenerateKey(rand.Reader)
		prv1, _  = crypto.GenerateKey(rand.Reader)
		fd0, fd1 = net.Pipe()
		c0, c1   = newRLPX(fd0).(*rlpx), newRLPX(fd1).(*rlpx)
		output   = make(chan result)
	)

	go func() {
		r := result{side: "initiator"}
		defer func() { output <- r }()
		defer fd0.Close()

		pub := ed448.Ed448DerivePublicKey(prv1)
		r.pubkey, r.err = c0.doEncHandshake(prv0, pub)
		if r.err != nil {
			return
		}
		if !reflect.DeepEqual(r.pubkey, pub) {
			r.err = fmt.Errorf("remote pubkey mismatch: got %v, want: %v", r.pubkey, &pub)
		}
	}()
	go func() {
		r := result{side: "receiver"}
		defer func() { output <- r }()
		defer fd1.Close()

		r.pubkey, r.err = c1.doEncHandshake(prv1, ed448.PublicKey{})
		if r.err != nil {
			return
		}
		pub := ed448.Ed448DerivePublicKey(prv0)
		if !reflect.DeepEqual(r.pubkey, pub) {
			r.err = fmt.Errorf("remote ID mismatch: got %v, want: %v", r.pubkey, &pub)
		}
	}()

	// wait for results from both sides
	r1, r2 := <-output, <-output
	if r1.err != nil {
		return fmt.Errorf("%s side error: %v", r1.side, r1.err)
	}
	if r2.err != nil {
		return fmt.Errorf("%s side error: %v", r2.side, r2.err)
	}

	// compare derived secrets
	if !reflect.DeepEqual(c0.rw.egressMAC, c1.rw.ingressMAC) {
		return fmt.Errorf("egress mac mismatch:\n c0.rw: %#v\n c1.rw: %#v", c0.rw.egressMAC, c1.rw.ingressMAC)
	}
	if !reflect.DeepEqual(c0.rw.ingressMAC, c1.rw.egressMAC) {
		return fmt.Errorf("ingress mac mismatch:\n c0.rw: %#v\n c1.rw: %#v", c0.rw.ingressMAC, c1.rw.egressMAC)
	}
	if !reflect.DeepEqual(c0.rw.enc, c1.rw.enc) {
		return fmt.Errorf("enc cipher mismatch:\n c0.rw: %#v\n c1.rw: %#v", c0.rw.enc, c1.rw.enc)
	}
	if !reflect.DeepEqual(c0.rw.dec, c1.rw.dec) {
		return fmt.Errorf("dec cipher mismatch:\n c0.rw: %#v\n c1.rw: %#v", c0.rw.dec, c1.rw.dec)
	}
	return nil
}

func TestProtocolHandshake(t *testing.T) {
	var (
		prv0, _ = crypto.GenerateKey(rand.Reader)
		pubb0   = ed448.Ed448DerivePublicKey(prv0)
		pub0    = crypto.FromEDDSAPub(pubb0)
		hs0     = &protoHandshake{Version: 3, ID: pub0, Caps: []Cap{{"a", 0}, {"b", 2}}}

		prv1, _ = crypto.GenerateKey(rand.Reader)
		pubb1   = ed448.Ed448DerivePublicKey(prv1)
		pub1    = crypto.FromEDDSAPub(pubb1)
		hs1     = &protoHandshake{Version: 3, ID: pub1, Caps: []Cap{{"c", 1}, {"d", 3}}}

		wg sync.WaitGroup
	)

	fd0, fd1, err := pipes.TCPPipe()
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(2)
	go func() {
		defer wg.Done()
		defer fd0.Close()
		rlpx := newRLPX(fd0)
		rpubkey, err := rlpx.doEncHandshake(prv0, pubb1)
		if err != nil {
			t.Errorf("dial side enc handshake failed: %v", err)
			return
		}
		if !reflect.DeepEqual(rpubkey, pubb1) {
			t.Errorf("dial side remote pubkey mismatch: got %v, want %v", rpubkey, &pubb1)
			return
		}

		phs, err := rlpx.doProtoHandshake(hs0)
		if err != nil {
			t.Errorf("dial side proto handshake error: %v", err)
			return
		}
		phs.Rest = nil
		if !reflect.DeepEqual(phs, hs1) {
			t.Errorf("dial side proto handshake mismatch:\ngot: %s\nwant: %s\n", spew.Sdump(phs), spew.Sdump(hs1))
			return
		}
		rlpx.close(DiscQuitting)
	}()
	go func() {
		defer wg.Done()
		defer fd1.Close()
		rlpx := newRLPX(fd1)
		rpubkey, err := rlpx.doEncHandshake(prv1, ed448.PublicKey{})
		if err != nil {
			t.Errorf("listen side enc handshake failed: %v", err)
			return
		}
		if !reflect.DeepEqual(rpubkey, pubb0) {
			t.Errorf("listen side remote pubkey mismatch: got %v, want %v", rpubkey, &pubb0)
			return
		}

		phs, err := rlpx.doProtoHandshake(hs1)
		if err != nil {
			t.Errorf("listen side proto handshake error: %v", err)
			return
		}
		phs.Rest = nil
		if !reflect.DeepEqual(phs, hs0) {
			t.Errorf("listen side proto handshake mismatch:\ngot: %s\nwant: %s\n", spew.Sdump(phs), spew.Sdump(hs0))
			return
		}

		if err := ExpectMsg(rlpx, discMsg, []DiscReason{DiscQuitting}); err != nil {
			t.Errorf("error receiving disconnect: %v", err)
		}
	}()
	wg.Wait()
}

func TestProtocolHandshakeErrors(t *testing.T) {
	tests := []struct {
		code uint64
		msg  interface{}
		err  error
	}{
		{
			code: discMsg,
			msg:  []DiscReason{DiscQuitting},
			err:  DiscQuitting,
		},
		{
			code: 0x989898,
			msg:  []byte{1},
			err:  errors.New("expected handshake, got 989898"),
		},
		{
			code: handshakeMsg,
			msg:  make([]byte, baseProtocolMaxMsgSize+2),
			err:  errors.New("message too big"),
		},
		{
			code: handshakeMsg,
			msg:  []byte{1, 2, 3},
			err:  newPeerError(errInvalidMsg, "(code 0) (size 4) rlp: expected input list for p2p.protoHandshake"),
		},
		{
			code: handshakeMsg,
			msg:  &protoHandshake{Version: 3},
			err:  DiscInvalidIdentity,
		},
	}

	for i, test := range tests {
		p1, p2 := MsgPipe()
		go Send(p1, test.code, test.msg)
		_, err := readProtocolHandshake(p2)
		if !reflect.DeepEqual(err, test.err) {
			t.Errorf("test %d: error mismatch: got %q, want %q", i, err, test.err)
		}
	}
}

func TestRLPXFrameFake(t *testing.T) {
	buf := new(bytes.Buffer)
	hash := fakeHash([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
	rw := newRLPXFrameRW(buf, secrets{
		AES:        crypto.SHA3(),
		MAC:        crypto.SHA3(),
		IngressMAC: hash,
		EgressMAC:  hash,
	})

	golden := unhex(`
ee4bdf598434add487b2845159762028
01010101010101010101010101010101
2ad9100d3c2ac6652fa799328a635bc2
01010101010101010101010101010101
`)

	// Check WriteMsg. This puts a message into the buffer.
	if err := Send(rw, 8, []uint{1, 2, 3, 4}); err != nil {
		t.Fatalf("WriteMsg error: %v", err)
	}
	written := buf.Bytes()
	if !bytes.Equal(written, golden) {
		t.Fatalf("output mismatch:\n  got:  %x\n  want: %x", written, golden)
	}

	// Check ReadMsg. It reads the message encoded by WriteMsg, which
	// is equivalent to the golden message above.
	msg, err := rw.ReadMsg()
	if err != nil {
		t.Fatalf("ReadMsg error: %v", err)
	}
	if msg.Size != 5 {
		t.Errorf("msg size mismatch: got %d, want %d", msg.Size, 5)
	}
	if msg.Code != 8 {
		t.Errorf("msg code mismatch: got %d, want %d", msg.Code, 8)
	}
	payload, _ := ioutil.ReadAll(msg.Payload)
	wantPayload := unhex("C401020304")
	if !bytes.Equal(payload, wantPayload) {
		t.Errorf("msg payload mismatch:\ngot  %x\nwant %x", payload, wantPayload)
	}
}

type fakeHash []byte

func (fakeHash) Write(p []byte) (int, error) { return len(p), nil }
func (fakeHash) Reset()                      {}
func (fakeHash) BlockSize() int              { return 0 }

func (h fakeHash) Size() int           { return len(h) }
func (h fakeHash) Sum(b []byte) []byte { return append(b, h...) }

func TestRLPXFrameRW(t *testing.T) {
	var (
		aesSecret      = make([]byte, 16)
		macSecret      = make([]byte, 16)
		egressMACinit  = make([]byte, 32)
		ingressMACinit = make([]byte, 32)
	)
	for _, s := range [][]byte{aesSecret, macSecret, egressMACinit, ingressMACinit} {
		rand.Read(s)
	}
	conn := new(bytes.Buffer)

	s1 := secrets{
		AES:        aesSecret,
		MAC:        macSecret,
		EgressMAC:  sha3.New256(),
		IngressMAC: sha3.New256(),
	}
	s1.EgressMAC.Write(egressMACinit)
	s1.IngressMAC.Write(ingressMACinit)
	rw1 := newRLPXFrameRW(conn, s1)

	s2 := secrets{
		AES:        aesSecret,
		MAC:        macSecret,
		EgressMAC:  sha3.New256(),
		IngressMAC: sha3.New256(),
	}
	s2.EgressMAC.Write(ingressMACinit)
	s2.IngressMAC.Write(egressMACinit)
	rw2 := newRLPXFrameRW(conn, s2)

	// send some messages
	for i := 0; i < 10; i++ {
		// write message into conn buffer
		wmsg := []interface{}{"foo", "bar", strings.Repeat("test", i)}
		err := Send(rw1, uint64(i), wmsg)
		if err != nil {
			t.Fatalf("WriteMsg error (i=%d): %v", i, err)
		}

		// read message that rw1 just wrote
		msg, err := rw2.ReadMsg()
		if err != nil {
			t.Fatalf("ReadMsg error (i=%d): %v", i, err)
		}
		if msg.Code != uint64(i) {
			t.Fatalf("msg code mismatch: got %d, want %d", msg.Code, i)
		}
		payload, _ := ioutil.ReadAll(msg.Payload)
		wantPayload, _ := rlp.EncodeToBytes(wmsg)
		if !bytes.Equal(payload, wantPayload) {
			t.Fatalf("msg payload mismatch:\ngot  %x\nwant %x", payload, wantPayload)
		}
	}
}

type handshakeAuthTest struct {
	input       string
	isPlain     bool
	wantVersion uint
	wantRest    []rlp.RawValue
}

var cip8HandshakeAuthTests = []handshakeAuthTest{
	// (Auth₁) RLPx v4 plain encoding
	{
		input:       `024cea4fe3ca154bf3dabef6c312e1d3ed34a01a45847d4a013108b396f373382c8486320a32f7d96a7f8ac7d354a32c1bb9e6d4fb0ac7272ce9806df1c08cffce1e329e7d9206fb4ce1f27d9943ff4d2ed7a3bde5847b10167e3b0027c2c48a1bc33db256ee572f05ea140c5c085e206b1362051e129a8a550b6878b28c1d8d981aaa37d07d5b8ac17eb887775fb2169889f9105ac6cc39860632af2fc9a27fde3201168855527ab189df8f0fbcc4c07354e3fd4dbfa6b776bdcebbd02ed2bb454fe6fa156e98a1440f0dc6c00c33683cde84ddf64c8c6a241f100d4813537c28227d42a315a3a4739910e3f9304190a591e6f826588b6f4cb496bc7625c2116555a8eb636c28397b29752e36d7e5d1240c167cd91e6066731cfe562325ad27760fa764340a0fcc18af361ce63251a890c20b7cc06730c4db82275712b4352c49842f772c35d8088efeae51e6f9a3baaaa225349286211b3fd2125068d483073519b91233dc1612cc58c9ed59f95166fddde5a05d4e5a2a119c40391ee1e8d2afcfd18158f4a47af90fa9e8c2db0b138022f3c621804bdd36b429f720df93ed16fd9c9932bd5d2548b4a48d6bf29ca0f641c6eb0b0044600aa75a4b0ad501a59d952b1a904f0c0d0468fdf58e5e9e56b91e56793d58e5150e035d311dd6c5fee6ea6a1d01660adb72e97a10d14a9a463bf583a39d19c0a40ad3b57f7e777391b7714488687b1c10a11958300e64afae8a9762f8c6f85c536156ad642dcfe78b61bc17eaca64d40890e679a172b14a7a7d8e11d55da029c8a8da33abbd11c307d05deb2486d36d08d08348b8bd83`,
		isPlain:     false,
		wantRest:    []rlp.RawValue{},
		wantVersion: 4,
	},
	// (Auth₂) CIP-8 encoding
	{
		input:       `024ccad299392df90734104b9ba7991f77db1052fa8b3545931b23e604d5b0da47c28e3fee12d3996cc86db94c742a04ef8b0829aea28823494f004f76a498f3ad51bd33b2b39186265910b63dd553f423a0b1efcb83395c4e7c7f0bfad98ba436e7da129a7edd05bb164066dd6d26502cf13d0ebb6c6219960648d4ecc9d2d6750f527a7e6248a1f33e92f99013a02f12cd87092af3875c1a9d9180d120db62686a5e6ca4d59438ba08ab8ba78de725f980156fb787c2aff1071eb85b2d5f17acde2cb145aef636ca4aabd55addc287baf8f380f18a2a14a17fcdabc8999182431097dd036b7ae2a5fda25003affcd7c5794cdf332c3b42390df741823e4e009afa57e04d7054313be6bec8ee2e5dea88fa5e9e0c272598e82d268cd90e4525faa817f4104d56b4f8b0b0f77d22bbb23ad30cd5992980ab1df5c2a0db3fb43f0a6019eeba9f8d4c37539968066b449817451a7d0f7da7a30e6dbad7c9ef4eb3faa02a8ad3c7da0c0e9f50fe997e640a492ee2b4a9999e1df7b6b36d67697ff19de5d63b074acf51bc624c9470237381beafd0ae51d0fc2bed5d09f86f2eaf65a4fbb8b9f7182f5c29f6db7ba2220b5d21bd21ae00cfd4e2296bfed3fc138c276139c0977b0054e8b01dfd8984dce11dd344e55f67b9f5d328eeea0e5832ee3c66491b3b52525c989165c3f2c3bd4308bb7cf86895b3b7f420135886c078fb013455b2007e5c1e4e523d14d0c880eb3f34da3818924a16feb3aaced85cea20900d2a89d107c1572cc5f8098838190a1681f11f8acc88e06a0aae35205756d4b08db01de3f6228cf8a3aae8e4ff21`,
		wantVersion: 4,
		wantRest:    []rlp.RawValue{},
	},
	// (Auth₃) RLPx v4 CIP-8 encoding with version 56, additional list elements
	{
		input:       `025138cbe53a20acc28881ca94d94f875d7457c69f02d3ae004e60b4e08fca7d4e5f5ba4b2d16f563a430c4360bf92681163ad692281eaba19c780fc6fd2ca3a826799eed17bb560433a3b78d2cb74f829d214a3f1cce087f16d4dc9ed5522db59bfbc80d640b490e5b537ad04660f80cf0cc89bf3a08a56df2ec532cfc832175320619038e14b622f4ca0359dee9fecaa5e1597914655feb2630721c766e146842fb5932d2e290ba19dad372751299526babfc4b33f6a024517833a8dd1d6206da34a8bdc0dd9d590e627957d6687c4749c068fa59f9fee57bef1a1ab8f8b8ee2bce13e04971a880de143049c798ea433db213f0ddcfd2548c7677e790dc47649eec27aa5bf64d2459e616cba3485e557e6ddef6ceb686edcb3d05de624625777ee516486c9ee68aaaca100e3d45a9a18dffe4d6a24b782fb50483af9424785e74da583470c76fe89822a5f15a2dc94fc108d30e7e426ade84746044a9a9000554c2b436c75afea49a3a04bff423b5cbe6ec15ce38c0a5328cc93b8cdebf4dc8130fb778a4d42de1c9c7b3d37ad6433d4faf36b48e110087295d4ece5143fecc0784affb4f1530a75af4ae757f3658f88316781ca40b081271e08a95f9a879980edb6e5ed29e0b4829f38a27c75408df324041612645e8509cf611013ae5f3b7ca807dbfe6666617962861feceb316bb169393bcf6e49ba9e2691d680218a87f9b87b8dad962f236fb2fd7085aab681c59621c0b3422fdbe4bbd0f840bfa6a87f12e9e140d84615e35673aa9a9e00c0a4d69e1eef4d1de4092572966d5f6d245be6080d984d8e9b44d55391a84280a04bc5ba`,
		wantVersion: 56,
		wantRest:    []rlp.RawValue{{0x01}, {0x02}, {0xC2, 0x04, 0x05}},
	},
}

type handshakeAckTest struct {
	input       string
	wantVersion uint
	wantRest    []rlp.RawValue
}

var cip8HandshakeRespTests = []handshakeAckTest{
	// (Ack₁) RLPx v4 plain encoding
	{
		input:       `1f25a25b85bf5bbaf2e18663e3776f41a72ca14b6f9ef65658f9da30cf0078b403269c740b918fb0a307dbbba6aedf1f80e612ac132661d200bc07a2a2e11197a2b6cac13ffeb12cf932372d9d98f4e22a5a940b00d09d726b04701e9b301789a1a7a431267cb388a0c8b7a034001c2da59df6b8061800772ecdb5c892d30f91d6095acfe772d0437507724b01fd869bcbc30fa8f7a9016522eb525b629df2aae7d1e3da2f3271d572db57240d76b383cc9886de28914757912b12799e3d42c608a7fb3106e83bbbdcbe66811aeb8d3eebd19fa685612f4c4f918aedcbf374f55312f2`,
		wantVersion: 4,
	},
	// (Ack₂) CIP-8 encoding
	{
		input:       `019efa17c5f8155a29be83d87450b2270125a8cd82afd2ba73439ff66886074053d79fea273534d1ddd327d8117bcb28f7c0e79cbebbfdbcc443002a7bd79ad4308b809797987c4775859819ffa2991a2301eedf87fb5e3bcdcefed9617a27f3bfa7de21bb41a21556dbea9b37e5b6619ca4bc2088125cbadf6543ab0daa82d42c38325de1f11be9515c809608274cc2a605687677aaa52654a37acd62e958862a90080ead6b483cde8ffa6cc25f6332b2e8ef8b2e4b79c2f83ca58e95f91cf184ff737af9a70ea17fe1b838ca3ef7d0f0e9d9a8c2f1952fc3a92177cbda75bcf5e5fd79d17645a03ab57420497bc574e64e7b1119b5ab624ebd8977140235b6f57596347612e3f2d22d0490a1b10fb8b064fcb55b6642aad0c5c850b50b410f2549ea0b0b8927506de68bcab802af69aeaf495fdce8c4ac39bae87ad00550f76606a1e6c21885a58a8dc23b206366be4188ca3292c804fd457123157b794ddd9decaa6fb4bf98e58176d2a8436f363b9fc9ff1fd16c057e764e5f89d177b14d3267523911b88f14c749698c691d67fd2508c45057bad86e644ade07c31321b6`,
		wantVersion: 4,
		wantRest:    []rlp.RawValue{},
	},
	// (Ack₃) CIP-8 encoding with version 57, additional list elements
	{
		input:       `01a4b7be1a43972fa4c7fe12585551f38a50d382e12b8d9184582847adf6e3ebfc424c21021a591aa9c591dc78c5b51ecac3916d68e386eda370803abef5c9372d7bb6212df673a74b22b312d699e62686efb2a4ad8d4bb0c11464edaac29c2e3543f1becbaa90a1f20ec863bce245e463e3977a2a08cfadaf40bc6e9220f0bfbf89ec7a2b135dd03dfb456a864bee985b5321b1aa18f50d9b5f2c5b0c473ffb071ddeb92682bfa9716912a7627e2cc216d9ff0d51132c28126a57c20dbf67d215368489f3f6a65545cecd814f0fbead04b3ca6119b3ebc6e1f4b4d8c5f6b12f9905165c85373c676d11bb353c946daa31070fa5acab924966511d45e3286d53cec90d34818ea12ab9a561283375b1f65ae6b70f3d1eed78070cbac05468cd1ad9209733fff27b32269907dd8dd950539f85b76e0ea3506afa5fbc3a2b3f0eb3cce2babb47911664a78d14b230e40b92a0367d98f62b7a19d61330836d69a628fb66eec9c132693b1ff3aad4547ab021895410048dd701253f3da5cfa3b1f8af0825a2fc10dd6addfce3872abb2b88d0a5d8f135968364fffb52fd1e384af997a11fd550f242`,
		wantVersion: 57,
		wantRest:    []rlp.RawValue{{0x06}, {0xC2, 0x07, 0x08}, {0x81, 0xFA}},
	},
}

func TestHandshakeForwardCompatibility(t *testing.T) {
	var (
		keyA, _       = crypto.HexToEDDSA("f2cdd78003bf733b7badb5001ea1fe9248346f5b7943a87b3c40528a8f10941eb29313c1abf703aaab4a46a94a0b2302ec89bf1998332d7c60")
		keyB, _       = crypto.HexToEDDSA("856a9af6b0b651dd2f43b5e12193652ec1701c4da6f1c0d2a366ac4b9dabc9433ef09e41ca129552bd2c029086d9b03604de872a3b3432041f")
		pubbA         = ed448.Ed448DerivePublicKey(keyA)
		pubbB         = ed448.Ed448DerivePublicKey(keyB)
		pubA          = crypto.FromEDDSAPub(pubbA)
		pubB          = crypto.FromEDDSAPub(pubbB)
		ephA, _       = crypto.HexToEDDSA("45516f2d6e60098e547e9b50d386e75f530805fb468c132bead2ce7b205208d895cb086fff390eff73c349a7e5caf1c8c8d8278ae31a6b175a")
		ephB, _       = crypto.HexToEDDSA("96b3c4485ef83aae585776685bed5d7d6373befb7b661f43592ac703b94ed543526a23d4de35af35c30690998993f140ed1fd9389bc99506b9")
		pubbEpthA     = ed448.Ed448DerivePublicKey(ephA)
		pubbEpthB     = ed448.Ed448DerivePublicKey(ephB)
		ephPubA       = crypto.FromEDDSAPub(pubbEpthA)
		ephPubB       = crypto.FromEDDSAPub(pubbEpthB)
		nonceA        = unhex("7e968bba13b6c50e2c4cd7f241cc0d64d1ac25c7f5952df231ac6a2bda8ee5d6")
		nonceB        = unhex("559aead08264d5795d3909718cdd05abd49572e84fe55590eef31a88a08fdffd")
		_, _, _, _    = pubA, pubB, ephPubA, ephPubB
		authSignature = unhex("c180a432dad22041c7c648776681b4b20cb040c58ffe8d58a344cb86e704eb26e68300b8c4f3f0836c7025eeb81f76b4952fa0a0f70effdb008ec4fd100e0e7b7bc30b5a5ac8aea7eb0d2092bbdd5a3763e78e7c1c6d13cc0856a6171640af4e7d1f013a08c4d474ecf8223d1c871a412700f6e2ad06289ee32fb5b673d332629f5aa30e553783812eaf76dc929e9785765e001e93e1fe5ac2c1ebbaca3a15fc814a4321959afd6d97f200")
		_             = authSignature
	)

	makeAuth := func(test handshakeAuthTest) *authMsgV4 {
		msg := &authMsgV4{Version: test.wantVersion, Rest: test.wantRest, gotPlain: test.isPlain}
		copy(msg.Signature[:], authSignature)
		copy(msg.InitiatorPubkey[:], pubA)
		copy(msg.Nonce[:], nonceA)
		return msg
	}
	makeAck := func(test handshakeAckTest) *authRespV4 {
		msg := &authRespV4{Version: test.wantVersion, Rest: test.wantRest}
		copy(msg.RandomPubkey[:], ephPubB)
		copy(msg.Nonce[:], nonceB)
		return msg
	}

	// check auth msg parsing
	for _, test := range cip8HandshakeAuthTests {
		r := bytes.NewReader(unhex(test.input))
		msg := new(authMsgV4)
		ciphertext, err := readHandshakeMsg(msg, encAuthMsgLen, keyB, r)
		if err != nil {
			t.Errorf("error for input %x:\n  %v", unhex(test.input), err)
			continue
		}
		if !bytes.Equal(ciphertext, unhex(test.input)) {
			t.Errorf("wrong ciphertext for input %x:\n  %x", unhex(test.input), ciphertext)
		}
		want := makeAuth(test)
		if !reflect.DeepEqual(msg, want) {
			t.Errorf("wrong msg for input %x:\ngot %s\nwant %s", unhex(test.input), spew.Sdump(msg), spew.Sdump(want))
		}
	}

	// check auth resp parsing
	for _, test := range cip8HandshakeRespTests {
		input := unhex(test.input)
		r := bytes.NewReader(input)
		msg := new(authRespV4)
		ciphertext, err := readHandshakeMsg(msg, encAuthRespLen, keyA, r)
		if err != nil {
			t.Errorf("error for input %x:\n  %v", input, err)
			continue
		}
		if !bytes.Equal(ciphertext, input) {
			t.Errorf("wrong ciphertext for input %x:\n  %x", input, err)
		}
		want := makeAck(test)
		if !reflect.DeepEqual(msg, want) {
			t.Errorf("wrong msg for input %x:\ngot %s\nwant %s", input, spew.Sdump(msg), spew.Sdump(want))
		}
	}

	// check derivation for (Auth₂, Ack₂) on recipient side
	var (
		hs = &encHandshake{
			initiator:     false,
			respNonce:     nonceB,
			randomPrivKey: ephB,
		}
		authCiphertext     = unhex(cip8HandshakeAuthTests[1].input)
		authRespCiphertext = unhex(cip8HandshakeRespTests[1].input)
		authMsg            = makeAuth(cip8HandshakeAuthTests[1])
		wantAES            = unhex("f66f0ebac0eb93bb131bbc91e8cfc996c4e923617b5d9a2aac25e610d7408279")
		wantMAC            = unhex("7e6e60c3bb303290a274b07f82269c8af55a5ae89891900ac42ad42064083866")
		wantFooIngressHash = unhex("38ae8889d5072cbaf71481c19f6ef5a3cafe4b48998c72d5ab8f24c8a754c87f")
	)
	if err := hs.handleAuthMsg(authMsg, keyB); err != nil {
		t.Fatalf("handleAuthMsg: %v", err)
	}
	derived, err := hs.secrets(authCiphertext, authRespCiphertext)
	if err != nil {
		t.Fatalf("secrets: %v", err)
	}
	if !bytes.Equal(derived.AES, wantAES) {
		t.Errorf("aes-secret mismatch:\ngot %x\nwant %x", derived.AES, wantAES)
	}
	if !bytes.Equal(derived.MAC, wantMAC) {
		t.Errorf("mac-secret mismatch:\ngot %x\nwant %x", derived.MAC, wantMAC)
	}
	io.WriteString(derived.IngressMAC, "foo")
	fooIngressHash := derived.IngressMAC.Sum(nil)
	if !bytes.Equal(fooIngressHash, wantFooIngressHash) {
		t.Errorf("ingress-mac('foo') mismatch:\ngot %x\nwant %x", fooIngressHash, wantFooIngressHash)
	}
}
