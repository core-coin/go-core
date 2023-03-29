// Copyright 2020 by the Authors
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

package rlpx

import (
	"bytes"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/rlp"
)

type message struct {
	code uint64
	data []byte
	err  error
}

func TestHandshake(t *testing.T) {
	p1, p2 := createPeers(t)
	p1.Close()
	p2.Close()
}

// This test checks that messages can be sent and received through WriteMsg/ReadMsg.
func TestReadWriteMsg(t *testing.T) {
	peer1, peer2 := createPeers(t)
	defer peer1.Close()
	defer peer2.Close()

	testCode := uint64(23)
	testData := []byte("test")
	checkMsgReadWrite(t, peer1, peer2, testCode, testData)

	t.Log("enabling snappy")
	peer1.SetSnappy(true)
	peer2.SetSnappy(true)
	checkMsgReadWrite(t, peer1, peer2, testCode, testData)
}

func checkMsgReadWrite(t *testing.T, p1, p2 *Conn, msgCode uint64, msgData []byte) {
	// Set up the reader.
	ch := make(chan message, 1)
	go func() {
		var msg message
		msg.code, msg.data, _, msg.err = p1.Read()
		ch <- msg
	}()

	// Write the message.
	_, err := p2.Write(msgCode, msgData)
	if err != nil {
		t.Fatal(err)
	}

	// Check it was received correctly.
	msg := <-ch
	assert.Equal(t, msgCode, msg.code, "wrong message code returned from ReadMsg")
	assert.Equal(t, msgData, msg.data, "wrong message data returned from ReadMsg")
}

func createPeers(t *testing.T) (peer1, peer2 *Conn) {
	conn1, conn2 := net.Pipe()
	key1, key2 := newkey(), newkey()
	peer1 = NewConn(conn1, key2.PublicKey()) // dialer
	peer2 = NewConn(conn2, nil)              // listener
	doHandshake(t, peer1, peer2, key1, key2)
	return peer1, peer2
}

func doHandshake(t *testing.T, peer1, peer2 *Conn, key1, key2 *crypto.PrivateKey) {
	keyChan := make(chan *crypto.PublicKey, 1)
	go func() {
		pubKey, err := peer2.Handshake(key2)
		if err != nil {
			t.Errorf("peer2 could not do handshake: %v", err)
		}
		keyChan <- pubKey
	}()

	pubKey2, err := peer1.Handshake(key1)
	if err != nil {
		t.Errorf("peer1 could not do handshake: %v", err)
	}
	pubKey1 := <-keyChan

	// Confirm the handshake was successful.
	if !reflect.DeepEqual(pubKey1, key1.PublicKey()) || !reflect.DeepEqual(pubKey2, key2.PublicKey()) {
		t.Fatal("unsuccessful handshake")
	}
}

// This test checks the frame data of written messages.
func TestFrameReadWrite(t *testing.T) {
	conn := NewConn(nil, nil)
	hash := fakeHash([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
	conn.InitWithSecrets(Secrets{
		AES:        crypto.SHA3(),
		MAC:        crypto.SHA3(),
		IngressMAC: hash,
		EgressMAC:  hash,
	})
	h := conn.handshake

	golden := unhex(`
		ee4bdf598434add487b2845159762028
		01010101010101010101010101010101
		2ad9100d3c2ac6652fa799328a635bc2
		01010101010101010101010101010101
	`)
	msgCode := uint64(8)
	msg := []uint{1, 2, 3, 4}
	msgEnc, _ := rlp.EncodeToBytes(msg)

	// Check writeFrame. The frame that's written should be equal to the test vector.
	buf := new(bytes.Buffer)
	if err := h.writeFrame(buf, msgCode, msgEnc); err != nil {
		t.Fatalf("WriteMsg error: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), golden) {
		t.Fatalf("output mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), golden)
	}

	// Check readFrame on the test vector.
	content, err := h.readFrame(bytes.NewReader(golden))
	if err != nil {
		t.Fatalf("ReadMsg error: %v", err)
	}
	wantContent := unhex("08C401020304")
	if !bytes.Equal(content, wantContent) {
		t.Errorf("frame content mismatch:\ngot  %x\nwant %x", content, wantContent)
	}
}

type fakeHash []byte

func (fakeHash) Write(p []byte) (int, error) { return len(p), nil }
func (fakeHash) Reset()                      {}
func (fakeHash) BlockSize() int              { return 0 }
func (h fakeHash) Size() int                 { return len(h) }
func (h fakeHash) Sum(b []byte) []byte       { return append(b, h...) }

type handshakeAuthTest struct {
	input       string
	isPlain     bool
	wantVersion uint
	wantRest    []rlp.RawValue
}

var cip8HandshakeAuthTests = []handshakeAuthTest{
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
		keyA, _       = crypto.UnmarshalPrivateKeyHex("f2cdd78003bf733b7badb5001ea1fe9248346f5b7943a87b3c40528a8f10941eb29313c1abf703aaab4a46a94a0b2302ec89bf1998332d7c60")
		keyB, _       = crypto.UnmarshalPrivateKeyHex("856a9af6b0b651dd2f43b5e12193652ec1701c4da6f1c0d2a366ac4b9dabc9433ef09e41ca129552bd2c029086d9b03604de872a3b3432041f")
		ephA, _       = crypto.UnmarshalPrivateKeyHex("45516f2d6e60098e547e9b50d386e75f530805fb468c132bead2ce7b205208d895cb086fff390eff73c349a7e5caf1c8c8d8278ae31a6b175a")
		ephB, _       = crypto.UnmarshalPrivateKeyHex("96b3c4485ef83aae585776685bed5d7d6373befb7b661f43592ac703b94ed543526a23d4de35af35c30690998993f140ed1fd9389bc99506b9")
		nonceA        = unhex("7e968bba13b6c50e2c4cd7f241cc0d64d1ac25c7f5952df231ac6a2bda8ee5d6")
		nonceB        = unhex("559aead08264d5795d3909718cdd05abd49572e84fe55590eef31a88a08fdffd")
		_, _, _, _    = keyA, keyB, ephA, ephB
		authSignature = unhex("c180a432dad22041c7c648776681b4b20cb040c58ffe8d58a344cb86e704eb26e68300b8c4f3f0836c7025eeb81f76b4952fa0a0f70effdb008ec4fd100e0e7b7bc30b5a5ac8aea7eb0d2092bbdd5a3763e78e7c1c6d13cc0856a6171640af4e7d1f013a08c4d474ecf8223d1c871a412700f6e2ad06289ee32fb5b673d332629f5aa30e553783812eaf76dc929e9785765e001e93e1fe5ac2c1ebbaca3a15fc814a4321959afd6d97f200")
		_             = authSignature
	)
	makeAuth := func(test handshakeAuthTest) *authMsgV4 {
		msg := &authMsgV4{Version: test.wantVersion, Rest: test.wantRest}
		copy(msg.Signature[:], authSignature)
		copy(msg.InitiatorPubkey[:], keyA.PublicKey()[:])
		copy(msg.Nonce[:], nonceA)
		return msg
	}
	makeAck := func(test handshakeAckTest) *authRespV4 {
		msg := &authRespV4{Version: test.wantVersion, Rest: test.wantRest}
		copy(msg.RandomPubkey[:], ephB.PublicKey()[:])
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
		wantFooIngressHash = unhex("457aba35ae6b27710c07589c86fc50f88338139a8b67429c608cfc6d7fab7b87")
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

func unhex(str string) []byte {
	r := strings.NewReplacer("\t", "", " ", "", "\n", "")
	b, err := hex.DecodeString(r.Replace(str))
	if err != nil {
		panic(fmt.Sprintf("invalid hex string: %q", str))
	}
	return b
}

func newkey() *crypto.PrivateKey {
	key, err := crypto.GenerateKey(crand.Reader)
	if err != nil {
		panic("couldn't generate key: " + err.Error())
	}
	return key
}
