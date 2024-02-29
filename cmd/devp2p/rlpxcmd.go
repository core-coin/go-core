// Copyright 2020 by the Authors
// This file is part of go-core.
//
// go-core is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-core is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-core. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	crand "crypto/rand"
	"fmt"
	"net"

	"gopkg.in/urfave/cli.v1"

	xcbtest "github.com/core-coin/go-core/v2/cmd/devp2p/internal/xcbtest"
	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/p2p"
	"github.com/core-coin/go-core/v2/p2p/rlpx"
	"github.com/core-coin/go-core/v2/rlp"
)

var (
	rlpxCommand = cli.Command{
		Name:  "rlpx",
		Usage: "RLPx Commands",
		Subcommands: []cli.Command{
			rlpxPingCommand,
			rlpxXcbTestCommand,
		},
	}
	rlpxPingCommand = cli.Command{
		Name:   "ping",
		Usage:  "ping <node>",
		Action: rlpxPing,
	}
	rlpxXcbTestCommand = cli.Command{
		Name:      "xcb-test",
		Usage:     "Runs tests against a node",
		ArgsUsage: "<node> <chain.rlp> <genesis.json>",
		Action:    rlpxXcbTest,
		Flags: []cli.Flag{
			testPatternFlag,
			testTAPFlag,
		},
	}
)

func rlpxPing(ctx *cli.Context) error {
	n := getNodeArg(ctx)
	fd, err := net.Dial("tcp", fmt.Sprintf("%v:%d", n.IP(), n.TCP()))
	if err != nil {
		return err
	}
	conn := rlpx.NewConn(fd, n.Pubkey())
	ourKey, _ := crypto.GenerateKey(crand.Reader)
	_, err = conn.Handshake(ourKey)
	if err != nil {
		return err
	}
	code, data, _, err := conn.Read()
	if err != nil {
		return err
	}
	switch code {
	case 0:
		var h xcbtest.Hello
		if err := rlp.DecodeBytes(data, &h); err != nil {
			return fmt.Errorf("invalid handshake: %v", err)
		}
		fmt.Printf("%+v\n", h)
	case 1:
		var msg []p2p.DiscReason
		if rlp.DecodeBytes(data, &msg); len(msg) == 0 {
			return fmt.Errorf("invalid disconnect message")
		}
		return fmt.Errorf("received disconnect message: %v", msg[0])
	default:
		return fmt.Errorf("invalid message code %d, expected handshake (code zero)", code)
	}
	return nil
}

// rlpxXcbTest runs the xcb protocol test suite.
func rlpxXcbTest(ctx *cli.Context) error {
	if ctx.NArg() < 3 {
		exit("missing path to chain.rlp as command-line argument")
	}
	common.DefaultNetworkID = 19763
	suite := xcbtest.NewSuite(getNodeArg(ctx), ctx.Args()[1], ctx.Args()[2])
	return runTests(ctx, suite.AllTests())
}
