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
	"errors"
	"math"
	"math/big"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/core/vm"
	"github.com/core-coin/go-core/log"
	"github.com/core-coin/go-core/params"
)

var (
	errInsufficientBalanceForEnergy = errors.New("insufficient balance to pay for energy")
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
			return 0, vm.ErrOutOfEnergy
		}
		energy += nz * nonZeroEnergy

		z := uint64(len(data)) - nz
		if (math.MaxUint64-energy)/params.TxDataZeroEnergy < z {
			return 0, vm.ErrOutOfEnergy
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
func ApplyMessage(cvm *vm.CVM, msg Message, gp *EnergyPool) ([]byte, uint64, bool, error) {
	return NewStateTransition(cvm, msg, gp).TransitionDb()
}

// to returns the recipient of the message.
func (st *StateTransition) to() common.Address {
	if st.msg == nil || st.msg.To() == nil /* contract creation */ {
		return common.Address{}
	}
	return *st.msg.To()
}

func (st *StateTransition) useEnergy(amount uint64) error {
	if st.energy < amount {
		return vm.ErrOutOfEnergy
	}
	st.energy -= amount

	return nil
}

func (st *StateTransition) buyEnergy() error {
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.msg.Energy()), st.energyPrice)
	if st.state.GetBalance(st.msg.From()).Cmp(mgval) < 0 {
		return errInsufficientBalanceForEnergy
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
		nonce := st.state.GetNonce(st.msg.From())
		if nonce < st.msg.Nonce() {
			return ErrNonceTooHigh
		} else if nonce > st.msg.Nonce() {
			return ErrNonceTooLow
		}
	}
	return st.buyEnergy()
}

// TransitionDb will transition the state by applying the current message and
// returning the result including the used energy. It returns an error if failed.
// An error indicates a consensus issue.
func (st *StateTransition) TransitionDb() (ret []byte, usedEnergy uint64, failed bool, err error) {
	if err = st.preCheck(); err != nil {
		return
	}
	msg := st.msg
	sender := vm.AccountRef(msg.From())
	contractCreation := msg.To() == nil

	// Pay intrinsic energy
	energy, err := IntrinsicEnergy(st.data, contractCreation)
	if err != nil {
		return nil, 0, false, err
	}
	if err = st.useEnergy(energy); err != nil {
		return nil, 0, false, err
	}

	var (
		cvm = st.cvm
		// vm errors do not effect consensus and are therefor
		// not assigned to err, except for insufficient balance
		// error.
		vmerr error
	)
	if contractCreation {
		ret, _, st.energy, vmerr = cvm.Create(sender, st.data, st.energy, st.value)
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
		ret, st.energy, vmerr = cvm.Call(sender, st.to(), st.data, st.energy, st.value)
	}
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, vmerr
		}
	}
	st.refundEnergy()
	st.state.AddBalance(st.cvm.Coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.energyUsed()), st.energyPrice))

	return ret, st.energyUsed(), vmerr != nil, err
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
