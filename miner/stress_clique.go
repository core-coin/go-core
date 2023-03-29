// Copyright 2018 by the Authors
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

//go:build none
// +build none

// This file contains a miner stress test based on the Clique consensus engine.
package main

import (
	"bytes"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"time"

	"github.com/core-coin/go-goldilocks"

	"github.com/core-coin/go-core/v2/accounts/keystore"
	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/fdlimit"
	"github.com/core-coin/go-core/v2/core"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/log"
	"github.com/core-coin/go-core/v2/miner"
	"github.com/core-coin/go-core/v2/node"
	"github.com/core-coin/go-core/v2/p2p"
	"github.com/core-coin/go-core/v2/p2p/enode"
	"github.com/core-coin/go-core/v2/params"
	"github.com/core-coin/go-core/v2/xcb"
	"github.com/core-coin/go-core/v2/xcb/downloader"
)

func main() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
	fdlimit.Raise(2048)

	// Generate a batch of accounts to seal and fund with
	faucets := make([]*goldilocks.PrivateKey, 128)
	for i := 0; i < len(faucets); i++ {
		faucets[i], _ = crypto.GenerateKey(crand.Reader)
	}
	sealers := make([]*goldilocks.PrivateKey, 4)
	for i := 0; i < len(sealers); i++ {
		sealers[i], _ = crypto.GenerateKey(crand.Reader)
	}
	// Create a Clique network based off of the Devin config
	genesis := makeGenesis(faucets, sealers)

	var (
		nodes  []*xcb.Core
		enodes []*enode.Node
	)

	for _, sealer := range sealers {
		// Start the node and wait until it's up
		stack, xcbBackend, err := makeSealer(genesis)
		if err != nil {
			panic(err)
		}
		defer stack.Close()

		for stack.Server().NodeInfo().Ports.Listener == 0 {
			time.Sleep(250 * time.Millisecond)
		}
		// Connect the node to all the previous ones
		for _, n := range enodes {
			stack.Server().AddPeer(n)
		}
		// Start tracking the node and its enode
		nodes = append(nodes, xcbBackend)
		enodes = append(enodes, stack.Server().Self())

		// Inject the signer key and start sealing with it
		store := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
		signer, err := store.ImportEDDSA(sealer, "")
		if err != nil {
			panic(err)
		}
		if err := store.Unlock(signer, ""); err != nil {
			panic(err)
		}
	}

	// Iterate over all the nodes and start signing on them
	time.Sleep(3 * time.Second)
	for _, node := range nodes {
		if err := node.StartMining(1); err != nil {
			panic(err)
		}
	}
	time.Sleep(3 * time.Second)

	// Start injecting transactions from the faucet like crazy
	nonces := make([]uint64, len(faucets))
	for {
		// Pick a random signer node
		index := rand.Intn(len(faucets))
		backend := nodes[index%len(nodes)]

		// Create a self transaction and inject into the pool
		tx, err := types.SignTx(types.NewTransaction(nonces[index], faucets[index].Address(), new(big.Int), 21000, big.NewInt(100000000000), nil), types.HomesteadSigner{}, faucets[index])
		if err != nil {
			panic(err)
		}
		if err := backend.TxPool().AddLocal(tx); err != nil {
			panic(err)
		}
		nonces[index]++

		// Wait if we're too saturated
		if pend, _ := backend.TxPool().Stats(); pend > 2048 {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// makeGenesis creates a custom Clique genesis block based on some pre-defined
// signer and faucet accounts.
func makeGenesis(faucets []*goldilocks.PrivateKey, sealers []*goldilocks.PrivateKey) *core.Genesis {
	// Create a Clique network based off of the Devin config
	genesis := core.DefaultDevinGenesisBlock()
	genesis.EnergyLimit = 25000000

	genesis.Config.NetworkID = big.NewInt(18)
	genesis.Config.Clique.Period = 1
	genesis.Config.CIP150Hash = common.Hash{}

	genesis.Alloc = core.GenesisAlloc{}
	for _, faucet := range faucets {
		genesis.Alloc[faucet.Address()] = core.GenesisAccount{
			Balance: new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil),
		}
	}
	// Sort the signers and embed into the extra-data section
	signers := make([]common.Address, len(sealers))
	for i, sealer := range sealers {
		signers[i] = sealer.Address()
	}
	for i := 0; i < len(signers); i++ {
		for j := i + 1; j < len(signers); j++ {
			if bytes.Compare(signers[i][:], signers[j][:]) > 0 {
				signers[i], signers[j] = signers[j], signers[i]
			}
		}
	}
	genesis.ExtraData = make([]byte, 32+len(signers)*common.AddressLength+crypto.ExtendedSignatureLength)
	for i, signer := range signers {
		copy(genesis.ExtraData[32+i*common.AddressLength:], signer[:])
	}
	// Return the genesis block for initialization
	return genesis
}

func makeSealer(genesis *core.Genesis) (*node.Node, *xcb.Core, error) {
	// Define the basic configurations for the Core node
	datadir, _ := ioutil.TempDir("", "")

	config := &node.Config{
		Name:    "gocore",
		Version: params.Version,
		DataDir: datadir,
		P2P: p2p.Config{
			ListenAddr:  "0.0.0.0:0",
			NoDiscovery: true,
			MaxPeers:    25,
		},
	}
	// Start the node and configure a full Core node on it
	stack, err := node.New(config)
	if err != nil {
		return nil, nil, err
	}
	// Create and register the backend
	xcbBackend, err := xcb.New(stack, &xcb.Config{
		Genesis:         genesis,
		NetworkId:       genesis.Config.NetworkID.Uint64(),
		SyncMode:        downloader.FullSync,
		DatabaseCache:   256,
		DatabaseHandles: 256,
		TxPool:          core.DefaultTxPoolConfig,
		GPO:             xcb.DefaultConfig.GPO,
		Miner: miner.Config{
			EnergyFloor: genesis.EnergyLimit * 9 / 10,
			EnergyCeil:  genesis.EnergyLimit * 11 / 10,
			EnergyPrice: big.NewInt(1),
			Recommit:    time.Second,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	err = stack.Start()
	return stack, xcbBackend, err
}
