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

package les

import (
	"context"
	"testing"
	"time"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/core/rawdb"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/light"
	"github.com/core-coin/go-core/xcbdb"
)

var testBankSecureTrieKey = secAddr(bankAddr)

func secAddr(addr common.Address) []byte {
	return crypto.SHA3(addr[:])
}

type accessTestFn func(db xcbdb.Database, bhash common.Hash, number uint64) light.OdrRequest

func TestBlockAccessLes2(t *testing.T) { testAccess(t, 2, tfBlockAccess) }
func TestBlockAccessLes3(t *testing.T) { testAccess(t, 3, tfBlockAccess) }

func tfBlockAccess(db xcbdb.Database, bhash common.Hash, number uint64) light.OdrRequest {
	return &light.BlockRequest{Hash: bhash, Number: number}
}

func TestReceiptsAccessLes2(t *testing.T) { testAccess(t, 2, tfReceiptsAccess) }
func TestReceiptsAccessLes3(t *testing.T) { testAccess(t, 3, tfReceiptsAccess) }

func tfReceiptsAccess(db xcbdb.Database, bhash common.Hash, number uint64) light.OdrRequest {
	return &light.ReceiptsRequest{Hash: bhash, Number: number}
}

func TestTrieEntryAccessLes2(t *testing.T) { testAccess(t, 2, tfTrieEntryAccess) }
func TestTrieEntryAccessLes3(t *testing.T) { testAccess(t, 3, tfTrieEntryAccess) }

func tfTrieEntryAccess(db xcbdb.Database, bhash common.Hash, number uint64) light.OdrRequest {
	if number := rawdb.ReadHeaderNumber(db, bhash); number != nil {
		return &light.TrieRequest{Id: light.StateTrieID(rawdb.ReadHeader(db, bhash, *number)), Key: testBankSecureTrieKey}
	}
	return nil
}

func TestCodeAccessLes2(t *testing.T) { testAccess(t, 2, tfCodeAccess) }
func TestCodeAccessLes3(t *testing.T) { testAccess(t, 3, tfCodeAccess) }

func tfCodeAccess(db xcbdb.Database, bhash common.Hash, num uint64) light.OdrRequest {
	number := rawdb.ReadHeaderNumber(db, bhash)
	if number != nil {
		return nil
	}
	header := rawdb.ReadHeader(db, bhash, *number)
	if header.Number.Uint64() < testContractDeployed {
		return nil
	}
	sti := light.StateTrieID(header)
	ci := light.StorageTrieID(sti, crypto.SHA3Hash(testContractAddr[:]), common.Hash{})
	return &light.CodeRequest{Id: ci, Hash: crypto.SHA3Hash(testContractCodeDeployed)}
}

func testAccess(t *testing.T, protocol int, fn accessTestFn) {
	t.Skip("skip long-running tests")
	// Assemble the test environment
	server, client, tearDown := newClientServerEnv(t, 4, protocol, nil, nil, 0, false, true, true)
	defer tearDown()

	// Ensure the client has synced all necessary data.
	clientHead := client.handler.backend.blockchain.CurrentHeader()
	if clientHead.Number.Uint64() != 4 {
		t.Fatalf("Failed to sync the chain with server, head: %v", clientHead.Number.Uint64())
	}

	test := func(expFail uint64) {
		for i := uint64(0); i <= server.handler.blockchain.CurrentHeader().Number.Uint64(); i++ {
			bhash := rawdb.ReadCanonicalHash(server.db, i)
			if req := fn(client.db, bhash, i); req != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				err := client.handler.backend.odr.Retrieve(ctx, req)
				cancel()

				got := err == nil
				exp := i < expFail
				if exp && !got {
					t.Errorf("object retrieval failed")
				}
				if !exp && got {
					t.Errorf("unexpected object retrieval success")
				}
			}
		}
	}
	test(5)
}
