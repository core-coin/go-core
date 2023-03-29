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

package bind_test

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/core-coin/go-core/v2/accounts/abi/bind"
	"github.com/core-coin/go-core/v2/accounts/abi/bind/backends"
	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/core"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/crypto"
)

var testKey, _ = crypto.UnmarshalPrivateKeyHex("c0b711eea422df26d5ffdcaae35fe0527cf647c5ce62d3efb5e09a0e14fc8afe57fac1a5daa330bc10bfa1d3db11e172a822dcfffb86a0b26d")

var addr, _ = common.HexToAddress("cb375a538daf54f2e568bb4237357b1cee1aa3cb7eba")

var waitDeployedTests = map[string]struct {
	code        string
	energy      uint64
	wantAddress common.Address
	wantErr     error
}{
	"successful deploy": {
		code:        `6060604052600a8060106000396000f360606040526008565b00`,
		energy:      3000000,
		wantAddress: addr,
	},
	"empty code": {
		code:        ``,
		energy:      300000,
		wantErr:     bind.ErrNoCodeAfterDeploy,
		wantAddress: addr,
	},
}

func TestWaitDeployed(t *testing.T) {
	for name, test := range waitDeployedTests {
		backend := backends.NewSimulatedBackend(
			core.GenesisAlloc{
				testKey.Address(): {Balance: big.NewInt(10000000000)},
			},
			10000000,
		)
		defer backend.Close()

		// Create the transaction.
		tx := types.NewContractCreation(0, big.NewInt(0), test.energy, big.NewInt(1), common.FromHex(test.code))
		tx, _ = types.SignTx(tx, types.NewNucleusSigner(backend.Blockchain().Config().NetworkID), testKey)

		// Wait for it to get mined in the background.
		var (
			err     error
			address common.Address
			mined   = make(chan struct{})
			ctx     = context.Background()
		)
		go func() {
			address, err = bind.WaitDeployed(ctx, backend, tx)
			close(mined)
		}()

		// Send and mine the transaction.
		backend.SendTransaction(ctx, tx)
		backend.Commit()

		select {
		case <-mined:
			if err != test.wantErr {
				t.Errorf("test %q: error mismatch: want %q, got %q", name, test.wantErr, err)
			}
			if address != test.wantAddress {
				t.Errorf("test %q: unexpected contract address %s", name, address.Hex())
			}
		case <-time.After(2 * time.Second):
			t.Errorf("test %q: timeout", name)
		}
	}
}

func TestWaitDeployedCornerCases(t *testing.T) {
	backend := backends.NewSimulatedBackend(
		core.GenesisAlloc{
			testKey.Address(): {Balance: big.NewInt(10000000000)},
		},
		10000000,
	)
	defer backend.Close()

	addr, _ := common.HexToAddress("cb270000000000000000000000000000000000000001")
	// Create a transaction to an account.
	code := "6060604052600a8060106000396000f360606040526008565b00"
	tx := types.NewTransaction(0, addr, big.NewInt(0), 3000000, big.NewInt(1), common.FromHex(code))
	tx, _ = types.SignTx(tx, types.NewNucleusSigner(backend.Blockchain().Config().NetworkID), testKey)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	backend.SendTransaction(ctx, tx)
	backend.Commit()
	notContentCreation := errors.New("tx is not contract creation")
	if _, err := bind.WaitDeployed(ctx, backend, tx); err.Error() != notContentCreation.Error() {
		t.Errorf("error missmatch: want %q, got %q, ", notContentCreation, err)
	}

	// Create a transaction that is not mined.
	tx = types.NewContractCreation(1, big.NewInt(0), 3000000, big.NewInt(1), common.FromHex(code))
	tx, _ = types.SignTx(tx, types.NewNucleusSigner(backend.Blockchain().Config().NetworkID), testKey)

	go func() {
		contextCanceled := errors.New("context canceled")
		if _, err := bind.WaitDeployed(ctx, backend, tx); err.Error() != contextCanceled.Error() {
			t.Errorf("error missmatch: want %q, got %q, ", contextCanceled, err)
		}
	}()

	backend.SendTransaction(ctx, tx)
	cancel()
}
