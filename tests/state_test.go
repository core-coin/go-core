// Copyright 2020 The CORE FOUNDATION, nadacia
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

package tests

import (
	"bufio"
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/core-coin/go-core/core/vm"
)

func TestState(t *testing.T) {
	t.Parallel()

	st := new(testMatcher)
	// Long tests:
	st.slow(`^stAttackTest/ContractCreationSpam`)
	st.slow(`^stBadOpcode/badOpcodes`)
	st.slow(`^stPreCompiledContracts/modexp`)
	st.slow(`^stQuadraticComplexityTest/`)
	st.slow(`^stStaticCall/static_Call50000`)
	st.slow(`^stStaticCall/static_Return50000`)
	st.slow(`^stStaticCall/static_Call1MB`)
	st.slow(`^stSystemOperationsTest/CallRecursiveBomb`)
	st.slow(`^stTransactionTest/Opcodes_TransactionInit`)

	// Very time consuming
	st.skipLoad(`^stTimeConsuming/`)

	// Older tests were moved into LegacyTests
	for _, dir := range []string{
		stateTestDir,
		legacyStateTestDir,
	} {
		st.walk(t, dir, func(t *testing.T, name string, test *StateTest) {
			for _, subtest := range test.Subtests() {
				subtest := subtest
				key := fmt.Sprintf("%s/%d", subtest.Fork, subtest.Index)
				name := name + "/" + key

				t.Run(key+"/trie", func(t *testing.T) {
					withTrace(t, test.energyLimit(subtest), func(vmconfig vm.Config) error {
						_, err := test.Run(subtest, vmconfig, false)
						return st.checkFailure(t, name+"/trie", err)
					})
				})
				t.Run(key+"/snap", func(t *testing.T) {
					withTrace(t, test.energyLimit(subtest), func(vmconfig vm.Config) error {
						_, err := test.Run(subtest, vmconfig, true)
						return st.checkFailure(t, name+"/snap", err)
					})
				})
			}
		})
	}
}

// Transactions with energyLimit above this value will not get a VM trace on failure.
const traceErrorLimit = 400000

func withTrace(t *testing.T, energyLimit uint64, test func(vm.Config) error) {
	// Use config from command line arguments.
	config := vm.Config{CVMInterpreter: *testCVM, EWASMInterpreter: *testEWASM}
	err := test(config)
	if err == nil {
		return
	}

	// Test failed, re-run with tracing enabled.
	t.Error(err)
	if energyLimit > traceErrorLimit {
		t.Log("energy limit too high for CVM trace")
		return
	}
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)
	tracer := vm.NewJSONLogger(&vm.LogConfig{DisableMemory: true}, w)
	config.Debug, config.Tracer = true, tracer
	err2 := test(config)
	if !reflect.DeepEqual(err, err2) {
		t.Errorf("different error for second run: %v", err2)
	}
	w.Flush()
	if buf.Len() == 0 {
		t.Log("no CVM operation logs generated")
	} else {
		t.Log("CVM operation log:\n" + buf.String())
	}
	//t.Logf("CVM output: 0x%x", tracer.Output())
	//t.Logf("CVM error: %v", tracer.Error())
}
