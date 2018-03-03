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
	"github.com/nebulasio/go-nebulas/common/dag"
	"github.com/nebulasio/go-nebulas/common/dag/pb"
	"runtime"
	"time"

	"github.com/nebulasio/go-nebulas/core/state"

	"github.com/nebulasio/go-nebulas/crypto"

	"github.com/nebulasio/go-nebulas/crypto/keystore"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core/pb"
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
	stateRoot     byteutils.Hash
	txsRoot       byteutils.Hash
	eventsRoot    byteutils.Hash
	consensusRoot byteutils.Hash

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
		Hash:          b.hash,
		ParentHash:    b.parentHash,
		StateRoot:     b.stateRoot,
		TxsRoot:       b.txsRoot,
		EventsRoot:    b.eventsRoot,
		ConsensusRoot: b.consensusRoot,
		Nonce:         b.nonce,
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
		b.consensusRoot = msg.ConsensusRoot
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
	dependency   *dag.Dag

	sealed      bool
	height      uint64
	parentBlock *Block

	worldState state.WorldState

	txPool *TransactionPool
	miner  *Address

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
		dependency, err := block.dependency.ToProto()
		if err != nil {
			return nil, err
		}
		if dependency := dependency.(*dagpb.Dag); ok {
			return &corepb.Block{
				Header:       header,
				Transactions: txs,
				Dependency:   dependency,
				Height:       block.height,
				Miner:        block.miner.Bytes(),
			}, nil
		}
		return nil, errors.New("Protobuf message cannot be converted into Dag")
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
		block.dependency = dag.NewDag()
		if err := block.dependency.FromProto(msg.Dependency); err != nil {
			return err
		}
		block.height = msg.Height
		block.miner = &Address{msg.Miner}
		return nil
	}
	return errors.New("Protobuf message cannot be converted into Block")
}

// SerializeTxByHash returns tx serialized bytes
func (block *Block) SerializeTxByHash(hash byteutils.Hash) (proto.Message, error) {
	tx, err := GetTransaction(hash, block.worldState)
	if err != nil {
		return nil, err
	}
	return tx.ToProto()
}

// NewBlock return new block.
func NewBlock(chainID uint32, coinbase *Address, parent *Block) (*Block, error) {
	worldState, err := parent.worldState.Clone()
	if err != nil {
		return nil, err
	}

	block := &Block{
		header: &BlockHeader{
			chainID:    chainID,
			parentHash: parent.Hash(),
			coinbase:   coinbase,
			timestamp:  time.Now().Unix(),
		},
		transactions: make(Transactions, 0),
		dependency:   dag.NewDag(),
		parentBlock:  parent,
		worldState:   worldState,
		txPool:       parent.txPool,
		height:       parent.height + 1,
		sealed:       false,
		storage:      parent.storage,
		eventEmitter: parent.eventEmitter,
	}

	block.Begin()
	block.rewardCoinbase()

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

// WorldState return the world state of the block
func (block *Block) WorldState() state.WorldState {
	return block.worldState
}

// EventsRoot return events root hash.
func (block *Block) EventsRoot() byteutils.Hash {
	return block.header.eventsRoot
}

// ConsensusRoot return the roothash of consensus state
func (block *Block) ConsensusRoot() byteutils.Hash {
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
	if block.ParentHash().Equals(parentBlock.Hash()) == false {
		return ErrLinkToWrongParentBlock
	}

	var err error
	if block.worldState, err = parentBlock.WorldState().Clone(); err != nil {
		return ErrCloneAccountState
	}

	elapsedSecond := block.Timestamp() - parentBlock.Timestamp()
	consensusState, err := parentBlock.worldState.NextConsensusState(elapsedSecond)
	if err != nil {
		return err
	}
	block.WorldState().SetConsensusState(consensusState)

	block.txPool = parentBlock.txPool
	block.parentBlock = parentBlock
	block.storage = parentBlock.storage
	block.height = parentBlock.height + 1
	block.eventEmitter = parentBlock.eventEmitter

	return nil
}

// Begin a batch task
func (block *Block) Begin() error {
	return block.WorldState().Begin()
}

// Commit a batch task
func (block *Block) Commit() {
	block.WorldState().Commit()
}

// RollBack a batch task
func (block *Block) RollBack() {
	block.WorldState().RollBack()
}

// Prepare a sub batch task
func (block *Block) Prepare(tx *Transaction) (state.TxWorldState, error) {
	return block.WorldState().Prepare(tx.Hash().String())
}

// CheckAndUpdate a batch task, threadsafe
func (block *Block) CheckAndUpdate(tx *Transaction) ([]interface{}, error) {
	return block.WorldState().CheckAndUpdate(tx.Hash().String())
}

// Reset a batch task
func (block *Block) Reset(tx *Transaction) error {
	return block.WorldState().Reset(tx.Hash().String())
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

type executedResult struct {
	transaction *Transaction
	dependency  []interface{}
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

	var givebacks []*Transaction
	pool := block.txPool

	packed := int64(0)
	unpacked := int64(0)

	dag := dag.NewDag()
	transactions := []*Transaction{}

	parallelCh := make(chan bool, 32)
	executedCh := make(chan *executedResult, 32)
	mergeCh := make(chan bool, 1)
	quitCh := false

	go func() {
		for !pool.Empty() {
			tx := pool.Pop()

			parallelCh <- true
			go func() {
				giveback, err := block.ExecuteTransaction(tx)
				if giveback {
					givebacks = append(givebacks, tx)
				}
				if err != nil {
					logging.CLog().WithFields(logrus.Fields{
						"tx":       tx,
						"err":      err,
						"giveback": giveback,
					}).Debug("invalid tx.")
					unpacked++
				} else {
					mergeCh <- true
					dependency, err := block.CheckAndUpdate(tx)
					if err != nil {
						logging.CLog().WithFields(logrus.Fields{
							"tx":       tx,
							"err":      err,
							"giveback": giveback,
						}).Debug("invalid tx.")
						unpacked++
					} else {
						logging.CLog().WithFields(logrus.Fields{
							"tx": tx,
						}).Debug("packed tx.")
						packed++
						executedCh <- &executedResult{
							transaction: tx,
							dependency:  dependency,
						}
					}
					<-mergeCh
				}
				<-parallelCh
			}()

			if quitCh && len(parallelCh) == 0 {
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
			}
		}
	}()

	for {
		select {
		case <-deadlineTimer.C:
			quitCh = true
			block.transactions = transactions
			block.dependency = dag
			return
		case result := <-executedCh:
			transactions = append(transactions, result.transaction)
			txid := result.transaction.Hash().String()
			dag.AddNode(txid)
			for _, node := range result.dependency {
				dag.AddEdge(node, txid)
			}
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

	var err error
	block.header.stateRoot, err = block.WorldState().AccountsRoot()
	if err != nil {
		return err
	}
	block.header.txsRoot, err = block.WorldState().TxsRoot()
	if err != nil {
		return err
	}
	block.header.eventsRoot, err = block.WorldState().EventsRoot()
	if err != nil {
		return err
	}
	block.header.consensusRoot, err = block.WorldState().ConsensusRoot()
	if err != nil {
		return err
	}

	block.header.hash = HashBlock(block)
	block.sealed = true

	logging.VLog().WithFields(logrus.Fields{
		"block": block,
	}).Info("Sealed Block.")
	block.RollBack()

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
	return fmt.Sprintf(`{"height": %d, "hash": "%s", "parent_hash": "%s", "state": "%s", "txs": "%s", "events": "%s", "timestamp": %d, "consensus": "%s", "tx": %d, "miner": "%s"}`,
		block.height,
		block.header.hash,
		block.header.parentHash,
		block.header.stateRoot,
		block.header.txsRoot,
		block.header.eventsRoot,
		block.header.timestamp,
		block.header.consensusRoot,
		len(block.transactions),
		miner,
	)
}

// VerifyExecution execute the block and verify the execution result.
func (block *Block) VerifyExecution(parent *Block, consensus Consensus) error {
	block.Begin()

	if err := block.execute(); err != nil {
		block.RollBack()
		return err
	}

	if err := block.verifyState(); err != nil {
		block.RollBack()
		return err
	}

	block.Commit()

	// release all events
	block.triggerEvent()

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
		event := &state.Event{
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

	e := &state.Event{
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
	accountsRoot, err := block.WorldState().AccountsRoot()
	if err != nil {
		return err
	}
	if !byteutils.Equal(accountsRoot, block.StateRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.StateRoot(),
			"actual": accountsRoot,
		}).Debug("Failed to verify state.")
		return ErrInvalidBlockStateRoot
	}

	// verify transaction root.
	txsRoot, err := block.WorldState().TxsRoot()
	if err != nil {
		return err
	}
	if !byteutils.Equal(txsRoot, block.TxsRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.TxsRoot(),
			"actual": txsRoot,
		}).Debug("Failed to verify txs.")
		return ErrInvalidBlockTxsRoot
	}

	// verify events root.
	eventsRoot, err := block.WorldState().EventsRoot()
	if err != nil {
		return err
	}
	if !byteutils.Equal(eventsRoot, block.EventsRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.EventsRoot(),
			"actual": eventsRoot,
		}).Debug("Failed to verify events.")
		return ErrInvalidBlockEventsRoot
	}

	// verify transaction root.
	consensusRoot, err := block.WorldState().ConsensusRoot()
	if err != nil {
		return err
	}
	if !byteutils.Equal(consensusRoot, block.ConsensusRoot()) {
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

	dispatcher := dag.NewDispatcher(block.dependency, runtime.NumCPU(), block, func(node *dag.Node, context interface{}) error {
		block := context.(*Block)
		tx := block.transactions[node.Index]
		metricsTxExecute.Mark(1)
		if _, err := block.ExecuteTransaction(tx); err != nil {
			return err
		}
		if _, err := block.CheckAndUpdate(tx); err != nil {
			return err
		}
		return nil
	})
	start := time.Now().UnixNano()
	if err := dispatcher.Run(); err != nil {
		return err
	}
	end := time.Now().UnixNano()

	if len(block.transactions) != 0 {
		metricsTxVerifiedTime.Update((end - start) / int64(len(block.transactions)))
	} else {
		metricsTxVerifiedTime.Update(0)
	}

	endAt := time.Now().UnixNano()
	metricsBlockVerifiedTime.Update(endAt - startAt)
	metricsTxsInBlock.Update(int64(len(block.transactions)))

	return nil
}

// GetBalance returns balance for the given address on this block.
func (block *Block) GetBalance(address byteutils.Hash) (*util.Uint128, error) {
	account, err := block.WorldState().GetOrCreateUserAccount(address)
	if err != nil {
		return nil, err
	}
	return account.Balance(), nil
}

// GetNonce returns nonce for the given address on this block.
func (block *Block) GetNonce(address byteutils.Hash) (uint64, error) {
	account, err := block.WorldState().GetOrCreateUserAccount(address)
	if err != nil {
		return 0, err
	}
	return account.Nonce(), nil
}

// RecordEvent record event's topic and data with txHash
func (block *Block) RecordEvent(txHash byteutils.Hash, topic, data string) error {
	event := &state.Event{Topic: topic, Data: data}
	return block.WorldState().RecordEvent(txHash, event)
}

// FetchEvents fetch events by txHash.
func (block *Block) FetchEvents(txHash byteutils.Hash) ([]*state.Event, error) {
	return block.WorldState().FetchEvents(txHash)
}

func (block *Block) rewardCoinbase() error {
	coinbaseAddr := block.Coinbase().Bytes()
	coinbaseAcc, err := block.WorldState().GetOrCreateUserAccount(coinbaseAddr)
	if err != nil {
		return err
	}
	coinbaseAcc.AddBalance(BlockReward)
	return nil
}

// ExecuteTransaction execute the transaction
func (block *Block) ExecuteTransaction(tx *Transaction) (bool, error) {
	txWorldState, err := block.Prepare(tx)
	if err != nil {
		return true, err
	}

	if giveback, err := CheckTransaction(tx, txWorldState); err != nil {
		return giveback, err
	}

	if err := VerifyExecution(tx, block, txWorldState); err != nil {
		return false, err
	}

	if err := AcceptTransaction(tx, txWorldState); err != nil {
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
	hasher.Write(block.ConsensusRoot())
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
func LoadBlockFromStorage(hash byteutils.Hash, chain *BlockChain) (*Block, error) {
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
	block.worldState, err = state.NewWorldState(chain.ConsensusHandler(), chain.storage)
	if err != nil {
		return nil, err
	}
	if err := block.WorldState().LoadAccountsRoot(block.StateRoot()); err != nil {
		return nil, err
	}
	if err := block.WorldState().LoadTxsRoot(block.TxsRoot()); err != nil {
		return nil, err
	}
	if err := block.WorldState().LoadEventsRoot(block.EventsRoot()); err != nil {
		return nil, err
	}
	if err := block.WorldState().LoadConsensusRoot(block.ConsensusRoot()); err != nil {
		return nil, err
	}

	block.txPool = chain.txPool
	block.storage = chain.storage
	block.sealed = true
	block.eventEmitter = chain.eventEmitter
	return block, nil
}

// Dispose dispose block.
func (block *Block) Dispose() {
	// cut off the parent block reference, prevent memory leak.
	block.parentBlock = nil
}
