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

// BalanceOf returns the balance of a given token holder for a given token contract.
func (sc *Client) BalanceOf(ctx context.Context, holderAddress, tokenAddress common.Address) (*hexutil.Big, error) {
	var result hexutil.Big
	err := sc.c.CallContext(ctx, &result, "sc_balanceOf", holderAddress, tokenAddress)
	return &result, err
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
func (sc *Client) TotalSupply(ctx context.Context, tokenAddress common.Address) (*hexutil.Big, error) {
	var result hexutil.Big
	err := sc.c.CallContext(ctx, &result, "sc_totalSupply", tokenAddress)
	return &result, err
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

// BalanceOfSubscription subscribes to real-time updates about token balances.
// It returns a subscription that will notify when the balance changes.
func (sc *Client) BalanceOfSubscription(ctx context.Context, holderAddress, tokenAddress common.Address) (Subscription, error) {
	return sc.c.XcbSubscribe(ctx, make(chan *big.Int), "balanceOf", holderAddress, tokenAddress)
}

// DecimalsSubscription subscribes to real-time updates about token decimal places.
// It returns a subscription that will notify when the decimals change.
func (sc *Client) DecimalsSubscription(ctx context.Context, tokenAddress common.Address) (Subscription, error) {
	return sc.c.XcbSubscribe(ctx, make(chan uint8), "decimals", tokenAddress)
}

// TotalSupplySubscription subscribes to real-time updates about token total supply.
// It returns a subscription that will notify when the total supply changes.
func (sc *Client) TotalSupplySubscription(ctx context.Context, tokenAddress common.Address) (Subscription, error) {
	return sc.c.XcbSubscribe(ctx, make(chan *big.Int), "totalSupply", tokenAddress)
}

// LengthSubscription subscribes to real-time updates about smart contract code size.
// It returns a subscription that will notify when the contract length changes.
func (sc *Client) LengthSubscription(ctx context.Context, tokenAddress common.Address) (Subscription, error) {
	return sc.c.XcbSubscribe(ctx, make(chan uint64), "length", tokenAddress)
}

// GetKVSubscription subscribes to real-time updates about key-value metadata.
// It returns a subscription that will notify when the metadata changes.
func (sc *Client) GetKVSubscription(ctx context.Context, key string, tokenAddress common.Address, sealed bool) (Subscription, error) {
	return sc.c.XcbSubscribe(ctx, make(chan *GetKVResult), "getKV", key, tokenAddress, sealed)
}

// ListKVSubscription subscribes to real-time updates about key-value metadata lists.
// It returns a subscription that will notify when the metadata list changes.
func (sc *Client) ListKVSubscription(ctx context.Context, tokenAddress common.Address, sealed bool) (Subscription, error) {
	return sc.c.XcbSubscribe(ctx, make(chan *ListKVResult), "listKV", tokenAddress, sealed)
}

// TokenURISubscription subscribes to real-time updates about NFT token URIs.
// It returns a subscription that will notify when the token URI changes.
func (sc *Client) TokenURISubscription(ctx context.Context, tokenAddress common.Address, tokenId *big.Int) (Subscription, error) {
	return sc.c.XcbSubscribe(ctx, make(chan string), "tokenURI", tokenAddress, tokenId)
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

// GetPrice returns the latest price or aggregated price from a PriceFeed contract.
// Based on CIP-104 for Core Blockchain PriceFeed contracts.
func (c *Client) GetPrice(ctx context.Context, tokenAddress common.Address, aggregated bool) (*hexutil.Big, error) {
	var result hexutil.Big
	err := c.c.CallContext(ctx, &result, "sc_getPrice", tokenAddress, aggregated)
	return &result, err
}

// GetPriceSubscription provides real-time updates for the latest price from a PriceFeed contract.
// Based on CIP-104 for Core Blockchain PriceFeed contracts.
func (c *Client) GetPriceSubscription(ctx context.Context, tokenAddress common.Address, aggregated bool) (Subscription, error) {
	return c.c.XcbSubscribe(ctx, "sc_getPrice", tokenAddress, aggregated)
}

// ExpiredResult represents the result of checking token expiration status
type ExpiredResult struct {
	Expired         bool     `json:"expired"`                   // Whether the token is expired
	TokenExpiration *big.Int `json:"tokenExpiration,omitempty"` // Unix timestamp when token expires
	TradingStop     *big.Int `json:"tradingStop,omitempty"`     // Unix timestamp when trading should stop
}

// Expired checks if a token is expired based on CIP-151 Token Lifecycle Metadata Standard.
// Based on CIP-151 for Core Blockchain token lifecycle management.
func (c *Client) Expired(ctx context.Context, tokenAddress common.Address, stopData bool) (*ExpiredResult, error) {
	var result *ExpiredResult
	err := c.c.CallContext(ctx, &result, "sc_expired", tokenAddress, stopData)
	return result, err
}

// ExpiredSubscription provides real-time updates for token expiration status.
// Based on CIP-151 for Core Blockchain token lifecycle management.
func (c *Client) ExpiredSubscription(ctx context.Context, tokenAddress common.Address, stopData bool) (Subscription, error) {
	return c.c.XcbSubscribe(ctx, "sc_expired", tokenAddress, stopData)
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
func (c *Client) GetKYC(ctx context.Context, tokenAddress, userAddress common.Address) (*KYCResult, error) {
	var result *KYCResult
	err := c.c.CallContext(ctx, &result, "sc_getKYC", tokenAddress, userAddress)
	return result, err
}

// GetKYCSubscription provides real-time updates for KYC verification status.
// Based on CorePass KYC verification system for Core Blockchain.
func (c *Client) GetKYCSubscription(ctx context.Context, tokenAddress, userAddress common.Address) (Subscription, error) {
	return c.c.XcbSubscribe(ctx, "sc_getKYC", tokenAddress, userAddress)
}
