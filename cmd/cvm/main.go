// Copyright 2014 by the Authors
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

// cvm executes CVM code snippets.
package main

import (
	"fmt"
	"math/big"
	"os"

	"gopkg.in/urfave/cli.v1"

	"github.com/core-coin/go-core/v2/cmd/cvm/internal/t8ntool"

	"github.com/core-coin/go-core/v2/cmd/utils"
	"github.com/core-coin/go-core/v2/internal/flags"
)

var gitTag = ""
var gitCommit = "" // Git SHA1 commit hash of the release (set via linker flags)
var gitDate = ""

var (
	app = flags.NewApp(gitTag, gitCommit, gitDate, "the cvm command line interface")

	DebugFlag = cli.BoolFlag{
		Name:  "debug",
		Usage: "output full trace logs",
	}
	MemProfileFlag = cli.StringFlag{
		Name:  "memprofile",
		Usage: "creates a memory profile at the given path",
	}
	CPUProfileFlag = cli.StringFlag{
		Name:  "cpuprofile",
		Usage: "creates a CPU profile at the given path",
	}
	StatDumpFlag = cli.BoolFlag{
		Name:  "statdump",
		Usage: "displays stack and heap memory information",
	}
	CodeFlag = cli.StringFlag{
		Name:  "code",
		Usage: "CVM code",
	}
	CodeFileFlag = cli.StringFlag{
		Name:  "codefile",
		Usage: "File containing CVM code. If '-' is specified, code is read from stdin ",
	}
	EnergyFlag = cli.Uint64Flag{
		Name:  "energy",
		Usage: "energy limit for the cvm",
		Value: 10000000000,
	}
	PriceFlag = utils.BigFlag{
		Name:  "price",
		Usage: "price set for the cvm",
		Value: new(big.Int),
	}
	ValueFlag = utils.BigFlag{
		Name:  "value",
		Usage: "value set for the cvm",
		Value: new(big.Int),
	}
	DumpFlag = cli.BoolFlag{
		Name:  "dump",
		Usage: "dumps the state after the run",
	}
	InputFlag = cli.StringFlag{
		Name:  "input",
		Usage: "input for the CVM",
	}
	InputFileFlag = cli.StringFlag{
		Name:  "inputfile",
		Usage: "file containing input for the CVM",
	}
	VerbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Usage: "sets the verbosity level",
	}
	BenchFlag = cli.BoolFlag{
		Name:  "bench",
		Usage: "benchmark the execution",
	}
	CreateFlag = cli.BoolFlag{
		Name:  "create",
		Usage: "indicates the action should be create rather than call",
	}
	GenesisFlag = cli.StringFlag{
		Name:  "prestate",
		Usage: "JSON file with prestate (genesis) config",
	}
	MachineFlag = cli.BoolFlag{
		Name:  "json",
		Usage: "output trace logs in machine readable format (json)",
	}
	SenderFlag = cli.StringFlag{
		Name:  "sender",
		Usage: "The transaction origin",
	}
	ReceiverFlag = cli.StringFlag{
		Name:  "receiver",
		Usage: "The transaction receiver (execution context)",
	}
	DisableMemoryFlag = cli.BoolFlag{
		Name:  "nomemory",
		Usage: "disable memory output",
	}
	DisableStackFlag = cli.BoolFlag{
		Name:  "nostack",
		Usage: "disable stack output",
	}
	DisableStorageFlag = cli.BoolFlag{
		Name:  "nostorage",
		Usage: "disable storage output",
	}
	DisableReturnDataFlag = cli.BoolFlag{
		Name:  "noreturndata",
		Usage: "disable return data output",
	}
	CVMInterpreterFlag = cli.StringFlag{
		Name:  "vm.cvm",
		Usage: "External CVM configuration (default = built-in interpreter)",
		Value: "",
	}
)

var stateTransitionCommand = cli.Command{
	Name:    "transition",
	Aliases: []string{"t8n"},
	Usage:   "executes a full state transition",
	Action:  t8ntool.Main,
	Flags: []cli.Flag{
		t8ntool.TraceFlag,
		t8ntool.TraceDisableMemoryFlag,
		t8ntool.TraceDisableStackFlag,
		t8ntool.TraceDisableReturnDataFlag,
		t8ntool.OutputBasedir,
		t8ntool.OutputAllocFlag,
		t8ntool.OutputResultFlag,
		t8ntool.InputAllocFlag,
		t8ntool.InputEnvFlag,
		t8ntool.InputTxsFlag,
		t8ntool.ForknameFlag,
		t8ntool.NetworkIDFlag,
		t8ntool.RewardFlag,
		t8ntool.VerbosityFlag,
	},
}

func init() {
	app.Flags = []cli.Flag{
		BenchFlag,
		CreateFlag,
		DebugFlag,
		VerbosityFlag,
		CodeFlag,
		CodeFileFlag,
		EnergyFlag,
		PriceFlag,
		ValueFlag,
		DumpFlag,
		InputFlag,
		InputFileFlag,
		MemProfileFlag,
		CPUProfileFlag,
		StatDumpFlag,
		GenesisFlag,
		MachineFlag,
		SenderFlag,
		ReceiverFlag,
		DisableMemoryFlag,
		DisableStackFlag,
		DisableStorageFlag,
		DisableReturnDataFlag,
		CVMInterpreterFlag,
		utils.NetworkIdFlag,
	}
	app.Commands = []cli.Command{
		compileCommand,
		disasmCommand,
		runCommand,
		stateTestCommand,
		stateTransitionCommand,
	}
	cli.CommandHelpTemplate = flags.OriginCommandHelpTemplate
}

func main() {
	if err := app.Run(os.Args); err != nil {
		code := 1
		if ec, ok := err.(*t8ntool.NumberedError); ok {
			code = ec.Code()
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(code)
	}
}
