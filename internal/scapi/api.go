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
	// ERC20 symbol() function selector: 0x95d89b41
	// But some contracts use 0x231782d8, so we'll try both
	symbolSelectors := []string{
		"0x95d89b41", // standard ERC20 symbol()
		"0x231782d8", // alternative symbol()
	}

	for _, selector := range symbolSelectors {
		// Create the call data
		data := hexutil.MustDecode(selector)

		// Make the contract call
		result, err := s.b.CallContract(ctx, xcbapi.CallMsg{
			ToAddr:    &tokenAddress,
			DataBytes: data,
		}, rpc.LatestBlockNumber)

		if err != nil {
			continue // Try next selector
		}

		// If we got a result, decode it
		if len(result) > 0 {
			decoded, err := decodeDynString(hexutil.Encode(result))
			if err == nil && decoded != "" {
				return decoded, nil
			}
		}
	}

	return "", fmt.Errorf("failed to get symbol from contract %s", tokenAddress.Hex())
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
}
