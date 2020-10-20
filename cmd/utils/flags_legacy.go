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

package utils

import (
	"fmt"
	"strings"

	"github.com/core-coin/go-core/v2/node"
	"github.com/core-coin/go-core/v2/xcb"
	"gopkg.in/urfave/cli.v1"
)

var ShowDeprecated = cli.Command{
	Action:      showDeprecated,
	Name:        "show-deprecated-flags",
	Usage:       "Show flags that have been deprecated",
	ArgsUsage:   " ",
	Category:    "MISCELLANEOUS COMMANDS",
	Description: "Show flags that have been deprecated and will soon be removed",
}

var DeprecatedFlags = []cli.Flag{
	LegacyLightPeersFlag,
	LegacyMinerThreadsFlag,
	LegacyMinerEnergyTargetFlag,
	LegacyMinerEnergyPriceFlag,
	LegacyMinerCorebaseFlag,
	LegacyMinerExtraDataFlag,
}

var (
	// (Deprecated April 2018)
	LegacyMinerThreadsFlag = cli.IntFlag{
		Name:  "minerthreads",
		Usage: "Number of CPU threads to use for mining (deprecated, use --miner.threads)",
		Value: 0,
	}
	LegacyMinerEnergyTargetFlag = cli.Uint64Flag{
		Name:  "targetenergylimit",
		Usage: "Target energy floor for mined blocks (deprecated, use --miner.energytarget)",
		Value: xcb.DefaultConfig.Miner.EnergyFloor,
	}
	LegacyMinerEnergyPriceFlag = BigFlag{
		Name:  "energyprice",
		Usage: "Minimum energy price for mining a transaction (deprecated, use --miner.energyprice)",
		Value: xcb.DefaultConfig.Miner.EnergyPrice,
	}
	LegacyMinerCorebaseFlag = cli.StringFlag{
		Name:  "corebase",
		Usage: "Public address for block mining rewards (default = first account, deprecated, use --miner.corebase)",
		Value: "0",
	}
	LegacyMinerExtraDataFlag = cli.StringFlag{
		Name:  "extradata",
		Usage: "Block extra data set by the miner (default = client version, deprecated, use --miner.extradata)",
	}

	LegacyLightPeersFlag = cli.IntFlag{
		Name:  "lightpeers",
		Usage: "Maximum number of light clients to serve, or light servers to attach to  (deprecated, use --light.maxpeers)",
		Value: xcb.DefaultConfig.LightPeers,
	}

	LegacyRPCEnabledFlag = cli.BoolFlag{
		Name:  "rpc",
		Usage: "Enable the HTTP-RPC server (deprecated, use --http)",
	}
	LegacyRPCListenAddrFlag = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface (deprecated, use --http.addr)",
		Value: node.DefaultHTTPHost,
	}
	LegacyRPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port (deprecated, use --http.port)",
		Value: node.DefaultHTTPPort,
	}
	LegacyRPCCORSDomainFlag = cli.StringFlag{
		Name:  "rpccorsdomain",
		Usage: "Comma separated list of domains from which to accept cross origin requests (browser enforced) (deprecated, use --http.corsdomain)",
		Value: "",
	}
	LegacyRPCVirtualHostsFlag = cli.StringFlag{
		Name:  "rpcvhosts",
		Usage: "Comma separated list of virtual hostnames from which to accept requests (server enforced). Accepts '*' wildcard. (deprecated, use --http.vhosts)",
		Value: strings.Join(node.DefaultConfig.HTTPVirtualHosts, ","),
	}
	LegacyRPCApiFlag = cli.StringFlag{
		Name:  "rpcapi",
		Usage: "API's offered over the HTTP-RPC interface (deprecated, use --http.api)",
		Value: "",
	}
	LegacyWSListenAddrFlag = cli.StringFlag{
		Name:  "wsaddr",
		Usage: "WS-RPC server listening interface (deprecated, use --ws.addr)",
		Value: node.DefaultWSHost,
	}
	LegacyWSPortFlag = cli.IntFlag{
		Name:  "wsport",
		Usage: "WS-RPC server listening port (deprecated, use --ws.port)",
		Value: node.DefaultWSPort,
	}
	LegacyWSApiFlag = cli.StringFlag{
		Name:  "wsapi",
		Usage: "API's offered over the WS-RPC interface (deprecated, use --ws.api)",
		Value: "",
	}
	LegacyWSAllowedOriginsFlag = cli.StringFlag{
		Name:  "wsorigins",
		Usage: "Origins from which to accept websockets requests (deprecated, use --ws.origins)",
		Value: "",
	}
	LegacyGpoBlocksFlag = cli.IntFlag{
		Name:  "gpoblocks",
		Usage: "Number of recent blocks to check for energy prices (deprecated, use --gpo.blocks)",
		Value: xcb.DefaultConfig.GPO.Blocks,
	}
	LegacyGpoPercentileFlag = cli.IntFlag{
		Name:  "gpopercentile",
		Usage: "Suggested energy price is the given percentile of a set of recent transaction energy prices (deprecated, use --gpo.percentile)",
		Value: xcb.DefaultConfig.GPO.Percentile,
	}
)

// showDeprecated displays deprecated flags that will be soon removed from the codebase.
func showDeprecated(*cli.Context) {
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("The following flags are deprecated and will be removed in the future!")
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println()

	for _, flag := range DeprecatedFlags {
		fmt.Println(flag.String())
	}
}
