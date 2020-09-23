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
	"enode://92fd9892e06692c9c4fb4d7e63ecadd67703c60c6082bfdfcd55a367b5782f0a0389cd7d9761aadb2ab2a08b35cfa2712c1ed59120af63b6@77.73.68.177:30300", // core-devin-ca-qc
	"enode://40b7524d2096b03af81279b3a3213e49aea27e8d4541949bae50a1f7482dd456576528c82fd7f4122eaeb487d4d88aa6d905a05fadbbfdda@77.73.68.195:30300", // core-devin-eu-nl
}

// KolibaBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Koliba test network.
var KolibaBootnodes = []string{
	// Upstream bootnodes
	"enode://7081d1a8a52a9d0b8bd/c0247cb7fd04277a8e76a30f85ce8a9daf41a4e8e1f79902d4959416a663a779cd28390b10a881f56d73469deb695@51.75.247.236:30303", // core-devin-eu-fr
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
