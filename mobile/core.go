// Copyright 2016 by the Authors
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

// Contains all the wrappers from the go-core root package.

package gocore

import (
	"errors"

	core "github.com/core-coin/go-core"
	"github.com/core-coin/go-core/common"
)

// Subscription represents an event subscription where events are
// delivered on a data channel.
type Subscription struct {
	sub core.Subscription
}

// Unsubscribe cancels the sending of events to the data channel
// and closes the error channel.
func (s *Subscription) Unsubscribe() {
	s.sub.Unsubscribe()
}

// CallMsg contains parameters for contract calls.
type CallMsg struct {
	msg core.CallMsg
}

// NewCallMsg creates an empty contract call parameter list.
func NewCallMsg() *CallMsg {
	return new(CallMsg)
}

func (msg *CallMsg) GetFrom() *Address       { return &Address{msg.msg.From} }
func (msg *CallMsg) GetEnergy() int64        { return int64(msg.msg.Energy) }
func (msg *CallMsg) GetEnergyPrice() *BigInt { return &BigInt{msg.msg.EnergyPrice} }
func (msg *CallMsg) GetValue() *BigInt       { return &BigInt{msg.msg.Value} }
func (msg *CallMsg) GetData() []byte         { return msg.msg.Data }
func (msg *CallMsg) GetTo() *Address {
	if to := msg.msg.To; to != nil {
		return &Address{*msg.msg.To}
	}
	return nil
}

func (msg *CallMsg) SetFrom(address *Address)     { msg.msg.From = address.address }
func (msg *CallMsg) SetEnergy(energy int64)       { msg.msg.Energy = uint64(energy) }
func (msg *CallMsg) SetEnergyPrice(price *BigInt) { msg.msg.EnergyPrice = price.bigint }
func (msg *CallMsg) SetValue(value *BigInt)       { msg.msg.Value = value.bigint }
func (msg *CallMsg) SetData(data []byte)          { msg.msg.Data = common.CopyBytes(data) }
func (msg *CallMsg) SetTo(address *Address) {
	if address == nil {
		msg.msg.To = nil
		return
	}
	msg.msg.To = &address.address
}

// SyncProgress gives progress indications when the node is synchronising with
// the Core network.
type SyncProgress struct {
	progress core.SyncProgress
}

func (p *SyncProgress) GetStartingBlock() int64 { return int64(p.progress.StartingBlock) }
func (p *SyncProgress) GetCurrentBlock() int64  { return int64(p.progress.CurrentBlock) }
func (p *SyncProgress) GetHighestBlock() int64  { return int64(p.progress.HighestBlock) }
func (p *SyncProgress) GetPulledStates() int64  { return int64(p.progress.PulledStates) }
func (p *SyncProgress) GetKnownStates() int64   { return int64(p.progress.KnownStates) }

// Topics is a set of topic lists to filter events with.
type Topics struct{ topics [][]common.Hash }

// NewTopics creates a slice of uninitialized Topics.
func NewTopics(size int) *Topics {
	return &Topics{
		topics: make([][]common.Hash, size),
	}
}

// NewTopicsEmpty creates an empty slice of Topics values.
func NewTopicsEmpty() *Topics {
	return NewTopics(0)
}

// Size returns the number of topic lists inside the set
func (t *Topics) Size() int {
	return len(t.topics)
}

// Get returns the topic list at the given index from the slice.
func (t *Topics) Get(index int) (hashes *Hashes, _ error) {
	if index < 0 || index >= len(t.topics) {
		return nil, errors.New("index out of bounds")
	}
	return &Hashes{t.topics[index]}, nil
}

// Set sets the topic list at the given index in the slice.
func (t *Topics) Set(index int, topics *Hashes) error {
	if index < 0 || index >= len(t.topics) {
		return errors.New("index out of bounds")
	}
	t.topics[index] = topics.hashes
	return nil
}

// Append adds a new topic list to the end of the slice.
func (t *Topics) Append(topics *Hashes) {
	t.topics = append(t.topics, topics.hashes)
}

// FilterQuery contains options for contract log filtering.
type FilterQuery struct {
	query core.FilterQuery
}

// NewFilterQuery creates an empty filter query for contract log filtering.
func NewFilterQuery() *FilterQuery {
	return new(FilterQuery)
}

func (fq *FilterQuery) GetFromBlock() *BigInt    { return &BigInt{fq.query.FromBlock} }
func (fq *FilterQuery) GetToBlock() *BigInt      { return &BigInt{fq.query.ToBlock} }
func (fq *FilterQuery) GetAddresses() *Addresses { return &Addresses{fq.query.Addresses} }
func (fq *FilterQuery) GetTopics() *Topics       { return &Topics{fq.query.Topics} }

func (fq *FilterQuery) SetFromBlock(fromBlock *BigInt)    { fq.query.FromBlock = fromBlock.bigint }
func (fq *FilterQuery) SetToBlock(toBlock *BigInt)        { fq.query.ToBlock = toBlock.bigint }
func (fq *FilterQuery) SetAddresses(addresses *Addresses) { fq.query.Addresses = addresses.addresses }
func (fq *FilterQuery) SetTopics(topics *Topics)          { fq.query.Topics = topics.topics }
