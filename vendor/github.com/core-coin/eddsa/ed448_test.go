package eddsa

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestEd448(t *testing.T) {
	privateKey, err := Ed448().GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	if len(privateKey.D) !=  ed448_privkey_size {
		t.Fatalf("bad private key length")
	}
	if len(privateKey.X) != ed448_pubkey_size {
		t.Fatalf("bad public key length")
	}
	if bytes.Compare(privateKey.D[56:112], privateKey.X) != 0 {
		t.Fatalf("bad private key: %v %v", privateKey.D[56:144], privateKey.X)
	}
	t.Logf("%v", privateKey.D[56:112])
	t.Logf("%v", privateKey.X)

	cmpPrivateKey, err := Ed448().UnmarshalPriv(privateKey.D)
	if err != nil {
		panic(err)
	}
	if bytes.Compare(privateKey.D, cmpPrivateKey.D) != 0 {
		t.Fatalf("bad private key: %v %v", privateKey.D, cmpPrivateKey.D)
	}
	if bytes.Compare(privateKey.X, cmpPrivateKey.X) != 0 {
		t.Fatalf("bad public key: %v %v", privateKey.X, cmpPrivateKey.X)
	}

	cmpPublicKey, err := Ed448().UnmarshalPub(privateKey.PublicKey.X)
	if err != nil {
		panic(err)
	}
	if bytes.Compare(privateKey.PublicKey.X, cmpPublicKey.X) != 0 {
		t.Fatalf("bad public key: %v %v", privateKey.PublicKey.X, cmpPublicKey.X)
	}

	b := make([]byte, 94)
	rand.Read(b)

	sig, err := privateKey.Sign(b)
	if err != nil {
		panic(err)
	}

	ok := privateKey.Verify(b, sig)
	if !ok {
		t.Fatalf("failed to verify")
	}

	b[0] ^= 0x40
	ok = privateKey.Verify(b, sig)
	if ok {
		t.Fatalf("verified when invalid")
	}

	b[0] ^= 0x40
	sig[0] ^= 0x40
	ok = privateKey.Verify(b, sig)
	if ok {
		t.Fatalf("verified when invalid")
	}

	pub, ok := privateKey.Public().(*PublicKey)
	if !ok || pub != &privateKey.PublicKey {
		t.Fatalf("...")
	}

	pa := PrivateKeyBuffer(privateKey)
	if bytes.Compare(pa[:], privateKey.D) != 0 {
		t.Fatalf("PrivateKeyBuffer is wrong")
	}

	puba := PublicKeyBuffer(&privateKey.PublicKey)
	if bytes.Compare(puba[:], privateKey.X) != 0 {
		t.Fatalf("PublicKeyBuffer is wrong")
	}

	pubb, err := privateKey.SigToPub(sig)
	if err != nil {
		panic(err)
	}
	if bytes.Compare(pubb[:], privateKey.X) != 0 {
		t.Fatalf("SigToPub is wrong")
	}

	privateKeyB, err := Ed448().GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	secretA := Ed448().ComputeSecret(privateKey, &privateKeyB.PublicKey)
	secretB := Ed448().ComputeSecret(privateKeyB, &privateKey.PublicKey)
	if bytes.Compare(secretA[:], secretB[:]) != 0 {
		t.Fatalf("ComputeSecret is wrong")
	}
}
