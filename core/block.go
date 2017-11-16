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

	// the hash of the parent of the block where dynasty works
	dynastyParentHash byteutils.Hash
	// validators in current dynasty
	dynastyRoot byteutils.Hash
	// validators in next dynasty
	nextDynastyRoot byteutils.Hash
	// candidates in second next dynasty
	dynastyCandidatesRoot byteutils.Hash
	// deposit charged by validators
	depositRoot byteutils.Hash
	// all prepare votes
	prepareVotesRoot byteutils.Hash
	// all commit votes
	commitVotesRoot byteutils.Hash
	// all change votes
	changeVotesRoot byteutils.Hash
	// all abdicate votes
	abdicateVotesRoot byteutils.Hash
	// reorganize all blocks based on their height
	blocksHeightRoot byteutils.Hash

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

		DynastyParentHash:     b.dynastyParentHash,
		DynastyRoot:           b.dynastyRoot,
		NextDynastyRoot:       b.nextDynastyRoot,
		DynastyCandidatesRoot: b.dynastyCandidatesRoot,
		DepositRoot:           b.depositRoot,
		PrepareVotesRoot:      b.prepareVotesRoot,
		CommitVotesRoot:       b.commitVotesRoot,
		ChangeVotesRoot:       b.changeVotesRoot,
		AbdicateVotesRoot:     b.abdicateVotesRoot,
		BlocksHeightRoot:      b.blocksHeightRoot,

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

		b.dynastyParentHash = msg.DynastyParentHash
		b.dynastyRoot = msg.DynastyRoot
		b.nextDynastyRoot = msg.NextDynastyRoot
		b.dynastyCandidatesRoot = msg.DynastyCandidatesRoot
		b.depositRoot = msg.DepositRoot
		b.prepareVotesRoot = msg.PrepareVotesRoot
		b.commitVotesRoot = msg.CommitVotesRoot
		b.changeVotesRoot = msg.ChangeVotesRoot
		b.abdicateVotesRoot = msg.AbdicateVotesRoot
		b.blocksHeightRoot = msg.BlocksHeightRoot

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

	sealed       bool
	height       uint64
	parenetBlock *Block
	accState     state.AccountState
	txsTrie      *trie.BatchTrie

	dynastyTrie           *trie.BatchTrie // key: addr
	nextDynastyTrie       *trie.BatchTrie // key: addr
	dynastyCandidatesTrie *trie.BatchTrie // key: addr
	depositTrie           *trie.BatchTrie // key: addr
	prepareVotesTrie      *trie.BatchTrie // key: block hash + addr
	commitVotesTrie       *trie.BatchTrie // key: block hash + addr
	changeVotesTrie       *trie.BatchTrie // key: block hash + N + addr
	abdicateVotesTrie     *trie.BatchTrie // key: dynasty parent hash + addr
	blocksHeightTrie      *trie.BatchTrie // key: height

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
func NewBlock(chainID uint32, coinbase *Address, parent *Block, txPool *TransactionPool, storage storage.Storage) *Block {
	accState, _ := parent.accState.Clone()
	txsTrie, _ := parent.txsTrie.Clone()

	dynastyTrie, _ := parent.dynastyTrie.Clone()
	nextDynastyTrie, _ := parent.nextDynastyTrie.Clone()
	dynastyCandidatesTrie, _ := parent.dynastyCandidatesTrie.Clone()
	depositTrie, _ := parent.depositTrie.Clone()
	prepareVotesTrie, _ := parent.prepareVotesTrie.Clone()
	commitVotesTrie, _ := parent.commitVotesTrie.Clone()
	changeVotesTrie, _ := parent.changeVotesTrie.Clone()
	abdicateVotesTrie, _ := parent.abdicateVotesTrie.Clone()
	blocksHeightTrie, _ := parent.blocksHeightTrie.Clone()

	block := &Block{
		header: &BlockHeader{
			parentHash:        parent.Hash(),
			dynastyParentHash: parent.DynastyParentHash(),
			coinbase:          coinbase,
			nonce:             0,
			timestamp:         time.Now().Unix(),
			chainID:           chainID,
		},
		transactions: make(Transactions, 0),
		parenetBlock: parent,
		accState:     accState,
		txsTrie:      txsTrie,

		dynastyTrie:           dynastyTrie,
		nextDynastyTrie:       nextDynastyTrie,
		dynastyCandidatesTrie: dynastyCandidatesTrie,
		depositTrie:           depositTrie,
		prepareVotesTrie:      prepareVotesTrie,
		commitVotesTrie:       commitVotesTrie,
		changeVotesTrie:       changeVotesTrie,
		abdicateVotesTrie:     abdicateVotesTrie,
		blocksHeightTrie:      blocksHeightTrie,

		txPool:  txPool,
		height:  parent.Height() + 1,
		sealed:  false,
		storage: storage,
	}

	if block.checkDynastyRuleEpochOver() {
		block.changeDynasty()
		block.header.dynastyParentHash = block.header.parentHash
	}

	return block
}

// return validators in the given dynasty
func validators(dynasty *trie.BatchTrie) map[byteutils.HexHash]bool {
	iterator, _ := dynasty.Iterator(nil)
	validators := make(map[byteutils.HexHash]bool)
	for iterator.Next() {
		validators[byteutils.HexHash(byteutils.Hex(iterator.Value()))] = true
	}
	return validators
}

// return all candidates in second next dynasty
func (block *Block) candidates() []byteutils.Hash {
	iterator, _ := block.dynastyCandidatesTrie.Iterator(nil)
	candidates := []byteutils.Hash{}
	for iterator.Next() {
		candidates = append(candidates, byteutils.Hash(byteutils.Hex(iterator.Value())))
	}
	return candidates
}

// dynasty rule: epoch over, create > 100 blocks
func (block *Block) checkDynastyRuleEpochOver() bool {
	parentBlock, err := LoadBlockFromStorage(block.header.dynastyParentHash, block.storage, block.txPool)
	if err != nil {
		panic("cannot find the birth place of the block's dynasty")
	}
	if block.height > parentBlock.height && block.height-parentBlock.height > EpochSize {
		return true
	}
	return false
}

// dynasty rule: remove too many validators, remove > 1/3 * N
func (block *Block) checkDynastyRuleTooFewValidators() bool {
	parentBlock, err := LoadBlockFromStorage(block.header.dynastyParentHash, block.storage, block.txPool)
	if err != nil {
		panic("cannot find the birth place of the block's dynasty")
	}
	if len(validators(block.dynastyTrie)) < 2/3*len(validators(parentBlock.nextDynastyTrie)) {
		return true
	}
	return false
}

// change current dynasty to next one
func (block *Block) changeDynasty() {
	block.dynastyTrie = block.nextDynastyTrie
	// select DynastySize validators from candidates
	vs := &ValidatorSort{validators: block.candidates(), seed: block.height}
	sort.Sort(vs)
	// TODO(roy): adjust the dynastysize dynamically
	// TODO(roy): limit the re-election validators
	count := 0
	for _, v := range vs.validators {
		if count < DynastySize {
			block.nextDynastyTrie.Put(v, v)
			// un-selected candidates will login automatically
			block.dynastyCandidatesTrie.Del(v)
			count++
		}
	}
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

// DynastyParentHash return block's dynasty's parent hash.
func (block *Block) DynastyParentHash() byteutils.Hash {
	return block.header.dynastyParentHash
}

// StateRoot return state root hash.
func (block *Block) StateRoot() byteutils.Hash {
	return block.header.stateRoot
}

// TxsRoot return txs root hash.
func (block *Block) TxsRoot() byteutils.Hash {
	return block.header.txsRoot
}

// DynastyRoot return dynasty root hash.
func (block *Block) DynastyRoot() byteutils.Hash {
	return block.header.dynastyRoot
}

// NextDynastyRoot return next dynasty root hash.
func (block *Block) NextDynastyRoot() byteutils.Hash {
	return block.header.nextDynastyRoot
}

// DynastyCandidatesRoot return dynasty candidates root hash.
func (block *Block) DynastyCandidatesRoot() byteutils.Hash {
	return block.header.dynastyCandidatesRoot
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

// BlocksHeightRoot return blocks height root hash.
func (block *Block) BlocksHeightRoot() byteutils.Hash {
	return block.header.blocksHeightRoot
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

	block.dynastyTrie, _ = parentBlock.dynastyTrie.Clone()
	block.nextDynastyTrie, _ = parentBlock.nextDynastyTrie.Clone()
	block.dynastyCandidatesTrie, _ = parentBlock.dynastyCandidatesTrie.Clone()
	block.depositTrie, _ = parentBlock.depositTrie.Clone()
	block.prepareVotesTrie, _ = parentBlock.prepareVotesTrie.Clone()
	block.commitVotesTrie, _ = parentBlock.commitVotesTrie.Clone()
	block.changeVotesTrie, _ = parentBlock.changeVotesTrie.Clone()
	block.abdicateVotesTrie, _ = parentBlock.abdicateVotesTrie.Clone()
	block.blocksHeightTrie, _ = parentBlock.blocksHeightTrie.Clone()

	block.txPool = parentBlock.txPool
	block.parenetBlock = parentBlock
	block.storage = parentBlock.storage

	// travel to calculate block height.
	depth := uint64(0)
	ancestorHeight := uint64(0)
	for ancestor := block; ancestor != nil; ancestor = ancestor.parenetBlock {
		depth++
		ancestorHeight = ancestor.height
		if ancestor.height > 0 {
			break
		}
	}

	for ancestor := block; ancestor != nil && depth > 1; ancestor = ancestor.parenetBlock {
		depth--
		ancestor.height = ancestorHeight + depth
	}

	if block.checkDynastyRuleEpochOver() {
		block.changeDynasty()
	}

	return true
}

func (block *Block) begin() {
	log.Info("Block Begin.")

	block.accState.BeginBatch()
	block.txsTrie.BeginBatch()

	block.dynastyTrie.BeginBatch()
	block.nextDynastyTrie.BeginBatch()
	block.dynastyCandidatesTrie.BeginBatch()
	block.depositTrie.BeginBatch()
	block.prepareVotesTrie.BeginBatch()
	block.commitVotesTrie.BeginBatch()
	block.changeVotesTrie.BeginBatch()
	block.abdicateVotesTrie.BeginBatch()
	block.blocksHeightTrie.BeginBatch()
}

func (block *Block) commit() {
	block.accState.Commit()
	block.txsTrie.Commit()

	block.dynastyTrie.Commit()
	block.nextDynastyTrie.Commit()
	block.dynastyCandidatesTrie.Commit()
	block.depositTrie.Commit()
	block.prepareVotesTrie.Commit()
	block.commitVotesTrie.Commit()
	block.changeVotesTrie.Commit()
	block.abdicateVotesTrie.Commit()
	block.blocksHeightTrie.Commit()

	log.WithFields(log.Fields{
		"block": block,
	}).Info("Block Commit.")
}

func (block *Block) rollback() {
	block.accState.RollBack()
	block.txsTrie.RollBack()

	block.dynastyTrie.RollBack()
	block.nextDynastyTrie.RollBack()
	block.dynastyCandidatesTrie.RollBack()
	block.depositTrie.RollBack()
	block.prepareVotesTrie.RollBack()
	block.commitVotesTrie.RollBack()
	block.changeVotesTrie.RollBack()
	block.abdicateVotesTrie.RollBack()
	block.blocksHeightTrie.RollBack()

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

	block.header.dynastyRoot = block.dynastyTrie.RootHash()
	block.header.nextDynastyRoot = block.nextDynastyTrie.RootHash()
	block.header.dynastyCandidatesRoot = block.dynastyCandidatesTrie.RootHash()
	block.header.depositRoot = block.depositTrie.RootHash()
	block.header.prepareVotesRoot = block.prepareVotesTrie.RootHash()
	block.header.commitVotesRoot = block.commitVotesTrie.RootHash()
	block.header.changeVotesRoot = block.changeVotesTrie.RootHash()
	block.header.abdicateVotesRoot = block.abdicateVotesTrie.RootHash()
	block.header.blocksHeightRoot = block.blocksHeightTrie.RootHash()

	block.header.hash = HashBlock(block)
	block.sealed = true
}

func (block *Block) String() string {
	return fmt.Sprintf(`Block %p { 
		height:%d; hash:%s; parentHash:%s; stateRoot:%s; txsRoot: %s; 
		dynastyRoot: %s; nextDynastyRoot: %s; dynastyCandidatesRoot: %s; depositRoot: %s; 
		prepareVotesRoot: %s; commitVotesRoot: %s; changeVotesRoot: %s; 
		abdicateVotesRoot: %s; blocksHeightRoot: %s;
		nonce:%d; timestamp: %d}`,
		block,
		block.height,
		byteutils.Hex(block.header.hash),
		byteutils.Hex(block.header.parentHash),
		byteutils.Hex(block.StateRoot()),
		byteutils.Hex(block.TxsRoot()),

		byteutils.Hex(block.DynastyRoot()),
		byteutils.Hex(block.NextDynastyRoot()),
		byteutils.Hex(block.DynastyCandidatesRoot()),
		byteutils.Hex(block.DepositRoot()),
		byteutils.Hex(block.PrepareVotesRoot()),
		byteutils.Hex(block.CommitVotesRoot()),
		byteutils.Hex(block.ChangeVotesRoot()),
		byteutils.Hex(block.AbdicateVotesRoot()),
		byteutils.Hex(block.BlocksHeightRoot()),

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
	hasher.Write(block.header.dynastyParentHash)
	hasher.Write(block.header.stateRoot)
	hasher.Write(block.header.txsRoot)

	hasher.Write(block.header.dynastyRoot)
	hasher.Write(block.header.nextDynastyRoot)
	hasher.Write(block.header.dynastyCandidatesRoot)
	hasher.Write(block.header.depositRoot)
	hasher.Write(block.header.prepareVotesRoot)
	hasher.Write(block.header.commitVotesRoot)
	hasher.Write(block.header.changeVotesRoot)
	hasher.Write(block.header.abdicateVotesRoot)
	hasher.Write(block.header.blocksHeightRoot)

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

	if block.dynastyTrie, err = trie.NewBatchTrie(block.DynastyRoot(), storage); err != nil {
		return nil, err
	}
	if block.nextDynastyTrie, err = trie.NewBatchTrie(block.NextDynastyRoot(), storage); err != nil {
		return nil, err
	}
	if block.dynastyCandidatesTrie, err = trie.NewBatchTrie(block.DynastyCandidatesRoot(), storage); err != nil {
		return nil, err
	}
	if block.depositTrie, err = trie.NewBatchTrie(block.DepositRoot(), storage); err != nil {
		return nil, err
	}
	if block.prepareVotesTrie, err = trie.NewBatchTrie(block.PrepareVotesRoot(), storage); err != nil {
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
	if block.blocksHeightTrie, err = trie.NewBatchTrie(block.BlocksHeightRoot(), storage); err != nil {
		return nil, err
	}

	block.txPool = txPool
	block.storage = storage
	return block, nil
}
