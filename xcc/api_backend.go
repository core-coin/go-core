// Copyright 2015 The go-core Authors
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

package xcc

import (
	"context"
	"errors"
	"math/big"

	"github.com/core-coin/go-core/accounts"
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/core"
	"github.com/core-coin/go-core/core/bloombits"
	"github.com/core-coin/go-core/core/rawdb"
	"github.com/core-coin/go-core/core/state"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/core/vm"
	"github.com/core-coin/go-core/event"
	"github.com/core-coin/go-core/params"
	"github.com/core-coin/go-core/rpc"
	"github.com/core-coin/go-core/xcc/downloader"
	"github.com/core-coin/go-core/xcc/energyprice"
	"github.com/core-coin/go-core/xccdb"
)

// XccAPIBackend implements xccapi.Backend for full nodes
type XccAPIBackend struct {
	extRPCEnabled bool
	xcc           *Core
	gpo           *energyprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *XccAPIBackend) ChainConfig() *params.ChainConfig {
	return b.xcc.blockchain.Config()
}

func (b *XccAPIBackend) CurrentBlock() *types.Block {
	return b.xcc.blockchain.CurrentBlock()
}

func (b *XccAPIBackend) SetHead(number uint64) {
	b.xcc.protocolManager.downloader.Cancel()
	b.xcc.blockchain.SetHead(number)
}

func (b *XccAPIBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		block := b.xcc.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if number == rpc.LatestBlockNumber {
		return b.xcc.blockchain.CurrentBlock().Header(), nil
	}
	return b.xcc.blockchain.GetHeaderByNumber(uint64(number)), nil
}

func (b *XccAPIBackend) HeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Header, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.HeaderByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header := b.xcc.blockchain.GetHeaderByHash(hash)
		if header == nil {
			return nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && b.xcc.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, errors.New("hash is not currently canonical")
		}
		return header, nil
	}
	return nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *XccAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.xcc.blockchain.GetHeaderByHash(hash), nil
}

func (b *XccAPIBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		block := b.xcc.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if number == rpc.LatestBlockNumber {
		return b.xcc.blockchain.CurrentBlock(), nil
	}
	return b.xcc.blockchain.GetBlockByNumber(uint64(number)), nil
}

func (b *XccAPIBackend) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.xcc.blockchain.GetBlockByHash(hash), nil
}

func (b *XccAPIBackend) BlockByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Block, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.BlockByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header := b.xcc.blockchain.GetHeaderByHash(hash)
		if header == nil {
			return nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && b.xcc.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, errors.New("hash is not currently canonical")
		}
		block := b.xcc.blockchain.GetBlock(hash, header.Number.Uint64())
		if block == nil {
			return nil, errors.New("header found, but block body is missing")
		}
		return block, nil
	}
	return nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *XccAPIBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if number == rpc.PendingBlockNumber {
		block, state := b.xcc.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, number)
	if err != nil {
		return nil, nil, err
	}
	if header == nil {
		return nil, nil, errors.New("header not found")
	}
	stateDb, err := b.xcc.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *XccAPIBackend) StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *types.Header, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.StateAndHeaderByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header, err := b.HeaderByHash(ctx, hash)
		if err != nil {
			return nil, nil, err
		}
		if header == nil {
			return nil, nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && b.xcc.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, nil, errors.New("hash is not currently canonical")
		}
		stateDb, err := b.xcc.BlockChain().StateAt(header.Root)
		return stateDb, header, err
	}
	return nil, nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *XccAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	return b.xcc.blockchain.GetReceiptsByHash(hash), nil
}

func (b *XccAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	receipts := b.xcc.blockchain.GetReceiptsByHash(hash)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *XccAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.xcc.blockchain.GetTdByHash(blockHash)
}

func (b *XccAPIBackend) GetCVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header) (*vm.CVM, func() error, error) {
	vmError := func() error { return nil }

	context := core.NewCVMContext(msg, header, b.xcc.BlockChain(), nil)
	return vm.NewCVM(context, state, b.xcc.blockchain.Config(), *b.xcc.blockchain.GetVMConfig()), vmError, nil
}

func (b *XccAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.xcc.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *XccAPIBackend) SubscribePendingLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.xcc.miner.SubscribePendingLogs(ch)
}

func (b *XccAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.xcc.BlockChain().SubscribeChainEvent(ch)
}

func (b *XccAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.xcc.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *XccAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.xcc.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *XccAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.xcc.BlockChain().SubscribeLogsEvent(ch)
}

func (b *XccAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.xcc.txPool.AddLocal(signedTx)
}

func (b *XccAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.xcc.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *XccAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.xcc.txPool.Get(hash)
}

func (b *XccAPIBackend) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error) {
	tx, blockHash, blockNumber, index := rawdb.ReadTransaction(b.xcc.ChainDb(), txHash)
	return tx, blockHash, blockNumber, index, nil
}

func (b *XccAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.xcc.txPool.Nonce(addr), nil
}

func (b *XccAPIBackend) Stats() (pending int, queued int) {
	return b.xcc.txPool.Stats()
}

func (b *XccAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.xcc.TxPool().Content()
}

func (b *XccAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.xcc.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *XccAPIBackend) Downloader() *downloader.Downloader {
	return b.xcc.Downloader()
}

func (b *XccAPIBackend) ProtocolVersion() int {
	return b.xcc.XccVersion()
}

func (b *XccAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *XccAPIBackend) ChainDb() xccdb.Database {
	return b.xcc.ChainDb()
}

func (b *XccAPIBackend) EventMux() *event.TypeMux {
	return b.xcc.EventMux()
}

func (b *XccAPIBackend) AccountManager() *accounts.Manager {
	return b.xcc.AccountManager()
}

func (b *XccAPIBackend) ExtRPCEnabled() bool {
	return b.extRPCEnabled
}

func (b *XccAPIBackend) RPCEnergyCap() *big.Int {
	return b.xcc.config.RPCEnergyCap
}

func (b *XccAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.xcc.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *XccAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.xcc.bloomRequests)
	}
}
