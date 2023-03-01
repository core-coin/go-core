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
	crand "crypto/rand"
	"math/big"
	"testing"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/rlp"
)

func TestCIP155Signing(t *testing.T) {
	key, _ := crypto.GenerateKey(crand.Reader)

	signer := NewNucleusSigner(big.NewInt(18))
	tx, err := SignTx(NewTransaction(0, key.Address(), new(big.Int), 0, new(big.Int), nil), signer, key)
	if err != nil {
		t.Fatal(err)
	}

	from, err := Sender(signer, tx)
	if err != nil {
		t.Fatal(err)
	}
	if from != key.Address() {
		t.Errorf("exected from and address to be equal. Got %x want %x", from, addr)
	}
}

func TestCIP155NetworkId(t *testing.T) {
	key, _ := crypto.GenerateKey(crand.Reader)

	signer := NewNucleusSigner(big.NewInt(18))
	tx, err := SignTx(NewTransaction(0, key.Address(), new(big.Int), 0, new(big.Int), nil), signer, key)
	if err != nil {
		t.Fatal(err)
	}

	if int64(tx.NetworkID()) != signer.networkId.Int64() {
		t.Error("expected networkId to be", signer.networkId.Int64(), "got", tx.NetworkID())
	}

	tx2 := NewTransaction(0, key.Address(), new(big.Int), 0, new(big.Int), nil)
	tx2, err = SignTx(tx2, NewNucleusSigner(big.NewInt(1)), key)
	if err != nil {
		t.Fatal(err)
	}

	if tx2.NetworkID() == tx.NetworkID() {
		t.Error("expected to have different network ids", tx2.NetworkID(), tx.NetworkID())
	}

}

func TestCIP155SigningVitalik(t *testing.T) {
	for i, test := range []struct {
		txRlp, addr string
	}{
		{"f8e2800a8252080196cb2257ca89938de3a1eefdadb722c50a088d36c86aa48096cb55165569a4ac81c00fc3a871fd93be6b6a3b24726fb8ab278df725689a750012962d7cd9fb1cba321fa6e8346336208f9530d6b305f0bd80ae9d1b561873ca62523b5b53fcb51189942760aabf7c4e003830230aab4be6acf5e03cc2ac97a3b34b4debd425edf7adc2c945fd8a0955b0267f1e27c38a25e334e6415a1de48ca28bcac4b2a02bc031004a017069c1d1bcbe64d755f6cad5596603d9d32f1b81e121d9a189a51fc6abdc683471c4a1c90da33cd9e8169665b0e0c984552ff6d1618a00", "cb55165569a4ac81c00fc3a871fd93be6b6a3b24726f"},
		{"f8e5010a8252080196cb40c57cbdc8a522effbeeff97dabca24ff941cf6386830186a096cb06502f6ed135c78e84dc667881d812a737afbcea76b8ab14e2f11e35cb99bdc71004ac26b22929d6c172feb2249d4cd609a7a545176d75a054f9da13c865f114f20c41d3e0f36620d53afd367cb560003121cca2f0f830ce22d39cec3779385620ba8159800d188047a793ffdd1b46a202a6d3e98e35c38dd87d582281cdaa7b54851a61b94ed92c00c25381590fc8d73f777b0e43eec8f01a9257a480a77c47dad15621e1e32c9d8ddad0209fa08b3efb0f81dd320fb379d68d7b2f5084c3e99f00", "cb06502f6ed135c78e84dc667881d812a737afbcea76"},
		{"f8e5020a8252080196cb46ce8427f044c60c224912490fc806e8bbf2de601b83030d4096cb80e7f4d60575708248773eb28e179a264afb675dd6b8ab36be77331f9ba1937bd90021c18b66a6b2c14932c69fe1d95131159bbd24fa0bcd4916f8468ebfed647fea25cbdbb19a7fecc19d0d5896a880a2448f7b0549bf57661aedee72ce45a346da1328f64857c957070610933012db2580723a622c89d5d5fb54f28cfc1b931b28bb2268aaca100073db9591feda8365a412747b20c8d27ddbdc21e95840ebb8625b907ba23fec1f82861a14df6cbf0677b4fbaaf084aa90a50d8a19985fcda680", "cb80e7f4d60575708248773eb28e179a264afb675dd6"},
		{"f8e5030a8252080196cb503ab33df9618e8665b2753cdc6e9de9b1e3662792830493e096cb307ad47bba267d35a7d1ee38f33c00c89a0585db72b8ab9a8f745a39e062e0214a67a53f4e58711eda07be8d03cb8c78c14a199b240287c66bbb8b437fe75de79e18cdae2c296c00f2921e87da137980c1063ffcee267a275cb438418604d42673017bf6cb5a7cedbc37cd98698f9ad3033d6abed58a29b6b1fae2222ecd91c670e3d94ab116c82a0004c23752bc45407225421b6911ee1e30f1f29df3fa78ab5176c880549d1f09d3f290d196e1b6acaa4a197163e4f191c71f7fc89fbb6d6e7b00", "cb307ad47bba267d35a7d1ee38f33c00c89a0585db72"},
		{"f8e5040a8252080196cb578095b88b484f7600342a8b481034c0ab6a2eac9483061a8096cb46b5fee7025f6b4ff725c45ba793103673101c3ab2b8ab68c1ec85bd17af625e536f5dea32018beb1d6dffc2b44a5fcc353baa3337a4976f8c387b00c5ac1aac9e501d46bcbc351a8584e96b797b088046d85d4ccf4892e2f6e897ff50147a710efc2cfae6602fcf7ffd40fbd5197191458da13cc2ee5696d2ce66c915dbff8143e65db35e35db3500a7a2b34a5b91bf721de00286d4639e0c80a3c9d93f1bac76d5f31f245bcb46a0447dcffa51d8cb5d75d059c260c6141eff8cc795ae67295680", "cb46b5fee7025f6b4ff725c45ba793103673101c3ab2"},
		{"f8e5050a8252080196cb391d26127650505c63e51235b32cc730f6e637ecb98307a12096cb48af26a6c7ed230f22d58bf0b07cb44304e731d31ab8abfd04ab20bc153b8878036d4e9beba26d3c453e23309ffbe4b97e3520756e8667f2fd70dd097a926edc50bc63f3d10000f5bfcf806f29aa72806dee6e7236191af6bb06047234edbeb2373bc7f63157198fd7410fa5254f7f8a3294d55ee7faf8f68bf62ba873c75eb41fea3e8670543b3d0019bf027c59d5963dcb2ff1e23ef20e4d2a2583d861eeb3e28ae5b5e864a630344ed11b462ce333ad4fb51094cccd85e79c4c7047a99c4d6a80", "cb48af26a6c7ed230f22d58bf0b07cb44304e731d31a"},
		{"f8e5060a8252080196cb276e92afdc9a7c6d87070aaa634e5f92209e14e1f7830927c096cb427389eac646ced63a40301934ffef01f49a36fd94b8abcab49cc200077d1a2c296efea5a199355d877cee65ae15489f153971d886f5fbe9c96313735c86d3b538274f9d0fcd05b486c1905ebcc6c4008caadee6b01d7c305145d0a73803654e7e079eb709d73f56be1ee47ffa3cef61e61b675506ba09c72b71a1f2bcb90a1f73fb39cf86b1ec380044ca81deeab3c328f5d3453b39464bd1967978f3a51c87d3df5806a722ed226322dfd265387545e934938a23035680ee5315e98e6dd7f01300", "cb427389eac646ced63a40301934ffef01f49a36fd94"},
		{"f8e5070a8252080196cb38c5c358a2be815039974dc5d45bf92ca764cf9365830aae6096cb614221bcc1c658aacd9755acea25589ba8fcf37833b8ab7b4152a33bd71638f04233c02e03d55e0be9bdf9dc5ae2e875461a42952318b715c757544e50aba17bbe99f22c0c73356a8874bd35f65c688075c488934313ba0a7c7e8cc1ddf6dfa42a583af952776ad5079ac5f22b140e133b3df3db5d97776ccdb9b6b9fc51406c9d7bebfb8ef8f20c00ba8b8cc4d28bc21cbd2f2c8a88ee4b7b485c4b034a0efda33410a1e2a34bcd421eb98af616f57690f7cdac904d193607aa2f75222673357900", "cb614221bcc1c658aacd9755acea25589ba8fcf37833"},
		{"f8e5080a8252080196cb230436a2286eb1059a1d919a143141682a8c041d62830c350096cb23845cf69afc1be6762b819f5273dba4b39a570c35b8ab01e5da5e55a4230af41a2e7956018e7c5fcb487f403381c8425941c0e5d62a94fc8599b6acb6712a8ef050f37868e8a65857573bb025fcb1000ade51bcdc34b77471945354ad7538ce69b80d8dae8526e23c8d3fe9d2e3c0ac99032f7d01e895a91a84c33f72e4b103698bc2438a2d7b1300054143772fb8b47e5d6d44f5b78001cb4af0013d2130a52425407e70b1556f21bf2029cdc7cd455a46e624d0bcf5411a033e0f2824afee8000", "cb23845cf69afc1be6762b819f5273dba4b39a570c35"},
		{"f8e5090a8252080196cb322b2fa0a3d789ae8ba24fd1831d2805f58059574e830dbba096cb41a20ca121312823522cc6e831566c1e56ff2172ebb8ab85c8dd1dfa72ed4e39a29c2706c7241d1747200ce33228208b1e88ba1ef3508e5bcb4c486973a8af168b6268e7fe53e0b6b3f3284be1453900e898975aae1bb3000bd0fd25e9622e6d8dab4059cf453ec26e9b373bd1f481bca4361b518a217b138f64c1f2c3ac9fabcf3a358bdb0de92d00c14f47d93acc4f2ca35df2a35665233ad93978b50106ea0fc3ead64fe256af05c61587e45e26baf7c0bb67955eaba4109de1bf77050324d800", "cb41a20ca121312823522cc6e831566c1e56ff2172eb"},
	} {
		signer := NewNucleusSigner(big.NewInt(1))

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

func TestNetworkId(t *testing.T) {
	key, _ := defaultTestKey()

	tx := NewTransaction(0, common.Address{}, new(big.Int), 0, new(big.Int), nil)

	var err error
	tx, err = SignTx(tx, NewNucleusSigner(big.NewInt(1)), key)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Sender(NewNucleusSigner(big.NewInt(2)), tx)
	if err != ErrInvalidNetworkId {
		t.Error("expected error:", ErrInvalidNetworkId)
	}

	_, err = Sender(NewNucleusSigner(big.NewInt(1)), tx)
	if err != nil {
		t.Error("expected no error")
	}
}
