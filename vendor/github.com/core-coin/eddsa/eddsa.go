// Package eddsa provides structures to represent EdDSA public and private
// keys.
//
// Currently, this only supports Ed25519, but could in the future support other
// EdDSA algorithms such as Ed449.
package eddsa

import (
	"crypto"
	"crypto/sha512"
	"errors"
	"io"
)

// An EdDSA curve implementation.
type Curve interface {
	// Generate a new keypair using the given entropy source.
	GenerateKey(rand io.Reader) (*PrivateKey, error)

	// Generate a signature. Unlike for crypto/ecdsa, there is no particular
	// requirement that the data you pass be the output of a hash function,
	// or that it be small. The only requirement is that you pass the data as
	// a slice. In other words, it must be practical to load the data contiguously
	// into memory.
	//
	// If you do not wish to pass the data itself, or it is not practical to
	// pass the data as a single contiguous memory block for signature generation,
	// you can instead pass a hash of the data. The IETF CFRG EdDSA specifications
	// specify how to construct a prehashed variant of EdX (for some value of X)
	// from EdX, in particular what hash function should be used. For example,
	// Ed25519 and Ed449 require the use of SHA512, and the 64-byte binary output
	// of SHA512 is then passed as the data to be signed. This construction is called
	// Ed25519ph/Ed449ph. You can use the PreferredPrehash method to get information
	// on the recommended prehash function to use.
	//
	// You should note that by doing this, you lose the collision resistance
	// property of EdDSA. Thus you are rendered more vulnerable to any vulnerability
	// in the prehash function to collisions. This is a rather academic concern.
	Sign(priv *PrivateKey, data []byte) ([]byte, error)

	// Verify a signature. Returns true if the signature is valid. If the signature
	// is invalid, or the wrong length, or the key given is for a different curve,
	// returns false.
	Verify(pub *PublicKey, data, sig []byte) bool

	// Returns the algorithm name, e.g. "Ed25519".
	Name() string

	// Returns the public key, private key and signature sizes in bytes.
	KeySize() (publicKeySize, privateKeySize, signatureSize int)

	// Provides information on the preferred prehash function.
	// You can use this to construct EdXph (for some supported X) from EdX.
	// If preferred prehash is unknown, returns zero values.
	PreferredPrehash() (prehash crypto.Hash, prehashName string) // e.g. "SHA512"

	// SigToPub returns the public key that created the given signature
	SigToPub(sig []byte) ([]byte, error)

	// converts bytes to a private key object
	UnmarshalPub(buffer []byte) (*PublicKey, error)

	// converts bytes to a public key object
	UnmarshalPriv(buffer []byte) (*PrivateKey, error)

	// Shared secret
	ComputeSecret(priv *PrivateKey, pub *PublicKey) (secret [sha512.Size]byte)
}

// PublicKey represents an EdDSA public key.
type PublicKey struct {
	Curve
	X []byte // Fixed-length public key, suitable for marshalling as-is.
}

// PrivateKey represents an EdDSA private key.
//
// The public key always appears at the end of D. That is, D[len(D)-len(X):]  is
// equal to X for all EdDSA keypairs. That the public key is duplicated as
// the X field in the PublicKey field of the PrivateKey is a matter of convenience
// for the purposes of defining a reasonable interface.
//
// It is possible to recover the public key from the private key. Thus, a
// private key can be compressed to its first part.
type PrivateKey struct {
	PublicKey
	D []byte // Fixed-length private key, suitable for marshalling as-is.
}

// Public returns the public key corresponding to priv.
func (priv *PrivateKey) Public() crypto.PublicKey {
	return &priv.PublicKey
}

// See Curve.Sign.
func (priv *PrivateKey) Sign(data []byte) ([]byte, error) {
	return priv.Curve.Sign(priv, data)
}

// See Curve.Verify.
func (pub *PublicKey) Verify(data, sig []byte) bool {
	return pub.Curve.Verify(pub, data, sig)
}

var errInvalidPublicKey = errors.New("invalid public key")
var errInvalidPrivateKey = errors.New("invalid private key")
var errInvalidSignature = errors.New("invalid signature")
