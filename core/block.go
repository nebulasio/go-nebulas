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
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"

	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// constants
const (
	EpochSize       = 100
	DynastySize     = 100
	CandidatesLimit = 3000
)

var (
	// BlockHashLength define a const of the length of Hash of Block in byte.
	BlockHashLength = 32

	// BlockReward given to coinbase
	// TODO: block reward should calculates dynamic.
	BlockReward = util.NewUint128FromInt(16)
)

// Errors in block
var (
	ErrInvalidBlockHash      = errors.New("invalid block hash")
	ErrInvalidBlockStateRoot = errors.New("invalid block state root hash")
	ErrInvalidBlockTxsRoot   = errors.New("invalid block txs root hash")
	ErrInvalidChainID        = errors.New("invalid transaction chainID")
	ErrDuplicatedTransaction = errors.New("duplicated transaction")
	ErrSmallTransactionNonce = errors.New("cannot accept a transaction with smaller nonce")
	ErrLargeTransactionNonce = errors.New("cannot accept a transaction with too bigger nonce")
	ErrGenesisHasNoParent    = errors.New("genesis block has no parent")
)

// ValidatorSort are able to sort validator
// Use hash(str)[0] to sort validators
type ValidatorSort struct {
	seed       uint64
	validators []byteutils.Hash
}

func (v *ValidatorSort) hash(validator byteutils.Hash) int64 {
	hasher := sha3.New256()
	hasher.Write(byteutils.FromUint64(v.seed))
	hasher.Write(validator)
	result := hasher.Sum(nil)
	return byteutils.Int64(result[:8])
}

func (v *ValidatorSort) Len() int {
	return len(v.validators)
}

func (v *ValidatorSort) Swap(i, j int) {
	v.validators[i], v.validators[j] = v.validators[j], v.validators[i]
}

func (v *ValidatorSort) Less(i, j int) bool {
	return v.hash(v.validators[i]) < v.hash(v.validators[j])
}

// BlockHeader of a block
type BlockHeader struct {
	hash       byteutils.Hash
	parentHash byteutils.Hash

	stateRoot byteutils.Hash
	txsRoot   byteutils.Hash

	// validators in current dynasty
	curDynastyRoot byteutils.Hash
	// validators in next dynasty
	nextDynastyRoot byteutils.Hash
	// candidates in second next dynasty
	dynastyCandidatesRoot byteutils.Hash
	// active validators in all dynasties
	validatorsRoot byteutils.Hash
	// deposit charged by validators
	depositRoot byteutils.Hash
	// all prepare votes
	prepareVotesRoot byteutils.Hash
	// all prepare votes are organized by their height
	heightPrepareVotesRoot byteutils.Hash
	// all commit votes
	commitVotesRoot byteutils.Hash
	// all change votes
	changeVotesRoot byteutils.Hash
	// all abdicate votes
	abdicateVotesRoot byteutils.Hash

	nonce     uint64
	coinbase  *Address
	timestamp int64
	chainID   uint32
}

// ToProto converts domain BlockHeader to proto BlockHeader
func (b *BlockHeader) ToProto() (proto.Message, error) {
	return &corepb.BlockHeader{
		Hash:       b.hash,
		ParentHash: b.parentHash,
		StateRoot:  b.stateRoot,
		TxsRoot:    b.txsRoot,

		CurDynastyRoot:         b.curDynastyRoot,
		NextDynastyRoot:        b.nextDynastyRoot,
		DynastyCandidatesRoot:  b.dynastyCandidatesRoot,
		ValidatorsRoot:         b.validatorsRoot,
		DepositRoot:            b.depositRoot,
		PrepareVotesRoot:       b.prepareVotesRoot,
		HeightPrepareVotesRoot: b.heightPrepareVotesRoot,
		CommitVotesRoot:        b.commitVotesRoot,
		ChangeVotesRoot:        b.changeVotesRoot,
		AbdicateVotesRoot:      b.abdicateVotesRoot,

		Nonce:     b.nonce,
		Coinbase:  b.coinbase.address,
		Timestamp: b.timestamp,
		ChainId:   b.chainID,
	}, nil
}

// FromProto converts proto BlockHeader to domain BlockHeader
func (b *BlockHeader) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.BlockHeader); ok {
		b.hash = msg.Hash
		b.parentHash = msg.ParentHash
		b.stateRoot = msg.StateRoot
		b.txsRoot = msg.TxsRoot

		b.curDynastyRoot = msg.CurDynastyRoot
		b.nextDynastyRoot = msg.NextDynastyRoot
		b.dynastyCandidatesRoot = msg.DynastyCandidatesRoot
		b.validatorsRoot = msg.ValidatorsRoot
		b.depositRoot = msg.DepositRoot
		b.prepareVotesRoot = msg.PrepareVotesRoot
		b.heightPrepareVotesRoot = msg.HeightPrepareVotesRoot
		b.commitVotesRoot = msg.CommitVotesRoot
		b.changeVotesRoot = msg.ChangeVotesRoot
		b.abdicateVotesRoot = msg.AbdicateVotesRoot

		b.nonce = msg.Nonce
		b.coinbase = &Address{msg.Coinbase}
		b.timestamp = msg.Timestamp
		b.chainID = msg.ChainId
		return nil
	}
	return errors.New("Pb Message cannot be converted into BlockHeader")
}

// Block structure
type Block struct {
	header       *BlockHeader
	transactions Transactions

	sealed      bool
	height      uint64
	parentBlock *Block
	accState    state.AccountState
	txsTrie     *trie.BatchTrie

	curDynastyTrie         *trie.BatchTrie // key: addr, value: addr
	nextDynastyTrie        *trie.BatchTrie // key: addr, value: addr
	dynastyCandidatesTrie  *trie.BatchTrie // key: addr, value: addr
	validatorsTrie         *trie.BatchTrie // key: dynasty hash + addr, value: addr
	depositTrie            *trie.BatchTrie // key: addr, value: deposit
	prepareVotesTrie       *trie.BatchTrie // key: block hash + addr, value: vote
	heightPrepareVotesTrie *trie.BatchTrie // key: height + addr, value: vote
	commitVotesTrie        *trie.BatchTrie // key: block hash + addr, value: vote
	changeVotesTrie        *trie.BatchTrie // key: block hash + addr, value: N
	abdicateVotesTrie      *trie.BatchTrie // key: dynasty hash + addr, value: vote

	txPool *TransactionPool

	storage storage.Storage
}

// ToProto converts domain Block into proto Block
func (block *Block) ToProto() (proto.Message, error) {
	header, _ := block.header.ToProto()
	if header, ok := header.(*corepb.BlockHeader); ok {
		var txs []*corepb.Transaction
		for _, v := range block.transactions {
			tx, err := v.ToProto()
			if err != nil {
				return nil, err
			}
			if tx, ok := tx.(*corepb.Transaction); ok {
				txs = append(txs, tx)
			} else {
				return nil, errors.New("Pb Message cannot be converted into Transaction")
			}
		}
		return &corepb.Block{
			Header:       header,
			Transactions: txs,
			Height:       block.height,
		}, nil
	}
	return nil, errors.New("Pb Message cannot be converted into BlockHeader")
}

// FromProto converts proto Block to domain Block
func (block *Block) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.Block); ok {
		block.header = new(BlockHeader)
		if err := block.header.FromProto(msg.Header); err != nil {
			return err
		}
		for _, v := range msg.Transactions {
			tx := new(Transaction)
			if err := tx.FromProto(v); err != nil {
				return err
			}
			block.transactions = append(block.transactions, tx)
		}
		block.height = msg.Height
		return nil
	}
	return errors.New("Pb Message cannot be converted into Block")
}

// NewBlock return new block.
func NewBlock(chainID uint32, coinbase *Address, parent *Block) *Block {
	accState, _ := parent.accState.Clone()
	txsTrie, _ := parent.txsTrie.Clone()

	curDynastyTrie, _ := parent.curDynastyTrie.Clone()
	nextDynastyTrie, _ := parent.nextDynastyTrie.Clone()
	dynastyCandidatesTrie, _ := parent.dynastyCandidatesTrie.Clone()
	validatorsTrie, _ := parent.validatorsTrie.Clone()
	depositTrie, _ := parent.depositTrie.Clone()
	prepareVotesTrie, _ := parent.prepareVotesTrie.Clone()
	heightPrepareVotesTrie, _ := parent.heightPrepareVotesTrie.Clone()
	commitVotesTrie, _ := parent.commitVotesTrie.Clone()
	changeVotesTrie, _ := parent.changeVotesTrie.Clone()
	abdicateVotesTrie, _ := parent.abdicateVotesTrie.Clone()

	block := &Block{
		header: &BlockHeader{
			parentHash: parent.Hash(),
			coinbase:   coinbase,
			nonce:      0,
			timestamp:  time.Now().Unix(),
			chainID:    chainID,
		},
		transactions: make(Transactions, 0),
		parentBlock:  parent,
		accState:     accState,
		txsTrie:      txsTrie,

		curDynastyTrie:         curDynastyTrie,
		nextDynastyTrie:        nextDynastyTrie,
		dynastyCandidatesTrie:  dynastyCandidatesTrie,
		validatorsTrie:         validatorsTrie,
		depositTrie:            depositTrie,
		prepareVotesTrie:       prepareVotesTrie,
		heightPrepareVotesTrie: heightPrepareVotesTrie,
		commitVotesTrie:        commitVotesTrie,
		changeVotesTrie:        changeVotesTrie,
		abdicateVotesTrie:      abdicateVotesTrie,

		txPool:  parent.txPool,
		height:  parent.Height() + 1,
		sealed:  false,
		storage: parent.storage,
	}

	change, err := parent.checkDynastyRule()
	if err != nil {
		panic("cannot create new block:" + err.Error())
	}
	if change {
		block.changeDynasty()
	}
	return block
}

func sortValidators(validators []byteutils.Hash, seed uint64) []byteutils.Hash {
	vs := &ValidatorSort{validators: validators, seed: seed}
	sort.Sort(vs)
	// TODO(roy): adjust the dynastysize dynamically
	// TODO(roy): limit the re-election validators
	return vs.validators
}

func traverseValidators(dynastyTrie *trie.BatchTrie, prefix []byte) ([]byteutils.Hash, error) {
	validators := []byteutils.Hash{}
	if dynastyTrie.Empty() {
		return validators, nil
	}
	iter, err := dynastyTrie.Iterator(prefix)
	if err != nil {
		return nil, err
	}
	for iter.Next() {
		validators = append(validators, iter.Value())
	}
	return validators, nil
}

func countValidators(dynastyTrie *trie.BatchTrie, prefix []byte) (int, error) {
	if dynastyTrie.Empty() {
		return 0, nil
	}
	iter, err := dynastyTrie.Iterator(prefix)
	if err != nil {
		return 0, err
	}
	count := 0
	for iter.Next() {
		count++
	}
	return count, nil
}

// NextBlockSortedValidators return the sorted validators who will propose and vote the next block
func (block *Block) NextBlockSortedValidators() ([]byteutils.Hash, error) {
	dynastyRoot, err := block.NextBlockDynastyRoot()
	if err != nil {
		return nil, err
	}
	validators, err := traverseValidators(block.validatorsTrie, dynastyRoot)
	if err != nil {
		return nil, err
	}
	return sortValidators(validators, block.height), nil
}

// NextBlockDynastyRoot return the dynasty root in next block
func (block *Block) NextBlockDynastyRoot() (byteutils.Hash, error) {
	change, err := block.checkDynastyRule()
	if err != nil {
		return nil, err
	}
	dynastyRoot := block.header.curDynastyRoot
	if change {
		dynastyRoot = block.header.nextDynastyRoot
	}
	return dynastyRoot, nil
}

func (block *Block) checkDynastyRule() (bool, error) {
	change, err := block.checkDynastyRuleEpochOver()
	if err != nil {
		return false, err
	}
	if change {
		return true, nil
	}
	return block.checkDynastyRuleTooFewValidators()
}

// dynasty rule: epoch over, create > 100 blocks
func (block *Block) checkDynastyRuleEpochOver() (bool, error) {
	var err error
	curBlock := block
	dynastyRoot := block.CurDynastyRoot()
	for !CheckGenesisBlock(curBlock) && curBlock.CurDynastyRoot().Equals(dynastyRoot) {
		curBlock, err = curBlock.ParentBlock()
		if err != nil {
			return false, err
		}
	}
	if curBlock.Height() < block.Height() && block.Height()-curBlock.Height() >= EpochSize {
		return true, nil
	}
	return false, nil
}

// dynasty rule: remove too many validators, remove > 1/3 * N
func (block *Block) checkDynastyRuleTooFewValidators() (bool, error) {
	if block.curDynastyTrie.Empty() {
		return true, nil
	}
	curSize, err := countValidators(block.validatorsTrie, block.curDynastyTrie.RootHash())
	if err != nil {
		return false, err
	}
	fullSize, err := countValidators(block.curDynastyTrie, nil)
	if err != nil {
		return false, err
	}
	if curSize < 2/3*fullSize {
		return true, nil
	}
	return false, nil
}

// change current dynasty to next one
// all candidates will login automatically
func (block *Block) changeDynasty() {
	block.curDynastyTrie = block.nextDynastyTrie
	block.nextDynastyTrie, _ = trie.NewBatchTrie(nil, block.storage)
	validators, _ := traverseValidators(block.dynastyCandidatesTrie, nil)
	validators = sortValidators(validators, block.height)
	count := 0
	for _, v := range validators {
		if count < DynastySize {
			block.nextDynastyTrie.Put(v, v)
			count++
		}
	}
	dynastyRoot := block.nextDynastyTrie.RootHash()
	count = 0
	for _, v := range validators {
		if count < DynastySize {
			key := append(dynastyRoot, v...)
			block.validatorsTrie.Put(key, v)
			count++
		}
	}
}

// ParentBlock return the parent block
func (block *Block) ParentBlock() (*Block, error) {
	if CheckGenesisBlock(block) {
		return nil, ErrGenesisHasNoParent
	}
	if block.parentBlock == nil {
		parentBlock, err := LoadBlockFromStorage(block.ParentHash(), block.storage, block.txPool)
		if err != nil {
			return nil, err
		}
		block.parentBlock = parentBlock
	}
	return block.parentBlock, nil
}

// Coinbase return block's coinbase
func (block *Block) Coinbase() *Address {
	return block.header.coinbase
}

// Nonce return nonce.
func (block *Block) Nonce() uint64 {
	return block.header.nonce
}

// SetNonce set nonce.
func (block *Block) SetNonce(nonce uint64) {
	if block.sealed {
		panic("Sealed block can't be changed.")
	}
	block.header.nonce = nonce
}

// SetTimestamp set timestamp
func (block *Block) SetTimestamp(timestamp int64) {
	if block.sealed {
		panic("Sealed block can't be changed.")
	}
	block.header.timestamp = timestamp
}

// Hash return block hash.
func (block *Block) Hash() byteutils.Hash {
	return block.header.hash
}

// StateRoot return state root hash.
func (block *Block) StateRoot() byteutils.Hash {
	return block.header.stateRoot
}

// TxsRoot return txs root hash.
func (block *Block) TxsRoot() byteutils.Hash {
	return block.header.txsRoot
}

// CurDynastyRoot return current dynasty root hash.
func (block *Block) CurDynastyRoot() byteutils.Hash {
	return block.header.curDynastyRoot
}

// NextDynastyRoot return next dynasty root hash.
func (block *Block) NextDynastyRoot() byteutils.Hash {
	return block.header.nextDynastyRoot
}

// DynastyCandidatesRoot return dynasty candidates root hash.
func (block *Block) DynastyCandidatesRoot() byteutils.Hash {
	return block.header.dynastyCandidatesRoot
}

// ValidatorsRoot return validators candidates root hash.
func (block *Block) ValidatorsRoot() byteutils.Hash {
	return block.header.validatorsRoot
}

// DepositRoot return deposit root hash.
func (block *Block) DepositRoot() byteutils.Hash {
	return block.header.depositRoot
}

// PrepareVotesRoot return prepare votes root hash.
func (block *Block) PrepareVotesRoot() byteutils.Hash {
	return block.header.prepareVotesRoot
}

// CommitVotesRoot return commit votes root hash.
func (block *Block) CommitVotesRoot() byteutils.Hash {
	return block.header.commitVotesRoot
}

// ChangeVotesRoot return change votes root hash.
func (block *Block) ChangeVotesRoot() byteutils.Hash {
	return block.header.changeVotesRoot
}

// AbdicateVotesRoot return abdicate votes root hash.
func (block *Block) AbdicateVotesRoot() byteutils.Hash {
	return block.header.abdicateVotesRoot
}

// HeightPrepareVotesRoot return height prepare votes root hash.
func (block *Block) HeightPrepareVotesRoot() byteutils.Hash {
	return block.header.heightPrepareVotesRoot
}

// ParentHash return parent hash.
func (block *Block) ParentHash() byteutils.Hash {
	return block.header.parentHash
}

// Height return height from genesis block.
func (block *Block) Height() uint64 {
	return block.height
}

// LinkParentBlock link parent block, return true if hash is the same; false otherwise.
func (block *Block) LinkParentBlock(parentBlock *Block) bool {
	if block.ParentHash().Equals(parentBlock.Hash()) == false {
		return false
	}

	block.accState, _ = parentBlock.accState.Clone()
	block.txsTrie, _ = parentBlock.txsTrie.Clone()

	block.curDynastyTrie, _ = parentBlock.curDynastyTrie.Clone()
	block.nextDynastyTrie, _ = parentBlock.nextDynastyTrie.Clone()
	block.dynastyCandidatesTrie, _ = parentBlock.dynastyCandidatesTrie.Clone()
	block.depositTrie, _ = parentBlock.depositTrie.Clone()
	block.prepareVotesTrie, _ = parentBlock.prepareVotesTrie.Clone()
	block.heightPrepareVotesTrie, _ = parentBlock.heightPrepareVotesTrie.Clone()
	block.commitVotesTrie, _ = parentBlock.commitVotesTrie.Clone()
	block.changeVotesTrie, _ = parentBlock.changeVotesTrie.Clone()
	block.abdicateVotesTrie, _ = parentBlock.abdicateVotesTrie.Clone()

	block.txPool = parentBlock.txPool
	block.parentBlock = parentBlock
	block.storage = parentBlock.storage

	// travel to calculate block height.
	depth := uint64(0)
	ancestorHeight := uint64(0)
	for ancestor := block; ancestor != nil; ancestor = ancestor.parentBlock {
		depth++
		ancestorHeight = ancestor.height
		if ancestor.height > 0 {
			break
		}
	}

	for ancestor := block; ancestor != nil && depth > 1; ancestor = ancestor.parentBlock {
		depth--
		ancestor.height = ancestorHeight + depth
	}

	return true
}

func (block *Block) begin() {
	log.Info("Block Begin.")

	block.accState.BeginBatch()
	block.txsTrie.BeginBatch()

	block.curDynastyTrie.BeginBatch()
	block.nextDynastyTrie.BeginBatch()
	block.dynastyCandidatesTrie.BeginBatch()
	block.depositTrie.BeginBatch()
	block.prepareVotesTrie.BeginBatch()
	block.heightPrepareVotesTrie.BeginBatch()
	block.commitVotesTrie.BeginBatch()
	block.changeVotesTrie.BeginBatch()
	block.abdicateVotesTrie.BeginBatch()
}

func (block *Block) commit() {
	block.accState.Commit()
	block.txsTrie.Commit()

	block.curDynastyTrie.Commit()
	block.nextDynastyTrie.Commit()
	block.dynastyCandidatesTrie.Commit()
	block.depositTrie.Commit()
	block.prepareVotesTrie.Commit()
	block.heightPrepareVotesTrie.Commit()
	block.commitVotesTrie.Commit()
	block.changeVotesTrie.Commit()
	block.abdicateVotesTrie.Commit()

	log.WithFields(log.Fields{
		"block": block,
	}).Info("Block Commit.")
}

func (block *Block) rollback() {
	block.accState.RollBack()
	block.txsTrie.RollBack()

	block.curDynastyTrie.RollBack()
	block.nextDynastyTrie.RollBack()
	block.dynastyCandidatesTrie.RollBack()
	block.depositTrie.RollBack()
	block.prepareVotesTrie.RollBack()
	block.heightPrepareVotesTrie.RollBack()
	block.commitVotesTrie.RollBack()
	block.changeVotesTrie.RollBack()
	block.abdicateVotesTrie.RollBack()

	log.WithFields(log.Fields{
		"block": block,
	}).Info("Block RollBack.")
}

// ReturnTransactions and giveback them to tx pool
// TODO(roy): optimize storage.
// if a block is reverted, we should erase all changes
// made by this block on storage. use refcount.
func (block *Block) ReturnTransactions() {
	for _, tx := range block.transactions {
		block.txPool.Push(tx)
	}
}

// CollectTransactions and add them to block.
func (block *Block) CollectTransactions(n int) {
	if block.sealed {
		panic("Sealed block can't be changed.")
	}

	pool := block.txPool
	var givebacks []*Transaction
	for !pool.Empty() && n > 0 {
		tx := pool.Pop()
		block.begin()
		giveback, err := block.executeTransaction(tx)
		if giveback {
			givebacks = append(givebacks, tx)
		}
		if err == nil {
			log.WithFields(log.Fields{
				"func":     "block.CollectionTransactions",
				"block":    block,
				"tx":       tx,
				"giveback": giveback,
			}).Info("tx is packed.")
			block.commit()
			block.transactions = append(block.transactions, tx)
			n--
		} else {
			log.WithFields(log.Fields{
				"func":     "block.CollectionTransactions",
				"block":    block,
				"tx":       tx,
				"err":      err,
				"giveback": giveback,
			}).Warn("invalid tx.")
			block.rollback()
		}
	}
	for _, tx := range givebacks {
		pool.Push(tx)
	}
}

// Sealed return true if block seals. Otherwise return false.
func (block *Block) Sealed() bool {
	return block.sealed
}

// Seal seal block, calculate stateRoot and block hash.
func (block *Block) Seal() {
	if block.sealed {
		return
	}

	log.WithFields(log.Fields{
		"block": block,
	}).Info("Block Seal.")

	block.begin()
	block.rewardCoinbase()
	block.commit()
	block.header.stateRoot = block.accState.RootHash()
	block.header.txsRoot = block.txsTrie.RootHash()

	block.header.curDynastyRoot = block.curDynastyTrie.RootHash()
	block.header.nextDynastyRoot = block.nextDynastyTrie.RootHash()
	block.header.dynastyCandidatesRoot = block.dynastyCandidatesTrie.RootHash()
	block.header.validatorsRoot = block.validatorsTrie.RootHash()
	block.header.depositRoot = block.depositTrie.RootHash()
	block.header.prepareVotesRoot = block.prepareVotesTrie.RootHash()
	block.header.heightPrepareVotesRoot = block.heightPrepareVotesTrie.RootHash()
	block.header.commitVotesRoot = block.commitVotesTrie.RootHash()
	block.header.changeVotesRoot = block.changeVotesTrie.RootHash()
	block.header.abdicateVotesRoot = block.abdicateVotesTrie.RootHash()

	block.header.hash = HashBlock(block)
	block.sealed = true
}

func (block *Block) String() string {
	return fmt.Sprintf(`Block %p { 
		height:%d; hash:%s; parentHash:%s; stateRoot:%s; txsRoot: %s; 
		dynastyRoot: %s; nextDynastyRoot: %s; 
		dynastyCandidatesRoot: %s; depositRoot: %s; validatorsRoot: %s;
		prepareVotesRoot: %s; heightPrepapreVotesRoot: %s; commitVotesRoot: %s; 
		changeVotesRoot: %s; abdicateVotesRoot: %s;
		nonce:%d; timestamp: %d}`,
		block,
		block.height,
		byteutils.Hex(block.header.hash),
		byteutils.Hex(block.header.parentHash),
		byteutils.Hex(block.StateRoot()),
		byteutils.Hex(block.TxsRoot()),

		byteutils.Hex(block.CurDynastyRoot()),
		byteutils.Hex(block.NextDynastyRoot()),
		byteutils.Hex(block.DynastyCandidatesRoot()),
		byteutils.Hex(block.ValidatorsRoot()),
		byteutils.Hex(block.DepositRoot()),
		byteutils.Hex(block.PrepareVotesRoot()),
		byteutils.Hex(block.HeightPrepareVotesRoot()),
		byteutils.Hex(block.CommitVotesRoot()),
		byteutils.Hex(block.ChangeVotesRoot()),
		byteutils.Hex(block.AbdicateVotesRoot()),

		block.header.nonce,
		block.header.timestamp,
	)
}

// Verify return block verify result, including Hash, Nonce and StateRoot.
func (block *Block) Verify(chainID uint32) error {
	if err := block.verifyHash(chainID); err != nil {
		return err
	}

	// begin state transaction.
	block.begin()

	if err := block.Execute(); err != nil {
		block.rollback()
		return err
	}

	if err := block.verifyState(); err != nil {
		block.rollback()
		return err
	}

	// commit.
	block.commit()

	return nil
}

// VerifyHash return hash verify result.
func (block *Block) verifyHash(chainID uint32) error {
	// check ChainID.
	if block.header.chainID != chainID {
		return ErrInvalidChainID
	}

	// verify hash.
	wantedHash := HashBlock(block)
	if !wantedHash.Equals(block.Hash()) {
		return ErrInvalidBlockHash
	}

	// verify transactions, check ChainID, hash & sign
	for _, tx := range block.transactions {
		if err := tx.Verify(block.header.chainID); err != nil {
			return err
		}
	}

	return nil
}

// verifyState return state verify result.
func (block *Block) verifyState() error {
	// verify state root.
	if !byteutils.Equal(block.accState.RootHash(), block.StateRoot()) {
		return ErrInvalidBlockStateRoot
	}

	// verify transaction root.
	if !byteutils.Equal(block.txsTrie.RootHash(), block.TxsRoot()) {
		return ErrInvalidBlockTxsRoot
	}

	return nil
}

// Execute block and return result.
func (block *Block) Execute() error {
	// execute transactions.
	for _, tx := range block.transactions {
		giveback, err := block.executeTransaction(tx)
		if giveback {
			block.txPool.Push(tx)
		}
		if err != nil {
			return err
		}
	}

	block.rewardCoinbase()

	return nil
}

// GetBalance returns balance for the given address on this block.
func (block *Block) GetBalance(address byteutils.Hash) *util.Uint128 {
	return block.accState.GetOrCreateUserAccount(address).Balance()
}

// GetNonce returns nonce for the given address on this block.
func (block *Block) GetNonce(address byteutils.Hash) uint64 {
	return block.accState.GetOrCreateUserAccount(address).Nonce()
}

func (block *Block) rewardCoinbase() {
	coinbaseAddr := block.header.coinbase.address
	coinbaseAcc := block.accState.GetOrCreateUserAccount(coinbaseAddr)
	coinbaseAcc.AddBalance(BlockReward)
}

// GetTransaction from txs Trie
func (block *Block) GetTransaction(hash byteutils.Hash) (*Transaction, error) {
	txBytes, err := block.txsTrie.Get(hash)
	if err != nil {
		return nil, err
	}
	pbTx := new(corepb.Transaction)
	if err := proto.Unmarshal(txBytes, pbTx); err != nil {
		return nil, err
	}

	tx := new(Transaction)
	if err = tx.FromProto(pbTx); err != nil {
		return nil, err
	}
	return tx, nil
}

// save tx in txs Trie
func (block *Block) acceptTransaction(tx *Transaction) error {
	pbTx, err := tx.ToProto()
	if err != nil {
		return err
	}
	txBytes, err := proto.Marshal(pbTx)
	if err != nil {
		return err
	}
	if _, err := block.txsTrie.Put(tx.hash, txBytes); err != nil {
		return err
	}
	return nil
}

// executeTransaction in block
func (block *Block) executeTransaction(tx *Transaction) (giveback bool, err error) {
	// check duplication
	if proof, _ := block.txsTrie.Prove(tx.hash); proof != nil {
		return false, ErrDuplicatedTransaction
	}

	// check nonce
	fromAcc := block.accState.GetOrCreateUserAccount(tx.from.address)
	if tx.nonce < fromAcc.Nonce()+1 {
		return false, ErrSmallTransactionNonce
	} else if tx.nonce > fromAcc.Nonce()+1 {
		return true, ErrLargeTransactionNonce
	}

	// execute.
	if err := tx.Execute(block); err != nil {
		return false, err
	}

	// save txs info in txs trie
	if err := block.acceptTransaction(tx); err != nil {
		return false, err
	}

	return false, nil
}

// HashBlock return the hash of block.
func HashBlock(block *Block) byteutils.Hash {
	hasher := sha3.New256()

	hasher.Write(block.header.parentHash)
	hasher.Write(block.header.stateRoot)
	hasher.Write(block.header.txsRoot)

	hasher.Write(block.header.curDynastyRoot)
	hasher.Write(block.header.nextDynastyRoot)
	hasher.Write(block.header.dynastyCandidatesRoot)
	hasher.Write(block.header.validatorsRoot)
	hasher.Write(block.header.depositRoot)
	hasher.Write(block.header.prepareVotesRoot)
	hasher.Write(block.header.heightPrepareVotesRoot)
	hasher.Write(block.header.commitVotesRoot)
	hasher.Write(block.header.changeVotesRoot)
	hasher.Write(block.header.abdicateVotesRoot)

	hasher.Write(byteutils.FromUint64(block.header.nonce))
	hasher.Write(block.header.coinbase.address)
	hasher.Write(byteutils.FromInt64(block.header.timestamp))
	hasher.Write(byteutils.FromUint32(block.header.chainID))

	for _, tx := range block.transactions {
		hasher.Write(tx.Hash())
	}

	return hasher.Sum(nil)
}

// LoadBlockFromStorage return a block from storage
func LoadBlockFromStorage(hash byteutils.Hash, storage storage.Storage, txPool *TransactionPool) (*Block, error) {
	value, err := storage.Get(hash)
	if err != nil {
		return nil, err
	}
	pbBlock := new(corepb.Block)
	block := new(Block)
	if err = proto.Unmarshal(value, pbBlock); err != nil {
		return nil, err
	}
	if err = block.FromProto(pbBlock); err != nil {
		return nil, err
	}

	if block.accState, err = state.NewAccountState(block.StateRoot(), storage); err != nil {
		return nil, err
	}
	if block.txsTrie, err = trie.NewBatchTrie(block.TxsRoot(), storage); err != nil {
		return nil, err
	}

	if block.curDynastyTrie, err = trie.NewBatchTrie(block.CurDynastyRoot(), storage); err != nil {
		return nil, err
	}
	if block.nextDynastyTrie, err = trie.NewBatchTrie(block.NextDynastyRoot(), storage); err != nil {
		return nil, err
	}
	if block.dynastyCandidatesTrie, err = trie.NewBatchTrie(block.DynastyCandidatesRoot(), storage); err != nil {
		return nil, err
	}
	if block.validatorsTrie, err = trie.NewBatchTrie(block.ValidatorsRoot(), storage); err != nil {
		return nil, err
	}
	if block.depositTrie, err = trie.NewBatchTrie(block.DepositRoot(), storage); err != nil {
		return nil, err
	}
	if block.prepareVotesTrie, err = trie.NewBatchTrie(block.PrepareVotesRoot(), storage); err != nil {
		return nil, err
	}
	if block.heightPrepareVotesTrie, err = trie.NewBatchTrie(block.HeightPrepareVotesRoot(), storage); err != nil {
		return nil, err
	}
	if block.commitVotesTrie, err = trie.NewBatchTrie(block.CommitVotesRoot(), storage); err != nil {
		return nil, err
	}
	if block.changeVotesTrie, err = trie.NewBatchTrie(block.ChangeVotesRoot(), storage); err != nil {
		return nil, err
	}
	if block.abdicateVotesTrie, err = trie.NewBatchTrie(block.AbdicateVotesRoot(), storage); err != nil {
		return nil, err
	}

	block.txPool = txPool
	block.storage = storage
	block.sealed = true
	return block, nil
}

// EraseBlockFromStorage remove a block from local storage
func EraseBlockFromStorage(block *Block) error {
	return block.storage.Del(block.Hash())
}
