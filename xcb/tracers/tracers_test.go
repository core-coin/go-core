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

package tracers

import (
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/common/math"
	"github.com/core-coin/go-core/v2/core"
	"github.com/core-coin/go-core/v2/core/rawdb"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/core/vm"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/params"
	"github.com/core-coin/go-core/v2/rlp"
	"github.com/core-coin/go-core/v2/tests"
)

// To generate a new callTracer test, copy paste the makeTest method below into
// a Gocore console and call it with a transaction hash you which to export.

/*
// makeTest generates a callTracer test by running a prestate reassembled and a
// call trace run, assembling all the gathered information into a test case.
var makeTest = function(tx, rewind) {
  // Generate the genesis block from the block, transaction and prestate data
  var block   = xcb.getBlock(xcb.getTransaction(tx).blockHash);
  var genesis = xcb.getBlock(block.parentHash);

  delete genesis.energyUsed;
  delete genesis.logsBloom;
  delete genesis.parentHash;
  delete genesis.receiptsRoot;
  delete genesis.sha3Uncles;
  delete genesis.size;
  delete genesis.transactions;
  delete genesis.transactionsRoot;
  delete genesis.uncles;

  genesis.energyLimit  = genesis.energyLimit.toString();
  genesis.number    = genesis.number.toString();
  genesis.timestamp = genesis.timestamp.toString();

  genesis.alloc = debug.traceTransaction(tx, {tracer: "prestateTracer", rewind: rewind});
  for (var key in genesis.alloc) {
    genesis.alloc[key].nonce = genesis.alloc[key].nonce.toString();
  }
  genesis.config = admin.nodeInfo.protocols.xcb.config;

  // Generate the call trace and produce the test input
  var result = debug.traceTransaction(tx, {tracer: "callTracer", rewind: rewind});
  delete result.time;

  console.log(JSON.stringify({
    genesis: genesis,
    context: {
      number:     block.number.toString(),
      difficulty: block.difficulty,
      timestamp:  block.timestamp.toString(),
      energyLimit:   block.energyLimit.toString(),
      miner:      block.miner,
    },
    input:  xcb.getRawTransaction(tx),
    result: result,
  }, null, 2));
}
*/

// callTrace is the result of a callTracer run.
type callTrace struct {
	Type       string          `json:"type"`
	From       common.Address  `json:"from"`
	To         common.Address  `json:"to"`
	Input      hexutil.Bytes   `json:"input"`
	Output     hexutil.Bytes   `json:"output"`
	Energy     *hexutil.Uint64 `json:"energy,omitempty"`
	EnergyUsed *hexutil.Uint64 `json:"energyUsed,omitempty"`
	Value      *hexutil.Big    `json:"value,omitempty"`
	Error      string          `json:"error,omitempty"`
	Calls      []callTrace     `json:"calls,omitempty"`
}

type callContext struct {
	Number      math.HexOrDecimal64   `json:"number"`
	Difficulty  *math.HexOrDecimal256 `json:"difficulty"`
	Time        math.HexOrDecimal64   `json:"timestamp"`
	EnergyLimit math.HexOrDecimal64   `json:"energyLimit"`
	Miner       common.Address        `json:"miner"`
}

// callTracerTest defines a single test to check the call tracer against.
type callTracerTest struct {
	Genesis *core.Genesis `json:"genesis"`
	Context *callContext  `json:"context"`
	Input   string        `json:"input"`
	Result  *callTrace    `json:"result"`
}

func TestPrestateTracerCreate2(t *testing.T) {
	addr, _ := common.HexToAddress("cb8000000000000000000000000000000000deadbeef")
	unsignedTx := types.NewTransaction(1, addr,
		new(big.Int), 5000000, big.NewInt(1), []byte{})

	privateKeyEDDSA, err := crypto.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	signer := types.NewNucleusSigner(big.NewInt(1))
	tx, err := types.SignTx(unsignedTx, signer, privateKeyEDDSA)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	/**
		This comes from one of the test-vectors on the Skinny Create2 - CIP

	    address 0x00000000000000000000000000000000deadbeef
	    salt 0x00000000000000000000000000000000000000000000000000000000cafebabe
	    init_code 0xdeadbeef
	    energy (assuming no mem expansion): 32006
	    result: 0x60f3f640a8508fC6a86d45DF051962668E1e8AC7
	*/
	origin, _ := signer.Sender(tx)
	txContext := vm.TxContext{
		Origin:      origin,
		EnergyPrice: big.NewInt(1),
	}
	context := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    common.Address{},
		BlockNumber: new(big.Int).SetUint64(8000000),
		Time:        new(big.Int).SetUint64(5),
		Difficulty:  big.NewInt(0x30000),
		EnergyLimit: uint64(6000000),
	}
	alloc := core.GenesisAlloc{}

	// The code pushes 'deadbeef' into memory, then the other params, and calls CREATE2, then returns
	// the address
	alloc[addr] = core.GenesisAccount{
		Nonce:   1,
		Code:    hexutil.MustDecode("0x63deadbeef60005263cafebabe6004601c6000F560005260206000F3"),
		Balance: big.NewInt(1),
	}
	alloc[origin] = core.GenesisAccount{
		Nonce:   1,
		Code:    []byte{},
		Balance: big.NewInt(500000000000000),
	}
	_, statedb := tests.MakePreState(rawdb.NewMemoryDatabase(), alloc, false)

	// Create the tracer, the CVM environment and run it
	tracer, err := New("prestateTracer")
	if err != nil {
		t.Fatalf("failed to create call tracer: %v", err)
	}
	cvm := vm.NewCVM(context, txContext, statedb, params.MainnetChainConfig, vm.Config{Debug: true, Tracer: tracer})

	msg, err := tx.AsMessage(signer)
	if err != nil {
		t.Fatalf("failed to prepare transaction for tracing: %v", err)
	}
	st := core.NewStateTransition(cvm, msg, new(core.EnergyPool).AddEnergy(tx.Energy()))
	if _, err = st.TransitionDb(); err != nil {
		t.Fatalf("failed to execute transaction: %v", err)
	}
	// Retrieve the trace result and compare against the etalon
	res, err := tracer.GetResult()
	if err != nil {
		t.Fatalf("failed to retrieve trace result: %v", err)
	}
	ret := make(map[string]interface{})
	if err := json.Unmarshal(res, &ret); err != nil {
		t.Fatalf("failed to unmarshal trace result: %v", err)
	}
	if _, has := ret["0xcb16050c2d2855fcee1a2ab8e4db79cc1015faf3b5d4"]; !has {
		t.Fatalf("Expected 0xcb16050c2d2855fcee1a2ab8e4db79cc1015faf3b5d4 in result")
	}
}

// Iterates over all the input-output datasets in the tracer test harness and
// runs the JavaScript tracers against them.
func TestCallTracer(t *testing.T) {
	t.Skip("temporary skip")
	files, err := ioutil.ReadDir("testdata")
	if err != nil {
		t.Fatalf("failed to retrieve tracer test suite: %v", err)
	}
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), "call_tracer_") {
			continue
		}
		file := file // capture range variable
		t.Run(camel(strings.TrimSuffix(strings.TrimPrefix(file.Name(), "call_tracer_"), ".json")), func(t *testing.T) {
			t.Parallel()

			// Call tracer test found, read if from disk
			blob, err := ioutil.ReadFile(filepath.Join("testdata", file.Name()))
			if err != nil {
				t.Fatalf("failed to read testcase: %v", err)
			}
			test := new(callTracerTest)
			if err := json.Unmarshal(blob, test); err != nil {
				t.Fatalf("failed to parse testcase: %v", err)
			}
			// Configure a blockchain with the given prestate
			tx := new(types.Transaction)
			if err := rlp.DecodeBytes(common.FromHex(test.Input), tx); err != nil {
				t.Fatalf("failed to parse testcase input: %v", err)
			}
			signer := types.MakeSigner(test.Genesis.Config.NetworkID)
			origin, _ := signer.Sender(tx)
			txContext := vm.TxContext{
				Origin:      origin,
				EnergyPrice: tx.EnergyPrice(),
			}
			context := vm.BlockContext{
				CanTransfer: core.CanTransfer,
				Transfer:    core.Transfer,
				Coinbase:    test.Context.Miner,
				BlockNumber: new(big.Int).SetUint64(uint64(test.Context.Number)),
				Time:        new(big.Int).SetUint64(uint64(test.Context.Time)),
				Difficulty:  (*big.Int)(test.Context.Difficulty),
				EnergyLimit: uint64(test.Context.EnergyLimit),
			}
			_, statedb := tests.MakePreState(rawdb.NewMemoryDatabase(), test.Genesis.Alloc, false)

			// Create the tracer, the CVM environment and run it
			tracer, err := New("callTracer")
			if err != nil {
				t.Fatalf("failed to create call tracer: %v", err)
			}
			cvm := vm.NewCVM(context, txContext, statedb, test.Genesis.Config, vm.Config{Debug: true, Tracer: tracer})

			msg, err := tx.AsMessage(signer)
			if err != nil {
				t.Fatalf("failed to prepare transaction for tracing: %v", err)
			}
			st := core.NewStateTransition(cvm, msg, new(core.EnergyPool).AddEnergy(tx.Energy()))
			if _, err = st.TransitionDb(); err != nil {
				t.Fatalf("failed to execute transaction: %v", err)
			}
			// Retrieve the trace result and compare against the etalon
			res, err := tracer.GetResult()
			if err != nil {
				t.Fatalf("failed to retrieve trace result: %v", err)
			}
			ret := new(callTrace)
			if err := json.Unmarshal(res, ret); err != nil {
				t.Fatalf("failed to unmarshal trace result: %v", err)
			}

			if !jsonEqual(ret, test.Result) {
				// uncomment this for easier debugging
				//have, _ := json.MarshalIndent(ret, "", " ")
				//want, _ := json.MarshalIndent(test.Result, "", " ")
				//t.Fatalf("trace mismatch: \nhave %+v\nwant %+v", string(have), string(want))
				t.Fatalf("trace mismatch: \nhave %+v\nwant %+v", ret, test.Result)
			}
		})
	}
}

// jsonEqual is similar to reflect.DeepEqual, but does a 'bounce' via json prior to
// comparison
func jsonEqual(x, y interface{}) bool {
	xTrace := new(callTrace)
	yTrace := new(callTrace)
	if xj, err := json.Marshal(x); err == nil {
		json.Unmarshal(xj, xTrace)
	} else {
		return false
	}
	if yj, err := json.Marshal(y); err == nil {
		json.Unmarshal(yj, yTrace)
	} else {
		return false
	}
	return reflect.DeepEqual(xTrace, yTrace)
}
