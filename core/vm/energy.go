// Copyright 2015 The go-core Authors
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

package vm

import (
	"math/big"
)

// Energy costs
const (
	EnergyQuickStep   uint64 = 2
	EnergyFastestStep uint64 = 3
	EnergyFastStep    uint64 = 5
	EnergyMidStep     uint64 = 8
	EnergySlowStep    uint64 = 10
	EnergyExtStep     uint64 = 20
)

// callEnergy returns the actual energy cost of the call.
//
// The returned energy is energy - base * 63 / 64.
func callEnergy(availableEnergy, base uint64, callCost *big.Int) (uint64, error) {
	availableEnergy = availableEnergy - base
	energy := availableEnergy - availableEnergy/64
	// If the bit length exceeds 64 bit we know that the newly calculated "energy" for CIP150
	// is smaller than the requested amount. Therefor we return the new energy instead
	// of returning an error.
	if !callCost.IsUint64() || energy < callCost.Uint64() {
		return energy, nil
	}
	if !callCost.IsUint64() {
		return 0, errEnergyUintOverflow
	}

	return callCost.Uint64(), nil
}
