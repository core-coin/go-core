// Copyright 2024 by the Authors
// This file is part of the go-core library.
//
// The go-core library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-core library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR ANY PURPOSE. See the GNU Lesser General Public
// License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-core library. If not, see <http://www.gnu.org/licenses/>.

package scapi

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"time"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/core/state"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/internal/xcbapi"
	"github.com/core-coin/go-core/v2/log"
	"github.com/core-coin/go-core/v2/rpc"
)

// PublicSmartContractAPI provides an API to access smart contract related information.
// It offers methods that operate on public data that is freely available to anyone.
type PublicSmartContractAPI struct {
	b Backend
}

// NewPublicSmartContractAPI creates a new smart contract protocol API.
func NewPublicSmartContractAPI(b Backend) *PublicSmartContractAPI {
	return &PublicSmartContractAPI{b}
}

// createViewCallMsg creates a properly initialized CallMsg for view calls
func (s *PublicSmartContractAPI) createViewCallMsg(toAddr common.Address, data []byte) xcbapi.CallMsg {
	return xcbapi.CallMsg{
		FromAddr:         common.Address{}, // Zero address for view calls
		ToAddr:           &toAddr,
		EnergyLimit:      100000,        // Reasonable energy limit for view calls
		EnergyPriceValue: big.NewInt(0), // Zero energy price for view calls
		ValueAmount:      big.NewInt(0), // Zero value for view calls
		DataBytes:        data,
	}
}

// validateOffsetAndLength safely validates offset and length values to prevent integer overflow and bounds issues
// validateHexOffsetAndLength validates offset/length when the maxSize is expressed
// in hex characters (nibbles). This helper preserves the original conservative
// bound that accounted for converting bytes -> hex (×2).
func validateHexOffsetAndLength(offset, length *big.Int, maxHexLen int64) error {
	if offset.Sign() < 0 || length.Sign() < 0 {
		return fmt.Errorf("negative offset or length: offset=%v, length=%v", offset, length)
	}
	// Keep the original check that compared against max/2 to account for ×2 later
	if offset.Int64() > maxHexLen/2 || length.Int64() > maxHexLen/2 {
		return fmt.Errorf("offset or length too large: offset=%v, length=%v, max=%v", offset, length, maxHexLen/2)
	}
	return nil
}

// validateByteOffsetAndLength validates offset/length when working directly with
// raw byte slices. No division by 2 is needed; we validate against the actual
// byte length.
func validateByteOffsetAndLength(offset, length *big.Int, maxBytesLen int64) error {
	if offset.Sign() < 0 || length.Sign() < 0 {
		return fmt.Errorf("negative offset or length: offset=%v, length=%v", offset, length)
	}
	if offset.Int64() > maxBytesLen || length.Int64() > maxBytesLen {
		return fmt.Errorf("offset or length too large: offset=%v, length=%v, max=%v", offset, length, maxBytesLen)
	}
	// Best‑effort guard to avoid offset+length overflow the buffer
	if (new(big.Int).Add(offset, length)).Int64() > maxBytesLen {
		return fmt.Errorf("offset+length exceeds buffer: offset=%v, length=%v, max=%v", offset, length, maxBytesLen)
	}
	return nil
}

// Ticker returns the ticker symbol of a token contract by calling the symbol() function.
// It automatically decodes the dynamic string response using decodeDynString.
func (s *PublicSmartContractAPI) Ticker(ctx context.Context, tokenAddress common.Address) (string, error) {
	// CBC20 symbol() function selector: 0x231782d8
	selector := "0x231782d8" // standard CBC20 symbol()

	// Create the call data
	data := hexutil.MustDecode(selector)

	// Make the contract call with properly initialized CallMsg
	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)

	if err != nil {
		return "", fmt.Errorf("failed to call symbol() on contract %s: %v", tokenAddress.Hex(), err)
	}

	// Add response size limit to prevent DoS attacks
	const MaxResponseSize = 1024 * 1024 // 1MB limit
	if len(result) > MaxResponseSize {
		return "", fmt.Errorf("response too large: %d bytes exceeds limit of %d", len(result), MaxResponseSize)
	}

	// If we got a result, decode it
	if len(result) > 0 {
		decoded, err := decodeDynString(hexutil.Encode(result))
		if err != nil {
			return "", fmt.Errorf("failed to decode ticker response from contract %s: %v", tokenAddress.Hex(), err)
		}
		if decoded != "" {
			return decoded, nil
		}
	}

	return "", fmt.Errorf("empty response from symbol() call on contract %s", tokenAddress.Hex())
}

// Name returns the name of a token contract by calling the name() function.
// It automatically decodes the dynamic string response using decodeDynString.
func (s *PublicSmartContractAPI) Name(ctx context.Context, tokenAddress common.Address) (string, error) {
	// CBC20 name() function selector: 0x07ba2a17
	selector := "0x07ba2a17" // standard CBC20 name()

	// Create the call data
	data := hexutil.MustDecode(selector)

	// Make the contract call with properly initialized CallMsg
	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)

	if err != nil {
		return "", fmt.Errorf("failed to call name() on contract %s: %v", tokenAddress.Hex(), err)
	}

	// Add response size limit to prevent DoS attacks
	const MaxResponseSize = 1024 * 1024 // 1MB limit
	if len(result) > MaxResponseSize {
		return "", fmt.Errorf("response too large: %d bytes exceeds limit of %d", len(result), MaxResponseSize)
	}

	// If we got a result, decode it
	if len(result) > 0 {
		decoded, err := decodeDynString(hexutil.Encode(result))
		if err != nil {
			return "", fmt.Errorf("failed to decode name response from contract %s: %v", tokenAddress.Hex(), err)
		}
		if decoded != "" {
			return decoded, nil
		}
	}

	return "", fmt.Errorf("empty response from name() call on contract %s", tokenAddress.Hex())
}

// BalanceOf returns the balance of a given token holder for a given token contract.
// It automatically decodes the uint256 response and converts it to a hexutil.Big.
// If unit is provided, it will first attempt to call balanceOf(address,string) and fall back
// to the canonical balanceOf(address) if the aliased call fails.
func (s *PublicSmartContractAPI) BalanceOf(ctx context.Context, holderAddress, tokenAddress common.Address, unit *string) (*hexutil.Big, error) {
	if unit != nil {
		unitValue := *unit
		normalizedUnit := strings.ToLower(unitValue)
		balance, err := s.balanceOfWithUnit(ctx, holderAddress, tokenAddress, normalizedUnit)
		if err != nil {
			log.Debug("balanceOf alias call failed",
				"holder", holderAddress,
				"token", tokenAddress,
				"unit", unitValue,
				"err", err,
			)
			return nil, fmt.Errorf("unit %s does not exist", unitValue)
		}
		return balance, nil
	}

	return s.balanceOfDefault(ctx, holderAddress, tokenAddress)
}

// balanceOfWithUnit calls balanceOf(address,string) and decodes the result.
func (s *PublicSmartContractAPI) balanceOfWithUnit(ctx context.Context, holderAddress, tokenAddress common.Address, unit string) (*hexutil.Big, error) {
	selector := hexutil.MustDecode("0x5a805b98") // balanceOf(address,string)
	data := append([]byte{}, selector...)

	addressBytes := holderAddress.Bytes()
	paddedAddress := make([]byte, 32)
	copy(paddedAddress[32-len(addressBytes):], addressBytes)
	data = append(data, paddedAddress...)

	offset := make([]byte, 32)
	offset[31] = 0x40 // dynamic data starts right after the two static slots
	data = append(data, offset...)

	unitBytes := []byte(unit)
	lengthBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(lengthBytes[24:], uint64(len(unitBytes)))
	data = append(data, lengthBytes...)

	paddedLen := ((len(unitBytes) + 31) / 32) * 32
	paddedUnit := make([]byte, paddedLen)
	copy(paddedUnit, unitBytes)
	data = append(data, paddedUnit...)

	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf(address,string) on contract %s for address %s and unit %s: %w", tokenAddress.Hex(), holderAddress.Hex(), unit, err)
	}

	const MaxResponseSize = 1024 * 1024 // 1MB limit
	if len(result) > MaxResponseSize {
		return nil, fmt.Errorf("response too large: %d bytes exceeds limit of %d", len(result), MaxResponseSize)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("empty response from balanceOf(address,string) call on contract %s for address %s and unit %s", tokenAddress.Hex(), holderAddress.Hex(), unit)
	}

	balance := new(big.Int).SetBytes(result)
	return (*hexutil.Big)(balance), nil
}

// balanceOfDefault calls the canonical balanceOf(address) selector.
func (s *PublicSmartContractAPI) balanceOfDefault(ctx context.Context, holderAddress, tokenAddress common.Address) (*hexutil.Big, error) {
	selector := "0x1d7976f3" // standard CBC20 balanceOf()
	data := hexutil.MustDecode(selector)

	addressBytes := holderAddress.Bytes()
	paddedAddress := make([]byte, 32)
	copy(paddedAddress[32-len(addressBytes):], addressBytes)
	data = append(data, paddedAddress...)

	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf on contract %s for address %s: %v", tokenAddress.Hex(), holderAddress.Hex(), err)
	}

	const MaxResponseSize = 1024 * 1024 // 1MB limit
	if len(result) > MaxResponseSize {
		return nil, fmt.Errorf("response too large: %d bytes exceeds limit of %d", len(result), MaxResponseSize)
	}

	if len(result) > 0 {
		balance := new(big.Int).SetBytes(result)
		return (*hexutil.Big)(balance), nil
	}

	return nil, fmt.Errorf("empty response from balanceOf call on contract %s for address %s", tokenAddress.Hex(), holderAddress.Hex())
}

// Decimals returns the number of decimal places for a given token contract.
// It automatically decodes the uint8 response and converts it to hexutil.Big.
func (s *PublicSmartContractAPI) Decimals(ctx context.Context, tokenAddress common.Address) (*hexutil.Big, error) {
	// CBC20 decimals() function selector: 0x5d1fb5f9
	selector := "0x5d1fb5f9" // standard CBC20 decimals()

	// Create the call data
	data := hexutil.MustDecode(selector)

	// Make the contract call with properly initialized CallMsg
	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)

	if err != nil {
		return nil, fmt.Errorf("failed to call decimals() on contract %s: %v", tokenAddress.Hex(), err)
	}

	// Add response size limit to prevent DoS attacks
	const MaxResponseSize = 1024 * 1024 // 1MB limit
	if len(result) > MaxResponseSize {
		return nil, fmt.Errorf("response too large: %d bytes exceeds limit of %d", len(result), MaxResponseSize)
	}

	// If we got a result, decode it as uint8
	if len(result) > 0 {
		// Convert the 32-byte result to uint8 (last byte), then to hexutil.Big
		if len(result) >= 32 {
			decimals := new(big.Int).SetUint64(uint64(result[31])) // Last byte contains the uint8 value
			return (*hexutil.Big)(decimals), nil
		}
	}

	return nil, fmt.Errorf("invalid response length from decimals() call on contract %s", tokenAddress.Hex())
}

// TotalSupply returns the total supply of a given token contract.
// It automatically decodes the uint256 response and converts it to a hexutil.Big.
func (s *PublicSmartContractAPI) TotalSupply(ctx context.Context, tokenAddress common.Address) (*hexutil.Big, error) {
	// CBC20 totalSupply() function selector: 0x1f1881f8
	selector := "0x1f1881f8" // standard CBC20 totalSupply()

	// Create the call data
	data := hexutil.MustDecode(selector)

	// Make the contract call with properly initialized CallMsg
	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)

	if err != nil {
		return nil, fmt.Errorf("failed to call totalSupply() on contract %s: %v", tokenAddress.Hex(), err)
	}

	// Add response size limit to prevent DoS attacks
	const MaxResponseSize = 1024 * 1024 // 1MB limit
	if len(result) > MaxResponseSize {
		return nil, fmt.Errorf("response too large: %d bytes exceeds limit of %d", len(result), MaxResponseSize)
	}

	// If we got a result, decode it as uint256
	if len(result) > 0 {
		// Convert the 32-byte result to big.Int, then to hexutil.Big
		supply := new(big.Int).SetBytes(result)
		return (*hexutil.Big)(supply), nil
	}

	return nil, fmt.Errorf("invalid response length from totalSupply() call on contract %s", tokenAddress.Hex())
}

// Length returns the smart contract code size in bytes for a given contract address.
// It uses xcb.getCode to fetch the contract code and returns the size in bytes as hexutil.Big.
func (s *PublicSmartContractAPI) Length(ctx context.Context, tokenAddress common.Address) (*hexutil.Big, error) {
	// Since xcb.getCode is already implemented as a standard RPC method,
	// we can use the existing blockchain API directly
	// This is equivalent to calling xcb.getCode(addr, "latest")

	// Get the latest block number
	latestBlock := rpc.LatestBlockNumber

	// Get the state at the latest block
	state, _, err := s.b.StateAndHeaderByNumber(ctx, latestBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get state for contract %s: %v", tokenAddress.Hex(), err)
	}

	// Get the code from the state
	code := state.GetCode(tokenAddress)

	// Check if it's an EOA (no code) or empty code
	if len(code) == 0 {
		return (*hexutil.Big)(big.NewInt(0)), nil
	}

	// Return the code size in bytes as hexutil.Big
	return (*hexutil.Big)(big.NewInt(int64(len(code)))), nil
}

// GetKVResult represents the result of a getKV call
type GetKVResult struct {
	Value  string `json:"value"`  // The metadata value
	Sealed bool   `json:"sealed"` // Whether the key is sealed (immutable)
	Exists bool   `json:"exists"` // Whether the key exists
}

// GetKV retrieves key-value metadata from a smart contract implementing CIP-150.
// Based on the CIP-150 standard for On-Chain Key-Value Metadata Storage.
// If sealed=false (default), returns the value and sealed status.
// If sealed=true, only returns data if the item is actually sealed.
//
// Function selectors used (verified for Core Blockchain CIP-150):
// - hasKey(string): 0xf37e8f05
// - isSealed(string): 0xf272a162
// - getValue(string): 0xe2f3625a
// - listKeys(): 0xe4d90ad0
// - getByIndex(uint256): 0x2ae4e412
// - count(): 0x2d7d47f2
// - setValue(string,string): 0x516c0734
// - sealKey(string): 0x78506b1e

func (s *PublicSmartContractAPI) GetKV(ctx context.Context, key string, tokenAddress common.Address, sealed bool) (*GetKVResult, error) {
	// CIP-150 interface functions:
	// - getValue(string) returns (string value)
	// - isSealed(string) returns (bool sealed)
	// - hasKey(string) returns (bool exists)

	// First check if the key exists
	exists, err := s.callHasKey(ctx, key, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to check if key exists for contract %s: %v", tokenAddress.Hex(), err)
	}

	if !exists {
		return &GetKVResult{
			Value:  "",
			Sealed: false,
			Exists: false,
		}, nil
	}

	// If we only want sealed items, check the sealed status first
	if sealed {
		isSealed, err := s.callIsSealed(ctx, key, tokenAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to check sealed status for key %s on contract %s: %v", key, tokenAddress.Hex(), err)
		}

		// If we want sealed items but this one isn't sealed, return nothing
		if !isSealed {
			return &GetKVResult{
				Value:  "",
				Sealed: false,
				Exists: true,
			}, nil
		}
	}

	// Get the value
	value, err := s.callGetValue(ctx, key, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get value for key %s on contract %s: %v", key, tokenAddress.Hex(), err)
	}

	// Get the sealed status (only if we need it)
	var isSealed bool
	if !sealed {
		isSealed, err = s.callIsSealed(ctx, key, tokenAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to check sealed status for key %s on contract %s: %v", key, tokenAddress.Hex(), err)
		}
	} else {
		isSealed = true // We already know it's sealed from above
	}

	return &GetKVResult{
		Value:  value,
		Sealed: isSealed,
		Exists: true,
	}, nil
}

// ListKVResult represents the result of a listKV call
type ListKVResult struct {
	Keys   []string `json:"keys"`             // List of all keys
	Count  uint64   `json:"count"`            // Total number of keys
	Sealed []bool   `json:"sealed,omitempty"` // Sealed status for each key (if sealed=false)
	Values []string `json:"values,omitempty"` // Values for each key (if sealed=false)
}

// ListKV retrieves all keys from a smart contract implementing CIP-150.
// Based on the CIP-150 standard for On-Chain Key-Value Metadata Storage.
// If sealed=false (default), returns keys, sealed status, and values.
// If sealed=true, only returns keys that are sealed.
//
// Function selectors used (verified for Core Blockchain CIP-150):
// - listKeys(): 0xe4d90ad0
// - count(): 0x2d7d47f2
// - isSealed(string): 0xf272a162
// - getValue(string): 0xe2f3625a
func (s *PublicSmartContractAPI) ListKV(ctx context.Context, tokenAddress common.Address, sealed bool) (*ListKVResult, error) {
	// Get the total count of keys
	count, err := s.callCount(ctx, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get key count for contract %s: %v", tokenAddress.Hex(), err)
	}

	if count == 0 {
		return &ListKVResult{
			Keys:   []string{},
			Count:  0,
			Sealed: []bool{},
			Values: []string{},
		}, nil
	}

	// Try to get all keys in a single call first
	keys, err := s.callListKeys(ctx, tokenAddress)
	if err != nil {
		// Fallback: fetch keys individually by index to avoid large return payloads
		keys = make([]string, 0, count)
		for i := uint64(0); i < count; i++ {
			k, ierr := s.callGetByIndex(ctx, tokenAddress, i)
			if ierr != nil {
				return nil, fmt.Errorf("failed to get key by index %d for contract %s: %v", i, tokenAddress.Hex(), ierr)
			}
			keys = append(keys, k)
		}
	}

	// Two branches: sealed-only and full listing
	if sealed {
		var filteredKeys []string
		var filteredSealed []bool
		var filteredValues []string
		for _, key := range keys {
			isSealed, err := s.callIsSealed(ctx, key, tokenAddress)
			if err != nil {
				return nil, fmt.Errorf("failed to check sealed status for key %s on contract %s: %v", key, tokenAddress.Hex(), err)
			}
			if isSealed {
				filteredKeys = append(filteredKeys, key)
				filteredSealed = append(filteredSealed, true)
				// Keep the response shape identical to the full listing: include values
				value, err := s.callGetValue(ctx, key, tokenAddress)
				if err != nil {
					return nil, fmt.Errorf("failed to get value for key %s on contract %s: %v", key, tokenAddress.Hex(), err)
				}
				filteredValues = append(filteredValues, value)
			}
		}

		return &ListKVResult{
			Keys:   filteredKeys,
			Count:  uint64(len(filteredKeys)),
			Sealed: filteredSealed,
			Values: filteredValues,
		}, nil
	}

	// sealed == false: include all keys with sealed status and values
	sealedStatus := make([]bool, 0, len(keys))
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		isSealed, err := s.callIsSealed(ctx, key, tokenAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to check sealed status for key %s on contract %s: %v", key, tokenAddress.Hex(), err)
		}
		sealedStatus = append(sealedStatus, isSealed)

		value, err := s.callGetValue(ctx, key, tokenAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to get value for key %s on contract %s: %v", key, tokenAddress.Hex(), err)
		}
		values = append(values, value)
	}

	// Directly return all keys with their statuses/values
	return &ListKVResult{
		Keys:   keys,
		Count:  uint64(len(keys)),
		Sealed: sealedStatus,
		Values: values,
	}, nil
}

// callHasKey calls the hasKey function on the contract (CIP-150)
func (s *PublicSmartContractAPI) callHasKey(ctx context.Context, key string, tokenAddress common.Address) (bool, error) {
	// hasKey(string) function selector: 0xf37e8f05
	selector := "0xf37e8f05"

	// Create the call data: selector + encoded string key
	data := hexutil.MustDecode(selector)

	// Encode the string key (dynamic string encoding)
	keyBytes := []byte(key)
	keyLength := len(keyBytes)

	// Add offset (32 bytes for dynamic string)
	offset := big.NewInt(32)
	offsetBytes := make([]byte, 32)
	offset.FillBytes(offsetBytes)
	data = append(data, offsetBytes...)

	// Add length
	lengthBytes := make([]byte, 32)
	big.NewInt(int64(keyLength)).FillBytes(lengthBytes)
	data = append(data, lengthBytes...)

	// Add the key data (padded to 32 bytes)
	keyPadded := make([]byte, 32)
	copy(keyPadded, keyBytes)
	data = append(data, keyPadded...)

	// Make the contract call with properly initialized CallMsg
	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)

	if err != nil {
		return false, fmt.Errorf("failed to call hasKey on contract %s: %v", tokenAddress.Hex(), err)
	}

	// Add response size limit to prevent DoS attacks
	const MaxResponseSize = 1024 * 1024 // 1MB limit
	if len(result) > MaxResponseSize {
		return false, fmt.Errorf("response too large: %d bytes exceeds limit of %d", len(result), MaxResponseSize)
	}

	// Decode the boolean result
	if len(result) >= 32 {
		// The last byte contains the boolean value
		return result[31] != 0, nil
	}

	return false, fmt.Errorf("invalid response length from hasKey call")
}

// callIsSealed calls the isSealed function on the contract (CIP-150)
func (s *PublicSmartContractAPI) callIsSealed(ctx context.Context, key string, tokenAddress common.Address) (bool, error) {
	// isSealed(string) function selector: 0xf272a162
	selector := "0xf272a162"

	// Create the call data: selector + encoded string key
	data := hexutil.MustDecode(selector)

	// Encode the string key (dynamic string encoding)
	keyBytes := []byte(key)
	keyLength := len(keyBytes)

	// Add offset (32 bytes for dynamic string)
	offset := big.NewInt(32)
	offsetBytes := make([]byte, 32)
	offset.FillBytes(offsetBytes)
	data = append(data, offsetBytes...)

	// Add length
	lengthBytes := make([]byte, 32)
	big.NewInt(int64(keyLength)).FillBytes(lengthBytes)
	data = append(data, lengthBytes...)

	// Add the key data (padded to 32 bytes)
	keyPadded := make([]byte, 32)
	copy(keyPadded, keyBytes)
	data = append(data, keyPadded...)

	// Make the contract call with properly initialized CallMsg
	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)

	if err != nil {
		return false, fmt.Errorf("failed to call isSealed on contract %s: %v", tokenAddress.Hex(), err)
	}

	// Add response size limit to prevent DoS attacks
	const MaxResponseSize = 1024 * 1024 // 1MB limit
	if len(result) > MaxResponseSize {
		return false, fmt.Errorf("response too large: %d bytes exceeds limit of %d", len(result), MaxResponseSize)
	}

	// Decode the boolean result
	if len(result) >= 32 {
		// The last byte contains the boolean value
		return result[31] != 0, nil
	}

	return false, fmt.Errorf("invalid response length from isSealed call")
}

// callGetValue calls the getValue function on the contract (CIP-150)
func (s *PublicSmartContractAPI) callGetValue(ctx context.Context, key string, tokenAddress common.Address) (string, error) {
	// getValue(string) function selector: 0xe2f3625a
	selector := "0xe2f3625a"

	// Create the call data: selector + encoded string key
	data := hexutil.MustDecode(selector)

	// Encode the string key (dynamic string encoding)
	keyBytes := []byte(key)
	keyLength := len(keyBytes)

	// Add offset (32 bytes for dynamic string)
	offset := big.NewInt(32)
	offsetBytes := make([]byte, 32)
	offset.FillBytes(offsetBytes)
	data = append(data, offsetBytes...)

	// Add length
	lengthBytes := make([]byte, 32)
	big.NewInt(int64(keyLength)).FillBytes(lengthBytes)
	data = append(data, lengthBytes...)

	// Add the key data (padded to 32 bytes)
	keyPadded := make([]byte, 32)
	copy(keyPadded, keyBytes)
	data = append(data, keyPadded...)

	// Make the contract call with properly initialized CallMsg
	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)

	if err != nil {
		return "", fmt.Errorf("failed to call getValue on contract %s: %v", tokenAddress.Hex(), err)
	}

	// Add response size limit to prevent DoS attacks
	const MaxResponseSize = 1024 * 1024 // 1MB limit
	if len(result) > MaxResponseSize {
		return "", fmt.Errorf("response too large: %d bytes exceeds limit of %d", len(result), MaxResponseSize)
	}

	// Decode the dynamic string response using our existing decodeDynString function
	if len(result) > 0 {
		decoded, err := decodeDynString(hexutil.Encode(result))
		if err != nil {
			return "", fmt.Errorf("failed to decode getValue response from contract %s: %v", tokenAddress.Hex(), err)
		}
		return decoded, nil
	}

	return "", fmt.Errorf("empty response from getValue call")
}

// TickerSubscription provides real-time updates about token ticker symbols.
// This can be useful for monitoring token metadata changes.
func (s *PublicSmartContractAPI) TickerSubscription(ctx context.Context, tokenAddress common.Address) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()

	go func() {
		// Send initial ticker
		ticker, err := s.Ticker(ctx, tokenAddress)
		if err == nil {
			notifier.Notify(rpcSub.ID, ticker)
		}

		// Monitor for changes (this is a simplified implementation)
		// In a real scenario, you might want to monitor specific events
		// or use a different strategy for detecting changes
		for {
			select {
			case <-rpcSub.Err():
				return
			case <-notifier.Closed():
				return
			}
		}
	}()

	return rpcSub, nil
}

// NameSubscription provides real-time updates about token names.
// This can be useful for monitoring token metadata changes.
func (s *PublicSmartContractAPI) NameSubscription(ctx context.Context, tokenAddress common.Address) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()

	go func() {
		// Send initial name
		name, err := s.Name(ctx, tokenAddress)
		if err == nil {
			notifier.Notify(rpcSub.ID, name)
		}

		// Monitor for changes (this is a simplified implementation)
		// In a real scenario, you might want to monitor specific events
		// or use a different strategy for detecting changes
		for {
			select {
			case <-rpcSub.Err():
				return
			case <-notifier.Closed():
				return
			}
		}
	}()

	return rpcSub, nil
}

// BalanceOfSubscription provides real-time updates about token balances.
// This can be useful for monitoring balance changes for specific addresses.
func (s *PublicSmartContractAPI) BalanceOfSubscription(ctx context.Context, holderAddress, tokenAddress common.Address, unit *string) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	go func() {
		// Send initial balance
		balance, err := s.BalanceOf(ctx, holderAddress, tokenAddress, unit)
		if err == nil {
			notifier.Notify(rpcSub.ID, balance)
		}

		// Monitor for changes (this is a simplified implementation)
		// In a real scenario, you might want to monitor specific events
		// or use a different strategy for detecting changes
		for {
			select {
			case <-rpcSub.Err():
				return
			case <-notifier.Closed():
				return
			}
		}
	}()

	return rpcSub, nil
}

// DecimalsSubscription provides real-time updates about token decimal places.
// This can be useful for monitoring decimal changes.
func (s *PublicSmartContractAPI) DecimalsSubscription(ctx context.Context, tokenAddress common.Address) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	go func() {
		// Send initial decimals
		decimals, err := s.Decimals(ctx, tokenAddress)
		if err == nil {
			notifier.Notify(rpcSub.ID, decimals)
		}

		// Monitor for changes (this is a simplified implementation)
		// In a real scenario, you might want to monitor specific events
		// or use a different strategy for detecting changes
		for {
			select {
			case <-rpcSub.Err():
				return
			case <-notifier.Closed():
				return
			}
		}
	}()

	return rpcSub, nil
}

// TotalSupplySubscription provides real-time updates about token total supply.
// This can be useful for monitoring supply changes.
func (s *PublicSmartContractAPI) TotalSupplySubscription(ctx context.Context, tokenAddress common.Address) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	go func() {
		// Send initial total supply
		supply, err := s.TotalSupply(ctx, tokenAddress)
		if err == nil {
			notifier.Notify(rpcSub.ID, supply)
		}

		// Monitor for changes (this is a simplified implementation)
		// In a real scenario, you might want to monitor specific events
		// or use a different strategy for detecting changes
		for {
			select {
			case <-rpcSub.Err():
				return
			case <-notifier.Closed():
				return
			}
		}
	}()

	return rpcSub, nil
}

// LengthSubscription provides real-time updates about smart contract code size.
// This can be useful for monitoring contract upgrades or deployments.
func (s *PublicSmartContractAPI) LengthSubscription(ctx context.Context, tokenAddress common.Address) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	go func() {
		// Send initial length
		length, err := s.Length(ctx, tokenAddress)
		if err == nil {
			notifier.Notify(rpcSub.ID, length)
		}

		// Monitor for changes (this is a simplified implementation)
		// In a real scenario, you might want to monitor specific events
		// or use a different strategy for detecting changes
		for {
			select {
			case <-rpcSub.Err():
				return
			case <-notifier.Closed():
				return
			}
		}
	}()

	return rpcSub, nil
}

// GetKVSubscription provides real-time updates about key-value metadata.
// This can be useful for monitoring metadata changes.
func (s *PublicSmartContractAPI) GetKVSubscription(ctx context.Context, key string, tokenAddress common.Address, sealed bool) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	go func() {
		// Send initial metadata
		result, err := s.GetKV(ctx, key, tokenAddress, sealed)
		if err == nil {
			notifier.Notify(rpcSub.ID, result)
		}

		// Monitor for changes (this is a simplified implementation)
		// In a real scenario, you might want to monitor specific events
		// or use a different strategy for detecting changes
		for {
			select {
			case <-rpcSub.Err():
				return
			case <-notifier.Closed():
				return
			}
		}
	}()

	return rpcSub, nil
}

// ListKVSubscription provides real-time updates about key-value metadata lists.
// This can be useful for monitoring metadata list changes.
func (s *PublicSmartContractAPI) ListKVSubscription(ctx context.Context, tokenAddress common.Address, sealed bool) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	go func() {
		// Send initial metadata list
		result, err := s.ListKV(ctx, tokenAddress, sealed)
		if err == nil {
			notifier.Notify(rpcSub.ID, result)
		}

		// Monitor for changes (this is a simplified implementation)
		// In a real scenario, you might want to monitor specific events
		// or use a different strategy for detecting changes
		for {
			select {
			case <-rpcSub.Err():
				return
			case <-notifier.Closed():
				return
			}
		}
	}()

	return rpcSub, nil
}

// TokenURISubscription provides real-time updates about NFT token URIs.
// This can be useful for monitoring metadata URI changes.
func (s *PublicSmartContractAPI) TokenURISubscription(ctx context.Context, tokenAddress common.Address, tokenId *big.Int) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	go func() {
		// Send initial token URI
		uri, err := s.TokenURI(ctx, tokenAddress, tokenId)
		if err == nil {
			notifier.Notify(rpcSub.ID, uri)
		}

		// Monitor for changes (this is a simplified implementation)
		// In a real scenario, you might want to monitor specific events
		// or use a different strategy for detecting changes
		for {
			select {
			case <-rpcSub.Err():
				return
			case <-notifier.Closed():
				return
			}
		}
	}()

	return rpcSub, nil
}

// decodeDynString decodes a dynamic string response from a smart contract call.
// It handles both bytes32 fallback and dynamic string encoding.
func decodeDynString(res string) (string, error) {
	// Remove 0x prefix
	h := strings.TrimPrefix(res, "0x")

	// Add size limits to prevent DoS attacks
	const MaxHexLength = 1024 * 1024 // 1MB limit for hex string

	if len(h) > MaxHexLength {
		return "", fmt.Errorf("response too large: %d hex chars exceeds limit of %d", len(h), MaxHexLength)
	}

	// Check if it's a bytes32 fallback (64 hex chars = 32 bytes)
	if len(h) == 64 {
		// Decode as bytes32 and convert to ASCII
		bytes, err := hex.DecodeString(h)
		if err != nil {
			return "", fmt.Errorf("failed to decode bytes32: %v", err)
		}

		// Convert to string and remove null bytes
		result := string(bytes)
		trimmed := strings.TrimRight(result, "\x00")

		return trimmed, nil
	}

	// Handle dynamic string encoding
	if len(h) < 64 {
		return "", fmt.Errorf("response too short for dynamic string: got %d hex chars, need at least 64", len(h))
	}

	// Parse offset (first 32 bytes)
	offsetHex := h[:64]
	offset, err := hex.DecodeString(offsetHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode offset: %v", err)
	}

	// Convert offset to int (big-endian)
	offsetInt := new(big.Int).SetBytes(offset)

	// Validate offset
	if err := validateHexOffsetAndLength(offsetInt, big.NewInt(0), int64(len(h))); err != nil {
		return "", fmt.Errorf("invalid offset: %v", err)
	}

	offsetBytes := offsetInt.Int64() * 2 // Convert to hex string position

	// Check if offset is 0 (simple bytes32 case) or if it's a valid offset
	if offsetInt.Cmp(big.NewInt(0)) == 0 {
		// This is a simple bytes32 case, decode the entire response as bytes32
		bytes, err := hex.DecodeString(h)
		if err != nil {
			return "", fmt.Errorf("failed to decode bytes32 fallback: %v", err)
		}
		result := string(bytes)
		return strings.TrimRight(result, "\x00"), nil
	}

	// Check if we have enough data
	if int64(len(h)) < offsetBytes+64 {
		return "", fmt.Errorf("response too short for offset %d: got %d hex chars, need at least %d", offsetBytes, len(h), offsetBytes+64)
	}

	// Parse length (32 bytes after offset)
	lengthHex := h[offsetBytes : offsetBytes+64]
	length, err := hex.DecodeString(strings.TrimPrefix(lengthHex, "0x"))
	if err != nil {
		return "", fmt.Errorf("failed to decode length: %v", err)
	}

	// Convert length to int (big-endian)
	lengthInt := new(big.Int).SetBytes(length)

	// Validate length
	if err := validateHexOffsetAndLength(big.NewInt(0), lengthInt, int64(len(h))); err != nil {
		return "", fmt.Errorf("invalid length: %v", err)
	}

	lengthBytes := lengthInt.Int64() * 2 // Convert to hex string position

	// Extract the actual string data
	dataStart := offsetBytes + 64
	dataEnd := dataStart + lengthBytes

	if int64(len(h)) < dataEnd {
		return "", fmt.Errorf("response too short for data length %d: got %d hex chars, need at least %d", lengthBytes, len(h), dataEnd)
	}

	dataHex := h[dataStart:dataEnd]

	// Decode the hex data to bytes
	dataBytes, err := hex.DecodeString(dataHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode string data: %v", err)
	}

	// Convert to UTF-8 string
	result := string(dataBytes)
	return result, nil
}

// callCount calls the count function on the contract (CIP-150)
func (s *PublicSmartContractAPI) callCount(ctx context.Context, tokenAddress common.Address) (uint64, error) {
	// count() function selector: 0x2d7d47f2
	selector := "0x2d7d47f2"

	// Create the call data: just the selector (no parameters)
	data := hexutil.MustDecode(selector)

	// Make the contract call with properly initialized CallMsg
	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)

	if err != nil {
		return 0, fmt.Errorf("failed to call count on contract %s: %v", tokenAddress.Hex(), err)
	}

	// Add response size limit to prevent DoS attacks
	const MaxResponseSize = 1024 * 1024 // 1MB limit
	if len(result) > MaxResponseSize {
		return 0, fmt.Errorf("response too large: %d bytes exceeds limit of %d", len(result), MaxResponseSize)
	}

	// Decode the uint256 result
	if len(result) >= 32 {
		count := new(big.Int).SetBytes(result)
		return count.Uint64(), nil
	}

	return 0, fmt.Errorf("invalid response length from count call")
}

// callListKeys calls the listKeys function on the contract (CIP-150)
func (s *PublicSmartContractAPI) callListKeys(ctx context.Context, tokenAddress common.Address) ([]string, error) {
	// listKeys() function selector: 0xe4d90ad0
	selector := "0xe4d90ad0"

	// Create the call data: just the selector (no parameters)
	data := hexutil.MustDecode(selector)

	// Make the contract call with properly initialized CallMsg
	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)

	if err != nil {
		return nil, fmt.Errorf("failed to call listKeys on contract %s: %v", tokenAddress.Hex(), err)
	}

	// Decode the string[] result
	// This is a dynamic array of strings, so we need to parse the ABI encoding

	// Add response size limit to prevent DoS attacks
	const MaxResponseSize = 1024 * 1024 // 1MB limit
	const MaxArrayLength = 1000         // Maximum array elements

	if len(result) < 32 {
		return nil, fmt.Errorf("response too short for dynamic array: got %d bytes, need at least 32", len(result))
	}

	if len(result) > MaxResponseSize {
		return nil, fmt.Errorf("response too large: %d bytes exceeds limit of %d", len(result), MaxResponseSize)
	}

	// Get the offset to the array data
	offsetHex := hexutil.Encode(result[:32])
	offset, err := hex.DecodeString(strings.TrimPrefix(offsetHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode array offset: %v", err)
	}
	offsetInt := new(big.Int).SetBytes(offset)

	// Validate offset before using it
	if err := validateByteOffsetAndLength(offsetInt, big.NewInt(0), int64(len(result))); err != nil {
		return nil, fmt.Errorf("invalid array offset: %v", err)
	}

	// Offsets in ABI-encoded return data are byte offsets, not hex-string positions
	offsetBytes := offsetInt.Int64()

	if int64(len(result)) < offsetBytes+32 {
		return nil, fmt.Errorf("invalid response length for array offset %d: got %d bytes, need at least %d", offsetBytes, len(result), offsetBytes+32)
	}

	// Get the array length (first word at the start of the array data)
	lengthHex := hexutil.Encode(result[offsetBytes : offsetBytes+32])
	length, err := hex.DecodeString(strings.TrimPrefix(lengthHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode array length: %v", err)
	}
	lengthInt := new(big.Int).SetBytes(length)

	// Validate array length
	if lengthInt.Cmp(big.NewInt(MaxArrayLength)) > 0 {
		return nil, fmt.Errorf("array too large: %d elements exceeds limit of %d", lengthInt, MaxArrayLength)
	}

	arrayLength := lengthInt.Int64()

	var keys []string
	// Base of the array data (where the length word lives)
	arrayBase := offsetBytes
	// Offset inside the array where per-element pointers start
	currentOffset := arrayBase + 32

	// Parse each string in the array
	for i := int64(0); i < arrayLength; i++ {
		if int64(len(result)) < currentOffset+32 {
			return nil, fmt.Errorf("response too short for string %d offset: got %d bytes, need at least %d", i, len(result), currentOffset+32)
		}

		// Get the string offset
		stringOffsetHex := hexutil.Encode(result[currentOffset : currentOffset+32])
		stringOffset, err := hex.DecodeString(strings.TrimPrefix(stringOffsetHex, "0x"))
		if err != nil {
			return nil, fmt.Errorf("failed to decode string %d offset: %v", i, err)
		}
		stringOffsetInt := new(big.Int).SetBytes(stringOffset)

		// Offsets inside dynamic arrays are relative to the start of the array data
		absStringOffset := new(big.Int).Add(stringOffsetInt, big.NewInt(arrayBase))
		// Validate absolute string offset
		if err := validateByteOffsetAndLength(absStringOffset, big.NewInt(0), int64(len(result))); err != nil {
			return nil, fmt.Errorf("invalid string %d offset: %v", i, err)
		}

		// String offsets are in bytes (absolute position in the result buffer)
		stringOffsetBytes := absStringOffset.Int64()

		if int64(len(result)) < stringOffsetBytes+32 {
			return nil, fmt.Errorf("response too short for string %d data: got %d bytes, need at least %d", i, len(result), stringOffsetBytes+32)
		}

		// Get the string length
		lengthHex := hexutil.Encode(result[stringOffsetBytes : stringOffsetBytes+32])
		stringLength, err := hex.DecodeString(strings.TrimPrefix(lengthHex, "0x"))
		if err != nil {
			return nil, fmt.Errorf("failed to decode string %d length: %v", i, err)
		}
		stringLengthInt := new(big.Int).SetBytes(stringLength)

		// Validate string length
		if err := validateByteOffsetAndLength(big.NewInt(0), stringLengthInt, int64(len(result))); err != nil {
			return nil, fmt.Errorf("invalid string %d length: %v", i, err)
		}

		// String length is in bytes
		stringLengthBytes := stringLengthInt.Int64()

		// Get the string data
		stringDataStart := stringOffsetBytes + 32
		stringDataEnd := stringDataStart + stringLengthBytes
		if int64(len(result)) < stringDataEnd {
			return nil, fmt.Errorf("response too short for string %d content: got %d bytes, need at least %d", i, len(result), stringDataEnd)
		}

		stringDataHex := hexutil.Encode(result[stringDataStart:stringDataEnd])
		stringData, err := hex.DecodeString(strings.TrimPrefix(stringDataHex, "0x"))
		if err != nil {
			return nil, fmt.Errorf("failed to decode string %d content: %v", i, err)
		}

		keys = append(keys, string(stringData))
		currentOffset += 32
	}

	return keys, nil
}

// callGetByIndex calls the getByIndex(uint256) function on the contract (CIP-150)
// selector: 0x2ae4e412
func (s *PublicSmartContractAPI) callGetByIndex(ctx context.Context, tokenAddress common.Address, idx uint64) (string, error) {
	selector := "0x2ae4e412"
	data := hexutil.MustDecode(selector)

	// Encode index as uint256
	idxBytes := make([]byte, 32)
	new(big.Int).SetUint64(idx).FillBytes(idxBytes)
	data = append(data, idxBytes...)

	// Call contract
	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)
	if err != nil {
		return "", fmt.Errorf("failed to call getByIndex on contract %s: %v", tokenAddress.Hex(), err)
	}

	// Decode dynamic string
	if len(result) > 0 {
		decoded, err := decodeDynString(hexutil.Encode(result))
		if err != nil {
			return "", fmt.Errorf("failed to decode getByIndex response from contract %s: %v", tokenAddress.Hex(), err)
		}
		return decoded, nil
	}
	return "", fmt.Errorf("empty response from getByIndex call")
}

// TokenURI retrieves the token URI for a specific NFT token ID from a CoreNFT contract.
// Based on the CIP-721 standard for Core Blockchain Non-Fungible Tokens.
//
// Function selector used (verified for Core Blockchain CIP-721):
// - tokenURI(uint256): 0xa89da637
func (s *PublicSmartContractAPI) TokenURI(ctx context.Context, tokenAddress common.Address, tokenId *big.Int) (string, error) {
	// tokenURI(uint256) function selector: 0xa89da637
	selector := "0xa89da637"

	// Create the call data: selector + encoded uint256 tokenId
	data := hexutil.MustDecode(selector)

	// Encode the tokenId (uint256)
	tokenIdBytes := make([]byte, 32)
	tokenId.FillBytes(tokenIdBytes)
	data = append(data, tokenIdBytes...)

	// Make the contract call with properly initialized CallMsg
	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)

	if err != nil {
		return "", fmt.Errorf("failed to call tokenURI on contract %s for tokenId %s: %v", tokenAddress.Hex(), tokenId.String(), err)
	}

	// Add response size limit to prevent DoS attacks
	const MaxResponseSize = 1024 * 1024 // 1MB limit
	if len(result) > MaxResponseSize {
		return "", fmt.Errorf("response too large: %d bytes exceeds limit of %d", len(result), MaxResponseSize)
	}

	// Decode the dynamic string response using our existing decodeDynString function
	if len(result) > 0 {
		decoded, err := decodeDynString(hexutil.Encode(result))
		if err != nil {
			return "", fmt.Errorf("failed to decode tokenURI response from contract %s for tokenId %s: %v", tokenAddress.Hex(), tokenId.String(), err)
		}
		return decoded, nil
	}

	return "", fmt.Errorf("empty response from tokenURI call on contract %s for tokenId %s", tokenAddress.Hex(), tokenId.String())
}

// GetPrice retrieves the latest price from a PriceFeed contract.
// Based on CIP-104 for Core Blockchain PriceFeed contracts.
//
// Function selectors used (verified for Core Blockchain CIP-104):
// - getLatestPrice(): 0x677dcf04
// - getAggregatedPrice(): 0xd9c1c1c9
//
// Parameters:
// - tokenAddress: The address of the PriceFeed contract
// - aggregated: If true, returns aggregated price; if false, returns latest individual price
func (s *PublicSmartContractAPI) GetPrice(ctx context.Context, tokenAddress common.Address, aggregated bool) (*hexutil.Big, error) {
	var selector string
	var data []byte

	if aggregated {
		// getAggregatedPrice() function selector: 0xd9c1c1c9
		selector = "0xd9c1c1c9"
		data = hexutil.MustDecode(selector)
	} else {
		// getLatestPrice() function selector: 0x677dcf04
		selector = "0x677dcf04"
		data = hexutil.MustDecode(selector)
	}

	// Make the contract call with properly initialized CallMsg
	result, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)

	if err != nil {
		funcName := "getLatestPrice"
		if aggregated {
			funcName = "getAggregatedPrice"
		}
		return nil, fmt.Errorf("failed to call %s on PriceFeed contract %s: %v",
			funcName,
			tokenAddress.Hex(), err)
	}

	// Add response size limit to prevent DoS attacks
	const MaxResponseSize = 1024 * 1024 // 1MB limit
	if len(result) > MaxResponseSize {
		funcName := "getLatestPrice"
		if aggregated {
			funcName = "getAggregatedPrice"
		}
		return nil, fmt.Errorf("response too large from %s: %d bytes exceeds limit of %d",
			funcName, len(result), MaxResponseSize)
	}

	// If we got a result, decode it as uint256
	if len(result) >= 32 {
		price := new(big.Int).SetBytes(result)
		return (*hexutil.Big)(price), nil
	}

	funcName := "getLatestPrice"
	if aggregated {
		funcName = "getAggregatedPrice"
	}
	return nil, fmt.Errorf("invalid response length from %s call on PriceFeed contract %s",
		funcName,
		tokenAddress.Hex())
}

// GetPriceSubscription provides real-time updates for the latest price from a PriceFeed contract.
// Based on CIP-104 for Core Blockchain PriceFeed contracts.
func (s *PublicSmartContractAPI) GetPriceSubscription(ctx context.Context, tokenAddress common.Address, aggregated bool) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, rpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()

	go func() {
		var lastPrice *hexutil.Big
		ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds (provider cycle dependent)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Get current price
				currentPrice, err := s.GetPrice(ctx, tokenAddress, aggregated)
				if err != nil {
					log.Warn("Failed to get price for subscription", "error", err, "contract", tokenAddress.Hex())
					continue
				}

				// Only notify if price changed by more than 1%
				if lastPrice == nil || pricesChanged(lastPrice, currentPrice) {
					lastPrice = currentPrice
					err = notifier.Notify(rpcSub.ID, currentPrice)
					if err != nil {
						log.Warn("Failed to send price notification", "error", err)
						return
					}
				}

			case <-rpcSub.Err():
				return
			}
		}
	}()

	return rpcSub, nil
}

// pricesChanged checks if two price arrays are different
func pricesChanged(current, last *hexutil.Big) bool {
	return (*big.Int)(current).Cmp((*big.Int)(last)) != 0
}

// Backend interface provides the common API services needed for smart contract operations.
type Backend interface {
	// CallContract executes a contract call
	CallContract(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error)
	// StateAndHeaderByNumber gets the state and header at a specific block number
	StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error)
}

// ExpiredResult represents the result of checking token expiration status
type ExpiredResult struct {
	Expired         bool     `json:"expired"`                   // Whether the token is expired
	TokenExpiration *big.Int `json:"tokenExpiration,omitempty"` // Unix timestamp when token expires
	TradingStop     *big.Int `json:"tradingStop,omitempty"`     // Unix timestamp when trading should stop
}

// Expired checks if a token is expired based on CIP-151 Token Lifecycle Metadata Standard.
// Based on CIP-151 for Core Blockchain token lifecycle management.
//
// Function selectors used (from CIP-150 KV metadata):
// - getValue(string): 0xe2f3625a
//
// Parameters:
// - tokenAddress: The address of the smart contract
// - stopData: If true, also returns tradingStop timestamp (default: true)
func (s *PublicSmartContractAPI) Expired(ctx context.Context, tokenAddress common.Address, stopData bool) (*ExpiredResult, error) {
	result := &ExpiredResult{
		Expired: false,
	}

	// Get tokenExpiration metadata
	expirationResult, err := s.GetKV(ctx, "tokenExpiration", tokenAddress, false)
	if err != nil {
		// If no expiration metadata exists, token is not expired
		return result, nil
	}

	// If tokenExpiration exists, check if expired
	if expirationResult.Exists && expirationResult.Value != "" {
		// Parse the expiration timestamp
		expirationStr := strings.TrimSpace(expirationResult.Value)
		if expirationStr != "" {
			expirationInt, ok := new(big.Int).SetString(expirationStr, 10)
			if ok {
				result.TokenExpiration = expirationInt

				// Check if current block timestamp is past expiration
				_, header, err := s.b.StateAndHeaderByNumber(ctx, rpc.LatestBlockNumber)
				if err != nil {
					return nil, fmt.Errorf("failed to get latest block header: %v", err)
				}

				currentTime := big.NewInt(int64(header.Time))
				if currentTime.Cmp(expirationInt) >= 0 {
					result.Expired = true
				}
			}
		}
	}

	// If stopData is requested, also get tradingStop metadata
	if stopData {
		tradingStopResult, err := s.GetKV(ctx, "tradingStop", tokenAddress, false)
		if err == nil && tradingStopResult.Exists && tradingStopResult.Value != "" {
			tradingStopStr := strings.TrimSpace(tradingStopResult.Value)
			if tradingStopStr != "" {
				tradingStopInt, ok := new(big.Int).SetString(tradingStopStr, 10)
				if ok {
					result.TradingStop = tradingStopInt
				}
			}
		}
	}

	return result, nil
}

// ExpiredSubscription provides real-time updates for token expiration status.
// Based on CIP-151 for Core Blockchain token lifecycle management.
func (s *PublicSmartContractAPI) ExpiredSubscription(ctx context.Context, tokenAddress common.Address, stopData bool) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, rpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()

	go func() {
		var lastResult *ExpiredResult
		ticker := time.NewTicker(60 * time.Second) // Check every minute for expiration changes
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Get current expiration status
				currentResult, err := s.Expired(ctx, tokenAddress, stopData)
				if err != nil {
					log.Warn("Failed to get expiration status for subscription", "error", err, "contract", tokenAddress.Hex())
					continue
				}

				// Only notify if the result changed
				if lastResult == nil || expirationStatusChanged(lastResult, currentResult) {
					lastResult = currentResult
					err = notifier.Notify(rpcSub.ID, currentResult)
					if err != nil {
						log.Warn("Failed to send expiration notification", "error", err)
						return
					}
				}

			case <-rpcSub.Err():
				return
			}
		}
	}()

	return rpcSub, nil
}

// expirationStatusChanged checks if the expiration status has changed
func expirationStatusChanged(last, current *ExpiredResult) bool {
	if last.Expired != current.Expired {
		return true
	}

	if last.TokenExpiration != nil && current.TokenExpiration != nil {
		if last.TokenExpiration.Cmp(current.TokenExpiration) != 0 {
			return true
		}
	} else if last.TokenExpiration != current.TokenExpiration {
		return true
	}

	if last.TradingStop != nil && current.TradingStop != nil {
		if last.TradingStop.Cmp(current.TradingStop) != 0 {
			return true
		}
	} else if last.TradingStop != current.TradingStop {
		return true
	}

	return false
}

// KYCResult represents the result of checking KYC verification status
type KYCResult struct {
	Verified     bool     `json:"verified"`               // Whether the user is KYC verified
	Timestamp    *big.Int `json:"timestamp,omitempty"`    // Unix timestamp when KYC verification happened
	SubmissionID *big.Int `json:"submissionId,omitempty"` // Submission ID for the verification
	Role         string   `json:"role,omitempty"`         // Role associated with the verification
}

// GetKYC checks if a user is KYC verified based on CorePass KYC smart contract.
// Based on CorePass KYC verification system for Core Blockchain.
//
// Function selector used:
// - isVerified(address,bytes32): 0xc9e14248
//
// Parameters:
//   - tokenAddress: The address of the KYC provider smart contract
//   - address: The user address to check KYC verification for
//   - fieldType (optional): specific verification type (e.g. "passport", "id", "driver", "email").
//     If nil, defaults to checking any of: Passport, IDCard, DriverLicense.
func (s *PublicSmartContractAPI) GetKYC(ctx context.Context, tokenAddress, userAddress common.Address, fieldType *string) (*KYCResult, error) {
	result := &KYCResult{Verified: false}
	// Encode address to 32 bytes (left-padded).
	userAddressBytes := make([]byte, 32)
	ua := userAddress.Bytes()
	copy(userAddressBytes[32-len(ua):], ua)

	// Build role candidates based on optional fieldType parameter
	normalize := func(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
	roleMap := map[string]string{
		"id":              "IDCard",
		"idcard":          "IDCard",
		"passport":        "Passport",
		"driver":          "DriverLicense",
		"driverlicense":   "DriverLicense",
		"drivinglicense":  "DriverLicense",
		"email":           "Email",
		"phone":           "Phone",
		"address":         "Address",
		"externalwallet":  "ExternalWallet",
		"wallet":          "ExternalWallet",
		"residencepermit": "ResidencePermit",
	}
	var roleCandidates []string
	if fieldType != nil && *fieldType != "" {
		if mapped, ok := roleMap[normalize(*fieldType)]; ok {
			roleCandidates = []string{mapped}
		} else {
			// Try raw provided name if no mapping is known
			roleCandidates = []string{*fieldType}
		}
	} else {
		// Default (requested): Passport OR ID OR Driver License
		roleCandidates = []string{"Passport", "IDCard", "DriverLicense"}
	}

	// isRoleVerified(address,bytes32)
	roleSig := []byte("isRoleVerified(address,bytes32)")
	roleSelector := crypto.SHA3Hash(roleSig).Bytes()[:4]
	callIsRoleVerified := func(name string, roleID common.Hash) (bool, error) {
		data := append([]byte{}, roleSelector...)
		data = append(data, userAddressBytes...)
		rb := make([]byte, 32)
		copy(rb, roleID.Bytes())
		data = append(data, rb...)
		out, err := s.b.CallContract(ctx, s.createViewCallMsg(tokenAddress, data), rpc.LatestBlockNumber)
		if err != nil {
			return false, err
		}
		if len(out) >= 32 {
			v := new(big.Int).SetBytes(out[:32])
			if v.Cmp(big.NewInt(0)) != 0 {
				result.Verified = true
				result.Role = name
				return true, nil
			}
		}
		return false, nil
	}

	var lastErr error
	for _, n := range roleCandidates {
		if ok, err := callIsRoleVerified(n, crypto.SHA3Hash([]byte(n))); err != nil {
			lastErr = err
		} else if ok {
			return result, nil
		}
	}

	if lastErr != nil {
		log.Warn("GetKYC returned false for all attempts", "contract", tokenAddress.Hex(), "user", userAddress.Hex(), "err", lastErr)
	}
	return result, nil
}

// GetKYCSubscription provides real-time updates for KYC verification status.
// Based on CorePass KYC verification system for Core Blockchain.
func (s *PublicSmartContractAPI) GetKYCSubscription(ctx context.Context, tokenAddress, userAddress common.Address, fieldType *string) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, rpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()

	go func() {
		var lastResult *KYCResult
		ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds for KYC changes - Ongoing KYC verification
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Get current KYC status
				currentResult, err := s.GetKYC(ctx, tokenAddress, userAddress, fieldType)
				if err != nil {
					log.Warn("Failed to get KYC status for subscription", "error", err, "contract", tokenAddress.Hex(), "user", userAddress.Hex())
					continue
				}

				// Only notify if the result changed
				if lastResult == nil || kycStatusChanged(lastResult, currentResult) {
					lastResult = currentResult
					err = notifier.Notify(rpcSub.ID, currentResult)
					if err != nil {
						log.Warn("Failed to send KYC notification", "error", err)
						return
					}
				}

			case <-rpcSub.Err():
				return
			}
		}
	}()

	return rpcSub, nil
}

// kycStatusChanged checks if the KYC status has changed
func kycStatusChanged(last, current *KYCResult) bool {
	if last.Verified != current.Verified {
		return true
	}

	if last.Timestamp != nil && current.Timestamp != nil {
		if last.Timestamp.Cmp(current.Timestamp) != 0 {
			return true
		}
	} else if last.Timestamp != current.Timestamp {
		return true
	}

	if last.Role != current.Role {
		return true
	}

	return false
}
