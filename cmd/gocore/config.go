// Copyright 2017 by the Authors
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
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"unicode"

	"gopkg.in/urfave/cli.v1"

	"github.com/core-coin/go-core/cmd/utils"
	"github.com/core-coin/go-core/node"
	"github.com/core-coin/go-core/params"
	"github.com/core-coin/go-core/xcb"
	"github.com/naoina/toml"
)

var (
	dumpConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(nodeFlags, rpcFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

type xcbstatsConfig struct {
	URL string `toml:",omitempty"`
}

type gocoreConfig struct {
	Xcb      xcb.Config
	Node     node.Config
	Xcbstats xcbstatsConfig
}

func loadConfig(file string, cfg *gocoreConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit, gitDate)
	cfg.HTTPModules = append(cfg.HTTPModules, "xcb")
	cfg.WSModules = append(cfg.WSModules, "xcb")
	cfg.IPCPath = "gocore.ipc"
	return cfg
}

func makeConfigNode(ctx *cli.Context) (*node.Node, gocoreConfig) {
	// Load defaults.
	cfg := gocoreConfig{
		Xcb:  xcb.DefaultConfig,
		Node: defaultNodeConfig(),
	}

	// Load config file.
	if file := ctx.GlobalString(configFileFlag.Name); file != "" {
		if err := loadConfig(file, &cfg); err != nil {
			utils.Fatalf("%v", err)
		}
	}

	// Apply flags.
	utils.SetNodeConfig(ctx, &cfg.Node)

	// Set default network id from --networkId, --devin, --koliba flags
	utils.SetDefaultNetworkID(ctx)

	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	utils.SetXcbConfig(ctx, stack, &cfg.Xcb)
	if ctx.GlobalIsSet(utils.XcbStatsURLFlag.Name) {
		cfg.Xcbstats.URL = ctx.GlobalString(utils.XcbStatsURLFlag.Name)
	}

	return stack, cfg
}

func makeFullNode(ctx *cli.Context) *node.Node {
	stack, cfg := makeConfigNode(ctx)
	utils.RegisterXcbService(stack, &cfg.Xcb)

	// Configure GraphQL if requested
	if ctx.GlobalIsSet(utils.GraphQLEnabledFlag.Name) {
		utils.RegisterGraphQLService(stack, cfg.Node.GraphQLEndpoint(), cfg.Node.GraphQLCors, cfg.Node.GraphQLVirtualHosts, cfg.Node.HTTPTimeouts)
	}
	// Add the Core Stats daemon if requested.
	if cfg.Xcbstats.URL != "" {
		utils.RegisterXcbStatsService(stack, cfg.Xcbstats.URL)
	}
	return stack
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	_, cfg := makeConfigNode(ctx)
	comment := ""

	if cfg.Xcb.Genesis != nil {
		cfg.Xcb.Genesis = nil
		comment += "# Note: this config doesn't contain the genesis block.\n\n"
	}

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}

	dump := os.Stdout
	if ctx.NArg() > 0 {
		dump, err = os.OpenFile(ctx.Args().Get(0), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer dump.Close()
	}
	dump.WriteString(comment)
	dump.Write(out)

	return nil
}
