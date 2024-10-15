// Copyright 2023 by the Authors
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

// Contains all the wrappers from the node package to support client side node
// management on mobile platforms.

package gocore

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/core-coin/go-core/v2/xcbstats"

	"github.com/core-coin/go-core/v2/core"
	"github.com/core-coin/go-core/v2/internal/debug"
	"github.com/core-coin/go-core/v2/les"
	"github.com/core-coin/go-core/v2/node"
	"github.com/core-coin/go-core/v2/p2p"
	"github.com/core-coin/go-core/v2/p2p/nat"
	"github.com/core-coin/go-core/v2/params"
	"github.com/core-coin/go-core/v2/xcb"
	"github.com/core-coin/go-core/v2/xcb/downloader"
	"github.com/core-coin/go-core/v2/xcbclient"
)

// NodeConfig represents the collection of configuration values to fine tune the Gocore
// node embedded into a mobile process. The available values are a subset of the
// entire API provided by go-core to reduce the maintenance surface and dev
// complexity.
type NodeConfig struct {
	// Bootstrap nodes used to establish connectivity with the rest of the network.
	BootstrapNodes *Enodes

	// MaxPeers is the maximum number of peers that can be connected. If this is
	// set to zero, then only the configured static and trusted peers can connect.
	MaxPeers int

	// CoreEnabled specifies whether the node should run the Core protocol.
	CoreEnabled bool

	// CoreNetworkID is the network identifier used by the Core protocol to
	// decide if remote peers should be accepted or not.
	CoreNetworkID int64 // uint64 in truth, but Java can't handle that...

	// CoreGenesis is the genesis JSON to use to seed the blockchain with. An
	// empty genesis state is equivalent to using the mainnet's state.
	CoreGenesis string

	// CoreDatabaseCache is the system memory in MB to allocate for database caching.
	// A minimum of 16MB is always reserved.
	CoreDatabaseCache int

	// CoreNetStats is a netstats connection string to use to report various
	// chain, transaction and node stats to a monitoring server.
	//
	// It has the form "nodename:secret@host:port"
	CoreNetStats string

	// Listening address of pprof server.
	PprofAddress string
}

// defaultNodeConfig contains the default node configuration values to use if all
// or some fields are missing from the user's specified list.
var defaultNodeConfig = &NodeConfig{
	BootstrapNodes:    FoundationBootnodes(),
	MaxPeers:          25,
	CoreEnabled:       true,
	CoreNetworkID:     1,
	CoreDatabaseCache: 16,
}

// NewNodeConfig creates a new node option set, initialized to the default values.
func NewNodeConfig() *NodeConfig {
	config := *defaultNodeConfig
	return &config
}

// AddBootstrapNode adds an additional bootstrap node to the node config.
func (conf *NodeConfig) AddBootstrapNode(node *Enode) {
	conf.BootstrapNodes.Append(node)
}

// EncodeJSON encodes a NodeConfig into a JSON data dump.
func (conf *NodeConfig) EncodeJSON() (string, error) {
	data, err := json.Marshal(conf)
	return string(data), err
}

// String returns a printable representation of the node config.
func (conf *NodeConfig) String() string {
	return encodeOrError(conf)
}

// Node represents a Gocore Core node instance.
type Node struct {
	node *node.Node
}

// NewNode creates and configures a new Gocore node.
func NewNode(datadir string, config *NodeConfig) (stack *Node, _ error) {
	// If no or partial configurations were specified, use defaults
	if config == nil {
		config = NewNodeConfig()
	}
	if config.MaxPeers == 0 {
		config.MaxPeers = defaultNodeConfig.MaxPeers
	}
	if config.BootstrapNodes == nil || config.BootstrapNodes.Size() == 0 {
		config.BootstrapNodes = defaultNodeConfig.BootstrapNodes
	}

	if config.PprofAddress != "" {
		debug.StartPProf(config.PprofAddress, true)
	}

	// Create the empty networking stack
	nodeConf := &node.Config{
		Name: clientIdentifier,
		//Version:     params.VersionWithMeta,
		DataDir:     datadir,
		KeyStoreDir: filepath.Join(datadir, "keystore"), // Mobile should never use internal keystores!
		P2P: p2p.Config{
			NoDiscovery:      true,
			DiscoveryV5:      true,
			BootstrapNodesV5: config.BootstrapNodes.nodes,
			ListenAddr:       ":0",
			NAT:              nat.Any(),
			MaxPeers:         config.MaxPeers,
		},
	}

	rawStack, err := node.New(nodeConf)
	if err != nil {
		return nil, err
	}

	var genesis *core.Genesis
	if config.CoreGenesis != "" {
		// Parse the user supplied genesis spec if not mainnet
		genesis = new(core.Genesis)
		if err := json.Unmarshal([]byte(config.CoreGenesis), genesis); err != nil {
			return nil, fmt.Errorf("invalid genesis spec: %v", err)
		}
		// If we have the Devin testnet, hard code the chain configs too
		if config.CoreGenesis == DevinGenesis() {
			genesis.Config = params.DevinChainConfig
			if config.CoreNetworkID == 1 {
				config.CoreNetworkID = 3
			}
		}
	}
	// Register the Core protocol if requested
	if config.CoreEnabled {
		xcbConf := xcb.DefaultConfig
		xcbConf.Genesis = genesis
		xcbConf.SyncMode = downloader.LightSync
		xcbConf.NetworkId = uint64(config.CoreNetworkID)
		xcbConf.DatabaseCache = config.CoreDatabaseCache
		lesBackend, err := les.New(rawStack, &xcbConf)
		if err != nil {
			return nil, fmt.Errorf("core init: %v", err)
		}
		// If netstats reporting is requested, do it
		if config.CoreNetStats != "" {
			if err := xcbstats.New(rawStack, lesBackend.ApiBackend, lesBackend.Engine(), config.CoreNetStats); err != nil {
				return nil, fmt.Errorf("netstats init: %v", err)
			}
		}
	}
	return &Node{rawStack}, nil
}

// Close terminates a running node along with all it's services, tearing internal state
// down. It is not possible to restart a closed node.
func (n *Node) Close() error {
	return n.node.Close()
}

// Start creates a live P2P node and starts running it.
func (n *Node) Start() error {
	// TODO: recreate the node so it can be started multiple times
	return n.node.Start()
}

// Stop terminates a running node along with all its services. If the node was not started,
// an error is returned. It is not possible to restart a stopped node.
//
// Deprecated: use Close()
func (n *Node) Stop() error {
	return n.node.Close()
}

// GetCoreClient retrieves a client to access the Core subsystem.
func (n *Node) GetCoreClient() (client *CoreClient, _ error) {
	rpc, err := n.node.Attach()
	if err != nil {
		return nil, err
	}
	return &CoreClient{xcbclient.NewClient(rpc)}, nil
}

// GetNodeInfo gathers and returns a collection of metadata known about the host.
func (n *Node) GetNodeInfo() *NodeInfo {
	return &NodeInfo{n.node.Server().NodeInfo()}
}

// GetPeersInfo returns an array of metadata objects describing connected peers.
func (n *Node) GetPeersInfo() *PeerInfos {
	return &PeerInfos{n.node.Server().PeersInfo()}
}
