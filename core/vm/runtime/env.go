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

package runtime

import (
	"github.com/core-coin/go-core/v2/core"
	"github.com/core-coin/go-core/v2/core/vm"
)

func NewEnv(cfg *Config) *vm.CVM {
	txContext := vm.TxContext{
		Origin:      cfg.Origin,
		EnergyPrice: cfg.EnergyPrice,
	}
	blockContext := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     cfg.GetHashFn,
		Coinbase:    cfg.Coinbase,
		BlockNumber: cfg.BlockNumber,
		Time:        cfg.Time,
		Difficulty:  cfg.Difficulty,
		EnergyLimit: cfg.EnergyLimit,
	}

	return vm.NewCVM(blockContext, txContext, cfg.State, cfg.ChainConfig, cfg.CVMConfig)
}
