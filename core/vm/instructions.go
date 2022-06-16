// Copyright 2015 by the Authors
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
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/params"
	"github.com/core-coin/uint256"
	"golang.org/x/crypto/sha3"
)

var (
	errWriteProtection       = errors.New("cvm: write protection")
	errReturnDataOutOfBounds = errors.New("cvm: return data out of bounds")
	errExecutionReverted     = errors.New("cvm: execution reverted")
	errMaxCodeSizeExceeded   = errors.New("cvm: max code size exceeded")
	errInvalidJump           = errors.New("cvm: invalid jump destination")
)

func opAdd(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Add(&x, y)
	return nil, nil
}

func opSub(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Sub(&x, y)
	return nil, nil
}

func opMul(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Mul(&x, y)

	return nil, nil
}

func opDiv(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Div(&x, y)
	return nil, nil
}

func opSdiv(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.SDiv(&x, y)
	return nil, nil
}

func opMod(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Mod(&x, y)
	return nil, nil
}

func opSmod(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.SMod(&x, y)
	return nil, nil
}

func opExp(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	base, exponent := stack.pop(), stack.peek()
	exponent.Exp(&base, exponent)
	return nil, nil
}

func opSignExtend(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	back, num := stack.pop(), stack.peek()
	num.ExtendSign(num, &back)
	return nil, nil
}

func opNot(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := stack.peek()
	x.Not(x)
	return nil, nil
}

func opLt(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Lt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opGt(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Gt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opSlt(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Slt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opSgt(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Sgt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opEq(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Eq(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opIszero(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := stack.peek()
	if x.IsZero() {
		x.SetOne()
	} else {
		x.Clear()
	}
	return nil, nil
}

func opAnd(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.And(&x, y)
	return nil, nil
}

func opOr(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Or(&x, y)
	return nil, nil
}

func opXor(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Xor(&x, y)
	return nil, nil
}

func opByte(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	th, val := stack.pop(), stack.peek()
	val.Byte(&th)
	return nil, nil
}

func opAddmod(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y, z := stack.pop(), stack.pop(), stack.peek()
	if z.IsZero() {
		z.Clear()
	} else {
		z.AddMod(&x, &y, z)
	}
	return nil, nil
}

func opMulmod(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y, z := stack.pop(), stack.pop(), stack.peek()
	z.MulMod(&x, &y, z)
	return nil, nil
}

// opSHL implements Shift Left
// The SHL instruction (shift left) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the left by arg1 number of bits.
func opSHL(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := stack.pop(), stack.peek()
	if shift.LtUint64(256) {
		value.Lsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}

	return nil, nil
}

// opSHR implements Logical Shift Right
// The SHR instruction (logical shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with zero fill.
func opSHR(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := stack.pop(), stack.peek()
	if shift.LtUint64(256) {
		value.Rsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}

	return nil, nil
}

// opSAR implements Arithmetic Shift Right
// The SAR instruction (arithmetic shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with sign extension.
func opSAR(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Note, S256 returns (potentially) a new bigint, so we're popping, not peeking this one
	shift, value := stack.pop(), stack.peek()
	if shift.GtUint64(256) {
		if value.Sign() >= 0 {
			value.Clear()
		} else {
			// Max negative shift: all bits set
			value.SetAllOne()
		}
		return nil, nil
	}
	n := uint(shift.Uint64())
	value.SRsh(value, n)

	return nil, nil
}

func opSha3(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.peek()
	data := memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))

	if interpreter.hasher == nil {
		interpreter.hasher = sha3.New256().(keccakState)
	} else {
		interpreter.hasher.Reset()
	}
	interpreter.hasher.Write(data)
	interpreter.hasher.Read(interpreter.hasherBuf[:])

	cvm := interpreter.cvm
	if cvm.vmConfig.EnablePreimageRecording {
		cvm.StateDB.AddPreimage(interpreter.hasherBuf, data)
	}
	size.SetBytes(interpreter.hasherBuf[:])
	return nil, nil
}

func opAddress(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(new(uint256.Int).SetBytes(contract.Address().Bytes()))
	return nil, nil
}

func opBalance(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	slot := stack.peek()
	address := common.Address(slot.Bytes22())
	slot.SetFromBig(interpreter.cvm.StateDB.GetBalance(address))
	return nil, nil
}

func opOrigin(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(new(uint256.Int).SetBytes(interpreter.cvm.Origin.Bytes()))
	return nil, nil
}

func opCaller(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(new(uint256.Int).SetBytes(contract.Caller().Bytes()))
	return nil, nil
}

func opCallValue(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	v, _ := uint256.FromBig(contract.value)
	stack.push(v)
	return nil, nil
}

func opCallDataLoad(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := stack.peek()
	if offset, overflow := x.Uint64WithOverflow(); !overflow {
		data := getData(contract.Input, offset, 32)
		x.SetBytes(data)
	} else {
		x.Clear()
	}
	return nil, nil
}

func opCallDataSize(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(new(uint256.Int).SetUint64(uint64(len(contract.Input))))
	return nil, nil
}

func opCallDataCopy(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		dataOffset = stack.pop()
		length     = stack.pop()
	)
	dataOffset64, overflow := dataOffset.Uint64WithOverflow()
	if overflow {
		dataOffset64 = 0xffffffffffffffff
	}
	// These values are checked for overflow during energy cost calculation
	memOffset64 := memOffset.Uint64()
	length64 := length.Uint64()
	memory.Set(memOffset64, length64, getData(contract.Input, dataOffset64, length64))
	return nil, nil
}

func opReturnDataSize(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(new(uint256.Int).SetUint64(uint64(len(interpreter.returnData))))
	return nil, nil
}

func opReturnDataCopy(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		dataOffset = stack.pop()
		length     = stack.pop()
	)

	offset64, overflow := dataOffset.Uint64WithOverflow()
	if overflow {
		return nil, errReturnDataOutOfBounds
	}
	// we can reuse dataOffset now (aliasing it for clarity)
	var end = dataOffset
	end.Add(&dataOffset, &length)
	end64, overflow := end.Uint64WithOverflow()
	if overflow || uint64(len(interpreter.returnData)) < end64 {
		return nil, errReturnDataOutOfBounds
	}
	memory.Set(memOffset.Uint64(), length.Uint64(), interpreter.returnData[offset64:end64])
	return nil, nil
}

func opExtCodeSize(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	slot := stack.peek()
	slot.SetUint64(uint64(interpreter.cvm.StateDB.GetCodeSize(common.Address(slot.Bytes22()))))

	return nil, nil
}

func opCodeSize(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	l := new(uint256.Int)
	l.SetUint64(uint64(len(contract.Code)))
	stack.push(l)

	return nil, nil
}

func opCodeCopy(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	uint64CodeOffset, overflow := codeOffset.Uint64WithOverflow()
	if overflow {
		uint64CodeOffset = 0xffffffffffffffff
	}
	codeCopy := getData(contract.Code, uint64CodeOffset, length.Uint64())
	memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	return nil, nil
}

func opExtCodeCopy(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		a          = stack.pop()
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	uint64CodeOffset, overflow := codeOffset.Uint64WithOverflow()
	if overflow {
		uint64CodeOffset = 0xffffffffffffffff
	}
	addr := common.Address(a.Bytes22())
	codeCopy := getData(interpreter.cvm.StateDB.GetCode(addr), uint64CodeOffset, length.Uint64())
	memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	return nil, nil
}

// opExtCodeHash returns the code hash of a specified account.
// There are several cases when the function is called, while we can relay everything
// to `state.GetCodeHash` function to ensure the correctness.
//   (1) Caller tries to get the code hash of a normal contract account, state
// should return the relative code hash and set it as the result.
//
//   (2) Caller tries to get the code hash of a non-existent account, state should
// return common.Hash{} and zero will be set as the result.
//
//   (3) Caller tries to get the code hash for an account without contract code,
// state should return emptyCodeHash(0xc5d246...) as the result.
//
//   (4) Caller tries to get the code hash of a precompiled account, the result
// should be zero or emptyCodeHash.
//
// It is worth noting that in order to avoid unnecessary create and clean,
// all precompile accounts on mainnet have been transferred 1 ore, so the return
// here should be emptyCodeHash.
// If the precompile account is not transferred any amount on a private or
// customized chain, the return value will be zero.
//
//   (5) Caller tries to get the code hash for an account which is marked as suicided
// in the current transaction, the code hash of this account should be returned.
//
//   (6) Caller tries to get the code hash for an account which is marked as deleted,
// this account should be regarded as a non-existent account and zero should be returned.
func opExtCodeHash(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	slot := stack.peek()
	address := common.Address(slot.Bytes22())
	if interpreter.cvm.StateDB.Empty(address) {
		slot.Clear()
	} else {
		slot.SetBytes(interpreter.cvm.StateDB.GetCodeHash(address).Bytes())
	}
	return nil, nil
}

func opEnergyprice(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.cvm.EnergyPrice)
	stack.push(v)
	return nil, nil
}

func opBlockhash(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	num := stack.peek()
	num64, overflow := num.Uint64WithOverflow()
	if overflow {
		num.Clear()
		return nil, nil
	}
	var upper, lower uint64
	upper = interpreter.cvm.BlockNumber.Uint64()
	if upper < 257 {
		lower = 0
	} else {
		lower = upper - 256
	}
	if num64 >= lower && num64 < upper {
		num.SetBytes(interpreter.cvm.GetHash(num64).Bytes())
	} else {
		num.Clear()
	}
	return nil, nil
}

func opCoinbase(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(new(uint256.Int).SetBytes(interpreter.cvm.Coinbase.Bytes()))
	return nil, nil
}

func opTimestamp(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.cvm.Time)
	stack.push(v)
	return nil, nil
}

func opNumber(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.cvm.BlockNumber)
	stack.push(v)
	return nil, nil
}

func opDifficulty(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.cvm.Difficulty)
	stack.push(v)
	return nil, nil
}

func opEnergyLimit(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(new(uint256.Int).SetUint64(interpreter.cvm.EnergyLimit))
	return nil, nil
}

func opPop(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.pop()
	return nil, nil
}

func opMload(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	v := stack.peek()
	offset := int64(v.Uint64())
	v.SetBytes(memory.GetPtr(offset, 32))
	return nil, nil
}

func opMstore(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// pop value of the stack
	mStart, val := stack.pop(), stack.pop()
	memory.Set32(mStart.Uint64(), &val)
	return nil, nil
}

func opMstore8(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	off, val := stack.pop(), stack.pop()
	memory.store[off.Uint64()] = byte(val.Uint64())
	return nil, nil
}

func opSload(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	loc := stack.peek()
	hash := common.Hash(loc.Bytes32())
	val := interpreter.cvm.StateDB.GetState(contract.Address(), hash)
	loc.SetBytes(val.Bytes())
	return nil, nil
}

func opSstore(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	loc := stack.pop()
	val := stack.pop()
	interpreter.cvm.StateDB.SetState(contract.Address(),
		common.Hash(loc.Bytes32()), common.Hash(val.Bytes32()))
	return nil, nil
}

func opJump(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	pos := stack.pop()
	if !contract.validJumpdest(&pos) {
		return nil, errInvalidJump
	}
	*pc = pos.Uint64()

	return nil, nil
}

func opJumpi(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	pos, cond := stack.pop(), stack.pop()
	if !cond.IsZero() {
		if !contract.validJumpdest(&pos) {
			return nil, errInvalidJump
		}
		*pc = pos.Uint64()
	} else {
		*pc++
	}

	return nil, nil
}

func opJumpdest(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	return nil, nil
}

func opPc(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(new(uint256.Int).SetUint64(*pc))
	return nil, nil
}

func opMsize(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(new(uint256.Int).SetUint64(uint64(memory.Len())))
	return nil, nil
}

func opEnergy(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(new(uint256.Int).SetUint64(contract.Energy))
	return nil, nil
}

func opCreate(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		value        = stack.pop()
		offset, size = stack.pop(), stack.pop()
		input        = memory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()))
		energy       = contract.Energy
	)

	energy -= energy / 64
	// reuse size int for stackvalue
	stackvalue := size

	contract.UseEnergy(energy)
	res, addr, returnEnergy, suberr := interpreter.cvm.Create(contract, input, energy, value.ToBig())
	// Push item on the stack based on the returned error. We must
	// ignore this error and pretend the operation was successful.
	if suberr == ErrCodeStoreOutOfEnergy {
		stackvalue.Clear()
	} else if suberr != nil && suberr != ErrCodeStoreOutOfEnergy {
		stackvalue.Clear()
	} else {
		stackvalue.SetBytes(addr.Bytes())
	}
	stack.push(&stackvalue)
	contract.Energy += returnEnergy

	if suberr == errExecutionReverted {
		return res, nil
	}
	return nil, nil
}

func opCreate2(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		endowment    = stack.pop()
		offset, size = stack.pop(), stack.pop()
		salt         = stack.pop()
		input        = memory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()))
		energy       = contract.Energy
	)

	// Apply CIP150
	energy -= energy / 64
	contract.UseEnergy(energy)
	// reuse size int for stackvalue
	stackvalue := size
	res, addr, returnEnergy, suberr := interpreter.cvm.Create2(contract, input, energy,
		endowment.ToBig(), salt.ToBig())
	// Push item on the stack based on the returned error.
	if suberr != nil {
		stackvalue.Clear()
	} else {
		stackvalue.SetBytes(addr.Bytes())
	}
	stack.push(&stackvalue)
	contract.Energy += returnEnergy

	if suberr == errExecutionReverted {
		return res, nil
	}
	return nil, nil
}

func opCall(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop energy. The actual energy in interpreter.cvm.callEnergyTemp.
	// We can use this as a temporary value
	temp := stack.pop()
	energy := interpreter.cvm.callEnergyTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes22())
	// Get the arguments from the memory.
	args := memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	if !value.IsZero() {
		energy += params.CallStipend
	}
	ret, returnEnergy, err := interpreter.cvm.Call(contract, toAddr, args, energy, value.ToBig())
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Energy += returnEnergy
	return ret, nil
}

func opCallCode(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop energy. The actual energy is in interpreter.cvm.callEnergyTemp.
	// We use it as a temporary value
	temp := stack.pop()
	energy := interpreter.cvm.callEnergyTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes22())
	// Get arguments from the memory.
	args := memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	if !value.IsZero() {
		energy += params.CallStipend
	}
	ret, returnEnergy, err := interpreter.cvm.CallCode(contract, toAddr, args, energy, value.ToBig())
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Energy += returnEnergy
	return ret, nil
}

func opDelegateCall(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop energy. The actual energy is in interpreter.cvm.callEnergyTemp.
	// We use it as a temporary value
	temp := stack.pop()
	energy := interpreter.cvm.callEnergyTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes22())
	// Get arguments from the memory.
	args := memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	ret, returnEnergy, err := interpreter.cvm.DelegateCall(contract, toAddr, args, energy)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Energy += returnEnergy
	return ret, nil
}

func opStaticCall(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop energy. The actual energy is in interpreter.cvm.callEnergyTemp.
	// We use it as a temporary value
	temp := stack.pop()
	energy := interpreter.cvm.callEnergyTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes22())
	// Get arguments from the memory.
	args := memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	ret, returnEnergy, err := interpreter.cvm.StaticCall(contract, toAddr, args, energy)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Energy += returnEnergy
	return ret, nil
}

func opReturn(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))
	return ret, nil
}

func opRevert(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))
	return ret, nil
}

func opStop(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	return nil, nil
}

func opSuicide(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	beneficiary := stack.pop()
	balance := interpreter.cvm.StateDB.GetBalance(contract.Address())
	interpreter.cvm.StateDB.AddBalance(common.Address(beneficiary.Bytes22()), balance)

	interpreter.cvm.StateDB.Suicide(contract.Address())
	return nil, nil
}

// following functions are used by the instruction jump  table

// make log instruction function
func makeLog(size int) executionFunc {
	return func(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		topics := make([]common.Hash, size)
		mStart, mSize := stack.pop(), stack.pop()
		for i := 0; i < size; i++ {
			addr := stack.pop()
			topics[i] = common.Hash(addr.Bytes32())
		}

		d := memory.GetCopy(int64(mStart.Uint64()), int64(mSize.Uint64()))
		interpreter.cvm.StateDB.AddLog(&types.Log{
			Address: contract.Address(),
			Topics:  topics,
			Data:    d,
			// This is a non-consensus field, but assigned here because
			// core/state doesn't know the current block number.
			BlockNumber: interpreter.cvm.BlockNumber.Uint64(),
		})

		return nil, nil
	}
}

// opPush1 is a specialized version of pushN
func opPush1(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		codeLen = uint64(len(contract.Code))
		integer = new(uint256.Int)
	)
	*pc += 1
	if *pc < codeLen {
		stack.push(integer.SetUint64(uint64(contract.Code[*pc])))
	} else {
		stack.push(integer.Clear())
	}
	return nil, nil
}

// make push instruction function
func makePush(size uint64, pushByteSize int) executionFunc {
	return func(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		codeLen := len(contract.Code)

		startMin := codeLen
		if int(*pc+1) < startMin {
			startMin = int(*pc + 1)
		}

		endMin := codeLen
		if startMin+pushByteSize < endMin {
			endMin = startMin + pushByteSize
		}

		integer := new(uint256.Int)
		stack.push(integer.SetBytes(common.RightPadBytes(
			contract.Code[startMin:endMin], pushByteSize)))

		*pc += size
		return nil, nil
	}
}

// make dup instruction function
func makeDup(size int64) executionFunc {
	return func(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		stack.dup(int(size))
		return nil, nil
	}
}

// make swap instruction function
func makeSwap(size int64) executionFunc {
	// switch n + 1 otherwise n would be swapped with n
	size++
	return func(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		stack.swap(int(size))
		return nil, nil
	}
}
