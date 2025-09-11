// Copyright 2015 by the Authors
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

// Package xcbapi implements the general Core API functions.
package xcbapi

import (
	"context"
	"math/big"

	"github.com/core-coin/go-core/v2/xcbdb"

	"github.com/core-coin/go-core/v2/accounts"
	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/consensus"
	"github.com/core-coin/go-core/v2/core"
	"github.com/core-coin/go-core/v2/core/bloombits"
	"github.com/core-coin/go-core/v2/core/state"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/core/vm"
	"github.com/core-coin/go-core/v2/event"
	"github.com/core-coin/go-core/v2/params"
	"github.com/core-coin/go-core/v2/rpc"
	"github.com/core-coin/go-core/v2/xcb/downloader"
)

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	// General Core API
	Downloader() *downloader.Downloader
	ProtocolVersion() int
	SuggestPrice(ctx context.Context) (*big.Int, error)
	ChainDb() xcbdb.Database
	AccountManager() *accounts.Manager
	ExtRPCEnabled() bool
	RPCEnergyCap() uint64 // global energy cap for xcb_call over rpc: DoS protection
	RPCTxFeeCap() float64 // global tx fee cap for all transaction related APIs

	// Blockchain API
	SetHead(number uint64)
	HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error)
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	HeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Header, error)
	CurrentHeader() *types.Header
	CurrentBlock() *types.Block
	BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error)
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	BlockByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Block, error)
	StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error)
	StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *types.Header, error)
	GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error)
	GetTd(ctx context.Context, hash common.Hash) *big.Int
	GetCVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header) (*vm.CVM, func() error, error)
	SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription
	SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription
	SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription

	// Transaction pool API
	SendTx(ctx context.Context, signedTx *types.Transaction) error
	GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error)
	GetPoolTransactions() (types.Transactions, error)
	GetPoolTransaction(txHash common.Hash) *types.Transaction
	GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error)
	Stats() (pending int, queued int)
	TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions)
	SubscribeNewTxsEvent(chan<- core.NewTxsEvent) event.Subscription

	// Filter API
	BloomStatus() (uint64, uint64)
	GetLogs(ctx context.Context, blockHash common.Hash) ([][]*types.Log, error)
	ServiceFilter(ctx context.Context, session *bloombits.MatcherSession)
	SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription
	SubscribePendingLogsEvent(ch chan<- []*types.Log) event.Subscription
	SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription

	ChainConfig() *params.ChainConfig
	Engine() consensus.Engine

	// Smart Contract API
	CallContract(ctx context.Context, call CallMsg, blockNumber rpc.BlockNumber) ([]byte, error)
}

// CallMsg contains parameters for contract calls.
type CallMsg struct {
	FromAddr         common.Address  // the sender of the 'transaction'
	ToAddr           *common.Address // the destination contract (nil for contract creation)
	EnergyLimit      uint64          // if 0, the call executes with near-infinite energy
	EnergyPriceValue *big.Int        // ore <-> energy exchange ratio
	ValueAmount      *big.Int        // amount of ore sent along with the call
	DataBytes        []byte          // input data, usually an ABI-encoded contract method invocation
}

// Implement core.Message interface
func (m CallMsg) Nonce() uint64         { return 0 }
func (m CallMsg) CheckNonce() bool      { return false }
func (m CallMsg) From() common.Address  { return m.FromAddr }
func (m CallMsg) To() *common.Address   { return m.ToAddr }
func (m CallMsg) EnergyPrice() *big.Int { return m.EnergyPriceValue }
func (m CallMsg) Energy() uint64        { return m.EnergyLimit }
func (m CallMsg) Value() *big.Int       { return m.ValueAmount }
func (m CallMsg) Data() []byte          { return m.DataBytes }

func GetAPIs(apiBackend Backend) []rpc.API {
	nonceLock := new(AddrLocker)
	return []rpc.API{
		{
			Namespace: "xcb",
			Version:   "1.0",
			Service:   NewPublicCoreAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "xcb",
			Version:   "1.0",
			Service:   NewPublicBlockChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "xcb",
			Version:   "1.0",
			Service:   NewPublicTransactionPoolAPI(apiBackend, nonceLock),
			Public:    true,
		}, {
			Namespace: "txpool",
			Version:   "1.0",
			Service:   NewPublicTxPoolAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(apiBackend),
		}, {
			Namespace: "xcb",
			Version:   "1.0",
			Service:   NewPublicAccountAPI(apiBackend.AccountManager()),
			Public:    true,
		}, {
			Namespace: "personal",
			Version:   "1.0",
			Service:   NewPrivateAccountAPI(apiBackend, nonceLock),
			Public:    false,
		},
	}
}
