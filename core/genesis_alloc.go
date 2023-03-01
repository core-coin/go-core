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
const mainnetAllocData = "\xf8\xfc\xe3\x96\xcb\x06+\r~\x13\xb4|@5\r+\xb7y@\bG7\u07aa\xb7U\x8bYU\xe3\xbb>t?\xec\x00\x00\x00\xe3\x96\xcb5}+*l\x1doAi\xf3\xb6\x18\xf9S\xea\x9e#q\xa9\u0632\x8b\x18\u043fB<\x03\xd8\xde\x00\x00\x00\xe3\x96\xcbUZ\xa3BQ\xab475\x9b\f\xa6_\xddU\xf6u\x85X\xac\xa1\x8b\x16Ux\xee\u03dd\x0f\xfb\x00\x00\x00\xe3\x96\u02c9\xe8In:\xab\x9bM\xee\x80\\\x92\xc5\u06c6\x057\x80\xc0\x13\xeb\x8b\x18\u043fB<\x03\xd8\xde\x00\x00\x00\xe3\x96\u02d3\x043h.\\\xd7&\xd9\xf6\x06\x9f\b\u017e/\xc6F\v\xaf\xf4\x8b\x04\xf6\x8c\xa6\xd8\u0351\xc6\x00\x00\x00\xe3\x96\u02d4\x85\xe8R=\xff\xd7P\x10,\xd0<\"\x87h\xe3\x00(\xd8\xf5\x03\x8b9\x13Q~\xbd<\fe\x00\x00\x00\xe3\x96\u02d5\x16\xeb\x8ae\xb7`\xd9\xd6&\xeb\xdc3\xc2\"\xfek^\x8bp\xe0\x8b\x18\u043fB<\x03\xd8\xde\x00\x00\x00"
const devinAllocData = "\xe5\u45ab\x92\xf8F\x86\xfc\x90\xf0\x0f\x1aP\xc1\xe7\xbf/\\\xf7'~\xbb\x8b\x10\x8c\x01\x9d\x97\x1eO\xe8@\x1et\x00\x00\x00"
