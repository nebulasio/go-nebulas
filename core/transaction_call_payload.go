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

// CallPayload carry function call information
type CallPayload struct {
	Function string
	Args     string
}

// LoadCallPayload from bytes
func LoadCallPayload(bytes []byte) (*CallPayload, error) {
	payload := &CallPayload{}
	if err := json.Unmarshal(bytes, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// NewCallPayload with function & args
func NewCallPayload(function, args string) *CallPayload {
	return &CallPayload{
		Function: function,
		Args:     args,
	}
}

// ToBytes serialize payload
func (payload *CallPayload) ToBytes() ([]byte, error) {
	return json.Marshal(payload)
}

// BaseGasCount returns base gas count
func (payload *CallPayload) BaseGasCount() *util.Uint128 {
	return util.NewUint128()
}

// Execute the call payload in tx, call a function
func (payload *CallPayload) Execute(block *Block, tx *Transaction) (*util.Uint128, string, error) {
	if block == nil || tx == nil {
		return util.NewUint128(), "", ErrNilArgument
	}

	//add gas limit and memory use limit
	payloadGasLimit, err := tx.PayloadGasLimit(payload)
	if err != nil {
		return util.NewUint128(), "", err
	}
	// payloadGasLimit <= 0, v8 engine not limit the execution instructions
	if payloadGasLimit.Cmp(util.NewUint128()) <= 0 {
		return util.NewUint128(), "", ErrOutOfGasLimit
	}

	contract, err := block.CheckContract(tx.to)
	if err != nil {
		return util.NewUint128(), "", err
	}

	birthTx, err := block.GetTransaction(contract.BirthPlace())
	if err != nil {
		return util.NewUint128(), "", err
	}
	owner, err := block.accState.GetOrCreateUserAccount(birthTx.from.Bytes())
	if err != nil {
		return util.NewUint128(), "", err
	}
	deploy, err := LoadDeployPayload(birthTx.data.Payload) // ToConfirm: move deploy payload in ctx.
	if err != nil {
		return util.NewUint128(), "", err
	}

	if err := block.nvm.StartEngine(block, tx, owner, contract, block.accState); err != nil {
		return util.NewUint128(), "", err
	}
	defer block.nvm.DisposeEngine()

	if err := block.nvm.SetEngineExecutionLimits(payloadGasLimit.Uint64()); err != nil {
		return util.NewUint128(), "", err
	}

	result, exeErr := block.nvm.CallEngine(deploy.Source, deploy.SourceType, payload.Function, payload.Args)
	gasCout, err := block.nvm.ExecutionInstructions()
	if err != nil {
		return util.NewUint128(), "", err
	}
	instructions, err := util.NewUint128FromInt(int64(gasCout))
	if err != nil {
		return util.NewUint128(), "", err
	}
	return instructions, result, exeErr
}
