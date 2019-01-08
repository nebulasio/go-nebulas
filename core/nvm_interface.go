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

type NVMConfig struct {
	PayloadSource string
	PayloadSourceType string
	FunctionName string
	ContractArgs string
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