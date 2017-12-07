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
	"github.com/nebulasio/go-nebulas/storage"
	log "github.com/sirupsen/logrus"
)

// VotePayload carry election information
type VotePayload struct {
	Delegatee *Address
}

var (
	vote = []byte("vote")
)

// LoadVotePayload from bytes
func LoadVotePayload(bytes []byte) (*VotePayload, error) {
	delegatee, err := AddressParseFromBytes(bytes)
	if err != nil {
		return nil, err
	}
	return &VotePayload{Delegatee: delegatee}, nil
}

// NewVotePayload with function & args
func NewVotePayload(addr string) (*VotePayload, error) {
	delegatee, err := AddressParse(addr)
	if err != nil {
		return nil, err
	}
	return &VotePayload{
		Delegatee: delegatee,
	}, nil
}

// ToBytes serialize payload
func (payload *VotePayload) ToBytes() ([]byte, error) {
	return payload.Delegatee.Bytes(), nil
}

// Execute the call payload in tx, call a function
func (payload *VotePayload) Execute(tx *Transaction, block *Block) error {
	delegator := tx.from.Bytes()
	delegatee := payload.Delegatee.Bytes()
	_, err := block.dposContext.candidateTrie.Get(delegatee)
	if err != nil && err != storage.ErrKeyNotFound {
		return err
	}
	if err != nil {
		log.WithFields(log.Fields{
			"func":      "VotePayload.Execute",
			"delegator": tx.from.ToHex(),
			"delegatee": payload.Delegatee.ToHex(),
		}).Error("cannot vote for a non-candidate")
		return nil
	}
	pre, err := block.dposContext.voteTrie.Get(delegator)
	if err != nil && err != storage.ErrKeyNotFound {
		return err
	}
	if err == nil {
		key := append(pre, delegator...)
		if _, err = block.dposContext.delegateTrie.Del(key); err != nil {
			return err
		}
	}
	key := append(delegatee, delegator...)
	if _, err = block.dposContext.delegateTrie.Put(key, delegator); err != nil {
		return err
	}
	_, err = block.dposContext.voteTrie.Put(delegator, delegatee)
	return err
}
