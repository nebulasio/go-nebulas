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

	"github.com/nebulasio/go-nebulas/nf/nvm"
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

// Execute the call payload in tx, call a function
func (payload *CallPayload) Execute(tx *Transaction, block *Block) error {
	context := block.accState
	contract, err := context.GetContractAccount(tx.to.Bytes())
	if err != nil {
		return err
	}
	birthTx, err := block.GetTransaction(contract.BirthPlace())
	if err != nil {
		return err
	}
	owner := context.GetOrCreateUserAccount(birthTx.from.Bytes())
	deploy, err := LoadDeployPayload(birthTx.data.Payload)
	if err != nil {
		return err
	}

	ctxBlock := &nvm.ContextBlock{
		Coinbase: block.Coinbase().String(),
		Nonce:    block.Nonce(),
		Hash:     block.Hash().String(),
		Height:   block.Height(),
	}
	ctxTx := &nvm.ContextTransaction{
		Nonce:    tx.Nonce(),
		Hash:     tx.Hash().String(),
		GasPrice: tx.GasPrice(),
	}
	ctx := nvm.NewContext(ctxBlock, ctxTx, owner, contract, context)
	engine := nvm.NewV8Engine(ctx)
	//add gas limit and memory use limit
	engine.SetExecutionLimits(tx.GasLimit().Uint64(), nvm.DefaultLimitsOfTotalMemorySize)
	defer engine.Dispose()

	return engine.Call(deploy.Source, payload.Function, payload.Args)
}
