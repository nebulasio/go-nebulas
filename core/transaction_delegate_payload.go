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

// DelegatePayload carry election information
type DelegatePayload struct {
	Delegatee *Address
}

// LoadDelegatePayload from bytes
func LoadDelegatePayload(bytes []byte) (*DelegatePayload, error) {
	delegatee, err := AddressParseFromBytes(bytes)
	if err != nil {
		return nil, err
	}
	return &DelegatePayload{Delegatee: delegatee}, nil
}

// NewDelegatePayload with function & args
func NewDelegatePayload(addr string) (*DelegatePayload, error) {
	delegatee, err := AddressParse(addr)
	if err != nil {
		return nil, err
	}
	return &DelegatePayload{
		Delegatee: delegatee,
	}, nil
}

// ToBytes serialize payload
func (payload *DelegatePayload) ToBytes() []byte {
	return payload.Delegatee.Bytes()
}

// Execute the call payload in tx, call a function
func (payload *DelegatePayload) Execute(tx *Transaction, block *Block) error {
	delegator := tx.from.Bytes()
	delegatee := payload.Delegatee.Bytes()
	key := append(delegatee, delegator...)
	_, err := block.dposContext.delegateTrie.Put(key, delegator)
	return err
}
