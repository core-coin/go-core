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
	"crypto/rand"
	"fmt"
	"net"

	"github.com/core-coin/go-core/common/hexutil"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/p2p"
	"github.com/core-coin/go-core/p2p/rlpx"
	"github.com/core-coin/go-core/rlp"
	"gopkg.in/urfave/cli.v1"
)

var (
	rlpxCommand = cli.Command{
		Name:  "rlpx",
		Usage: "RLPx Commands",
		Subcommands: []cli.Command{
			rlpxPingCommand,
		},
	}
	rlpxPingCommand = cli.Command{
		Name:      "ping",
		Usage:     "Perform a RLPx handshake",
		ArgsUsage: "<node>",
		Action:    rlpxPing,
	}
)

func rlpxPing(ctx *cli.Context) error {
	n := getNodeArg(ctx)

	fd, err := net.Dial("tcp", fmt.Sprintf("%v:%d", n.IP(), n.TCP()))
	if err != nil {
		return err
	}
	conn := rlpx.NewConn(fd, n.Pubkey())

	ourKey, _ := crypto.GenerateKey(rand.Reader)
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
		var h devp2pHandshake
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

// devp2pHandshake is the RLP structure of the devp2p protocol handshake.
type devp2pHandshake struct {
	Version    uint64
	Name       string
	Caps       []p2p.Cap
	ListenPort uint64
	ID         hexutil.Bytes // secp256k1 public key
	// Ignore additional fields (for forward compatibility).
	Rest []rlp.RawValue `rlp:"tail"`
}