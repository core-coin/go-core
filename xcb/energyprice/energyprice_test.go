package energyprice

import (
	"context"
	"github.com/core-coin/go-core/accounts"
	"github.com/core-coin/go-core/core/bloombits"
	"github.com/core-coin/go-core/core/state"
	"github.com/core-coin/go-core/event"
	"github.com/core-coin/go-core/xcb/downloader"
	"github.com/core-coin/go-core/xcbdb"
	eddsa "github.com/core-coin/go-goldilocks"
	"math"
	"math/big"
	"testing"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/consensus/cryptore"
	"github.com/core-coin/go-core/core"
	"github.com/core-coin/go-core/core/rawdb"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/core/vm"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/params"
	"github.com/core-coin/go-core/rpc"
)

type testBackend struct {
	chain *core.BlockChain
}

func (b *testBackend) Downloader() *downloader.Downloader {
	panic("implement me")
}

func (b *testBackend) ProtocolVersion() int {
	panic("implement me")
}

func (b *testBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	panic("implement me")
}

func (b *testBackend) ChainDb() xcbdb.Database {
	panic("implement me")
}

func (b *testBackend) AccountManager() *accounts.Manager {
	panic("implement me")
}

func (b *testBackend) ExtRPCEnabled() bool {
	panic("implement me")
}

func (b *testBackend) RPCEnergyCap() *big.Int {
	panic("implement me")
}

func (b *testBackend) SetHead(number uint64) {
	panic("implement me")
}

func (b *testBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	panic("implement me")
}

func (b *testBackend) HeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Header, error) {
	panic("implement me")
}

func (b *testBackend) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	panic("implement me")
}

func (b *testBackend) BlockByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Block, error) {
	panic("implement me")
}

func (b *testBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	panic("implement me")
}

func (b *testBackend) StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *types.Header, error) {
	panic("implement me")
}

func (b *testBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	panic("implement me")
}

func (b *testBackend) GetTd(hash common.Hash) *big.Int {
	panic("implement me")
}

func (b *testBackend) GetCVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header) (*vm.CVM, func() error, error) {
	panic("implement me")
}

func (b *testBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	panic("implement me")
}

func (b *testBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	panic("implement me")
}

func (b *testBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	panic("implement me")
}

func (b *testBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	panic("implement me")
}

func (b *testBackend) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error) {
	panic("implement me")
}

func (b *testBackend) GetPoolTransactions() (types.Transactions, error) {
	panic("implement me")
}

func (b *testBackend) GetPoolTransaction(txHash common.Hash) *types.Transaction {
	panic("implement me")
}

func (b *testBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	panic("implement me")
}

func (b *testBackend) Stats() (pending int, queued int) {
	panic("implement me")
}

func (b *testBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	panic("implement me")
}

func (b *testBackend) SubscribeNewTxsEvent(events chan<- core.NewTxsEvent) event.Subscription {
	panic("implement me")
}

func (b *testBackend) BloomStatus() (uint64, uint64) {
	panic("implement me")
}

func (b *testBackend) GetLogs(ctx context.Context, blockHash common.Hash) ([][]*types.Log, error) {
	panic("implement me")
}

func (b *testBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	panic("implement me")
}

func (b *testBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	panic("implement me")
}

func (b *testBackend) SubscribePendingLogsEvent(ch chan<- []*types.Log) event.Subscription {
	panic("implement me")
}

func (b *testBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	panic("implement me")
}

func (b *testBackend) CurrentBlock() *types.Block {
	panic("implement me")
}

func (b *testBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error) {
	if number == rpc.LatestBlockNumber {
		return b.chain.CurrentBlock().Header(), nil
	}
	return b.chain.GetHeaderByNumber(uint64(number)), nil
}

func (b *testBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error) {
	if number == rpc.LatestBlockNumber {
		return b.chain.CurrentBlock(), nil
	}
	return b.chain.GetBlockByNumber(uint64(number)), nil
}

func (b *testBackend) ChainConfig() *params.ChainConfig {
	return b.chain.Config()
}

func newTestBackend(t *testing.T) *testBackend {
	var (
		key, _ = crypto.HexToEDDSA("856a9af6b0b651dd2f43b5e12193652ec1701c4da6f1c0d2a366ac4b9dabc9433ef09e41ca129552bd2c029086d9b03604de872a3b3432041f")
		pub    = eddsa.Ed448DerivePublicKey(*key)
		addr   = crypto.PubkeyToAddress(pub)
		gspec  = &core.Genesis{
			Config: params.TestChainConfig,
			Alloc:  core.GenesisAlloc{addr: {Balance: big.NewInt(math.MaxInt64)}},
		}
		signer = types.NewNucleusSigner(gspec.Config.NetworkID)
	)
	engine := cryptore.NewFaker()
	db := rawdb.NewMemoryDatabase()
	genesis, _ := gspec.Commit(db)

	addr, err := common.HexToAddress("cb79c0b7afce8c1aa50eb259d87fdf72f6220f173ab6")

	// Generate testing blocks
	blocks, _ := core.GenerateChain(params.TestChainConfig, genesis, engine, db, 32, func(i int, b *core.BlockGen) {
		b.SetCoinbase(common.Address{1})

		tx, err := types.SignTx(types.NewTransaction(b.TxNonce(addr), addr, big.NewInt(100), 21000, big.NewInt(int64(i+1)*params.Nucle), nil), signer, key)
		if err != nil {
			t.Fatalf("failed to create tx: %v", err)
		}
		b.AddTx(tx)
	})
	// Construct testing chain
	diskdb := rawdb.NewMemoryDatabase()
	gspec.Commit(diskdb)
	chain, err := core.NewBlockChain(diskdb, nil, params.TestChainConfig, engine, vm.Config{}, nil)
	if err != nil {
		t.Fatalf("Failed to create local chain, %v", err)
	}
	chain.InsertChain(blocks)
	return &testBackend{chain: chain}
}

func (b *testBackend) CurrentHeader() *types.Header {
	return b.chain.CurrentHeader()
}

func (b *testBackend) GetBlockByNumber(number uint64) *types.Block {
	return b.chain.GetBlockByNumber(number)
}

func TestSuggestPrice(t *testing.T) {
	config := Config{
		Blocks:     3,
		Percentile: 60,
		Default:    big.NewInt(params.Nucle),
	}
	backend := newTestBackend(t)
	oracle := NewOracle(backend, config)

	// The energy price sampled is: 32G, 31G, 30G, 29G, 28G, 27G
	got, err := oracle.SuggestPrice(context.Background())
	if err != nil {
		t.Fatalf("Failed to retrieve recommended energy price: %v", err)
	}
	expect := big.NewInt(params.Nucle * int64(31))
	if got.Cmp(expect) != 0 {
		t.Fatalf("energy price mismatch, want %d, got %d", expect, got)
	}
}
