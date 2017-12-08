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
	"github.com/nebulasio/go-nebulas/util"
	log "github.com/sirupsen/logrus"
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
	ctx, deployPayload, err := generateCallContext(tx, block)
	if err != nil {
		return err
	}

	engine := nvm.NewV8Engine(ctx)
	defer engine.Dispose()

	//add gas limit and memory use limit
	engine.SetExecutionLimits(tx.GasLimit().Uint64(), nvm.DefaultLimitsOfTotalMemorySize)

	err = engine.Call(deployPayload.Source, deployPayload.SourceType, payload.Function, payload.Args)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err,
			"block":       block,
			"transaction": tx,
		}).Error("CallPayload Execute.")
	} else {
		block.accState = ctx.State()
	}
	return gasCombustion(engine, tx, block)
}

// EstimateGas the payload in tx
func (payload *CallPayload) EstimateGas(tx *Transaction, block *Block) (*util.Uint128, error) {
	// TODO: @larry by @robin. since we can rollback all changes after Execute, so we don't need such function.

	ctx, deployPayload, err := generateCallContext(tx, block)
	if err != nil {
		return nil, err
	}

	engine := nvm.NewV8Engine(ctx)
	defer engine.Dispose()

	executionInstructions := util.NewUint128()
	executionInstructions.Sub(tx.gasLimit.Int, tx.CalculateGas().Int)
	engine.SetExecutionLimits(executionInstructions.Uint64(), nvm.DefaultLimitsOfTotalMemorySize)

	err = engine.SimulationRun(deployPayload.Source, deployPayload.SourceType, payload.Function, payload.Args)
	if err != nil {
		return nil, err
	}

	return util.NewUint128FromInt(int64(engine.ExecutionInstructions())), nil
}

func generateCallContext(tx *Transaction, block *Block) (*nvm.Context, *DeployPayload, error) {
	context, err := block.accState.Clone()
	if err != nil {
		return nil, nil, err
	}
	contract, err := context.GetContractAccount(tx.to.Bytes())
	if err != nil {
		return nil, nil, err
	}
	birthTx, err := block.GetTransaction(contract.BirthPlace())
	if err != nil {
		return nil, nil, err
	}
	owner := context.GetOrCreateUserAccount(birthTx.from.Bytes())
	deploy, err := LoadDeployPayload(birthTx.data.Payload)
	if err != nil {
		return nil, nil, err
	}

	ctx := nvm.NewContext(block, convertNvmTx(tx), owner, contract, context)
	return ctx, deploy, nil
}
