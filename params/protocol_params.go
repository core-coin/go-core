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

package params

import "math/big"

const (
	EnergyLimitBoundDivisor uint64 = 1024    // The bound divisor of the energy limit, used in update calculations.
	MinEnergyLimit          uint64 = 5000    // Minimum the energy limit may ever be.
	GenesisEnergyLimit      uint64 = 4712388 // Energy limit of the Genesis block.

	MaximumExtraDataSize     uint64 = 32    // Maximum size extra data may be after Genesis.
	CallValueTransferEnergy  uint64 = 9000  // Paid for CALL when the value transfer is non-zero.
	CallNewAccountEnergy     uint64 = 25000 // Paid for CALL when the destination address didn't exist prior.
	TxEnergy                 uint64 = 21000 // Per transaction not creating a contract. NOTE: Not payable on data of calls between transactions.
	TxEnergyContractCreation uint64 = 53000 // Per transaction that creates a contract. NOTE: Not payable on data of calls between transactions.
	TxDataZeroEnergy         uint64 = 4     // Per byte of data attached to a transaction that equals zero. NOTE: Not payable on data of calls between transactions.
	QuadCoeffDiv             uint64 = 512   // Divisor for the quadratic particle of the memory cost equation.
	LogDataEnergy            uint64 = 8     // Per byte in a LOG* operation's data.
	CallStipend              uint64 = 2300  // Free energy given at beginning of call.

	Sha3Energy     uint64 = 30 // Once per SHA3 operation.
	Sha3WordEnergy uint64 = 6  // Once per word of the SHA3 operation's data.

	SstoreSentryEnergy uint64 = 2300  // Minimum energy required to be present for an SSTORE call, not consumed
	SstoreNoopEnergy   uint64 = 800   // Once per SSTORE operation if the value doesn't change.
	SstoreDirtyEnergy  uint64 = 800   // Once per SSTORE operation if a dirty value is changed.
	SstoreInitEnergy   uint64 = 20000 // Once per SSTORE operation from clean zero to non-zero
	SstoreInitRefund   uint64 = 19200 // Once per SSTORE operation for resetting to the original zero value
	SstoreCleanEnergy  uint64 = 5000  // Once per SSTORE operation from clean non-zero to something else
	SstoreCleanRefund  uint64 = 4200  // Once per SSTORE operation for resetting to the original non-zero value
	SstoreClearRefund  uint64 = 15000 // Once per SSTORE operation for clearing an originally existing storage slot

	JumpdestEnergy uint64 = 1 // Once per JUMPDEST operation.

	CreateDataEnergy         uint64 = 200   //
	CallCreateDepth          uint64 = 1024  // Maximum depth of call/create stack.
	ExpEnergy                uint64 = 10    // Once per EXP instruction
	LogEnergy                uint64 = 375   // Per LOG* operation.
	CopyEnergy               uint64 = 3     //
	StackLimit               uint64 = 1024  // Maximum size of VM stack allowed.
	LogTopicEnergy           uint64 = 375   // Multiplied by the * of the LOG*, per LOG transaction. e.g. LOG0 incurs 0 * c_txLogTopicEnergy, LOG4 incurs 4 * c_txLogTopicEnergy.
	CreateEnergy             uint64 = 32000 // Once per CREATE operation & contract-creation transaction.
	Create2Energy            uint64 = 32000 // Once per CREATE2 operation
	SelfdestructRefundEnergy uint64 = 24000 // Refunded following a selfdestruct operation.
	MemoryEnergy             uint64 = 3     // Times the address of the (highest referenced byte in memory + 1). NOTE: referencing happens on read, write and in instructions such as RETURN and CALL.
	TxDataNonZeroEnergy      uint64 = 16    // Per byte of non zero data attached to a transaction

	// These have been changed during the course of the chain
	CallEnergy         uint64 = 700  // Static portion of energy for CALL-derivates
	BalanceEnergy      uint64 = 700  // The cost of a BALANCE operation
	ExtcodeSizeEnergy  uint64 = 700  // Cost of EXTCODESIZE
	SloadEnergy        uint64 = 800  // Cost of SLOAD
	ExtcodeHashEnergy  uint64 = 700  // Cost of EXTCODEHASH
	SelfdestructEnergy uint64 = 5000 // Cost of SELFDESTRUCT

	// EXP has a dynamic portion depending on the size of the exponent
	ExpByte uint64 = 50 // was raised to 50

	// Extcodecopy has a dynamic AND a static cost. This represents only the
	// static portion of the energy.
	ExtcodeCopyBase uint64 = 700

	// CreateBySelfdestructEnergy is used when the refunded account is one that does
	// not exist. This logic is similar to call.
	CreateBySelfdestructEnergy uint64 = 25000

	MaxCodeSize = 24576 // Maximum bytecode to permit for a contract

	// Precompiled contract energy prices

	EcrecoverEnergy        uint64 = 3000 // Elliptic curve sender recovery energy price
	Sha256BaseEnergy       uint64 = 60   // Base price for a SHA256 operation
	Sha256PerWordEnergy    uint64 = 12   // Per-word price for a SHA256 operation
	Ripemd160BaseEnergy    uint64 = 600  // Base price for a RIPEMD160 operation
	Ripemd160PerWordEnergy uint64 = 120  // Per-word price for a RIPEMD160 operation
	IdentityBaseEnergy     uint64 = 15   // Base price for a data copy operation
	IdentityPerWordEnergy  uint64 = 3    // Per-work price for a data copy operation
	ModExpQuadCoeffDiv     uint64 = 20   // Divisor for the quadratic particle of the big int modular exponentiation

	Bn256AddEnergy             uint64 = 150   // Energy needed for an elliptic curve addition
	Bn256ScalarMulEnergy       uint64 = 6000  // Energy needed for an elliptic curve scalar multiplication
	Bn256PairingBaseEnergy     uint64 = 45000 // Base price for an elliptic curve pairing check
	Bn256PairingPerPointEnergy uint64 = 34000 // Per-point price for an elliptic curve pairing check
)

var (
	DifficultyBoundDivisor = big.NewInt(2)    // The bound divisor of the difficulty, used in the update calculations.
	GenesisDifficulty      = big.NewInt(32)   // Difficulty of the Genesis block.
	MinimumDifficulty      = big.NewInt(0x20) // The minimum that the difficulty may ever be.
	DurationLimit          = big.NewInt(6)    // The decision boundary on the blocktime duration used to determine whether difficulty should go up or not.
)
