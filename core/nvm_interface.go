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
	"github.com/nebulasio/go-nebulas/core/state"
)

type NVMConfig struct {

	// block related info
	Block    *Block
	Tx       *Transaction
	ContractAccount state.Account
	State    WorldState

	// other limitations
	LimitedGas uint64
	DefaultLimitsOfTotalMemorySize uint64
	PayloadSource string
	PayloadSourceType string
	FunctionName string
	ContractArgs string
}

type NVMExeResponse struct {
	GasCount int64
	Result string
	ExeError error

	ActualCountOfExecutionInstructions uint64
	ActualTotalMemorySize uint64
}


func (c *NVMConfig) GetBlock() *Block{
	return c.Block
}

func (c *NVMConfig) GetTransaction() *Transaction{
	return c.Tx
}

func (c *NVMConfig) GetContractAccount() state.Account{
	return c.ContractAccount
}

func (c *NVMConfig) GetState() WorldState {
	return c.State
}

func (c *NVMConfig) GetLimitedGas() uint64 {
	return c.LimitedGas	
}

func (c *NVMConfig) GetDefaultLimitsOfTotalMemorySize() uint64 {
	return c.DefaultLimitsOfTotalMemorySize
}

func (c *NVMConfig) GetPayloadSource() string {
	return c.PayloadSource
}

func (c *NVMConfig) GetPayloadSourceType() string {
	return c.PayloadSourceType
}

func (c *NVMConfig) GetContractArgs() string {
	return c.ContractArgs
}

func (c *NVMConfig) SetFunctionName(functionName string){
	c.FunctionName = functionName
}


func (response *NVMExeResponse) GetActualCountOfExecutionInstructions() uint64 {
	return response.ActualCountOfExecutionInstructions
}

func (response *NVMExeResponse) GetActualTotalMemorySize() uint64 {
	return response.ActualTotalMemorySize
}