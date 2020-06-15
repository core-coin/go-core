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
	"github.com/core-coin/go-core/params"
	"math/big"
	"testing"

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
		{"f839808504a817c800825208943535353535353535353535353535353535353535808094f0f6f18bca1b28cd68e4357452947e021241e9ce820539", "0xf0f6f18bca1b28cd68e4357452947e021241e9ce"},
		{"f839018504a817c80182a41094353535353535353535353535353535353535353501809423ef145a395ea3fa3deb533b8a9e1b4c6c25d112820539", "0x23ef145a395ea3fa3deb533b8a9e1b4c6c25d112"},
		{"f7028504a817c80282f6189435353535353535353535353535353535353535350880942e485e0c23b4c3c542628a5f672eeab0ad4888be01", "0x2e485e0c23b4c3c542628a5f672eeab0ad4888be"},
		{"f838038504a817c803830148209435353535353535353535353535353535353535351b809482a88539669a3fd524d669e858935de5e5410cf001", "0x82a88539669a3fd524d669e858935de5e5410cf0"},
		{"f838048504a817c80483019a28943535353535353535353535353535353535353535408094f9358f2538fd5ccfeb848b64a96b743fcc93055401", "0xf9358f2538fd5ccfeb848b64a96b743fcc930554"},
		{"f838058504a817c8058301ec309435353535353535353535353535353535353535357d8094a8f7aba377317440bc5b26198a363ad22af1f3a401", "0xa8f7aba377317440bc5b26198a363ad22af1f3a4"},
		{"f83b068504a817c80683023e3894353535353535353535353535353535353535353581d88094f1f571dc362a0e5b2696b8e775f8491d3e50de35820539", "0xf1f571dc362a0e5b2696b8e775f8491d3e50de35"},
		{"f83c078504a817c807830290409435353535353535353535353535353535353535358201578094d37922162ab7cea97c97a87551ed02c9a38b7332820539", "0xd37922162ab7cea97c97a87551ed02c9a38b7332"},
		{"f83c088504a817c8088302e24894353535353535353535353535353535353535353582020080949bddad43f934d313c2b79ca28a432dd2b7281029820539", "0x9bddad43f934d313c2b79ca28a432dd2b7281029"},
		{"f83c098504a817c809830334509435353535353535353535353535353535353535358202d980943c24d7329e92f84f08556ceb6df1cdb0104ca49f820539", "0x3c24d7329e92f84f08556ceb6df1cdb0104ca49f"},
	} {
		signer := NewNucleusSigner(params.AllCryptoreProtocolChanges.ChainID)
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
