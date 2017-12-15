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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nebulasio/go-nebulas/crypto/keystore"

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

var (
	// BlockHashLength define a const of the length of Hash of Block in byte.
	BlockHashLength = 32

	// BlockReward given to coinbase, default is:16*(10**18)
	// TODO: block reward should calculates dynamic.
	BlockReward = util.NewUint128FromBigInt(util.NewUint128().Mul(util.NewUint128FromInt(16).Int,
		util.NewUint128().Exp(util.NewUint128FromInt(10).Int, util.NewUint128FromInt(18).Int, nil)))
)

// BlockHeader of a block
type BlockHeader struct {
	hash       byteutils.Hash
	parentHash byteutils.Hash

	// world state
	stateRoot   byteutils.Hash
	txsRoot     byteutils.Hash
	eventsRoot  byteutils.Hash
	dposContext *corepb.DposContext

	coinbase  *Address
	nonce     uint64
	timestamp int64
	chainID   uint32

	// sign
	alg  uint8
	sign byteutils.Hash
}

// ToProto converts domain BlockHeader to proto BlockHeader
func (b *BlockHeader) ToProto() (proto.Message, error) {
	return &corepb.BlockHeader{
		Hash:        b.hash,
		ParentHash:  b.parentHash,
		StateRoot:   b.stateRoot,
		TxsRoot:     b.txsRoot,
		EventsRoot:  b.eventsRoot,
		DposContext: b.dposContext,
		Nonce:       b.nonce,
		Coinbase:    b.coinbase.address,
		Timestamp:   b.timestamp,
		ChainId:     b.chainID,
		Alg:         uint32(b.alg),
		Sign:        b.sign,
	}, nil
}

// FromProto converts proto BlockHeader to domain BlockHeader
func (b *BlockHeader) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.BlockHeader); ok {
		b.hash = msg.Hash
		b.parentHash = msg.ParentHash
		b.stateRoot = msg.StateRoot
		b.txsRoot = msg.TxsRoot
		b.eventsRoot = msg.EventsRoot
		b.dposContext = msg.DposContext
		b.nonce = msg.Nonce
		b.coinbase = &Address{msg.Coinbase}
		b.timestamp = msg.Timestamp
		b.chainID = msg.ChainId
		b.alg = uint8(msg.Alg)
		b.sign = msg.Sign
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
	eventsTrie   *trie.BatchTrie
	dposContext  *DposContext
	txPool       *TransactionPool
	miner        *Address

	storage      storage.Storage
	eventEmitter *EventEmitter
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

// SerializeTxByHash returns tx serialized bytes
func (block *Block) SerializeTxByHash(hash byteutils.Hash) (proto.Message, error) {
	tx, err := block.GetTransaction(hash)
	if err != nil {
		return nil, err
	}
	return tx.ToProto()
}

// NewBlock return new block.
func NewBlock(chainID uint32, coinbase *Address, parent *Block) (*Block, error) {
	accState, err := parent.accState.Clone()
	if err != nil {
		return nil, err
	}
	txsTrie, err := parent.txsTrie.Clone()
	if err != nil {
		return nil, err
	}
	eventsTrie, err := parent.eventsTrie.Clone()
	if err != nil {
		return nil, err
	}
	dposContext, err := parent.dposContext.Clone()
	if err != nil {
		return nil, err
	}
	block := &Block{
		header: &BlockHeader{
			parentHash:  parent.Hash(),
			dposContext: &corepb.DposContext{},
			coinbase:    coinbase,
			nonce:       0,
			timestamp:   time.Now().Unix(),
			chainID:     chainID,
		},
		transactions: make(Transactions, 0),
		parenetBlock: parent,
		accState:     accState,
		txsTrie:      txsTrie,
		eventsTrie:   eventsTrie,
		dposContext:  dposContext,
		txPool:       parent.txPool,
		height:       parent.height + 1,
		sealed:       false,
		storage:      parent.storage,
		eventEmitter: parent.eventEmitter,
	}
	return block, nil
}

// Sign sign transaction,sign algorithm is
func (block *Block) Sign(signature keystore.Signature) error {
	sign, err := signature.Sign(block.header.hash)
	if err != nil {
		return err
	}
	block.header.alg = uint8(signature.Algorithm())
	block.header.sign = sign
	return nil
}

// Coinbase return block's coinbase
func (block *Block) Coinbase() *Address {
	return block.header.coinbase
}

// Alg return block's alg
func (block *Block) Alg() uint8 {
	return block.header.alg
}

// Signature return block's signature
func (block *Block) Signature() byteutils.Hash {
	return block.header.sign
}

// CoinbaseHash return block's coinbase hash
func (block *Block) CoinbaseHash() byteutils.Hash {
	return block.header.coinbase.address
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

// Timestamp return timestamp
func (block *Block) Timestamp() int64 {
	return block.header.timestamp
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

// EventsRoot return events root hash.
func (block *Block) EventsRoot() byteutils.Hash {
	return block.header.eventsRoot
}

// DposContext return dpos context
func (block *Block) DposContext() *corepb.DposContext {
	return block.header.dposContext
}

// DposContextHash hash dpos context
func (block *Block) DposContextHash() byteutils.Hash {
	hasher := sha3.New256()

	hasher.Write(block.header.dposContext.DynastyRoot)
	hasher.Write(block.header.dposContext.NextDynastyRoot)
	hasher.Write(block.header.dposContext.DelegateRoot)
	hasher.Write(block.header.dposContext.VoteRoot)
	hasher.Write(block.header.dposContext.CandidateRoot)
	hasher.Write(block.header.dposContext.MintCntRoot)

	return hasher.Sum(nil)
}

// ParentHash return parent hash.
func (block *Block) ParentHash() byteutils.Hash {
	return block.header.parentHash
}

// ParentBlock return the parent block.
func (block *Block) ParentBlock() (*Block, error) {
	if block.parenetBlock != nil {
		return block.parenetBlock, nil
	}
	parentBlock, err := LoadBlockFromStorage(block.ParentHash(), block.storage, block.txPool, block.eventEmitter)
	if err != nil {
		log.WithFields(log.Fields{
			"func":  "block.ParentBlock",
			"block": block,
			"err":   err,
		}).Error("cannot find the parent block.")
		return nil, err
	}
	return parentBlock, nil
}

// Height return height
func (block *Block) Height() uint64 {
	return block.height
}

// Miner return miner
func (block *Block) Miner() *Address {
	return block.miner
}

// SetMiner return miner
func (block *Block) SetMiner(miner *Address) {
	block.miner = miner
}

// VerifyAddress returns if the addr string is valid
func (block *Block) VerifyAddress(str string) bool {
	_, err := AddressParse(str)
	return err == nil
}

// LinkParentBlock link parent block, return true if hash is the same; false otherwise.
func (block *Block) LinkParentBlock(parentBlock *Block) bool {
	if block.ParentHash().Equals(parentBlock.Hash()) == false {
		return false
	}

	var err error
	if block.accState, err = parentBlock.accState.Clone(); err != nil {
		log.WithFields(log.Fields{
			"func":  "block.LinkParentBlock",
			"block": parentBlock,
			"err":   err,
		}).Error("cannot clone account state.")
		return false
	}
	if block.txsTrie, err = parentBlock.txsTrie.Clone(); err != nil {
		log.WithFields(log.Fields{
			"func":  "block.LinkParentBlock",
			"block": parentBlock,
			"err":   err,
		}).Error("cannot clone txs state.")
		return false
	}
	if block.eventsTrie, err = parentBlock.eventsTrie.Clone(); err != nil {
		log.WithFields(log.Fields{
			"func":  "block.LinkParentBlock",
			"block": parentBlock,
			"err":   err,
		}).Error("cannot clone events state.")
		return false
	}
	elapsedSecond := block.Timestamp() - parentBlock.Timestamp()
	context, err := parentBlock.NextDynastyContext(elapsedSecond)
	if err != nil {
		log.WithFields(log.Fields{
			"func":  "block.LinkParentBlock",
			"block": parentBlock,
			"err":   err,
		}).Error("calculate next dynasty context.")
		return false
	}
	block.LoadDynastyContext(context)
	block.txPool = parentBlock.txPool
	block.parenetBlock = parentBlock
	block.storage = parentBlock.storage
	block.height = parentBlock.height + 1
	block.eventEmitter = parentBlock.eventEmitter

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
	log.Info("Block Begin.")
	block.accState.BeginBatch()
	block.txsTrie.BeginBatch()
	block.eventsTrie.BeginBatch()
	block.dposContext.BeginBatch()
}

func (block *Block) commit() {
	block.accState.Commit()
	block.txsTrie.Commit()
	block.eventsTrie.Commit()
	block.dposContext.Commit()
	log.WithFields(log.Fields{
		"block": block,
	}).Info("Block Commit.")
}

func (block *Block) rollback() {
	block.accState.RollBack()
	block.txsTrie.RollBack()
	block.eventsTrie.RollBack()
	block.dposContext.RollBack()
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
func (block *Block) Seal() error {
	if block.sealed {
		return ErrDoubleSealBlock
	}

	block.begin()
	block.rewardCoinbase()
	err := block.recordMintCnt()
	if err != nil {
		block.rollback()
		return err
	}
	block.commit()
	block.header.stateRoot = block.accState.RootHash()
	block.header.txsRoot = block.txsTrie.RootHash()
	block.header.eventsRoot = block.eventsTrie.RootHash()
	if block.header.dposContext, err = block.dposContext.ToProto(); err != nil {
		return err
	}
	block.header.hash = HashBlock(block)
	block.sealed = true

	log.WithFields(log.Fields{
		"block": block,
	}).Info("Block Seal.")
	return nil
}

func (block *Block) String() string {
	return fmt.Sprintf("Block %p {height:%d; hash:%s; parentHash:%s; accState: %s; nonce:%d, timestamp: %d; coinbase: %s}",
		block,
		block.height,
		byteutils.Hex(block.header.hash),
		byteutils.Hex(block.header.parentHash),
		byteutils.Hex(block.header.stateRoot),
		block.header.nonce,
		block.header.timestamp,
		block.header.coinbase.String(),
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

	// trigger event.
	for _, v := range block.transactions {
		var topic string
		switch v.Type() {
		case TxPayloadBinaryType:
			topic = TopicSendTransaction
		case TxPayloadDeployType:
			topic = TopicDeploySmartContract
		case TxPayloadCallType:
			topic = TopicCallSmartContract
		case TxPayloadDelegateType:
			topic = TopicDelegate
		case TxPayloadCandidateType:
			topic = TopicCandidate
		}
		e := &Event{
			Topic: topic,
			Data:  v.String(),
		}
		block.eventEmitter.Trigger(e)
	}

	e := &Event{
		Topic: TopicLinkBlock,
		Data:  block.String(),
	}
	block.eventEmitter.Trigger(e)

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
	log.Info(block.accState.RootHash())
	log.Info(block.StateRoot())
	if !byteutils.Equal(block.accState.RootHash(), block.StateRoot()) {
		return ErrInvalidBlockStateRoot
	}

	// verify transaction root.
	if !byteutils.Equal(block.txsTrie.RootHash(), block.TxsRoot()) {
		return ErrInvalidBlockTxsRoot
	}

	// verify events root.
	if !byteutils.Equal(block.eventsTrie.RootHash(), block.EventsRoot()) {
		return ErrInvalidBlockEventsRoot
	}

	// verify transaction root.
	if !byteutils.Equal(block.dposContext.RootHash(), block.DposContextHash()) {
		return ErrInvalidBlockDposContextRoot
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
	return block.recordMintCnt()
}

// GetBalance returns balance for the given address on this block.
func (block *Block) GetBalance(address byteutils.Hash) *util.Uint128 {
	return block.accState.GetOrCreateUserAccount(address).Balance()
}

// GetNonce returns nonce for the given address on this block.
func (block *Block) GetNonce(address byteutils.Hash) uint64 {
	return block.accState.GetOrCreateUserAccount(address).Nonce()
}

func (block *Block) recordEvent(tx *Transaction, event *Event) error {
	iter, err := block.eventsTrie.Iterator(tx.Hash())
	if err != nil && err != storage.ErrKeyNotFound {
		return err
	}
	cnt := int64(0)
	if err != storage.ErrKeyNotFound {
		exist, err := iter.Next()
		if err != nil {
			return err
		}
		for exist {
			cnt++
			exist, err = iter.Next()
			if err != nil {
				return err
			}
		}
	}
	cnt++
	key := append(tx.Hash(), byteutils.FromInt64(cnt)...)
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = block.eventsTrie.Put(key, bytes)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"block": block,
		"tx":    tx,
		"event": event,
	}).Debug("Record Event.")
	return nil
}

func (block *Block) fetchEvents(tx *Transaction) ([]*Event, error) {
	events := []*Event{}
	iter, err := block.eventsTrie.Iterator(tx.Hash())
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	if err != storage.ErrKeyNotFound {
		exist, err := iter.Next()
		if err != nil {
			return nil, err
		}
		for exist {
			event := new(Event)
			err = json.Unmarshal(iter.Value(), event)
			if err != nil {
				return nil, err
			}
			events = append(events, event)
			exist, err = iter.Next()
			if err != nil {
				return nil, err
			}
		}
	}
	return events, nil
}

func (block *Block) recordMintCnt() error {
	key := append(byteutils.FromInt64(block.Timestamp()/DynastyInterval), block.miner.Bytes()...)
	bytes, err := block.dposContext.mintCntTrie.Get(key)
	if err != nil && err != storage.ErrKeyNotFound {
		return err
	}
	cnt := int64(0)
	if err != storage.ErrKeyNotFound {
		cnt = byteutils.Int64(bytes)
	}
	cnt++
	_, err = block.dposContext.mintCntTrie.Put(key, byteutils.FromInt64(cnt))
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"dynasty": block.Timestamp() / DynastyInterval,
		"miner":   block.miner.String(),
		"count":   cnt,
	}).Debug("Record Mint Count.")
	return nil
}

func (block *Block) rewardCoinbase() {
	coinbaseAddr := block.header.coinbase.address
	coinbaseAcc := block.accState.GetOrCreateUserAccount(coinbaseAddr)
	coinbaseAcc.AddBalance(BlockReward)
	log.WithFields(log.Fields{
		"coinbase": coinbaseAddr.Hex(),
		"balance":  coinbaseAcc.Balance().Int64(),
	}).Debug("Coinbase Reward.")
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
	if _, err := tx.Execute(block); err != nil {
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

	hasher.Write(block.ParentHash())
	hasher.Write(block.StateRoot())
	hasher.Write(block.TxsRoot())
	hasher.Write(block.EventsRoot())
	hasher.Write(block.DposContextHash())
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
func LoadBlockFromStorage(hash byteutils.Hash, storage storage.Storage, txPool *TransactionPool, eventEmitter *EventEmitter) (*Block, error) {
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
	block.accState, err = state.NewAccountState(block.StateRoot(), storage)
	if err != nil {
		return nil, err
	}
	block.txsTrie, err = trie.NewBatchTrie(block.TxsRoot(), storage)
	if err != nil {
		return nil, err
	}
	block.eventsTrie, err = trie.NewBatchTrie(block.EventsRoot(), storage)
	if err != nil {
		return nil, err
	}
	if block.dposContext, err = NewDposContext(storage); err != nil {
		return nil, err
	}
	if block.dposContext.FromProto(block.DposContext()) != nil {
		return nil, err
	}
	block.txPool = txPool
	block.storage = storage
	block.sealed = true
	block.eventEmitter = eventEmitter
	return block, nil
}
