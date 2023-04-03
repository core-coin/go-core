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

package core

import (
	"fmt"
	"math"
	"math/big"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/core/vm"
	"github.com/core-coin/go-core/v2/params"
)

/*
The State Transitioning Model

A state transition is a change made when a transaction is applied to the current world state
The state transitioning model does all the necessary work to work out a valid new state root.

1) Nonce handling
2) Pre pay energy
3) Create a new state object if the recipient is \0*32
4) Value transfer
== If contract creation ==

	4a) Attempt to run transaction data
	4b) If valid, use result as code for the new state object

== end ==
5) Run Script section
6) Derive new state root
*/
type StateTransition struct {
	gp            *EnergyPool
	msg           Message
	energy        uint64
	energyPrice   *big.Int
	initialEnergy uint64
	value         *big.Int
	data          []byte
	state         vm.StateDB
	cvm           *vm.CVM
}

// Message represents a message sent to a contract.
type Message interface {
	From() common.Address
	To() *common.Address

	EnergyPrice() *big.Int
	Energy() uint64
	Value() *big.Int

	Nonce() uint64
	CheckNonce() bool
	Data() []byte
}

// ExecutionResult includes all output after executing given cvm
// message no matter the execution itself is successful or not.
type ExecutionResult struct {
	UsedEnergy uint64 // Total used energy but include the refunded energy
	Err        error  // Any error encountered during the execution(listed in core/vm/errors.go)
	ReturnData []byte // Returned data from cvm(function result or data supplied with revert opcode)
}

// Unwrap returns the internal cvm error which allows us for further
// analysis outside.
func (result *ExecutionResult) Unwrap() error {
	return result.Err
}

// Failed returns the indicator whether the execution is successful or not
func (result *ExecutionResult) Failed() bool { return result.Err != nil }

// Return is a helper function to help caller distinguish between revert reason
// and function return. Return returns the data after execution if no error occurs.
func (result *ExecutionResult) Return() []byte {
	if result.Err != nil {
		return nil
	}
	return common.CopyBytes(result.ReturnData)
}

// Revert returns the concrete revert reason if the execution is aborted by `REVERT`
// opcode. Note the reason can be nil if no data supplied with revert opcode.
func (result *ExecutionResult) Revert() []byte {
	if result.Err != vm.ErrExecutionReverted {
		return nil
	}
	return common.CopyBytes(result.ReturnData)
}

// IntrinsicEnergy computes the 'intrinsic energy' for a message with the given data.
func IntrinsicEnergy(data []byte, contractCreation bool) (uint64, error) {
	// Set the starting energy for the raw transaction
	var energy uint64
	if contractCreation {
		energy = params.TxEnergyContractCreation
	} else {
		energy = params.TxEnergy
	}
	// Bump the required energy by the amount of transactional data
	if len(data) > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		// Make sure we don't exceed uint64 for all data combinations
		nonZeroEnergy := params.TxDataNonZeroEnergy
		if (math.MaxUint64-energy)/nonZeroEnergy < nz {
			return 0, ErrEnergyUintOverflow
		}
		energy += nz * nonZeroEnergy

		z := uint64(len(data)) - nz
		if (math.MaxUint64-energy)/params.TxDataZeroEnergy < z {
			return 0, ErrEnergyUintOverflow
		}
		energy += z * params.TxDataZeroEnergy
	}
	return energy, nil
}

// NewStateTransition initialises and returns a new state transition object.
func NewStateTransition(cvm *vm.CVM, msg Message, gp *EnergyPool) *StateTransition {
	return &StateTransition{
		gp:          gp,
		cvm:         cvm,
		msg:         msg,
		energyPrice: msg.EnergyPrice(),
		value:       msg.Value(),
		data:        msg.Data(),
		state:       cvm.StateDB,
	}
}

// ApplyMessage computes the new state by applying the given message
// against the old state within the environment.
//
// ApplyMessage returns the bytes returned by any CVM execution (if it took place),
// the energy used (which includes energy refunds) and an error if it failed. An error always
// indicates a core error meaning that the message would always fail for that particular
// state and would never be accepted within a block.
func ApplyMessage(cvm *vm.CVM, msg Message, gp *EnergyPool) (*ExecutionResult, error) {
	return NewStateTransition(cvm, msg, gp).TransitionDb()
}

// to returns the recipient of the message.
func (st *StateTransition) to() common.Address {
	if st.msg == nil || st.msg.To() == nil /* contract creation */ {
		return common.Address{}
	}
	return *st.msg.To()
}

func (st *StateTransition) buyEnergy() error {
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.msg.Energy()), st.energyPrice)
	if have, want := st.state.GetBalance(st.msg.From()), mgval; have.Cmp(want) < 0 {
		return fmt.Errorf("%w: address %v have %v want %v", ErrInsufficientFunds, st.msg.From().Hex(), have, want)
	}
	if err := st.gp.SubEnergy(st.msg.Energy()); err != nil {
		return err
	}
	st.energy += st.msg.Energy()

	st.initialEnergy = st.msg.Energy()
	st.state.SubBalance(st.msg.From(), mgval)
	return nil
}

func (st *StateTransition) preCheck() error {
	// Make sure this transaction's nonce is correct.
	if st.msg.CheckNonce() {
		stNonce := st.state.GetNonce(st.msg.From())
		if msgNonce := st.msg.Nonce(); stNonce < msgNonce {
			return fmt.Errorf("%w: address %v, tx: %d state: %d", ErrNonceTooHigh,
				st.msg.From().Hex(), msgNonce, stNonce)
		} else if stNonce > msgNonce {
			return fmt.Errorf("%w: address %v, tx: %d state: %d", ErrNonceTooLow,
				st.msg.From().Hex(), msgNonce, stNonce)
		}
	}
	return st.buyEnergy()
}

// TransitionDb will transition the state by applying the current message and
// returning the cvm execution result with following fields.
//
//   - used energy:
//     total energy used (including energy being refunded)
//   - returndata:
//     the returned data from cvm
//   - concrete execution error:
//     various **CVM** error which aborts the execution,
//     e.g. ErrOutOfEnergy, ErrExecutionReverted
//
// However if any consensus issue encountered, return the error directly with
// nil cvm execution result.
func (st *StateTransition) TransitionDb() (*ExecutionResult, error) {
	// First check this message satisfies all consensus rules before
	// applying the message. The rules include these clauses
	//
	// 1. the nonce of the message caller is correct
	// 2. caller has enough balance to cover transaction fee(energylimit * energyprice)
	// 3. the amount of energy required is available in the block
	// 4. the purchased energy is enough to cover intrinsic usage
	// 5. there is no overflow when calculating intrinsic energy
	// 6. caller has enough balance to cover asset transfer for **topmost** call

	// Check clauses 1-3, buy energy if everything is correct
	if err := st.preCheck(); err != nil {
		return nil, err
	}
	msg := st.msg
	sender := vm.AccountRef(msg.From())
	contractCreation := msg.To() == nil

	// Check clauses 4-5, subtract intrinsic energy if everything is correct
	energy, err := IntrinsicEnergy(st.data, contractCreation)
	if err != nil {
		return nil, err
	}
	if st.energy < energy {
		return nil, fmt.Errorf("%w: have %d, want %d", ErrIntrinsicEnergy, st.energy, energy)
	}
	st.energy -= energy

	// Check clause 6
	if msg.Value().Sign() > 0 && !st.cvm.Context.CanTransfer(st.state, msg.From(), msg.Value()) {
		return nil, fmt.Errorf("%w: address %v", ErrInsufficientFundsForTransfer, msg.From().Hex())
	}
	var (
		ret   []byte
		vmerr error // vm errors do not effect consensus and are therefore not assigned to err
	)
	if contractCreation {
		ret, _, st.energy, vmerr = st.cvm.Create(sender, st.data, st.energy, st.value)
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
		ret, st.energy, vmerr = st.cvm.Call(sender, st.to(), st.data, st.energy, st.value)
	}
	st.refundEnergy()
	st.state.AddBalance(st.cvm.Context.Coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.energyUsed()), st.energyPrice))

	return &ExecutionResult{
		UsedEnergy: st.energyUsed(),
		Err:        vmerr,
		ReturnData: ret,
	}, nil
}

func (st *StateTransition) refundEnergy() {
	// Apply refund counter, capped to half of the used energy.
	refund := st.energyUsed() / 2
	if refund > st.state.GetRefund() {
		refund = st.state.GetRefund()
	}
	st.energy += refund

	// Return XCB for remaining energy, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.energy), st.energyPrice)
	st.state.AddBalance(st.msg.From(), remaining)

	// Also return remaining energy to the block energy counter so it is
	// available for the next transaction.
	st.gp.AddEnergy(st.energy)
}

// energyUsed returns the amount of energy used up by the state transition.
func (st *StateTransition) energyUsed() uint64 {
	return st.initialEnergy - st.energy
}
