// Copyright 2016 by the Authors
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
	"errors"
	"github.com/core-coin/ed448"
	"math/big"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
)

var ErrInvalidNetworkId = errors.New("invalid network id for signer")

// sigCache is used to cache the derived sender and contains
// the signer used to derive it.
type sigCache struct {
	signer Signer
	from   common.Address
}

// MakeSigner returns a Signer based on the given chain config and block number.
func MakeSigner(networkID *big.Int) Signer {
	var signer = NewNucleusSigner(networkID)
	return signer
}

// SignTx signs the transaction using the given signer and private key
func SignTx(tx *Transaction, s Signer, prv ed448.PrivateKey) (*Transaction, error) {
	tx.data.NetworkID = uint(s.NetworkID())
	h := s.Hash(tx)
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	return tx.WithSignature(s, sig[:])
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
	// NetworkID returns network id stored in signer
	NetworkID() int
}

// NucleusSigner implements Signer with network id.
type NucleusSigner struct {
	networkId *big.Int
}

func NewNucleusSigner(networkId *big.Int) NucleusSigner {
	if networkId == nil {
		networkId = new(big.Int)
	}
	return NucleusSigner{
		networkId: networkId,
	}
}

func (s NucleusSigner) Equal(s2 Signer) bool {
	nucleus, ok := s2.(NucleusSigner)
	return ok && nucleus.networkId.Cmp(s.networkId) == 0
}

func (s NucleusSigner) Sender(tx *Transaction) (common.Address, error) {
	if tx.data.NetworkID != 0 && s.NetworkID() != 0 {
		if tx.data.NetworkID != uint(s.networkId.Int64()) {
			return common.Address{}, ErrInvalidNetworkId
		}
	}
	return recoverPlain(s, tx)
}

func (s NucleusSigner) NetworkID() int {
	return int(s.networkId.Int64())
}

// Hash returns the hash to be signed by the sender.
// It does not uniquely identify the transaction.
func (s NucleusSigner) Hash(tx *Transaction) common.Hash {
	return rlpHash([]interface{}{
		tx.data.AccountNonce,
		tx.data.Price,
		tx.data.EnergyLimit,
		tx.data.Recipient,
		tx.data.Amount,
		tx.data.Payload,
		tx.data.NetworkID,
	})
}

func recoverPlain(signer Signer, tx *Transaction) (common.Address, error) {
	if len(tx.data.Signature) != crypto.ExtendedSignatureLength {
		return common.Address{}, ErrInvalidSig
	}
	pubk, err := crypto.SigToPub(signer.Hash(tx).Bytes(), tx.data.Signature[:])
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(pubk), nil
}
