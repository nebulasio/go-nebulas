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

package nvm

import (
	"encoding/json"

	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

var (
	// DefaultLimitsOfTotalMemorySize default limits of total menmory size
	DefaultLimitsOfTotalMemorySize uint64 = 20 * 1024 * 1024
)

// Blockchain interface breaks cycle import dependency and hides unused services.
type Blockchain interface {
	VerifyAddress(str string) bool
	SerializeBlockByHash(hash byteutils.Hash) ([]byte, error)
	SerializeTxByHash(hash byteutils.Hash) ([]byte, error)
}

// ContextParams warpper block & transaction info
type ContextParams struct {
	Coinbase    string `json:"coinbase"`
	BlockNonce  uint64 `json:"blockNonce"`
	BlockHash   string `json:"blockHash"`
	BlockHeight uint64 `json:"blockHeight"`
	TxNonce     uint64 `json:"txNonce"`
	TxHash      string `json:"txHash"`
}

// Context nvm engine context
type Context struct {
	params   *ContextParams
	owner    state.Account
	contract state.Account
	state    state.AccountState
	chain    Blockchain
}

// NewContext create a engine context
func NewContext(params *ContextParams, owner state.Account, contract state.Account, state state.AccountState) *Context {
	ctx := &Context{
		params:   params,
		owner:    owner,
		contract: contract,
		state:    state,
	}
	return ctx
}

// getParamsJson returns a json string with block & transaction info
func (c *Context) getParamsJSON() string {
	jsonStr, _ := json.Marshal(c.params)
	return string(jsonStr)
}
