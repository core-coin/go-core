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

package light

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/core-coin/go-core/v2/xcbdb"

	"github.com/core-coin/go-core/v2/consensus/cryptore"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/math"
	"github.com/core-coin/go-core/v2/core"
	"github.com/core-coin/go-core/v2/core/rawdb"
	"github.com/core-coin/go-core/v2/core/state"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/core/vm"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/params"
	"github.com/core-coin/go-core/v2/rlp"
	"github.com/core-coin/go-core/v2/trie"
)

var (
	testBankKey, _ = crypto.UnmarshalPrivateKeyHex("89bdfaa2b6f9c30b94ee98fec96c58ff8507fabf49d36a6267e6cb5516eaa2a9e854eccc041f9f67e109d0eb4f653586855355c5b2b87bb313")
	testBankFunds  = big.NewInt(100000000)

	acc1Key, _ = crypto.UnmarshalPrivateKeyHex("ab856a9af6b0b651dd2f43b5e12193652ec1701c4da6f1c0d2a366ac4b9dabc9433ef09e41ca129552bd2c029086d9b03604de872a3b343204")
	acc2Key, _ = crypto.UnmarshalPrivateKeyHex("c0b711eea422df26d5ffdcaae35fe0527cf647c5ce62d3efb5e09a0e14fc8afe57fac1a5daa330bc10bfa1d3db11e172a822dcfffb86a0b26d")

	testContractCode = common.Hex2Bytes("606060405260cc8060106000396000f360606040526000357c01000000000000000000000000000000000000000000000000000000009004806360cd2685146041578063c16431b914606b57603f565b005b6055600480803590602001909190505060a9565b6040518082815260200191505060405180910390f35b60886004808035906020019091908035906020019091905050608a565b005b80600060005083606481101560025790900160005b50819055505b5050565b6000600060005082606481101560025790900160005b5054905060c7565b91905056")
	testContractAddr common.Address
)

type testOdr struct {
	OdrBackend
	indexerConfig *IndexerConfig
	sdb, ldb      xcbdb.Database
	disable       bool
}

func (odr *testOdr) Database() xcbdb.Database {
	return odr.ldb
}

var ErrOdrDisabled = errors.New("ODR disabled")

func (odr *testOdr) Retrieve(ctx context.Context, req OdrRequest) error {
	if odr.disable {
		return ErrOdrDisabled
	}
	switch req := req.(type) {
	case *BlockRequest:
		number := rawdb.ReadHeaderNumber(odr.sdb, req.Hash)
		if number != nil {
			req.Rlp = rawdb.ReadBodyRLP(odr.sdb, req.Hash, *number)
		}
	case *ReceiptsRequest:
		number := rawdb.ReadHeaderNumber(odr.sdb, req.Hash)
		if number != nil {
			req.Receipts = rawdb.ReadRawReceipts(odr.sdb, req.Hash, *number)
		}
	case *TrieRequest:
		t, _ := trie.New(req.Id.Root, trie.NewDatabase(odr.sdb))
		nodes := NewNodeSet()
		t.Prove(req.Key, 0, nodes)
		req.Proof = nodes
	case *CodeRequest:
		req.Data = rawdb.ReadCode(odr.sdb, req.Hash)
	}
	req.StoreResult(odr.ldb)
	return nil
}

func (odr *testOdr) IndexerConfig() *IndexerConfig {
	return odr.indexerConfig
}

type odrTestFn func(ctx context.Context, db xcbdb.Database, bc *core.BlockChain, lc *LightChain, bhash common.Hash) ([]byte, error)

func TestOdrGetBlockLes2(t *testing.T) { testChainOdr(t, 1, odrGetBlock) }

func odrGetBlock(ctx context.Context, db xcbdb.Database, bc *core.BlockChain, lc *LightChain, bhash common.Hash) ([]byte, error) {
	var block *types.Block
	if bc != nil {
		block = bc.GetBlockByHash(bhash)
	} else {
		block, _ = lc.GetBlockByHash(ctx, bhash)
	}
	if block == nil {
		return nil, nil
	}
	rlp, _ := rlp.EncodeToBytes(block)
	return rlp, nil
}

func TestOdrGetReceiptsLes2(t *testing.T) { testChainOdr(t, 1, odrGetReceipts) }

func odrGetReceipts(ctx context.Context, db xcbdb.Database, bc *core.BlockChain, lc *LightChain, bhash common.Hash) ([]byte, error) {
	var receipts types.Receipts
	if bc != nil {
		number := rawdb.ReadHeaderNumber(db, bhash)
		if number != nil {
			receipts = rawdb.ReadReceipts(db, bhash, *number, bc.Config())
		}
	} else {
		number := rawdb.ReadHeaderNumber(db, bhash)
		if number != nil {
			receipts, _ = GetBlockReceipts(ctx, lc.Odr(), bhash, *number)
		}
	}
	if receipts == nil {
		return nil, nil
	}
	rlp, _ := rlp.EncodeToBytes(receipts)
	return rlp, nil
}

func TestOdrAccountsLes2(t *testing.T) { testChainOdr(t, 1, odrAccounts) }

func odrAccounts(ctx context.Context, db xcbdb.Database, bc *core.BlockChain, lc *LightChain, bhash common.Hash) ([]byte, error) {
	dummyAddr, err := common.HexToAddress("cb721234567812345678123456781234567812345678")
	if err != nil {
		panic(err)
	}
	acc := []common.Address{testBankKey.Address(), acc1Key.Address(), acc2Key.Address(), dummyAddr}

	var st *state.StateDB
	if bc == nil {
		header := lc.GetHeaderByHash(bhash)
		st = NewState(ctx, header, lc.Odr())
	} else {
		header := bc.GetHeaderByHash(bhash)
		st, _ = state.New(header.Root, state.NewDatabase(db), nil)
	}

	var res []byte
	for _, addr := range acc {
		bal := st.GetBalance(addr)
		rlp, _ := rlp.EncodeToBytes(bal)
		res = append(res, rlp...)
	}
	return res, st.Error()
}

func TestOdrContractCallLes2(t *testing.T) { testChainOdr(t, 1, odrContractCall) }

type callmsg struct {
	types.Message
}

func (callmsg) CheckNonce() bool { return false }

func odrContractCall(ctx context.Context, db xcbdb.Database, bc *core.BlockChain, lc *LightChain, bhash common.Hash) ([]byte, error) {
	data := common.Hex2Bytes("60CD26850000000000000000000000000000000000000000000000000000000000000000")
	config := params.MainnetChainConfig

	var res []byte
	for i := 0; i < 3; i++ {
		data[35] = byte(i)

		var (
			st     *state.StateDB
			header *types.Header
			chain  core.ChainContext
		)
		if bc == nil {
			chain = lc
			header = lc.GetHeaderByHash(bhash)
			st = NewState(ctx, header, lc.Odr())
		} else {
			chain = bc
			header = bc.GetHeaderByHash(bhash)
			st, _ = state.New(header.Root, state.NewDatabase(db), nil)
		}

		// Perform read-only call.
		st.SetBalance(testBankKey.Address(), math.MaxBig256)
		msg := callmsg{types.NewMessage(testBankKey.Address(), &testContractAddr, 0, new(big.Int), 1000000, new(big.Int), data, false)}
		txContext := core.NewCVMTxContext(msg)
		context := core.NewCVMBlockContext(header, chain, nil)
		vmenv := vm.NewCVM(context, txContext, st, config, vm.Config{})
		gp := new(core.EnergyPool).AddEnergy(math.MaxUint64)
		result, _ := core.ApplyMessage(vmenv, msg, gp)
		res = append(res, result.Return()...)
		if st.Error() != nil {
			return res, st.Error()
		}
	}
	return res, nil
}

func testChainGen(i int, block *core.BlockGen) {
	signer := types.NewNucleusSigner(params.MainnetChainConfig.NetworkID)
	switch i {
	case 0:
		// In block 1, the test bank sends account #1 some core.
		tx, _ := types.SignTx(types.NewTransaction(block.TxNonce(testBankKey.Address()), acc1Key.Address(), big.NewInt(10000), params.TxEnergy, nil, nil), signer, testBankKey)
		block.AddTx(tx)
	case 1:
		// In block 2, the test bank sends some more core to account #1.
		// acc1Key.Address() passes it on to account #2.
		// acc1Key.Address() creates a test contract.
		tx1, _ := types.SignTx(types.NewTransaction(block.TxNonce(testBankKey.Address()), acc1Key.Address(), big.NewInt(1000), params.TxEnergy, nil, nil), signer, testBankKey)
		nonce := block.TxNonce(acc1Key.Address())
		tx2, _ := types.SignTx(types.NewTransaction(nonce, acc2Key.Address(), big.NewInt(1000), params.TxEnergy, nil, nil), signer, acc1Key)
		nonce++
		tx3, _ := types.SignTx(types.NewContractCreation(nonce, big.NewInt(0), 1000000, big.NewInt(0), testContractCode), signer, acc1Key)
		testContractAddr = crypto.CreateAddress(acc1Key.Address(), nonce)
		block.AddTx(tx1)
		block.AddTx(tx2)
		block.AddTx(tx3)
	case 2:
		// Block 3 is empty but was mined by account #2.
		block.SetCoinbase(acc2Key.Address())
		block.SetExtra([]byte("yeehaw"))
		data := common.Hex2Bytes("C16431B900000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001")
		tx, _ := types.SignTx(types.NewTransaction(block.TxNonce(testBankKey.Address()), testContractAddr, big.NewInt(0), 100000, nil, data), signer, testBankKey)
		block.AddTx(tx)
	case 3:
		// Block 4 includes blocks 2 and 3 as uncle headers (with modified extra data).
		b2 := block.PrevBlock(1).Header()
		b2.Extra = []byte("foo")
		block.AddUncle(b2)
		b3 := block.PrevBlock(2).Header()
		b3.Extra = []byte("foo")
		block.AddUncle(b3)
		data := common.Hex2Bytes("C16431B900000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000002")
		tx, _ := types.SignTx(types.NewTransaction(block.TxNonce(testBankKey.Address()), testContractAddr, big.NewInt(0), 100000, nil, data), signer, testBankKey)
		block.AddTx(tx)
	}
}

func testChainOdr(t *testing.T, protocol int, fn odrTestFn) {
	var (
		sdb     = rawdb.NewMemoryDatabase()
		ldb     = rawdb.NewMemoryDatabase()
		gspec   = core.Genesis{Alloc: core.GenesisAlloc{testBankKey.Address(): {Balance: testBankFunds}}}
		genesis = gspec.MustCommit(sdb)
	)
	gspec.MustCommit(ldb)
	// Assemble the test environment
	blockchain, _ := core.NewBlockChain(sdb, nil, params.MainnetChainConfig, cryptore.NewFullFaker(), vm.Config{}, nil, nil)
	gchain, _ := core.GenerateChain(params.MainnetChainConfig, genesis, cryptore.NewFaker(), sdb, 4, testChainGen)
	if _, err := blockchain.InsertChain(gchain); err != nil {
		t.Fatal(err)
	}

	odr := &testOdr{sdb: sdb, ldb: ldb, indexerConfig: TestClientIndexerConfig}
	lightchain, err := NewLightChain(odr, params.MainnetChainConfig, cryptore.NewFullFaker(), nil)
	if err != nil {
		t.Fatal(err)
	}
	headers := make([]*types.Header, len(gchain))
	for i, block := range gchain {
		headers[i] = block.Header()
	}
	if _, err := lightchain.InsertHeaderChain(headers, 1); err != nil {
		t.Fatal(err)
	}

	test := func(expFail int) {
		for i := uint64(0); i <= blockchain.CurrentHeader().Number.Uint64(); i++ {
			bhash := rawdb.ReadCanonicalHash(sdb, i)
			b1, err := fn(NoOdr, sdb, blockchain, nil, bhash)
			if err != nil {
				t.Fatalf("error in full-node test for block %d: %v", i, err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			exp := i < uint64(expFail)
			b2, err := fn(ctx, ldb, nil, lightchain, bhash)
			if err != nil && exp {
				t.Errorf("error in ODR test for block %d: %v", i, err)
			}

			eq := bytes.Equal(b1, b2)
			if exp && !eq {
				t.Errorf("ODR test output for block %d doesn't match full node", i)
			}
		}
	}

	// expect retrievals to fail (except genesis block) without a les peer
	t.Log("checking without ODR")
	odr.disable = true
	test(1)

	// expect all retrievals to pass with ODR enabled
	t.Log("checking with ODR")
	odr.disable = false
	test(len(gchain))

	// still expect all retrievals to pass, now data should be cached locally
	t.Log("checking without ODR, should be cached")
	odr.disable = true
	test(len(gchain))
}
