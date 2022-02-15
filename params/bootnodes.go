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
	"enode://f9cf86278b202b396f07eb3206092674635a387a7b5e03e2ed8073267fd9dfb70082d7d14d9ea4dddd2891363a1c32303de55b3fbdedae9a80@78.47.203.14:30300",    // core-devin-node-10 core-devin-eu-ge-sn
	"enode://1b5b5a269e98cca3353077e393d46370a2daa21b3da5d3fe2e1edbd42d400021660990a690bed90c0b5777899b51183f5aa5737cc5b5a64800@51.15.51.158:30300",    // core-devin-node-18 core-devin-eu-nl-ams
	"enode://ad8212b20b930a6f8609ccff235e4bcce9189711132f9b76609d51b12fe2900601fc2f81f87991ec0c74d1114fb6b88f019bdb48b8b490e380@51.158.123.140:30300",  // core-devin-node-20 core-devin-eu-fr-idf
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
}
