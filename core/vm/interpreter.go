// Copyright 2014 by the Authors
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

package vm

import (
	"hash"
	"sync/atomic"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/math"
)

// Config are the configuration options for the Interpreter
type Config struct {
	Debug                   bool   // Enables debugging
	Tracer                  Tracer // Opcode logger
	NoRecursion             bool   // Disables call, callcode, delegate call and create
	EnablePreimageRecording bool   // Enables recording of SHA3/keccak preimages

	JumpTable [256]*operation // CVM instruction table, automatically populated if unset

	EWASMInterpreter string // External EWASM interpreter options
	CVMInterpreter   string // External CVM interpreter options

	ExtraCips []int // Additional CIPS that are to be enabled
}

// Interpreter is used to run Core based contracts and will utilise the
// passed environment to query external sources for state information.
// The Interpreter will run the byte code VM based on the passed
// configuration.
type Interpreter interface {
	// Run loops and evaluates the contract's code with the given input data and returns
	// the return byte-slice and an error if one occurred.
	Run(contract *Contract, input []byte, static bool) ([]byte, error)
	// CanRun tells if the contract, passed as an argument, can be
	// run by the current interpreter. This is meant so that the
	// caller can do something like:
	//
	// ```golang
	// for _, interpreter := range interpreters {
	//   if interpreter.CanRun(contract.code) {
	//     interpreter.Run(contract.code, input)
	//   }
	// }
	// ```
	CanRun([]byte) bool
}

// callCtx contains the things that are per-call, such as stack and memory,
// but not transients like pc and energy
type callCtx struct {
	memory   *Memory
	stack    *Stack
	rstack   *ReturnStack
	contract *Contract
}

// keccakState wraps sha3.state. In addition to the usual hash methods, it also supports
// Read to get a variable amount of data from the hash state. Read is faster than Sum
// because it doesn't copy the internal state, but also modifies the internal state.
type keccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

// CVMInterpreter represents an CVM interpreter
type CVMInterpreter struct {
	cvm *CVM
	cfg Config

	hasher    keccakState // SHA3 hasher instance shared across opcodes
	hasherBuf common.Hash // SHA3 hasher result array shared aross opcodes

	readOnly   bool   // Whether to throw on stateful modifications
	returnData []byte // Last CALL's return data for subsequent reuse
}

// NewCVMInterpreter returns a new instance of the Interpreter.
func NewCVMInterpreter(cvm *CVM, cfg Config) *CVMInterpreter {
	// We use the STOP instruction whether to see
	// the jump table was initialised. If it was not
	// we'll set the default jump table.
	if cfg.JumpTable[STOP] == nil {
		var jt JumpTable
		switch {
		default:
			jt = InstructionSet
		}
		cfg.JumpTable = jt
	}

	return &CVMInterpreter{
		cvm: cvm,
		cfg: cfg,
	}
}

// Run loops and evaluates the contract's code with the given input data and returns
// the return byte-slice and an error if one occurred.
//
// It's important to note that any errors returned by the interpreter should be
// considered a revert-and-consume-all-energy operation except for
// ErrExecutionReverted which means revert-and-keep-energy-left.
func (in *CVMInterpreter) Run(contract *Contract, input []byte, readOnly bool) (ret []byte, err error) {
	// Increment the call depth which is restricted to 1024
	in.cvm.depth++
	defer func() { in.cvm.depth-- }()

	// Make sure the readOnly is only set if we aren't in readOnly yet.
	// This makes also sure that the readOnly flag isn't removed for child calls.
	if readOnly && !in.readOnly {
		in.readOnly = true
		defer func() { in.readOnly = false }()
	}

	// Reset the previous call's return data. It's unimportant to preserve the old buffer
	// as every returning call will return new data anyway.
	in.returnData = nil

	// Don't bother with the execution if there's no code.
	if len(contract.Code) == 0 {
		return nil, nil
	}

	var (
		op          OpCode             // current opcode
		mem         = NewMemory()      // bound memory
		stack       = newstack()       // local stack
		returns     = newReturnStack() // local returns stack
		callContext = &callCtx{
			memory:   mem,
			stack:    stack,
			rstack:   returns,
			contract: contract,
		}
		// For optimisation reason we're using uint64 as the program counter.
		// It's theoretically possible to go above 2^64. The YP defines the PC
		// to be uint256. Practically much less so feasible.
		pc   = uint64(0) // program counter
		cost uint64
		// copies used by tracer
		pcCopy     uint64 // needed for the deferred Tracer
		energyCopy uint64 // for Tracer to log energy remaining before execution
		logged     bool   // deferred Tracer should ignore already logged steps
		res        []byte // result of the opcode execution function
	)
	// Don't move this deferrred function, it's placed before the capturestate-deferred method,
	// so that it get's executed _after_: the capturestate needs the stacks before
	// they are returned to the pools
	defer func() {
		returnStack(stack)
		returnRStack(returns)
	}()
	contract.Input = input

	if in.cfg.Debug {
		defer func() {
			if err != nil {
				if !logged {
					in.cfg.Tracer.CaptureState(in.cvm, pcCopy, op, energyCopy, cost, mem, stack, returns, in.returnData, contract, in.cvm.depth, err)
				} else {
					in.cfg.Tracer.CaptureFault(in.cvm, pcCopy, op, energyCopy, cost, mem, stack, returns, contract, in.cvm.depth, err)
				}
			}
		}()
	}
	// The Interpreter main run loop (contextual). This loop runs until either an
	// explicit STOP, RETURN or SELFDESTRUCT is executed, an error occurred during
	// the execution of one of the operations or until the done flag is set by the
	// parent context.
	steps := 0
	for {
		steps++
		if steps%1000 == 0 && atomic.LoadInt32(&in.cvm.abort) != 0 {
			break
		}
		if in.cfg.Debug {
			// Capture pre-execution values for tracing.
			logged, pcCopy, energyCopy = false, pc, contract.Energy
		}

		// Get the operation from the jump table and validate the stack to ensure there are
		// enough stack items available to perform the operation.
		op = contract.GetOp(pc)
		operation := in.cfg.JumpTable[op]
		if operation == nil {
			return nil, &ErrInvalidOpCode{opcode: op}
		}
		// Validate stack
		if sLen := stack.len(); sLen < operation.minStack {
			return nil, &ErrStackUnderflow{stackLen: sLen, required: operation.minStack}
		} else if sLen > operation.maxStack {
			return nil, &ErrStackOverflow{stackLen: sLen, limit: operation.maxStack}
		}
		// If the operation is valid, enforce and write restrictions
		if in.readOnly {
			// If the interpreter is operating in readonly mode, make sure no
			// state-modifying operation is performed. The 3rd stack item
			// for a call operation is the value. Transferring value from one
			// account to the others means the state is modified and should also
			// return with an error.
			if operation.writes || (op == CALL && stack.Back(2).Sign() != 0) {
				return nil, ErrWriteProtection
			}
		}
		// Static portion of energy
		cost = operation.constantEnergy // For tracing
		if !contract.UseEnergy(operation.constantEnergy) {
			return nil, ErrOutOfEnergy
		}

		var memorySize uint64
		// calculate the new memory size and expand the memory to fit
		// the operation
		// Memory check needs to be done prior to evaluating the dynamic energy portion,
		// to detect calculation overflows
		if operation.memorySize != nil {
			memSize, overflow := operation.memorySize(stack)
			if overflow {
				return nil, ErrEnergyUintOverflow
			}
			// memory is expanded in words of 32 bytes. Energy
			// is also calculated in words.
			if memorySize, overflow = math.SafeMul(toWordSize(memSize), 32); overflow {
				return nil, ErrEnergyUintOverflow
			}
		}
		// Dynamic portion of energy
		// consume the energy and return an error if not enough energy is available.
		// cost is explicitly set so that the capture state defer method can get the proper cost
		if operation.dynamicEnergy != nil {
			var dynamicCost uint64
			dynamicCost, err = operation.dynamicEnergy(in.cvm, contract, stack, mem, memorySize)
			cost += dynamicCost // total cost, for debug tracing
			if err != nil || !contract.UseEnergy(dynamicCost) {
				return nil, ErrOutOfEnergy
			}
		}
		if memorySize > 0 {
			mem.Resize(memorySize)
		}

		if in.cfg.Debug {
			in.cfg.Tracer.CaptureState(in.cvm, pc, op, energyCopy, cost, mem, stack, returns, in.returnData, contract, in.cvm.depth, err)
			logged = true
		}

		// execute the operation
		res, err = operation.execute(&pc, in, callContext)
		// if the operation clears the return data (e.g. it has returning data)
		// set the last return to the result of the operation.
		if operation.returns {
			in.returnData = common.CopyBytes(res)
		}

		switch {
		case err != nil:
			return nil, err
		case operation.reverts:
			return res, ErrExecutionReverted
		case operation.halts:
			return res, nil
		case !operation.jumps:
			pc++
		}
	}
	return nil, nil
}

// CanRun tells if the contract, passed as an argument, can be
// run by the current interpreter.
func (in *CVMInterpreter) CanRun(code []byte) bool {
	return true
}
