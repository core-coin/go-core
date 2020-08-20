// Copyright 2020 The CORE FOUNDATION, nadacia
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
	address, errAddr = common.HexToAddress("cb08095e7baea6a6c7c4c2dfeb977efac326af552d87")
	key, _           = crypto.HexToEDDSA("2da94fd47e8369ffe88850654de266727ff284c3f78d61b04153cb9a908ed3b61248ac5172d3caabbc3493807c0297645ae328e10eb9543bdbcc413b5871d83426cd5b3a0083e6f589f60c1177b287b8f4f764acfcba7dfceadc51ef37b40d6182e3fe6bce148c8a48e07379754ebbf83236643837663566326266393833356639613437656566616535373162633039")

	emptyTx = NewTransaction(
		0,
		address,
		big.NewInt(0), 0, big.NewInt(0),
		nil,
	)

	rightvrsTx, _ = NewTransaction(
		0,
		crypto.PubkeyToAddress(key.PublicKey),
		big.NewInt(10),
		50000,
		big.NewInt(10),
		common.FromHex("1123"),
	).WithSignature(
		NewNucleusSigner(params.AllCryptoreProtocolChanges.ChainID),
		common.Hex2Bytes("b7ea2c0222ad2cf32dc5671dcff5b6d3190d328b2696cca92b5b17d56b76fb26b6e3e303f3d1c428828b0e86616f783b4fc0dcc9d60157f820d2ce5b6b2709734fcf4188bc1020db6e8b972c63531832d0ee4fd66a1c2cee858540e237ff4351d8b2ed0c08edf15a9851cfd532191c39825e4ad8d405878998a67c949b3f94e3445f1d6e61f69d091be53f4326eb9d9d05627375ef943ae9f1763689984aa377ed84cd8923973ab0"),
	)
)

func TestTransactionSigHash(t *testing.T) {
	if errAddr != nil {
		t.Error(errAddr)
	}
	var nucleus = NewNucleusSigner(params.AllCryptoreProtocolChanges.ChainID)
	if nucleus.Hash(emptyTx) != common.HexToHash("0x0064d7a2aa08686b4f36a2188352ba162ff2b5bdce72335f4a0e25a6c5f47af7") {
		t.Errorf("empty transaction hash mismatch, got %x", emptyTx.Hash())
	}
	if nucleus.Hash(rightvrsTx) != common.HexToHash("0x600b4f389ba51f30ed9baae7b8880a283d5f460c88b32aafd9fbe98f42a23ff3") {
		t.Errorf("RightVRS transaction hash mismatch, got %x", rightvrsTx.Hash())
	}
}

func TestTransactionEncode(t *testing.T) {
	txb, err := rlp.EncodeToBytes(rightvrsTx)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	should := common.FromHex("f8cb800a82c3500196cb723ce833507cdfb66b76332523f034b38d8b077f560a821123b8a8b7ea2c0222ad2cf32dc5671dcff5b6d3190d328b2696cca92b5b17d56b76fb26b6e3e303f3d1c428828b0e86616f783b4fc0dcc9d60157f820d2ce5b6b2709734fcf4188bc1020db6e8b972c63531832d0ee4fd66a1c2cee858540e237ff4351d8b2ed0c08edf15a9851cfd532191c39825e4ad8d405878998a67c949b3f94e3445f1d6e61f69d091be53f4326eb9d9d05627375ef943ae9f1763689984aa377ed84cd8923973ab0")
	if !bytes.Equal(txb, should) {
		t.Errorf("encoded RLP mismatch, got %x", txb)
	}
}

func decodeTx(data []byte) (*Transaction, error) {
	var tx Transaction
	t, err := &tx, rlp.Decode(bytes.NewReader(data), &tx)

	return t, err
}

func TestRecipientEmpty(t *testing.T) {
	tx, err := decodeTx(common.Hex2Bytes("f8bb078504a817c80783029040808082015780b8a82131c93431f0f0bdd8b4a32f927ec73f06ad1204f2cbff1203875a4cb6bf5e3d0133c2706e45c17b54651c028674f9cd7cfd7cfa5ff054ea96214d5093f0e3bf4314264cf221b3d86958dceb19ea181c39a36ce44bc5a82c2a6dfff0478b948a4dccbf494a1e9316fe4db7fee7b2802058fd718628e64c5f0e138656aecadfc57a1d6b735127b0bccdd8e5ff775cc72565b49b9613566acdecc2946bbf7cbcf52e697cdbdef3484e"))
	if err != nil {
		t.Fatal(err)
	}
	from, err := Sender(NewNucleusSigner(params.TestChainConfig.ChainID), tx)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := common.HexToAddress("cb11973f51daa415dd5b1ce5cb84f8d20ad85f0e4c09")
	if err != nil {
		t.Fatal(err)
	}
	if expected != from {
		t.Fatal("derived address doesn't match")
	}
}

func TestRecipientNormal(t *testing.T) {
	tx, err := decodeTx(common.Hex2Bytes("f8bb078504a817c80783029040808082015780b8a82131c93431f0f0bdd8b4a32f927ec73f06ad1204f2cbff1203875a4cb6bf5e3d0133c2706e45c17b54651c028674f9cd7cfd7cfa5ff054ea96214d5093f0e3bf4314264cf221b3d86958dceb19ea181c39a36ce44bc5a82c2a6dfff0478b948a4dccbf494a1e9316fe4db7fee7b2802058fd718628e64c5f0e138656aecadfc57a1d6b735127b0bccdd8e5ff775cc72565b49b9613566acdecc2946bbf7cbcf52e697cdbdef3484e"))
	if err != nil {
		t.Fatal(err)
	}

	from, err := Sender(NewNucleusSigner(params.TestChainConfig.ChainID), tx)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := common.HexToAddress("cb11973f51daa415dd5b1ce5cb84f8d20ad85f0e4c09")
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
