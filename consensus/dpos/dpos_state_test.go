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
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

func checkDynasty(t *testing.T, consensus core.Consensus, consensusRoot byteutils.Hash, storage storage.Storage) {
	consensusState, err := consensus.NewState(consensusRoot, storage)
	assert.Nil(t, err)
	dynasty, err := consensusState.Dynasty()
	assert.Nil(t, err)
	nextDynasty, err := consensusState.NextDynasty()
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(dynasty[i].Hex()), DefaultOpenDynasty[i])
		assert.Equal(t, string(nextDynasty[i].Hex()), DefaultOpenDynasty[i])
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

func TestBlock_ElectNewDynasty(t *testing.T) {
	neb := mockNeb(t)
	chain := neb.chain

	block, _ := core.LoadBlockFromStorage(core.GenesisHash, chain)
	block.Begin()
	kickout, _ := core.AddressParse(DefaultOpenDynasty[1])
	v, err := core.AddressParse(DefaultOpenDynasty[len(DefaultOpenDynasty)-1])
	assert.Nil(t, err)
	acc, err := block.WorldState().GetOrCreateUserAccount(v.Bytes())
	assert.Nil(t, err)
	acc.AddBalance(util.NewUint128FromInt(2000000))
	delegatePayload := core.NewDelegatePayload(core.DelegateAction, v.String())
	bytes, _ := delegatePayload.ToBytes()
	tx := core.NewTransaction(0, kickout, kickout, util.NewUint128FromInt(1), 1, core.TxPayloadDelegateType, bytes, core.TransactionGasPrice, util.NewUint128FromInt(200000))
	_, err = block.ExecuteTransaction(tx)
	assert.Nil(t, err)
	candidatePayload := core.NewCandidatePayload(core.LogoutAction)
	bytes, _ = candidatePayload.ToBytes()
	tx = core.NewTransaction(0, kickout, kickout, util.NewUint128FromInt(1), 2, core.TxPayloadCandidateType, bytes, core.TransactionGasPrice, util.NewUint128FromInt(200000))
	_, err = block.ExecuteTransaction(tx)
	assert.Nil(t, err)
	block.Commit()
	context, err := block.WorldState().NextConsensusState(DynastyInterval)
	assert.Nil(t, err)
	nextDynasty, err := context.NextDynasty()
	assert.Nil(t, err)
	kickoutExist := false
	validatorExist := false
	for _, validator := range nextDynasty {
		if validator.Equals(kickout.Bytes()) {
			kickoutExist = true
		}
		if validator.Equals(v.Bytes()) {
			validatorExist = true
		}
	}
	assert.Equal(t, kickoutExist, false)
	assert.Equal(t, validatorExist, true)
}

func TestBlock_Kickout(t *testing.T) {
	neb := mockNeb(t)
	chain := neb.chain
	validators, _ := chain.TailBlock().WorldState().Dynasty()
	coinbase, err := core.AddressParseFromBytes(validators[0])
	assert.Nil(t, err)
	assert.Nil(t, neb.am.Unlock(coinbase, []byte("passphrase"), keystore.DefaultUnlockDuration))

	block, _ := core.NewBlock(0, coinbase, chain.TailBlock())
	block.SetTimestamp(DynastyInterval)
	context, err := chain.TailBlock().WorldState().NextConsensusState(block.Timestamp() - chain.TailBlock().Timestamp())
	assert.Nil(t, err)
	block.WorldState().SetConsensusState(context)
	block.SetMiner(coinbase)
	assert.Equal(t, block.Seal(), nil)
	assert.Nil(t, neb.am.SignBlock(coinbase, block))
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(chain, chain.TailBlock()), nil)
	block.SetMiner(coinbase)
	assert.Nil(t, block.VerifyExecution(chain.TailBlock(), chain.ConsensusHandler()))
	chain.SetTailBlock(block)
	consensusRoot, err := chain.TailBlock().WorldState().ConsensusRoot()
	assert.Nil(t, err)
	checkDynasty(t, neb.consensus, consensusRoot, neb.Storage())

	block, _ = core.NewBlock(0, coinbase, block)
	block.SetTimestamp(DynastyInterval * 2)
	context, err = chain.TailBlock().WorldState().NextConsensusState(block.Timestamp() - chain.TailBlock().Timestamp())
	assert.Nil(t, err)
	block.WorldState().SetConsensusState(context)
	block.SetMiner(coinbase)
	assert.Equal(t, block.Seal(), nil)
	assert.Nil(t, neb.am.SignBlock(coinbase, block))
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(chain, chain.TailBlock()), nil)
	block.SetMiner(coinbase)
	assert.Nil(t, block.VerifyExecution(chain.TailBlock(), chain.ConsensusHandler()))
	chain.SetTailBlock(block)
	consensusRoot, err = chain.TailBlock().WorldState().ConsensusRoot()
	assert.Nil(t, err)
	checkDynasty(t, neb.consensus, consensusRoot, neb.Storage())
}

func TestTallyVotes(t *testing.T) {
	neb := mockNeb(t)
	chain := neb.chain

	worldState := chain.TailBlock().WorldState()
	consensusRoot, err := worldState.ConsensusRoot()
	assert.Nil(t, err)
	consensusState, err := neb.consensus.NewState(consensusRoot, neb.Storage())
	assert.Nil(t, err)
	tester := "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8"
	candidate, err := core.AddressParse(tester)
	assert.Nil(t, err)
	worldState.Begin()
	acc, err := worldState.GetOrCreateUserAccount(candidate.Bytes())
	assert.Nil(t, err)
	acc.AddBalance(util.NewUint128FromInt(10000))
	worldState.Commit()
	assert.Nil(t, err)
	// empty candidates
	votes, err := consensusState.(*State).tallyVotes(worldState)
	assert.Nil(t, err)
	assert.Equal(t, votes[tester].String(), "10000000000000000010000")
	consensusState.DelVote(candidate.Bytes())
	consensusState.DelDelegate(candidate.Bytes(), candidate.Bytes())
	votes, err = consensusState.(*State).tallyVotes(worldState)
	assert.Nil(t, err)
	assert.Equal(t, votes[tester], util.NewUint128())
}

func TestChooseCandidates(t *testing.T) {
	neb := mockNeb(t)
	chain := neb.chain
	worldState := chain.TailBlock().WorldState()
	consensusState, err := worldState.NextConsensusState(0)
	assert.Nil(t, err)
	dposState := consensusState.(*State)
	protectTrie := dposState.candidateTrie
	votes, err := dposState.tallyVotes(worldState)
	assert.Nil(t, err)
	tester := "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8"
	candidate, err := core.AddressParse(tester)
	assert.Nil(t, err)
	dposState.kickoutCandidate(candidate.Bytes())
	dposState.protectTrie = protectTrie
	lenVotes := len(votes)
	votes2, err := dposState.chooseCandidates(votes)
	assert.Nil(t, err)
	lenVotes2 := len(votes2)
	assert.Equal(t, lenVotes, lenVotes2)
	assert.Equal(t, votes2[lenVotes2-1].Address.String(), tester)
}

func TestKickoutDynastyActuallyKickoutCandidates(t *testing.T) {
	neb := mockNeb(t)
	chain := neb.chain
	consensusState, err := chain.TailBlock().WorldState().NextConsensusState(0)
	assert.Nil(t, err)
	dposState := consensusState.(*State)
	protectTrie := dposState.candidateTrie
	tester := "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8"
	candidate, err := core.AddressParse(tester)
	assert.Nil(t, err)
	dposState.kickoutCandidate(candidate.Bytes())
	dposState.candidateTrie = protectTrie
	assert.Nil(t, dposState.kickoutDynasty(0))
	candidates, err := TraverseDynasty(dposState.candidateTrie)
	assert.Nil(t, err)
	assert.Equal(t, len(candidates), len(neb.Genesis().Consensus.Dpos.Dynasty)-1)
}

func TestCheckActiveBootstrapValidators(t *testing.T) {
	neb := mockNeb(t)
	chain := neb.chain
	consensusState, err := chain.TailBlock().WorldState().NextConsensusState(0)
	assert.Nil(t, err)
	dposState := consensusState.(*State)
	protectTrie := dposState.candidateTrie
	tester := "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8"
	candidate, err := core.AddressParse(tester)
	assert.Nil(t, err)
	candidates, err := trie.NewBatchTrie(nil, neb.storage)
	assert.Nil(t, err)
	_, err = candidates.Put(candidate.Bytes(), candidate.Bytes())
	assert.Nil(t, err)
	active, err := checkActiveBootstrapValidator(candidate.Bytes(), protectTrie, candidates)
	assert.Equal(t, active, true)

	dposState.kickoutCandidate(candidate.Bytes())
	active, err = checkActiveBootstrapValidator(candidate.Bytes(), protectTrie, candidates)
	assert.Equal(t, active, false)
	assert.Nil(t, err)

	candidates.Del(candidate.Bytes())
	active, err = checkActiveBootstrapValidator(candidate.Bytes(), protectTrie, candidates)
	assert.Equal(t, active, false)
	assert.Nil(t, err)
}

func TestElectNextDynastyOnBaseDynastyWhenTooFewCandidates(t *testing.T) {
	neb := mockNeb(t)
	chain := neb.chain
	worldState := chain.TailBlock().WorldState()
	consensusState, err := worldState.NextConsensusState(0)
	assert.Nil(t, err)
	dposState := consensusState.(*State)
	members, err := TraverseDynasty(dposState.candidateTrie)
	assert.Nil(t, err)
	for i := 0; i < len(members)-SafeSize+1; i++ {
		assert.Nil(t, dposState.kickoutCandidate(members[i]))
	}
	assert.Equal(t, dposState.electNextDynastyOnBaseDynasty(worldState, 0, 1, false), ErrTooFewCandidates)
}

func TestTraverseDynasty(t *testing.T) {
	stor, err := storage.NewMemoryStorage()
	assert.Nil(t, err)
	dynasty, err := trie.NewBatchTrie(nil, stor)
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
