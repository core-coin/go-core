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
	"container/heap"
	"errors"
	"io"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/hexutil"
	"github.com/core-coin/go-core/rlp"
)

//go:generate gencodec -type txdata -field-override txdataMarshaling -out gen_tx_json.go

var (
	ErrInvalidSig = errors.New("invalid signature values")
)

type Transaction struct {
	data txdata    // Consensus contents of a transaction
	time time.Time // Time first seen locally (spam avoidance)

	// caches
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

type txdata struct {
	AccountNonce uint64          `json:"nonce"    gencodec:"required"`
	Price        *big.Int        `json:"energyPrice" gencodec:"required"`
	EnergyLimit  uint64          `json:"energy"      gencodec:"required"`
	NetworkID    uint            `json:"chain_id" gencodec:"required"`
	Recipient    *common.Address `json:"to"       rlp:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"    gencodec:"required"`
	Payload      []byte          `json:"input"    gencodec:"required"`
	Signature    []byte          `json:"signature"    gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

type txdataMarshaling struct {
	AccountNonce hexutil.Uint64
	Price        *hexutil.Big
	EnergyLimit  hexutil.Uint64
	NetworkID    hexutil.Uint64
	Signature    hexutil.Bytes
	Amount       *hexutil.Big
	Payload      hexutil.Bytes
}

func NewTransaction(nonce uint64, to common.Address, amount *big.Int, energyLimit uint64, energyPrice *big.Int, data []byte) *Transaction {
	return newTransaction(nonce, &to, amount, energyLimit, energyPrice, data)
}

func NewContractCreation(nonce uint64, amount *big.Int, energyLimit uint64, energyPrice *big.Int, data []byte) *Transaction {
	return newTransaction(nonce, nil, amount, energyLimit, energyPrice, data)
}

func newTransaction(nonce uint64, to *common.Address, amount *big.Int, energyLimit uint64, energyPrice *big.Int, data []byte) *Transaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		AccountNonce: nonce,
		Recipient:    to,
		Payload:      data,
		Amount:       new(big.Int),
		EnergyLimit:  energyLimit,
		Price:        new(big.Int),
		Signature:    []byte{},
		NetworkID:    0,
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if energyPrice != nil {
		d.Price.Set(energyPrice)
	}

	return &Transaction{
		data: d,
		time: time.Now(),
	}
}

// EncodeRLP implements rlp.Encoder
func (tx *Transaction) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &tx.data)
}

// DecodeRLP implements rlp.Decoder
func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&tx.data)
	if err == nil {
		tx.size.Store(common.StorageSize(rlp.ListSize(size)))
		tx.time = time.Now()
	}
	return err
}

// MarshalJSON encodes the web3 RPC transaction format.
func (tx *Transaction) MarshalJSON() ([]byte, error) {
	hash := tx.Hash()
	data := tx.data
	data.Hash = &hash
	return data.MarshalJSON()
}

// UnmarshalJSON decodes the web3 RPC transaction format.
func (tx *Transaction) UnmarshalJSON(input []byte) error {
	var dec txdata
	if err := dec.UnmarshalJSON(input); err != nil {
		return err
	}
	*tx = Transaction{
		data: dec,
		time: time.Now(),
	}
	return nil
}

func (tx *Transaction) Data() []byte                { return common.CopyBytes(tx.data.Payload) }
func (tx *Transaction) Energy() uint64              { return tx.data.EnergyLimit }
func (tx *Transaction) EnergyPrice() *big.Int       { return new(big.Int).Set(tx.data.Price) }
func (tx *Transaction) Value() *big.Int             { return new(big.Int).Set(tx.data.Amount) }
func (tx *Transaction) Nonce() uint64               { return tx.data.AccountNonce }
func (tx *Transaction) CheckNonce() bool            { return true }
func (tx *Transaction) NetworkID() uint             { return tx.data.NetworkID }
func (tx *Transaction) Signature() []byte           { return tx.data.Signature }
func (tx *Transaction) SetNetworkID(networkID uint) { tx.data.NetworkID = networkID }

// To returns the recipient address of the transaction.
// It returns nil if the transaction is a contract creation.
func (tx *Transaction) To() *common.Address {
	if tx.data.Recipient == nil {
		return nil
	}
	to := *tx.data.Recipient
	return &to
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := rlpHash(tx)
	tx.hash.Store(v)
	return v
}

// Size returns the true RLP encoded storage size of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *Transaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, &tx.data)
	tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

// AsMessage returns the transaction as a core.Message.
//
// AsMessage requires a signer to derive the sender.
//
// XXX Rename message to something less arbitrary?
func (tx *Transaction) AsMessage(s Signer) (Message, error) {
	from, err := recoverPlain(s, tx)
	if err != nil {
		return Message{}, err
	}
	msg := Message{
		nonce:       tx.data.AccountNonce,
		energyLimit: tx.data.EnergyLimit,
		energyPrice: new(big.Int).Set(tx.data.Price),
		to:          tx.data.Recipient,
		amount:      tx.data.Amount,
		data:        tx.data.Payload,
		from:        from,
		checkNonce:  true,
	}
	return msg, nil
}

// WithSignature returns a new transaction with the given signature.
func (tx *Transaction) WithSignature(signer Signer, sig []byte) (*Transaction, error) {
	cpy := &Transaction{
		data: tx.data,
		time: tx.time,
	}
	cpy.data.NetworkID = uint(signer.NetworkID())
	cpy.data.Signature = sig[:]
	return cpy, nil
}

// Cost returns amount + energyprice * energylimit.
func (tx *Transaction) Cost() *big.Int {
	total := new(big.Int).Mul(tx.data.Price, new(big.Int).SetUint64(tx.data.EnergyLimit))
	total.Add(total, tx.data.Amount)
	return total
}

// Transactions is a Transaction slice type for basic sorting.
type Transactions []*Transaction

// Len returns the length of s.
func (s Transactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp.
func (s Transactions) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}

// TxDifference returns a new set which is the difference between a and b.
func TxDifference(a, b Transactions) Transactions {
	keep := make(Transactions, 0, len(a))

	remove := make(map[common.Hash]struct{})
	for _, tx := range b {
		remove[tx.Hash()] = struct{}{}
	}

	for _, tx := range a {
		if _, ok := remove[tx.Hash()]; !ok {
			keep = append(keep, tx)
		}
	}

	return keep
}

// TxByNonce implements the sort interface to allow sorting a list of transactions
// by their nonces. This is usually only useful for sorting transactions from a
// single account, otherwise a nonce comparison doesn't make much sense.
type TxByNonce Transactions

func (s TxByNonce) Len() int           { return len(s) }
func (s TxByNonce) Less(i, j int) bool { return s[i].data.AccountNonce < s[j].data.AccountNonce }
func (s TxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// TxByPriceAndTime implements both the sort and the heap interface, making it useful
// for all at once sorting as well as individually adding and removing elements.
type TxByPriceAndTime Transactions

func (s TxByPriceAndTime) Len() int { return len(s) }
func (s TxByPriceAndTime) Less(i, j int) bool {
	// If the prices are equal, use the time the transaction was first seen for
	// deterministic sorting
	cmp := s[i].data.Price.Cmp(s[j].data.Price)
	if cmp == 0 {
		return s[i].time.Before(s[j].time)
	}
	return cmp > 0
}
func (s TxByPriceAndTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s *TxByPriceAndTime) Push(x interface{}) {
	*s = append(*s, x.(*Transaction))
}

func (s *TxByPriceAndTime) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

// TransactionsByPriceAndNonce represents a set of transactions that can return
// transactions in a profit-maximizing sorted order, while supporting removing
// entire batches of transactions for non-executable accounts.
type TransactionsByPriceAndNonce struct {
	txs    map[common.Address]Transactions // Per account nonce-sorted list of transactions
	heads  TxByPriceAndTime                // Next transaction for each unique account (price heap)
	signer Signer                          // Signer for the set of transactions
}

// NewTransactionsByPriceAndNonce creates a transaction set that can retrieve
// price sorted transactions in a nonce-honouring way.
//
// Note, the input map is reowned so the caller should not interact any more with
// if after providing it to the constructor.
func NewTransactionsByPriceAndNonce(signer Signer, txs map[common.Address]Transactions) *TransactionsByPriceAndNonce {
	// Initialize a price and received time based heap with the head transactions
	heads := make(TxByPriceAndTime, 0, len(txs))
	for from, accTxs := range txs {
		heads = append(heads, accTxs[0])
		// Ensure the sender address is from the signer
		acc, _ := Sender(signer, accTxs[0])
		txs[acc] = accTxs[1:]
		if from != acc {
			delete(txs, from)
		}
	}
	heap.Init(&heads)

	// Assemble and return the transaction set
	return &TransactionsByPriceAndNonce{
		txs:    txs,
		heads:  heads,
		signer: signer,
	}
}

// Peek returns the next transaction by price.
func (t *TransactionsByPriceAndNonce) Peek() *Transaction {
	if len(t.heads) == 0 {
		return nil
	}
	return t.heads[0]
}

// Shift replaces the current best head with the next one from the same account.
func (t *TransactionsByPriceAndNonce) Shift() {
	acc, _ := Sender(t.signer, t.heads[0])
	if txs, ok := t.txs[acc]; ok && len(txs) > 0 {
		t.heads[0], t.txs[acc] = txs[0], txs[1:]
		heap.Fix(&t.heads, 0)
	} else {
		heap.Pop(&t.heads)
	}
}

// Pop removes the best transaction, *not* replacing it with the next one from
// the same account. This should be used when a transaction cannot be executed
// and hence all subsequent ones should be discarded from the same account.
func (t *TransactionsByPriceAndNonce) Pop() {
	heap.Pop(&t.heads)
}

// Message is a fully derived transaction and implements core.Message
//
// NOTE: In a future PR this will be removed.
type Message struct {
	to          *common.Address
	from        common.Address
	nonce       uint64
	amount      *big.Int
	energyLimit uint64
	energyPrice *big.Int
	data        []byte
	checkNonce  bool
}

func NewMessage(from common.Address, to *common.Address, nonce uint64, amount *big.Int, energyLimit uint64, energyPrice *big.Int, data []byte, checkNonce bool) Message {
	return Message{
		from:        from,
		to:          to,
		nonce:       nonce,
		amount:      amount,
		energyLimit: energyLimit,
		energyPrice: energyPrice,
		data:        data,
		checkNonce:  checkNonce,
	}
}

func (m Message) From() common.Address  { return m.from }
func (m Message) To() *common.Address   { return m.to }
func (m Message) EnergyPrice() *big.Int { return m.energyPrice }
func (m Message) Value() *big.Int       { return m.amount }
func (m Message) Energy() uint64        { return m.energyLimit }
func (m Message) Nonce() uint64         { return m.nonce }
func (m Message) Data() []byte          { return m.data }
func (m Message) CheckNonce() bool      { return m.checkNonce }
