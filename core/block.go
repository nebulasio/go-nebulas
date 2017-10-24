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
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/batch_trie"
	"github.com/nebulasio/go-nebulas/core/pb"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"

	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
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
)

// Account info in state Trie
type Account struct {
	Balance *util.Uint128
	Nonce   uint64
}

// ToProto converts domain Account to proto Account
func (acc *Account) ToProto() (proto.Message, error) {
	value, err := acc.Balance.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	return &corepb.Account{
		Balance: value,
		Nonce:   acc.Nonce,
	}, nil
}

// FromProto converts proto Account to domain Account
func (acc *Account) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.Account); ok {
		value, err := util.NewUint128FromFixedSizeByteSlice(msg.Balance)
		if err != nil {
			return err
		}
		acc.Balance = value
		acc.Nonce = msg.Nonce
		return nil
	}
	return errors.New("Pb Message cannot be converted into Account")
}

// BlockHeader of a block
type BlockHeader struct {
	hash       Hash
	parentHash Hash
	stateRoot  Hash
	txsRoot    Hash
	nonce      uint64
	coinbase   *Address
	timestamp  int64
	chainID    uint32
}

// ToProto converts domain BlockHeader to proto BlockHeader
func (b *BlockHeader) ToProto() (proto.Message, error) {
	return &corepb.BlockHeader{
		Hash:       b.hash,
		ParentHash: b.parentHash,
		StateRoot:  b.stateRoot,
		TxsRoot:    b.txsRoot,
		Nonce:      b.nonce,
		Coinbase:   b.coinbase.address,
		Timestamp:  b.timestamp,
		ChainId:    b.chainID,
	}, nil
}

// FromProto converts proto BlockHeader to domain BlockHeader
func (b *BlockHeader) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.BlockHeader); ok {
		b.hash = msg.Hash
		b.parentHash = msg.ParentHash
		b.stateRoot = msg.StateRoot
		b.txsRoot = msg.TxsRoot
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
	stateTrie    *batchtrie.BatchTrie
	txsTrie      *batchtrie.BatchTrie
	txPool       *TransactionPool
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
		return nil
	}
	return errors.New("Pb Message cannot be converted into Block")
}

// NewBlock return new block.
func NewBlock(chainID uint32, coinbase *Address, parent *Block, txPool *TransactionPool) *Block {
	stateTrie, _ := parent.stateTrie.Clone()
	txsTrie, _ := parent.txsTrie.Clone()
	block := &Block{
		header: &BlockHeader{
			parentHash: parent.Hash(),
			coinbase:   coinbase,
			nonce:      0,
			timestamp:  time.Now().Unix(),
			chainID:    chainID,
		},
		transactions: make(Transactions, 0),
		parenetBlock: parent,
		stateTrie:    stateTrie,
		txsTrie:      txsTrie,
		txPool:       txPool,
		height:       parent.height + 1,
		sealed:       false,
	}
	return block
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
func (block *Block) Hash() Hash {
	return block.header.hash
}

// StateRoot return state root hash.
func (block *Block) StateRoot() Hash {
	return block.header.stateRoot
}

// TxsRoot return txs root hash.
func (block *Block) TxsRoot() Hash {
	return block.header.txsRoot
}

// ParentHash return parent hash.
func (block *Block) ParentHash() Hash {
	return block.header.parentHash
}

// ParentBlock return parent block.
func (block *Block) ParentBlock() *Block {
	return block.parenetBlock
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

	block.stateTrie, _ = parentBlock.stateTrie.Clone()
	block.txsTrie, _ = parentBlock.txsTrie.Clone()
	block.txPool = parentBlock.txPool
	block.parenetBlock = parentBlock

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

	return true
}

func (block *Block) begin() {
	block.stateTrie.BeginBatch()
	block.txsTrie.BeginBatch()
}

func (block *Block) commit() {
	block.stateTrie.Commit()
	block.txsTrie.Commit()
}

func (block *Block) rollback() {
	block.stateTrie.RollBack()
	block.txsTrie.RollBack()
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
				"tx":       tx,
				"giveback": giveback,
			}).Info("tx is packed.")
			block.commit()
			block.transactions = append(block.transactions, tx)
			n--
		} else {
			log.WithFields(log.Fields{
				"func":     "block.CollectionTransactions",
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

	block.rewardCoinbase()
	block.header.hash = HashBlock(block)
	block.header.stateRoot = block.stateTrie.RootHash()
	block.header.txsRoot = block.txsTrie.RootHash()
	block.sealed = true
}

func (block *Block) String() string {
	return fmt.Sprintf("Block {height:%d; hash:%s; parentHash:%s; stateRoot:%s, nonce:%d, timestamp: %d}",
		block.height,
		byteutils.Hex(block.header.hash),
		byteutils.Hex(block.header.parentHash),
		byteutils.Hex(block.StateRoot()),
		block.header.nonce,
		block.header.timestamp,
	)
}

// Verify return block verify result, including Hash, Nonce and StateRoot.
func (block *Block) Verify() error {
	if err := block.verifyHash(); err != nil {
		return err
	}

	block.begin()
	if err := block.verifyTransactions(); err != nil {
		block.rollback()
		return err
	}
	block.commit()

	return nil
}

// VerifyHash return hash verify result.
func (block *Block) verifyHash() error {
	// verify hash.
	wantedHash := HashBlock(block)
	if !wantedHash.Equals(block.Hash()) {
		return ErrInvalidBlockHash
	}

	return nil
}

// VerifyTransactions return hash verify result.
func (block *Block) verifyTransactions() error {
	if err := block.executeTransactions(); err != nil {
		return err
	}

	block.rewardCoinbase()

	if !byteutils.Equal(block.stateTrie.RootHash(), block.StateRoot()) {
		return ErrInvalidBlockStateRoot
	}
	if !byteutils.Equal(block.txsTrie.RootHash(), block.TxsRoot()) {
		return ErrInvalidBlockTxsRoot
	}

	return nil
}

// GetBalance returns balance for the given address on this block.
func (block *Block) GetBalance(address Hash) *util.Uint128 {
	return block.FindAccount(&Address{address}).Balance
}

// GetNonce returns nonce for the given address on this block.
func (block *Block) GetNonce(address Hash) uint64 {
	return block.FindAccount(&Address{address}).Nonce
}

func (block *Block) rewardCoinbase() {
	coinbaseAddr := block.header.coinbase
	coinbaseAcc := block.FindAccount(coinbaseAddr)
	coinbaseAcc.Balance.Add(coinbaseAcc.Balance.Int, BlockReward.Int)
	block.saveAccount(coinbaseAddr, coinbaseAcc)
}

func (block *Block) executeTransactions() error {
	for _, tx := range block.transactions {
		giveback, err := block.executeTransaction(tx)
		if giveback {
			block.txPool.Push(tx)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// FindAccount return account info in state Trie
// if not found, return a new account
func (block *Block) FindAccount(address *Address) *Account {
	acc := new(Account)
	if accBytes, err := block.stateTrie.Get(address.address); err != nil {
		// new account
		acc.Balance = util.NewUint128()
		acc.Nonce = 0
	} else {
		pbAcc := new(corepb.Account)
		if err := proto.Unmarshal(accBytes, pbAcc); err != nil {
			panic("account in stateTrie cannot be unmarshaled correctly")
		}
		value, err := util.NewUint128FromFixedSizeByteSlice(pbAcc.Balance)
		if err != nil {
			panic("account's balance in stateTrie, convert []byte to uint128 failed")
		}
		acc.Balance = value
		acc.Nonce = pbAcc.Nonce
	}
	return acc
}

// saveAccount update account info in state Trie
func (block *Block) saveAccount(address *Address, acc *Account) {
	pbAcc, _ := acc.ToProto()
	accBytes, _ := proto.Marshal(pbAcc)
	block.stateTrie.Put(address.address, accBytes)
}

// saveTxs record tx in txs Trie
func (block *Block) saveTransaction(tx *Transaction) {
	pbTx, err := tx.ToProto()
	if err != nil {
		panic("tx cannot be converted into []byte")
	}
	txBytes, err := proto.Marshal(pbTx)
	if err != nil {
		panic("tx cannot be converted into []byte")
	}
	if _, err := block.txsTrie.Put(tx.hash, txBytes); err != nil {
		panic("tx cannot be put into txs trie")
	}
}

// executeTransaction in block
// 0. check chainID
// 1. check hash
// 2. check sign
// 3. check duplication
// 4. check nonce increase by 1
// 5. check balance
func (block *Block) executeTransaction(tx *Transaction) (giveback bool, err error) {
	// check chainID
	if tx.chainID != block.header.chainID {
		return false, errors.New("invalid transaction chainID")
	}
	// check hash & sign
	if err := tx.Verify(); err != nil {
		return false, err
	}
	// check duplication
	if proof, _ := block.txsTrie.Prove(tx.hash); proof != nil {
		return false, errors.New("cannot execute an existed transaction")
	}
	// check nonce
	fromAcc := block.FindAccount(tx.from)
	if tx.nonce < fromAcc.Nonce+1 {
		return false, errors.New("cannot accept a transaction with smaller nonce")
	} else if tx.nonce > fromAcc.Nonce+1 {
		return true, errors.New("cannot accept a transaction with too bigger nonce")
	}
	// check balance
	toAcc := block.FindAccount(tx.to)
	if fromAcc.Balance.Cmp(tx.value.Int) < 0 {
		return false, ErrInsufficientBalance
	}
	// accept the transaction
	fromAcc.Balance.Sub(fromAcc.Balance.Int, tx.value.Int)
	toAcc.Balance.Add(toAcc.Balance.Int, tx.value.Int)
	fromAcc.Nonce++

	// save account info in state trie
	block.saveAccount(tx.from, fromAcc)
	block.saveAccount(tx.to, toAcc)
	// save txs info in txs trie
	block.saveTransaction(tx)

	return false, nil
}

// HashBlock return the hash of block.
func HashBlock(block *Block) Hash {
	hasher := sha3.New256()

	hasher.Write(block.header.parentHash)
	hasher.Write(byteutils.FromUint64(block.header.nonce))
	hasher.Write(block.header.coinbase.address)
	hasher.Write(byteutils.FromInt64(block.header.timestamp))
	hasher.Write(byteutils.FromUint32(block.header.chainID))

	for _, tx := range block.transactions {
		hasher.Write(tx.Hash())
	}

	return hasher.Sum(nil)
}
