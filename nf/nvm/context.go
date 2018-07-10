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
	"math/rand"
	"unsafe"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
)

// SerializableAccount serializable account state
type SerializableAccount struct {
	Nonce   uint64 `json:"nonce"`
	Balance string `json:"balance"`
}

// SerializableBlock serializable block
type SerializableBlock struct {
	Timestamp int64  `json:"timestamp"`
	Hash      string `json:"hash"`
	Height    uint64 `json:"height"`
	Seed      string `json:"seed,omitempty"`
}

// SerializableTransaction serializable transaction
type SerializableTransaction struct {
	Hash      string `json:"hash"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     string `json:"value"`
	Nonce     uint64 `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	GasPrice  string `json:"gasPrice"`
	GasLimit  string `json:"gasLimit"`
}

// ContextRand ..
type ContextRand struct {
	rand *rand.Rand
}

// Context nvm engine context
type Context struct {
	block       Block
	tx          Transaction
	contract    Account
	state       WorldState
	head        unsafe.Pointer
	index       uint32
	contextRand *ContextRand
}

// NewContext create a engine context
func NewContext(block Block, tx Transaction, contract Account, state WorldState) (*Context, error) {
	if block == nil || tx == nil || contract == nil || state == nil {
		return nil, ErrContextConstructArrEmpty
	}
	ctx := &Context{
		block:       block,
		tx:          tx,
		contract:    contract,
		state:       state,
		contextRand: &ContextRand{},
	}
	return ctx, nil
}

// NewInnerContext create a child engine context
func NewInnerContext(block Block, tx Transaction, contract Account, state WorldState, head unsafe.Pointer, index uint32, ctxRand *ContextRand) (*Context, error) {
	if block == nil || tx == nil || contract == nil || state == nil || head == nil {
		return nil, ErrContextConstructArrEmpty
	}
	ctx := &Context{
		block:       block,
		tx:          tx,
		contract:    contract,
		state:       state,
		head:        head,
		index:       index,
		contextRand: ctxRand,
	}
	return ctx, nil
}
func toSerializableAccount(acc Account) *SerializableAccount {
	sAcc := &SerializableAccount{
		Nonce:   acc.Nonce(),
		Balance: acc.Balance().String(),
	}
	return sAcc
}

func toSerializableBlock(block Block) *SerializableBlock {
	sBlock := &SerializableBlock{
		Timestamp: block.Timestamp(),
		Hash:      "",
		Height:    block.Height(),
	}
	if core.V8BlockSeedAvailableAtHeight(block.Height()) {
		sBlock.Seed = block.RandomSeed()
	}
	return sBlock
}

func toSerializableTransaction(tx Transaction) *SerializableTransaction {
	return &SerializableTransaction{
		From:      tx.From().String(),
		To:        tx.To().String(),
		Value:     tx.Value().String(),
		Timestamp: tx.Timestamp(),
		Nonce:     tx.Nonce(),
		Hash:      tx.Hash().String(),
		GasPrice:  tx.GasPrice().String(),
		GasLimit:  tx.GasLimit().String(),
	}
}

func toSerializableTransactionFromBytes(txBytes []byte) (*SerializableTransaction, error) {
	pbTx := new(corepb.Transaction)
	if err := proto.Unmarshal(txBytes, pbTx); err != nil {
		return nil, err
	}
	tx := new(core.Transaction)
	if err := tx.FromProto(pbTx); err != nil {
		return nil, err
	}

	return &SerializableTransaction{
		From:      tx.From().String(),
		To:        tx.To().String(),
		Value:     tx.Value().String(),
		Timestamp: tx.Timestamp(),
		Nonce:     tx.Nonce(),
		Hash:      tx.Hash().String(),
		GasPrice:  tx.GasPrice().String(),
		GasLimit:  tx.GasLimit().String(),
	}, nil
}
