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
		{"f8b180808002808080b8a8ead54829c6381a7d49572467b1ed0a69fa74814c47f827ff15bcb941fbcca2d1b1067bc66a136e2ffc024915fef6e5995da25742458814765d00461e4c35384c86bb1db16da79ecf93c7ee0a1f11161825ff2d7e0e48b11226120a7b2b425e86729bf61bb3f133942629994527b4102f8b0ffe571aab715f65e928fd6819923e7040db30f11db10170932f5a5fec0fbd574a7205f8cdf1d0196762ff1f8f8534bcd0752e5223ffa1", "cb0382be1b379199b2ef27ffcc887e08724b00e9869e"},
		{"f8b180808002808080b8a8b64ffc1dadb778b7d2792660e51c45b404ee0ecf7b21ca0ec2ae187fb03ea03e5f3de7bd838ec0bfec0ae86d6a7d24a6debcf0844f14ec6c7f58360deb5a387c571344c86427f7b8f9bda37dc7a9cadc76b6df9ce301b34ba02204db6d9cc7c714abb13c4c594bc1f08ada733208712cb508f9f47f22eac56445a6bc1436756e84d9124e4d6edc4512de455dfd909121ee2cb68799a3e293351cf596fd30d12686e8e9e308484e58", "cb253f3d22cbf2ef9901f2af8bd750c0155b529405e6"},
		{"f8b180808002808080b8a8c0304b8555eb5b33573c999ce7b93188001a06685b2be10d40dbc48be58fdd3377407f30944213e3888f13b826353bf589038517c60c1742e4244610b9f3cc1322cbf58aa766fa5520bf96021a7dab5c096557d853c24a3ca224a72a8701c86aea64e5295816ef49cef31728b91bd517335e314a824a2bb21fc74c6b58df82dbc81ecf275b59db21833e817687549d4fb743ca1c83b247433c5cb9a548152342d5c6a113dc3e1975", "cb34401123c3a2188ff239fd364a8aa7eb8eb0e7e685"},
		{"f8b180808002808080b8a83fd066ceaa989651103dbf6ce5c8ae917906683dea8ebe7cc94a31e75a0b37a851e470c0a5b38f2bfda31c71c00bfd5b4fd0ea0b28786f402b7cb105b25ddc581d37adf07395a765a0492913eb6c1587c86c23283f2f73e212e8287ce820d7846ecec266724d7e44012ad528e6c6ce09738ea3616c15c51db40686b7e6cbc125ac403d1268eccbffd701eae8bda24677916fcd57dda3dd41d526223745fd9a39f281042063e463d3", "cb57c2a67864901b9d24e302591aea8ab462107745d7"},
		{"f8b180808002808080b8a81182a03856b6e01b3596b85cdc772880c3bf15c41ed15d6539eacd789661f6d6068fac05a2eb712697fd56e68c10612ee7d5b9f9822d30eded3f3a087be7b3f5450a594ed213a3d2d68c2707bbbfab3d22e2d34931f70953537d2fe7e8a98537b930f580c6280d63a837893df40531080f724cfaef8a9acc451ca983a94b28398cc140357dd94dc014932f916e0e8a9785cf6854a70d018778a5d60f812b6f6968f5651152d41378", "cb025587a15f430b1d81f4e141d0a9e7b77cce5c8bfa"},
		{"f8b180808002808080b8a8920b4e882c97f783a4dec7dad39939294d0f43a8bf49ec36061c00a82d53fa99feaeffa883351db9935eee0fab0b4aa23fb183b672d39188518765683199bf2406d3fb922b6d16916ebe062fc51d15800e70f9ed14d4080bf932249b38f921bcbbe477daaf4c077a05ce41e57159de11a107ee40d9df20632246f720e8bd3b1eb9615e757aa686a98908b6b7016eaa717ed16ddb61a8d11128b810ce95f9ddb2bc7ed72584200a13", "cb53feabc0176fd0ae89ab12f32c7f993e72178cf8d9"},
		{"f8b180808002808080b8a810c62f45e025c77d3ad5edd1a2df59961991b9ab90e01aa823ca6c3419334371e01935d6428d51a01aaac2424df018896b86d4ca98cb8fa1e3dca96751f2e27a008e2308ed0b374a21328d366a705ebb44892e040ce0a27c34c382b5a3c08c278d5184fc8bb3525a18dd2189e4f3b81e49ec0775d53be25eee94c8db82516e46f7c6454e612603dd9c028ef5a1a5583c0a37e6bc88de76af536c0de3e9c8237f060175559a0045a0", "cb71e778054e5fd831eeab4a03a25f7c419834d979dd"},
		{"f8b180808002808080b8a89675b2c767563374a8f459b26bce5e059bd888ac074b43d702aaf0ace7db94270cc79cb2ab481f0051b964101b2b415b75f43212ac9f716131746a6f34064b05b27dc1f9bd275c5752f0cbae3ebb45365f8dd96bdea56801ea4263f3a3bf23982799f2c9764394aea58fe16dbbc28132ed23f0810406dbd4d2ec7719803e7aa3d17d88df0b34a34ea7dbe42d29b3729ae31408001820745ddf3c13284b517c489e46ec54a6e2e847", "cb696dbda0e93be5d2f91b860be373313a009a55b341"},
		{"f8b180808002808080b8a88dca2b3a7ba243490a91505e37c8fc9a0ac12f598dc295620492650b9df62943d77eb5eb083e8940c9595171fc0b3382f09524347aa33bdd16457fbd94d73b6b6be1d28a804c53cc4c3fece4e782155d7487fa5ad29a83ee99f8e768ef1caaca1c691a87f22069e84d24d58180e8b230101e2e63efeaf089481257758d5f826133d8a3daa7fc8c705c30aeff5eb9a8e7d99be06f45d075d1401434241c36ff6de405bb31910fbec8", "cb6639f355dc499aea7c3860204c860981fc4b6c1bd6"},
		{"f8b180808002808080b8a8be1551a61f7d7ffd23e5694f2c954917385a00288b470a890383bcd97f9aba49ea6b680afba6e956dab607cc0cb4410cd924aeb5769fcb1027a34e7b9093a20aa4f8d54aa09a08c9362072c1bffc1196c3c8e263afa602373f8116bfd39a4694f4ddf261ae104d19fabb2169f0beef3a958910d788c109965d6bfac100e4263378532ffc63a11e79e0d54e327de81e37ae03f39baf57f7c554ffb911b33898578dfeec87255385fd", "cb40e70ae1383fd3e3b9bdf0c5359a6b6b66b13c453c"},
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
