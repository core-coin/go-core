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
	"github.com/core-coin/eddsa"
	"github.com/core-coin/go-core/params"
	"math/big"
	"testing"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/rlp"
)

// The values in those tests are from the Transaction Tests
// at github.com/core-coin/tests.
var (
	emptyTx = NewTransaction(
		0,
		common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
		big.NewInt(0), 0, big.NewInt(0),
		nil,
	)

	rightvrsTx, _ = NewTransaction(
		3,
		common.HexToAddress("0x1A1f598a1b3f1614C7c5F3AD27D0ef4875A874Ec"),
		big.NewInt(10),
		2000,
		big.NewInt(1),
		common.FromHex("1123"),
	).WithSignature(
		NewNucleusSigner(params.AllCryptoreProtocolChanges.ChainID),
		common.Hex2Bytes("b7ea2c0222ad2cf32dc5671dcff5b6d3190d328b2696cca92b5b17d56b76fb26b6e3e303f3d1c428828b0e86616f783b4fc0dcc9d60157f820d2ce5b6b2709734fcf4188bc1020db6e8b972c63531832d0ee4fd66a1c2cee858540e237ff4351d8b2ed0c08edf15a9851cfd532191c39825e4ad8d405878998a67c949b3f94e3445f1d6e61f69d091be53f4326eb9d9d05627375ef943ae9f1763689984aa377ed84cd8923973ab0"),
	)
)

func TestTransactionSigHash(t *testing.T) {
	var nucleus = NewNucleusSigner(params.AllCryptoreProtocolChanges.ChainID)
	if nucleus.Hash(emptyTx) != common.HexToHash("0xacb6be0ac9d62bb0bcbd2d3cb3fcb0fc35f750ca1f37d37e0bec0bbd1ddfb52f") {
		t.Errorf("empty transaction hash mismatch, got %x", emptyTx.Hash())
	}
	if nucleus.Hash(rightvrsTx) != common.HexToHash("0x034afa3236428cde891112ba48944cf6f35efe70a6c114ede1d12d9a55e0b3a7") {
		t.Errorf("RightVRS transaction hash mismatch, got %x", rightvrsTx.Hash())
	}
}

func TestTransactionEncode(t *testing.T) {
	txb, err := rlp.EncodeToBytes(rightvrsTx)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	should := common.FromHex("f603018207d0941a1f598a1b3f1614c7c5f3ad27d0ef4875a874ec0a82112394a0d85dc279e54182d97d4d3aa365445133a8d32d820539")
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
	tx, err := decodeTx(common.Hex2Bytes("de80808080800194858a65a40fa13231ba88c574db2c9539124e6e1c820539"))
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
	tx, err := decodeTx(common.Hex2Bytes("f2808080940000000000000000000000000000000000000000800194858a65a40fa13231ba88c574db2c9539124e6e1c820539"))
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
