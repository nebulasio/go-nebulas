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

	"github.com/nebulasio/go-nebulas/crypto"

	"github.com/nebulasio/go-nebulas/crypto/keystore"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"

	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

var (
	// BlockHashLength define a const of the length of Hash of Block in byte.
	BlockHashLength = 32

	// BlockReward given to coinbase
	// rule: 3% per year, 3,000,000. 1 block per 5 seconds
	// value: 10^8 * 3% / (365*24*3600/5) * 10^18 ≈ 16 * 3% * 10*18 = 48 * 10^16
	BlockReward = util.NewUint128FromBigInt(util.NewUint128().Mul(util.NewUint128FromInt(48).Int,
		util.NewUint128().Exp(util.NewUint128FromInt(10).Int, util.NewUint128FromInt(16).Int, nil)))
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
	return errors.New("Protobuf message cannot be converted into BlockHeader")
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
	eventsTrie  *trie.BatchTrie
	dposContext *DposContext
	txPool      *TransactionPool
	miner       *Address

	storage      storage.Storage
	eventEmitter *EventEmitter
}

// ToProto converts domain Block into proto Block
func (block *Block) ToProto() (proto.Message, error) {
	header, err := block.header.ToProto()
	if err != nil {
		return nil, err
	}
	if header, ok := header.(*corepb.BlockHeader); ok {
		txs := make([]*corepb.Transaction, len(block.transactions))
		for idx, v := range block.transactions {
			tx, err := v.ToProto()
			if err != nil {
				return nil, err
			}
			if tx, ok := tx.(*corepb.Transaction); ok {
				txs[idx] = tx
			} else {
				return nil, errors.New("Protobuf message cannot be converted into Transaction")
			}
		}
		return &corepb.Block{
			Header:       header,
			Transactions: txs,
			Height:       block.height,
			Miner:        block.miner.Bytes(),
		}, nil
	}
	return nil, errors.New("Protobuf message cannot be converted into BlockHeader")
}

// FromProto converts proto Block to domain Block
func (block *Block) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.Block); ok {
		block.header = new(BlockHeader)
		if err := block.header.FromProto(msg.Header); err != nil {
			return err
		}

		block.transactions = make(Transactions, len(msg.Transactions))
		for idx, v := range msg.Transactions {
			tx := new(Transaction)
			if err := tx.FromProto(v); err != nil {
				return err
			}
			block.transactions[idx] = tx
		}
		block.height = msg.Height
		block.miner = &Address{msg.Miner}
		return nil
	}
	return errors.New("Protobuf message cannot be converted into Block")
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
		parentBlock:  parent,
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

	block.begin()
	block.rewardCoinbase()
	block.commit()

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

// ChainID returns block's chainID
func (block *Block) ChainID() uint32 {
	return block.header.chainID
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
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
		}).Fatal("Sealed block can't be changed.")
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
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
		}).Fatal("Sealed block can't be changed.")
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

// Storage return storage.
func (block *Block) Storage() storage.Storage {
	return block.storage
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

// Height return height
func (block *Block) Height() uint64 {
	return block.height
}

// Transactions returns block transactions
func (block *Block) Transactions() Transactions {
	return block.transactions
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
func (block *Block) LinkParentBlock(chain *BlockChain, parentBlock *Block) error {
	// startAt := time.Now().Unix()

	if block.ParentHash().Equals(parentBlock.Hash()) == false {
		return ErrLinkToWrongParentBlock
	}

	var err error
	if block.accState, err = parentBlock.accState.Clone(); err != nil {
		return ErrCloneAccountState
	}
	if block.txsTrie, err = parentBlock.txsTrie.Clone(); err != nil {
		return ErrCloneTxsState
	}
	if block.eventsTrie, err = parentBlock.eventsTrie.Clone(); err != nil {
		return ErrCloneEventsState
	}

	elapsedSecond := block.Timestamp() - parentBlock.Timestamp()
	context, err := parentBlock.NextDynastyContext(chain, elapsedSecond)
	if err != nil {
		return err
	}
	// nextAt := time.Now().Unix()

	if err := block.LoadDynastyContext(context); err != nil {
		return ErrLoadNextDynastyContext
	}
	// loadAt := time.Now().Unix()

	block.txPool = parentBlock.txPool
	block.parentBlock = parentBlock
	block.storage = parentBlock.storage
	block.height = parentBlock.height + 1
	block.eventEmitter = parentBlock.eventEmitter

	/* 	logging.VLog().WithFields(logrus.Fields{
		"parent":    parentBlock,
		"block":     block,
		"err":       err,
		"time.next": nextAt - startAt,
		"time.load": loadAt - nextAt,
		"time.all":  time.Now().Unix() - startAt,
	}).Info("Linked the parent block.") */

	return nil
}

func (block *Block) begin() {
	// logging.VLog().Debug("Block Begin.")
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
	/* 	logging.VLog().WithFields(logrus.Fields{
		"block": block,
	}).Debug("Block Commit.") */
}

func (block *Block) rollback() {
	block.accState.RollBack()
	block.txsTrie.RollBack()
	block.eventsTrie.RollBack()
	block.dposContext.RollBack()
	/* 	logging.VLog().WithFields(logrus.Fields{
		"block": block,
	}).Debug("Block RollBack.") */
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
func (block *Block) CollectTransactions(deadline int64) {
	metricsBlockPackTxTime.Update(0)
	if block.sealed {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
		}).Fatal("Sealed block can't be changed.")
	}

	now := time.Now().Unix()
	elapse := deadline - now
	logging.VLog().WithFields(logrus.Fields{
		"elapse": elapse,
	}).Info("Packing tx elapsed time.")
	metricsBlockPackTxTime.Update(elapse)
	if elapse <= 0 {
		return
	}

	deadlineTimer := time.NewTimer(time.Duration(elapse) * time.Second)
	executedTxBlocksCh := make(chan *Block, 64)
	notifyCh := make(chan bool, 1)

	var givebacks []*Transaction
	pool := block.txPool

	packed := int64(0)
	unpacked := int64(0)

	// fast skip small nonce tx.
	currentNonceOfFromAddress := make(map[string]uint64)

	// execute transaction.
	go func() {
		for !pool.Empty() {
			tx := pool.Pop()

			// get current nonce for fast skip.
			currentNonce := currentNonceOfFromAddress[tx.From().String()]
			if tx.nonce <= currentNonce {
				continue
			}

			txBlock, err := block.Clone()
			if err != nil {
				return
			}

			txBlock.begin()

			giveback, currentNonce, err := txBlock.executeTransaction(tx)
			if giveback {
				givebacks = append(givebacks, tx)
			}

			// set current nonce.
			if currentNonce > 0 {
				currentNonceOfFromAddress[tx.From().String()] = currentNonce
			}

			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"block":    block,
					"tx":       tx,
					"err":      err,
					"giveback": giveback,
				}).Debug("invalid tx.")
				unpacked++
				txBlock.rollback()
				executedTxBlocksCh <- nil
			} else {
				/* 				logging.VLog().WithFields(logrus.Fields{
					"block":    block,
					"tx":       tx,
					"giveback": giveback,
				}).Info("tx is packed.") */
				packed++
				txBlock.commit()
				txBlock.transactions = append(txBlock.transactions, tx)
				executedTxBlocksCh <- txBlock
			}

			select {
			case flag := <-notifyCh:
				if flag == true {
					metricsTxPackedCount.Update(packed)
					metricsTxUnpackedCount.Update(unpacked)
					metricsTxGivebackCount.Update(int64(len(givebacks)))
					// deadline is up, put current tx back and quit.
					if giveback == false && err == nil {
						err := pool.Push(tx)
						if err != nil {
							logging.VLog().WithFields(logrus.Fields{
								"block": block,
								"tx":    tx,
								"err":   err,
							}).Debug("Failed to giveback the tx.")
						}
					}
					return
				}
			}
		}
	}()

	// consume the executedTxBlocksCh, or wait for the deadline.
	for {
		select {
		case <-deadlineTimer.C:
			// notify transaction execution goroutine to quit.
			notifyCh <- true

			// put tx back to transaction_pool.
			go func() {
				// giveback transactions.
				for _, tx := range givebacks {
					err := pool.Push(tx)
					if err != nil {
						logging.VLog().WithFields(logrus.Fields{
							"block": block,
							"tx":    tx,
							"err":   err,
						}).Debug("Failed to giveback the tx.")
					}
				}

				// execute succeed but not included in block.
				for {
					select {
					case txBlock := <-executedTxBlocksCh:
						if txBlock != nil {
							for _, tx := range txBlock.transactions {
								err := pool.Push(tx)
								if err != nil {
									logging.VLog().WithFields(logrus.Fields{
										"block": block,
										"tx":    tx,
										"err":   err,
									}).Debug("Failed to giveback the tx.")
								}
							}
						}
					default:
						return
					}
				}

			}()
			return
		case txBlock := <-executedTxBlocksCh:
			if txBlock != nil {
				metricsTxSubmit.Mark(1)
				block.Merge(txBlock)
			}

			// continue.
			notifyCh <- false
		}
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
	err := block.recordMintCnt()
	if err != nil {
		block.rollback()
		return err
	}
	block.commit()

	block.header.stateRoot, err = block.accState.RootHash()
	if err != nil {
		return err
	}
	block.header.txsRoot = block.txsTrie.RootHash()
	block.header.eventsRoot = block.eventsTrie.RootHash()
	if block.header.dposContext, err = block.dposContext.ToProto(); err != nil {
		return err
	}
	block.header.hash = HashBlock(block)
	block.sealed = true

	logging.VLog().WithFields(logrus.Fields{
		"block": block,
	}).Info("Sealed Block.")

	metricsTxPackedCount.Update(0)
	metricsTxUnpackedCount.Update(0)
	metricsTxGivebackCount.Update(0)

	return nil
}

func (block *Block) String() string {
	miner := ""
	if block.miner != nil {
		miner = block.miner.String()
	}
	return fmt.Sprintf(`{"height": %d, "hash": "%s", "parent_hash": "%s", "state": "%s", "txs": "%s", "events": "%s", "timestamp": %d, "dynasty": "%s", "tx": %d, "miner": "%s"}`,
		block.height,
		block.header.hash,
		block.header.parentHash,
		block.header.stateRoot,
		block.header.txsRoot,
		block.header.eventsRoot,
		block.header.timestamp,
		byteutils.Hex(block.header.dposContext.DynastyRoot),
		len(block.transactions),
		miner,
	)
}

// VerifyExecution execute the block and verify the execution result.
func (block *Block) VerifyExecution(parent *Block, consensus Consensus) error {
	// verify the block is acceptable by consensus
	if err := consensus.VerifyBlock(block, parent); err != nil {
		return err
	}
	// verifyAt := time.Now().Unix()

	block.begin()
	// beginAt := time.Now().Unix()

	if err := block.execute(); err != nil {
		block.rollback()
		return err
	}
	// executeAt := time.Now().Unix()

	if err := block.verifyState(); err != nil {
		block.rollback()
		return err
	}
	// stateAt := time.Now().Unix()

	block.commit()
	// commitAt := time.Now().Unix()

	// release all events
	block.triggerEvent()

	/* 	logging.VLog().WithFields(logrus.Fields{
		"time.verify":  verifyAt - startAt,
		"time.begin":   beginAt - verifyAt,
		"time.execute": executeAt - beginAt,
		"time.state":   stateAt - executeAt,
		"time.commit":  commitAt - stateAt,
		"time.event":   endAt - commitAt,
		"tail.all":     endAt - startAt,
		"block":        block,
	}).Info("Succeed to verify the block.") */
	return nil
}

func (block *Block) triggerEvent() {
	logging.VLog().WithFields(logrus.Fields{
		"count": len(block.eventEmitter.eventCh),
	}).Debug("Start TriggerEvent")

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
		event := &Event{
			Topic: topic,
			Data:  v.String(),
		}
		block.eventEmitter.Trigger(event)

		events, err := block.FetchEvents(v.hash)
		if err != nil {
			for _, e := range events {
				block.eventEmitter.Trigger(e)
			}
		}
	}

	e := &Event{
		Topic: TopicLinkBlock,
		Data:  block.String(),
	}
	block.eventEmitter.Trigger(e)

	logging.VLog().WithFields(logrus.Fields{
		"count": len(block.eventEmitter.eventCh),
	}).Debug("Stop TriggerEvent")
}

// VerifyIntegrity verify block's hash, txs' integrity and consensus acceptable.
func (block *Block) VerifyIntegrity(chainID uint32, consensus Consensus) error {
	// check ChainID.
	if block.header.chainID != chainID {
		logging.VLog().WithFields(logrus.Fields{
			"expect": chainID,
			"actual": block.header.chainID,
		}).Debug("Failed to check chainid.")
		return ErrInvalidChainID
	}

	// verify block hash.
	wantedHash := HashBlock(block)
	if !wantedHash.Equals(block.Hash()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": wantedHash,
			"actual": block.Hash(),
		}).Debug("Failed to check block's hash.")
		return ErrInvalidBlockHash
	}

	// verify transactions integrity.
	for _, tx := range block.transactions {
		if err := tx.VerifyIntegrity(block.header.chainID); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"tx":  tx,
				"err": err,
			}).Debug("Failed to verify tx's integrity.")
			return err
		}
	}

	// verify the block is acceptable by consensus.
	if err := consensus.FastVerifyBlock(block); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
			"err":   err,
		}).Debug("Failed to fast verify block.")
		metricsInvalidBlock.Inc(1)
		return err
	}

	return nil
}

// verifyState return state verify result.
func (block *Block) verifyState() error {
	// verify state root.
	stateRoot, err := block.accState.RootHash()
	if err != nil {
		return err
	}
	if !byteutils.Equal(stateRoot, block.StateRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.StateRoot(),
			"actual": stateRoot,
		}).Debug("Failed to verify state.")
		return ErrInvalidBlockStateRoot
	}

	// verify transaction root.
	if !byteutils.Equal(block.txsTrie.RootHash(), block.TxsRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.TxsRoot(),
			"actual": block.txsTrie.RootHash(),
		}).Debug("Failed to verify txs.")
		return ErrInvalidBlockTxsRoot
	}

	// verify events root.
	if !byteutils.Equal(block.eventsTrie.RootHash(), block.EventsRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.EventsRoot(),
			"actual": block.eventsTrie.RootHash(),
		}).Debug("Failed to verify events.")
		return ErrInvalidBlockEventsRoot
	}

	// verify transaction root.
	if !byteutils.Equal(block.dposContext.RootHash(), block.DposContextHash()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.DposContextHash(),
			"actual": block.dposContext.RootHash(),
		}).Debug("Failed to verify dpos context.")
		return ErrInvalidBlockDposContextRoot
	}

	return nil
}

// Execute block and return result.
func (block *Block) execute() error {
	startAt := time.Now().UnixNano()
	block.rewardCoinbase()

	start := time.Now().UnixNano()
	for _, tx := range block.transactions {
		metricsTxExecute.Mark(1)

		giveback, _, err := block.executeTransaction(tx)
		if giveback {
			err := block.txPool.Push(tx)
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}

	}
	txs := int64(len(block.transactions))
	end := time.Now().UnixNano()
	if txs != 0 {
		metricsTxVerifiedTime.Update((end - start) / txs)
	} else {
		metricsTxVerifiedTime.Update(0)
	}

	if err := block.recordMintCnt(); err != nil {
		return err
	}

	endAt := time.Now().UnixNano()
	metricsBlockVerifiedTime.Update(endAt - startAt)
	metricsTxsInBlock.Update(txs)

	return nil
}

// GetBalance returns balance for the given address on this block.
func (block *Block) GetBalance(address byteutils.Hash) (*util.Uint128, error) {
	account, err := block.accState.GetOrCreateUserAccount(address)
	if err != nil {
		return nil, err
	}
	return account.Balance(), nil
}

// GetNonce returns nonce for the given address on this block.
func (block *Block) GetNonce(address byteutils.Hash) (uint64, error) {
	account, err := block.accState.GetOrCreateUserAccount(address)
	if err != nil {
		return 0, err
	}
	return account.Nonce(), nil
}

// RecordEvent record event's topic and data with txHash
func (block *Block) RecordEvent(txHash byteutils.Hash, topic, data string) error {
	event := &Event{Topic: topic, Data: data}
	return block.recordEvent(txHash, event)
}

func (block *Block) recordEvent(txHash byteutils.Hash, event *Event) error {
	iter, err := block.eventsTrie.Iterator(txHash)
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
	key := append(txHash, byteutils.FromInt64(cnt)...)
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = block.eventsTrie.Put(key, bytes)
	if err != nil {
		return err
	}
	/* 	logging.VLog().WithFields(logrus.Fields{
		"block": block,
		"tx":    txHash.Hex(),
		"event": event,
	}).Info("Recorded event.") */
	return nil
}

// FetchEvents fetch events by txHash.
func (block *Block) FetchEvents(txHash byteutils.Hash) ([]*Event, error) {
	events := []*Event{}
	iter, err := block.eventsTrie.Iterator(txHash)
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
	// startAt := time.Now().Unix()
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
	// endAt := time.Now().Unix()

	/* 	logging.VLog().WithFields(logrus.Fields{
		"dynasty": block.Timestamp() / DynastyInterval,
		"miner":   block.miner,
		"count":   cnt,
		"time":    endAt - startAt,
	}).Debug("Recorded the block minted by the miner in the dynasty.") */
	return nil
}

func (block *Block) rewardCoinbase() error {
	stateRoot, err := block.accState.RootHash()
	logging.VLog().WithFields(logrus.Fields{
		"state": stateRoot,
		"dirty": block.accState.DirtyAccountSize(),
		"err":   err,
	}).Info("block execution.")

	coinbaseAddr := block.header.coinbase.address
	coinbaseAcc, err := block.accState.GetOrCreateUserAccount(coinbaseAddr)
	if err != nil {
		return err
	}
	coinbaseAcc.AddBalance(BlockReward)

	stateRoot, err = block.accState.RootHash()
	logging.VLog().WithFields(logrus.Fields{
		"state":    stateRoot,
		"dirty":    block.accState.DirtyAccountSize(),
		"coinbase": coinbaseAddr.Hex(),
		"balance":  coinbaseAcc.Balance(),
		"reward":   BlockReward,
	}).Info("Rewarded the coinbase.")
	return nil
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

func (block *Block) acceptTransaction(tx *Transaction) error {
	// record tx
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
	// incre nonce
	fromAcc, err := block.accState.GetOrCreateUserAccount(tx.from.address)
	if err != nil {
		return err
	}
	fromAcc.IncrNonce()
	return nil
}

func (block *Block) checkTransaction(tx *Transaction) (bool, uint64, error) {
	// check duplication
	/* 	if proof, _ := block.txsTrie.Prove(tx.hash); proof != nil {
		return false, ErrDuplicatedTransaction
	} */

	// check nonce
	fromAcc, err := block.accState.GetOrCreateUserAccount(tx.from.address)
	if err != nil {
		return true, 0, err
	}

	// pass current Nonce.
	currentNonce := fromAcc.Nonce()

	if tx.nonce < currentNonce+1 {
		return false, currentNonce, ErrSmallTransactionNonce
	} else if tx.nonce > currentNonce+1 {
		return true, currentNonce, ErrLargeTransactionNonce
	}

	return false, currentNonce, nil
}

func (block *Block) executeTransaction(tx *Transaction) (bool, uint64, error) {
	if giveback, currentNonce, err := block.checkTransaction(tx); err != nil {
		return giveback, currentNonce, err
	}

	if _, err := tx.VerifyExecution(block); err != nil {
		return false, uint64(0), err
	}

	if err := block.acceptTransaction(tx); err != nil {
		return false, uint64(0), err
	}

	return false, uint64(0), nil
}

// CheckContract check if contract is valid
func (block *Block) CheckContract(addr *Address) error {
	contract, err := block.accState.GetContractAccount(addr.Bytes())
	if err != nil {
		return err
	}

	birthEvents, err := block.FetchEvents(contract.BirthPlace())
	if err != nil {
		return err
	}

	result := false
	for _, v := range birthEvents {
		// TODO: later update event change topic
		if v.Topic == TopicExecuteTxSuccess {
			result = true
			break
		}
	}
	if !result {
		return ErrContractDeployFailed
	}

	return nil
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

// HashPbBlock return the hash of pb block.
func HashPbBlock(pbBlock *corepb.Block) byteutils.Hash {
	block := new(Block)
	block.FromProto(pbBlock)
	return block.Hash()
}

// RecoverMiner return miner from block
func RecoverMiner(block *Block) (*Address, error) {
	signature, err := crypto.NewSignature(keystore.Algorithm(block.Alg()))
	if err != nil {
		return nil, err
	}
	pub, err := signature.RecoverPublic(block.Hash(), block.Signature())
	if err != nil {
		return nil, err
	}
	pubdata, err := pub.Encoded()
	if err != nil {
		return nil, err
	}
	addr, err := NewAddressFromPublicKey(pubdata)
	if err != nil {
		return nil, err
	}
	return addr, nil
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

// Clone return new Block, with cloned state.
func (block *Block) Clone() (*Block, error) {
	accState, err := block.accState.Clone()
	if err != nil {
		return nil, ErrCloneAccountState
	}

	txsTrie, err := block.txsTrie.Clone()
	if err != nil {
		return nil, ErrCloneTxsState
	}

	eventsTrie, err := block.eventsTrie.Clone()
	if err != nil {
		return nil, ErrCloneEventsState
	}

	dposContext, err := block.dposContext.Clone()
	if err != nil {
		return nil, err
	}

	return &Block{
		header:       block.header,
		sealed:       block.sealed,
		height:       block.height,
		parentBlock:  block.parentBlock,
		txPool:       block.txPool,
		miner:        block.miner,
		storage:      block.storage,
		eventEmitter: block.eventEmitter,
		transactions: make(Transactions, 0),

		accState:    accState,
		txsTrie:     txsTrie,
		eventsTrie:  eventsTrie,
		dposContext: dposContext,
	}, nil
}

// Merge merge the state from source block.
func (block *Block) Merge(source *Block) {
	block.accState = source.accState
	block.txsTrie = source.txsTrie
	block.eventsTrie = source.eventsTrie
	block.dposContext = source.dposContext
	block.transactions = append(block.transactions, source.transactions...)
}

// Dispose dispose block.
func (block *Block) Dispose() {
	// cut off the parent block reference, prevent memory leak.
	block.parentBlock = nil
}
