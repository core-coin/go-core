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
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/core/state"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/internal/xcbapi"
	"github.com/core-coin/go-core/v2/rpc"
)

// PublicSmartContractAPI provides an API to access smart contract related information.
// It offers methods that operate on smart contract data that can be available to anyone.
type PublicSmartContractAPI struct {
	b Backend
}

// NewPublicSmartContractAPI creates a new smart contract protocol API.
func NewPublicSmartContractAPI(b Backend) *PublicSmartContractAPI {
	return &PublicSmartContractAPI{b}
}

// Symbol returns the symbol of a token contract by calling the symbol() function.
// It automatically decodes the dynamic string response using decodeDynString.
func (s *PublicSmartContractAPI) Symbol(ctx context.Context, tokenAddress common.Address) (string, error) {
	// CBC20 symbol() function selector: 0x231782d8
	selector := "0x231782d8" // standard CBC20 symbol()

	// Create the call data
	data := hexutil.MustDecode(selector)

	// Make the contract call
	result, err := s.b.CallContract(ctx, xcbapi.CallMsg{
		ToAddr:    &tokenAddress,
		DataBytes: data,
	}, rpc.LatestBlockNumber)

	if err != nil {
		return "", fmt.Errorf("failed to call symbol() on contract %s: %v", tokenAddress.Hex(), err)
	}

	// If we got a result, decode it
	if len(result) > 0 {
		decoded, err := decodeDynString(hexutil.Encode(result))
		if err != nil {
			return "", fmt.Errorf("failed to decode symbol response from contract %s: %v", tokenAddress.Hex(), err)
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

	// Make the contract call
	result, err := s.b.CallContract(ctx, xcbapi.CallMsg{
		ToAddr:    &tokenAddress,
		DataBytes: data,
	}, rpc.LatestBlockNumber)

	if err != nil {
		return "", fmt.Errorf("failed to call name() on contract %s: %v", tokenAddress.Hex(), err)
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

// BalanceOf returns the token balance of a specific address for a given token contract.
// It automatically decodes the uint256 response and converts it to a big.Int.
func (s *PublicSmartContractAPI) BalanceOf(ctx context.Context, holderAddress, tokenAddress common.Address) (*big.Int, error) {
	// CBC20 balanceOf(address) function selector: 0x1d7976f3
	selector := "0x1d7976f3" // standard CBC20 balanceOf(address)

	// Create the call data: selector + padded address (32 bytes)
	data := hexutil.MustDecode(selector)

	// Pad the holder address to 32 bytes (left-pad with zeros)
	addressBytes := holderAddress.Bytes()
	paddedAddress := make([]byte, 32)
	copy(paddedAddress[32-len(addressBytes):], addressBytes)

	// Append the padded address to the selector
	data = append(data, paddedAddress...)

	// Make the contract call
	result, err := s.b.CallContract(ctx, xcbapi.CallMsg{
		ToAddr:    &tokenAddress,
		DataBytes: data,
	}, rpc.LatestBlockNumber)

	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf on contract %s for address %s: %v", tokenAddress.Hex(), holderAddress.Hex(), err)
	}

	// If we got a result, decode it as uint256
	if len(result) > 0 {
		// Convert the 32-byte result to big.Int
		balance := new(big.Int).SetBytes(result)
		return balance, nil
	}

	return nil, fmt.Errorf("empty response from balanceOf call on contract %s for address %s", tokenAddress.Hex(), holderAddress.Hex())
}

// Decimals returns the number of decimal places for a given token contract.
// It automatically decodes the uint8 response and converts it to uint8.
func (s *PublicSmartContractAPI) Decimals(ctx context.Context, tokenAddress common.Address) (uint8, error) {
	// CBC20 decimals() function selector: 0x5d1fb5f9
	selector := "0x5d1fb5f9" // standard CBC20 decimals()

	// Create the call data
	data := hexutil.MustDecode(selector)

	// Make the contract call
	result, err := s.b.CallContract(ctx, xcbapi.CallMsg{
		ToAddr:    &tokenAddress,
		DataBytes: data,
	}, rpc.LatestBlockNumber)

	if err != nil {
		return 0, fmt.Errorf("failed to call decimals() on contract %s: %v", tokenAddress.Hex(), err)
	}

	// If we got a result, decode it as uint8
	if len(result) > 0 {
		// Convert the 32-byte result to uint8 (last byte)
		if len(result) >= 32 {
			decimals := result[31] // Last byte contains the uint8 value
			return decimals, nil
		}
	}

	return 0, fmt.Errorf("invalid response length from decimals() call on contract %s", tokenAddress.Hex())
}

// TotalSupply returns the total supply of a given token contract.
// It automatically decodes the uint256 response and converts it to a big.Int.
func (s *PublicSmartContractAPI) TotalSupply(ctx context.Context, tokenAddress common.Address) (*big.Int, error) {
	// CBC20 totalSupply() function selector: 0x1f1881f8
	selector := "0x1f1881f8" // standard CBC20 totalSupply()

	// Create the call data
	data := hexutil.MustDecode(selector)

	// Make the contract call
	result, err := s.b.CallContract(ctx, xcbapi.CallMsg{
		ToAddr:    &tokenAddress,
		DataBytes: data,
	}, rpc.LatestBlockNumber)

	if err != nil {
		return nil, fmt.Errorf("failed to call totalSupply() on contract %s: %v", tokenAddress.Hex(), err)
	}

	// If we got a result, decode it as uint256
	if len(result) > 0 {
		// Convert the 32-byte result to big.Int
		supply := new(big.Int).SetBytes(result)
		return supply, nil
	}

	return nil, fmt.Errorf("empty response from totalSupply() call on contract %s", tokenAddress.Hex())
}

// Length returns the smart contract code size in bytes for a given contract address.
// It uses xcb.getCode to fetch the contract code and returns the size in bytes.
func (s *PublicSmartContractAPI) Length(ctx context.Context, tokenAddress common.Address) (uint64, error) {
	// Since xcb.getCode is already implemented as a standard RPC method,
	// we can use the existing blockchain API directly
	// This is equivalent to calling xcb.getCode(addr, "latest")

	// Get the latest block number
	latestBlock := rpc.LatestBlockNumber

	// Get the state at the latest block
	state, _, err := s.b.StateAndHeaderByNumber(ctx, latestBlock)
	if err != nil {
		return 0, fmt.Errorf("failed to get state for contract %s: %v", tokenAddress.Hex(), err)
	}

	// Get the code from the state
	code := state.GetCode(tokenAddress)

	// Check if it's an EOA (no code) or empty code
	if len(code) == 0 {
		return 0, nil
	}

	// Return the code size in bytes
	return uint64(len(code)), nil
}

// SymbolSubscription provides real-time updates about token symbols.
// This can be useful for monitoring token metadata changes.
func (s *PublicSmartContractAPI) SymbolSubscription(ctx context.Context, tokenAddress common.Address) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()

	go func() {
		// Send initial symbol
		symbol, err := s.Symbol(ctx, tokenAddress)
		if err == nil {
			notifier.Notify(rpcSub.ID, symbol)
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

// decodeDynString decodes a dynamic string from contract call response.
// It handles both bytes32 fallback and dynamic string encoding.
func decodeDynString(res string) (string, error) {
	// Remove 0x prefix
	h := strings.TrimPrefix(res, "0x")

	// Check if it's a bytes32 fallback (64 hex chars = 32 bytes)
	if len(h) == 64 {
		// Decode as bytes32 and convert to ASCII
		bytes, err := hex.DecodeString(h)
		if err != nil {
			return "", err
		}

		// Convert to string and remove null bytes
		result := string(bytes)
		return strings.TrimRight(result, "\x00"), nil
	}

	// Handle dynamic string encoding
	if len(h) < 64 {
		return "", fmt.Errorf("response too short for dynamic string")
	}

	// Parse offset (first 32 bytes)
	offsetHex := h[:64]
	offset, err := hex.DecodeString(offsetHex)
	if err != nil {
		return "", err
	}

	// Convert offset to int (big-endian)
	offsetInt := new(big.Int).SetBytes(offset)
	offsetBytes := offsetInt.Int64() * 2 // Convert to hex string position

	// Check if offset is 0 (simple bytes32 case) or if it's a valid offset
	if offsetInt.Cmp(big.NewInt(0)) == 0 {
		// This is a simple bytes32 case, decode the entire response as bytes32
		bytes, err := hex.DecodeString(h)
		if err != nil {
			return "", err
		}
		result := string(bytes)
		return strings.TrimRight(result, "\x00"), nil
	}

	// Check if we have enough data
	if int64(len(h)) < offsetBytes+64 {
		return "", fmt.Errorf("response too short for offset %d", offsetBytes)
	}

	// Parse length (32 bytes after offset)
	lengthHex := h[offsetBytes : offsetBytes+64]
	length, err := hex.DecodeString(lengthHex)
	if err != nil {
		return "", err
	}

	// Convert length to int (big-endian)
	lengthInt := new(big.Int).SetBytes(length)
	lengthBytes := lengthInt.Int64() * 2 // Convert to hex string position

	// Extract the actual string data
	dataStart := offsetBytes + 64
	dataEnd := dataStart + lengthBytes

	if int64(len(h)) < dataEnd {
		return "", fmt.Errorf("response too short for data length %d", lengthBytes)
	}

	dataHex := h[dataStart:dataEnd]

	// Decode the hex data to bytes
	dataBytes, err := hex.DecodeString(dataHex)
	if err != nil {
		return "", err
	}

	// Convert to UTF-8 string
	return string(dataBytes), nil
}

// Backend interface provides the common API services needed for smart contract operations.
type Backend interface {
	// CallContract executes a contract call
	CallContract(ctx context.Context, call xcbapi.CallMsg, blockNumber rpc.BlockNumber) ([]byte, error)
	// StateAndHeaderByNumber gets the state and header at a specific block number
	StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error)
}
