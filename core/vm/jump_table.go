// Copyright 2015 The go-core Authors
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

	"github.com/core-coin/go-core/params"
)

type (
	executionFunc func(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error)
	energyFunc       func(*CVM, *Contract, *Stack, *Memory, uint64) (uint64, error) // last parameter is the requested memory size as a uint64
	// memorySizeFunc returns the required size, and whether the operation overflowed a uint64
	memorySizeFunc func(*Stack) (size uint64, overflow bool)
)

var errEnergyUintOverflow = errors.New("energy uint64 overflow")

type operation struct {
	// execute is the operation function
	execute     executionFunc
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
	valid   bool // indication whether the retrieved operation is valid and known
	reverts bool // determines whether the operation reverts state (implicitly halts)
	returns bool // determines whether the operations sets the return data content
}

var (
	frontierInstructionSet         = newFrontierInstructionSet()
	homesteadInstructionSet        = newHomesteadInstructionSet()
	tangerineWhistleInstructionSet = newTangerineWhistleInstructionSet()
	spuriousDragonInstructionSet   = newSpuriousDragonInstructionSet()
	byzantiumInstructionSet        = newByzantiumInstructionSet()
	constantinopleInstructionSet   = newConstantinopleInstructionSet()
	istanbulInstructionSet         = newIstanbulInstructionSet()
)

// JumpTable contains the CVM opcodes supported at a given fork.
type JumpTable [256]operation

// newIstanbulInstructionSet returns the frontier, homestead
// byzantium, contantinople and petersburg instructions.
func newIstanbulInstructionSet() JumpTable {
	instructionSet := newConstantinopleInstructionSet()

	enable1344(&instructionSet) // ChainID opcode - https://eips.coreblockchain.cc/EIPS/eip-1344
	enable1884(&instructionSet) // Reprice reader opcodes - https://eips.coreblockchain.cc/EIPS/eip-1884
	enable2200(&instructionSet) // Net metered SSTORE - https://eips.coreblockchain.cc/EIPS/eip-2200

	return instructionSet
}

// newConstantinopleInstructionSet returns the frontier, homestead
// byzantium and contantinople instructions.
func newConstantinopleInstructionSet() JumpTable {
	instructionSet := newByzantiumInstructionSet()
	instructionSet[SHL] = operation{
		execute:     opSHL,
		constantEnergy: EnergyFastestStep,
		minStack:    minStack(2, 1),
		maxStack:    maxStack(2, 1),
		valid:       true,
	}
	instructionSet[SHR] = operation{
		execute:     opSHR,
		constantEnergy: EnergyFastestStep,
		minStack:    minStack(2, 1),
		maxStack:    maxStack(2, 1),
		valid:       true,
	}
	instructionSet[SAR] = operation{
		execute:     opSAR,
		constantEnergy: EnergyFastestStep,
		minStack:    minStack(2, 1),
		maxStack:    maxStack(2, 1),
		valid:       true,
	}
	instructionSet[EXTCODEHASH] = operation{
		execute:     opExtCodeHash,
		constantEnergy: params.ExtcodeHashEnergyConstantinople,
		minStack:    minStack(1, 1),
		maxStack:    maxStack(1, 1),
		valid:       true,
	}
	instructionSet[CREATE2] = operation{
		execute:     opCreate2,
		constantEnergy: params.Create2Energy,
		dynamicEnergy:  energyCreate2,
		minStack:    minStack(4, 1),
		maxStack:    maxStack(4, 1),
		memorySize:  memoryCreate2,
		valid:       true,
		writes:      true,
		returns:     true,
	}
	return instructionSet
}

// newByzantiumInstructionSet returns the frontier, homestead and
// byzantium instructions.
func newByzantiumInstructionSet() JumpTable {
	instructionSet := newSpuriousDragonInstructionSet()
	instructionSet[STATICCALL] = operation{
		execute:     opStaticCall,
		constantEnergy: params.CallEnergyEIP150,
		dynamicEnergy:  energyStaticCall,
		minStack:    minStack(6, 1),
		maxStack:    maxStack(6, 1),
		memorySize:  memoryStaticCall,
		valid:       true,
		returns:     true,
	}
	instructionSet[RETURNDATASIZE] = operation{
		execute:     opReturnDataSize,
		constantEnergy: EnergyQuickStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
		valid:       true,
	}
	instructionSet[RETURNDATACOPY] = operation{
		execute:     opReturnDataCopy,
		constantEnergy: EnergyFastestStep,
		dynamicEnergy:  energyReturnDataCopy,
		minStack:    minStack(3, 0),
		maxStack:    maxStack(3, 0),
		memorySize:  memoryReturnDataCopy,
		valid:       true,
	}
	instructionSet[REVERT] = operation{
		execute:    opRevert,
		dynamicEnergy: energyRevert,
		minStack:   minStack(2, 0),
		maxStack:   maxStack(2, 0),
		memorySize: memoryRevert,
		valid:      true,
		reverts:    true,
		returns:    true,
	}
	return instructionSet
}

// EIP 158 a.k.a Spurious Dragon
func newSpuriousDragonInstructionSet() JumpTable {
	instructionSet := newTangerineWhistleInstructionSet()
	instructionSet[EXP].dynamicEnergy = energyExpEIP158
	return instructionSet

}

// EIP 150 a.k.a Tangerine Whistle
func newTangerineWhistleInstructionSet() JumpTable {
	instructionSet := newHomesteadInstructionSet()
	instructionSet[BALANCE].constantEnergy = params.BalanceEnergyEIP150
	instructionSet[EXTCODESIZE].constantEnergy = params.ExtcodeSizeEnergyEIP150
	instructionSet[SLOAD].constantEnergy = params.SloadEnergyEIP150
	instructionSet[EXTCODECOPY].constantEnergy = params.ExtcodeCopyBaseEIP150
	instructionSet[CALL].constantEnergy = params.CallEnergyEIP150
	instructionSet[CALLCODE].constantEnergy = params.CallEnergyEIP150
	instructionSet[DELEGATECALL].constantEnergy = params.CallEnergyEIP150
	return instructionSet
}

// newHomesteadInstructionSet returns the frontier and homestead
// instructions that can be executed during the homestead phase.
func newHomesteadInstructionSet() JumpTable {
	instructionSet := newFrontierInstructionSet()
	instructionSet[DELEGATECALL] = operation{
		execute:     opDelegateCall,
		dynamicEnergy:  energyDelegateCall,
		constantEnergy: params.CallEnergyFrontier,
		minStack:    minStack(6, 1),
		maxStack:    maxStack(6, 1),
		memorySize:  memoryDelegateCall,
		valid:       true,
		returns:     true,
	}
	return instructionSet
}

// newFrontierInstructionSet returns the frontier instructions
// that can be executed during the frontier phase.
func newFrontierInstructionSet() JumpTable {
	return JumpTable{
		STOP: {
			execute:     opStop,
			constantEnergy: 0,
			minStack:    minStack(0, 0),
			maxStack:    maxStack(0, 0),
			halts:       true,
			valid:       true,
		},
		ADD: {
			execute:     opAdd,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		MUL: {
			execute:     opMul,
			constantEnergy: EnergyFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		SUB: {
			execute:     opSub,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		DIV: {
			execute:     opDiv,
			constantEnergy: EnergyFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		SDIV: {
			execute:     opSdiv,
			constantEnergy: EnergyFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		MOD: {
			execute:     opMod,
			constantEnergy: EnergyFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		SMOD: {
			execute:     opSmod,
			constantEnergy: EnergyFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		ADDMOD: {
			execute:     opAddmod,
			constantEnergy: EnergyMidStep,
			minStack:    minStack(3, 1),
			maxStack:    maxStack(3, 1),
			valid:       true,
		},
		MULMOD: {
			execute:     opMulmod,
			constantEnergy: EnergyMidStep,
			minStack:    minStack(3, 1),
			maxStack:    maxStack(3, 1),
			valid:       true,
		},
		EXP: {
			execute:    opExp,
			dynamicEnergy: energyExpFrontier,
			minStack:   minStack(2, 1),
			maxStack:   maxStack(2, 1),
			valid:      true,
		},
		SIGNEXTEND: {
			execute:     opSignExtend,
			constantEnergy: EnergyFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		LT: {
			execute:     opLt,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		GT: {
			execute:     opGt,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		SLT: {
			execute:     opSlt,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		SGT: {
			execute:     opSgt,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		EQ: {
			execute:     opEq,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		ISZERO: {
			execute:     opIszero,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
			valid:       true,
		},
		AND: {
			execute:     opAnd,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		XOR: {
			execute:     opXor,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		OR: {
			execute:     opOr,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		NOT: {
			execute:     opNot,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
			valid:       true,
		},
		BYTE: {
			execute:     opByte,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			valid:       true,
		},
		SHA3: {
			execute:     opSha3,
			constantEnergy: params.Sha3Energy,
			dynamicEnergy:  energySha3,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			memorySize:  memorySha3,
			valid:       true,
		},
		ADDRESS: {
			execute:     opAddress,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		BALANCE: {
			execute:     opBalance,
			constantEnergy: params.BalanceEnergyFrontier,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
			valid:       true,
		},
		ORIGIN: {
			execute:     opOrigin,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		CALLER: {
			execute:     opCaller,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		CALLVALUE: {
			execute:     opCallValue,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		CALLDATALOAD: {
			execute:     opCallDataLoad,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
			valid:       true,
		},
		CALLDATASIZE: {
			execute:     opCallDataSize,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		CALLDATACOPY: {
			execute:     opCallDataCopy,
			constantEnergy: EnergyFastestStep,
			dynamicEnergy:  energyCallDataCopy,
			minStack:    minStack(3, 0),
			maxStack:    maxStack(3, 0),
			memorySize:  memoryCallDataCopy,
			valid:       true,
		},
		CODESIZE: {
			execute:     opCodeSize,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		CODECOPY: {
			execute:     opCodeCopy,
			constantEnergy: EnergyFastestStep,
			dynamicEnergy:  energyCodeCopy,
			minStack:    minStack(3, 0),
			maxStack:    maxStack(3, 0),
			memorySize:  memoryCodeCopy,
			valid:       true,
		},
		ENERGYPRICE: {
			execute:     opEnergyprice,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		EXTCODESIZE: {
			execute:     opExtCodeSize,
			constantEnergy: params.ExtcodeSizeEnergyFrontier,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
			valid:       true,
		},
		EXTCODECOPY: {
			execute:     opExtCodeCopy,
			constantEnergy: params.ExtcodeCopyBaseFrontier,
			dynamicEnergy:  energyExtCodeCopy,
			minStack:    minStack(4, 0),
			maxStack:    maxStack(4, 0),
			memorySize:  memoryExtCodeCopy,
			valid:       true,
		},
		BLOCKHASH: {
			execute:     opBlockhash,
			constantEnergy: EnergyExtStep,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
			valid:       true,
		},
		COINBASE: {
			execute:     opCoinbase,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		TIMESTAMP: {
			execute:     opTimestamp,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		NUMBER: {
			execute:     opNumber,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		DIFFICULTY: {
			execute:     opDifficulty,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		ENERGYLIMIT: {
			execute:     opEnergyLimit,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		POP: {
			execute:     opPop,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(1, 0),
			maxStack:    maxStack(1, 0),
			valid:       true,
		},
		MLOAD: {
			execute:     opMload,
			constantEnergy: EnergyFastestStep,
			dynamicEnergy:  energyMLoad,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
			memorySize:  memoryMLoad,
			valid:       true,
		},
		MSTORE: {
			execute:     opMstore,
			constantEnergy: EnergyFastestStep,
			dynamicEnergy:  energyMStore,
			minStack:    minStack(2, 0),
			maxStack:    maxStack(2, 0),
			memorySize:  memoryMStore,
			valid:       true,
		},
		MSTORE8: {
			execute:     opMstore8,
			constantEnergy: EnergyFastestStep,
			dynamicEnergy:  energyMStore8,
			memorySize:  memoryMStore8,
			minStack:    minStack(2, 0),
			maxStack:    maxStack(2, 0),

			valid: true,
		},
		SLOAD: {
			execute:     opSload,
			constantEnergy: params.SloadEnergyFrontier,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
			valid:       true,
		},
		SSTORE: {
			execute:    opSstore,
			dynamicEnergy: energySStore,
			minStack:   minStack(2, 0),
			maxStack:   maxStack(2, 0),
			valid:      true,
			writes:     true,
		},
		JUMP: {
			execute:     opJump,
			constantEnergy: EnergyMidStep,
			minStack:    minStack(1, 0),
			maxStack:    maxStack(1, 0),
			jumps:       true,
			valid:       true,
		},
		JUMPI: {
			execute:     opJumpi,
			constantEnergy: EnergySlowStep,
			minStack:    minStack(2, 0),
			maxStack:    maxStack(2, 0),
			jumps:       true,
			valid:       true,
		},
		PC: {
			execute:     opPc,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		MSIZE: {
			execute:     opMsize,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		ENERGY: {
			execute:     opEnergy,
			constantEnergy: EnergyQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		JUMPDEST: {
			execute:     opJumpdest,
			constantEnergy: params.JumpdestEnergy,
			minStack:    minStack(0, 0),
			maxStack:    maxStack(0, 0),
			valid:       true,
		},
		PUSH1: {
			execute:     opPush1,
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH2: {
			execute:     makePush(2, 2),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH3: {
			execute:     makePush(3, 3),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH4: {
			execute:     makePush(4, 4),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH5: {
			execute:     makePush(5, 5),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH6: {
			execute:     makePush(6, 6),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH7: {
			execute:     makePush(7, 7),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH8: {
			execute:     makePush(8, 8),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH9: {
			execute:     makePush(9, 9),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH10: {
			execute:     makePush(10, 10),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH11: {
			execute:     makePush(11, 11),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH12: {
			execute:     makePush(12, 12),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH13: {
			execute:     makePush(13, 13),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH14: {
			execute:     makePush(14, 14),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH15: {
			execute:     makePush(15, 15),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH16: {
			execute:     makePush(16, 16),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH17: {
			execute:     makePush(17, 17),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH18: {
			execute:     makePush(18, 18),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH19: {
			execute:     makePush(19, 19),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH20: {
			execute:     makePush(20, 20),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH21: {
			execute:     makePush(21, 21),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH22: {
			execute:     makePush(22, 22),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH23: {
			execute:     makePush(23, 23),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH24: {
			execute:     makePush(24, 24),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH25: {
			execute:     makePush(25, 25),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH26: {
			execute:     makePush(26, 26),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH27: {
			execute:     makePush(27, 27),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH28: {
			execute:     makePush(28, 28),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH29: {
			execute:     makePush(29, 29),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH30: {
			execute:     makePush(30, 30),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH31: {
			execute:     makePush(31, 31),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		PUSH32: {
			execute:     makePush(32, 32),
			constantEnergy: EnergyFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
			valid:       true,
		},
		DUP1: {
			execute:     makeDup(1),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(1),
			maxStack:    maxDupStack(1),
			valid:       true,
		},
		DUP2: {
			execute:     makeDup(2),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(2),
			maxStack:    maxDupStack(2),
			valid:       true,
		},
		DUP3: {
			execute:     makeDup(3),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(3),
			maxStack:    maxDupStack(3),
			valid:       true,
		},
		DUP4: {
			execute:     makeDup(4),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(4),
			maxStack:    maxDupStack(4),
			valid:       true,
		},
		DUP5: {
			execute:     makeDup(5),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(5),
			maxStack:    maxDupStack(5),
			valid:       true,
		},
		DUP6: {
			execute:     makeDup(6),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(6),
			maxStack:    maxDupStack(6),
			valid:       true,
		},
		DUP7: {
			execute:     makeDup(7),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(7),
			maxStack:    maxDupStack(7),
			valid:       true,
		},
		DUP8: {
			execute:     makeDup(8),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(8),
			maxStack:    maxDupStack(8),
			valid:       true,
		},
		DUP9: {
			execute:     makeDup(9),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(9),
			maxStack:    maxDupStack(9),
			valid:       true,
		},
		DUP10: {
			execute:     makeDup(10),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(10),
			maxStack:    maxDupStack(10),
			valid:       true,
		},
		DUP11: {
			execute:     makeDup(11),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(11),
			maxStack:    maxDupStack(11),
			valid:       true,
		},
		DUP12: {
			execute:     makeDup(12),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(12),
			maxStack:    maxDupStack(12),
			valid:       true,
		},
		DUP13: {
			execute:     makeDup(13),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(13),
			maxStack:    maxDupStack(13),
			valid:       true,
		},
		DUP14: {
			execute:     makeDup(14),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(14),
			maxStack:    maxDupStack(14),
			valid:       true,
		},
		DUP15: {
			execute:     makeDup(15),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(15),
			maxStack:    maxDupStack(15),
			valid:       true,
		},
		DUP16: {
			execute:     makeDup(16),
			constantEnergy: EnergyFastestStep,
			minStack:    minDupStack(16),
			maxStack:    maxDupStack(16),
			valid:       true,
		},
		SWAP1: {
			execute:     makeSwap(1),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(2),
			maxStack:    maxSwapStack(2),
			valid:       true,
		},
		SWAP2: {
			execute:     makeSwap(2),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(3),
			maxStack:    maxSwapStack(3),
			valid:       true,
		},
		SWAP3: {
			execute:     makeSwap(3),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(4),
			maxStack:    maxSwapStack(4),
			valid:       true,
		},
		SWAP4: {
			execute:     makeSwap(4),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(5),
			maxStack:    maxSwapStack(5),
			valid:       true,
		},
		SWAP5: {
			execute:     makeSwap(5),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(6),
			maxStack:    maxSwapStack(6),
			valid:       true,
		},
		SWAP6: {
			execute:     makeSwap(6),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(7),
			maxStack:    maxSwapStack(7),
			valid:       true,
		},
		SWAP7: {
			execute:     makeSwap(7),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(8),
			maxStack:    maxSwapStack(8),
			valid:       true,
		},
		SWAP8: {
			execute:     makeSwap(8),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(9),
			maxStack:    maxSwapStack(9),
			valid:       true,
		},
		SWAP9: {
			execute:     makeSwap(9),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(10),
			maxStack:    maxSwapStack(10),
			valid:       true,
		},
		SWAP10: {
			execute:     makeSwap(10),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(11),
			maxStack:    maxSwapStack(11),
			valid:       true,
		},
		SWAP11: {
			execute:     makeSwap(11),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(12),
			maxStack:    maxSwapStack(12),
			valid:       true,
		},
		SWAP12: {
			execute:     makeSwap(12),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(13),
			maxStack:    maxSwapStack(13),
			valid:       true,
		},
		SWAP13: {
			execute:     makeSwap(13),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(14),
			maxStack:    maxSwapStack(14),
			valid:       true,
		},
		SWAP14: {
			execute:     makeSwap(14),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(15),
			maxStack:    maxSwapStack(15),
			valid:       true,
		},
		SWAP15: {
			execute:     makeSwap(15),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(16),
			maxStack:    maxSwapStack(16),
			valid:       true,
		},
		SWAP16: {
			execute:     makeSwap(16),
			constantEnergy: EnergyFastestStep,
			minStack:    minSwapStack(17),
			maxStack:    maxSwapStack(17),
			valid:       true,
		},
		LOG0: {
			execute:    makeLog(0),
			dynamicEnergy: makeEnergyLog(0),
			minStack:   minStack(2, 0),
			maxStack:   maxStack(2, 0),
			memorySize: memoryLog,
			valid:      true,
			writes:     true,
		},
		LOG1: {
			execute:    makeLog(1),
			dynamicEnergy: makeEnergyLog(1),
			minStack:   minStack(3, 0),
			maxStack:   maxStack(3, 0),
			memorySize: memoryLog,
			valid:      true,
			writes:     true,
		},
		LOG2: {
			execute:    makeLog(2),
			dynamicEnergy: makeEnergyLog(2),
			minStack:   minStack(4, 0),
			maxStack:   maxStack(4, 0),
			memorySize: memoryLog,
			valid:      true,
			writes:     true,
		},
		LOG3: {
			execute:    makeLog(3),
			dynamicEnergy: makeEnergyLog(3),
			minStack:   minStack(5, 0),
			maxStack:   maxStack(5, 0),
			memorySize: memoryLog,
			valid:      true,
			writes:     true,
		},
		LOG4: {
			execute:    makeLog(4),
			dynamicEnergy: makeEnergyLog(4),
			minStack:   minStack(6, 0),
			maxStack:   maxStack(6, 0),
			memorySize: memoryLog,
			valid:      true,
			writes:     true,
		},
		CREATE: {
			execute:     opCreate,
			constantEnergy: params.CreateEnergy,
			dynamicEnergy:  energyCreate,
			minStack:    minStack(3, 1),
			maxStack:    maxStack(3, 1),
			memorySize:  memoryCreate,
			valid:       true,
			writes:      true,
			returns:     true,
		},
		CALL: {
			execute:     opCall,
			constantEnergy: params.CallEnergyFrontier,
			dynamicEnergy:  energyCall,
			minStack:    minStack(7, 1),
			maxStack:    maxStack(7, 1),
			memorySize:  memoryCall,
			valid:       true,
			returns:     true,
		},
		CALLCODE: {
			execute:     opCallCode,
			constantEnergy: params.CallEnergyFrontier,
			dynamicEnergy:  energyCallCode,
			minStack:    minStack(7, 1),
			maxStack:    maxStack(7, 1),
			memorySize:  memoryCall,
			valid:       true,
			returns:     true,
		},
		RETURN: {
			execute:    opReturn,
			dynamicEnergy: energyReturn,
			minStack:   minStack(2, 0),
			maxStack:   maxStack(2, 0),
			memorySize: memoryReturn,
			halts:      true,
			valid:      true,
		},
		SELFDESTRUCT: {
			execute:    opSuicide,
			dynamicEnergy: energySelfdestruct,
			minStack:   minStack(1, 0),
			maxStack:   maxStack(1, 0),
			halts:      true,
			valid:      true,
			writes:     true,
		},
	}
}
