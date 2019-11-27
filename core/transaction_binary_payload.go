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
	"fmt"

	"github.com/nebulasio/go-nebulas/util"
)

// BinaryPayload carry some data
type BinaryPayload struct {
	Data []byte
}

// LoadBinaryPayload from bytes
func LoadBinaryPayload(bytes []byte) (*BinaryPayload, error) {
	return NewBinaryPayload(bytes), nil
}

// NewBinaryPayload with data
func NewBinaryPayload(data []byte) *BinaryPayload {
	return &BinaryPayload{
		Data: data,
	}
}

// ToBytes serialize payload
func (payload *BinaryPayload) ToBytes() ([]byte, error) {
	return payload.Data, nil
}

// BaseGasCount returns base gas count
func (payload *BinaryPayload) BaseGasCount() *util.Uint128 {
	return util.NewUint128()
}

// Execute the payload in tx
func (payload *BinaryPayload) Execute(limitedGas *util.Uint128, tx *Transaction, block *Block, ws WorldState) (*util.Uint128, string, error) {
	if block == nil || tx == nil || tx.to == nil {
		return util.NewUint128(), "", ErrNilArgument
	}

	// transfer to contract
	if tx.to.Type() == ContractAddress && AcceptAvailableAtHeight(block.height) {

		instructions, result, exeErr := executeContract(limitedGas, tx, block, ws, ContractAcceptFunc, "")
		if exeErr == ErrExecutionFailed && len(result) > 0 {
			exeErr = fmt.Errorf("Binary: %s", result)
		}
		if exeErr != nil {
			return instructions, "", exeErr
		}
		if NodeUpdateAtHeight(block.height) {
			return instructions, result, nil
		}
	}

	return util.NewUint128(), "", nil
}
