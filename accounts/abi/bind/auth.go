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

package bind

import (
	"errors"
	"github.com/core-coin/go-core/log"
	"io"
	"io/ioutil"
	"math/big"

	eddsa "github.com/core-coin/go-goldilocks"

	"github.com/core-coin/go-core/accounts"
	"github.com/core-coin/go-core/accounts/external"
	"github.com/core-coin/go-core/accounts/keystore"
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/crypto"
)

// ErrNoNetworkID is returned whenever the user failed to specify a network id.
var ErrNoNetworkID = errors.New("no network id specified")

// ErrNotAuthorized is returned when an account is not properly unlocked.
var ErrNotAuthorized = errors.New("not authorized to sign this account")

// NewTransactor is a utility method to easily create a transaction signer from
// an encrypted json key stream and the associated passphrase.
//
// Deprecated: Use NewTransactorWithNetworkID instead.
func NewTransactor(keyin io.Reader, passphrase string) (*TransactOpts, error) {
	log.Warn("WARNING: NewTransactor has been deprecated in favour of NewTransactorWithNetworkID")
	json, err := ioutil.ReadAll(keyin)
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(json, passphrase)
	if err != nil {
		return nil, err
	}
	return NewKeyedTransactor(key.PrivateKey), nil
}

// NewKeyStoreTransactor is a utility method to easily create a transaction signer from
// an decrypted key from a keystore.
//
// Deprecated: Use NewKeyStoreTransactorWithNetworkID instead.
func NewKeyStoreTransactor(keystore *keystore.KeyStore, account accounts.Account) (*TransactOpts, error) {
	log.Warn("WARNING: NewKeyStoreTransactor has been deprecated in favour of NewTransactorWithNetworkID")
	signer := types.NucleusSigner{}
	return &TransactOpts{
		From: account.Address,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != account.Address {
				return nil, ErrNotAuthorized
			}
			tx.SetNetworkID(uint(signer.NetworkID()))
			signature, err := keystore.SignHash(account, signer.Hash(tx).Bytes())
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)
		},
	}, nil
}

// NewKeyedTransactor is a utility method to easily create a transaction signer
// from a single private key.
//
// Deprecated: Use NewKeyedTransactorWithNetworkID instead.
func NewKeyedTransactor(key *eddsa.PrivateKey) *TransactOpts {
	log.Warn("WARNING: NewKeyedTransactor has been deprecated in favour of NewKeyedTransactorWithNetworkID")
	pub := eddsa.Ed448DerivePublicKey(*key)
	keyAddr := crypto.PubkeyToAddress(pub)
	signer := types.NucleusSigner{}
	return &TransactOpts{
		From: keyAddr,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != keyAddr {
				return nil, ErrNotAuthorized
			}
			tx.SetNetworkID(uint(signer.NetworkID()))
			signature, err := crypto.Sign(signer.Hash(tx).Bytes(), key)
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)
		},
	}
}

// NewTransactorWithNetworkID is a utility method to easily create a transaction signer from
// an encrypted json key stream and the associated passphrase.
func NewTransactorWithNetworkID(keyin io.Reader, passphrase string, networkID *big.Int) (*TransactOpts, error) {
	json, err := ioutil.ReadAll(keyin)
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(json, passphrase)
	if err != nil {
		return nil, err
	}
	return NewKeyedTransactorWithNetworkID(key.PrivateKey, networkID)
}

// NewKeyStoreTransactorWithNetworkID is a utility method to easily create a transaction signer from
// an decrypted key from a keystore.
func NewKeyStoreTransactorWithNetworkID(keystore *keystore.KeyStore, account accounts.Account, networkID *big.Int) (*TransactOpts, error) {
	if networkID == nil {
		return nil, ErrNoNetworkID
	}
	signer := types.NewNucleusSigner(networkID)
	return &TransactOpts{
		From: account.Address,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != account.Address {
				return nil, ErrNotAuthorized
			}
			tx.SetNetworkID(uint(signer.NetworkID()))
			signature, err := keystore.SignHash(account, signer.Hash(tx).Bytes())
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)
		},
	}, nil
}

// NewKeyedTransactorWithNetworkID is a utility method to easily create a transaction signer
// from a single private key.
func NewKeyedTransactorWithNetworkID(key *eddsa.PrivateKey, networkID *big.Int) (*TransactOpts, error) {
	pub := eddsa.Ed448DerivePublicKey(*key)
	keyAddr := crypto.PubkeyToAddress(pub)
	if networkID == nil {
		return nil, ErrNoNetworkID
	}
	signer := types.NewNucleusSigner(networkID)
	return &TransactOpts{
		From: keyAddr,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != keyAddr {
				return nil, ErrNotAuthorized
			}
			tx.SetNetworkID(uint(signer.NetworkID()))
			signature, err := crypto.Sign(signer.Hash(tx).Bytes(), key)
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)
		},
	}, nil
}

// NewClefTransactor is a utility method to easily create a transaction signer
// with a clef backend.
func NewClefTransactor(clef *external.ExternalSigner, account accounts.Account) *TransactOpts {
	return &TransactOpts{
		From: account.Address,
		Signer: func(address common.Address, transaction *types.Transaction) (*types.Transaction, error) {
			if address != account.Address {
				return nil, ErrNotAuthorized
			}
			return clef.SignTx(account, transaction, nil) // Clef enforces its own network id
		},
	}
}
