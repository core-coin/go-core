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
