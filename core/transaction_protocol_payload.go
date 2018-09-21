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
	"github.com/nebulasio/go-nebulas/util"
)

// ProtocolPayload carry ir data
type ProtocolPayload struct {
	Data []byte
}

// LoadProtocolPayload from bytes
func LoadProtocolPayload(bytes []byte) (*ProtocolPayload, error) {
	return NewProtocolPayload(bytes), nil
}

// NewProtocolPayload with data
func NewProtocolPayload(data []byte) *ProtocolPayload {
	return &ProtocolPayload{
		Data: data,
	}
}

// ToBytes serialize payload
func (payload *ProtocolPayload) ToBytes() ([]byte, error) {
	return payload.Data, nil
}

// BaseGasCount returns base gas count
func (payload *ProtocolPayload) BaseGasCount() *util.Uint128 {
	return util.NewUint128()
}

// Execute the payload in tx
func (payload *ProtocolPayload) Execute(limitedGas *util.Uint128, tx *Transaction, block *Block, ws WorldState) (*util.Uint128, string, error) {
	return util.NewUint128(), "", nil
}
