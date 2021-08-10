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
	"crypto/rand"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/core-coin/go-core/params"
	eddsa "github.com/core-coin/go-goldilocks"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/rlp"
)

// The values in those tests are from the Transaction Tests
// at github.com/core-coin/tests.
var (
	address, errAddr = common.HexToAddress("cb08095e7baea6a6c7c4c2dfeb977efac326af552d87")
	key, _           = crypto.HexToEDDSA("2da94fd47e8369ffe88850654de266727ff284c3f78d61b04153cb9a908ed3b61248ac5172d3caabbc3493807c0297645ae328e10eb9543bdb")

	emptyTx = NewTransaction(
		0,
		address,
		big.NewInt(0), 0, big.NewInt(0),
		nil,
	)
	pub           = eddsa.Ed448DerivePublicKey(*key)
	rightvrsTx, _ = NewTransaction(
		0,
		crypto.PubkeyToAddress(pub),
		big.NewInt(10),
		50000,
		big.NewInt(10),
		common.FromHex("1123"),
	).WithSignature(
		NewNucleusSigner(params.AllCryptoreProtocolChanges.NetworkID),
		common.Hex2Bytes("b7ea2c0222ad2cf32dc5671dcff5b6d3190d328b2696cca92b5b17d56b76fb26b6e3e303f3d1c428828b0e86616f783b4fc0dcc9d60157f820d2ce5b6b2709734fcf4188bc1020db6e8b972c63531832d0ee4fd66a1c2cee858540e237ff4351d8b2ed0c08edf15a9851cfd532191c39825e4ad8d405878998a67c949b3f94e3445f1d6e61f69d091be53f4326eb9d9d05627375ef943ae9f1763689984aa377ed84cd8923973ab0"),
	)
)

func TestTransactionSigHash(t *testing.T) {
	if errAddr != nil {
		t.Error(errAddr)
	}
	var nucleus = NewNucleusSigner(params.AllCryptoreProtocolChanges.NetworkID)
	if nucleus.Hash(emptyTx) != common.HexToHash("0x0064d7a2aa08686b4f36a2188352ba162ff2b5bdce72335f4a0e25a6c5f47af7") {
		t.Errorf("empty transaction hash mismatch, got %x", emptyTx.Hash())
	}
	if nucleus.Hash(rightvrsTx) != common.HexToHash("0xc5606a8629133547d2fe010fd1f2d2a144704e207e9eb35add40e49dac635300") {
		t.Errorf("RightVRS transaction hash mismatch, got %x", rightvrsTx.Hash())
	}
}

func TestTransactionEncode(t *testing.T) {
	txb, err := rlp.EncodeToBytes(rightvrsTx)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	should := common.FromHex("f8cb800a82c3500196cb4896f17ca7f974ce8b1d596292c83ba7f38257101b0a821123b8a8b7ea2c0222ad2cf32dc5671dcff5b6d3190d328b2696cca92b5b17d56b76fb26b6e3e303f3d1c428828b0e86616f783b4fc0dcc9d60157f820d2ce5b6b2709734fcf4188bc1020db6e8b972c63531832d0ee4fd66a1c2cee858540e237ff4351d8b2ed0c08edf15a9851cfd532191c39825e4ad8d405878998a67c949b3f94e3445f1d6e61f69d091be53f4326eb9d9d05627375ef943ae9f1763689984aa377ed84cd8923973ab0")
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
	tx, err := decodeTx(common.Hex2Bytes("f8b480808002808080b8ab96244b19508ebae6a6e0784798a7133c1a450f67ebc7261639814a7564a44770487e1d0152f9bc122624c9f47cf268a48eb73bedc4bda99f00151fc77aac317bf4f03f3ed376950e6376fc1eced1c16c433e2466e1484f08ecb34369959a561a4cf7767a93a10b0ff52fbce0cca5acfc2600661c16da7981ec4c3007bde037c36a426ae7171c2a2403f1d20f5ce999f869271de6d65af86ad72a5ada09b12ec60b154ff72a0ccec9ce2080"))
	if err != nil {
		t.Fatal(err)
	}
	from, err := Sender(NewNucleusSigner(params.TestChainConfig.NetworkID), tx)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := common.HexToAddress("cb76a631db606f1452ddc2432931d611f1d5b126f848")
	if err != nil {
		t.Fatal(err)
	}
	if expected != from {
		t.Fatal("derived address doesn't match")
	}
}

func TestRecipientNormal(t *testing.T) {
	tx, err := decodeTx(common.Hex2Bytes("f8b480808002808080b8abc9df7a9e4c69e12f72cc27060ae919a470a9d14ce9590a4c8477354e437cb5da3a72cb7f9c398f30ff2cb32ea47bd4fa44f8e2369674d2b70070d5ed40d4d2125a518fd1f345f263a65f0aa49db7efb70d0c5b7878f877b4ed036c2749507fe66824e15f0b11c24a690f8cac25cb4b2101009f5bdee59b65cf8fc8d088814cb08ab501c8e53a1878f5b781402e3bdbe4cf47c989286ff3d568d9007200a47bf79b5eedb937688214096500"))
	if err != nil {
		t.Fatal(err)
	}

	from, err := Sender(NewNucleusSigner(params.TestChainConfig.NetworkID), tx)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := common.HexToAddress("cb74992308477b09a42fdd8051fd1bfd325bd7c0d16a")
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
	signer := NewNucleusSigner(params.AllCryptoreProtocolChanges.NetworkID)

	// Generate a batch of transactions with overlapping values, but shifted nonces
	groups := map[common.Address]Transactions{}
	for start, key := range keys {
		pub := eddsa.Ed448DerivePublicKey(*key)
		addr := crypto.PubkeyToAddress(pub)
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

// Tests that if multiple transactions have the same price, the ones seen earlier
// are prioritized to avoid network spam attacks aiming for a specific ordering.
func TestTransactionTimeSort(t *testing.T) {
	// Generate a batch of accounts to start with
	keys := make([]*eddsa.PrivateKey, 5)
	for i := 0; i < len(keys); i++ {
		keys[i], _ = crypto.GenerateKey(rand.Reader)
	}
	signer := NewNucleusSigner(params.AllCryptoreProtocolChanges.NetworkID)

	// Generate a batch of transactions with overlapping prices, but different creation times
	groups := map[common.Address]Transactions{}
	for start, key := range keys {
		pub := eddsa.Ed448DerivePublicKey(*key)
		addr := crypto.PubkeyToAddress(pub)

		tx, _ := SignTx(NewTransaction(0, common.Address{}, big.NewInt(100), 100, big.NewInt(1), nil), signer, key)
		tx.time = time.Unix(0, int64(len(keys)-start))

		groups[addr] = append(groups[addr], tx)
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
				t.Errorf("invalid gasprice ordering: tx #%d (A=%x P=%v) < tx #%d (A=%x P=%v)", i, fromi[:4], txi.EnergyPrice(), i+1, fromNext[:4], next.EnergyPrice())
			}
			// Make sure time order is ascending if the txs have the same gas price
			if txi.EnergyPrice().Cmp(next.EnergyPrice()) == 0 && txi.time.After(next.time) {
				t.Errorf("invalid received time ordering: tx #%d (A=%x T=%v) > tx #%d (A=%x T=%v)", i, fromi[:4], txi.time, i+1, fromNext[:4], next.time)
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
