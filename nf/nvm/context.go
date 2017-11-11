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
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// Block interface breaks cycle import dependency and hides unused properties.
type Block interface {
	Coinbase() []byte
	Nonce() uint64
	Hash() byteutils.Hash
	ParentHash() byteutils.Hash
	Height() uint64
}

// Transaction interface breaks cycle import dependency and hides unused properties.
type Transaction interface {
	ChainID() uint32
	Nonce() uint64
	Hash() byteutils.Hash
}

// Context nvm engine context
type Context struct {
	block    Block       //contract execute block
	tx       Transaction //contract execute transaction
	owner    state.Account
	contract state.Account
	state    state.AccountState
}

// NewContext create a engine context
func NewContext(block Block, tx Transaction, owner state.Account, contract state.Account, state state.AccountState) *Context {
	ctx := &Context{
		block:    block,
		tx:       tx,
		owner:    owner,
		contract: contract,
		state:    state,
	}
	return ctx
}
