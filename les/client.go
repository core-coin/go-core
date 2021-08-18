// Copyright 2019 by the Authors
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

// Package les implements the Light Core Subprotocol.
package les

import (
	"fmt"
	"math/big"

	"github.com/core-coin/go-core/accounts"
	"github.com/core-coin/go-core/accounts/abi/bind"
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/hexutil"
	"github.com/core-coin/go-core/common/mclock"
	"github.com/core-coin/go-core/consensus"
	"github.com/core-coin/go-core/core"
	"github.com/core-coin/go-core/core/bloombits"
	"github.com/core-coin/go-core/core/rawdb"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/event"
	"github.com/core-coin/go-core/internal/xcbapi"
	"github.com/core-coin/go-core/les/checkpointoracle"
	"github.com/core-coin/go-core/light"
	"github.com/core-coin/go-core/log"
	"github.com/core-coin/go-core/node"
	"github.com/core-coin/go-core/p2p"
	"github.com/core-coin/go-core/p2p/enode"
	"github.com/core-coin/go-core/params"
	"github.com/core-coin/go-core/rpc"
	"github.com/core-coin/go-core/xcb"
	"github.com/core-coin/go-core/xcb/downloader"
	"github.com/core-coin/go-core/xcb/energyprice"
	"github.com/core-coin/go-core/xcb/filters"
)

type LightCore struct {
	lesCommons

	peers      *serverPeerSet
	reqDist    *requestDistributor
	retriever  *retrieveManager
	odr        *LesOdr
	relay      *lesTxRelay
	handler    *clientHandler
	txPool     *light.TxPool
	blockchain *light.LightChain
	serverPool *serverPool
	pruner         *pruner

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend     *LesApiBackend
	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager
	netRPCService  *xcbapi.PublicNetAPI
}

func New(ctx *node.ServiceContext, config *xcb.Config) (*LightCore, error) {
	chainDb, err := ctx.OpenDatabase("lightchaindata", config.DatabaseCache, config.DatabaseHandles, "xcb/db/chaindata/")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}

	chainConfig.NetworkID = big.NewInt(int64(config.NetworkId))

	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newServerPeerSet()
	lxcb := &LightCore{
		lesCommons: lesCommons{
			genesis:     genesisHash,
			config:      config,
			chainConfig: chainConfig,
			iConfig:     light.DefaultClientIndexerConfig,
			chainDb:     chainDb,
			closeCh:     make(chan struct{}),
		},
		peers:          peers,
		eventMux:       ctx.EventMux,
		reqDist:        newRequestDistributor(peers, &mclock.System{}),
		accountManager: ctx.AccountManager,
		engine:         xcb.CreateConsensusEngine(ctx, chainConfig, &config.Cryptore, nil, false, chainDb),
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   xcb.NewBloomIndexer(chainDb, params.BloomBitsBlocksClient, params.HelperTrieConfirmations),
		serverPool:     newServerPool(chainDb, config.UltraLightServers),
	}
	lxcb.retriever = newRetrieveManager(peers, lxcb.reqDist, lxcb.serverPool)
	lxcb.relay = newLesTxRelay(peers, lxcb.retriever)

	lxcb.odr = NewLesOdr(chainDb, light.DefaultClientIndexerConfig, lxcb.retriever)
	lxcb.chtIndexer = light.NewChtIndexer(chainDb, lxcb.odr, params.CHTFrequency, params.HelperTrieConfirmations, config.LightNoPrune)
	lxcb.bloomTrieIndexer = light.NewBloomTrieIndexer(chainDb, lxcb.odr, params.BloomBitsBlocksClient, params.BloomTrieFrequency, config.LightNoPrune)
	lxcb.odr.SetIndexers(lxcb.chtIndexer, lxcb.bloomTrieIndexer, lxcb.bloomIndexer)

	checkpoint := config.Checkpoint
	if checkpoint == nil {
		checkpoint = params.TrustedCheckpoints[genesisHash]
	}
	// Note: NewLightChain adds the trusted checkpoint so it needs an ODR with
	// indexers already set but not started yet
	if lxcb.blockchain, err = light.NewLightChain(lxcb.odr, lxcb.chainConfig, lxcb.engine, checkpoint); err != nil {
		return nil, err
	}
	lxcb.chainReader = lxcb.blockchain
	lxcb.txPool = light.NewTxPool(lxcb.chainConfig, lxcb.blockchain, lxcb.relay)

	// Set up checkpoint oracle.
	oracle := config.CheckpointOracle
	if oracle == nil {
		oracle = params.CheckpointOracles[genesisHash]
	}
	lxcb.oracle = checkpointoracle.New(oracle, lxcb.localCheckpoint)

	// Note: AddChildIndexer starts the update process for the child
	lxcb.bloomIndexer.AddChildIndexer(lxcb.bloomTrieIndexer)
	lxcb.chtIndexer.Start(lxcb.blockchain)
	lxcb.bloomIndexer.Start(lxcb.blockchain)

	lxcb.handler = newClientHandler(config.UltraLightServers, config.UltraLightFraction, checkpoint, lxcb)
	if lxcb.handler.ulc != nil {
		log.Warn("Ultra light client is enabled", "trustedNodes", len(lxcb.handler.ulc.keys), "minTrustedFraction", lxcb.handler.ulc.fraction)
		lxcb.blockchain.DisableCheckFreq()
	}

	// Start a light chain pruner to delete useless historical data.
	lxcb.pruner = newPruner(chainDb, lxcb.chtIndexer, lxcb.bloomTrieIndexer)

	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		lxcb.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	lxcb.ApiBackend = &LesApiBackend{ctx.ExtRPCEnabled(), lxcb, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.Miner.EnergyPrice
	}
	lxcb.ApiBackend.gpo = energyprice.NewOracle(lxcb.ApiBackend, gpoParams)

	return lxcb, nil
}

type LightDummyAPI struct{}

// Corebase is the address that mining rewards will be send to
func (s *LightDummyAPI) Corebase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("mining is not supported in light mode")
}

// Coinbase is the address that mining rewards will be send to (alias for Corebase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("mining is not supported in light mode")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the core package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightCore) APIs() []rpc.API {
	apis := xcbapi.GetAPIs(s.ApiBackend)
	apis = append(apis, s.engine.APIs(s.BlockChain().HeaderChain())...)
	return append(apis, []rpc.API{
		{
			Namespace: "xcb",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "xcb",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.handler.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "xcb",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		}, {
			Namespace: "les",
			Version:   "1.0",
			Service:   NewPrivateLightAPI(&s.lesCommons),
			Public:    false,
		},
	}...)
}

func (s *LightCore) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightCore) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightCore) TxPool() *light.TxPool              { return s.txPool }
func (s *LightCore) Engine() consensus.Engine           { return s.engine }
func (s *LightCore) LesVersion() int                    { return int(ClientProtocolVersions[0]) }
func (s *LightCore) Downloader() *downloader.Downloader { return s.handler.downloader }
func (s *LightCore) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightCore) Protocols() []p2p.Protocol {
	return s.makeProtocols(ClientProtocolVersions, s.handler.runPeer, func(id enode.ID) interface{} {
		if p := s.peers.peer(peerIdToString(id)); p != nil {
			return p.Info()
		}
		return nil
	})
}

// Start implements node.Service, starting all internal goroutines needed by the
// light core protocol implementation.
func (s *LightCore) Start(srvr *p2p.Server) error {
	log.Warn("Light client mode is an experimental feature")

	// Start bloom request workers.
	s.wg.Add(bloomServiceThreads)
	s.startBloomHandlers(params.BloomBitsBlocksClient)

	s.netRPCService = xcbapi.NewPublicNetAPI(srvr, s.config.NetworkId)

	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Core protocol.
func (s *LightCore) Stop() error {
	close(s.closeCh)
	s.peers.close()
	s.reqDist.close()
	s.odr.Stop()
	s.relay.Stop()
	s.bloomIndexer.Close()
	s.chtIndexer.Close()
	s.blockchain.Stop()
	s.handler.stop()
	s.txPool.Stop()
	s.engine.Close()
	s.pruner.close()
	s.eventMux.Stop()
	s.serverPool.stop()
	s.chainDb.Close()
	s.wg.Wait()
	log.Info("Light core stopped")
	return nil
}

// SetClient sets the rpc client and binds the registrar contract.
func (s *LightCore) SetContractBackend(backend bind.ContractBackend) {
	if s.oracle == nil {
		return
	}
	s.oracle.Start(backend)
}
