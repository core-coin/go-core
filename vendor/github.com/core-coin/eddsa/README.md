eddsa
=====

[![GoDoc](https://godoc.org/github.com/hlandau/eddsa?status.svg)](https://godoc.org/github.com/hlandau/eddsa) [![Build Status](https://travis-ci.org/hlandau/eddsa.svg?branch=master)](https://travis-ci.org/hlandau/eddsa)

`crypto/rsa` and `crypto/ecdsa` provide `PublicKey` and `PrivateKey` structures
which can be used to unambiguously represent RSA and ECDSA keys.

There is [an implementation of Ed25519](https://github.com/agl/ed25519) for Go,
but it provides basic functions which take pointers to fixed-length arrays. It
is undesirable for code which does type-switches on `interface{}` values to
have to assume that a value of type `*[32]byte` is an Ed25519 public key and a
value of type `*[64]byte` is an Ed25519 private key.

This package wraps [agl/ed25519](https://github.com/agl/ed25519) with a saner
interface much more like `crypto/rsa`, `crypto/ecdsa` and `crypto/elliptic`,
while still allowing you to get the public and private keys as pointers to
fixed-length arrays if you need to.

It is designed to allow other curves to be implemented in future, such as Curve448.
In this regard, the design of this package closely follows `crypto/elliptic`.

Build
-------
export GOPATH=$PWD
git clone https://github.com/core-coin/eddsa.git src/eddsa
cd src/eddsa
go build eddsa.go ed448.go

Licence
-------
    © 2015 Hugo Landau <hlandau@devever.net>  MIT License

