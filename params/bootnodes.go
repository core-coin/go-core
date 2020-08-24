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
	"enode://7081d1a8a52a9d0b8bdc0247cb7fd04277a8e76a30f85ce8a9daf41a4e8e1f79902d4959416a663a779cd28390b10a881f56d73469deb695@51.75.247.236:30303", // core-devin-eu-fr
}

// DevinBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Devin test network.
var DevinBootnodes = []string{
	"enode://7081d1a8a52a9d0b8bdc0247cb7fd04277a8e76a30f85ce8a9daf41a4e8e1f79902d4959416a663a779cd28390b10a881f56d73469deb695@51.75.247.236:30303", // core-devin-eu-fr
}

// KolibaBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Koliba test network.
var KolibaBootnodes = []string{
	// Upstream bootnodes
	"enode://7081d1a8a52a9d0b8bdc0247cb7fd04277a8e76a30f85ce8a9daf41a4e8e1f79902d4959416a663a779cd28390b10a881f56d73469deb695@51.75.247.236:30303", // core-devin-eu-fr
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
