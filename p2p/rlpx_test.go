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
	eddsa "github.com/core-coin/ed448"
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
	pub0 := eddsa.Ed448DerivePublicKey(*prv0)
	prv1, _ := crypto.GenerateKey(rand.Reader)
	pub1 := eddsa.Ed448DerivePublicKey(*prv1)

	ss0 := crypto.ComputeSecret(prv0, &pub1)
	ss1 := crypto.ComputeSecret(prv1, &pub0)
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
		pubkey *eddsa.PublicKey
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

		pub := eddsa.Ed448DerivePublicKey(*prv1)
		r.pubkey, r.err = c0.doEncHandshake(prv0, &pub)
		if r.err != nil {
			return
		}
		if !reflect.DeepEqual(r.pubkey, &pub) {
			r.err = fmt.Errorf("remote pubkey mismatch: got %v, want: %v", r.pubkey, &pub)
		}
	}()
	go func() {
		r := result{side: "receiver"}
		defer func() { output <- r }()
		defer fd1.Close()

		r.pubkey, r.err = c1.doEncHandshake(prv1, nil)
		if r.err != nil {
			return
		}
		pub := eddsa.Ed448DerivePublicKey(*prv0)
		if !reflect.DeepEqual(r.pubkey, &pub) {
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
		pubb0   = eddsa.Ed448DerivePublicKey(*prv0)
		pub0    = crypto.FromEDDSAPub(&pubb0)
		hs0     = &protoHandshake{Version: 3, ID: pub0, Caps: []Cap{{"a", 0}, {"b", 2}}}

		prv1, _ = crypto.GenerateKey(rand.Reader)
		pubb1   = eddsa.Ed448DerivePublicKey(*prv1)
		pub1    = crypto.FromEDDSAPub(&pubb1)
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
		rpubkey, err := rlpx.doEncHandshake(prv0, &pubb1)
		if err != nil {
			t.Errorf("dial side enc handshake failed: %v", err)
			return
		}
		if !reflect.DeepEqual(rpubkey, &pubb1) {
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
		rpubkey, err := rlpx.doEncHandshake(prv1, nil)
		if err != nil {
			t.Errorf("listen side enc handshake failed: %v", err)
			return
		}
		if !reflect.DeepEqual(rpubkey, &pubb0) {
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
		input:       `024c996a6dc06dc9b270be2a756422a9c8b0cd5c8d1fff40f6c5c7fbddb1ad399c8809137f6083a77143bde0d555d28a9deb062343d3a84242a500d6d6cc8db82eef9fb170cd450944c5ea5568b820c890429401cad1772d982d92c55182550c2341eaf4e960aae5a87127efb7545b65450164001ee32a2cbb01404512d174cafc6cf808eab83e875cc6ef668163dd48a3865e1c240db7a11d90a9e8d8b9533a3169946f14d19bf15d714b29b020eafca689add206bf94afa27231bfa1945c6a70357a4dd31af09adab31ffd6c18bdcd16b45e92d24a0d6501b569c4ae0e86a3bebd015c4c22373ce86dfc9da9907b35d8f978e8d47fb005b20b13531edbcb0ff1c776aee6042f55e6a42f12089e6a41f82f1ad18bbd6c1c7d601ef81dc8363d5da7689c11f6e6e19da9b759b72215cb7387cedacb9380c148e06f873746b96fc6048c09ebbb736e53bb0ece94444bc9f2fc44ddbd520a1f665804a26c9f03d1e365e7fc1e235854e4dd83d9c439a81411cec80b1c98f5850cc92a8aa40bce3f6e71836358d9dbfdb9300b411086efcacff2f8bbb5548bfe5042411b278f5c42d8ceae3821ec93d92a7e76508e638eacbca223728b9569d8ddd4e167edba287665cdc21cfbbb2e8e77730a4d577ec14a4274d1b627b832595c9b6f2f4993dbd6f2b11b1e4cc3b3a1d93a8a6bf8580e85c4e570cb32a8c512e20abb00a73cbaf0143b561fff11f2d8c557badb391b2031f7d29d2c75fa9670fc890a8d9f6a4cc095c97e82823eb3efd707b94b3207d2149bdb2dd5e5d84754548194923ec13839d43cafc28bb111734ba4854bb0a2`,
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
		input:       `eca7a771a547ddbc4075532b85d9f83a6ec927b0572a3c222a39fba8184d8729a6b14b14c26f23b4c509d174b5477a73a8cb8bf95a814da780987def344a0dee794c78073873228500d123f8472ea008fd89bf184d6375cbc08116b0cf5d46b1c1c125edf00ef1ce2a3f5535667a0813cac568e32e5ae7f1c68df69dba2a1a24a77f6c6b148e3fd04f7753cb9343d5abfe9a7be00b4fdc92bd8ec999b82b6da33ccebc3f8fbb3960d4272b7c9f1ef6bf93906b04306e05a6e28b6dc9c19f7c8d26ca4f1de3b7edd524fc8dcb9fafe4d4da34fcb6dad601af263af7d6b0132387f35f11`,
		wantVersion: 4,
	},
	// (Ack₂) CIP-8 encoding
	{
		input:       `01a4dc5bb0d46b3ef7e35124fe30913961ff4b962ac49328a1f17c734ba40a8c0d2b4007a81ab3257da60bb7c5ef1c2b883dd58ea1b51f6e6a318037dac7d00c6b57a658e804a033e3a61921160c00ec11fb2719fd11b2ea8882eb4d1fe15b6ef5c4747e9af92c27d18f99bddc824d576d8650257a421b5ec059aeb7feec0174209c8f057d95a6991ae36ca72547de2e089eee5319b342cc4f8b12cb7628806236af54295d7d21c9e00aa81f063f5dc852b7721f942bf3b12313716297824cb1e635f9934793d92520970fc2071349e3163b2364ff7fd9d2079c2a5bb39d5b312546c2de60815e1152760122e2ae43cff2b0774a9879471e83872f1d86fa5fbab0d27e59722717ddee8881834671d1ae6a5eb09019df64bb5297bb4dc476d860a9f4ce813382c47d434380616de332e25d03a7f993d0576c76b63b95f0157fb2e915c7c0ce189ab854b026afbcd80699acf862dd589c628705c7c310fd1c551d22aac8ffbe3c332c3c46f6df4f887f552c10116dabeb83cca00ced3bb517cf9c5c425e5c9aa77b6f29876632398dff93123485fe3778ee63a85bc41e72eab9e4980ee25b6a69`,
		wantVersion: 4,
		wantRest:    []rlp.RawValue{},
	},
	// (Ack₃) CIP-8 encoding with version 57, additional list elements
	{
		input:       `018298af236a6774f8f6497412c6529aa4d867612c22bd96942bc502319a74d0ea6d07a5802b5997a5abd917153c0e860ceb703a869c9716f85100db6100d8ae296f52abb054e8f9caa5c060f656576e5d9effe4ab40fb49e156c4765ace4d047cfc8b4432fe281d5508f536943fb769b6fa060b6faaba7ac13d8b84317b8dabd4d53883393a539314df3671ae076cd5da05728313cfbd8956c6c973a46fed1b2fd0520b3658b06637d061d84377ff2cebd387e5ef965c76c4810283e496a53c954d2af9b955a73f31fd4bc63ab373d22bf5f0b14955312d7ffb7ad335c90bd8ac563e8d7576ee5ceefbff16f2e4a9c5d4de3a2192985b7aa2522a5bc92b221cb2f036ba427039e83cbba7acfcc60596bd06afc875cf355fe2f8a2d8a35539684341635a348bec7a4445a38adc9999ebaf195fb325dc131665f77fe1b7d9e1f037e3e82cebf90aaf92f7042c6802dded28657a4e5281adedf98d62a14de5b88c518d5337bd1d483d459e8d49c1ff3ec211cd2f09717b0acbadebe7a0266e3c466c356ee0`,
		wantVersion: 57,
		wantRest:    []rlp.RawValue{{0x06}, {0xC2, 0x07, 0x08}, {0x81, 0xFA}},
	},
}

func TestHandshakeForwardCompatibility(t *testing.T) {
	var (
		keyA, _       = crypto.HexToEDDSA("f2cdd78003bf733b7badb5001ea1fe9248346f5b7943a87b3c40528a8f10941eb29313c1abf703aaab4a46a94a0b2302ec89bf1998332d7c60")
		keyB, _       = crypto.HexToEDDSA("856a9af6b0b651dd2f43b5e12193652ec1701c4da6f1c0d2a366ac4b9dabc9433ef09e41ca129552bd2c029086d9b03604de872a3b3432041f")
		pubbA         = eddsa.Ed448DerivePublicKey(*keyA)
		pubbB         = eddsa.Ed448DerivePublicKey(*keyB)
		pubA          = crypto.FromEDDSAPub(&pubbA)
		pubB          = crypto.FromEDDSAPub(&pubbB)
		ephA, _       = crypto.HexToEDDSA("45516f2d6e60098e547e9b50d386e75f530805fb468c132bead2ce7b205208d895cb086fff390eff73c349a7e5caf1c8c8d8278ae31a6b175a")
		ephB, _       = crypto.HexToEDDSA("96b3c4485ef83aae585776685bed5d7d6373befb7b661f43592ac703b94ed543526a23d4de35af35c30690998993f140ed1fd9389bc99506b9")
		pubbEpthA     = eddsa.Ed448DerivePublicKey(*ephA)
		pubbEpthB     = eddsa.Ed448DerivePublicKey(*ephB)
		ephPubA       = crypto.FromEDDSAPub(&pubbEpthA)
		ephPubB       = crypto.FromEDDSAPub(&pubbEpthB)
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
		wantAES            = unhex("02a8f660fd5c452851980694e076d64b2168605201638e95c47d3c696000b1d7")
		wantMAC            = unhex("96a446be84d374707892788f7de25dcf9b175f4bc9dd7a6c119af31cc2b30f72")
		wantFooIngressHash = unhex("82dffd828360fc5c2c049f28d1860bab35d64b590d225f27e2a9af5d76da7451")
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
