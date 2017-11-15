// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package core

import (
	"encoding/json"

	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// Action types
const (
	PrepareAction  = "prepare"
	CommitAction   = "commit"
	ChangeAction   = "change"
	AbdicateAction = "abdicate"
)

// VotePayload carry function call information
type VotePayload struct {
	Action    string
	BlockHash byteutils.Hash
}

// LoadVotePayload from bytes
func LoadVotePayload(bytes []byte) (*VotePayload, error) {
	payload := &VotePayload{}
	if err := json.Unmarshal(bytes, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// NewVotePayload with function & args
func NewVotePayload(action string, blockHash byteutils.Hash) *VotePayload {
	return &VotePayload{
		Action:    action,
		BlockHash: blockHash,
	}
}

// ToBytes serialize payload
func (payload *VotePayload) ToBytes() ([]byte, error) {
	return json.Marshal(payload)
}

// Execute the call payload in tx, call a function
func (payload *VotePayload) Execute(tx *Transaction, block *Block) error {
	return nil
}
