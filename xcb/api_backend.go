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

package xcb

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
	"github.com/core-coin/go-core/xcb/downloader"
	"github.com/core-coin/go-core/xcb/energyprice"
	"github.com/core-coin/go-core/xcbdb"
)

// XcbAPIBackend implements xcbapi.Backend for full nodes
type XcbAPIBackend struct {
	extRPCEnabled bool
	xcb           *Core
	gpo           *energyprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *XcbAPIBackend) ChainConfig() *params.ChainConfig {
	return b.xcb.blockchain.Config()
}

func (b *XcbAPIBackend) CurrentBlock() *types.Block {
	return b.xcb.blockchain.CurrentBlock()
}

func (b *XcbAPIBackend) SetHead(number uint64) {
	b.xcb.protocolManager.downloader.Cancel()
	b.xcb.blockchain.SetHead(number)
}

func (b *XcbAPIBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		block := b.xcb.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if number == rpc.LatestBlockNumber {
		return b.xcb.blockchain.CurrentBlock().Header(), nil
	}
	return b.xcb.blockchain.GetHeaderByNumber(uint64(number)), nil
}

func (b *XcbAPIBackend) HeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Header, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.HeaderByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header := b.xcb.blockchain.GetHeaderByHash(hash)
		if header == nil {
			return nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && b.xcb.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, errors.New("hash is not currently canonical")
		}
		return header, nil
	}
	return nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *XcbAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.xcb.blockchain.GetHeaderByHash(hash), nil
}

func (b *XcbAPIBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		block := b.xcb.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if number == rpc.LatestBlockNumber {
		return b.xcb.blockchain.CurrentBlock(), nil
	}
	return b.xcb.blockchain.GetBlockByNumber(uint64(number)), nil
}

func (b *XcbAPIBackend) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.xcb.blockchain.GetBlockByHash(hash), nil
}

func (b *XcbAPIBackend) BlockByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Block, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.BlockByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header := b.xcb.blockchain.GetHeaderByHash(hash)
		if header == nil {
			return nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && b.xcb.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, errors.New("hash is not currently canonical")
		}
		block := b.xcb.blockchain.GetBlock(hash, header.Number.Uint64())
		if block == nil {
			return nil, errors.New("header found, but block body is missing")
		}
		return block, nil
	}
	return nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *XcbAPIBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if number == rpc.PendingBlockNumber {
		block, state := b.xcb.miner.Pending()
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
	stateDb, err := b.xcb.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *XcbAPIBackend) StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *types.Header, error) {
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
		if blockNrOrHash.RequireCanonical && b.xcb.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, nil, errors.New("hash is not currently canonical")
		}
		stateDb, err := b.xcb.BlockChain().StateAt(header.Root)
		return stateDb, header, err
	}
	return nil, nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *XcbAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	return b.xcb.blockchain.GetReceiptsByHash(hash), nil
}

func (b *XcbAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	receipts := b.xcb.blockchain.GetReceiptsByHash(hash)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *XcbAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.xcb.blockchain.GetTdByHash(blockHash)
}

func (b *XcbAPIBackend) GetCVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header) (*vm.CVM, func() error, error) {
	vmError := func() error { return nil }

	context := core.NewCVMContext(msg, header, b.xcb.BlockChain(), nil)
	return vm.NewCVM(context, state, b.xcb.blockchain.Config(), *b.xcb.blockchain.GetVMConfig()), vmError, nil
}

func (b *XcbAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.xcb.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *XcbAPIBackend) SubscribePendingLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.xcb.miner.SubscribePendingLogs(ch)
}

func (b *XcbAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.xcb.BlockChain().SubscribeChainEvent(ch)
}

func (b *XcbAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.xcb.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *XcbAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.xcb.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *XcbAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.xcb.BlockChain().SubscribeLogsEvent(ch)
}

func (b *XcbAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.xcb.txPool.AddLocal(signedTx)
}

func (b *XcbAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.xcb.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *XcbAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.xcb.txPool.Get(hash)
}

func (b *XcbAPIBackend) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error) {
	tx, blockHash, blockNumber, index := rawdb.ReadTransaction(b.xcb.ChainDb(), txHash)
	return tx, blockHash, blockNumber, index, nil
}

func (b *XcbAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.xcb.txPool.Nonce(addr), nil
}

func (b *XcbAPIBackend) Stats() (pending int, queued int) {
	return b.xcb.txPool.Stats()
}

func (b *XcbAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.xcb.TxPool().Content()
}

func (b *XcbAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.xcb.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *XcbAPIBackend) Downloader() *downloader.Downloader {
	return b.xcb.Downloader()
}

func (b *XcbAPIBackend) ProtocolVersion() int {
	return b.xcb.XcbVersion()
}

func (b *XcbAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *XcbAPIBackend) ChainDb() xcbdb.Database {
	return b.xcb.ChainDb()
}

func (b *XcbAPIBackend) EventMux() *event.TypeMux {
	return b.xcb.EventMux()
}

func (b *XcbAPIBackend) AccountManager() *accounts.Manager {
	return b.xcb.AccountManager()
}

func (b *XcbAPIBackend) ExtRPCEnabled() bool {
	return b.extRPCEnabled
}

func (b *XcbAPIBackend) RPCEnergyCap() *big.Int {
	return b.xcb.config.RPCEnergyCap
}

func (b *XcbAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.xcb.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *XcbAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.xcb.bloomRequests)
	}
}
