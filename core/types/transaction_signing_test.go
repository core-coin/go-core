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
		{"f8c68080800195927a8d07705e47bda1a5fbb769fb898b1455ef2ec68080b8a85b550dae30951a091f9b4d8212af2ca595aab87dff883940ebf3447f4c8f5e969e492ccfb456629bfe2a8ffb8e9806be557293a288cecf2d034e8b1994fdd6aa0d5984d7b46fcec71950e9914f5efe80cebddee5afc62ad1c94c24ed8929ad748ac2d61302f63b20595d9ff804673d267e0250080bd9a1a679b16b10f0825074abe6f78f11f592e1e78e18e0383c68cb239d899d38951f4772125a6c5e3da7c0427d12eb5414a095", "927a8d07705e47bda1a5fbb769fb898b1455ef2ec6"},
		{"f8c6808080019507a498bc6279f5c0a1fdb3a8cca81e2d6e03cd67e48080b8a8a304e46ebb3023a07cfc13081172978ee85b153ba9adb984d78c64edf6271995957d788fac6dbe9eb389d46b1e895f8b244766e0d0d5ee6cab937a0e2c28fb71c566d152e3453a5c4e0f07a8a47bbe55d095457640082277fd97ce43d6b38798cf3c7dee9b683a57723162f35fa6861990b32d225e138500649d57946571556d9ce4da42aae8f1e2e67788d9dcd1033ac90415edac143c64d3df3dda9e86411b55d64bb7eed60d52", "07a498bc6279f5c0a1fdb3a8cca81e2d6e03cd67e4"},
		{"f8c6808080019571474664fc9462eb77770e24933efe31873ca7b2c08080b8a85585d8ec8d78f6053ac596a2bb550024474457d5709f42a9599e3d634e799caa2887b810f13b3b56448f3fc992269ad6471150fe96be8161eee31d9fec3535d43505b86a3e6e1994b025d3379e5870b6e41cf7d63f4763028e82e1ba66770f5e6c372d4e71f9f07fd8ac1cc743c1c51f8e68969e6292f81af00f577d8b19334b2107875fafc3fe06fea5f7bb45f771f1700b6990cb2db5513d1c5f7548b0786750c926a82948e7a7", "71474664fc9462eb77770e24933efe31873ca7b2c0"},
		{"f8c68080800195043b384c74294e070e4386519d8182eca23da6db418080b8a80c73c1f0b7085b8a939f420b247ed08bc911e0179eb40251273e84f58536ab24fcae375a079642dce2e6c342c43e1e351c67312381f450adea37615356ad5927e27844db37d80f7a1273fd9f8a0e43e1c940e07507b5139dbff52d078f1e183cb4f9cdc4823c82f960b35797547b4d0a981ecad6ae671160d18372a02c111149a302b7543d12892d4433954cdae3f3505c8676af8f18fe14b5b849e47e62f29a4e276617019daf8a", "043b384c74294e070e4386519d8182eca23da6db41"},
		{"f8c68080800195160576e3772e57d5ebf18a7b22d20cdaf24b3b542f8080b8a8219115872aa69f31798c53ae312a96e758284f50108d42ed21b1bc6ec6bd43c2c9bd3040400b7096a0d5bb7ba2fef25a8e940ed017ca3911a39d58b012fcdf6428d03ae42f1142011ff89a8e9fa42f5faa64e27619e6d5b7bc7cb0d26af73b822b8abc34cf720811930bbfe05bcac421abb692fbb02f0cb2051c5d91020d99aaae4040fc601f32ca6ff29496b4db96d085f55f6e2c8e89fe6900f03d8290dc01998d8d9ac00c402f", "160576e3772e57d5ebf18a7b22d20cdaf24b3b542f"},
		{"f8c680808001951286ce8184d8f6e9dcd649830ce07ef7c42f56b97a8080b8a8dcfa851059280a1761e4eb3803d57916003e07218d1f1d90f144169e9ece1416aa72791d317313797ba4da3a5d602edf971ba2c56d226e8e212a874aac0802c3f762fee2c71dad2fa73c6defe90e911dfed18b554530bc40761624c2933b3a5acede86adef636bb4ae5e7e6f1ffe38313cf519e6ecc91002afd2af31fc1786867e2b9c77d8cee4b3017737b3ac430926ee0c02486469acbeb96dc9de32d20f339577b017eb54abd8", "1286ce8184d8f6e9dcd649830ce07ef7c42f56b97a"},
		{"f8c68080800195954755728390b83474ec0af138790320aafd33e7578080b8a82861d242d4badafebc6331ea0dc0407ac9ff6e579fc62a06a7f113dc413d0604693809e4547d5abbb2e2613b34280916c17ae85b331f3a3ff4efa41f96befd5dc49c77ea7ace4bb95f3c5cfeb517ee43178b9726c540f93c874a0ffda8045132b77ec35163c347090eb985b6da487a25445e93761088401f312bb83e30885162a4de2cc3d22d2d7a68a215da0835dcddbffae2fad5b3eb9f5db691ff99789b3fd98ceb6b8a9adf05", "954755728390b83474ec0af138790320aafd33e757"},
		{"f8c68080800195287f3ff2292b30d66cbcb5ae8c0e1e92e528e8bb168080b8a8f2b30ee61369895d9523289f4c30b7f0063fd96206d12bc456352acc1d96932b9f5af384f40b451603ff50578c734b2c9026d53f6578216c519e88eeabbc478e7b3da4dd6af8f53ca0c210beef30ae79c274bdaeb81468ce8eca0bb3ba81f2a2309b2799137df8eb8fab1ac3618c61017e70ee29192bbe1f769ec59184ee6fda2887484edd1608044c767d678c0884b16315967e5f70aa129a7914f077d1f1079b0f2445ff944271", "287f3ff2292b30d66cbcb5ae8c0e1e92e528e8bb16"},
		{"f8c6808080019520b4a7f4d4162f89ed0382f24df9b5b0e11d928aec8080b8a83e838b33677f2170c67ba9d79f43e4ab37deb690629df9f9aacd3e10db4ae8620a300af9cb15e9aa3e902bf2980b4f6298a0f56e3287af6b1391cf54a897e7aba0575b2b1ac2fd17c2c9b3ec10e1c711ba16a136d25a86abb7c0afdd125805ea234682f2f39b1565a801464bd3b5473b47a6bcd8dfa57770ca5a5cb704a757614be31cf54f9f24f9b78828f34399551418638c494e440191d4dcf4a7effff9178848e9ba15b2accb", "20b4a7f4d4162f89ed0382f24df9b5b0e11d928aec"},
		{"f8c6808080019547bdd419e718e88154b155b0c14bba72f55e2f090a8080b8a83e9992d93c81afc0828c2c6c4017d0dc6dc89e2f1de1373c3cba27cf2dad8ac6111dd12a9e5c94e8c45727c74d78865114691a6e0497e46a391ce0d562487762aa4caa4ccf9e79755d746a9f73f0e1c93f00bcfdbce626f8ca9cbbbf06624d99f8df222e59f82b8af92ab38500833a211b3f757ebae0c4c783b1dc6b140c0c5cb57882ea9ba78e129d24bb48d8259e3c707ac124cd26f1b1810ee872c386b97aa786e7e787ec4ef2", "47bdd419e718e88154b155b0c14bba72f55e2f090a"},
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

		addr, err := common.HexToAddress(test.addr)
		if err != nil {
			t.Error(err)
		}
		if from != addr {
			t.Errorf("%d: expected %x got %x", i, addr, from)
		}

	}
}
