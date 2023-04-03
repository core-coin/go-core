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

package core

import (
	"fmt"
	"math/big"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/common/math"
)

// GnosisSafeTx is a type to parse the safe-tx returned by the relayer,
// it also conforms to the API required by the Gnosis Safe tx relay service.
// See 'SafeMultisigTransaction' on https://safe-transaction.mainnet.gnosis.io/
type GnosisSafeTx struct {
	// These fields are only used on output
	Signature  hexutil.Bytes  `json:"signature"`
	SafeTxHash common.Hash    `json:"contractTransactionHash"`
	Sender     common.Address `json:"sender"`
	// These fields are used both on input and output
	Safe           common.Address  `json:"safe"`
	To             common.Address  `json:"to"`
	Value          math.Decimal256 `json:"value"`
	EnergyPrice    math.Decimal256 `json:"energyPrice"`
	Data           *hexutil.Bytes  `json:"data"`
	Operation      uint8           `json:"operation"`
	EnergyToken    common.Address  `json:"energyToken"`
	RefundReceiver common.Address  `json:"refundReceiver"`
	BaseEnergy     big.Int         `json:"baseEnergy"`
	SafeTxEnergy   big.Int         `json:"safeTxEnergy"`
	Nonce          big.Int         `json:"nonce"`
	InputExpHash   common.Hash     `json:"safeTxHash"`
}

// ToTypedData converts the tx to a CIP-712 Typed Data structure for signing
func (tx *GnosisSafeTx) ToTypedData() TypedData {
	var data hexutil.Bytes
	if tx.Data != nil {
		data = *tx.Data
	}
	gnosisTypedData := TypedData{
		Types: Types{
			"CIP712Domain": []Type{{Name: "verifyingContract", Type: "address"}},
			"SafeTx": []Type{
				{Name: "to", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "data", Type: "bytes"},
				{Name: "operation", Type: "uint8"},
				{Name: "safeTxEnergy", Type: "uint256"},
				{Name: "baseEnergy", Type: "uint256"},
				{Name: "energyPrice", Type: "uint256"},
				{Name: "energyToken", Type: "address"},
				{Name: "refundReceiver", Type: "address"},
				{Name: "nonce", Type: "uint256"},
			},
		},
		Domain: TypedDataDomain{
			VerifyingContract: tx.Safe.Hex(),
		},
		PrimaryType: "SafeTx",
		Message: TypedDataMessage{
			"to":             tx.To.Hex(),
			"value":          tx.Value.String(),
			"data":           data,
			"operation":      fmt.Sprintf("%d", tx.Operation),
			"safeTxEnergy":   fmt.Sprintf("%#d", &tx.SafeTxEnergy),
			"baseEnergy":     fmt.Sprintf("%#d", &tx.BaseEnergy),
			"energyPrice":    tx.EnergyPrice.String(),
			"energyToken":    tx.EnergyToken.Hex(),
			"refundReceiver": tx.RefundReceiver.Hex(),
			"nonce":          fmt.Sprintf("%d", tx.Nonce.Uint64()),
		},
	}
	return gnosisTypedData
}

// ArgsForValidation returns a SendTxArgs struct, which can be used for the
// common validations, e.g. look up 4byte destinations
func (tx *GnosisSafeTx) ArgsForValidation() *SendTxArgs {
	args := &SendTxArgs{
		From:        tx.Safe,
		To:          &tx.To,
		Energy:      hexutil.Uint64(tx.SafeTxEnergy.Uint64()),
		EnergyPrice: hexutil.Big(tx.EnergyPrice),
		Value:       hexutil.Big(tx.Value),
		Nonce:       hexutil.Uint64(tx.Nonce.Uint64()),
		Data:        tx.Data,
		Input:       nil,
	}
	return args
}
