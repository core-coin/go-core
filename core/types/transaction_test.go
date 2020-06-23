// Copyright 2014 The go-core Authors
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
	"bytes"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/core-coin/eddsa"
	"github.com/core-coin/go-core/params"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/rlp"
)

// The values in those tests are from the Transaction Tests
// at github.com/core-coin/tests.
var (
	key, _ = crypto.HexToEDDSA("2da94fd47e8369ffe88850654de266727ff284c3f78d61b04153cb9a908ed3b61248ac5172d3caabbc3493807c0297645ae328e10eb9543bdbcc413b5871d83426cd5b3a0083e6f589f60c1177b287b8f4f764acfcba7dfceadc51ef37b40d6182e3fe6bce148c8a48e07379754ebbf83236643837663566326266393833356639613437656566616535373162633039")

	emptyTx = NewTransaction(
		0,
		common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
		big.NewInt(0), 0, big.NewInt(0),
		nil,
	)

	rightvrsTx, _ = NewTransaction(
		0,
		crypto.PubkeyToAddress(key.PublicKey),
		big.NewInt(10),
		50000,
		big.NewInt(10),
		nil,
	).WithSignature(NewNucleusSigner(params.AllCryptoreProtocolChanges.ChainID),
		common.Hex2Bytes("48e82f6b21d0d2ec0a71c6fbdb2ce2dd25ecf9c4a5c30de0bad198bcffebab0d1f77b7cbcff93aecfd552183ff2518f7ee06e96afcc9dd8ec413f27297fd7d5ae1c32c31da707bc40e05c048a4b6b81e0e0c5b6534009035f0f6d7be955a416a9d877189dbaa1365f18dc20a58ec9a30dbcc413b5871d83426cd5b3a0083e6f589f60c1177b287b8f4f764acfcba7dfceadc51ef37b40d6182e3fe6bce148c8a48e07379754ebbf8"),
	)
)

func TestTransactionSigHash(t *testing.T) {
	var nucleus = NewNucleusSigner(params.AllCryptoreProtocolChanges.ChainID)
	if nucleus.Hash(emptyTx) != common.HexToHash("0xacb6be0ac9d62bb0bcbd2d3cb3fcb0fc35f750ca1f37d37e0bec0bbd1ddfb52f") {
		t.Errorf("empty transaction hash mismatch, got %x", emptyTx.Hash())
	}
	if nucleus.Hash(rightvrsTx) != common.HexToHash("0xcf01db304a3550e823e6f69a768eb75442271e363cfea7d41732b0bdb8f9f6e1") {
		t.Errorf("RightVRS transaction hash mismatch, got %x", rightvrsTx.Hash())
	}
}

func TestTransactionEncode(t *testing.T) {
	txb, err := rlp.EncodeToBytes(rightvrsTx)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	should := common.FromHex("f1800a82c35094f8989a12c39d55e40ca9466fbc3963d9914a44880a8094f8989a12c39d55e40ca9466fbc3963d9914a4488")
	if !bytes.Equal(txb, should) {
		t.Errorf("encoded RLP mismatch, got %x", txb)
	}
}

func decodeTx(data []byte) (*Transaction, error) {
	var tx Transaction
	t, err := &tx, rlp.Decode(bytes.NewReader(data), &tx)

	return t, err
}

func defaultTestKey() (*eddsa.PrivateKey, common.Address) {
	key, _ := crypto.HexToEDDSA("07e988804055546babfb00e34d015314a21a76a1cb049cad4adeb3d931af355f2393ba45bfda9aeb7ca40c1e0a4e63ba4639e43957a54109f2bcef60235cab768f2c3f8f301e8cfee41e04f26d8d01c46a10b08dd3d27c61ca48dfc7433d2d5e3bf60b862cedd3197e82c0b709c75c47ced2896631075043550b8d6b0cfb0ec165d178df945ff8038f30c9ada2e7a69e")
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return key, addr
}

func TestRecipientEmpty(t *testing.T) {
	_, addr := defaultTestKey()
	tx, err := decodeTx(common.Hex2Bytes("db80808080800194858a65a40fa13231ba88c574db2c9539124e6e1c"))
	if err != nil {
		t.Fatal(err)
	}

	from, err := Sender(NewNucleusSigner(params.AllCryptoreProtocolChanges.ChainID), tx)
	if err != nil {
		t.Fatal(err)
	}
	if addr != from {
		t.Fatal("derived address doesn't match")
	}
}

func TestRecipientNormal(t *testing.T) {
	_, addr := defaultTestKey()

	tx, err := decodeTx(common.Hex2Bytes("ef808080940000000000000000000000000000000000000000800194858a65a40fa13231ba88c574db2c9539124e6e1c"))
	if err != nil {
		t.Fatal(err)
	}

	from, err := Sender(NewNucleusSigner(params.AllCryptoreProtocolChanges.ChainID), tx)
	if err != nil {
		t.Fatal(err)
	}
	if addr != from {
		t.Fatal("derived address doesn't match")
	}
}

// Tests that transactions can be correctly sorted according to their price in
// decreasing order, but at the same time with increasing nonces when issued by
// the same account.
func TestTransactionPriceNonceSort(t *testing.T) {
	// Generate a batch of accounts to start with
	keys := make([]*eddsa.PrivateKey, 25)
	for i := 0; i < len(keys); i++ {
		keys[i], _ = crypto.GenerateKey(rand.Reader)
	}

	signer := NewNucleusSigner(params.AllCryptoreProtocolChanges.ChainID)
	// Generate a batch of transactions with overlapping values, but shifted nonces
	groups := map[common.Address]Transactions{}
	for start, key := range keys {
		addr := crypto.PubkeyToAddress(key.PublicKey)
		for i := 0; i < 25; i++ {
			tx, _ := SignTx(NewTransaction(uint64(start+i), common.Address{}, big.NewInt(100), 100, big.NewInt(int64(start+i)), nil), signer, key)
			groups[addr] = append(groups[addr], tx)
		}
	}
	// Sort the transactions and cross check the nonce ordering
	txset := NewTransactionsByPriceAndNonce(signer, groups)

	txs := Transactions{}
	for tx := txset.Peek(); tx != nil; tx = txset.Peek() {
		txs = append(txs, tx)
		txset.Shift()
	}
	if len(txs) != 25*25 {
		t.Errorf("expected %d transactions, found %d", 25*25, len(txs))
	}
	for i, txi := range txs {
		fromi, _ := Sender(signer, txi)

		// Make sure the nonce order is valid
		for j, txj := range txs[i+1:] {
			fromj, _ := Sender(signer, txj)

			if fromi == fromj && txi.Nonce() > txj.Nonce() {
				t.Errorf("invalid nonce ordering: tx #%d (A=%x N=%v) < tx #%d (A=%x N=%v)", i, fromi[:4], txi.Nonce(), i+j, fromj[:4], txj.Nonce())
			}
		}

		// If the next tx has different from account, the price must be lower than the current one
		if i+1 < len(txs) {
			next := txs[i+1]
			fromNext, _ := Sender(signer, next)
			if fromi != fromNext && txi.EnergyPrice().Cmp(next.EnergyPrice()) < 0 {
				t.Errorf("invalid energyprice ordering: tx #%d (A=%x P=%v) < tx #%d (A=%x P=%v)", i, fromi[:4], txi.EnergyPrice(), i+1, fromNext[:4], next.EnergyPrice())
			}
		}
	}
}

// TestTransactionJSON tests serializing/de-serializing to/from JSON.
func TestTransactionJSON(t *testing.T) {
	t.Skip()
	key, err := crypto.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("could not generate key: %v", err)
	}
	signer := NewNucleusSigner(common.Big1)

	transactions := make([]*Transaction, 0, 50)
	for i := uint64(0); i < 25; i++ {
		var tx *Transaction
		switch i % 2 {
		case 0:
			tx = NewTransaction(i, common.Address{1}, common.Big0, 1, common.Big2, []byte("abcdef"))
		case 1:
			tx = NewContractCreation(i, common.Big0, 1, common.Big2, []byte("abcdef"))
		}
		transactions = append(transactions, tx)

		signedTx, err := SignTx(tx, signer, key)
		if err != nil {
			t.Fatalf("could not sign transaction: %v", err)
		}

		transactions = append(transactions, signedTx)
	}

	for _, tx := range transactions {
		data, err := json.Marshal(tx)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}

		var parsedTx *Transaction
		if err := json.Unmarshal(data, &parsedTx); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}

		// compare nonce, price, energylimit, recipient, amount, payload, V, R, S
		if tx.Hash() != parsedTx.Hash() {
			t.Errorf("parsed tx differs from original tx, want %v, got %v", tx, parsedTx)
		}
	}
}
