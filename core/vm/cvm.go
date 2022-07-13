// Copyright 2020 by the Authors
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
	"errors"
	"github.com/core-coin/uint256"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/params"
)

// emptyCodeHash is used by create to ensure deployment is disallowed to already
// deployed contract addresses (relevant after the account abstraction).
var emptyCodeHash = crypto.SHA3Hash(nil)

type (
	// CanTransferFunc is the signature of a transfer guard function
	CanTransferFunc func(StateDB, common.Address, *big.Int) bool
	// TransferFunc is the signature of a transfer function
	TransferFunc func(StateDB, common.Address, common.Address, *big.Int)
	// GetHashFunc returns the n'th block hash in the blockchain
	// and is used by the BLOCKHASH CVM op code.
	GetHashFunc func(uint64) common.Hash
)

func (cvm *CVM) precompile(addr common.Address) (PrecompiledContract, bool) {
	var precompiles map[common.Address]PrecompiledContract
	precompiles = PrecompiledContracts
	p, ok := precompiles[addr]
	return p, ok
}

// run runs the given contract and takes care of running precompiles with a fallback to the byte code interpreter.
func run(cvm *CVM, contract *Contract, input []byte, readOnly bool) ([]byte, error) {
	for _, interpreter := range cvm.interpreters {
		if interpreter.CanRun(contract.Code) {
			if cvm.interpreter != interpreter {
				// Ensure that the interpreter pointer is set back
				// to its current value upon return.
				defer func(i Interpreter) {
					cvm.interpreter = i
				}(cvm.interpreter)
				cvm.interpreter = interpreter
			}
			return interpreter.Run(contract, input, readOnly)
		}
	}
	return nil, errors.New("no compatible interpreter")
}

// Context provides the CVM with auxiliary information. Once provided
// it shouldn't be modified.
type Context struct {
	// CanTransfer returns whether the account contains
	// sufficient core to transfer the value
	CanTransfer CanTransferFunc
	// Transfer transfers core from one account to the other
	Transfer TransferFunc
	// GetHash returns the hash corresponding to n
	GetHash GetHashFunc

	// Message information
	Origin      common.Address // Provides information for ORIGIN
	EnergyPrice *big.Int       // Provides information for ENERGYPRICE

	// Block information
	Coinbase    common.Address // Provides information for COINBASE
	EnergyLimit uint64         // Provides information for ENERGYLIMIT
	BlockNumber *big.Int       // Provides information for NUMBER
	Time        *big.Int       // Provides information for TIME
	Difficulty  *big.Int       // Provides information for DIFFICULTY
}

// CVM is the Core Virtual Machine base object and provides
// the necessary tools to run a contract on the given state with
// the provided context. It should be noted that any error
// generated through any of the calls should be considered a
// revert-state-and-consume-all-energy operation, no checks on
// specific errors should ever be performed. The interpreter makes
// sure that any errors generated are to be considered faulty code.
//
// The CVM should never be reused and is not thread safe.
type CVM struct {
	// Context provides auxiliary blockchain related information
	Context
	// StateDB gives access to the underlying state
	StateDB StateDB
	// Depth is the current call stack
	depth int

	// chainConfig contains information about the current chain
	chainConfig *params.ChainConfig
	// chain rules contains the chain rules for the current epoch
	chainRules params.Rules
	// virtual machine configuration options used to initialise the
	// cvm.
	vmConfig Config
	// global (to this context) Core Virtual Machine
	// used throughout the execution of the tx.
	interpreters []Interpreter
	interpreter  Interpreter
	// abort is used to abort the CVM calling operations
	// NOTE: must be set atomically
	abort int32
	// callEnergyTemp holds the energy available for the current call. This is needed because the
	// available energy is calculated in energyCall* according to the 63/64 rule and later
	// applied in opCall*.
	callEnergyTemp uint64
}

// NewCVM returns a new CVM. The returned CVM is not thread safe and should
// only ever be used *once*.
func NewCVM(ctx Context, statedb StateDB, chainConfig *params.ChainConfig, vmConfig Config) *CVM {
	cvm := &CVM{
		Context:      ctx,
		StateDB:      statedb,
		vmConfig:     vmConfig,
		chainConfig:  chainConfig,
		chainRules:   chainConfig.Rules(ctx.BlockNumber),
		interpreters: make([]Interpreter, 0, 1),
	}

	if chainConfig.IsEWASM(ctx.BlockNumber) {
		// to be implemented by CVM-C and Wagon PRs.
		// if vmConfig.EWASMInterpreter != "" {
		//  extIntOpts := strings.Split(vmConfig.EWASMInterpreter, ":")
		//  path := extIntOpts[0]
		//  options := []string{}
		//  if len(extIntOpts) > 1 {
		//    options = extIntOpts[1..]
		//  }
		//  cvm.interpreters = append(cvm.interpreters, NewCVMVCInterpreter(cvm, vmConfig, options))
		// } else {
		// 	cvm.interpreters = append(cvm.interpreters, NewEWASMInterpreter(cvm, vmConfig))
		// }
		panic("No supported ewasm interpreter yet.")
	}

	// vmConfig.CVMInterpreter will be used by CVM-C, it won't be checked here
	// as we always want to have the built-in CVM as the failover option.
	cvm.interpreters = append(cvm.interpreters, NewCVMInterpreter(cvm, vmConfig))
	cvm.interpreter = cvm.interpreters[0]

	return cvm
}

// Cancel cancels any running CVM operation. This may be called concurrently and
// it's safe to be called multiple times.
func (cvm *CVM) Cancel() {
	atomic.StoreInt32(&cvm.abort, 1)
}

// Cancelled returns true if Cancel has been called
func (cvm *CVM) Cancelled() bool {
	return atomic.LoadInt32(&cvm.abort) == 1
}

// Interpreter returns the current interpreter
func (cvm *CVM) Interpreter() Interpreter {
	return cvm.interpreter
}

// Call executes the contract associated with the addr with the given input as
// parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
func (cvm *CVM) Call(caller ContractRef, addr common.Address, input []byte, energy uint64, value *big.Int) (ret []byte, leftOverEnergy uint64, err error) {
	if cvm.vmConfig.NoRecursion && cvm.depth > 0 {
		return nil, energy, nil
	}

	// Fail if we're trying to execute above the call depth limit
	if cvm.depth > int(params.CallCreateDepth) {
		return nil, energy, ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	if value.Sign() != 0 && !cvm.Context.CanTransfer(cvm.StateDB, caller.Address(), value) {
		return nil, energy, ErrInsufficientBalance
	}

	snapshot := cvm.StateDB.Snapshot()
	p, isPrecompile := cvm.precompile(addr)

	if !cvm.StateDB.Exist(addr) {
		if !isPrecompile && value.Sign() == 0 {
			// Calling a non existing account, don't do anything, but ping the tracer
			if cvm.vmConfig.Debug && cvm.depth == 0 {
				cvm.vmConfig.Tracer.CaptureStart(caller.Address(), addr, false, input, energy, value)
				cvm.vmConfig.Tracer.CaptureEnd(ret, 0, 0, nil)
			}
			return nil, energy, nil
		}
		cvm.StateDB.CreateAccount(addr)
	}
	cvm.Transfer(cvm.StateDB, caller.Address(), addr, value)

	// Capture the tracer start/end events in debug mode
	if cvm.vmConfig.Debug && cvm.depth == 0 {
		cvm.vmConfig.Tracer.CaptureStart(caller.Address(), addr, false, input, energy, value)

		defer func(startEnergy uint64, startTime time.Time) { // Lazy evaluation of the parameters
			cvm.vmConfig.Tracer.CaptureEnd(ret, startEnergy-energy, time.Since(startTime), err)
		}(energy, time.Now())
	}
	if isPrecompile {
		ret, energy, err = RunPrecompiledContract(p, input, energy)
	} else {
		// Initialise a new contract and set the code that is to be used by the CVM.
		// The contract is a scoped environment for this execution context only.
		code := cvm.StateDB.GetCode(addr)
		if len(code) == 0 {
			ret, err = nil, nil // energy is unchanged
		} else {
			addrCopy := addr
			// If the account has no code, we can abort here
			// The depth-check is already done, and precompiles handled above
			contract := NewContract(caller, AccountRef(addrCopy), value, energy)
			contract.SetCallCode(&addrCopy, cvm.StateDB.GetCodeHash(addrCopy), code)
			ret, err = run(cvm, contract, input, false)
			energy = contract.Energy
		}
	}

	// When an error was returned by the CVM or when setting the creation code
	// above we revert to the snapshot and consume any energy remaining.
	if err != nil {
		cvm.StateDB.RevertToSnapshot(snapshot)
		if err != ErrExecutionReverted {
			energy = 0
		}
		// TODO: consider clearing up unused snapshots:
		//} else {
		//	cvm.StateDB.DiscardSnapshot(snapshot)
	}
	return ret, energy, err
}

// CallCode executes the contract associated with the addr with the given input
// as parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
//
// CallCode differs from Call in the sense that it executes the given address'
// code with the caller as context.
func (cvm *CVM) CallCode(caller ContractRef, addr common.Address, input []byte, energy uint64, value *big.Int) (ret []byte, leftOverEnergy uint64, err error) {
	if cvm.vmConfig.NoRecursion && cvm.depth > 0 {
		return nil, energy, nil
	}

	// Fail if we're trying to execute above the call depth limit
	if cvm.depth > int(params.CallCreateDepth) {
		return nil, energy, ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	// Note although it's noop to transfer X core to caller itself. But
	// if caller doesn't have enough balance, it would be an error to allow
	// over-charging itself. So the check here is necessary.
	if !cvm.Context.CanTransfer(cvm.StateDB, caller.Address(), value) {
		return nil, energy, ErrInsufficientBalance
	}

	var snapshot = cvm.StateDB.Snapshot()

	// It is allowed to call precompiles, even via delegatecall
	if p, isPrecompile := cvm.precompile(addr); isPrecompile {
		ret, energy, err = RunPrecompiledContract(p, input, energy)
	} else {
		addrCopy := addr
		// Initialise a new contract and set the code that is to be used by the CVM.
		// The contract is a scoped environment for this execution context only.
		contract := NewContract(caller, AccountRef(caller.Address()), value, energy)
		contract.SetCallCode(&addrCopy, cvm.StateDB.GetCodeHash(addrCopy), cvm.StateDB.GetCode(addrCopy))
		ret, err = run(cvm, contract, input, false)
		energy = contract.Energy
	}
	if err != nil {
		cvm.StateDB.RevertToSnapshot(snapshot)
		if err != ErrExecutionReverted {
			energy = 0
		}
	}
	return ret, energy, err
}

// DelegateCall executes the contract associated with the addr with the given input
// as parameters. It reverses the state in case of an execution error.
//
// DelegateCall differs from CallCode in the sense that it executes the given address'
// code with the caller as context and the caller is set to the caller of the caller.
func (cvm *CVM) DelegateCall(caller ContractRef, addr common.Address, input []byte, energy uint64) (ret []byte, leftOverEnergy uint64, err error) {
	if cvm.vmConfig.NoRecursion && cvm.depth > 0 {
		return nil, energy, nil
	}
	// Fail if we're trying to execute above the call depth limit
	if cvm.depth > int(params.CallCreateDepth) {
		return nil, energy, ErrDepth
	}

	var snapshot = cvm.StateDB.Snapshot()

	// It is allowed to call precompiles, even via delegatecall
	if p, isPrecompile := cvm.precompile(addr); isPrecompile {
		ret, energy, err = RunPrecompiledContract(p, input, energy)
	} else {
		addrCopy := addr
		// Initialise a new contract and make initialise the delegate values
		contract := NewContract(caller, AccountRef(caller.Address()), nil, energy).AsDelegate()
		contract.SetCallCode(&addrCopy, cvm.StateDB.GetCodeHash(addrCopy), cvm.StateDB.GetCode(addrCopy))
		ret, err = run(cvm, contract, input, false)
		energy = contract.Energy
	}
	if err != nil {
		cvm.StateDB.RevertToSnapshot(snapshot)
		if err != ErrExecutionReverted {
			energy = 0
		}
	}
	return ret, energy, err
}

// StaticCall executes the contract associated with the addr with the given input
// as parameters while disallowing any modifications to the state during the call.
// Opcodes that attempt to perform such modifications will result in exceptions
// instead of performing the modifications.
func (cvm *CVM) StaticCall(caller ContractRef, addr common.Address, input []byte, energy uint64) (ret []byte, leftOverEnergy uint64, err error) {
	if cvm.vmConfig.NoRecursion && cvm.depth > 0 {
		return nil, energy, nil
	}
	// Fail if we're trying to execute above the call depth limit
	if cvm.depth > int(params.CallCreateDepth) {
		return nil, energy, ErrDepth
	}
	// We take a snapshot here. This is a bit counter-intuitive, and could probably be skipped.
	// However, even a staticcall is considered a 'touch'. On mainnet, static calls were introduced
	// after all empty accounts were deleted, so this is not required. However, if we omit this,
	// then certain tests start failing; stRevertTest/RevertPrecompiledTouchExactOOG.json.
	// We could change this, but for now it's left for legacy reasons
	var snapshot = cvm.StateDB.Snapshot()

	// We do an AddBalance of zero here, just in order to trigger a touch.
	// but is the correct thing to do and matters on other networks, in tests, and potential
	// future scenarios
	cvm.StateDB.AddBalance(addr, big0)

	if p, isPrecompile := cvm.precompile(addr); isPrecompile {
		ret, energy, err = RunPrecompiledContract(p, input, energy)
	} else {
		// At this point, we use a copy of address. If we don't, the go compiler will
		// leak the 'contract' to the outer scope, and make allocation for 'contract'
		// even if the actual execution ends on RunPrecompiled above.
		addrCopy := addr
		// Initialise a new contract and set the code that is to be used by the CVM.
		// The contract is a scoped environment for this execution context only.
		contract := NewContract(caller, AccountRef(addrCopy), new(big.Int), energy)
		contract.SetCallCode(&addrCopy, cvm.StateDB.GetCodeHash(addrCopy), cvm.StateDB.GetCode(addrCopy))
		// When an error was returned by the CVM or when setting the creation code
		// above we revert to the snapshot and consume any energy remaining.
		ret, err = run(cvm, contract, input, true)
		energy = contract.Energy
	}
	if err != nil {
		cvm.StateDB.RevertToSnapshot(snapshot)
		if err != ErrExecutionReverted {
			energy = 0
		}
	}
	return ret, energy, err
}

type codeAndHash struct {
	code []byte
	hash common.Hash
}

func (c *codeAndHash) Hash() common.Hash {
	if c.hash == (common.Hash{}) {
		c.hash = crypto.SHA3Hash(c.code)
	}
	return c.hash
}

// create creates a new contract using code as deployment code.
func (cvm *CVM) create(caller ContractRef, codeAndHash *codeAndHash, energy uint64, value *big.Int, address common.Address) ([]byte, common.Address, uint64, error) {
	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if cvm.depth > int(params.CallCreateDepth) {
		return nil, common.Address{}, energy, ErrDepth
	}
	if !cvm.CanTransfer(cvm.StateDB, caller.Address(), value) {
		return nil, common.Address{}, energy, ErrInsufficientBalance
	}
	nonce := cvm.StateDB.GetNonce(caller.Address())
	cvm.StateDB.SetNonce(caller.Address(), nonce+1)

	// Ensure there's no existing contract already at the designated address
	contractHash := cvm.StateDB.GetCodeHash(address)
	if cvm.StateDB.GetNonce(address) != 0 || (contractHash != (common.Hash{}) && contractHash != emptyCodeHash) {
		return nil, common.Address{}, 0, ErrContractAddressCollision
	}
	// Create a new account on the state
	snapshot := cvm.StateDB.Snapshot()
	cvm.StateDB.CreateAccount(address)
	cvm.StateDB.SetNonce(address, 1)
	cvm.Transfer(cvm.StateDB, caller.Address(), address, value)

	// Initialise a new contract and set the code that is to be used by the CVM.
	// The contract is a scoped environment for this execution context only.
	contract := NewContract(caller, AccountRef(address), value, energy)
	contract.SetCodeOptionalHash(&address, codeAndHash)

	if cvm.vmConfig.NoRecursion && cvm.depth > 0 {
		return nil, address, energy, nil
	}

	if cvm.vmConfig.Debug && cvm.depth == 0 {
		cvm.vmConfig.Tracer.CaptureStart(caller.Address(), address, true, codeAndHash.code, energy, value)
	}
	start := time.Now()

	ret, err := run(cvm, contract, nil, false)

	// check whether the max code size has been exceeded
	maxCodeSizeExceeded := len(ret) > params.MaxCodeSize
	// if the contract creation ran successfully and no errors were returned
	// calculate the energy required to store the code. If the code could not
	// be stored due to not enough energy set an error and let it be handled
	// by the error checking condition below.
	if err == nil && !maxCodeSizeExceeded {
		createDataEnergy := uint64(len(ret)) * params.CreateDataEnergy
		if contract.UseEnergy(createDataEnergy) {
			cvm.StateDB.SetCode(address, ret)
		} else {
			err = ErrCodeStoreOutOfEnergy
		}
	}

	// When an error was returned by the CVM or when setting the creation code
	// above we revert to the snapshot and consume any energy remaining.
	if maxCodeSizeExceeded || (err != nil && err != ErrCodeStoreOutOfEnergy) {
		cvm.StateDB.RevertToSnapshot(snapshot)
		if err != ErrExecutionReverted {
			contract.UseEnergy(contract.Energy)
		}
	}
	// Assign err if contract code size exceeds the max while the err is still empty.
	if maxCodeSizeExceeded && err == nil {
		err = ErrMaxCodeSizeExceeded
	}
	if cvm.vmConfig.Debug && cvm.depth == 0 {
		cvm.vmConfig.Tracer.CaptureEnd(ret, energy-contract.Energy, time.Since(start), err)
	}
	return ret, address, contract.Energy, err

}

// Create creates a new contract using code as deployment code.
func (cvm *CVM) Create(caller ContractRef, code []byte, energy uint64, value *big.Int) (ret []byte, contractAddr common.Address, leftOverEnergy uint64, err error) {
	contractAddr = crypto.CreateAddress(caller.Address(), cvm.StateDB.GetNonce(caller.Address()))
	return cvm.create(caller, &codeAndHash{code: code}, energy, value, contractAddr)
}

// Create2 creates a new contract using code as deployment code.
//
// The different between Create2 with Create is Create2 uses sha3(0xff ++ msg.sender ++ salt ++ sha3(init_code))[12:]
// instead of the usual sender-and-nonce-hash as the address where the contract is initialized at.
func (cvm *CVM) Create2(caller ContractRef, code []byte, energy uint64, endowment *big.Int, salt *uint256.Int) (ret []byte, contractAddr common.Address, leftOverEnergy uint64, err error) {
	codeAndHash := &codeAndHash{code: code}
	contractAddr = crypto.CreateAddress2(caller.Address(), common.Hash(salt.Bytes32()), codeAndHash.Hash().Bytes())
	return cvm.create(caller, codeAndHash, energy, endowment, contractAddr)
}

// ChainConfig returns the environment's chain configuration
func (cvm *CVM) ChainConfig() *params.ChainConfig { return cvm.chainConfig }
