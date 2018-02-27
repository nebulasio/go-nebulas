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

	"errors"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

const (
	// DefaultLimitsOfTotalMemorySize default limits of total memory size
	DefaultLimitsOfTotalMemorySize uint64 = 40 * 1000 * 1000
)

// Block interface breaks cycle import dependency and hides unused services.
type Block interface {
	CoinbaseHash() byteutils.Hash
	Hash() byteutils.Hash
	Height() uint64
	VerifyAddress(str string) bool
	SerializeTxByHash(hash byteutils.Hash) (proto.Message, error)
	RecordEvent(txHash byteutils.Hash, topic, data string) error
}

// AccountState context account state
type AccountState struct {
	Nonce   uint64 `json:"nonce"`
	Balance string `json:"balance"`
}

// ContextBlock warpper block
type ContextBlock struct {
	Coinbase string `json:"coinbase"`
	Hash     string `json:"hash"`
	Height   uint64 `json:"height"`
}

// ContextTransaction warpper transaction
type ContextTransaction struct {
	Hash      string `json:"hash"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     string `json:"value"`
	Nonce     uint64 `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	GasPrice  string `json:"gasPrice"`
	GasLimit  string `json:"gasLimit"`
}

// Context nvm engine context
type Context struct {
	block    Block
	tx       *ContextTransaction
	owner    state.Account
	contract state.Account
	state    state.AccountState
}

// NewContext create a engine context
func NewContext(block Block, tx *ContextTransaction, owner state.Account, contract state.Account, state state.AccountState) *Context {
	ctx := &Context{
		block:    block,
		tx:       tx,
		owner:    owner,
		contract: contract,
		state:    state,
	}
	return ctx
}

// State returns account state
func (ctx *Context) State() state.AccountState {
	return ctx.state
}

// Owner returns contract owner account
func (ctx *Context) Owner() state.Account {
	return ctx.owner
}

// Contract returns contract account
func (ctx *Context) Contract() state.Account {
	return ctx.contract
}

// SerializeContextBlock Serialize current block
func (ctx *Context) SerializeContextBlock() ([]byte, error) {

	if ctx.block != nil {
		block := &ContextBlock{
			Coinbase: ctx.block.CoinbaseHash().String(),
			Hash:     ctx.block.Hash().String(),
			Height:   ctx.block.Height(),
		}
		return json.Marshal(block)
	}
	return nil, errors.New("no block in context")
}

// SerializeContextTx Serialize current tx
func (ctx *Context) SerializeContextTx() ([]byte, error) {
	return json.Marshal(ctx.tx)
}

// SerializeTxByHash Serialize tx
func (ctx *Context) SerializeTxByHash(hash byteutils.Hash) ([]byte, error) {
	msg, err := ctx.block.SerializeTxByHash(hash)
	if err != nil {
		return nil, err
	}
	txMsg := msg.(*corepb.Transaction)
	value, _ := util.NewUint128FromFixedSizeByteSlice(txMsg.Value)
	gasPrice, _ := util.NewUint128FromFixedSizeByteSlice(txMsg.GasPrice)
	gasLimit, _ := util.NewUint128FromFixedSizeByteSlice(txMsg.GasLimit)
	tx := &ContextTransaction{
		Hash:      byteutils.Hex(txMsg.Hash),
		From:      byteutils.Hex(txMsg.From),
		To:        byteutils.Hex(txMsg.To),
		Value:     value.String(),
		Nonce:     txMsg.Nonce,
		Timestamp: txMsg.Timestamp,
		GasPrice:  gasPrice.String(),
		GasLimit:  gasLimit.String(),
	}
	return json.Marshal(tx)
}
