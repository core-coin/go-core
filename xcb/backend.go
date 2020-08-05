// Copyright 2014 The go-core Authors
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

// Package xcb implements the Core protocol.
package xcb

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/core-coin/go-core/accounts"
	"github.com/core-coin/go-core/accounts/abi/bind"
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/hexutil"
	"github.com/core-coin/go-core/consensus"
	"github.com/core-coin/go-core/consensus/clique"
	"github.com/core-coin/go-core/consensus/cryptore"
	"github.com/core-coin/go-core/core"
	"github.com/core-coin/go-core/core/bloombits"
	"github.com/core-coin/go-core/core/rawdb"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/core/vm"
	"github.com/core-coin/go-core/event"
	"github.com/core-coin/go-core/internal/xcbapi"
	"github.com/core-coin/go-core/log"
	"github.com/core-coin/go-core/miner"
	"github.com/core-coin/go-core/node"
	"github.com/core-coin/go-core/p2p"
	"github.com/core-coin/go-core/p2p/enode"
	"github.com/core-coin/go-core/p2p/enr"
	"github.com/core-coin/go-core/params"
	"github.com/core-coin/go-core/rlp"
	"github.com/core-coin/go-core/rpc"
	"github.com/core-coin/go-core/xcb/downloader"
	"github.com/core-coin/go-core/xcb/energyprice"
	"github.com/core-coin/go-core/xcb/filters"
	"github.com/core-coin/go-core/xcbdb"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	APIs() []rpc.API
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
	SetContractBackend(bind.ContractBackend)
}

// Core implements the Core full node service.
type Core struct {
	config *Config

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer
	dialCandiates   enode.Iterator

	// DB interfaces
	chainDb xcbdb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests     chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer      *core.ChainIndexer             // Bloom indexer operating during block imports
	closeBloomHandler chan struct{}

	APIBackend *XcbAPIBackend

	miner       *miner.Miner
	energyPrice *big.Int
	corebase    common.Address

	networkID     uint64
	netRPCService *xcbapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. energy price and corebase)
}

func (s *Core) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// SetClient sets a rpc client which connecting to our local node.
func (s *Core) SetContractBackend(backend bind.ContractBackend) {
	// Pass the rpc client to les server if it is enabled.
	if s.lesServer != nil {
		s.lesServer.SetContractBackend(backend)
	}
}

// New creates a new Core object (including the
// initialisation of the common Core object)
func New(ctx *node.ServiceContext, config *Config) (*Core, error) {
	// Ensure configuration values are compatible and sane
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run xcb.Core in light sync mode, use les.LightCore")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	if config.Miner.EnergyPrice == nil || config.Miner.EnergyPrice.Cmp(common.Big0) <= 0 {
		log.Warn("Sanitizing invalid miner energy price", "provided", config.Miner.EnergyPrice, "updated", DefaultConfig.Miner.EnergyPrice)
		config.Miner.EnergyPrice = new(big.Int).Set(DefaultConfig.Miner.EnergyPrice)
	}
	if config.NoPruning && config.TrieDirtyCache > 0 {
		config.TrieCleanCache += config.TrieDirtyCache * 3 / 5
		config.SnapshotCache += config.TrieDirtyCache * 3 / 5
		config.TrieDirtyCache = 0
	}
	log.Info("Allocated trie memory caches", "clean", common.StorageSize(config.TrieCleanCache)*1024*1024, "dirty", common.StorageSize(config.TrieDirtyCache)*1024*1024)

	// Assemble the Core object
	chainDb, err := ctx.OpenDatabaseWithFreezer("chaindata", config.DatabaseCache, config.DatabaseHandles, config.DatabaseFreezer, "xcb/db/chaindata/")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	xcb := &Core{
		config:            config,
		chainDb:           chainDb,
		eventMux:          ctx.EventMux,
		accountManager:    ctx.AccountManager,
		engine:            CreateConsensusEngine(ctx, chainConfig, &config.Cryptore, config.Miner.Notify, config.Miner.Noverify, chainDb),
		closeBloomHandler: make(chan struct{}),
		networkID:         config.NetworkId,
		energyPrice:       config.Miner.EnergyPrice,
		corebase:          config.Miner.Corebase,
		bloomRequests:     make(chan chan *bloombits.Retrieval),
		bloomIndexer:      NewBloomIndexer(chainDb, params.BloomBitsBlocks, params.BloomConfirms),
	}

	bcVersion := rawdb.ReadDatabaseVersion(chainDb)
	var dbVer = "<nil>"
	if bcVersion != nil {
		dbVer = fmt.Sprintf("%d", *bcVersion)
	}
	log.Info("Initialising Core protocol", "versions", ProtocolVersions, "network", config.NetworkId, "dbversion", dbVer)

	if !config.SkipBcVersionCheck {
		if bcVersion != nil && *bcVersion > core.BlockChainVersion {
			return nil, fmt.Errorf("database version is v%d, Gocore %s only supports v%d", *bcVersion, params.VersionWithMeta, core.BlockChainVersion)
		} else if bcVersion == nil || *bcVersion < core.BlockChainVersion {
			log.Warn("Upgrade blockchain database version", "from", dbVer, "to", core.BlockChainVersion)
			rawdb.WriteDatabaseVersion(chainDb, core.BlockChainVersion)
		}
	}
	var (
		vmConfig = vm.Config{
			EnablePreimageRecording: config.EnablePreimageRecording,
			EWASMInterpreter:        config.EWASMInterpreter,
			CVMInterpreter:          config.CVMInterpreter,
		}
		cacheConfig = &core.CacheConfig{
			TrieCleanLimit:      config.TrieCleanCache,
			TrieCleanNoPrefetch: config.NoPrefetch,
			TrieDirtyLimit:      config.TrieDirtyCache,
			TrieDirtyDisabled:   config.NoPruning,
			TrieTimeLimit:       config.TrieTimeout,
			SnapshotLimit:       config.SnapshotCache,
		}
	)
	xcb.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, chainConfig, xcb.engine, vmConfig, xcb.shouldPreserve)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		xcb.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	xcb.bloomIndexer.Start(xcb.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	xcb.txPool = core.NewTxPool(config.TxPool, chainConfig, xcb.blockchain)

	// Permit the downloader to use the trie cache allowance during fast sync
	cacheLimit := cacheConfig.TrieCleanLimit + cacheConfig.TrieDirtyLimit + cacheConfig.SnapshotLimit
	checkpoint := config.Checkpoint
	if checkpoint == nil {
		checkpoint = params.TrustedCheckpoints[genesisHash]
	}
	if xcb.protocolManager, err = NewProtocolManager(chainConfig, checkpoint, config.SyncMode, config.NetworkId, xcb.eventMux, xcb.txPool, xcb.engine, xcb.blockchain, chainDb, cacheLimit, config.Whitelist); err != nil {
		return nil, err
	}
	xcb.miner = miner.New(xcb, &config.Miner, chainConfig, xcb.EventMux(), xcb.engine, xcb.isLocalBlock)
	xcb.miner.SetExtra(makeExtraData(config.Miner.ExtraData))

	xcb.APIBackend = &XcbAPIBackend{ctx.ExtRPCEnabled(), xcb, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.Miner.EnergyPrice
	}
	xcb.APIBackend.gpo = energyprice.NewOracle(xcb.APIBackend, gpoParams)

	xcb.dialCandiates, err = xcb.setupDiscovery(&ctx.Config.P2P)
	if err != nil {
		return nil, err
	}

	return xcb, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"gocore",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Core service
func CreateConsensusEngine(ctx *node.ServiceContext, chainConfig *params.ChainConfig, config *cryptore.Config, notify []string, noverify bool, db xcbdb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch config.PowMode {
	case cryptore.ModeFake:
		log.Warn("Cryptore used in fake mode")
		return cryptore.NewFaker()
	case cryptore.ModeTest:
		log.Warn("Cryptore used in test mode")
		return cryptore.NewTester(nil, noverify)
	case cryptore.ModeShared:
		log.Warn("Cryptore used in shared mode")
		return cryptore.NewShared()
	default:
		engine := cryptore.New(cryptore.Config{}, notify, noverify)
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs return the collection of RPC services the core package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Core) APIs() []rpc.API {
	apis := xcbapi.GetAPIs(s.APIBackend)

	// Append any APIs exposed explicitly by the les server
	if s.lesServer != nil {
		apis = append(apis, s.lesServer.APIs()...)
	}
	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append any APIs exposed explicitly by the les server
	if s.lesServer != nil {
		apis = append(apis, s.lesServer.APIs()...)
	}

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "xcb",
			Version:   "1.0",
			Service:   NewPublicCoreAPI(s),
			Public:    true,
		}, {
			Namespace: "xcb",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "xcb",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "xcb",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.APIBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Core) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Core) Corebase() (eb common.Address, err error) {
	s.lock.RLock()
	corebase := s.corebase
	s.lock.RUnlock()

	if corebase != (common.Address{}) {
		return corebase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			corebase := accounts[0].Address

			s.lock.Lock()
			s.corebase = corebase
			s.lock.Unlock()

			log.Info("Corebase automatically configured", "address", corebase)
			return corebase, nil
		}
	}
	return common.Address{}, fmt.Errorf("corebase must be explicitly specified")
}

// isLocalBlock checks whether the specified block is mined
// by local miner accounts.
//
// We regard two types of accounts as local miner account: corebase
// and accounts specified via `txpool.locals` flag.
func (s *Core) isLocalBlock(block *types.Block) bool {
	author, err := s.engine.Author(block.Header())
	if err != nil {
		log.Warn("Failed to retrieve block author", "number", block.NumberU64(), "hash", block.Hash(), "err", err)
		return false
	}
	// Check whether the given address is corebase.
	s.lock.RLock()
	corebase := s.corebase
	s.lock.RUnlock()
	if author == corebase {
		return true
	}
	// Check whether the given address is specified by `txpool.local`
	// CLI flag.
	for _, account := range s.config.TxPool.Locals {
		if account == author {
			return true
		}
	}
	return false
}

// shouldPreserve checks whether we should preserve the given block
// during the chain reorg depending on whether the author of block
// is a local account.
func (s *Core) shouldPreserve(block *types.Block) bool {
	// The reason we need to disable the self-reorg preserving for clique
	// is it can be probable to introduce a deadlock.
	//
	// e.g. If there are 7 available signers
	//
	// r1   A
	// r2     B
	// r3       C
	// r4         D
	// r5   A      [X] F G
	// r6    [X]
	//
	// In the round5, the inturn signer E is offline, so the worst case
	// is A, F and G sign the block of round5 and reject the block of opponents
	// and in the round6, the last available signer B is offline, the whole
	// network is stuck.
	if _, ok := s.engine.(*clique.Clique); ok {
		return false
	}
	return s.isLocalBlock(block)
}

// SetCorebase sets the mining reward address.
func (s *Core) SetCorebase(corebase common.Address) {
	s.lock.Lock()
	s.corebase = corebase
	s.lock.Unlock()

	s.miner.SetCorebase(corebase)
}

// StartMining starts the miner with the given number of CPU threads. If mining
// is already running, this method adjust the number of threads allowed to use
// and updates the minimum price required by the transaction pool.
func (s *Core) StartMining(threads int) error {
	// Update the thread count within the consensus engine
	type threaded interface {
		SetThreads(threads int)
	}
	if th, ok := s.engine.(threaded); ok {
		log.Info("Updated mining threads", "threads", threads)
		if threads == 0 {
			threads = -1 // Disable the miner from within
		}
		th.SetThreads(threads)
	}
	// If the miner was not running, initialize it
	if !s.IsMining() {
		// Propagate the initial price point to the transaction pool
		s.lock.RLock()
		price := s.energyPrice
		s.lock.RUnlock()
		s.txPool.SetEnergyPrice(price)

		// Configure the local mining address
		eb, err := s.Corebase()
		if err != nil {
			log.Error("Cannot start mining without corebase", "err", err)
			return fmt.Errorf("corebase missing: %v", err)
		}
		if clique, ok := s.engine.(*clique.Clique); ok {
			wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
			if wallet == nil || err != nil {
				log.Error("Corebase account unavailable locally", "err", err)
				return fmt.Errorf("signer missing: %v", err)
			}
			clique.Authorize(eb, wallet.SignData)
		}
		// If mining is started, we can disable the transaction rejection mechanism
		// introduced to speed sync times.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)

		go s.miner.Start(eb)
	}
	return nil
}

// StopMining terminates the miner, both at the consensus engine level as well as
// at the block creation level.
func (s *Core) StopMining() {
	// Update the thread count within the consensus engine
	type threaded interface {
		SetThreads(threads int)
	}
	if th, ok := s.engine.(threaded); ok {
		th.SetThreads(-1)
	}
	// Stop the block creating itself
	s.miner.Stop()
}

func (s *Core) IsMining() bool      { return s.miner.Mining() }
func (s *Core) Miner() *miner.Miner { return s.miner }

func (s *Core) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Core) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Core) TxPool() *core.TxPool               { return s.txPool }
func (s *Core) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Core) Engine() consensus.Engine           { return s.engine }
func (s *Core) ChainDb() xcbdb.Database            { return s.chainDb }
func (s *Core) IsListening() bool                  { return true } // Always listening
func (s *Core) XcbVersion() int                    { return int(ProtocolVersions[0]) }
func (s *Core) NetVersion() uint64                 { return s.networkID }
func (s *Core) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *Core) Synced() bool                       { return atomic.LoadUint32(&s.protocolManager.acceptTxs) == 1 }
func (s *Core) ArchiveMode() bool                  { return s.config.NoPruning }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Core) Protocols() []p2p.Protocol {
	protos := make([]p2p.Protocol, len(ProtocolVersions))
	for i, vsn := range ProtocolVersions {
		protos[i] = s.protocolManager.makeProtocol(vsn)
		protos[i].Attributes = []enr.Entry{s.currentXcbEntry()}
		protos[i].DialCandidates = s.dialCandiates
	}
	if s.lesServer != nil {
		protos = append(protos, s.lesServer.Protocols()...)
	}
	return protos
}

// Start implements node.Service, starting all internal goroutines needed by the
// Core protocol implementation.
func (s *Core) Start(srvr *p2p.Server) error {
	s.startXcbEntryUpdate(srvr.LocalNode())

	// Start the bloom bits servicing goroutines
	s.startBloomHandlers(params.BloomBitsBlocks)

	// Start the RPC service
	s.netRPCService = xcbapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		if s.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= s.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Core protocol.
func (s *Core) Stop() error {
	// Stop all the peer-related stuff first.
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}

	// Then stop everything else.
	s.bloomIndexer.Close()
	close(s.closeBloomHandler)
	s.txPool.Stop()
	s.miner.Stop()
	s.blockchain.Stop()
	s.engine.Close()
	s.chainDb.Close()
	s.eventMux.Stop()
	return nil
}
