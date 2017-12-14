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

package dpos

import (
	"testing"

	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"
)

type Neb struct {
	genesis *corepb.Genesis
	storage storage.Storage
	emitter *core.EventEmitter
}

func (n *Neb) Genesis() *corepb.Genesis {
	return n.genesis
}

func (n *Neb) Storage() storage.Storage {
	return n.storage
}

func (n *Neb) EventEmitter() *core.EventEmitter {
	return n.emitter
}

func TestDpos_mintBlock(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	eventEmitter := core.NewEventEmitter()
	neb := &Neb{
		genesis: core.DefaultGenesisConf(0),
		storage: storage,
		emitter: eventEmitter,
	}

	chain, _ := core.NewBlockChain(neb)
	tail := chain.TailBlock()
	elapsedSecond := int64(core.DynastySize*core.BlockInterval + core.DynastyInterval)
	context, err := tail.NextDynastyContext(elapsedSecond)
	assert.Nil(t, err)
	coinbase, err := core.AddressParse("1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c")
	assert.Nil(t, err)
	block, err := core.NewBlock(chain.ChainID(), coinbase, tail)
	assert.Nil(t, err)
	block.LoadDynastyContext(context)
	block.SetMiner(coinbase)
	block.Seal()
	manager := account.NewManager(nil)
	miner, err := core.AddressParseFromBytes(context.Proposer)
	assert.Nil(t, err)
	assert.Nil(t, manager.Unlock(miner, []byte("passphrase")))
	assert.Nil(t, manager.SignBlock(miner, block))
	dpos := Dpos{blockInterval: core.BlockInterval, chain: chain}
	assert.Nil(t, dpos.VerifyBlock(block, tail))
}
