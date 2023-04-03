// Copyright 2014 by the Authors
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
	crand "crypto/rand"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/params"
	"github.com/core-coin/go-core/v2/rlp"
)

// The values in those tests are from the Transaction Tests
// at github.com/core-coin/tests.
var (
	address, errAddr = common.HexToAddress("cb08095e7baea6a6c7c4c2dfeb977efac326af552d87")
	key, _           = crypto.UnmarshalPrivateKeyHex("2da94fd47e8369ffe88850654de266727ff284c3f78d61b04153cb9a908ed3b61248ac5172d3caabbc3493807c0297645ae328e10eb9543bdb")

	emptyTx = NewTransaction(
		0,
		address,
		big.NewInt(0), 0, big.NewInt(0),
		nil,
	)
	rightvrsTx, _ = NewTransaction(
		0,
		key.Address(),
		big.NewInt(10),
		50000,
		big.NewInt(10),
		common.FromHex("1123"),
	).WithSignature(
		NewNucleusSigner(params.MainnetChainConfig.NetworkID),
		common.Hex2Bytes("b7ea2c0222ad2cf32dc5671dcff5b6d3190d328b2696cca92b5b17d56b76fb26b6e3e303f3d1c428828b0e86616f783b4fc0dcc9d60157f820d2ce5b6b2709734fcf4188bc1020db6e8b972c63531832d0ee4fd66a1c2cee858540e237ff4351d8b2ed0c08edf15a9851cfd532191c39825e4ad8d405878998a67c949b3f94e3445f1d6e61f69d091be53f4326eb9d9d05627375ef943ae9f1763689984aa377ed84cd8923973ab0"),
	)
)

func TestTransactionSigHash(t *testing.T) {
	if errAddr != nil {
		t.Error(errAddr)
	}
	var nucleus = NewNucleusSigner(params.MainnetChainConfig.NetworkID)
	if nucleus.Hash(emptyTx) != common.HexToHash("0x0064d7a2aa08686b4f36a2188352ba162ff2b5bdce72335f4a0e25a6c5f47af7") {
		t.Errorf("empty transaction hash mismatch, got %x", emptyTx.Hash())
	}
	if nucleus.Hash(rightvrsTx) != common.HexToHash("0x4da44e1f55eb0aa1958ce6a47fdf503b9cc9421fb1c4312b9eba1a3c5587e647") {
		t.Errorf("RightVRS transaction hash mismatch, got %x", rightvrsTx.Hash())
	}
}

func TestTransactionEncode(t *testing.T) {
	txb, err := rlp.EncodeToBytes(rightvrsTx)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	should := common.FromHex("f8cb800a82c3500196cb8238748ee459bc0c1d86eab1d3f6d83bb433cdad9c0a821123b8a8b7ea2c0222ad2cf32dc5671dcff5b6d3190d328b2696cca92b5b17d56b76fb26b6e3e303f3d1c428828b0e86616f783b4fc0dcc9d60157f820d2ce5b6b2709734fcf4188bc1020db6e8b972c63531832d0ee4fd66a1c2cee858540e237ff4351d8b2ed0c08edf15a9851cfd532191c39825e4ad8d405878998a67c949b3f94e3445f1d6e61f69d091be53f4326eb9d9d05627375ef943ae9f1763689984aa377ed84cd8923973ab0")
	if !bytes.Equal(txb, should) {
		t.Errorf("encoded RLP mismatch, got %x", txb)
	}
}

func decodeTx(data []byte) (*Transaction, error) {
	var tx Transaction
	t, err := &tx, rlp.Decode(bytes.NewReader(data), &tx)

	return t, err
}

func defaultTestKey() (*crypto.PrivateKey, common.Address) {
	key, _ := crypto.UnmarshalPrivateKeyHex("2da94fd47e8369ffe88850654de266727ff284c3f78d61b04153cb9a908ed3b61248ac5172d3caabbc3493807c0297645ae328e10eb9543bdb")
	return key, key.Address()
}

func TestRecipientEmpty(t *testing.T) {
	tx, err := decodeTx(common.Hex2Bytes("f8b803018207d001800a825544b8ab850769578ccf6328306c6cc42661de2d2fa4679dce8e68901ab0920d5b0f81c7dd214a184c34f9b1293305f94a171015d470b6eacbac4de380b812810a9d6ab60e6668e01cb4d74d93ab0a25ffa55c8ef962902c8669bac1b64bdd124ecd5e00bdfad1e45c88a65e4d3a659be43fc4b33800ba277941fcb9ac8063a9b6ed64fbc86c51dd5ae6cf1f01f7bcf533cf0b0cfc5dc3fdc5bc7eaa99366ada5e7127331b862586a46c12a85f9580"))
	if err != nil {
		t.Fatal(err)
	}
	from, err := Sender(NewNucleusSigner(params.MainnetChainConfig.NetworkID), tx)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := common.HexToAddress("cb8238748ee459bc0c1d86eab1d3f6d83bb433cdad9c")
	if err != nil {
		t.Fatal(err)
	}
	if expected != from {
		t.Fatal("derived address doesn't match")
	}
}

func TestRecipientNormal(t *testing.T) {
	tx, err := decodeTx(common.Hex2Bytes("f8ce03018207d00196cb8238748ee459bc0c1d86eab1d3f6d83bb433cdad9c0a825544b8abe0b9531cf735fa66d60fb08231780a6de7de847a3d0de4eef012615f1194562e25a84aacfe3151cb39bdbabadf82f620b661f5deb100fe7d00bc2eaf3e7c47ed2e40e6e32f7f8172c4e27c1a561d4410abbc89b90f4a33c76b325fa625844b03a68538b4c870c3e617b5fc959ac7b8671500ba277941fcb9ac8063a9b6ed64fbc86c51dd5ae6cf1f01f7bcf533cf0b0cfc5dc3fdc5bc7eaa99366ada5e7127331b862586a46c12a85f9580"))
	if err != nil {
		t.Fatal(err)
	}

	from, err := Sender(NewNucleusSigner(params.MainnetChainConfig.NetworkID), tx)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := common.HexToAddress("cb8238748ee459bc0c1d86eab1d3f6d83bb433cdad9c")
	if err != nil {
		t.Fatal(err)
	}
	if expected != from {
		t.Fatal("derived address doesn't match")
	}
}

// Tests that transactions can be correctly sorted according to their price in
// decreasing order, but at the same time with increasing nonces when issued by
// the same account.
func TestTransactionPriceNonceSort(t *testing.T) {
	// Generate a batch of accounts to start with
	keys := make([]*crypto.PrivateKey, 25)
	for i := 0; i < len(keys); i++ {
		keys[i], _ = crypto.GenerateKey(crand.Reader)
	}
	signer := NewNucleusSigner(params.MainnetChainConfig.NetworkID)

	// Generate a batch of transactions with overlapping values, but shifted nonces
	groups := map[common.Address]Transactions{}
	for start, key := range keys {
		for i := 0; i < 25; i++ {
			tx, _ := SignTx(NewTransaction(uint64(start+i), common.Address{}, big.NewInt(100), 100, big.NewInt(int64(start+i)), nil), signer, key)
			groups[key.Address()] = append(groups[key.Address()], tx)
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

// Tests that if multiple transactions have the same price, the ones seen earlier
// are prioritized to avoid network spam attacks aiming for a specific ordering.
func TestTransactionTimeSort(t *testing.T) {
	// Generate a batch of accounts to start with
	keys := make([]*crypto.PrivateKey, 5)
	for i := 0; i < len(keys); i++ {
		keys[i], _ = crypto.GenerateKey(crand.Reader)
	}
	signer := NewNucleusSigner(params.MainnetChainConfig.NetworkID)

	// Generate a batch of transactions with overlapping prices, but different creation times
	groups := map[common.Address]Transactions{}
	for start, key := range keys {
		tx, _ := SignTx(NewTransaction(0, common.Address{}, big.NewInt(100), 100, big.NewInt(1), nil), signer, key)
		tx.time = time.Unix(0, int64(len(keys)-start))

		groups[key.Address()] = append(groups[key.Address()], tx)
	}
	// Sort the transactions and cross check the nonce ordering
	txset := NewTransactionsByPriceAndNonce(signer, groups)

	txs := Transactions{}
	for tx := txset.Peek(); tx != nil; tx = txset.Peek() {
		txs = append(txs, tx)
		txset.Shift()
	}
	if len(txs) != len(keys) {
		t.Errorf("expected %d transactions, found %d", len(keys), len(txs))
	}
	for i, txi := range txs {
		fromi, _ := Sender(signer, txi)
		if i+1 < len(txs) {
			next := txs[i+1]
			fromNext, _ := Sender(signer, next)

			if txi.EnergyPrice().Cmp(next.EnergyPrice()) < 0 {
				t.Errorf("invalid energyprice ordering: tx #%d (A=%x P=%v) < tx #%d (A=%x P=%v)", i, fromi[:4], txi.EnergyPrice(), i+1, fromNext[:4], next.EnergyPrice())
			}
			// Make sure time order is ascending if the txs have the same energy price
			if txi.EnergyPrice().Cmp(next.EnergyPrice()) == 0 && txi.time.After(next.time) {
				t.Errorf("invalid received time ordering: tx #%d (A=%x T=%v) > tx #%d (A=%x T=%v)", i, fromi[:4], txi.time, i+1, fromNext[:4], next.time)
			}
		}
	}
}

// TestTransactionJSON tests serializing/de-serializing to/from JSON.
func TestTransactionJSON(t *testing.T) {
	key, err := crypto.GenerateKey(crand.Reader)
	if err != nil {
		t.Fatalf("could not generate key: %v", err)
	}
	signer := NewNucleusSigner(common.Big1)

	transactions := make([]*Transaction, 0, 50)
	for i := uint64(0); i < 25; i++ {
		var tx *Transaction
		switch i % 2 {
		case 0:
			tx = NewTransaction(i, common.BigToAddress(common.Big1), common.Big0, 1, common.Big2, []byte("abcdef"))
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
		if tx.NetworkID() != tx.NetworkID() {
			t.Errorf("invalid chain id, want %d, got %d", tx.NetworkID(), parsedTx.NetworkID())
		}
	}
}
