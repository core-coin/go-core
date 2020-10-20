// Copyright 2016 by the Authors
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

package types

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/core-coin/go-core/v2/params"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/rlp"
)

func TestNucleusSigning(t *testing.T) {
	key, _ := crypto.GenerateKey(rand.Reader)
	addr := crypto.PubkeyToAddress(key.PublicKey)

	signer := NewNucleusSigner(params.AllCryptoreProtocolChanges.ChainID)
	tx, err := SignTx(NewTransaction(0, addr, new(big.Int), 0, new(big.Int), nil), signer, key)
	if err != nil {
		t.Fatal(err)
	}

	from, err := Sender(signer, tx)
	if err != nil {
		t.Fatal(err)
	}
	if from != addr {
		t.Errorf("exected from and address to be equal. Got %x want %x", from, addr)
	}
}

func TestNucleusSigningDecoding(t *testing.T) {
	for i, test := range []struct {
		txRlp, addr string
	}{
		{"f8b180808002808080b8a8c5c40330f25b191ac61a620ca88333acf9b4b07d84ad7577473dc2e0d52caf4fe2bf3e13520d092ba0d190ea21b32125d903ee18fbe186a94e321e7dc7c2d74df5a6b0150cc01b3ba7d9ec810b721b94e8323925ec695574d360a6b491025f1e8fc8c2bd3afc705d285ff4ee2f4e733989ff1b2a741482d9c0f7a63891a4fa98b4f120fbbf0ef1a1938b8d6c35202f1b7595455d2bea8e26d7905a23b9b34484263c24e8c6ad65d5", "cb48e8e3cbdb63c9a183882f2549db7dea52a8bcc7d5"},
		{"f8b180808002808080b8a824219a12c54a749ba01a6c204520b6bcc1d1434336a2c5134074a8ff9c4c08267fe68be1922717cea336b27b78288d457ecd9f0255529549646c65aebe6cc75c468180bd32a3e8082f9a20f8175e0cfc3505a59ed71b4247c59c883ace62cdca015cb67cb10b661d8b9be207c09e72067b07b5dd099ecbd7e2c620098ad7e81653ffa949c3b145c774b43ce534bc3e9be27a0cdfd40fe8c785986a309dd3eebcade91a95247368f3", "cb133b72340f010ad0372eaaccb5f60dcdb533595952"},
		{"f8b180808002808080b8a85558067e26aac67770cde54e7a03437c351b36ae84d9224abc7dd84c988f65c1037293032abf1e5d30ca23b6535d52e2d337b68635536fec4f1c663d4044182bf4a4e0de80f7cba24021defa3a6a64d41c1ccfaf22642d927358d86f91dfe79ff14297b6508cb26a312907d3f3f2d5336d6683edd01cdc2650d5bd2316d8371d04f0b856ad88a033b831506f57994a5b9d84a7e63766a548a00bcd3f383988ad5a5c3687470afc04", "cb74f5f73365e237909c70de76667f31410d23d5bb7f"},
		{"f8b180808002808080b8a816c783ef8d45703a43715dc5242d9f79f9c1f956b3f98ff456adde46a93337a33e5dfda6463f8f41b0448e5163393da9863009b85ba2ac4f0c59243957b4ef6c41e02f5900860400c40032b38da3afeb77874fa367c697232f422c197f311a5b24eb89fff745e70a14e58f8cd9e2d100f5cd96a9e704975c722d8aad8471d501a5cb8d271755878f70b306f2d25a8122d81db6d1a8eae6b2b10572dab4c1259db42a03a6cb2a2d4a", "cb78073f912610f64a12216ce68ff9468bb86728b3ff"},
		{"f8b180808002808080b8a8806854edafb41c7909a64e480c01c32dec14fa31a20c354138ea329b2d6076fad2c6644d2b19461815f952a5e58fb9dba6774418eb73ba835c851ab951f168aa787a5058dc285b6432205d97f32b75e19ef47815cf1c91f3516d412f92f01b5d997ba21285476f3b3efb684d5c92da1bd23dba21c6943cb13d749e0b88f5c13d651e1fa50192d27a8a269942a1896b71abcbfd77fc250e72aa5aa70a03f1fe91594ccc89cf2f2265", "cb15815de6a9b1a367664d288be2746bf1cbf31b6d93"},
		{"f8b180808002808080b8a895b2cb4573464b013bbd907d9095015fc0e9b7bd0d44f18b2aadbd89f8c1fc3d9247d9174970b49eb747ed6a4b4655ce1ba91c3af127d264cdeef3a9ba4267ad7486dc02e3a2bc92a55539e77451706732a1ab11504e700d64faa8bec358e22ec45c86af58215ea9205a6f9fe401892837841cb1ac8c343f0f9c7a4947aedda5485b49b10f70203fa189801ef80c74fbd43096d554d33becd53e2e6a5cd12ff0cc804baf43d24b06", "cb52d04b987c5240fafc57e4f61878492290f7509dce"},
		{"f8b180808002808080b8a85cd71ac85a780d3a29d5b685560943a01c12059f349364e932aa9ebee669b9807c2fcdf706cacf8489ef2c24dd73d7a5a52f5ab1affa444d19340c96c73ed9735b13d303ac15e4ea8c9b3d34e8a8bcf852677aaa4e84ca4e5c499829b6835dd59da1da54c3095c036f7a72ca44afc011ac27cfaaa25b439beabf60989acf9dd793682ac6a113d68a30798c58a6ab7d6510b2c88823712e0c3d4764dcd73e59bfd41552143c279b3c", "cb178f3076930fdf26862a1b61a09e0535d302ec16f4"},
		{"f8b180808002808080b8a8cbe7886534d7eb40e726cb6ac3e7a1edda5762296d53230880594858faadff3f52632fb0f73a9d09d948488dd9eb0ee1fb6441cefd0f5e899c64481ef148b5da1a05e2e8a28f7152ec3cedee424bcc39d083024a6e4c98252de0b34bdc1982cd039303e93be1e00f7ebfcefe35027a299d32d9df3c6d8cf081612aa2b018ce475a2568b7dfc4d9330dae6f6fb79cff55845daca82db3cbf6396d4ca310819a722186597720079101", "cb756b0c743e330a533bfd4d9f03e3f05a12303c4cc6"},
		{"f8b180808002808080b8a8005630ffe856e915888e86e89c1d3b091cd701f9a7f8d33f9e92ca659d7723eeff16bc687c9abcdca20f60c8a1904bd9f69083aea3e11a0bccc74bf223aaa1f7994ea384f79e6fdacfc5b5871e4bc2b0e488541fe167b16a28a5bc6af8b5dc9bce15cf8d090eb3e3aa9bdd756b6440324b8ccef44e90fcd9324895f015193f176766477ce3f9708705f0a5c65709006be15bee5e73fb9942eab700e3c9b1ff22d535ee69e6a4680c", "cb5777cd6d1a2133d9e2346b3c4ace0c527402a317c5"},
		{"f8b180808002808080b8a8b4f596bf063e50bb49457dd7bc8045b51f18c9a9b9afb182a751680ca4eb6af5569827d2aedd31afa94827174a2baff9f1bdc462c6735f7a362ee85933cbb5337812c67bb8ee393d63ffc79b31e687458436cc2311c3b02ffde68abac6a35d8c269f69b2abfad2e093c026d80f65d4032cea82da9b70bfdee04afed2dc7885f64ec6bfb666474e5b695f55ea901e9868d9cdcc240366d699417d7b582dc65a0a3a27c5c7f8dccbab", "cb6060f68800aa777981c3b512b67c8a11ea25c53f7b"},
	} {
		signer := NewNucleusSigner(params.TestChainConfig.ChainID)
		var tx *Transaction
		err := rlp.DecodeBytes(common.Hex2Bytes(test.txRlp), &tx)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		from, err := Sender(signer, tx)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		addr, _ := common.HexToAddress(test.addr)
		if from != addr {
			t.Errorf("%d: expected %x got %x", i, addr, from)
		}
	}
}
