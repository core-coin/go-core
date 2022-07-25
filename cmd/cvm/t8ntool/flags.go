// Copyright 2022 by the Authors
// This file is part of the go-core library.
//
// The go-core library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-core library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-core library. If not, see <http://www.gnu.org/licenses/>.

package t8ntool

import (
	"gopkg.in/urfave/cli.v1"
)

var (
	TraceFlag = cli.BoolFlag{
		Name:  "trace",
		Usage: "Output full trace logs to files <txhash>.jsonl",
	}
	TraceDisableMemoryFlag = cli.BoolFlag{
		Name:  "trace.nomemory",
		Usage: "Disable full memory dump in traces",
	}
	TraceDisableStackFlag = cli.BoolFlag{
		Name:  "trace.nostack",
		Usage: "Disable stack output in traces",
	}
	OutputAllocFlag = cli.StringFlag{
		Name: "output.alloc",
		Usage: "Determines where to put the `alloc` of the post-state.\n" +
			"\t`stdout` - into the stdout output\n" +
			"\t`stderr` - into the stderr output\n" +
			"\t<file> - into the file <file> ",
		Value: "alloc.json",
	}
	OutputResultFlag = cli.StringFlag{
		Name: "output.result",
		Usage: "Determines where to put the `result` (stateroot, txroot etc) of the post-state.\n" +
			"\t`stdout` - into the stdout output\n" +
			"\t`stderr` - into the stderr output\n" +
			"\t<file> - into the file <file> ",
		Value: "result.json",
	}
	InputAllocFlag = cli.StringFlag{
		Name:  "input.alloc",
		Usage: "`stdin` or file name of where to find the prestate alloc to use.",
		Value: "alloc.json",
	}
	InputEnvFlag = cli.StringFlag{
		Name:  "input.env",
		Usage: "`stdin` or file name of where to find the prestate env to use.",
		Value: "env.json",
	}
	InputTxsFlag = cli.StringFlag{
		Name:  "input.txs",
		Usage: "`stdin` or file name of where to find the transactions to apply.",
		Value: "txs.json",
	}
	RewardFlag = cli.Int64Flag{
		Name:  "state.reward",
		Usage: "Mining reward. Set to -1 to disable",
		Value: 0,
	}
	NetworkIDFlag = cli.Int64Flag{
		Name:  "state.networkid",
		Usage: "NetworkID to use",
		Value: 1,
	}
	VerbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Usage: "sets the verbosity level",
		Value: 3,
	}
)
