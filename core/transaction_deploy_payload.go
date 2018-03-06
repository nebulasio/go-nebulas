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
	"github.com/nebulasio/go-nebulas/core/state"

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
func (payload *DeployPayload) Execute(tx *Transaction, block *Block, txWorldState state.TxWorldState) (*util.Uint128, string, error) {
	nvmctx, err := generateDeployContext(tx, block, txWorldState)
	if err != nil {
		return util.NewUint128(), "", err
	}

	payloadGasLimit, err := tx.PayloadGasLimit(payload)
	if err != nil {
		return util.NewUint128(), "", err
	}
	// payloadGasLimit <= 0, v8 engine not limit the execution instructions
	if payloadGasLimit.Cmp(util.NewUint128()) <= 0 {
		return util.NewUint128(), "", ErrOutOfGasLimit
	}

	engine := nvm.NewV8Engine(nvmctx)
	defer engine.Dispose()

	engine.SetExecutionLimits(payloadGasLimit.Uint64(), nvm.DefaultLimitsOfTotalMemorySize)

	// Deploy and Init.iutu
	result, exeErr := engine.DeployAndInit(payload.Source, payload.SourceType, payload.Args)
	instructions, _ := util.NewUint128FromInt(int64(engine.ExecutionInstructions()))
	return instructions, result, exeErr
}

func generateDeployContext(tx *Transaction, block *Block, txWorldState state.TxWorldState) (*nvm.Context, error) {
	if block.height > NewOptimizeHeight {
		if !tx.From().Equals(tx.To()) {
			return nil, ErrContractTransactionAddressNotEqual
		}
	}

	addr, err := tx.GenerateContractAddress()
	if err != nil {
		return nil, err
	}
	owner, err := txWorldState.GetOrCreateUserAccount(tx.from.Bytes())
	if err != nil {
		return nil, err
	}
	contract, err := txWorldState.CreateContractAccount(addr.Bytes(), tx.Hash())
	if err != nil {
		return nil, err
	}
	nvmctx := nvm.NewContext(block, convertNvmTx(tx), owner, contract, txWorldState)
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
