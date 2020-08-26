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
	"enode://ff03f84d279eb52ff63441ad9ad9d0a082ed815c0e5964aaa6f37ebe4bffed10827f987802da77a37fbff33d5f425c7fd48c75467efda645@178.33.40.162:30300",  // core-devin-eu-fr-ges
	"enode://953802c24e04e6760e7b0261bdedf54d4de410d6ebb75352f04415e402e1e95c7947f1fdf33a347c4d823f788b323c191f4139aa183717a9@178.32.40.223:30300",  // core-devin-eu-fr-hdf
	"enode://e4c5df04c43da91f810def6fb67a555fd1eeb0663ebba017f0415bc098d6599c9162ea99d132c11b5824f04a4d09e6bebd2d1beafda0c2ce@198.100.153.13:30300", // core-devin-ca-qc
	"enode://70f5059f1ad90e89b74e5c0099b25f9cb053473c97f0a60c98df7c861a98ad3c0738d989ed0e388d1a6636a30cc6666190e8ca1411e210db@51.159.93.3:30300",    // core-devin-eu-nl
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
