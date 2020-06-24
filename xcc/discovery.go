// Copyright 2019 The go-core Authors
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
	"github.com/core-coin/go-core/core"
	"github.com/core-coin/go-core/core/forkid"
	"github.com/core-coin/go-core/p2p"
	"github.com/core-coin/go-core/p2p/dnsdisc"
	"github.com/core-coin/go-core/p2p/enode"
	"github.com/core-coin/go-core/rlp"
)

// xccEntry is the "xcc" ENR entry which advertises xcc protocol
// on the discovery network.
type xccEntry struct {
	ForkID forkid.ID // Fork identifier per CIP-2124

	// Ignore additional fields (for forward compatibility).
	Rest []rlp.RawValue `rlp:"tail"`
}

// ENRKey implements enr.Entry.
func (e xccEntry) ENRKey() string {
	return "xcc"
}

// startXccEntryUpdate starts the ENR updater loop.
func (xcc *Core) startXccEntryUpdate(ln *enode.LocalNode) {
	var newHead = make(chan core.ChainHeadEvent, 10)
	sub := xcc.blockchain.SubscribeChainHeadEvent(newHead)

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case <-newHead:
				ln.Set(xcc.currentXccEntry())
			case <-sub.Err():
				// Would be nice to sync with xcc.Stop, but there is no
				// good way to do that.
				return
			}
		}
	}()
}

func (xcc *Core) currentXccEntry() *xccEntry {
	return &xccEntry{ForkID: forkid.NewID(xcc.blockchain)}
}

// setupDiscovery creates the node discovery source for the xcc protocol.
func (xcc *Core) setupDiscovery(cfg *p2p.Config) (enode.Iterator, error) {
	if cfg.NoDiscovery || len(xcc.config.DiscoveryURLs) == 0 {
		return nil, nil
	}
	client := dnsdisc.NewClient(dnsdisc.Config{})
	return client.NewIterator(xcc.config.DiscoveryURLs...)
}
