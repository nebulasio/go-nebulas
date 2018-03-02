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

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/keystore"

	"github.com/stretchr/testify/assert"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

func checkDynasty(t *testing.T, consensus core.Consensus, consensusRoot byteutils.Hash, storage storage.Storage) {
	consensusState, err := consensus.NewState(consensusRoot, storage)
	assert.Nil(t, err)
	dynasty, err := consensusState.Dynasty()
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(dynasty[i].Hex()), DefaultOpenDynasty[i])
	}
}

func TestBlock_NextDynastyContext(t *testing.T) {
	neb := mockNeb(t)
	block := neb.chain.GenesisBlock()

	context, err := block.WorldState().NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	validators, _ := block.WorldState().Dynasty()
	assert.Equal(t, context.Proposer(), validators[1])
	// check dynasty
	consensusRoot, err := context.RootHash()
	assert.Nil(t, err)
	checkDynasty(t, neb.consensus, consensusRoot, neb.Storage())

	context, err = block.WorldState().NextConsensusState(BlockInterval + DynastyInterval)
	assert.Nil(t, err)
	validators, _ = block.WorldState().Dynasty()
	assert.Equal(t, context.Proposer(), validators[1])
	// check dynasty
	consensusRoot, err = context.RootHash()
	assert.Nil(t, err)
	checkDynasty(t, neb.consensus, consensusRoot, neb.Storage())

	context, err = block.WorldState().NextConsensusState(DynastyInterval / 2)
	assert.Nil(t, err)
	validators, _ = block.WorldState().Dynasty()
	assert.Equal(t, context.Proposer(), validators[int(DynastyInterval/2/BlockInterval)%DynastySize])
	// check dynasty
	consensusRoot, err = context.RootHash()
	assert.Nil(t, err)
	checkDynasty(t, neb.consensus, consensusRoot, neb.Storage())

	context, err = block.WorldState().NextConsensusState(DynastyInterval*2 + DynastyInterval/3)
	assert.Nil(t, err)
	validators, _ = block.WorldState().Dynasty()
	index := int((DynastyInterval*2+DynastyInterval/3)%DynastyInterval) / int(BlockInterval) % DynastySize
	assert.Equal(t, context.Proposer(), validators[index])
	// check dynasty
	consensusRoot, err = context.RootHash()
	assert.Nil(t, err)
	checkDynasty(t, neb.consensus, consensusRoot, neb.Storage())

	// new block
	coinbase, err := core.AddressParseFromBytes(validators[4])
	assert.Nil(t, err)
	assert.Nil(t, neb.am.Unlock(coinbase, []byte("passphrase"), keystore.DefaultUnlockDuration))

	newBlock, _ := core.NewBlock(neb.chain.ChainID(), coinbase, neb.chain.TailBlock())
	newBlock.SetTimestamp(DynastyInterval*2 + DynastyInterval/3)
	newBlock.WorldState().SetConsensusState(context)
	newBlock.SetMiner(coinbase)
	assert.Equal(t, newBlock.Seal(), nil)
	assert.Nil(t, neb.am.SignBlock(coinbase, newBlock))
	newBlock, _ = mockBlockFromNetwork(newBlock)
	newBlock.LinkParentBlock(neb.chain, neb.chain.TailBlock())
	newBlock.SetMiner(coinbase)
	assert.Nil(t, newBlock.VerifyExecution(neb.chain.TailBlock(), neb.chain.ConsensusHandler()))
}

func TestTraverseDynasty(t *testing.T) {
	stor, err := storage.NewMemoryStorage()
	assert.Nil(t, err)
	dynasty, err := trie.NewTrie(nil, stor)
	assert.Nil(t, err)
	members, err := TraverseDynasty(dynasty)
	assert.Nil(t, err)
	assert.Equal(t, members, []byteutils.Hash{})
}

func TestInitialDynastyNotEnough(t *testing.T) {
	neb := mockNeb(t)
	neb.genesis.Consensus.Dpos.Dynasty = []string{}
	chain, err := core.NewBlockChain(neb)
	assert.Nil(t, err)
	assert.Equal(t, chain.Setup(neb), core.ErrGenesisConfNotMatch)
	neb.storage, _ = storage.NewMemoryStorage()
	chain, err = core.NewBlockChain(neb)
	assert.Nil(t, err)
	assert.Equal(t, chain.Setup(neb), ErrInitialDynastyNotEnough)
}

/*
func TestBlock_DposCandidates(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain

	tail := bc.tailBlock

	ks := keystore.DefaultKS
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata)
	ks.SetKey(from.String(), priv, []byte("passphrase"))
	ks.Unlock(from.String(), []byte("passphrase"), time.Second*60*60*24*365)

	key, _ := ks.GetUnlocked(from.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))

	priv1 := secp256k1.GeneratePrivateKey()
	pubdata1, _ := priv1.PublicKey().Encoded()
	to, _ := NewAddressFromPublicKey(pubdata1)
	priv2 := secp256k1.GeneratePrivateKey()
	pubdata2, _ := priv2.PublicKey().Encoded()
	coinbase, _ := NewAddressFromPublicKey(pubdata2)

	block0, _ := NewBlock(bc.ChainID(), from, tail)
	block0.header.timestamp = BlockInterval
	block0.SetMiner(from)
	block0.Seal()
	assert.Nil(t, bc.storeBlockToStorage(block0))
	assert.Nil(t, bc.SetTailBlock(block0))

	block, _ := NewBlock(bc.ChainID(), coinbase, block0)
	block.header.timestamp = BlockInterval * 2
	bytes, _ := NewCandidatePayload(LoginAction).ToBytes()
	tx := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 1, TxPayloadCandidateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	tx.Sign(signature)
	bc.txPool.Push(tx)
	payload := NewDelegatePayload(DelegateAction, from.String())
	bytes, _ = payload.ToBytes()
	tx = NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 2, TxPayloadDelegateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	tx.Sign(signature)
	bc.txPool.Push(tx)
	assert.Equal(t, len(block.transactions), 0)
	assert.Equal(t, bc.txPool.cache.Len(), 2)
	block.CollectTransactions(time.Now().Unix() + 2)
	assert.Equal(t, len(block.transactions), 2)
	assert.Equal(t, block.txPool.cache.Len(), 0)
	block.SetMiner(coinbase)
	assert.Equal(t, block.Seal(), nil)
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(bc, bc.tailBlock), nil)
	block.SetMiner(coinbase)
	assert.Nil(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()))
	bytes, _ = block.worldState.GetCandidate(from.Bytes())
	assert.Equal(t, bytes, from.Bytes())
	bytes, _ = block.worldState.GetVote(from.Bytes())
	assert.Equal(t, bytes, from.Bytes())
	exist, _ := block.worldState.HasDelegate(from.Bytes(), from.Bytes())
	assert.Equal(t, exist, true)
	assert.Nil(t, bc.storeBlockToStorage(block))
	assert.Nil(t, bc.SetTailBlock(block))

	block, _ = NewBlock(bc.ChainID(), coinbase, block)
	block.header.timestamp = BlockInterval * 3
	payload = NewDelegatePayload(UnDelegateAction, from.String())
	bytes, _ = payload.ToBytes()
	tx = NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 3, TxPayloadDelegateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	tx.Sign(signature)
	bc.txPool.Push(tx)
	assert.Equal(t, len(block.transactions), 0)
	assert.Equal(t, bc.txPool.cache.Len(), 1)
	block.CollectTransactions(time.Now().Unix() + 2)
	assert.Equal(t, len(block.transactions), 1)
	assert.Equal(t, block.txPool.cache.Len(), 0)
	block.SetMiner(coinbase)
	assert.Equal(t, block.Seal(), nil)
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(bc, bc.tailBlock), nil)
	block.SetMiner(coinbase)
	assert.Nil(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()))
	_, err := block.worldState.GetCandidate(from.Bytes())
	assert.Equal(t, err, nil)
	_, err = block.worldState.GetVote(from.Bytes())
	assert.Equal(t, err, storage.ErrKeyNotFound)
	_, err = block.worldState.IterDelegate(from.Bytes())
	assert.Equal(t, err, storage.ErrKeyNotFound)
	assert.Nil(t, bc.storeBlockToStorage(block))
	assert.Nil(t, bc.SetTailBlock(block))

	block, _ = NewBlock(bc.ChainID(), coinbase, block)
	block.header.timestamp = BlockInterval * 4
	payload = NewDelegatePayload(DelegateAction, from.String())
	bytes, _ = payload.ToBytes()
	tx = NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 4, TxPayloadDelegateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	tx.Sign(signature)
	bc.txPool.Push(tx)
	bytes, _ = NewCandidatePayload(LogoutAction).ToBytes()
	tx = NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 5, TxPayloadCandidateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	tx.Sign(signature)
	bc.txPool.Push(tx)
	assert.Equal(t, len(block.transactions), 0)
	assert.Equal(t, bc.txPool.cache.Len(), 2)
	block.CollectTransactions(time.Now().Unix() + 2)
	assert.Equal(t, len(block.transactions), 2)
	assert.Equal(t, block.txPool.cache.Len(), 0)
	block.SetMiner(coinbase)
	assert.Equal(t, block.Seal(), nil)
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(bc, bc.tailBlock), nil)
	block.SetMiner(coinbase)
	assert.Nil(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()))
	_, err = block.worldState.GetCandidate(from.Bytes())
	assert.Equal(t, err, storage.ErrKeyNotFound)
	_, err = block.worldState.GetVote(from.Bytes())
	assert.Equal(t, err, storage.ErrKeyNotFound)
	_, err = block.worldState.IterDelegate(from.Bytes())
	assert.Equal(t, err, storage.ErrKeyNotFound)
	assert.Nil(t, bc.storeBlockToStorage(block))
	assert.Nil(t, bc.SetTailBlock(block))
}
*/

/*
func TestNewGenesisBlock(t *testing.T) {
	neb := testNeb(t)
	chain := neb.chain
	genesis := neb.chain.genesisBlock
	conf := MockGenesisConf()

	iter, err := genesis.worldState.IterVote()
	assert.Nil(t, err)
	exist, err := iter.Next()
	i := 0
	for exist {
		var addr byteutils.Hash
		addr, _ = byteutils.FromHex(MockDynasty[i])
		assert.Equal(t, addr, byteutils.Hash(iter.Value()))
		i++
		exist, err = iter.Next()
	}

	iter, err = genesis.worldState.IterDelegate(nil)
	assert.Nil(t, err)
	exist, err = iter.Next()
	i = 0
	for exist {
		var addr byteutils.Hash
		addr, _ = byteutils.FromHex(MockDynasty[i])
		assert.Equal(t, addr, byteutils.Hash(iter.Value()))
		i++
		exist, err = iter.Next()
	}

	for _, v := range conf.TokenDistribution {
		addr, _ := byteutils.FromHex(v.Address)
		acc, err := genesis.worldState.GetOrCreateUserAccount(addr)
		assert.Nil(t, err)
		assert.Equal(t, acc.Balance().String(), v.Value)
	}

	dumpConf, err := DumpGenesis(chain)
	assert.Nil(t, err)
	assert.Equal(t, dumpConf.Meta.ChainId, conf.Meta.ChainId)
	assert.Equal(t, dumpConf.Consensus.Dpos.Dynasty, conf.Consensus.Dpos.Dynasty)
	assert.Equal(t, dumpConf.TokenDistribution, conf.TokenDistribution)
}
*/
