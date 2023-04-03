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

package vm

import (
	"errors"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/math"
	"github.com/core-coin/go-core/v2/params"
)

// memoryEnergyCost calculates the quadratic energy for memory expansion. It does so
// only for the memory region that is expanded, not the total memory.
func memoryEnergyCost(mem *Memory, newMemSize uint64) (uint64, error) {
	if newMemSize == 0 {
		return 0, nil
	}
	// The maximum that will fit in a uint64 is max_word_count - 1. Anything above
	// that will result in an overflow. Additionally, a newMemSize which results in
	// a newMemSizeWords larger than 0xFFFFFFFF will cause the square operation to
	// overflow. The constant 0x1FFFFFFFE0 is the highest number that can be used
	// without overflowing the energy calculation.
	if newMemSize > 0x1FFFFFFFE0 {
		return 0, ErrEnergyUintOverflow
	}
	newMemSizeWords := toWordSize(newMemSize)
	newMemSize = newMemSizeWords * 32

	if newMemSize > uint64(mem.Len()) {
		square := newMemSizeWords * newMemSizeWords
		linCoef := newMemSizeWords * params.MemoryEnergy
		quadCoef := square / params.QuadCoeffDiv
		newTotalFee := linCoef + quadCoef

		fee := newTotalFee - mem.lastEnergyCost
		mem.lastEnergyCost = newTotalFee

		return fee, nil
	}
	return 0, nil
}

// memoryCopierEnergy creates the energy functions for the following opcodes, and takes
// the stack position of the operand which determines the size of the data to copy
// as argument:
// CALLDATACOPY (stack position 2)
// CODECOPY (stack position 2)
// EXTCODECOPY (stack poition 3)
// RETURNDATACOPY (stack position 2)
func memoryCopierEnergy(stackpos int) energyFunc {
	return func(cvm *CVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
		// Energy for expanding the memory
		energy, err := memoryEnergyCost(mem, memorySize)
		if err != nil {
			return 0, err
		}
		// And energy for copying data, charged per word at param.CopyEnergy
		words, overflow := stack.Back(stackpos).Uint64WithOverflow()
		if overflow {
			return 0, ErrEnergyUintOverflow
		}

		if words, overflow = math.SafeMul(toWordSize(words), params.CopyEnergy); overflow {
			return 0, ErrEnergyUintOverflow
		}

		if energy, overflow = math.SafeAdd(energy, words); overflow {
			return 0, ErrEnergyUintOverflow
		}
		return energy, nil
	}
}

var (
	energyCallDataCopy   = memoryCopierEnergy(2)
	energyCodeCopy       = memoryCopierEnergy(2)
	energyExtCodeCopy    = memoryCopierEnergy(3)
	energyReturnDataCopy = memoryCopierEnergy(2)
)

//  0. If *energyleft* is less than or equal to 2300, fail the current call.
//  1. If current value equals new value (this is a no-op), SSTORE_NOOP_ENERGY energy is deducted.
//  2. If current value does not equal new value:
//     2.1. If original value equals current value (this storage slot has not been changed by the current execution context):
//     2.1.1. If original value is 0, SSTORE_INIT_ENERGY energy is deducted.
//     2.1.2. Otherwise, SSTORE_CLEAN_ENERGY energy is deducted. If new value is 0, add SSTORE_CLEAR_REFUND to refund counter.
//     2.2. If original value does not equal current value (this storage slot is dirty), SSTORE_DIRTY_ENERGY energy is deducted. Apply both of the following clauses:
//     2.2.1. If original value is not 0:
//     2.2.1.1. If current value is 0 (also means that new value is not 0), subtract SSTORE_CLEAR_REFUND energy from refund counter. We can prove that refund counter will never go below 0.
//     2.2.1.2. If new value is 0 (also means that current value is not 0), add SSTORE_CLEAR_REFUND energy to refund counter.
//     2.2.2. If original value equals new value (this storage slot is reset):
//     2.2.2.1. If original value is 0, add SSTORE_INIT_REFUND to refund counter.
//     2.2.2.2. Otherwise, add SSTORE_CLEAN_REFUND energy to refund counter.
func energySStore(cvm *CVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	// If we fail the minimum energy availability invariant, fail (0)
	if contract.Energy <= params.SstoreSentryEnergy {
		return 0, errors.New("not enough energy for reentrancy sentry")
	}
	// Energy sentry honoured, do the actual energy calculation based on the stored value
	var (
		y, x    = stack.Back(1), stack.Back(0)
		current = cvm.StateDB.GetState(contract.Address(), x.Bytes32())
	)
	value := common.Hash(y.Bytes32())

	if current == value { // noop (1)
		return params.SstoreNoopEnergy, nil
	}
	original := cvm.StateDB.GetCommittedState(contract.Address(), x.Bytes32())
	if original == current {
		if original == (common.Hash{}) { // create slot (2.1.1)
			return params.SstoreInitEnergy, nil
		}
		if value == (common.Hash{}) { // delete slot (2.1.2b)
			cvm.StateDB.AddRefund(params.SstoreClearRefund)
		}
		return params.SstoreCleanEnergy, nil // write existing slot (2.1.2)
	}
	if original != (common.Hash{}) {
		if current == (common.Hash{}) { // recreate slot (2.2.1.1)
			cvm.StateDB.SubRefund(params.SstoreClearRefund)
		} else if value == (common.Hash{}) { // delete slot (2.2.1.2)
			cvm.StateDB.AddRefund(params.SstoreClearRefund)
		}
	}
	if original == value {
		if original == (common.Hash{}) { // reset to original inexistent slot (2.2.2.1)
			cvm.StateDB.AddRefund(params.SstoreInitRefund)
		} else { // reset to original existing slot (2.2.2.2)
			cvm.StateDB.AddRefund(params.SstoreCleanRefund)
		}
	}
	return params.SstoreDirtyEnergy, nil // dirty update (2.2)
}

func makeEnergyLog(n uint64) energyFunc {
	return func(cvm *CVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
		requestedSize, overflow := stack.Back(1).Uint64WithOverflow()
		if overflow {
			return 0, ErrEnergyUintOverflow
		}

		energy, err := memoryEnergyCost(mem, memorySize)
		if err != nil {
			return 0, err
		}

		if energy, overflow = math.SafeAdd(energy, params.LogEnergy); overflow {
			return 0, ErrEnergyUintOverflow
		}
		if energy, overflow = math.SafeAdd(energy, n*params.LogTopicEnergy); overflow {
			return 0, ErrEnergyUintOverflow
		}

		var memorySizeEnergy uint64
		if memorySizeEnergy, overflow = math.SafeMul(requestedSize, params.LogDataEnergy); overflow {
			return 0, ErrEnergyUintOverflow
		}
		if energy, overflow = math.SafeAdd(energy, memorySizeEnergy); overflow {
			return 0, ErrEnergyUintOverflow
		}
		return energy, nil
	}
}

func energySha3(cvm *CVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	energy, err := memoryEnergyCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	wordEnergy, overflow := stack.Back(1).Uint64WithOverflow()
	if overflow {
		return 0, ErrEnergyUintOverflow
	}
	if wordEnergy, overflow = math.SafeMul(toWordSize(wordEnergy), params.Sha3WordEnergy); overflow {
		return 0, ErrEnergyUintOverflow
	}
	if energy, overflow = math.SafeAdd(energy, wordEnergy); overflow {
		return 0, ErrEnergyUintOverflow
	}
	return energy, nil
}

// pureMemoryEnergycost is used by several operations, which aside from their
// static cost have a dynamic cost which is solely based on the memory
// expansion
func pureMemoryEnergycost(cvm *CVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return memoryEnergyCost(mem, memorySize)
}

var (
	energyReturn  = pureMemoryEnergycost
	energyRevert  = pureMemoryEnergycost
	energyMLoad   = pureMemoryEnergycost
	energyMStore8 = pureMemoryEnergycost
	energyMStore  = pureMemoryEnergycost
	energyCreate  = pureMemoryEnergycost
)

func energyCreate2(cvm *CVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	energy, err := memoryEnergyCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	wordEnergy, overflow := stack.Back(2).Uint64WithOverflow()
	if overflow {
		return 0, ErrEnergyUintOverflow
	}
	if wordEnergy, overflow = math.SafeMul(toWordSize(wordEnergy), params.Sha3WordEnergy); overflow {
		return 0, ErrEnergyUintOverflow
	}
	if energy, overflow = math.SafeAdd(energy, wordEnergy); overflow {
		return 0, ErrEnergyUintOverflow
	}
	return energy, nil
}

func energyExp(cvm *CVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	expByteLen := uint64((stack.data[stack.len()-2].BitLen() + 7) / 8)

	var (
		energy   = expByteLen * params.ExpByte // no overflow check required. Max is 256 * ExpByte energy
		overflow bool
	)
	if energy, overflow = math.SafeAdd(energy, params.ExpEnergy); overflow {
		return 0, ErrEnergyUintOverflow
	}
	return energy, nil
}

func energyCall(cvm *CVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var (
		energy         uint64
		transfersValue = !stack.Back(2).IsZero()
		address        = common.Address(stack.Back(1).Bytes22())
	)
	if transfersValue && cvm.StateDB.Empty(address) {
		energy += params.CallNewAccountEnergy
	}
	if transfersValue {
		energy += params.CallValueTransferEnergy
	}
	memoryEnergy, err := memoryEnergyCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if energy, overflow = math.SafeAdd(energy, memoryEnergy); overflow {
		return 0, ErrEnergyUintOverflow
	}

	cvm.callEnergyTemp, err = callEnergy(contract.Energy, energy, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if energy, overflow = math.SafeAdd(energy, cvm.callEnergyTemp); overflow {
		return 0, ErrEnergyUintOverflow
	}
	return energy, nil
}

func energyCallCode(cvm *CVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	memoryEnergy, err := memoryEnergyCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var (
		energy   uint64
		overflow bool
	)
	if stack.Back(2).Sign() != 0 {
		energy += params.CallValueTransferEnergy
	}
	if energy, overflow = math.SafeAdd(energy, memoryEnergy); overflow {
		return 0, ErrEnergyUintOverflow
	}
	cvm.callEnergyTemp, err = callEnergy(contract.Energy, energy, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if energy, overflow = math.SafeAdd(energy, cvm.callEnergyTemp); overflow {
		return 0, ErrEnergyUintOverflow
	}
	return energy, nil
}

func energyDelegateCall(cvm *CVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	energy, err := memoryEnergyCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	cvm.callEnergyTemp, err = callEnergy(contract.Energy, energy, stack.Back(0))
	if err != nil {
		return 0, err
	}
	var overflow bool
	if energy, overflow = math.SafeAdd(energy, cvm.callEnergyTemp); overflow {
		return 0, ErrEnergyUintOverflow
	}
	return energy, nil
}

func energyStaticCall(cvm *CVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	energy, err := memoryEnergyCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	cvm.callEnergyTemp, err = callEnergy(contract.Energy, energy, stack.Back(0))
	if err != nil {
		return 0, err
	}
	var overflow bool
	if energy, overflow = math.SafeAdd(energy, cvm.callEnergyTemp); overflow {
		return 0, ErrEnergyUintOverflow
	}
	return energy, nil
}

func energySelfdestruct(cvm *CVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var energy uint64
	energy = params.SelfdestructEnergy
	var address = common.Address(stack.Back(0).Bytes22())

	// if empty and transfers value
	if cvm.StateDB.Empty(address) && cvm.StateDB.GetBalance(contract.Address()).Sign() != 0 {
		energy += params.CreateBySelfdestructEnergy
	}

	if !cvm.StateDB.HasSuicided(contract.Address()) {
		cvm.StateDB.AddRefund(params.SelfdestructRefundEnergy)
	}
	return energy, nil
}
