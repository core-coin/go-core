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
	eddsa "github.com/core-coin/go-goldilocks"
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
		input:       `024ccaa282473fe2718cfcf4d5d6ce65499d8809440081990379e0ca3aa12eca48009ad477f937e715195edc2b3226695a4ee5a30a9610fa82d500b5bb179807a200f161a5c1c514acc3b7044eeccf4a7c7381c56854e52e2a7ba8ccdd1378ff23898ce6736baba741213dec92df504f5d00c7793db58ad4138eccbb431412acc2edb4005d7842821d63dfecf649953a55af5d748283862d8acea675caab55f53410a2a2c3d6d667fc005fd38dff87a277f53214e6a864ee7ec4a84fdccbc29ee5ce0b68a804dd3c3c40c3cb28abe33aa593c195112bb853d4016997cba6c1b79dcb7e2396a558b07c79eeb1ed46649389825f071eda9a76c7ecd9e849e863774c51c118b0e23d0960a4b04e606d4bc140160dbd502e92cf96140ac6f4ccca6b7c9c5b89dc8e0375f4ab299e84f68ed87a7270567b80d66ee42cdaadeb914163b9b6b3ca59c3f6379332052a8a6d6d7b9a426bcfb7929b096155c2d1c69b839b0d266c4329b69ac8d1c82fdac64327f223ee539537df4ebaae7deb0aa6cb8b5542e336fe23ac5accf7637c17b9b112ccfd316657fc21aac1408db58aceaa9c21b43574ee0d19489268061006631d37a6e002998f0dd5dcffe8fe72b40e7f6aec61115570db8ab099793802fe6d43872dfb802b44a0413f9a93aa26661cc976c0952e0ea31d4727e1c96279f9d836e277f87f40c81f07eedd1311c1a5e31473d904c5cf61c7295341747b72e5d4e6e5794e999067c836ae8582574fbf80dcfbb97fdd91cbbe3ba835847f8174242f83a2e638246c0c9c1a2c85f2c1c29ba68d45ec2932030229df79a3f6e0d9825a`,
		isPlain:     false,
		wantRest:    []rlp.RawValue{},
		wantVersion: 4,
	},
	// (Auth₂) CIP-8 encoding
	{
		input:       `024c00e693f8ac1a8f9fa5954ca75e33bda54d95d16a64730bacd4c05a026a9898bf1d95f2615ded585fd3b764ec20dc2afa97e82fc4724a378600b2886c70206df3c9ec6bac0fc42ee5aa1a035d5b6c7e2252ea40f02d653c0a4efa7b26ca74bbd71f8110dc50718e3ffd8e722e79a7809b065888c466f5664d855ef647e0f4269af519269e706b16f5a0ea5eea7a6ceaca40b04600a30360e35d8dc24cc1227980f82e3cae6f3d9b428e67d4d519b5c7de6663771f69ef49a47eb42830028471cbb7944cfe7af011af9d4c6168de19dcffdf35edabb6c05f9ca336793b90f1ef1437f84abeaaaecd66522cf984469c9bf3adc15f1559a52d86a1e3bbd516cdc7cc5fa9fa94b7d52b823e40e5f8bd35d04a5d62c1bec939c6a33e8ad961cfcf18a6506f79569ff38a23fded109d8159dbcb149f4376f7f6bca9c765518b4a65cff5c3b54b1d2ea7ba03aa8302b62a972d2557b0528e4437f870fc5202a11de0caca781e964cdfe27f6bb2ea4af92397e133b04d3e43b4c28d672b31b75edc14edbcdb63dd0a76c7ba16c53981cc3ab3d22e9f852ff5ee42d2380a5b0e9992dcf0f76e35064db506de7551140084035baf01a8d3c48f923ee3d7428e99c9a364979c4fbae430b2916ce4be3cff09bdd793c3b7ea056251119d8fb3ae63c352fa3644e5f86a56e3534367625e5a9e638fc24e9ea4b22d14e617d1486d4862019b29464bb43afc10b8fbfd93c0f76e38d17037fe38f7137216ab4d1ca0899b7e9c365056f742d2ef150943676c4731aa0f7a7ce2877b070f7666147e85e6f0502d43ad35be1f067baa036d8ea81500`,
		wantVersion: 4,
		wantRest:    []rlp.RawValue{},
	},
	// (Auth₃) RLPx v4 CIP-8 encoding with version 56, additional list elements
	{
		input:       `02515e429fabaa78028bb306d98e1f15bffbdda86e19853e5ad2b907b865c5c7d9dc437d289ae6da99993e38e1945077df131813eab07a93f48c8086015630a81aaac26d69d31906d570070f36229a339750b4bb6fec927c0a80d1697e72426c0ac93b86eff07fcacd5f6b87141797450e70cafa60c9a52b91bcf2d1a249f46a315a3c7d77e0630bbd5d2b91311817cb18a51c50c0a14ed302c8738b259b94e91ef1885218b71a588016d6f4adc6e1de638b57ea5e1940cbf2317afc200a7c40fa4c4fe67c46a9f3982b4be1d014831aeb4ba23fc7a54f54539c8b261d72958667c84b6b461aa96e65b80f642aa0201070e3b6e8644fb14fc0295f41ea66ebfb1ec365bc15127318b8f1b9f434da9dff2433efb9d9459d9cda16eb9e762d58127e0680f796f3655c28cb2b4a4320ad7faca17464e4a9f529b6549c917dcb6d5c32a72cc887c39a960b3b8991a90cd6055e5812b4b5af880b65980101891468cc078ea0c67cfe99126bc837acdf6c08757b0b3fdea4f792980bd09774694c856dfb159fd22a89cd3cc56f81ec6b829fddcc59dd01424d560f473378c0d85a02ff0a2a14d92ebdc49072872945847e20dd86865c51ade87b1a63d9c71f7e647ebb4df318f70ea3c3cd97115a1bd0e8ad164451f19dc736ac532fd1e6d0c5956cb7512e2d5ea00396f88689bf148808b3bd78bee2159a7bc57e3ae0f325ced6c3bb8f1b06e88b44b26f3ffa6e41cff8acfb668652e11f228816c24afd4b7d94eb58508c7148755b5b2df7f262997eb8681b81be8016fe7189440a680289936c9c874bdbad7db04a218ca8bfebf33641bca3681aae`,
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
		input:       `d406a79a5e5d5eca83ccb30403fd63c799ddd802c83865e120119d81f1eab26f919f9ce12f276ae242948073e0407097e31fa131b34499f4800f09fbfbf8e5f209b793928f3d12f3426659a548320907b25fb7609a6a5aabc2dcb021238e2e925be75ce64a9ff9c6c7907218622cbda77c19095eee7d6bf86942320e078d82cb15f4b10397c9c3679fe00ede416370ad79367d91dd3e56d65968fa1394754dd3790ed13d2b621e9d3b96b24f157be486a91b066b9fb0c93b92edd1fde89e42d4c15f579a3d503c3fb65aafff96b15e28d2179ab654e06338dbfe0b412c99d4848621c4`,
		wantVersion: 4,
	},
	// (Ack₂) CIP-8 encoding
	{
		input:       `019e8f9bf08d580cc5d80b82e7a411986536ec6ab649ebbaf29ff194396a76b34218a58b86df203752ced226549cab712acacff20328cdae1327804812d2ded3675fd695a4581c8672b550f9bce745fef58a9e0ddf23fdb4cd583d8a74f407af55386dbb26edb34007fdd5d7fe6d4c1d65b2b0288293d18cbfb38a60db3cd8879e4a3fe4595d347a17b13594bfeb9571ecc4d1e7cc046ee1d1dbcf05d44206233e2c43873d659d5e8ead3d2907f2154e0ac5bd3f4cabdd5b5fc0c321e443aefa3de1d537025e6aed9840ec4300301eff0afdbccdd01b0684c34ef0adc1d4a28d542a6daf4fc3a973038e33e6b55a14338a433624e2b7840bd0258eec90cef96dd27d6f300acba015e1781215c1f654d47f433385c1e64f9fd6aba4fa2b78e4a172a1614f6680477f8c4fd273503c9d8ed18b8deb37fd4d44f11d13602c7cc272cdaf57073e003bb48163be085f431f3a23694a59fa09aeac17f3b10c5d601d03967c638fd5b2c704db95cf410d02f38123a9555079b5d866ce9535d1b97b7e784fcf9451c8495986e50299b21aa5d9ab8c7b1f3fe0535a3dcadbe53007c0a651`,
		wantVersion: 4,
		wantRest:    []rlp.RawValue{},
	},
	// (Ack₃) CIP-8 encoding with version 57, additional list elements
	{
		input:       `01a43466491884d97389b884cf285d37ccd13063a626c390c9ff95924b0a12f21e7ae155c8eac52105d255ee761ef42778eda5fcf6d73a620cc9803159a00d3b778427c2e3776679ede784f35d6db5630b7f45461eb5024b7ec11ae29e8c72a76ce4d2f1374222b32942b986ffcab6ce7ad0ff0988bdd80112c3122f7883a3913e234a15344f89883cd0f33f03a7dd4f3899b5e46199694225311ac6783d818db8b271b4cd373d980faf739cf374544bf8ed33e6b86fa64f022c6975d29ee83550eec7117d90323ef37fa40e7991902663508c2234892f42cc7693942ca382d31b5624fd87de8c665235d20cdd49846be439c22fdcc65852c7115bbe1a772be15f1e14177a3cb5d6c5c776f20587b933d16329ffa6705af00c21be4bac6062276cf525d5ff0d9d85363639dd4fe6d3893a6a31ef79c1a21cc50bd63c4ccfb610e34e129982a03af543e60e483e96ccc5eb459d65057689a111f8d3f3334084ccfae120731f212209dcba30f8a9ad1bffca54067231d1e5dd7073699d446a182d354c00b32c8e2e2ceeeaeca2d1b9dae01e1b9e26b3bfc215c5081bd071c7f189a910a096c28c`,
		wantVersion: 57,
		wantRest:    []rlp.RawValue{{0x06}, {0xC2, 0x07, 0x08}, {0x81, 0xFA}},
	},
}

func TestHandshakeForwardCompatibility(t *testing.T) {
	var (
		keyA, _       = crypto.HexToEDDSA("ec4f51f2db12a88c2675cb1241e83b83dbe13df604a4c3d4d4482099273e2b07e2e812ed9d035938d5c0a5ee1c4be5602a3fb82cfe6a9b2383")
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
		authSignature = unhex("299ca6acfd35e3d72d8ba3d1e2b60b5561d5af5218eb5bc182045769eb4226910a301acae3b369fffc4a4899d6b02531e89fd4fe36a2cf0d93607ba470b50f7800")
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
		wantAES            = unhex("cdcc598bb62e495bdb9074838d6be8ef43f348924bf307b147f32682ef1d1023")
		wantMAC            = unhex("00081dc2481ba5423b56703bce94d27c502543d202424db5b0933c6352c486c3")
		wantFooIngressHash = unhex("c4823b0caa3b2105551b286fc7f176108f76c7db3f5ce46ff0091bc1f55db53d")
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
