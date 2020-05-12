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

package params

import "math/big"

const (
	EnergyLimitBoundDivisor uint64 = 1024    // The bound divisor of the energy limit, used in update calculations.
	MinEnergyLimit          uint64 = 5000    // Minimum the energy limit may ever be.
	GenesisEnergyLimit      uint64 = 4712388 // Energy limit of the Genesis block.

	MaximumExtraDataSize  uint64 = 32    // Maximum size extra data may be after Genesis.
	ExpByteEnergy            uint64 = 10    // Times ceil(log256(exponent)) for the EXP instruction.
	SloadEnergy              uint64 = 50    // Multiplied by the number of 32-byte words that are copied (round up) for any *COPY operation and added.
	CallValueTransferEnergy  uint64 = 9000  // Paid for CALL when the value transfer is non-zero.
	CallNewAccountEnergy     uint64 = 25000 // Paid for CALL when the destination address didn't exist prior.
	TxEnergy                 uint64 = 21000 // Per transaction not creating a contract. NOTE: Not payable on data of calls between transactions.
	TxEnergyContractCreation uint64 = 53000 // Per transaction that creates a contract. NOTE: Not payable on data of calls between transactions.
	TxDataZeroEnergy         uint64 = 4     // Per byte of data attached to a transaction that equals zero. NOTE: Not payable on data of calls between transactions.
	QuadCoeffDiv          uint64 = 512   // Divisor for the quadratic particle of the memory cost equation.
	LogDataEnergy            uint64 = 8     // Per byte in a LOG* operation's data.
	CallStipend           uint64 = 2300  // Free energy given at beginning of call.

	Sha3Energy     uint64 = 30 // Once per SHA3 operation.
	Sha3WordEnergy uint64 = 6  // Once per word of the SHA3 operation's data.

	SstoreSetEnergy    uint64 = 20000 // Once per SLOAD operation.
	SstoreResetEnergy  uint64 = 5000  // Once per SSTORE operation if the zeroness changes from zero.
	SstoreClearEnergy  uint64 = 5000  // Once per SSTORE operation if the zeroness doesn't change.
	SstoreRefundEnergy uint64 = 15000 // Once per SSTORE operation if the zeroness changes to zero.

	NetSstoreNoopEnergy  uint64 = 200   // Once per SSTORE operation if the value doesn't change.
	NetSstoreInitEnergy  uint64 = 20000 // Once per SSTORE operation from clean zero.
	NetSstoreCleanEnergy uint64 = 5000  // Once per SSTORE operation from clean non-zero.
	NetSstoreDirtyEnergy uint64 = 200   // Once per SSTORE operation from dirty.

	NetSstoreClearRefund      uint64 = 15000 // Once per SSTORE operation for clearing an originally existing storage slot
	NetSstoreResetRefund      uint64 = 4800  // Once per SSTORE operation for resetting to the original non-zero value
	NetSstoreResetClearRefund uint64 = 19800 // Once per SSTORE operation for resetting to the original zero value

	SstoreSentryEnergyCIP2200   uint64 = 2300  // Minimum energy required to be present for an SSTORE call, not consumed
	SstoreNoopEnergyCIP2200     uint64 = 800   // Once per SSTORE operation if the value doesn't change.
	SstoreDirtyEnergyCIP2200    uint64 = 800   // Once per SSTORE operation if a dirty value is changed.
	SstoreInitEnergyCIP2200     uint64 = 20000 // Once per SSTORE operation from clean zero to non-zero
	SstoreInitRefundCIP2200  uint64 = 19200 // Once per SSTORE operation for resetting to the original zero value
	SstoreCleanEnergyCIP2200    uint64 = 5000  // Once per SSTORE operation from clean non-zero to something else
	SstoreCleanRefundCIP2200 uint64 = 4200  // Once per SSTORE operation for resetting to the original non-zero value
	SstoreClearRefundCIP2200 uint64 = 15000 // Once per SSTORE operation for clearing an originally existing storage slot

	JumpdestEnergy   uint64 = 1     // Once per JUMPDEST operation.
	EpochDuration uint64 = 30000 // Duration between proof-of-work epochs.

	CreateDataEnergy            uint64 = 200   //
	CallCreateDepth          uint64 = 1024  // Maximum depth of call/create stack.
	ExpEnergy                   uint64 = 10    // Once per EXP instruction
	LogEnergy                   uint64 = 375   // Per LOG* operation.
	CopyEnergy                  uint64 = 3     //
	StackLimit               uint64 = 1024  // Maximum size of VM stack allowed.
	TierStepEnergy              uint64 = 0     // Once per operation, for a selection of them.
	LogTopicEnergy              uint64 = 375   // Multiplied by the * of the LOG*, per LOG transaction. e.g. LOG0 incurs 0 * c_txLogTopicEnergy, LOG4 incurs 4 * c_txLogTopicEnergy.
	CreateEnergy                uint64 = 32000 // Once per CREATE operation & contract-creation transaction.
	Create2Energy               uint64 = 32000 // Once per CREATE2 operation
	SelfdestructRefundEnergy    uint64 = 24000 // Refunded following a selfdestruct operation.
	MemoryEnergy                uint64 = 3     // Times the address of the (highest referenced byte in memory + 1). NOTE: referencing happens on read, write and in instructions such as RETURN and CALL.
	TxDataNonZeroEnergyFrontier uint64 = 68    // Per byte of data attached to a transaction that is not equal to zero. NOTE: Not payable on data of calls between transactions.
	TxDataNonZeroEnergyCIP2028  uint64 = 16    // Per byte of non zero data attached to a transaction after CIP 2028 (part in Istanbul)

	// These have been changed during the course of the chain
	CallEnergyFrontier              uint64 = 40  // Once per CALL operation & message call transaction.
	CallEnergyCIP150                uint64 = 700 // Static portion of energy for CALL-derivates after CIP 150 (Tangerine)
	BalanceEnergyFrontier           uint64 = 20  // The cost of a BALANCE operation
	BalanceEnergyCIP150             uint64 = 400 // The cost of a BALANCE operation after Tangerine
	BalanceEnergyCIP1884            uint64 = 700 // The cost of a BALANCE operation after CIP 1884 (part of Istanbul)
	ExtcodeSizeEnergyFrontier       uint64 = 20  // Cost of EXTCODESIZE before CIP 150 (Tangerine)
	ExtcodeSizeEnergyCIP150         uint64 = 700 // Cost of EXTCODESIZE after CIP 150 (Tangerine)
	SloadEnergyFrontier             uint64 = 50
	SloadEnergyCIP150               uint64 = 200
	SloadEnergyCIP1884              uint64 = 800  // Cost of SLOAD after CIP 1884 (part of Istanbul)
	SloadEnergyCIP2200              uint64 = 800  // Cost of SLOAD after CIP 2200 (part of Istanbul)
	ExtcodeHashEnergyConstantinople uint64 = 400  // Cost of EXTCODEHASH (introduced in Constantinople)
	ExtcodeHashEnergyCIP1884        uint64 = 700  // Cost of EXTCODEHASH after CIP 1884 (part in Istanbul)
	SelfdestructEnergyCIP150        uint64 = 5000 // Cost of SELFDESTRUCT post CIP 150 (Tangerine)

	// EXP has a dynamic portion depending on the size of the exponent
	ExpByteFrontier uint64 = 10 // was set to 10 in Frontier
	ExpByteCIP158   uint64 = 50 // was raised to 50 during Cip158 (Spurious Dragon)

	// Extcodecopy has a dynamic AND a static cost. This represents only the
	// static portion of the energy. It was changed during CIP 150 (Tangerine)
	ExtcodeCopyBaseFrontier uint64 = 20
	ExtcodeCopyBaseCIP150   uint64 = 700

	// CreateBySelfdestructEnergy is used when the refunded account is one that does
	// not exist. This logic is similar to call.
	// Introduced in Tangerine Whistle (Cip 150)
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
	ModExpQuadCoeffDiv  uint64 = 20   // Divisor for the quadratic particle of the big int modular exponentiation

	Bn256AddEnergyByzantium             uint64 = 500    // Byzantium energy needed for an elliptic curve addition
	Bn256AddEnergyIstanbul              uint64 = 150    // Energy needed for an elliptic curve addition
	Bn256ScalarMulEnergyByzantium       uint64 = 40000  // Byzantium energy needed for an elliptic curve scalar multiplication
	Bn256ScalarMulEnergyIstanbul        uint64 = 6000   // Energy needed for an elliptic curve scalar multiplication
	Bn256PairingBaseEnergyByzantium     uint64 = 100000 // Byzantium base price for an elliptic curve pairing check
	Bn256PairingBaseEnergyIstanbul      uint64 = 45000  // Base price for an elliptic curve pairing check
	Bn256PairingPerPointEnergyByzantium uint64 = 80000  // Byzantium per-point price for an elliptic curve pairing check
	Bn256PairingPerPointEnergyIstanbul  uint64 = 34000  // Per-point price for an elliptic curve pairing check
)

var (
	DifficultyBoundDivisor = big.NewInt(2)   // The bound divisor of the difficulty, used in the update calculations.
	GenesisDifficulty      = big.NewInt(32) // Difficulty of the Genesis block.
	MinimumDifficulty      = big.NewInt(0x20) // The minimum that the difficulty may ever be.
	DurationLimit          = big.NewInt(6)     // The decision boundary on the blocktime duration used to determine whether difficulty should go up or not.
)
