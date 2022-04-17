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

package core

// Constants containing the genesis allocation of built-in genesis blocks.
// Their content is an RLP-encoded list of (address, balance) tuples.
// Use mkalloc.go to create/update them.

// nolint: misspell
const mainnetAllocData = "\xf8\xd8\xe3\x96\xcb\x06\x12&\xdc\xcePr\x0e\x86\xedl>\xbbi\xe3M$\xb9\xae\x83X\x8b)[\xe9nd\x06ir\x00\x00\x00\xe3\x96\xcb\x18\xdc\xc4\u071b\xe11\x95\t\x9f\xef=7\x9c+\xd9\t\xae\u0ca2\x8b)[\xe9nd\x06ir\x00\x00\x00\xe3\x96\xcb)\xad\x8f\x93y\x83h\xc3i+C\xfa\u06e1\xb6\xb4\x8aq_\x9a;\x8b)[\xe9nd\x06ir\x00\x00\x00\xe3\x96\xcb@\xed\xea\xefW)8c\x90\x16\xb2^\xbb\x95~C\x87\xec6\x82<\x8b)[\xe9nd\x06ir\x00\x00\x00\xe3\x96\xcbG\x96\xef\xad C-un\xe5'\xe7\xd1M\xb3\xc3\xf2\x9e\x8a0\b\x8b)[\xe9nd\x06ir\x00\x00\x00\xe3\x96\xcbc\xf4\u0611\xb4x\x8f\xcf[+E\xc4\x14/\xf5BY\xb2\xf3\xa9\u028b)[\xe9nd\x06ir\x00\x00\x00"
const devinAllocData = "\xe5\u45ab\x92\xf8F\x86\xfc\x90\xf0\x0f\x1aP\xc1\xe7\xbf/\\\xf7'~\xbb\x8b\x10\x8c\x01\x9d\x97\x1eO\xe8@\x1et\x00\x00\x00"
