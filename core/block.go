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
	"fmt"
	"reflect"
	"time"

	"github.com/nebulasio/go-nebulas/crypto"

	"github.com/nebulasio/go-nebulas/consensus/pb"
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
	BlockReward, _ = util.NewUint128FromString("480000000000000000")
)

// BlockHeader of a block
type BlockHeader struct {
	hash       byteutils.Hash
	parentHash byteutils.Hash

	// world state
	stateRoot     byteutils.Hash
	txsRoot       byteutils.Hash
	eventsRoot    byteutils.Hash
	consensusRoot *consensuspb.ConsensusRoot

	coinbase  *Address
	timestamp int64
	chainID   uint32

	// sign
	alg  keystore.Algorithm
	sign byteutils.Hash
}

// ToProto converts domain BlockHeader to proto BlockHeader
func (b *BlockHeader) ToProto() (proto.Message, error) {
	return &corepb.BlockHeader{
		Hash:          b.hash,
		ParentHash:    b.parentHash,
		StateRoot:     b.stateRoot,
		TxsRoot:       b.txsRoot,
		EventsRoot:    b.eventsRoot,
		ConsensusRoot: b.consensusRoot,
		Coinbase:      b.coinbase.address,
		Timestamp:     b.timestamp,
		ChainId:       b.chainID,
		Alg:           uint32(b.alg),
		Sign:          b.sign,
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
		if msg.ConsensusRoot == nil {
			return ErrInvalidProtoToBlockHeader
		}
		b.consensusRoot = msg.ConsensusRoot
		coinbase, err := AddressParseFromBytes(msg.Coinbase)
		if err != nil {
			return ErrInvalidProtoToBlockHeader
		}
		b.coinbase = coinbase
		b.timestamp = msg.Timestamp
		b.chainID = msg.ChainId
		b.alg = keystore.Algorithm(msg.Alg)
		b.sign = msg.Sign
		return nil
	}
	return ErrInvalidProtoToBlockHeader
}

// Block structure
type Block struct {
	header       *BlockHeader
	transactions Transactions

	sealed         bool
	height         uint64
	parentBlock    *Block
	accState       state.AccountState
	txsState       *trie.BatchTrie
	eventsState    *trie.BatchTrie
	consensusState state.ConsensusState
	txPool         *TransactionPool

	storage      storage.Storage
	eventEmitter *EventEmitter
	nvm          Engine
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
				return nil, ErrCannotConvertTransaction
			}
		}
		return &corepb.Block{
			Header:       header,
			Transactions: txs,
			Height:       block.height,
		}, nil
	}
	return nil, ErrInvalidBlockToProto
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
		return nil
	}
	return ErrInvalidProtoToBlock
}

// NewBlock return new block.
func NewBlock(chainID uint32, coinbase *Address, parent *Block) (*Block, error) { // ToCheck: check args. // ToCheck: check full-functional block.
	accState, err := parent.accState.Clone()
	if err != nil {
		return nil, err
	}
	txsState, err := parent.txsState.Clone()
	if err != nil {
		return nil, err
	}
	eventsState, err := parent.eventsState.Clone()
	if err != nil {
		return nil, err
	}
	consensusState, err := parent.consensusState.Clone()
	if err != nil {
		return nil, err
	}
	block := &Block{
		header: &BlockHeader{
			parentHash:    parent.Hash(),
			coinbase:      coinbase,
			timestamp:     time.Now().Unix(),
			chainID:       chainID,
			consensusRoot: &consensuspb.ConsensusRoot{},
		},
		transactions:   make(Transactions, 0),
		parentBlock:    parent,
		accState:       accState,
		txsState:       txsState,
		eventsState:    eventsState,
		consensusState: consensusState,
		txPool:         parent.txPool,
		height:         parent.height + 1,
		sealed:         false,
		storage:        parent.storage,
		eventEmitter:   parent.eventEmitter,
		nvm:            parent.nvm,
	}

	block.begin()
	block.rewardCoinbase()
	block.commit() // ToDelete

	return block, nil
}

// Sign sign transaction,sign algorithm is
func (block *Block) Sign(signature keystore.Signature) error {
	sign, err := signature.Sign(block.header.hash)
	if err != nil {
		return err
	}
	block.header.alg = keystore.Algorithm(signature.Algorithm())
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
func (block *Block) Alg() keystore.Algorithm {
	return block.header.alg
}

// Signature return block's signature
func (block *Block) Signature() byteutils.Hash {
	return block.header.sign
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

// ConsensusRoot return consensus root
func (block *Block) ConsensusRoot() *consensuspb.ConsensusRoot {
	return block.header.consensusRoot
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

// AccountState return account state
func (block *Block) AccountState() state.AccountState {
	return block.accState
}

// LinkParentBlock link parent block, return true if hash is the same; false otherwise.
func (block *Block) LinkParentBlock(chain *BlockChain, parentBlock *Block) error {
	if !block.ParentHash().Equals(parentBlock.Hash()) {
		return ErrLinkToWrongParentBlock
	}

	var err error
	if block.accState, err = parentBlock.accState.Clone(); err != nil {
		return ErrCloneAccountState
	}
	if block.txsState, err = parentBlock.txsState.Clone(); err != nil {
		return ErrCloneTxsState
	}
	if block.eventsState, err = parentBlock.eventsState.Clone(); err != nil {
		return ErrCloneEventsState
	}

	elapsedSecond := block.Timestamp() - parentBlock.Timestamp()
	consensusState, err := parentBlock.consensusState.NextState(elapsedSecond)
	if err != nil {
		return err
	}
	block.LoadConsensusState(consensusState)

	block.txPool = parentBlock.txPool
	block.parentBlock = parentBlock
	block.storage = parentBlock.storage
	block.height = parentBlock.height + 1
	block.eventEmitter = parentBlock.eventEmitter
	block.nvm = parentBlock.nvm

	return nil
}

func (block *Block) begin() {
	block.accState.Begin()
	block.txsState.Begin()
	block.eventsState.Begin()
	block.consensusState.Begin()
}

func (block *Block) commit() {
	block.accState.Commit()
	block.txsState.Commit()
	block.eventsState.Commit()
	block.consensusState.Commit()
}

func (block *Block) rollback() {
	block.accState.Rollback()
	block.txsState.Rollback()
	block.eventsState.Rollback()
	block.consensusState.Rollback()
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

	elapse := deadline - time.Now().Unix()
	logging.VLog().WithFields(logrus.Fields{
		"elapse": elapse,
	}).Info("Time to pack txs.")
	metricsBlockPackTxTime.Update(elapse)
	if elapse <= 0 {
		return
	}
	deadlineTimer := time.NewTimer(time.Duration(elapse) * time.Second)

	packed := int64(0)
	unpacked := int64(0)

	pool := block.txPool
	inprogress := make(map[byteutils.HexHash]bool)

	parallelCh := make(chan bool, 1)
	exclusiveCh := make(chan bool, 1)
	over := false

	go func() {
		for !pool.Empty() {
			// pop a valid tx
			exclusiveCh <- true
			if over {
				<-exclusiveCh
				return
			}

			tx := pool.PopWithBlacklist(inprogress)
			if tx == nil {
				<-exclusiveCh
				time.Sleep(time.Millisecond)
				continue
			}
			from := tx.from.address.Hex()
			inprogress[from] = true
			<-exclusiveCh

			parallelCh <- true
			go func() {
				defer func() { <-parallelCh }()

				// prepare independent environment
				exclusiveCh <- true
				txBlock, err := block.Clone()
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"block": block,
						"tx":    tx,
						"err":   err,
					}).Debug("Failed to prepare tx execution environment.")
					delete(inprogress, from)
					<-exclusiveCh
					return
				}
				<-exclusiveCh

				// execute tx
				merge := false
				txBlock.begin()
				giveback, err := txBlock.executeTransaction(tx)
				if err != nil {
					logging.CLog().WithFields(logrus.Fields{
						"tx":       tx,
						"err":      err,
						"giveback": giveback,
					}).Debug("invalid tx.")
					unpacked++

					txBlock.rollback()
					if giveback {
						if err := pool.Push(tx); err != nil {
							logging.VLog().WithFields(logrus.Fields{
								"block": block,
								"tx":    tx,
								"err":   err,
							}).Debug("Failed to giveback the tx.")
						}
					}
				} else {
					logging.CLog().WithFields(logrus.Fields{
						"tx": tx,
					}).Debug("packed tx.")
					packed++

					txBlock.transactions = append(txBlock.transactions, tx)
					txBlock.commit()
					merge = true
				}

				// merge tx
				exclusiveCh <- true
				if over {
					<-exclusiveCh
					return
				}
				if merge {
					delete(inprogress, from)
					block.Merge(txBlock)
				}
				<-exclusiveCh
			}()
		}
	}()

	<-deadlineTimer.C
	exclusiveCh <- true
	over = true
	<-exclusiveCh
	return
}

// Sealed return true if block seals. Otherwise return false.
func (block *Block) Sealed() bool {
	return block.sealed
}

// Seal seal block, calculate stateRoot and block hash.
func (block *Block) Seal() error {
	if block.sealed {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
		}).Fatal("cannot seal a block twice.")
	}

	defer block.rollback()

	var err error
	block.header.stateRoot, err = block.accState.RootHash()
	if err != nil {
		return err
	}
	block.header.txsRoot = block.txsState.RootHash()
	block.header.eventsRoot = block.eventsState.RootHash()
	if block.header.consensusRoot, err = block.consensusState.RootHash(); err != nil {
		return err
	}
	block.header.hash, err = HashBlock(block)
	if err != nil {
		return err
	}
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
	return fmt.Sprintf(`{"height": %d, "hash": "%s", "parent_hash": "%s", "acc_root": "%s", "timestamp": %d, "tx": %d, "miner": "%s"}`,
		block.height,
		block.header.hash,
		block.header.parentHash,
		block.header.stateRoot,
		block.header.timestamp,
		len(block.transactions),
		byteutils.Hex(block.header.consensusRoot.Proposer), //miner
	)
}

// VerifyExecution execute the block and verify the execution result.
func (block *Block) VerifyExecution() error {
	block.begin()

	if err := block.execute(); err != nil {
		block.rollback()
		return err
	}

	if err := block.verifyState(); err != nil {
		block.rollback()
		return err
	}

	block.commit()

	return nil
}

// VerifyIntegrity verify block's hash, txs' integrity and consensus acceptable.
func (block *Block) VerifyIntegrity(chainID uint32, consensus Consensus) error {

	if consensus == nil {
		metricsInvalidBlock.Inc(1)
		return ErrNilArgument
	}

	// check ChainID.
	if block.header.chainID != chainID {
		logging.VLog().WithFields(logrus.Fields{
			"expect": chainID,
			"actual": block.header.chainID,
		}).Debug("Failed to check chainid.")
		metricsInvalidBlock.Inc(1)
		return ErrInvalidChainID
	}

	// verify block hash.
	wantedHash, err := HashBlock(block)
	if err != nil {
		return err
	}
	if !wantedHash.Equals(block.Hash()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": wantedHash,
			"actual": block.Hash(),
		}).Debug("Failed to check block's hash.")
		metricsInvalidBlock.Inc(1)
		return ErrInvalidBlockHash
	}

	// verify transactions integrity.
	for _, tx := range block.transactions {
		if err := tx.VerifyIntegrity(block.header.chainID); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"tx":  tx,
				"err": err,
			}).Debug("Failed to verify tx's integrity.")
			metricsInvalidBlock.Inc(1)
			return err
		}
	}

	// verify the block is acceptable by consensus.
	if err := consensus.VerifyBlock(block); err != nil {
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
	if !byteutils.Equal(block.txsState.RootHash(), block.TxsRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.TxsRoot(),
			"actual": byteutils.Hex(block.txsState.RootHash()),
		}).Debug("Failed to verify txs.")
		return ErrInvalidBlockTxsRoot
	}

	// verify events root.
	if !byteutils.Equal(block.eventsState.RootHash(), block.EventsRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.EventsRoot(),
			"actual": byteutils.Hex(block.eventsState.RootHash()),
		}).Debug("Failed to verify events.")
		return ErrInvalidBlockEventsRoot
	}

	// verify transaction root.
	consensusRoot, err := block.consensusState.RootHash()
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(consensusRoot, block.ConsensusRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.ConsensusRoot(),
			"actual": consensusRoot,
		}).Debug("Failed to verify dpos context.")
		return ErrInvalidBlockConsensusRoot
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

		giveback, err := block.executeTransaction(tx)
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
	iter, err := block.eventsState.Iterator(txHash)
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
	_, err = block.eventsState.Put(key, bytes)
	if err != nil {
		return err
	}
	return nil
}

// FetchEvents fetch events by txHash.
func (block *Block) FetchEvents(txHash byteutils.Hash) ([]*Event, error) {
	events := []*Event{}
	iter, err := block.eventsState.Iterator(txHash)
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

func (block *Block) rewardCoinbase() error {
	coinbaseAddr := block.header.coinbase.address
	coinbaseAcc, err := block.accState.GetOrCreateUserAccount(coinbaseAddr)
	if err != nil {
		return err
	}
	return coinbaseAcc.AddBalance(BlockReward)
}

// GetTransaction from txs Trie
func (block *Block) GetTransaction(hash byteutils.Hash) (*Transaction, error) {
	txBytes, err := block.txsState.Get(hash) // ToCheck, hash is valid
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
	if _, err := block.txsState.Put(tx.hash, txBytes); err != nil {
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

func (block *Block) checkTransaction(tx *Transaction) (bool, error) {
	// check nonce
	fromAcc, err := block.accState.GetOrCreateUserAccount(tx.from.address)
	if err != nil {
		return true, err
	}

	// pass current Nonce.
	currentNonce := fromAcc.Nonce()

	if tx.nonce < currentNonce+1 {
		return false, ErrSmallTransactionNonce
	} else if tx.nonce > currentNonce+1 {
		return true, ErrLargeTransactionNonce
	}

	return false, nil
}

func (block *Block) executeTransaction(tx *Transaction) (bool, error) {
	if giveback, err := block.checkTransaction(tx); err != nil {
		return giveback, err
	}

	if _, err := tx.VerifyExecution(block); err != nil {
		return false, err
	}

	if err := block.acceptTransaction(tx); err != nil {
		return false, err
	}

	return false, nil
}

// Dynasty return the validators in current dynasty
func (block *Block) Dynasty() ([]byteutils.Hash, error) {
	return block.consensusState.Dynasty()
}

// NextConsensusState return the next consensus state
func (block *Block) NextConsensusState(elapsed int64) (state.ConsensusState, error) {
	return block.consensusState.NextState(elapsed)
}

// LoadConsensusState load the consensusState
func (block *Block) LoadConsensusState(consensusState state.ConsensusState) {
	block.consensusState = consensusState
	block.SetTimestamp(consensusState.TimeStamp())
}

// CheckContract check if contract is valid
func (block *Block) CheckContract(addr *Address) (state.Account, error) {

	contract, err := block.accState.GetContractAccount(addr.Bytes())
	if err != nil {
		return nil, err
	}

	birthEvents, err := block.FetchEvents(contract.BirthPlace())
	if err != nil {
		return nil, err
	}

	result := false
	for _, v := range birthEvents {

		if v.Topic == TopicTransactionExecutionResult {
			txEvent := TransactionEvent{}
			err = json.Unmarshal([]byte(v.Data), &txEvent)
			if err != nil {
				return nil, err
			}
			if txEvent.Status == TxExecutionSuccess {
				result = true
				break
			}
		}
	}
	if !result {
		return nil, ErrContractCheckFailed
	}

	return contract, nil
}

// HashBlock return the hash of block.
func HashBlock(block *Block) (byteutils.Hash, error) {
	if block == nil {
		return nil, ErrNilArgument
	}

	hasher := sha3.New256()

	consensusRoot, err := proto.Marshal(block.ConsensusRoot())
	if err != nil {
		return nil, err
	}

	hasher.Write(block.ParentHash())
	hasher.Write(block.StateRoot())
	hasher.Write(block.TxsRoot())
	hasher.Write(block.EventsRoot())
	hasher.Write(consensusRoot)
	hasher.Write(block.header.coinbase.address)
	hasher.Write(byteutils.FromInt64(block.header.timestamp))
	hasher.Write(byteutils.FromUint32(block.header.chainID))

	for _, tx := range block.transactions {
		hasher.Write(tx.Hash())
	}

	return hasher.Sum(nil), nil
}

// HashPbBlock return the hash of pb block.
func HashPbBlock(pbBlock *corepb.Block) byteutils.Hash { //ToAdd nil check
	block := new(Block) // ToFix: hash pbBlock directly, avoid catching fromproto err
	if err := block.FromProto(pbBlock); err != nil {
		if hash, err := HashBlock(block); err != nil {
			return hash
		}
	}
	return nil
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
func LoadBlockFromStorage(hash byteutils.Hash, chain *BlockChain) (*Block, error) {
	if chain == nil {
		return nil, ErrNilArgument
	}

	value, err := chain.storage.Get(hash)
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
	block.accState, err = state.NewAccountState(block.StateRoot(), chain.storage)
	if err != nil {
		return nil, err
	}
	block.txsState, err = trie.NewBatchTrie(block.TxsRoot(), chain.storage)
	if err != nil {
		return nil, err
	}
	block.eventsState, err = trie.NewBatchTrie(block.EventsRoot(), chain.storage)
	if err != nil {
		return nil, err
	}
	consensusState, err := chain.consensusHandler.NewState(block.ConsensusRoot(), chain.storage)
	if err != nil {
		return nil, err
	}
	block.LoadConsensusState(consensusState)
	block.txPool = chain.txPool
	block.storage = chain.storage
	block.sealed = true
	block.eventEmitter = chain.eventEmitter
	block.nvm = chain.nvm
	return block, nil
}

// Clone return new Block, with cloned state.
func (block *Block) Clone() (*Block, error) {
	accState, err := block.accState.Clone()
	if err != nil {
		return nil, ErrCloneAccountState
	}

	txsState, err := block.txsState.Clone()
	if err != nil {
		return nil, ErrCloneTxsState
	}

	eventsState, err := block.eventsState.Clone()
	if err != nil {
		return nil, ErrCloneEventsState
	}

	consensusState, err := block.consensusState.Clone()
	if err != nil {
		return nil, err
	}

	transactions := []*Transaction{}
	for _, tx := range block.transactions {
		transactions = append(transactions, tx)
	}

	return &Block{
		header:         block.header,
		sealed:         block.sealed,
		height:         block.height,
		parentBlock:    block.parentBlock,
		txPool:         block.txPool,
		storage:        block.storage,
		eventEmitter:   block.eventEmitter,
		nvm:            block.nvm,
		transactions:   transactions,
		accState:       accState,
		txsState:       txsState,
		eventsState:    eventsState,
		consensusState: consensusState,
	}, nil
}

// Merge merge the state from source block.
func (block *Block) Merge(source *Block) {
	block.accState = source.accState
	block.txsState = source.txsState
	block.eventsState = source.eventsState
	block.consensusState = source.consensusState
	block.transactions = source.transactions
}

// Dispose dispose block.
func (block *Block) Dispose() {
	// cut off the parent block reference, prevent memory leak.
	block.parentBlock = nil
}
