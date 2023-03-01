// Copyright 2020 by the Authors
// This file is part of go-core.
//
// go-core is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-core is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-core. If not, see <http://www.gnu.org/licenses/>.

package xcbtest

import (
	"math/big"
	"time"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/internal/utesting"
)

// var faucetAddr = common.HexToAddress("0x71562b71999873DB5b286dF957af199Ec94617F7")
var faucetKey, _ = crypto.UnmarshalPrivateKeyHex("89bdfaa2b6f9c30b94ee98fec96c58ff8507fabf49d36a6267e6cb5516eaa2a9e854eccc041f9f67e109d0eb4f653586855355c5b2b87bb313")

func sendSuccessfulTx(t *utesting.T, s *Suite, tx *types.Transaction) {
	sendConn := s.setupConnection(t)
	t.Logf("sending tx: %v %v %v\n", tx.Hash().String(), tx.EnergyPrice(), tx.Energy())
	// Send the transaction
	if err := sendConn.Write(Transactions([]*types.Transaction{tx})); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	recvConn := s.setupConnection(t)
	// Wait for the transaction announcement
	switch msg := recvConn.ReadAndServe(s.chain, timeout).(type) {
	case *Transactions:
		recTxs := *msg
		if len(recTxs) < 1 {
			t.Fatalf("received transactions do not match send: %v", recTxs)
		}
		if tx.Hash() != recTxs[len(recTxs)-1].Hash() {
			t.Fatalf("received transactions do not match send: got %v want %v", recTxs, tx)
		}
	case *NewPooledTransactionHashes:
		txHashes := *msg
		if len(txHashes) < 1 {
			t.Fatalf("received transactions do not match send: %v", txHashes)
		}
		if tx.Hash() != txHashes[len(txHashes)-1] {
			t.Fatalf("wrong announcement received, wanted %v got %v", tx, txHashes)
		}
	default:
		t.Fatalf("unexpected message in sendSuccessfulTx: %s", pretty.Sdump(msg))
	}
}

func sendFailingTx(t *utesting.T, s *Suite, tx *types.Transaction) {
	sendConn, recvConn := s.setupConnection(t), s.setupConnection(t)
	// Wait for a transaction announcement
	switch msg := recvConn.ReadAndServe(s.chain, timeout).(type) {
	case *NewPooledTransactionHashes:
		break
	default:
		t.Logf("unexpected message, logging: %v", pretty.Sdump(msg))
	}
	// Send the transaction
	if err := sendConn.Write(Transactions([]*types.Transaction{tx})); err != nil {
		t.Fatal(err)
	}
	// Wait for another transaction announcement
	switch msg := recvConn.ReadAndServe(s.chain, timeout).(type) {
	case *Transactions:
		t.Fatalf("Received unexpected transaction announcement: %v", msg)
	case *NewPooledTransactionHashes:
		t.Fatalf("Received unexpected pooledTx announcement: %v", msg)
	case *Error:
		// Transaction should not be announced -> wait for timeout
		return
	default:
		t.Fatalf("unexpected message in sendFailingTx: %s", pretty.Sdump(msg))
	}
}

func unknownTx(t *utesting.T, s *Suite) *types.Transaction {
	tx := getNextTxFromChain(t, s)
	var to common.Address
	if tx.To() != nil {
		to = *tx.To()
	}
	txNew := types.NewTransaction(tx.Nonce()+1, to, tx.Value(), tx.Energy(), tx.EnergyPrice(), tx.Data())
	return signWithFaucet(t, txNew, s.chain.chainConfig.NetworkID)
}

func getNextTxFromChain(t *utesting.T, s *Suite) *types.Transaction {
	// Get a new transaction
	var tx *types.Transaction
	for _, blocks := range s.fullChain.blocks[s.chain.Len():] {
		txs := blocks.Transactions()
		if txs.Len() != 0 {
			tx = txs[0]
			break
		}
	}
	if tx == nil {
		t.Fatal("could not find transaction")
	}
	return tx
}

func getOldTxFromChain(t *utesting.T, s *Suite) *types.Transaction {
	var tx *types.Transaction
	for _, blocks := range s.fullChain.blocks[:s.chain.Len()-1] {
		txs := blocks.Transactions()
		if txs.Len() != 0 {
			tx = txs[0]
			break
		}
	}
	if tx == nil {
		t.Fatal("could not find transaction")
	}
	return tx
}

func invalidNonceTx(t *utesting.T, s *Suite) *types.Transaction {
	tx := getNextTxFromChain(t, s)
	var to common.Address
	if tx.To() != nil {
		to = *tx.To()
	}
	txNew := types.NewTransaction(tx.Nonce()-2, to, tx.Value(), tx.Energy(), tx.EnergyPrice(), tx.Data())
	return signWithFaucet(t, txNew, s.chain.chainConfig.NetworkID)
}

func hugeAmount(t *utesting.T, s *Suite) *types.Transaction {
	tx := getNextTxFromChain(t, s)
	amount := largeNumber(2)
	var to common.Address
	if tx.To() != nil {
		to = *tx.To()
	}
	txNew := types.NewTransaction(tx.Nonce(), to, amount, tx.Energy(), tx.EnergyPrice(), tx.Data())
	return signWithFaucet(t, txNew, s.chain.chainConfig.NetworkID)
}

func hugeEnergyPrice(t *utesting.T, s *Suite) *types.Transaction {
	tx := getNextTxFromChain(t, s)
	energyPrice := largeNumber(2)
	var to common.Address
	if tx.To() != nil {
		to = *tx.To()
	}
	txNew := types.NewTransaction(tx.Nonce(), to, tx.Value(), tx.Energy(), energyPrice, tx.Data())
	return signWithFaucet(t, txNew, s.chain.chainConfig.NetworkID)
}

func hugeData(t *utesting.T, s *Suite) *types.Transaction {
	tx := getNextTxFromChain(t, s)
	var to common.Address
	if tx.To() != nil {
		to = *tx.To()
	}
	txNew := types.NewTransaction(tx.Nonce(), to, tx.Value(), tx.Energy(), tx.EnergyPrice(), largeBuffer(2))
	return signWithFaucet(t, txNew, s.chain.chainConfig.NetworkID)
}

func signWithFaucet(t *utesting.T, tx *types.Transaction, networkID *big.Int) *types.Transaction {
	signer := types.NewNucleusSigner(networkID)
	signedTx, err := types.SignTx(tx, signer, faucetKey)
	if err != nil {
		t.Fatalf("could not sign tx: %v\n", err)
	}
	return signedTx
}
