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
	"github.com/core-coin/ed448"
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/rlp"
	"math/big"
	"testing"

	"github.com/core-coin/go-core/params"

	"github.com/core-coin/go-core/crypto"
)

func TestNucleusSigning(t *testing.T) {
	key, _ := crypto.GenerateKey(rand.Reader)
	pub := ed448.Ed448DerivePublicKey(key)
	addr := crypto.PubkeyToAddress(pub)

	signer := NewNucleusSigner(params.AllCryptoreProtocolChanges.NetworkID)
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
		{"f8b480808002808080b8ab409545554f2140fb8809caa989416735d6f84f0ee343c5f7a5411347b270e123b45978fa23b21480d3c23607aaf97f49ecd12d44633c7fa300b6d977bba893344e5ae550715d38ba6a10510caed5f1acd71b59b222f27ccca272a1ba05cf3cbd1a0bd559d453f73d4740607b720286651800e1373d7335038a2655c9b2a295e6fc44166137e54c1a3774ef033b6f72acf778b68956199ceb45f7b5b07062c676ded8cc1adf95140ba94200", "cb979f85b68cc4365531562bbfe8594d13221d6677b9"},
		{"f8b480808002808080b8aba738b17be98e7806910ec817283119aacfdfc39cc97fbdd9d6a0ab1ab4fd47f11e411925de751ae4c029ba46b2e4da5047fe9fde1e7b9f70808b3f92314f43a6028d0a9f4500b2a2597689e0f927758283382c322c747a47ff91d933323cbcf96895c194796b54a94ba483fb75ce0b523c006a92bafe431dd06988c2bf90ab1b9d01fdf550578ca945aa9c51157120aa128616ef8d661000ceded82e8ea6d229a6cfa01e8955bd0cf4ff80", "cb9169fd913782244a964124dfa86f65c084fca25f4a"},
		{"f8b480808002808080b8abbb15dbe9da1093ab22d5df5f1c8ee727dc36ede5d8093a51b506febbc2fb79faecd7dd82cd0f95bdab442bd01761b68841b3c512019ee4b18006ff010112ee439bd155c11173c2167c6f77ca35c25bd733c334ddfc36cc2ed906a032e570841a8e4db75bf236a3f955b9e7ef180e083c050014cfc2849ba1582c015a1ef59eb9f645b15f5f18442c5b57f2f7adf3f8907b7418e8b76dcacd132a2aea12dcef82d7ac27edbf82c80ecd3900", "cb585688591fcc640af17b49e8345801f48c270c761d"},
		{"f8b480808002808080b8ab60f56a4901d66fa655298664fdf9c2b69ca176395dd9fe02a963fa58633087dc2219a90d604568ec329b3b65a1855bd59b41b07eab7027e3806c8c8553deab337c2a78d58b92c8a30250ffa82541be21dc6011d04f0ea006f9a5cd9baf15e94ebc192fa0e7682e6f087f0679a6f403f32e00109ae3b3a330005c6227e6b499a62265cb4c4ed3d2eee554352a3039fe378a6b114d1e6c73d69e5840b2a7563815c61689c902fe583aa47980", "cb787b9cca8e2c23de58ba6adf9240ad52ed8e6042a1"},
		{"f8b480808002808080b8abe2a4d01d8500b6d688476754cfdabcea043118d1cbefd946aaf9a7271a7270e61a6be2b38ad5df69a11f86cec496cfdfb95730c4eb92e8e6809df5a365238838a843514c217a2ad1a526d1eca204e7e59260f2e5eaa460924423b1fa574e1d238f50084628e6e68a33908deea19447930d00f16831cc188e540fbf240fdbbb958e288daed3dcbc7337e9dbb5ac10e386212e1c029da6f7a9f5352da9389163fdfcf6411f1af17ec0e18280", "cb03fb18ef1250b2987fb83f401d732acf28955e6238"},
		{"f8b480808002808080b8ab617295f8046ddea94bee347042cd793ccc191a31173f40e40a2b1dffd544f4f733492c4696589fc8841bf2db0964799aa46003f66fefa380809b1c576db447ab01e0e4cc3c867a5f1e7b42adbb2e1fe4cb22daaaa1a19b5b93616e4824faa48e992ba0f750215f6a40a513bd5304a2561900531d42aa0cc157c50164b5d4ac9d0845e2cc8a99713eb523874ca8b36708f50bd76df6e9887ba02333bb623eda938d9d4481fcb232a870e780", "cb4036d36dd98cdaa7c41de2de40d40548a4abce464a"},
		{"f8b480808002808080b8abbf235d17bc989afe7e97e8aaf72b8a234c35a876620b0b732b1845c2102b5f5ad6f39900f7834a599cd38d40f6b04374bba03700a0ac939600509e2a427ea3e81be4f599f2707d41e5f97926a03ae6a2484edd1f8bac35613d506ffbe775f7336683cf43f510dbd6cb3dd04261c5562f1000ab82a50e4a370cd44243e5e6bacde8fa22330e12c2f04511433edd09d2cef30bcaaa01c6142467d4bd65fa9c0e8560d281563bc49035e98400", "cb091e23d0b9f7912466fa079ed1f9439b34cd40f448"},
		{"f8b480808002808080b8abd0b443c1fa4a3458b1ac49d098ba0b6fb8dcc6cbe04a97bfb19a81793407b450b35ca98f884a1d30bdc474cb8cf9683b2cbce26f58f58861804314ba94db58aa589bdc9811571bdbd7c25c7eb131c4ceb7570fcc25388614a608dd3315c2e2672a2edd3d8bd4ac968ce1b164246336301a00fd65f6b5eed6dbade7e3a157357b5c1fd5610c8a6fdca21d8e01860fea002b7bbb14c6574da6950a50dd2883e18a07cbdcd258627fa29caa00", "cb06b92cb364f7e08dfcf8c1010de3361aaf00ff8f30"},
		{"f8b480808002808080b8ab7278c118edb600e7839377fd54a49411f4edbd93f1ac70dcab877ed16d5b923172fba2ce14bd86738fa9a6a74880e25166f528747b8607e9002293a85b47e914ba84578e8ec546fa59cac7f54db8c9e02a4bef3cfdc9a770d89d58327cf5d32e42d8ec337ac7f7a42d23e69ca6fb8476240086cc46c7d4a48c0c1297f265cd18320c9f347d259d6c6e4d7361d4e40d556a90dc6e176c528ada874f3f2c97c5a67ae058d86f521714e70300", "cb9145a5c8329ff959edbec447423eabf596042f6809"},
		{"f8b480808002808080b8abe305e6c12e6341bf4ada4ea59250c5e9e62152dd691e8ec80bf21956126ccf2403c3d82f9e0e21e61df86d0d3e19e0a309431fd1b4c2e615807744bcdd444308ec24578a81cf44d55672c6756b0d0b2d11250750f5e800fc6d521cb17129d093500582716e640ef5f56a06af5b5c610b35007382d3be65302784c087d9733b2fe7e600b4cfdc25819e3f1ff09cb693b7f0f40a7af12685c14866eda8eeab33b56404ff590c2c6699913400", "cb329269856d0826f232333f9b4f93ea4283a647894c"},
	} {
		signer := NewNucleusSigner(params.TestChainConfig.NetworkID)
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
