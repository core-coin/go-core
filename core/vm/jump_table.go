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
	"github.com/core-coin/go-core/v2/params"
)

type (
	executionFunc func(pc *uint64, interpreter *CVMInterpreter, callContext *callCtx) ([]byte, error)
	energyFunc    func(*CVM, *Contract, *Stack, *Memory, uint64) (uint64, error) // last parameter is the requested memory size as a uint64
	// memorySizeFunc returns the required size, and whether the operation overflowed a uint64
	memorySizeFunc func(*Stack) (size uint64, overflow bool)
)

type operation struct {
	// execute is the operation function
	execute        executionFunc
	constantEnergy uint64
	dynamicEnergy  energyFunc
	// minStack tells how many stack items are required
	minStack int
	// maxStack specifies the max length the stack can have for this operation
	// to not overflow the stack.
	maxStack int

	// memorySize returns the memory size required for the operation
	memorySize memorySizeFunc

	halts   bool // indicates whether the operation should halt further execution
	jumps   bool // indicates whether the program counter should not increment
	writes  bool // determines whether this a state modifying operation
	reverts bool // determines whether the operation reverts state (implicitly halts)
	returns bool // determines whether the operations sets the return data content
}

var (
	InstructionSet = newInstructionSet()
)

// JumpTable contains the CVM opcodes supported at a given fork.
type JumpTable [256]*operation

// newInstructionSet returns the instructions.
func newInstructionSet() JumpTable {
	return JumpTable{
		STOP: {
			execute:        opStop,
			constantEnergy: 0,
			minStack:       minStack(0, 0),
			maxStack:       maxStack(0, 0),
			halts:          true,
		},
		ADD: {
			execute:        opAdd,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		MUL: {
			execute:        opMul,
			constantEnergy: EnergyFastStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		SUB: {
			execute:        opSub,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		DIV: {
			execute:        opDiv,
			constantEnergy: EnergyFastStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		SDIV: {
			execute:        opSdiv,
			constantEnergy: EnergyFastStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		MOD: {
			execute:        opMod,
			constantEnergy: EnergyFastStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		SMOD: {
			execute:        opSmod,
			constantEnergy: EnergyFastStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		ADDMOD: {
			execute:        opAddmod,
			constantEnergy: EnergyMidStep,
			minStack:       minStack(3, 1),
			maxStack:       maxStack(3, 1),
		},
		MULMOD: {
			execute:        opMulmod,
			constantEnergy: EnergyMidStep,
			minStack:       minStack(3, 1),
			maxStack:       maxStack(3, 1),
		},
		EXP: {
			execute:       opExp,
			dynamicEnergy: energyExp,
			minStack:      minStack(2, 1),
			maxStack:      maxStack(2, 1),
		},
		SIGNEXTEND: {
			execute:        opSignExtend,
			constantEnergy: EnergyFastStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		LT: {
			execute:        opLt,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		GT: {
			execute:        opGt,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		SLT: {
			execute:        opSlt,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		SGT: {
			execute:        opSgt,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		EQ: {
			execute:        opEq,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		ISZERO: {
			execute:        opIszero,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(1, 1),
			maxStack:       maxStack(1, 1),
		},
		AND: {
			execute:        opAnd,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		XOR: {
			execute:        opXor,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		OR: {
			execute:        opOr,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		NOT: {
			execute:        opNot,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(1, 1),
			maxStack:       maxStack(1, 1),
		},
		BYTE: {
			execute:        opByte,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		SHA3: {
			execute:        opSha3,
			constantEnergy: params.Sha3Energy,
			dynamicEnergy:  energySha3,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
			memorySize:     memorySha3,
		},
		ADDRESS: {
			execute:        opAddress,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		BALANCE: {
			execute:        opBalance,
			constantEnergy: params.BalanceEnergy,
			minStack:       minStack(1, 1),
			maxStack:       maxStack(1, 1),
		},
		ORIGIN: {
			execute:        opOrigin,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		CALLER: {
			execute:        opCaller,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		CALLVALUE: {
			execute:        opCallValue,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		CALLDATALOAD: {
			execute:        opCallDataLoad,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(1, 1),
			maxStack:       maxStack(1, 1),
		},
		CALLDATASIZE: {
			execute:        opCallDataSize,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		CALLDATACOPY: {
			execute:        opCallDataCopy,
			constantEnergy: EnergyFastestStep,
			dynamicEnergy:  energyCallDataCopy,
			minStack:       minStack(3, 0),
			maxStack:       maxStack(3, 0),
			memorySize:     memoryCallDataCopy,
		},
		CODESIZE: {
			execute:        opCodeSize,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		CODECOPY: {
			execute:        opCodeCopy,
			constantEnergy: EnergyFastestStep,
			dynamicEnergy:  energyCodeCopy,
			minStack:       minStack(3, 0),
			maxStack:       maxStack(3, 0),
			memorySize:     memoryCodeCopy,
		},
		ENERGYPRICE: {
			execute:        opEnergyprice,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		EXTCODESIZE: {
			execute:        opExtCodeSize,
			constantEnergy: params.ExtcodeSizeEnergy,
			minStack:       minStack(1, 1),
			maxStack:       maxStack(1, 1),
		},
		EXTCODECOPY: {
			execute:        opExtCodeCopy,
			constantEnergy: params.ExtcodeCopyBase,
			dynamicEnergy:  energyExtCodeCopy,
			minStack:       minStack(4, 0),
			maxStack:       maxStack(4, 0),
			memorySize:     memoryExtCodeCopy,
		},
		BLOCKHASH: {
			execute:        opBlockhash,
			constantEnergy: EnergyExtStep,
			minStack:       minStack(1, 1),
			maxStack:       maxStack(1, 1),
		},
		COINBASE: {
			execute:        opCoinbase,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		TIMESTAMP: {
			execute:        opTimestamp,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		NUMBER: {
			execute:        opNumber,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		DIFFICULTY: {
			execute:        opDifficulty,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		ENERGYLIMIT: {
			execute:        opEnergyLimit,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		POP: {
			execute:        opPop,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(1, 0),
			maxStack:       maxStack(1, 0),
		},
		MLOAD: {
			execute:        opMload,
			constantEnergy: EnergyFastestStep,
			dynamicEnergy:  energyMLoad,
			minStack:       minStack(1, 1),
			maxStack:       maxStack(1, 1),
			memorySize:     memoryMLoad,
		},
		MSTORE: {
			execute:        opMstore,
			constantEnergy: EnergyFastestStep,
			dynamicEnergy:  energyMStore,
			minStack:       minStack(2, 0),
			maxStack:       maxStack(2, 0),
			memorySize:     memoryMStore,
		},
		MSTORE8: {
			execute:        opMstore8,
			constantEnergy: EnergyFastestStep,
			dynamicEnergy:  energyMStore8,
			memorySize:     memoryMStore8,
			minStack:       minStack(2, 0),
			maxStack:       maxStack(2, 0),
		},
		SLOAD: {
			execute:        opSload,
			constantEnergy: params.SloadEnergy,
			minStack:       minStack(1, 1),
			maxStack:       maxStack(1, 1),
		},
		SSTORE: {
			execute:       opSstore,
			dynamicEnergy: energySStore,
			minStack:      minStack(2, 0),
			maxStack:      maxStack(2, 0),
			writes:        true,
		},
		JUMP: {
			execute:        opJump,
			constantEnergy: EnergyMidStep,
			minStack:       minStack(1, 0),
			maxStack:       maxStack(1, 0),
			jumps:          true,
		},
		JUMPI: {
			execute:        opJumpi,
			constantEnergy: EnergySlowStep,
			minStack:       minStack(2, 0),
			maxStack:       maxStack(2, 0),
			jumps:          true,
		},
		PC: {
			execute:        opPc,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		MSIZE: {
			execute:        opMsize,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		ENERGY: {
			execute:        opEnergy,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		JUMPDEST: {
			execute:        opJumpdest,
			constantEnergy: params.JumpdestEnergy,
			minStack:       minStack(0, 0),
			maxStack:       maxStack(0, 0),
		},
		PUSH1: {
			execute:        opPush1,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH2: {
			execute:        makePush(2, 2),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH3: {
			execute:        makePush(3, 3),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH4: {
			execute:        makePush(4, 4),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH5: {
			execute:        makePush(5, 5),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH6: {
			execute:        makePush(6, 6),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH7: {
			execute:        makePush(7, 7),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH8: {
			execute:        makePush(8, 8),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH9: {
			execute:        makePush(9, 9),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH10: {
			execute:        makePush(10, 10),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH11: {
			execute:        makePush(11, 11),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH12: {
			execute:        makePush(12, 12),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH13: {
			execute:        makePush(13, 13),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH14: {
			execute:        makePush(14, 14),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH15: {
			execute:        makePush(15, 15),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH16: {
			execute:        makePush(16, 16),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH17: {
			execute:        makePush(17, 17),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH18: {
			execute:        makePush(18, 18),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH19: {
			execute:        makePush(19, 19),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH20: {
			execute:        makePush(20, 20),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH21: {
			execute:        makePush(21, 21),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH22: {
			execute:        makePush(22, 22),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH23: {
			execute:        makePush(23, 23),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH24: {
			execute:        makePush(24, 24),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH25: {
			execute:        makePush(25, 25),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH26: {
			execute:        makePush(26, 26),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH27: {
			execute:        makePush(27, 27),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH28: {
			execute:        makePush(28, 28),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH29: {
			execute:        makePush(29, 29),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH30: {
			execute:        makePush(30, 30),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH31: {
			execute:        makePush(31, 31),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		PUSH32: {
			execute:        makePush(32, 32),
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		DUP1: {
			execute:        makeDup(1),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(1),
			maxStack:       maxDupStack(1),
		},
		DUP2: {
			execute:        makeDup(2),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(2),
			maxStack:       maxDupStack(2),
		},
		DUP3: {
			execute:        makeDup(3),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(3),
			maxStack:       maxDupStack(3),
		},
		DUP4: {
			execute:        makeDup(4),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(4),
			maxStack:       maxDupStack(4),
		},
		DUP5: {
			execute:        makeDup(5),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(5),
			maxStack:       maxDupStack(5),
		},
		DUP6: {
			execute:        makeDup(6),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(6),
			maxStack:       maxDupStack(6),
		},
		DUP7: {
			execute:        makeDup(7),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(7),
			maxStack:       maxDupStack(7),
		},
		DUP8: {
			execute:        makeDup(8),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(8),
			maxStack:       maxDupStack(8),
		},
		DUP9: {
			execute:        makeDup(9),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(9),
			maxStack:       maxDupStack(9),
		},
		DUP10: {
			execute:        makeDup(10),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(10),
			maxStack:       maxDupStack(10),
		},
		DUP11: {
			execute:        makeDup(11),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(11),
			maxStack:       maxDupStack(11),
		},
		DUP12: {
			execute:        makeDup(12),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(12),
			maxStack:       maxDupStack(12),
		},
		DUP13: {
			execute:        makeDup(13),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(13),
			maxStack:       maxDupStack(13),
		},
		DUP14: {
			execute:        makeDup(14),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(14),
			maxStack:       maxDupStack(14),
		},
		DUP15: {
			execute:        makeDup(15),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(15),
			maxStack:       maxDupStack(15),
		},
		DUP16: {
			execute:        makeDup(16),
			constantEnergy: EnergyFastestStep,
			minStack:       minDupStack(16),
			maxStack:       maxDupStack(16),
		},
		SWAP1: {
			execute:        makeSwap(1),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(2),
			maxStack:       maxSwapStack(2),
		},
		SWAP2: {
			execute:        makeSwap(2),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(3),
			maxStack:       maxSwapStack(3),
		},
		SWAP3: {
			execute:        makeSwap(3),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(4),
			maxStack:       maxSwapStack(4),
		},
		SWAP4: {
			execute:        makeSwap(4),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(5),
			maxStack:       maxSwapStack(5),
		},
		SWAP5: {
			execute:        makeSwap(5),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(6),
			maxStack:       maxSwapStack(6),
		},
		SWAP6: {
			execute:        makeSwap(6),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(7),
			maxStack:       maxSwapStack(7),
		},
		SWAP7: {
			execute:        makeSwap(7),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(8),
			maxStack:       maxSwapStack(8),
		},
		SWAP8: {
			execute:        makeSwap(8),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(9),
			maxStack:       maxSwapStack(9),
		},
		SWAP9: {
			execute:        makeSwap(9),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(10),
			maxStack:       maxSwapStack(10),
		},
		SWAP10: {
			execute:        makeSwap(10),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(11),
			maxStack:       maxSwapStack(11),
		},
		SWAP11: {
			execute:        makeSwap(11),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(12),
			maxStack:       maxSwapStack(12),
		},
		SWAP12: {
			execute:        makeSwap(12),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(13),
			maxStack:       maxSwapStack(13),
		},
		SWAP13: {
			execute:        makeSwap(13),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(14),
			maxStack:       maxSwapStack(14),
		},
		SWAP14: {
			execute:        makeSwap(14),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(15),
			maxStack:       maxSwapStack(15),
		},
		SWAP15: {
			execute:        makeSwap(15),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(16),
			maxStack:       maxSwapStack(16),
		},
		SWAP16: {
			execute:        makeSwap(16),
			constantEnergy: EnergyFastestStep,
			minStack:       minSwapStack(17),
			maxStack:       maxSwapStack(17),
		},
		LOG0: {
			execute:       makeLog(0),
			dynamicEnergy: makeEnergyLog(0),
			minStack:      minStack(2, 0),
			maxStack:      maxStack(2, 0),
			memorySize:    memoryLog,
			writes:        true,
		},
		LOG1: {
			execute:       makeLog(1),
			dynamicEnergy: makeEnergyLog(1),
			minStack:      minStack(3, 0),
			maxStack:      maxStack(3, 0),
			memorySize:    memoryLog,
			writes:        true,
		},
		LOG2: {
			execute:       makeLog(2),
			dynamicEnergy: makeEnergyLog(2),
			minStack:      minStack(4, 0),
			maxStack:      maxStack(4, 0),
			memorySize:    memoryLog,
			writes:        true,
		},
		LOG3: {
			execute:       makeLog(3),
			dynamicEnergy: makeEnergyLog(3),
			minStack:      minStack(5, 0),
			maxStack:      maxStack(5, 0),
			memorySize:    memoryLog,
			writes:        true,
		},
		LOG4: {
			execute:       makeLog(4),
			dynamicEnergy: makeEnergyLog(4),
			minStack:      minStack(6, 0),
			maxStack:      maxStack(6, 0),
			memorySize:    memoryLog,
			writes:        true,
		},
		CREATE: {
			execute:        opCreate,
			constantEnergy: params.CreateEnergy,
			dynamicEnergy:  energyCreate,
			minStack:       minStack(3, 1),
			maxStack:       maxStack(3, 1),
			memorySize:     memoryCreate,
			writes:         true,
			returns:        true,
		},
		CALL: {
			execute:        opCall,
			constantEnergy: params.CallEnergy,
			dynamicEnergy:  energyCall,
			minStack:       minStack(7, 1),
			maxStack:       maxStack(7, 1),
			memorySize:     memoryCall,
			returns:        true,
		},
		CALLCODE: {
			execute:        opCallCode,
			constantEnergy: params.CallEnergy,
			dynamicEnergy:  energyCallCode,
			minStack:       minStack(7, 1),
			maxStack:       maxStack(7, 1),
			memorySize:     memoryCall,
			returns:        true,
		},
		RETURN: {
			execute:       opReturn,
			dynamicEnergy: energyReturn,
			minStack:      minStack(2, 0),
			maxStack:      maxStack(2, 0),
			memorySize:    memoryReturn,
			halts:         true,
		},
		SELFDESTRUCT: {
			execute:       opSuicide,
			dynamicEnergy: energySelfdestruct,
			minStack:      minStack(1, 0),
			maxStack:      maxStack(1, 0),
			halts:         true,
			writes:        true,
		},
		DELEGATECALL: {
			execute:        opDelegateCall,
			dynamicEnergy:  energyDelegateCall,
			constantEnergy: params.CallEnergy,
			minStack:       minStack(6, 1),
			maxStack:       maxStack(6, 1),
			memorySize:     memoryDelegateCall,
			returns:        true,
		},
		STATICCALL: {
			execute:        opStaticCall,
			constantEnergy: params.CallEnergy,
			dynamicEnergy:  energyStaticCall,
			minStack:       minStack(6, 1),
			maxStack:       maxStack(6, 1),
			memorySize:     memoryStaticCall,
			returns:        true,
		},
		RETURNDATASIZE: {
			execute:        opReturnDataSize,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		RETURNDATACOPY: {
			execute:        opReturnDataCopy,
			constantEnergy: EnergyFastestStep,
			dynamicEnergy:  energyReturnDataCopy,
			minStack:       minStack(3, 0),
			maxStack:       maxStack(3, 0),
			memorySize:     memoryReturnDataCopy,
		},
		REVERT: {
			execute:       opRevert,
			dynamicEnergy: energyRevert,
			minStack:      minStack(2, 0),
			maxStack:      maxStack(2, 0),
			memorySize:    memoryRevert,
			reverts:       true,
			returns:       true,
		},
		SHL: {
			execute:        opSHL,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		SHR: {
			execute:        opSHR,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		SAR: {
			execute:        opSAR,
			constantEnergy: EnergyFastestStep,
			minStack:       minStack(2, 1),
			maxStack:       maxStack(2, 1),
		},
		EXTCODEHASH: {
			execute:        opExtCodeHash,
			constantEnergy: params.ExtcodeHashEnergy,
			minStack:       minStack(1, 1),
			maxStack:       maxStack(1, 1),
		},
		CREATE2: {
			execute:        opCreate2,
			constantEnergy: params.Create2Energy,
			dynamicEnergy:  energyCreate2,
			minStack:       minStack(4, 1),
			maxStack:       maxStack(4, 1),
			memorySize:     memoryCreate2,
			writes:         true,
			returns:        true,
		},
		NETWORKID: {
			execute:        opNetworkID,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		SELFBALANCE: {
			execute:        opSelfBalance,
			constantEnergy: EnergyFastStep,
			minStack:       minStack(0, 1),
			maxStack:       maxStack(0, 1),
		},
		BEGINSUB: {
			execute:        opBeginSub,
			constantEnergy: EnergyQuickStep,
			minStack:       minStack(0, 0),
			maxStack:       maxStack(0, 0),
		},
		JUMPSUB: {
			execute:        opJumpSub,
			constantEnergy: EnergySlowStep,
			minStack:       minStack(1, 0),
			maxStack:       maxStack(1, 0),
			jumps:          true,
		},
		RETURNSUB: {
			execute:        opReturnSub,
			constantEnergy: EnergyFastStep,
			minStack:       minStack(0, 0),
			maxStack:       maxStack(0, 0),
			jumps:          true,
		},
	}
}
