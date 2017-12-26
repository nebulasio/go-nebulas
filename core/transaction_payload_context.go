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

import "github.com/nebulasio/go-nebulas/core/state"

// PayloadContext transaction payload context
type PayloadContext struct {
	block *Block
	tx    *Transaction

	accState state.AccountState
}

// NewPayloadContext returns new payloadcontxt
func NewPayloadContext(block *Block, tx *Transaction) *PayloadContext {
	ctx := &PayloadContext{block: block, tx: tx}
	return ctx
}

// Block returns ctx block
func (ctx *PayloadContext) Block() *Block {
	return ctx.block
}

// Transaction returns ctx transaction
func (ctx *PayloadContext) Transaction() *Transaction {
	return ctx.tx
}

// BeginBatch begin a batch task
func (ctx *PayloadContext) BeginBatch() (err error) {
	ctx.accState, err = ctx.block.accState.Clone()
	return err
}

// Commit a batch task
func (ctx *PayloadContext) Commit() {
	ctx.block.accState = ctx.accState
}

// RollBack a batch task
func (ctx *PayloadContext) RollBack() {
	// do nothing
}
