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
)

// DeployPayload carry contract deploy information
type DeployPayload struct {
	SourceType string
	Source     string
	Args       string
}

// LoadDeployPayload from bytes
func LoadDeployPayload(bytes []byte) (*DeployPayload, error) {
	payload := &DeployPayload{}
	if err := json.Unmarshal(bytes, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// NewDeployPayload with source & args
func NewDeployPayload(source, sourceType, args string) *DeployPayload {
	return &DeployPayload{
		Source:     source,
		SourceType: sourceType,
		Args:       args,
	}
}

// ToBytes serialize payload
func (payload *DeployPayload) ToBytes() ([]byte, error) {
	return json.Marshal(payload)
}

// BaseGasCount returns base gas count
func (payload *DeployPayload) BaseGasCount() *util.Uint128 {
	return util.NewUint128()
}

// Execute deploy payload in tx, deploy a new contract
func (payload *DeployPayload) Execute(block *Block, tx *Transaction) (*util.Uint128, string, error) {
	nvmctx, err := generateDeployContext(block, tx)
	if err != nil {
		return util.NewUint128(), "", err
	}

	engine := nvm.NewV8Engine(nvmctx)
	defer engine.Dispose()

	engine.SetExecutionLimits(tx.PayloadGasLimit(payload).Uint64(), nvm.DefaultLimitsOfTotalMemorySize)

	// Deploy and Init.
	result, err := engine.DeployAndInit(payload.Source, payload.SourceType, payload.Args)
	return util.NewUint128FromInt(int64(engine.ExecutionInstructions())), result, err
}

func generateDeployContext(block *Block, tx *Transaction) (*nvm.Context, error) {
	addr, err := tx.GenerateContractAddress()
	if err != nil {
		return nil, err
	}
	owner, err := block.accState.GetOrCreateUserAccount(tx.from.Bytes())
	if err != nil {
		return nil, err
	}
	contract, err := block.accState.CreateContractAccount(addr.Bytes(), tx.Hash())
	if err != nil {
		return nil, err
	}
	nvmctx := nvm.NewContext(block, convertNvmTx(tx), owner, contract, block.accState)
	return nvmctx, nil
}

func convertNvmTx(tx *Transaction) *nvm.ContextTransaction {
	ctxTx := &nvm.ContextTransaction{
		From:      tx.from.String(),
		To:        tx.to.String(),
		Value:     tx.value.String(),
		Timestamp: tx.timestamp,
		Nonce:     tx.Nonce(),
		Hash:      tx.Hash().String(),
		GasPrice:  tx.GasPrice().String(),
		GasLimit:  tx.GasLimit().String(),
	}
	return ctxTx
}
