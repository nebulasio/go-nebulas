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

	"github.com/nebulasio/go-nebulas/util"
)

// DipPayload carry ir data
type DipPayload struct {
	StartHeight uint64
	EndHeight   uint64
	Version     uint64
}

// LoadDipPayload from bytes
func LoadDipPayload(bytes []byte) (*DipPayload, error) {
	payload := &DipPayload{}
	if err := json.Unmarshal(bytes, payload); err != nil {
		return nil, ErrInvalidArgument
	}
	return NewDipPayload(payload.StartHeight, payload.EndHeight, payload.Version)
}

// NewDipPayload with data
func NewDipPayload(start, end, version uint64) (*DipPayload, error) {
	if end < start {
		return nil, ErrInvalidArgument
	}
	return &DipPayload{
		StartHeight: start,
		EndHeight:   end,
		Version:     version,
	}, nil
}

// ToBytes serialize payload
func (payload *DipPayload) ToBytes() ([]byte, error) {
	return json.Marshal(payload)
}

// BaseGasCount returns base gas count
func (payload *DipPayload) BaseGasCount() *util.Uint128 {
	return util.NewUint128()
}

// Execute the payload in tx
func (payload *DipPayload) Execute(limitedGas *util.Uint128, tx *Transaction, block *Block, ws WorldState) (*util.Uint128, string, error) {
	if block == nil || tx == nil {
		return util.NewUint128(), "", ErrNilArgument
	}

	return util.NewUint128(), "", nil
}
