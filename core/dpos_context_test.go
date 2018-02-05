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

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

func checkDynasty(t *testing.T, dynasty *trie.BatchTrie) {
	delegatees, err := TraverseDynasty(dynasty)
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), MockDynasty[i])
	}
}

func TestBlock_NextDynastyContext(t *testing.T) {
	neb := testNeb()
	chain, _ := NewBlockChain(neb)
	var c MockConsensus
	chain.SetConsensusHandler(c)
	block, _ := LoadBlockFromStorage(GenesisHash, chain.storage, chain.txPool, neb.emitter)

	context, err := block.NextDynastyContext(chain, BlockInterval)
	assert.Nil(t, err)
	validators, _ := TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[1])
	// check dynasty
	checkDynasty(t, context.DynastyTrie)
	checkDynasty(t, context.NextDynastyTrie)

	context, err = block.NextDynastyContext(chain, BlockInterval+DynastyInterval)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[1])
	// check dynasty
	checkDynasty(t, context.DynastyTrie)
	checkDynasty(t, context.NextDynastyTrie)

	context, err = block.NextDynastyContext(chain, DynastyInterval/2)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[int(DynastyInterval/2/BlockInterval)%DynastySize])
	// check dynasty
	checkDynasty(t, context.DynastyTrie)
	checkDynasty(t, context.NextDynastyTrie)

	context, err = block.NextDynastyContext(chain, DynastyInterval*2+DynastyInterval/3)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	index := int((DynastyInterval*2+DynastyInterval/3)%DynastyInterval) / int(BlockInterval) % DynastySize
	assert.Equal(t, context.Proposer, validators[index])
	// check dynasty
	checkDynasty(t, context.DynastyTrie)
	checkDynasty(t, context.NextDynastyTrie)

	// new block
	coinbase := &Address{validators[1]}
	newBlock, _ := NewBlock(chain.ChainID(), coinbase, chain.tailBlock)
	newBlock.LoadDynastyContext(context)
	newBlock.SetMiner(coinbase)
	newBlock.Seal()
	newBlock, _ = mockBlockFromNetwork(newBlock)
	newBlock.LinkParentBlock(chain, chain.tailBlock)
	newBlock.SetMiner(coinbase)
	assert.Nil(t, newBlock.VerifyExecution(chain.tailBlock, chain.ConsensusHandler()))
}

func TestBlock_ElectNewDynasty(t *testing.T) {
	neb := testNeb()
	chain, _ := NewBlockChain(neb)
	block, _ := LoadBlockFromStorage(GenesisHash, chain.storage, chain.txPool, neb.emitter)
	block.begin()
	kickout, _ := AddressParse(MockDynasty[1])
	v, err := AddressParse(MockDynasty[len(MockDynasty)-1])
	assert.Nil(t, err)
	acc, err := block.accState.GetOrCreateUserAccount(v.Bytes())
	assert.Nil(t, err)
	acc.AddBalance(util.NewUint128FromInt(2000000))
	delegatePayload := NewDelegatePayload(DelegateAction, v.String())
	bytes, _ := delegatePayload.ToBytes()
	tx := NewTransaction(0, kickout, kickout, util.NewUint128FromInt(1), 1, TxPayloadDelegateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	_, _, err = block.executeTransaction(tx)
	assert.Nil(t, err)
	candidatePayload := NewCandidatePayload(LogoutAction)
	bytes, _ = candidatePayload.ToBytes()
	tx = NewTransaction(0, kickout, kickout, util.NewUint128FromInt(1), 2, TxPayloadCandidateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	_, _, err = block.executeTransaction(tx)
	assert.Nil(t, err)
	block.commit()
	context, err := block.NextDynastyContext(chain, DynastyInterval)
	assert.Nil(t, err)
	_, err = context.NextDynastyTrie.Get(kickout.Bytes())
	assert.Equal(t, storage.ErrKeyNotFound, err)
	_, err = context.NextDynastyTrie.Get(v.Bytes())
	assert.Equal(t, err, nil)
}

func TestBlock_Kickout(t *testing.T) {
	neb := testNeb()
	chain, _ := NewBlockChain(neb)
	var c MockConsensus
	chain.SetConsensusHandler(c)
	validators, _ := TraverseDynasty(chain.tailBlock.dposContext.dynastyTrie)
	coinbase := &Address{validators[2]}

	block, _ := NewBlock(0, coinbase, chain.tailBlock)
	block.header.timestamp = DynastyInterval
	context, err := chain.tailBlock.NextDynastyContext(chain, block.Timestamp()-chain.tailBlock.Timestamp())
	assert.Nil(t, err)
	block.LoadDynastyContext(context)
	block.SetMiner(coinbase)
	assert.Equal(t, block.Seal(), nil)
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(chain, chain.tailBlock), nil)
	block.SetMiner(coinbase)
	assert.Nil(t, block.VerifyExecution(chain.tailBlock, chain.ConsensusHandler()))
	chain.SetTailBlock(block)
	checkDynasty(t, chain.tailBlock.dposContext.dynastyTrie)
	checkDynasty(t, chain.tailBlock.dposContext.nextDynastyTrie)

	block, _ = NewBlock(0, coinbase, block)
	block.header.timestamp = DynastyInterval * 2
	context, err = chain.tailBlock.NextDynastyContext(chain, block.Timestamp()-chain.tailBlock.Timestamp())
	assert.Nil(t, err)
	block.LoadDynastyContext(context)
	block.SetMiner(coinbase)
	assert.Equal(t, block.Seal(), nil)
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(chain, chain.tailBlock), nil)
	block.SetMiner(coinbase)
	assert.Nil(t, block.VerifyExecution(chain.tailBlock, chain.ConsensusHandler()))
	chain.SetTailBlock(block)
	checkDynasty(t, chain.tailBlock.dposContext.dynastyTrie)
	checkDynasty(t, chain.tailBlock.dposContext.nextDynastyTrie)
}

func TestTallyVotes(t *testing.T) {
	neb := testNeb()
	chain, err := NewBlockChain(neb)
	assert.Nil(t, err)
	var c MockConsensus
	chain.SetConsensusHandler(c)

	dc, err := GenesisDynastyContext(chain, neb.Genesis())
	assert.Nil(t, err)
	dc.Accounts, err = state.NewAccountState(nil, neb.storage)
	tester := "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8"
	candidate, err := AddressParse(tester)
	assert.Nil(t, err)
	dc.Accounts.BeginBatch()
	acc, err := dc.Accounts.GetOrCreateUserAccount(candidate.Bytes())
	assert.Nil(t, err)
	acc.AddBalance(util.NewUint128FromInt(10000))
	dc.Accounts.Commit()
	assert.Nil(t, err)
	// empty candidates
	candidates := dc.CandidateTrie
	dc.CandidateTrie, err = trie.NewBatchTrie(nil, neb.storage)
	votes, err := dc.tallyVotes()
	assert.Nil(t, err)
	assert.Equal(t, votes, make(map[string]*util.Uint128))
	dc.CandidateTrie = candidates
	dc.VoteTrie.Del(candidate.Bytes())
	dc.DelegateTrie.Del(append(candidate.Bytes(), candidate.Bytes()...))
	votes, err = dc.tallyVotes()
	assert.Nil(t, err)
	assert.Equal(t, votes[tester], util.NewUint128())
}

func TestChooseCandidates(t *testing.T) {
	neb := testNeb()
	chain, err := NewBlockChain(neb)
	dc, err := chain.TailBlock().NextDynastyContext(chain, 0)
	assert.Nil(t, err)
	votes, err := dc.tallyVotes()
	assert.Nil(t, err)
	tester := "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8"
	candidate, err := AddressParse(tester)
	assert.Nil(t, err)
	chain.genesisBlock.dposContext.kickoutCandidate(candidate.Bytes())
	dc.ProtectTrie = chain.genesisBlock.dposContext.candidateTrie
	lenVotes := len(votes)
	votes2, err := dc.chooseCandidates(votes)
	assert.Nil(t, err)
	lenVotes2 := len(votes2)
	assert.Equal(t, lenVotes, lenVotes2)
	assert.Equal(t, votes2[lenVotes2-1].Address.String(), tester)
}

func TestKickoutDynastyActuallyKickoutCandidates(t *testing.T) {
	neb := testNeb()
	chain, err := NewBlockChain(neb)
	dc, err := chain.TailBlock().NextDynastyContext(chain, 0)
	assert.Nil(t, err)
	tester := "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8"
	candidate, err := AddressParse(tester)
	assert.Nil(t, err)
	chain.genesisBlock.dposContext.kickoutCandidate(candidate.Bytes())
	dc.ProtectTrie = chain.genesisBlock.dposContext.candidateTrie
	assert.Nil(t, dc.kickoutDynasty(0))
	candidates, err := TraverseDynasty(dc.CandidateTrie)
	assert.Nil(t, err)
	assert.Equal(t, len(candidates), len(neb.Genesis().Consensus.Dpos.Dynasty)-1)
}

func TestCheckActiveBootstrapValidators(t *testing.T) {
	neb := testNeb()
	chain, err := NewBlockChain(neb)
	protect := chain.genesisBlock.dposContext.candidateTrie

	tester := "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8"
	candidate, err := AddressParse(tester)
	assert.Nil(t, err)
	candidates, err := trie.NewBatchTrie(nil, neb.storage)
	assert.Nil(t, err)
	active, err := checkActiveBootstrapValidator(candidate.Bytes(), protect, candidates)
	assert.Equal(t, active, false)

	candidates = chain.TailBlock().dposContext.candidateTrie
	chain.genesisBlock.dposContext.kickoutCandidate(candidate.Bytes())
	protect = chain.genesisBlock.dposContext.candidateTrie
	active, err = checkActiveBootstrapValidator(candidate.Bytes(), protect, candidates)
	assert.Equal(t, active, false)
	assert.Nil(t, err)

	candidates.Del(candidate.Bytes())
	active, err = checkActiveBootstrapValidator(candidate.Bytes(), protect, candidates)
	assert.Equal(t, active, false)
	assert.Nil(t, err)
}

func TestElectNextDynastyOnBaseDynastyWhenTooFewCandidates(t *testing.T) {
	neb := testNeb()
	chain, err := NewBlockChain(neb)
	dc, err := chain.TailBlock().NextDynastyContext(chain, 0)
	members, err := TraverseDynasty(dc.CandidateTrie)
	assert.Nil(t, err)
	for i := 0; i < len(members)-SafeSize+1; i++ {
		assert.Nil(t, dc.kickoutCandidate(members[i]))
	}
	assert.Equal(t, dc.electNextDynastyOnBaseDynasty(0, 1, false), ErrTooFewCandidates)
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
	neb := testNeb()
	neb.genesis.Consensus.Dpos.Dynasty = []string{}
	_, err := NewBlockChain(neb)
	assert.Equal(t, err, ErrInitialDynastyNotEnough)
}
