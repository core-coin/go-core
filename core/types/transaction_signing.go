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
	"github.com/core-coin/eddsa"
	"math/big"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
)

// sigCache is used to cache the derived sender and contains
// the signer used to derive it.
type sigCache struct {
	signer Signer
	from   common.Address
}

// MakeSigner returns a Signer based on the given chain config and block number.
func MakeSigner() Signer {
	var signer Signer
	signer = NucleusSigner{}
	return signer
}

// SignTx signs the transaction using the given signer and private key
func SignTx(tx *Transaction, s Signer, prv *eddsa.PrivateKey) (*Transaction, error) {
	h := s.Hash(tx)
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	return tx.WithSignature(s, sig)
}

// Sender may cache the address, allowing it to be used regardless of
// signing method. The cache is invalidated if the cached signer does
// not match the signer used in the current call.
func Sender(signer Signer, tx *Transaction) (common.Address, error) {
	if sc := tx.from.Load(); sc != nil {
		sigCache := sc.(sigCache)
		// If the signer used to derive from in a previous
		// call is not the same as used current, invalidate
		// the cache.
		if sigCache.signer.Equal(signer) {
			return sigCache.from, nil
		}
	}

	addr, err := signer.Sender(tx)
	if err != nil {
		return common.Address{}, err
	}
	tx.from.Store(sigCache{signer: signer, from: addr})
	return addr, nil
}

// Signer encapsulates transaction signature handling. Note that this interface is not a
// stable API and may change at any time to accommodate new protocol rules.
type Signer interface {
	// Sender returns the sender address of the transaction.
	Sender(tx *Transaction) (common.Address, error)
	// Hash returns the hash to be signed.
	Hash(tx *Transaction) common.Hash
	// Equal returns true if the given signer is the same as the receiver.
	Equal(Signer) bool
}

// CIP155Transaction implements Signer using the CIP155 rules.
type CIP155Signer struct {
	chainId *big.Int
}

func NewCIP155Signer(chainId *big.Int) CIP155Signer {
	if chainId == nil {
		chainId = new(big.Int)
	}
	return CIP155Signer{
		chainId: chainId,
	}
}

func (s CIP155Signer) Equal(s2 Signer)bool {
	cip155, ok := s2.(CIP155Signer)
	return ok && cip155.chainId.Cmp(s.chainId) == 0
}

func (s CIP155Signer) Sender(tx *Transaction) (common.Address, error) {
	return tx.data.Spender, nil
}

// Hash returns the hash to be signed by the sender.
// It does not uniquely identify the transaction.
func (s CIP155Signer) Hash(tx *Transaction) common.Hash {
	return rlpHash([]interface{}{
		tx.data.AccountNonce,
		tx.data.Price,
		tx.data.EnergyLimit,
		tx.data.Recipient,
		tx.data.Amount,
		tx.data.Payload,
		s.chainId, uint(0), uint(0),
	})
}

type NucleusSigner struct{}

func (s NucleusSigner) Equal(s2 Signer) bool {
	_, ok := s2.(NucleusSigner)
	return ok
}

func (hs NucleusSigner) Sender(tx *Transaction) (common.Address, error) {
	return tx.data.Spender, nil
}

// Hash returns the hash to be signed by the sender.
// It does not uniquely identify the transaction.
func (hs NucleusSigner) Hash(tx *Transaction) common.Hash {
	return rlpHash([]interface{}{
		tx.data.AccountNonce,
		tx.data.Price,
		tx.data.EnergyLimit,
		tx.data.Recipient,
		tx.data.Amount,
		tx.data.Payload,
	})
}
