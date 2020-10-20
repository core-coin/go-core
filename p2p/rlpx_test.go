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
	"github.com/core-coin/eddsa"
	"io"
	"io/ioutil"
	"net"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/crypto/ecies"
	"github.com/core-coin/go-core/v2/p2p/simulations/pipes"
	"github.com/core-coin/go-core/v2/rlp"
	"github.com/davecgh/go-spew/spew"
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
		input:       `b8ec62567ab29777109c00b363791fd31b8cd21bac29f35dc2e65fe01b0e470768c7f07337b4f69d2cf8574a73be25734832714f9b9655ffac57daddf65c6d60956072e30635c6312771ce9d5d8b1b452e4a622a26730a811ef0d9656c7df25f408b0a8d783b61c31868a26480b645b8c7a3a8e8eef3752614334e93394ec8870b49eec84afb0ee5ae52584078097dd99891e981599d43975de1b6478bc0293f8e63248d77e9e1c16abe2a21e4beca7a9350f15371063f3c843c5ee90e945048715a395552dec6273994872100119f343b0a34c3d970993559f7f09aecb4cc078ba6a0fe4d1a2821ba3a234feb509519cc21212b1965d0f821d1f5e26ecd48422460902ad44a3a80f665e26a6dd1206f951c82b44875c0c39bb137e609b09c1b70d253e871f43fe11124ec57d5c517fe0849b822a46c01174da965c2b6e2f77f9070202e961978adcb465c183368cb52ae4bd5aefe79d08114d72b468ef984a408c4c4fb99a2cd06282f89a25e8a8db551d9ea1e11c55edd4fdcefc5ece3de4c861e401a57166f886c238fe6233bb9d49ee1437da876476051e1f4990b38a9ad3be5d4e1df994ec0226f874a24479b098f4f1296435a44898bc6471962ae1fe9247cad1fbacb5811ff`,
		isPlain:     true,
		wantVersion: 4,
	},
	// (Auth₂) CIP-8 encoding
	{
		input:       `02470b949f1bc7213dcf8b25340d76442fb3593c7118efd449685e83d62cffc6be2aa68be8ee23239a39b0f2a43b623b8fbaee8fbf536e742a760ecab19efc637ea9b0690465502fc542b0b33509c4b6246a4d36f3867b12e29d5f2b75bfd608c40afb4a0833c96c53199fc7a2cd530782a227112d7540bc98570b16e76c117dd789c339e3e51899b0ee221af3a1d5579209ad5e5730200e32d14e438cc842e6b83961dd2d7ee4900c2276fc870911b1c2da3380f4dd4590f4060adb52de62770a82b3c6f9271aba47ec61bb84692061d0653e0d50d3d743ada97abdbdb7606924cdd138a666e0179caf9b2bbf0edea0d9bdd833c99311bcad2d74471d4527223d9a2c56e3efb104a824f30d2d31f2da64d72ade2042f53e679e44abaa4aae5df3acc1b179e0bb63ff869c88e358f9506136824a9d911260cc9b68484682b3099579714a863a1a9e2f6c6161618afc89e5200b0a87b8be0e024ed4159080c4e197772b6f2d48525acbe11355f3140de50373dfa6b3cf2c5d71565234637abe12d4ff81b73260c6f226fd0b3e3be80c0e0ba5d25e9f4ed157ae583dc1e0c99f10be4d615c00484ebffbfc3615310358abb2444bc97168bcd0e27cb349598af35c9a388e94b63152ac96e1d3853253474d41d3010ffc8f22a06c9988f84a710988ef4dc2e9b18d32138aab56f2ea75a7c0970b506297ebe18e3f5198a55349d3c810a35a020483603ac47040ba2a9505bc5cd48d9fdc6571054d3852491f2e036d7f5f5d925da391e51a1ae583bb5bd44f930d26f88f4b2d4439c66cac206f53fa3fb2b96adb32399573`,
		wantVersion: 4,
		wantRest:    []rlp.RawValue{},
	},
	// (Auth₃) RLPx v4 CIP-8 encoding with version 56, additional list elements
	{
		input:       `024cfc58214e0f9842ced5a3eab08ab00fa9fecf1d0edcebd36edae3992bc5e12c0e05aeacf7ad4808b9fa1f7031e69698d0979f04ade73d75dcb9b87fdc161ed14a02eea618cb7017195c30634cb2deb3f654597e6a6b4365191620dd842a39e7291ca79093881818110423ce70d2a5b98c8a3cd52f01f8b7d1565c1b9ad286f83a91625c52f0cf9cff347e44a9706801292ff84e2547e9f84a2e1adf7c2fadb836bcc941705dbad5c20065997de4a11c1ad4df8d4e8671f6fac1cc36ad075c7c8948d70383015ad1fc4d0b8abf9f67da11c17a2653983ab157255114f567da72bc04249032b591322492d0e96f7f841497f89d53850cd7a65a5422ef28a1f534467f0a1effafb2b798e4339e8cfeda9c27b9aec5556ebf6c605574ddeabd5de506340216c1b79dfd1f8807a3ba9e6ef99540c2f0b1ce1778102e66c2b79c5818a642eda45852dfa8934b0a4bab54753d2b7189e395a5850d0f75f2df71d778b8b931f716011edfa867546fa5ea23cf54db9cc6ee08318bb2c313325cf73fe610c7974762fb280cd5416a1b51ed8311ab62595141c7ea89dbca0e6710f585b22b6081143f132bf85909f03d7db475b9b5d447bf436a12bf17ccd17667639c117b44d7c7c15eb7ead08d9110e681a4470a2a5c6466859ea455d1f085bbb0fb84dd507a02e12f8d61703d0fff6f8d2146af3e93da29ae381425cfb092fde50056cc1fa5010dcdd359010ccad760cecf24d301ef3fdfab5a8d95feea59924c60d4aa2afaee6a6702f7a0adcf3ae6b706fafa9a68f115f585d08a1b123c4a43a18e7e3cbb22b2b4109971fb8ad564d7`,
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
		input:       `cf3f09b06c6739e072d89b6d860df06e5b9b01291c1af8dc52c63a0c061e6fe9e5e81ec5046e4c29795af2a890f69d7b743aef409a0d40d3e9cffe3fb602a1378db3e27848cf2a71ed861f2aedc5ea89fd8f5717d240ab04bde0d9512061980f93f3f04bb0d9588947b92df7ae09f3cd0e04f28376c8eab685bae7b2ec81ad389f75479e20464f7245bbeb127251569df91843bc316b5003cd2a5458a5ca3a6ceba9eebcc0d3ab4c5942fa6152a123591c3374107593878e1622288f35eb8501d4fdb0d1429e9392e09d6e445f77c7b47ae6dcc52bce41b7a138f9ab6de21a159d`,
		wantVersion: 4,
	},
	// (Ack₂) CIP-8 encoding
	{
		input:       `019cbc523128ebd1970d46a6b21ac8cbeb1cb7b4ba767f4ed2be5d3568aaa5dab8bb51b0c3f016765a903deb2027090eb0b50aa8fa3eb24631587ff09101e5ea130d47fb4460d7e172b158bfe8c53e06d19d685018bf1f644427d0f4ffaa1a2931845c34abf557bc5b00860dfbc426e71056dd534af7a925ed0d63e5f6b644f99c3a3d08dbfa7fe34dc2572f8f3100ddc87e88b62fab1c9ab1ccf30715a68f519e4e4edc1cf2b661102ee754e47ea05299a411c132be4c0cee32ac7de314885d19a859c56646fb6030179aad56a12cdddfb8617bdc8dc68b51b9a17528c7f6bf3fc9da31151df44187e1976985e44e7db33bcf451504dd248fddb165c3d847f4f1bd54f8033ddc45efc7b55a83441b28ec99943a9153ada3e7c0b55e96ef9499708f1f45d2a783f6c19848cd125637f87f4876b4110e39e2d864d057b58de1b364d9a9086c2f874b036e7b4853fd9b95beb31510580c971877aa92b814dce29843853cbd38361859b730990f9de4dc31dd70d0f92859e284d50ab975148b0f99d9dfd42fa0df31166a4b52a141ec8d0a9a3f775b28e4c0b4630edb7ae6c8`,
		wantVersion: 4,
		wantRest:    []rlp.RawValue{},
	},
	// (Ack₃) CIP-8 encoding with version 57, additional list elements
	{
		input:       `01a276aebf7cf63d53295230ebed1f1c5e20e9badbd0b5f26f9e40d606d58297dc346557289afb0cb60f945d583ee64783b51e1a782cd3c6c08e046e9ade544193663b2f36593eba6b909348f241986a3ba7a3ea9c5204a3970b489d46d5b2e93b003d5b182fcc3299340823ea475bcc5ab71708b53cd4684539a9997c3c6776a6037fca0e6d4bbff3cd11958fbe0945a8a02793a54568b350c8560ceaacf604a9f36ed62a3d9b1cdeb6f12b8fca0800a342cef65da01d24bdd28ea75f958398b0e63cab020ab3632b647c8717a75c9410bc12acda6a7cb7bdbc94e6779932c6363f25178b68c1f64c75bd463855fd73d6f55a8ef88382a97224009e94a9d7689e0daae9af58f8b3213651f44be865af1c2bf9324817bc945c19a20b230a6086c368d96ba5f4a08866fbf5d4c721419e9db6dc008254aacdcf9e28b009db09d130204d43d85d05944e8295fec3b66d746856a0a64e8d96338d164e28a8685d013d650757ffcbeffaa52d4a8980036fa1719d75a856ed8b44e99bbd9c3f9ce3bf40777800462ddc696c88bd2ae31cd14de204e769905f91cd59fe1a48cd5ea521c2afcfa2`,
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
		wantAES            = unhex("ac14d7f2b36e3869666788316d6b8beeea2af3debdee374a12d43f6807e398aa")
		wantMAC            = unhex("c0119ca14d507c205da36a00bb52199efde3dcdea718229e9f2d0eeb4aa0a068")
		wantFooIngressHash = unhex("79c4970f30ae34fdad070c2efb8ab46000d1f0fbb207586d668f18a4cc2d4e74")
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
