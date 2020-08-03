// Copyright 2018 The go-core Authors
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

// +build none

// This file contains a miner stress test based on the Cryptore consensus engine.
package main

import (
	"github.com/core-coin/eddsa"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"time"

	"github.com/core-coin/go-core/accounts/keystore"
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/fdlimit"
	"github.com/core-coin/go-core/core"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/log"
	"github.com/core-coin/go-core/miner"
	"github.com/core-coin/go-core/node"
	"github.com/core-coin/go-core/p2p"
	"github.com/core-coin/go-core/p2p/enode"
	"github.com/core-coin/go-core/params"
	"github.com/core-coin/go-core/xcb"
	"github.com/core-coin/go-core/xcb/downloader"
)

func main() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
	fdlimit.Raise(2048)

	// Generate a batch of accounts to seal and fund with
	faucets := make([]*eddsa.PrivateKey, 128)
	for i := 0; i < len(faucets); i++ {
		faucets[i], _ = crypto.GenerateKey(rand.Reader)
	}
	// Create an Cryptore network based off of the Devin config
	genesis := makeGenesis(faucets)

	var (
		nodes  []*node.Node
		enodes []*enode.Node
	)
	for i := 0; i < 4; i++ {
		// Start the node and wait until it's up
		node, err := makeMiner(genesis)
		if err != nil {
			panic(err)
		}
		defer node.Close()

		for node.Server().NodeInfo().Ports.Listener == 0 {
			time.Sleep(250 * time.Millisecond)
		}
		// Connect the node to al the previous ones
		for _, n := range enodes {
			node.Server().AddPeer(n)
		}
		// Start tracking the node and it's enode
		nodes = append(nodes, node)
		enodes = append(enodes, node.Server().Self())

		// Inject the signer key and start sealing with it
		store := node.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
		if _, err := store.NewAccount(""); err != nil {
			panic(err)
		}
	}
	// Iterate over all the nodes and start signing with them
	time.Sleep(3 * time.Second)

	for _, node := range nodes {
		var core *xcb.Core
		if err := node.Service(&core); err != nil {
			panic(err)
		}
		if err := core.StartMining(1); err != nil {
			panic(err)
		}
	}
	time.Sleep(3 * time.Second)

	// Start injecting transactions from the faucets like crazy
	nonces := make([]uint64, len(faucets))
	for {
		index := rand.Intn(len(faucets))

		// Fetch the accessor for the relevant signer
		var core *xcb.Core
		if err := nodes[index%len(nodes)].Service(&core); err != nil {
			panic(err)
		}
		// Create a self transaction and inject into the pool
		tx, err := types.SignTx(types.NewTransaction(nonces[index], crypto.PubkeyToAddress(faucets[index].PublicKey), new(big.Int), 21000, big.NewInt(100000000000+rand.Int63n(65536)), nil), types.NewNucleusSigner(genesis.Config.ChainID), faucets[index])
		if err != nil {
			panic(err)
		}
		if err := core.TxPool().AddLocal(tx); err != nil {
			panic(err)
		}
		nonces[index]++

		// Wait if we're too saturated
		if pend, _ := core.TxPool().Stats(); pend > 2048 {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// makeGenesis creates a custom Cryptore genesis block based on some pre-defined
// faucet accounts.
func makeGenesis(faucets []*eddsa.PrivateKey) *core.Genesis {
	genesis := core.DefaultDevinGenesisBlock()
	genesis.Difficulty = params.MinimumDifficulty
	genesis.EnergyLimit = 25000000

	genesis.Config.ChainID = big.NewInt(18)
	genesis.Config.CIP150Hash = common.Hash{}

	genesis.Alloc = core.GenesisAlloc{}
	for _, faucet := range faucets {
		genesis.Alloc[crypto.PubkeyToAddress(faucet.PublicKey)] = core.GenesisAccount{
			Balance: new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil),
		}
	}
	return genesis
}

func makeMiner(genesis *core.Genesis) (*node.Node, error) {
	// Define the basic configurations for the Core node
	datadir, _ := ioutil.TempDir("", "")

	config := &node.Config{
		Name:    "gcore",
		Version: params.Version,
		DataDir: datadir,
		P2P: p2p.Config{
			ListenAddr:  "0.0.0.0:0",
			NoDiscovery: true,
			MaxPeers:    25,
		},
		NoUSB:             true,
		UseLightweightKDF: true,
	}
	// Start the node and configure a full Core node on it
	stack, err := node.New(config)
	if err != nil {
		return nil, err
	}
	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return xcb.New(ctx, &xcb.Config{
			Genesis:         genesis,
			NetworkId:       genesis.Config.ChainID.Uint64(),
			SyncMode:        downloader.FullSync,
			DatabaseCache:   256,
			DatabaseHandles: 256,
			TxPool:          core.DefaultTxPoolConfig,
			GPO:             xcb.DefaultConfig.GPO,
			Cryptore:        xcb.DefaultConfig.Cryptore,
			Miner: miner.Config{
				EnergyFloor: genesis.EnergyLimit * 9 / 10,
				EnergyCeil:  genesis.EnergyLimit * 11 / 10,
				EnergyPrice: big.NewInt(1),
				Recommit:    time.Second,
			},
		})
	}); err != nil {
		return nil, err
	}
	// Start the node and return if successful
	return stack, stack.Start()
}
