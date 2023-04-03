// Copyright 2015 by the Authors
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

package tests

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/core-coin/go-core/v2/params"
)

// Forks table defines supported forks and their chain config.
var Forks = map[string]*params.ChainConfig{
	"Frontier": {
		NetworkID: big.NewInt(1),
	},
	"Homestead": {
		NetworkID: big.NewInt(1),
	},
	"CIP150": {
		NetworkID: big.NewInt(1),
	},
	"CIP158": {
		NetworkID: big.NewInt(1),
	},
	"Byzantium": {
		NetworkID: big.NewInt(1),
	},
	"Constantinople": {
		NetworkID: big.NewInt(1),
	},
	"ConstantinopleFix": {
		NetworkID: big.NewInt(1),
	},
	"Istanbul": {
		NetworkID: big.NewInt(1),
	},
	"FrontierToHomesteadAt5": {
		NetworkID: big.NewInt(1),
	},
	"HomesteadToCIP150At5": {
		NetworkID: big.NewInt(1),
	},
	"HomesteadToDaoAt5": {
		NetworkID: big.NewInt(1),
	},
	"CIP158ToByzantiumAt5": {
		NetworkID: big.NewInt(1),
	},
	"ByzantiumToConstantinopleAt5": {
		NetworkID: big.NewInt(1),
	},
	"ByzantiumToConstantinopleFixAt5": {
		NetworkID: big.NewInt(1),
	},
	"ConstantinopleFixToIstanbulAt5": {
		NetworkID: big.NewInt(1),
	},
	"Berlin": {
		NetworkID: big.NewInt(1),
	},
}

// Returns the set of defined fork names
func AvailableForks() []string {
	var availableForks []string
	for k := range Forks {
		availableForks = append(availableForks, k)
	}
	sort.Strings(availableForks)
	return availableForks
}

// UnsupportedForkError is returned when a test requests a fork that isn't implemented.
type UnsupportedForkError struct {
	Name string
}

func (e UnsupportedForkError) Error() string {
	return fmt.Sprintf("unsupported fork %q", e.Name)
}
