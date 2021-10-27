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

package params

import "github.com/core-coin/go-core/common"

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main Core network.
var MainnetBootnodes = []string{
	// Core Foundation Go Bootnodes
	"", // core-devin-eu-fr
}

// DevinBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Devin test network.
var DevinBootnodes = []string{
	"enode://caeef3fe131f1ed322735e4195ed8a887c3a5316dbd13595e09d98f02ef816c375a8be2baf9dbfac571d6d497c0d1afeb5b014d8df04cc5c00@168.119.231.210:30300", // core-devin-node-9 core-devin-eu-ge-by
	"enode://d67d5db87c8e6bc8139df78c45819675bbd992eded7bbc5c18b1868c53beb7a0b5242e95d7e9ddfdd0ad5f833531302616059fd4f60ce24a80@78.47.203.14:30300",    // core-devin-node-10 core-devin-eu-ge-sn
	"enode://4a19d5ac99da155f54cc5d56751f069a8a7e475c8d0a5fafdcabeb409a9fb71f33008592571cfdc949d8db8733a6ba7ecf8b759c42799baa80@51.15.51.158:30300",    // core-devin-node-18 core-devin-eu-nl-ams
	"enode://6819914de7cc04e0cd8521974e0c0910acf89d862bce0b284afda70d0fe8e7b5115e058d33b54cc6eb73e9f8ccdc3a6400fb1499109769d680@51.158.123.140:30300",  // core-devin-node-20 core-devin-eu-fr-idf
}

// KolibaBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Koliba test network.
var KolibaBootnodes = []string{
	// Upstream bootnodes
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{}

const dnsPrefix = "enrtree://AKA3AM6LPBYEUDMVNU3BSVQJ5AD45Y7YPOHJLEF6W26QOE4VTUDPE@"

// These DNS names provide bootstrap connectivity for public testnets and the mainnet.
// See https://github.com/core-coin/discv4-dns-lists for more information.
var KnownDNSNetworks = map[common.Hash]string{
	MainnetGenesisHash: dnsPrefix + "all.mainnet.corenode.stream",
	DevinGenesisHash:   dnsPrefix + "all.devin.corenode.stream",
	KolibaGenesisHash:  dnsPrefix + "all.koliba.corenode.stream",
}
