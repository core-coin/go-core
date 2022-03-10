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
const devinAllocData = "\xf8\xd8\u35ab\"\x80\xa5\x82j\f\u06cbO\xf5q\xde\x06\x94\x0fe\xb0\xe3\xcc\xf5\u00cb)[\xe9nd\x06ir\x00\x00\x00\u35ab2zH'\x19\xbdN\xe9\u5f3a\u0332\x80\x18|\xce\u06d8\u704b)[\xe9nd\x06ir\x00\x00\x00\u35ab5\xa5\xb4U|T\xd4Vz\xa9I@\r;4(yO\xe53\xff\x8b)[\xe9nd\x06ir\x00\x00\x00\u35abfx\xc2\x19=\x9b\xd0\x0f\xf4\x94\x97\f\x99 d\xb2\u06ca\xfd\x8b\u00cb)[\xe9nd\x06ir\x00\x00\x00\u35abr\xfd\x9f\xe1U\x180\xa9\x06k|\u03bepA\xba\xe5K3\x9ca\x8b)[\xe9nd\x06ir\x00\x00\x00\u35ab\x91\xbd\xba^\xa0\x9f\x80\xecK\xb6\x82\xe6\xe9-\x1fC\x13\xcc\x04\x96j\x8b)[\xe9nd\x06ir\x00\x00\x00"
