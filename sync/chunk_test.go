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

package sync

import (
	"time"

	"github.com/nebulasio/go-nebulas/account"
	corepb "github.com/nebulasio/go-nebulas/core/pb"

	"github.com/nebulasio/go-nebulas/consensus/dpos"
	"github.com/nebulasio/go-nebulas/util"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestChunk_generateChunkMeta(t *testing.T) {

	consensus := dpos.NewDpos()
	am, err := account.NewManager(nil)
	assert.Nil(t, err)
	neb := core.NewMockNeb(am, consensus, nil)
	chain := neb.BlockChain()

	ck := NewChunk(chain)

	source := `"use strict";var DepositeContent=function(text){if(text){var o=JSON.parse(text);this.balance=new BigNumber(o.balance);this.expiryHeight=new BigNumber(o.expiryHeight)}else{this.balance=new BigNumber(0);this.expiryHeight=new BigNumber(0)}};DepositeContent.prototype={toString:function(){return JSON.stringify(this)}};var BankVaultContract=function(){LocalContractStorage.defineMapProperty(this,"bankVault",{parse:function(text){return new DepositeContent(text)},stringify:function(o){return o.toString()}})};BankVaultContract.prototype={init:function(){},save:function(height){var from=Blockchain.transaction.from;var value=Blockchain.transaction.value;var bk_height=new BigNumber(Blockchain.block.height);var orig_deposit=this.bankVault.get(from);if(orig_deposit){value=value.plus(orig_deposit.balance)}var deposit=new DepositeContent();deposit.balance=value;deposit.expiryHeight=bk_height.plus(height);this.bankVault.put(from,deposit)},takeout:function(value){var from=Blockchain.transaction.from;var bk_height=new BigNumber(Blockchain.block.height);var amount=new BigNumber(value);var deposit=this.bankVault.get(from);if(!deposit){throw new Error("No deposit before.")}if(bk_height.lt(deposit.expiryHeight)){throw new Error("Can not takeout before expiryHeight.")}if(amount.gt(deposit.balance)){throw new Error("Insufficient balance.")}var result=Blockchain.transfer(from,amount);if(result!=0){throw new Error("transfer failed.")}Event.Trigger("BankVault",{Transfer:{from:Blockchain.transaction.to,to:from,value:amount.toString()}});deposit.balance=deposit.balance.sub(amount);this.bankVault.put(from,deposit)},balanceOf:function(){var from=Blockchain.transaction.from;return this.bankVault.get(from)}};module.exports=BankVaultContract;`
	sourceType := "js"
	argsDeploy := ""
	payload, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
	payloadDeploy, _ := payload.ToBytes()

	from, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
	assert.Nil(t, neb.AccountManager().Unlock(from, []byte("passphrase"), time.Second*60*60*24*365))

	blocks := []*core.Block{}
	for i := 0; i < 96; i++ {
		context, err := chain.TailBlock().WorldState().NextConsensusState(dpos.BlockIntervalInMs / dpos.SecondInMs)
		assert.Nil(t, err)
		coinbase, err := core.AddressParseFromBytes(context.Proposer())
		assert.Nil(t, err)
		assert.Nil(t, neb.AccountManager().Unlock(coinbase, []byte("passphrase"), time.Second*60*60*24*365))
		block, err := chain.NewBlock(coinbase)
		assert.Nil(t, err)
		block.WorldState().SetConsensusState(context)
		block.SetTimestamp(chain.TailBlock().Timestamp() + dpos.BlockIntervalInMs/dpos.SecondInMs)
		value, _ := util.NewUint128FromInt(1)
		gasLimit, _ := util.NewUint128FromInt(200000)
		txDeploy, _ := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, uint64(i+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, neb.AccountManager().SignTransaction(from, txDeploy))
		assert.Nil(t, neb.BlockChain().TransactionPool().Push(txDeploy))
		if i == 95 {
			block.CollectTransactions(time.Now().Unix()*1000 + 4000)
			assert.Equal(t, len(block.Transactions()), 96)
		}
		assert.Nil(t, block.Seal())
		assert.Nil(t, neb.AccountManager().SignBlock(coinbase, block))
		assert.Nil(t, chain.BlockPool().Push(block))
		blocks = append(blocks, block)
	}

	meta, err := ck.generateChunkHeaders(blocks[1].Hash())
	assert.Nil(t, err)
	assert.Equal(t, len(meta.ChunkHeaders), 3)
	chunks, err := ck.generateChunkData(meta.ChunkHeaders[1])
	assert.Nil(t, err)
	assert.Equal(t, len(chunks.Blocks), core.ChunkSize)
	for i := 0; i < core.ChunkSize; i++ {
		index := core.ChunkSize + 2 + i
		assert.Equal(t, int(chunks.Blocks[i].Height), index)
	}

	meta, err = ck.generateChunkHeaders(chain.GenesisBlock().Hash())
	assert.Nil(t, err)
	assert.Equal(t, len(meta.ChunkHeaders), 3)

	meta, err = ck.generateChunkHeaders(blocks[0].Hash())
	assert.Nil(t, err)
	assert.Equal(t, len(meta.ChunkHeaders), 3)

	meta, err = ck.generateChunkHeaders(blocks[31].Hash())
	assert.Nil(t, err)
	assert.Equal(t, int(blocks[31].Height()), 33)
	assert.Equal(t, len(meta.ChunkHeaders), 2)

	meta, err = ck.generateChunkHeaders(blocks[62].Hash())
	assert.Nil(t, err)
	assert.Equal(t, int(blocks[62].Height()), 64)
	assert.Equal(t, len(meta.ChunkHeaders), 2)

	neb2 := core.NewMockNeb(am, consensus, nil)
	chain2 := neb2.BlockChain()
	meta, err = ck.generateChunkHeaders(blocks[0].Hash())
	assert.Nil(t, err)
	for _, header := range meta.ChunkHeaders {
		chunk, err := ck.generateChunkData(header)
		assert.Nil(t, err)
		for _, v := range chunk.Blocks {
			block := new(core.Block)
			assert.Nil(t, block.FromProto(v))
			assert.Nil(t, chain2.BlockPool().Push(block))
			pbBlockHash, err := core.HashPbBlock(v)
			assert.Nil(t, err)
			assert.Equal(t, block.Hash(), pbBlockHash)
		}
	}
	tail := blocks[len(blocks)-1]
	bytes, err := chain2.Storage().Get(tail.Hash())
	assert.Nil(t, err)
	pbBlock := new(corepb.Block)
	checkBlock := new(core.Block)
	assert.Nil(t, proto.Unmarshal(bytes, pbBlock))
	assert.Nil(t, checkBlock.FromProto(pbBlock))
	assert.Equal(t, checkBlock.Hash(), tail.Hash())
}
