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

package xcbclient

import (
	"errors"
	"math/big"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/core/types"
)

// senderFromServer is a types.Signer that remembers the sender address returned by the RPC
// server. It is stored in the transaction's sender address cache to avoid an additional
// request in TransactionSender.
type senderFromServer struct {
	addr      common.Address
	blockhash common.Hash
	networkID int
}

var errNotCached = errors.New("sender not cached")

func setSenderFromServer(tx *types.Transaction, addr common.Address, block common.Hash, networkID int) {
	// Use types.Sender for side-effect to store our signer into the cache.
	types.Sender(&senderFromServer{addr, block, networkID}, tx)
}

func (s *senderFromServer) Equal(other types.Signer) bool {
	os, ok := other.(*senderFromServer)
	return ok && os.blockhash == s.blockhash
}

func (s *senderFromServer) Sender(tx *types.Transaction) (common.Address, error) {
	if s.blockhash == (common.Hash{}) {
		return common.Address{}, errNotCached
	}
	if int(tx.NetworkID()) != int(s.networkID) {
		return common.Address{}, types.ErrInvalidNetworkId
	}
	return s.addr, nil
}

func (s *senderFromServer) NetworkID() int {
	return s.networkID
}

func (s *senderFromServer) Hash(tx *types.Transaction) common.Hash {
	panic("can't sign with senderFromServer")
}
func (s *senderFromServer) SignatureValues(tx *types.Transaction, sig []byte) (R, S, V *big.Int, err error) {
	panic("can't sign with senderFromServer")
}
