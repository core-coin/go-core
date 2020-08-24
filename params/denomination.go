// Copyright 2017 by the Authors
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

package params

// These are the multipliers for core denominations.
// Example: To get the ore value of an amount in 'nucle', use
//
//    new(big.Int).Mul(value, big.NewInt(params.Nucle))
//
const (
	Ore         = 1
	Wav         = 1e3
	Grav        = 1e6
	Nucle       = 1e9
	Atom        = 1e12
	Moli        = 1e15
	Core        = 1e18
	Aer         = 1e21
	Orb         = 1e24
	Plano       = 1e27
	Terra       = 1e30
	Sola        = 1e33
	Galx        = 1e36
	Cluster     = 1e39
	Supermatter = 1e42
)
