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

package p2p

import (
	"bytes"
	eddsa "github.com/core-coin/eddsa"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/crypto/ecies"
	"github.com/core-coin/go-core/p2p/simulations/pipes"
	"github.com/core-coin/go-core/rlp"
	"golang.org/x/crypto/sha3"
)

func TestSharedSecret(t *testing.T) {
	prv0, _ := crypto.GenerateKey(rand.Reader) // = eddsa.GenerateKey(crypto.S256(), rand.Reader)
	pub0 := &prv0.PublicKey
	prv1, _ := crypto.GenerateKey(rand.Reader)
	pub1 := &prv1.PublicKey

	ss0, err := ecies.ImportEDDSA(prv0).GenerateShared(ecies.ImportEDDSAPublic(pub1), sskLen, sskLen)
	if err != nil {
		return
	}
	ss1, err := ecies.ImportEDDSA(prv1).GenerateShared(ecies.ImportEDDSAPublic(pub0), sskLen, sskLen)
	if err != nil {
		return
	}
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

		r.pubkey, r.err = c0.doEncHandshake(prv0, &prv1.PublicKey)
		if r.err != nil {
			return
		}
		if !reflect.DeepEqual(r.pubkey, &prv1.PublicKey) {
			r.err = fmt.Errorf("remote pubkey mismatch: got %v, want: %v", r.pubkey, &prv1.PublicKey)
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
		if !reflect.DeepEqual(r.pubkey, &prv0.PublicKey) {
			r.err = fmt.Errorf("remote ID mismatch: got %v, want: %v", r.pubkey, &prv0.PublicKey)
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
		pub0    = crypto.FromEDDSAPub(&prv0.PublicKey)
		hs0     = &protoHandshake{Version: 3, ID: pub0, Caps: []Cap{{"a", 0}, {"b", 2}}}

		prv1, _ = crypto.GenerateKey(rand.Reader)
		pub1    = crypto.FromEDDSAPub(&prv1.PublicKey)
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
		rpubkey, err := rlpx.doEncHandshake(prv0, &prv1.PublicKey)
		if err != nil {
			t.Errorf("dial side enc handshake failed: %v", err)
			return
		}
		if !reflect.DeepEqual(rpubkey, &prv1.PublicKey) {
			t.Errorf("dial side remote pubkey mismatch: got %v, want %v", rpubkey, &prv1.PublicKey)
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
		if !reflect.DeepEqual(rpubkey, &prv0.PublicKey) {
			t.Errorf("listen side remote pubkey mismatch: got %v, want %v", rpubkey, &prv0.PublicKey)
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
		AES:        crypto.Keccak256(),
		MAC:        crypto.Keccak256(),
		IngressMAC: hash,
		EgressMAC:  hash,
	})

	golden := unhex(`
00828ddae471818bb0bfa6b551d1cb42
01010101010101010101010101010101
ba628a4ba590cb43f7848f41c4382885
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
		EgressMAC:  sha3.NewLegacyKeccak256(),
		IngressMAC: sha3.NewLegacyKeccak256(),
	}
	s1.EgressMAC.Write(egressMACinit)
	s1.IngressMAC.Write(ingressMACinit)
	rw1 := newRLPXFrameRW(conn, s1)

	s2 := secrets{
		AES:        aesSecret,
		MAC:        macSecret,
		EgressMAC:  sha3.NewLegacyKeccak256(),
		IngressMAC: sha3.NewLegacyKeccak256(),
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
		input: `8ecb2481c2ee7018dd29f045c93f49e644352633064b18c1a847a96f114461650616a457ea69275dbeb38a53e2f9147db89abe85ac951f8b1739d07273bea0968e646cb08e212f0641f7a43c311b40936541923ad7ae6f3927168e69571cd5b7a61d0ddec45876057309b4120a4d3de24e090bfac314b9286c043eea94c5e3bdcf541cd744802fa028d4bf1b0841934a98f47d323afd96c46be5ce401a5077f2fe84e958e5f52beecfd1eb5352f029f3ba9d743e1495d65b29a549c861ea269b9fde4456f6f15838b7d0049d2c6ecc5307dd690140cfeeb37e440e54796a03cda729bff51c1fbf66a71d2a85931f1ac4a8c14c69e0433f59bb10a9065f24eb91911a365d76f75ad67befa737a3f48019869206d5e9f5500de6fc7e5e2487837389454fd1b5e4c84180ab5af28317989b93814fc3ec7b0ef61fe17ccfce1914a1a66a9d6471e299dbe06cab902ff4b9d87e0cadcbb3acfee82d6801a48facc0214433b3e5d342bfccdc919cdaec1548eb420a1067f0dfbece4420b93030d0d4a3d5a817913f23ff3c7f940dd08ebc2ac1f4fcdf923a0d7876830e5a81c7d9036105ffd14b3fcf826428c3afc9859f3c3555442d3e1e58c64844794a1af1014fc238d5d1a46b4953800e`,
		isPlain:     true,
		wantVersion: 4,
	},
	// (Auth₂) CIP-8 encoding
	{
		input: `024724846bb478d4e8123254ee2c321215458b4c47e451d0eb9bc5b4503bcfe39500ec6b1f516cdcca7d2dcbf8833107efe9ce82594c5913f60b71f48339709af3ccd79a7ceaf4f95c3a328fc19bab7ec9c901785b32ae20f9c4f239dea5987bb19eaa6fec73df64c7372a941c7d0532856fdd21c34fd1a2aaa19c7122977e0ad4b3c249da462a04546ac3c170639e828bacd9171c79b0ef018a9c0802c36237fba1d470dc30f39ae495a3f3cb06cbc7d7b5611c2b5cfab0088ef0ab9d478688f44c99e0581e2846e4c5fc4d4f7d5b50991a252076604ba3a6303a3d0185dc3d663e8ed852a2a820c33ad1154e1b8140b87f3f2bbdfdd8855416cbb857be64b3b72ccb7e7fe8eaca4a45f45dcbbaf611f4b78ffa5bc83fe77a5d0d24792936197ba89bad0946d2b542b519f415b8f65fae797e9c318ec0c887ee9a5171983e6e7e7f19388718ed8c2b4d609519545541779c627ae4e0ee376db9701f7fb0f9fd289d45db397636b6f763115099adea1c70233d35e145e750c3b2f13957b08b9d324a16041db9a333596b88de5036c24d9319ad8b7a7cecee6add0730046b46daf24c659ae0fa14ddd7e3785fe014df4139eca4fc8a7d3f7a085216affee62963143cf993c2934e8c48003f1644ba986ac163b79471c522a2d9997372d2f86354f30b099cb429115d8825027dcee11ce83a74372a38e87e96fabed0816c827f9efdf095139cff6920d7d1a84a2c453d2e9be50020334df00b1a70a87624336bf8c08b6bdfb4f04e37c5a93810a300ce50257029d29c8cf3409494c46e648de6a36b2dbb5d93da3a058d`,
		wantVersion: 4,
		wantRest:    []rlp.RawValue{},
	},
	// (Auth₃) RLPx v4 CIP-8 encoding with version 56, additional list elements
	{
		input: `024c2f7a121aa610c3a0ac371536571df2a1782ae62f8ae5b91f14a9e7a1f6ca887f2d2ee92ba7ca49395089fc010dd44ce65b729878d8114766a3449f9477c869f147e1e9e5658559a75d51431b21e1db17f13cefa5ed0c6ca32b5828fa6288b96062fe22a76980e322029b33d07213644e58cb3045d52f12bfac2c2280573c486bac2c74de5d5da0638c75fe376111def0f9e9da2b03d7cd56493dbf5ae3b3bc0ac32aeced270ec07a068d93ec683df877f207b12aa03eba490ca3ad1240ef903db1c41318577c272b68f68ce5f10d5e6cc8bec568c00c8d7c3c899f42adce48aeb2621247afc638dac77f3ce048baf1b75bd9e36d94411f3a380769c73fd6b9e5290946e47191ef1f4c87b184e7296a14c2365db25c5bb2f0882b1d36a6043e94455959474b2745b1d0edc5058861c4f76af92446b2de4d1bfbae1e34873b805dd74a5c5c54895f2313a3b91f9e0e708fefbe180cdf48e74f5089964727d3a8ed6f881639d3b73ce876bd1dabc789c4bd68e6b0db3bbc7b53f9f5c54af94302ed13124603f2d45f05b03f2beb14990e64ea1c0ff5833036bbb9669eff96cd69e3162dd7d9b8a09b2d49f65468d65ba498b79e7a3be6d0788e103bfbf8321f1f8d5ea27e34f67aa05f80c74f880b65eda58dd6a84b61de25d608a94b2ce3c4d77247f8859d96ea107a569be85e20c128679636d12d4945e26a28fff3a2d9b32ee66f3bf4ed05267ae9f76eba53dce7ed8c630026f0114ebdf5126eb088bd0c2bdd76fcede1a4ef4ff3e231f8a12c2eceb47a52742aabb778e6a3a763ee4abe1815f4e9c5f146f6ea832f231262`,
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
		input: `206e18faa45f217fe75d59c3f6ea05779a20a36e297ba9046e413b5b22c2800fe540b7d61e41098978a9dc63f0c6feab012313c70e897a4aa698e40d80783d2cc142ad607ae1faa6f5e95e710ff89aee7ce0f43935458d31ea709c12d23c4eaac8895523773185e7419675f11ebcb87e222125f7f4330da0f5af6dc8d3e259c05d580ce0e4ec8aa6c1142c1dc575547b4cfbe94bbb62954416bd34e61175543b1b5614cac7abcba6b76a5fd868efa58a255d3e6bcd5e33b9b76a040d6c2d0d1fcadea74812fc6a91e9e073a516c5f527b5990621fb1c4f455385279f91d33a16ba`,
		wantVersion: 4,
	},
	// (Ack₂) CIP-8 encoding
	{
		input: `019c6449a9404d13da141579c6dc4e8a288acc6cf599362af25d83335fd8aba13871137362730e615ef8b8c1d0b453bbc46e199fe69a061172583975883b2f89c1fd2f974f090a9e2060970155a4a3b53b0fa5d272cde53f4d8a0aa2e2a21408fc6693cfd968c236c374f7571a116c3e4ea5e399732c826a96626af7ade6264d45181e2e2911c901201290b502e97e9c9c4b2d06c5ce16b5aa8780ef17ea8d7ae3d02f6b43509b74918dd6ed9b86b27c27ba77a426986206cab09ce46471350df65eac051c7496c9def976fd534c6df8c80926c1945280bbf10f1e2f8dd8f521409d0b2ddda22cca5730aaa53f783b96098b20d4022e98ee93a1537048b577d0d7a0d82858ca0b4f649d9c2b5ac28a2063bc07467656e190a015121926265d7b01ccda43a082c27c401a985d1f18340d1ad8d1d173e215707fa2320ce611223faf75a25052ab981d378c4962e83fda2ed7df5616ce01298f6713252c97e7723cca1155af2ca9f9a45ef3321fddfb35db49377d715dd072627b7bd8807b3c82b50f10bb4499077d5d4694e4a71d9855724c230a33c0cd798768bba940e453`,
		wantVersion: 4,
		wantRest:    []rlp.RawValue{},
	},
	// (Ack₃) CIP-8 encoding with version 57, additional list elements
	{
		input: `01a286450a5c21857fc9ed284593dd35bb33183d7d79ae506163e1a4071edbd4664d478913762422fe921ab748d7c1a7eca88b4acebcf5fecf3f0ae72429dbf4a701dc0d0c9ec18b49f3a14b588ddca4cbe185f3c5820b625f2f41684b69e01ec114ed896c862a729bef660679ac3f1f44f0875639bdcb46d630f28be60c7de35d65ddb880c4bedb72606a9173633697027c306697beca92d354a7e21102ae4e2ed7d2748a800ec8a22b1f7d6938ca8a0157ac6fbf4ab6e32757706bb7f4317f7b97c1feeee487167206899f83d3ca53a05eaeb5bb58f8b287e9684085ef656326caf9451ed23210792434edfee6c8d4f89d44c2bf624c79fc776590ae7ce45516c0d75780395924162781f37899639561aa31205129be552544b2b54415cfe02017d959ecbedcd12c65c346c8ebee5c794755d0929879ecde042e278a8491247dd8a06a3d5b94f98dc6b50de96f71bc5a3334c92f22fabed2662af9928fa30a22c9eb58b358f7d2aa3257e4e7033366ba94f710fa6a439bed1d8b7347279ee74813f705eaac680b84a27a602adc5864e41fb26f19c73cf11a1c91dfa84c95c7691544be`,
		wantVersion: 57,
		wantRest:    []rlp.RawValue{{0x06}, {0xC2, 0x07, 0x08}, {0x81, 0xFA}},
	},
}

func TestHandshakeForwardCompatibility(t *testing.T) {
	var (
		keyA, _       = crypto.HexToEDDSA("ec4f51f2db12a88c2675cb1241e83b83dbe13df604a4c3d4d4482099273e2b07e2e812ed9d035938d5c0a5ee1c4be5602a3fb82cfe6a9b2383e6c839b66f15fd1b172bd0ccf0a00e5a4ca1f8675a9aa1251c5375d2dd8eccb3d637820a0204faf8e110911a25501a6a8200c633d5b7f8553c5662abd270756f096b04e0a834a49cf218c5fce341ec9af5e47d1fe7bf6d")
		keyB, _       = crypto.HexToEDDSA("856a9af6b0b651dd2f43b5e12193652ec1701c4da6f1c0d2a366ac4b9dabc9433ef09e41ca129552bd2c029086d9b03604de872a3b3432041f0b5df32640f4fff3e5160c27e9cfb1eae29afaa950d53885c63a2bdca47e0e49a8f69896e632e4b23e9d956f51d2f90adf22dae8e922b99bbeddf50472f9a08908167d9eddce7077f0bf6b3baaab2ebe66a80e0b0466a4")
		pubA          = crypto.FromEDDSAPub(&keyA.PublicKey)
		pubB          = crypto.FromEDDSAPub(&keyB.PublicKey)
		ephA, _       = crypto.HexToEDDSA("45516f2d6e60098e547e9b50d386e75f530805fb468c132bead2ce7b205208d895cb086fff390eff73c349a7e5caf1c8c8d8278ae31a6b175a5280ba4b5fd3f28a70138c81a4334eb1d16a35b09f0e272667f320a26c40fe22117f34d131d217b3b172a04532a33eb0cf148d501887293956ab04737a0d08e21fc151203a8ab402afa497899d16b2a84c7736ef1d07b1")
		ephB, _       = crypto.HexToEDDSA("96b3c4485ef83aae585776685bed5d7d6373befb7b661f43592ac703b94ed543526a23d4de35af35c30690998993f140ed1fd9389bc99506b98ac408e75d35449de00b8fc89c042f3cea4dfdd3dcc7878a836edb2a5516163ee5218a0af44e80c7d4ad114d5302109289c29925de77c82fb0e081f0732c15dbe54440ea719327d13acbbb3aaebd58dbc6e0a5c83c5c06")
		ephPubA       = crypto.FromEDDSAPub(&ephA.PublicKey)
		ephPubB       = crypto.FromEDDSAPub(&ephB.PublicKey)
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
			randomPrivKey: ecies.ImportEDDSA(ephB),
		}
		authCiphertext     = unhex(cip8HandshakeAuthTests[1].input)
		authRespCiphertext = unhex(cip8HandshakeRespTests[1].input)
		authMsg            = makeAuth(cip8HandshakeAuthTests[1])
		wantAES            = unhex("d5b5c6435460380778791811be8d3a354a17cbcc8a49a742bfc679e5db852c5e")
		wantMAC            = unhex("f4a7d65c026ba34aeaf0c7446a2c36fbfc88918bcf9541bbff84ac0bf48fbbd7")
		wantFooIngressHash = unhex("0d6d8c38e8befa00cbc118765cdc7c31bb8224d5109f72e80d9bc42a499a70d0")
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
