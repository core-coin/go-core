// Copyright 2016 The go-core Authors
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

	"github.com/core-coin/go-core/params"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/rlp"
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
		{"f8cf098504a817c80983033450019435353535353535353535353535353535353535358202d980b8a80a7a63489f5629e9175e85fc8c4236cf2ef4b281db9986e8d7fdaee913efe67bd15a4b8d50472653e2713d007a3ef547d30d60fc6885e035f332b6468d4f742942a1f5d4cc1130b13f21181ebd004a5de130e0444eef3b721bfc249048e1f354bf2e1b069182bada2ca32906549067373dd22759044d65ceb6e2326e6590c8568bfd34edf2c4a83e9bb0d6b6a1a3d4d3193c783e487a7d218b7c609f6bde76b6919a20d02fa93551", "0xA6Bc6603C0c124694A7CaA8C7163724c22dc70EE"},
		{"f8cf088504a817c8088302e2480194353535353535353535353535353535353535353582020080b8a89e2c4c1ca6aad38072c64b13c2c5da4c5fa671a2186990bbe35bc052ea6bd8e8d75463411c2ccbfddf86d2a4df52fd40dfed62eb13ab32fb363ef5ab7dfe0bb307642eedaf9e3f8e37b5fab94a8c9d09aa007d4514caa53ff4ad5a228ecd2e88a72a8872409a6c44aca2e3b5bbc6cb2c17ac454401a42be7884a78c0a06e28dea79837cb5b143cf7b219e5809d573aec65128fee4217a26612353f9cc67e9a0d9109629375fe2d4d", "0x3A4179Af2B416991BB9120990531F5754F7DD228"},
		{"f8cf078504a817c807830290400194353535353535353535353535353535353535353582015780b8a82131c93431f0f0bdd8b4a32f927ec73f06ad1204f2cbff1203875a4cb6bf5e3d0133c2706e45c17b54651c028674f9cd7cfd7cfa5ff054ea96214d5093f0e3bf4314264cf221b3d86958dceb19ea181c39a36ce44bc5a82c2a6dfff0478b948a4dccbf494a1e9316fe4db7fee7b2802058fd718628e64c5f0e138656aecadfc57a1d6b735127b0bccdd8e5ff775cc72565b49b9613566acdecc2946bbf7cbcf52e697cdbdef3484e", "0x089200Aa2a542254e53814647E9fa2F8E03Ad44C"},
		{"f8ce068504a817c80683023e380194353535353535353535353535353535353535353581d880b8a8ab7ecb40909834a557e6a0cedc1f85b5a853f37199fbee57978cea4462369f8f503d53b43609f2c0ec384317062583dd84a544e7cee4a905275b6f1b8e2abe5de5934e2fab318b14bea9cb837a218c17017af032d99017bbc715104acb0d9b2ec4a1281498b06aadb2a9048799b61d36eb332a523a5df2efc844a0ef5e06084193047096053148beed0d5f550ef2dedd755a8384eaafa4213ee9015631d5c1a64ccdc3de68264238", "0x5112ced45570578c12ed27881A1F708630Ee105D"},
		{"f8cd058504a817c8058301ec30019435353535353535353535353535353535353535357d80b8a8bdf53b2183f70653721927a6ece9aaaad11b53356999f0186444bc6dc36c6de35ce221c3cb98aa7d7b18a4ff827e62475a9f63aaa72b79c5895f2b1b6d58b03e76907faa4034989323b72f987bf5b9b3b3920d784ac02556a77babe91849b1d01731337c575a127f97e56b2dcd687a340d31c4ae124294b1811cb2d147501b0d4d6ae3c26e95846a6373ed6cd7c65d6654b027330c361c2b2bdaa0775b93885e394db6dd888da5bb", "0x51477e3eE012a350D40dc2CDf8c1F84a80eA2a18"},
		{"f8cd048504a817c80483019a28019435353535353535353535353535353535353535354080b8a8309cd3dcdd01ec7bb41f0ea264ebd787036fb260a513192f76f6a68b25004192b12968d96f7e13fae2189a5d32fd22ae3560e74fb694422c8c9662d543004b75ae1ef45a84a0ed15c079fb0a8e3c446c49e53e3203c0ed076c7d2b0a4a93762848ee78d1bccc7dda3917d891f991713e3dc1289040d3dd428db9b332bdaafa47c4b982428bdd1fa0c2a0477426f496d1d9d3ceebc6f980aae8579065d6c39d03e02f27114d39fbae", "0xECD1880B77B453E236D87acD9F4bc6D8f47377da"},
		{"f8cd038504a817c80383014820019435353535353535353535353535353535353535351b80b8a8b1550c8cfcc07602ba8e479ecf7d3924230efec88c4531775aaa98fa2f85aa8ba611cf25da16b64368fd7d07c41e912dee58e27d25058334c829c55917cd8f51743dd0e9551a306f2839576e3070f16eebd013059201b4956b18fc939c314d572a2bc28a35b555e8dc41e37c72efdc12add69e9785be8858ce42b2c634556a5522e5f21a51763d4e274cec582f73a06bd1801653324459400ee3d11ec527af1e50bfefde83eed3de", "0xD73e20EAb43946DF1975D4f7c4f6e880532D25CA"},
		{"f8cc028504a817c80282f618019435353535353535353535353535353535353535350880b8a86c9ec1da66ba8617065c0b7bc9c0e0321d5663c53454c9de37d0059cd1b3f032a7ae4b9d4728d461d972bd61e93295c5b0cdbeb92aa4ea5d4bc729f215c6e367218603a013363f23477c19ab53a3e5158a543a88ecb0d51b5f1e3a1e31c1f52577f777c98e7a31fa6d7e35818094541a9177c477cb9d0c8aff86fd0c213d5b6baf2f302c9b32a9b1955ed7750ec895be454e8369e7b0d79f4bdca251a67ece830ef7ff01c3d5a96c", "0x7E365E2657a05AB015195dBeAD9f4f1E14349B77"},
		{"f8cc018504a817c80182a410019435353535353535353535353535353535353535350180b8a8d99395e10077903dc2df0448b9f99c1fe6d11157fa169d4093c75dfb841e037893a8260f50ed9407da910a1ef8823f27f4d98dfa630181b8ceb94375044399c8f76c8cc13c75bc029c81af11fe8f81af0ed2ea950f7de1b548b9d507a051e6338071bd6be042f8c384ac547a38e25e29f360fe65062218cc32df1739af72cb829b806f8ba2db0c6206bd37431ae0e129a662b09b44c85cd17fbb0bd265e26152ae82153f15ef2ff4", "0xb6d1Bf6647133C4E942d51bc843AEFA4a430CAB5"},
		{"f8cc808504a817c800825208019435353535353535353535353535353535353535358080b8a85f22778b13b8b63d89e930e9f60ef7694ded31cbad1f4044bd250a923359a3136af64c09f7b004128c8b804b254f3dfd91ebff688b1cedf394acd35cb50a305df99a512de9f9188298dcbe2ba3d78fe83a109bf8d4498e88fa2955b08dc4af6743886d9cf0b97cebb3a5d33f1001592755cc053ea1efefa0ac324f153468319f1ea696d7e29a5e6a728951dbb8ff8e0ae07a62805803e93339c75754e7eae0da3cc17893d8fe1785", "0xD911E544D390Bf6F3e7F2F31c2d857527DA03d84"},
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

		addr := common.HexToAddress(test.addr)
		if from != addr {
			t.Errorf("%d: expected %x got %x", i, addr, from)
		}

	}
}
