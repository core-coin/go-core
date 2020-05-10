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
	eddsa "github.com/core-coin/eddsa"
	"net"
	"time"

	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/p2p/enode"
)

// node represents a host on the network.
// The fields of Node may not be modified.
type node struct {
	enode.Node
	addedAt        time.Time // time when the node was added to the table
	livenessChecks uint      // how often liveness was checked
}

type encPubkey [56]byte

func encodePubkey(pub *eddsa.PublicKey) (key encPubkey) {
	copy(key[:], pub.X[:])
	return key
}

func decodePubkey(e encPubkey) (*eddsa.PublicKey, error) {
	return crypto.UnmarshalPubkey(e[:])
}

func (e encPubkey) id() enode.ID {
	return enode.ID(crypto.Keccak256Hash(e[:]))
}

// recoverNodeKey computes the public key used to sign the
// given hash from the signature.
func recoverNodeKey(hash, sig []byte) (key encPubkey, err error) {
	pubkey, err := crypto.Ecrecover(hash, sig)
	if err != nil {
		return key, err
	}
	copy(key[:], pubkey[:])
	return key, nil
}

func wrapNode(n *enode.Node) *node {
	return &node{Node: *n}
}

func wrapNodes(ns []*enode.Node) []*node {
	result := make([]*node, len(ns))
	for i, n := range ns {
		result[i] = wrapNode(n)
	}
	return result
}

func unwrapNode(n *node) *enode.Node {
	return &n.Node
}

func unwrapNodes(ns []*node) []*enode.Node {
	result := make([]*enode.Node, len(ns))
	for i, n := range ns {
		result[i] = unwrapNode(n)
	}
	return result
}

func (n *node) addr() *net.UDPAddr {
	return &net.UDPAddr{IP: n.IP(), Port: n.UDP()}
}

func (n *node) String() string {
	return n.Node.String()
}
