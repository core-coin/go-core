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
	"github.com/core-coin/go-core/xce/downloader"
	"github.com/core-coin/go-core/xce/energyprice"
	"github.com/core-coin/go-core/xcedb"
)

// XceAPIBackend implements xceapi.Backend for full nodes
type XceAPIBackend struct {
	extRPCEnabled bool
	xce           *Core
	gpo           *energyprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *XceAPIBackend) ChainConfig() *params.ChainConfig {
	return b.xce.blockchain.Config()
}

func (b *XceAPIBackend) CurrentBlock() *types.Block {
	return b.xce.blockchain.CurrentBlock()
}

func (b *XceAPIBackend) SetHead(number uint64) {
	b.xce.protocolManager.downloader.Cancel()
	b.xce.blockchain.SetHead(number)
}

func (b *XceAPIBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		block := b.xce.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if number == rpc.LatestBlockNumber {
		return b.xce.blockchain.CurrentBlock().Header(), nil
	}
	return b.xce.blockchain.GetHeaderByNumber(uint64(number)), nil
}

func (b *XceAPIBackend) HeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Header, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.HeaderByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header := b.xce.blockchain.GetHeaderByHash(hash)
		if header == nil {
			return nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && b.xce.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, errors.New("hash is not currently canonical")
		}
		return header, nil
	}
	return nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *XceAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.xce.blockchain.GetHeaderByHash(hash), nil
}

func (b *XceAPIBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		block := b.xce.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if number == rpc.LatestBlockNumber {
		return b.xce.blockchain.CurrentBlock(), nil
	}
	return b.xce.blockchain.GetBlockByNumber(uint64(number)), nil
}

func (b *XceAPIBackend) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.xce.blockchain.GetBlockByHash(hash), nil
}

func (b *XceAPIBackend) BlockByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Block, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.BlockByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header := b.xce.blockchain.GetHeaderByHash(hash)
		if header == nil {
			return nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && b.xce.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, errors.New("hash is not currently canonical")
		}
		block := b.xce.blockchain.GetBlock(hash, header.Number.Uint64())
		if block == nil {
			return nil, errors.New("header found, but block body is missing")
		}
		return block, nil
	}
	return nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *XceAPIBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if number == rpc.PendingBlockNumber {
		block, state := b.xce.miner.Pending()
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
	stateDb, err := b.xce.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *XceAPIBackend) StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *types.Header, error) {
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
		if blockNrOrHash.RequireCanonical && b.xce.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, nil, errors.New("hash is not currently canonical")
		}
		stateDb, err := b.xce.BlockChain().StateAt(header.Root)
		return stateDb, header, err
	}
	return nil, nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *XceAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	return b.xce.blockchain.GetReceiptsByHash(hash), nil
}

func (b *XceAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	receipts := b.xce.blockchain.GetReceiptsByHash(hash)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *XceAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.xce.blockchain.GetTdByHash(blockHash)
}

func (b *XceAPIBackend) GetCVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header) (*vm.CVM, func() error, error) {
	vmError := func() error { return nil }

	context := core.NewCVMContext(msg, header, b.xce.BlockChain(), nil)
	return vm.NewCVM(context, state, b.xce.blockchain.Config(), *b.xce.blockchain.GetVMConfig()), vmError, nil
}

func (b *XceAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.xce.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *XceAPIBackend) SubscribePendingLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.xce.miner.SubscribePendingLogs(ch)
}

func (b *XceAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.xce.BlockChain().SubscribeChainEvent(ch)
}

func (b *XceAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.xce.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *XceAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.xce.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *XceAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.xce.BlockChain().SubscribeLogsEvent(ch)
}

func (b *XceAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.xce.txPool.AddLocal(signedTx)
}

func (b *XceAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.xce.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *XceAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.xce.txPool.Get(hash)
}

func (b *XceAPIBackend) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error) {
	tx, blockHash, blockNumber, index := rawdb.ReadTransaction(b.xce.ChainDb(), txHash)
	return tx, blockHash, blockNumber, index, nil
}

func (b *XceAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.xce.txPool.Nonce(addr), nil
}

func (b *XceAPIBackend) Stats() (pending int, queued int) {
	return b.xce.txPool.Stats()
}

func (b *XceAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.xce.TxPool().Content()
}

func (b *XceAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.xce.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *XceAPIBackend) Downloader() *downloader.Downloader {
	return b.xce.Downloader()
}

func (b *XceAPIBackend) ProtocolVersion() int {
	return b.xce.XceVersion()
}

func (b *XceAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *XceAPIBackend) ChainDb() xcedb.Database {
	return b.xce.ChainDb()
}

func (b *XceAPIBackend) EventMux() *event.TypeMux {
	return b.xce.EventMux()
}

func (b *XceAPIBackend) AccountManager() *accounts.Manager {
	return b.xce.AccountManager()
}

func (b *XceAPIBackend) ExtRPCEnabled() bool {
	return b.extRPCEnabled
}

func (b *XceAPIBackend) RPCEnergyCap() *big.Int {
	return b.xce.config.RPCEnergyCap
}

func (b *XceAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.xce.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *XceAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.xce.bloomRequests)
	}
}
