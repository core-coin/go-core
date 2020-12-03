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
	"enode://e0595284b9c7401656e3673bc2b36eebf1ddf6b8e66d7b302d08e4d49bff5d092cb53eed3057daceb2700f87fc872de8a7dc7c576a5f19bd@51.91.144.151:30300",  // core-devin-eu-fr-ges
	"enode://dae877207371079c53192408b7125a106f82b536f7a4afba975999dde1b739d68b94ed9860f9186398dda7ada0f1f0d2369f6da3b69b43ac@51.161.81.114:30300",  // core-devin-ca-qc
	"enode://cd23371d6a7cb76e5f5acf05d849e917f9e346cddf21ce3ed89139e687cdba8eb90e26512ea5c76f2214e3a66b6623fa0bb9a827c74fa18e@51.158.123.140:30300", // core-devin-eu-fr-idf
	"enode://2337cf2845586bb535f388c8acb70459f0ec93c8e33f4f385af3d2886833e9cfdf848233aa794d254a584cda0698240bcca187e70f1670d6@51.15.51.158:30300",   // core-devin-eu-nl-ams
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
