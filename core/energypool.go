// Copyright 2020 by the Authors
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

package core

import (
	"fmt"
	"math"
)

// EnergyPool tracks the amount of energy available during execution of the transactions
// in a block. The zero value is a pool with zero energy available.
type EnergyPool uint64

// AddEnergy makes energy available for execution.
func (gp *EnergyPool) AddEnergy(amount uint64) *EnergyPool {
	if uint64(*gp) > math.MaxUint64-amount {
		panic("energy pool pushed above uint64")
	}
	*(*uint64)(gp) += amount
	return gp
}

// SubEnergy deducts the given amount from the pool if enough energy is
// available and returns an error otherwise.
func (gp *EnergyPool) SubEnergy(amount uint64) error {
	if uint64(*gp) < amount {
		return ErrEnergyLimitReached
	}
	*(*uint64)(gp) -= amount
	return nil
}

// Energy returns the amount of energy remaining in the pool.
func (gp *EnergyPool) Energy() uint64 {
	return uint64(*gp)
}

func (gp *EnergyPool) String() string {
	return fmt.Sprintf("%d", *gp)
}
