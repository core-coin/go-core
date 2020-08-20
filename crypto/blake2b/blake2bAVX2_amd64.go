// Copyright 2020 The CORE FOUNDATION, nadacia
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

// +build go1.7,amd64,!gccgo,!appengine

package blake2b

import "golang.org/x/sys/cpu"

func init() {
	useAVX2 = cpu.X86.HasAVX2
	useAVX = cpu.X86.HasAVX
	useSSE4 = cpu.X86.HasSSE41
}

//go:noescape
func fAVX2(h *[8]uint64, m *[16]uint64, c0, c1 uint64, flag uint64, rounds uint64)

//go:noescape
func fAVX(h *[8]uint64, m *[16]uint64, c0, c1 uint64, flag uint64, rounds uint64)

//go:noescape
func fSSE4(h *[8]uint64, m *[16]uint64, c0, c1 uint64, flag uint64, rounds uint64)

func f(h *[8]uint64, m *[16]uint64, c0, c1 uint64, flag uint64, rounds uint64) {
	switch {
	case useAVX2:
		fAVX2(h, m, c0, c1, flag, rounds)
	case useAVX:
		fAVX(h, m, c0, c1, flag, rounds)
	case useSSE4:
		fSSE4(h, m, c0, c1, flag, rounds)
	default:
		fGeneric(h, m, c0, c1, flag, rounds)
	}
}
