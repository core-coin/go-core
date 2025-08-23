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

package scclient

import (
	"context"
	"math/big"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/rpc"
)

// Client represents a smart contracts client that can interact with the sc namespace.
type Client struct {
	c *rpc.Client
}

// NewClient creates a new smart contracts client.
func NewClient(c *rpc.Client) *Client {
	return &Client{c: c}
}

// Symbol returns the symbol of a token contract by calling the sc.symbol RPC method.
func (sc *Client) Symbol(ctx context.Context, tokenAddress common.Address) (string, error) {
	var result string
	err := sc.c.CallContext(ctx, &result, "sc_symbol", tokenAddress)
	return result, err
}

// Name returns the name of a token contract by calling the sc.name RPC method.
func (sc *Client) Name(ctx context.Context, tokenAddress common.Address) (string, error) {
	var result string
	err := sc.c.CallContext(ctx, &result, "sc_name", tokenAddress)
	return result, err
}

// BalanceOf returns the token balance of a specific address for a given token contract.
func (sc *Client) BalanceOf(ctx context.Context, holderAddress, tokenAddress common.Address) (*big.Int, error) {
	var result hexutil.Big
	err := sc.c.CallContext(ctx, &result, "sc_balanceOf", holderAddress, tokenAddress)
	if err != nil {
		return nil, err
	}
	return (*big.Int)(&result), nil
}

// Decimals returns the number of decimal places for a given token contract.
func (sc *Client) Decimals(ctx context.Context, tokenAddress common.Address) (uint8, error) {
	var result hexutil.Uint64
	err := sc.c.CallContext(ctx, &result, "sc_decimals", tokenAddress)
	if err != nil {
		return 0, err
	}
	return uint8(result), nil
}

// TotalSupply returns the total supply of a given token contract.
func (sc *Client) TotalSupply(ctx context.Context, tokenAddress common.Address) (*big.Int, error) {
	var result hexutil.Big
	err := sc.c.CallContext(ctx, &result, "sc_totalSupply", tokenAddress)
	if err != nil {
		return nil, err
	}
	return (*big.Int)(&result), nil
}

// Length returns the smart contract code size in bytes for a given contract address.
func (sc *Client) Length(ctx context.Context, tokenAddress common.Address) (uint64, error) {
	var result hexutil.Uint64
	err := sc.c.CallContext(ctx, &result, "sc_length", tokenAddress)
	if err != nil {
		return 0, err
	}
	return uint64(result), nil
}

// GetKVResult represents the result of a getKV call
type GetKVResult struct {
	Value  string `json:"value"`  // The metadata value
	Sealed bool   `json:"sealed"` // Whether the key is sealed (immutable)
	Exists bool   `json:"exists"` // Whether the key exists
}

// ListKVResult represents the result of a listKV call
type ListKVResult struct {
	Keys   []string `json:"keys"`   // List of all keys
	Count  uint64   `json:"count"`  // Total number of keys
	Sealed []bool   `json:"sealed"` // Sealed status for each key (if sealed=false)
	Values []string `json:"values"` // Values for each key (if sealed=false)
}

// GetKV retrieves key-value metadata from a smart contract implementing CIP-150.
// Based on the CIP-150 standard for On-Chain Key-Value Metadata Storage.
// If sealed=false (default), returns the value and sealed status.
// If sealed=true, only returns data if the item is actually sealed.
func (sc *Client) GetKV(ctx context.Context, key string, tokenAddress common.Address, sealed bool) (*GetKVResult, error) {
	var result GetKVResult
	err := sc.c.CallContext(ctx, &result, "sc_getKV", key, tokenAddress, sealed)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListKV retrieves all keys from a smart contract implementing CIP-150.
// Based on the CIP-150 standard for On-Chain Key-Value Metadata Storage.
// If sealed=false (default), returns keys, sealed status, and values.
// If sealed=true, only returns keys that are sealed.
func (sc *Client) ListKV(ctx context.Context, tokenAddress common.Address, sealed bool) (*ListKVResult, error) {
	var result ListKVResult
	err := sc.c.CallContext(ctx, &result, "sc_listKV", tokenAddress, sealed)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// TokenURI retrieves the token URI for a specific NFT token ID from a CoreNFT contract.
// Based on the CIP-721 standard for Core Blockchain Non-Fungible Tokens.
func (sc *Client) TokenURI(ctx context.Context, tokenAddress common.Address, tokenId *big.Int) (string, error) {
	var result string
	err := sc.c.CallContext(ctx, &result, "sc_tokenURI", tokenAddress, tokenId)
	if err != nil {
		return "", err
	}
	return result, nil
}

// SymbolSubscription subscribes to real-time updates about token symbols.
// It returns a subscription that will notify when the symbol changes.
func (sc *Client) SymbolSubscription(ctx context.Context, tokenAddress common.Address) (Subscription, error) {
	return sc.c.XcbSubscribe(ctx, make(chan string), "symbol", tokenAddress)
}

// NameSubscription subscribes to real-time updates about token names.
// It returns a subscription that will notify when the name changes.
func (sc *Client) NameSubscription(ctx context.Context, tokenAddress common.Address) (Subscription, error) {
	return sc.c.XcbSubscribe(ctx, make(chan string), "name", tokenAddress)
}

// Dial connects to a smart contracts RPC endpoint.
func Dial(rawurl string) (*Client, error) {
	c, err := rpc.Dial(rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// DialContext connects to a smart contracts RPC endpoint with context.
func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// Close closes the underlying RPC client.
func (sc *Client) Close() {
	sc.c.Close()
}

// Subscription represents an event subscription where events are
// delivered on a data channel.
type Subscription interface {
	// Unsubscribe cancels the sending of events to the data channel
	// and closes the error channel.
	Unsubscribe()
	// Err returns the subscription error channel. The error channel receives
	// a value if there is an issue with the subscription (e.g. the network connection
	// delivering the events has been closed). Only one value will ever be sent.
	// The error channel is closed by Unsubscribe.
	Err() <-chan error
}
