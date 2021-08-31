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
	"github.com/core-coin/ed448"
	"math/big"
	"testing"
	"time"

	"github.com/core-coin/go-core/accounts/abi/bind"
	"github.com/core-coin/go-core/accounts/abi/bind/backends"
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/core"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/crypto"
)

var testKey, _ = crypto.HexToEDDSA("c0b711eea422df26d5ffdcaae35fe0527cf647c5ce62d3efb5e09a0e14fc8afe57fac1a5daa330bc10bfa1d3db11e172a822dcfffb86a0b26d")

var addr, addrErr = common.HexToAddress("cb375a538daf54f2e568bb4237357b1cee1aa3cb7eba")

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
	if addrErr != nil {
		t.Error(addrErr)
	}
	for name, test := range waitDeployedTests {
		pub := ed448.Ed448DerivePublicKey(testKey)
		backend := backends.NewSimulatedBackend(
			core.GenesisAlloc{
				crypto.PubkeyToAddress(pub): {Balance: big.NewInt(10000000000)},
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
				t.Errorf("test %q: error mismatch: got %q, want %q", name, err, test.wantErr)
			}
			if address != test.wantAddress {
				t.Errorf("test %q: unexpected contract address %s", name, address.Hex())
			}
		case <-time.After(2 * time.Second):
			t.Errorf("test %q: timeout", name)
		}
	}
}
