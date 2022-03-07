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
	"enode://d74ae95f2967a731ff1341da75110daa050b603ad6dbf21c91a7d07e2b52674ce8018ba1ef529051e859b7cee358b80f5c0afae83b88f4b280@86.110.248.212:30300",
	"enode://4b2b16a1bf3ee11f87cc45d2039d59936fd91e9b20a55f5e307d8aad9cd1428416d935d306be6cc5b640106877c90f6ca1f12a82362ccc2500@86.110.248.213:30300",
	"enode://12b01bbbadf76c3fc4f1229c6a1af3430b61e8cbdda7d2de39413c949230496bda83ede613ac09f352b87fa7b0526dabaa366f8991b23f9300@86.110.248.214:30300",

	"enode://7fcacb7a3f70e12db2accbc554ffce7889d4b42bfa3f94210c75960a498b8bb8062fe191f9bc4aa33d683dba2e0492ce7335675da912b9a300@n1.coremainnet.net:30300",
	"enode://65749e415a8ebe4f584f7c95de3f5d84dd2f4f1cb789ccae35e67286aab00041797a630ffc0c21781bd10aefc51423f03fd2412b265e0d2d80@n2.coremainnet.net:30300",
	"enode://1806510e9fcbce7f7bbf56c27298c1bbd3d6981bba81194d81882a4d3b6a9fd5d98f037feb6f08d09dc07df646a9c3d2e340196a8d4f0a9280@n3.coremainnet.net:30300",
	"enode://9558dc01f8f804cf7c8ad819baf1182c743214acb6482143f593d2ac787b90c85a0e3965381ce3593c5041ac3aaf06282f24a9f6e8811c7080@n4.coremainnet.net:30300",
	"enode://3dac96b50ca74c3de40d8141f0f58f83fbf98cf6a7126edf6766656242a7f943d92e9d1bb8b0a507aeb05d361499eb7f7df35d05aa53152d80@n5.coremainnet.net:30300",
	"enode://7bf7c9093034622e11099cbb6ab2bb94aa9946e70717eee303bd9ca1f993b5b60c5aef74dc0f0f4f34000f3facb52e21fed732a0caea931180@n6.coremainnet.net:30300",
	"enode://26f0bae496bdebb666ebbe7e787423af7afbd39914bc5b1d4f116ac963e5b5d123c0b9d8a437cd38d70c42ddd239932399e789426fab000880@n7.coremainnet.net:30300",
	"enode://e549dcb77cecb405d7680744a00d3152cd423e258ada735de1cb70260303a2268a81fa7def1d2ccb8c09dbe44894bd7e8516b1b4293d292d00@n8.coremainnet.net:30300",
	"enode://ef4951bb7f967b4d2fbb0ba17cba29d23cc9e05fe19bdcaa0f7a110a4f8e3db19783c9bba5f9626ff7018cf37a9b57258e7c50c1a486d5e080@n9.coremainnet.net:30300",
	"enode://60c38b803e5e1493679a7c6cfc4a685bc6623cec5ca31f104b442edacbacb75bcbf2d7310eb00335fb6f74296b97e1f6a1d395bba3ba8bb580@n10.coremainnet.net:30300",
	"enode://bbb3fb84b4a65a8099521dc1d7716754505b815b2a5b018a1435e3ab080a40d87a75a49516585812f59ad9021d9ca644878680cb0e75df5280@n11.coremainnet.net:30300",
	"enode://a491871a099a255317b44635f68797445e71a30fb00e0766724bc3b57f7a693b563fa67dd8cf40cc52c9b4b41c80cb7348eebae23af93fed80@n12.coremainnet.net:30300",
	"enode://617ef29e72bc1be3a4078208f9fef0826fe9aa2be2f77d5d6e9aa80d83c3571d52fa85334315ad49a1691eca5455d65f40ca009342ceb71f00@n13.coremainnet.net:30300",
	"enode://57ba4f7f74741bc76edb9d9ce6e490da28c2ebb9051d8f1bedfe5f10ff432c950aa6ab54d0fc969f59abdabf7f0053c71734113c37d4ec5b00@n100.coremainnet.net:30300",
}

// DevinBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Devin test network.
var DevinBootnodes = []string{
	"enode://93ecfbf2093dbb87f0bbee0ccdafa3afb74003aeb816513b0ca428acff9dd1ac15966f3e34ef5fb2acb98a477ffc289015d23446f0ec195b00@86.110.248.210:30300",
	"enode://e0c9b6e9df3d45c90e83fb40135eefbc35a0ce411dd04acd005fa1ef4b97e635133629eee486b8f0a65084139a4ac31e8f8c88c8edbb600a00@86.110.248.211:30300",

	"enode://e24c18eca931475b267b511c62e19cc8aa428919111d195e250d59f6edcda4a361d102de1c6c2eb5061aa7b246f820c19d03407d38b975b980@n1.coredevin.net:30300",
	"enode://110aa01954eda50ae7ae1c018fdc9772c72108e0d6e9b0571e6ac5f766d99a3d07156894cb288e9b6ea756d660569743b847b385a539720c00@n2.coredevin.net:30300",
	"enode://e5bcf68387187de4c533bfddde09538e8d0bc35798d6f0d8e84c9e40cb3b645b052920c4b89ec6527a33ef1903a44e9eae545521b50a07d980@n3.coredevin.net:30300",
	"enode://936d469b555223d2cd207b975b9c9a82775513c1b57e3ecdf24c61430cbee7406e3eba1696329002f353f1c5e309e753abd6ff0a815ce96780@n4.coredevin.net:30300",
	"enode://5c9182238e1b46c1f2346d5be703ae02dd1deedae4bf3cce4b57e42b9713067c752832ba94a0e59494442bd48bcb37a7892cf87aa097672f00@5.coredevin.net:30300",
	"enode://f77f9bcb352a64e3aa06ceefe2eaca267fbba6486f654505e88c44d85d3f434e4ba7736bd23ec1e0b5ef48819e5ca0d4d5f3bac344faf4f800@6.coredevin.net:30300",
	"enode://24c539cd87e8f4c37747776ae96842dc6c637a99c8311aef4ea56a81b977c5412f1d5f16694580c602bd53a3985d7dd87252b996c4c5dc2d00@7.coredevin.net:30300",
	"enode://4dc065ffd94b9f9b448a7c7d26277b2fc27f12348b765909d06d6f9d12c93c0d63f782cb6c2c3dc95f20bd97525ae8e6692f3df4e4c62abb80@n100.coredevin.net:30300",
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
