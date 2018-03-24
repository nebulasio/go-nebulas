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
	"fmt"

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
func NewDeployPayload(source, sourceType, args string) *DeployPayload { // ToCheck: add version in sourceType // TODO check sourceType
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
	base, _ := util.NewUint128FromInt(60)
	return base
}

// Execute deploy payload in tx, deploy a new contract
func (payload *DeployPayload) Execute(limitedGas *util.Uint128, tx *Transaction, block *Block, ws WorldState) (*util.Uint128, string, error) {
	if block == nil || tx == nil {
		return util.NewUint128(), "", ErrNilArgument
	}

	if !tx.From().Equals(tx.To()) {
		return util.NewUint128(), "", ErrContractTransactionAddressNotEqual
	}

	// payloadGasLimit <= 0, v8 engine not limit the execution instructions
	if limitedGas.Cmp(util.NewUint128()) <= 0 {
		return util.NewUint128(), "", ErrOutOfGasLimit
	}

	addr, err := tx.GenerateContractAddress()
	if err != nil {
		return util.NewUint128(), "", err
	}
	owner, err := ws.GetOrCreateUserAccount(tx.from.Bytes())
	if err != nil {
		return util.NewUint128(), "", err
	}
	contract, err := ws.CreateContractAccount(addr.Bytes(), tx.Hash())
	if err != nil {
		return util.NewUint128(), "", err
	}

	engine, err := block.nvm.CreateEngine(block, tx, owner, contract, ws)
	if err != nil {
		return util.NewUint128(), "", err
	}
	defer engine.Dispose()

	if err := engine.SetExecutionLimits(limitedGas.Uint64(), DefaultLimitsOfTotalMemorySize); err != nil {
		return util.NewUint128(), "", err
	}

	// Deploy and Init.
	result, exeErr := engine.DeployAndInit(payload.Source, payload.SourceType, payload.Args)
	gasCout := engine.ExecutionInstructions()
	instructions, err := util.NewUint128FromInt(int64(gasCout))
	if err != nil {
		return util.NewUint128(), "", err
	}
	if exeErr != nil && len(result) > 0 {
		exeErr = fmt.Errorf("Deploy: %s", result)
	}
	return instructions, result, exeErr
}
