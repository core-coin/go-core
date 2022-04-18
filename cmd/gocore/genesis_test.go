// Copyright 2016 by the Authors
// This file is part of go-core.
//
// go-core is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-core is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-core. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var customGenesisTests = []struct {
	genesis string
	query   string
	result  string
}{
	// Plain genesis file without anything extra
	{
		genesis: `{
			"alloc"      : {},
			"coinbase"   : "0x0000000000000000000000000000000000000000",
			"difficulty" : "0x20000",
			"extraData"  : "",
			"energyLimit"   : "0x2fefd8",
			"nonce"      : "0x0000000000000043",
			"parentHash" : "0x0000000000000000000000000000000000000000000000000000000000000000",
			"timestamp"  : "0x00"
		}`,
		query:  "xcb.getBlock(0).nonce",
		result: "0x0000000000000043",
	},
	// Genesis file with an empty chain configuration (ensure missing fields work)
	{
		genesis: `{
			"alloc"      : {},
			"coinbase"   : "0x0000000000000000000000000000000000000000",
			"difficulty" : "0x20000",
			"extraData"  : "",
			"energyLimit"   : "0x2fefd8",
			"nonce"      : "0x0000000000000043",
			"parentHash" : "0x0000000000000000000000000000000000000000000000000000000000000000",
			"timestamp"  : "0x00",
			"config"     : {}
		}`,
		query:  "xcb.getBlock(0).nonce",
		result: "0x0000000000000043",
	},
	// Genesis file with specific chain configurations
	{
		genesis: `{
			"alloc"       : {},
			"coinbase"    : "0x0000000000000000000000000000000000000000",
			"difficulty"  : "0x20000",
			"extraData"   : "",
			"energyLimit" : "0x2fefd8",
			"nonce"       : "0x0000000000000043",
			"parentHash"  : "0x0000000000000000000000000000000000000000000000000000000000000000",
			"timestamp"   : "0x00",
			"config"      : {},
		}`,
		query:  "xcb.getBlock(0).nonce",
		result: "0x0000000000000043",
	},
}

// Tests that initializing Gocore with a custom genesis block and chain definitions
// work properly.
func TestCustomGenesis(t *testing.T) {
	for i, tt := range customGenesisTests {
		// Create a temporary data directory to use and inspect later
		datadir := tmpdir(t)
		defer os.RemoveAll(datadir)

		// Initialize the data directory with the custom genesis block
		json := filepath.Join(datadir, "genesis.json")
		if err := ioutil.WriteFile(json, []byte(tt.genesis), 0600); err != nil {
			t.Fatalf("test %d: failed to write genesis file: %v", i, err)
		}
		runGocore(t, "--nousb", "--datadir", datadir, "init", json).WaitExit()

		// Query the custom genesis block
		gocore := runGocore(t,
			"--nousb",
			"--datadir", datadir, "--maxpeers", "0", "--port", "0",
			"--nodiscover", "--nat", "none", "--ipcdisable",
			"--exec", tt.query, "console")
		gocore.ExpectRegexp(tt.result)
		gocore.ExpectExit()
	}
}
