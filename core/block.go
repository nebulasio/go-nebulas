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
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/dag"
	dagpb "github.com/nebulasio/go-nebulas/common/dag/pb"
	consensuspb "github.com/nebulasio/go-nebulas/consensus/pb"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

var (
	// BlockHashLength define a const of the length of Hash of Block in byte.
	BlockHashLength = 32

	// ParallelNum num
	PackedParallelNum = 1

	// verify thread parallel num
	VerifyParallelNum = 1 //runtime.NumCPU() * 2

	// VerifyExecutionTimeout 0 means unlimited
	VerifyExecutionTimeout = 0

	// BlockReward given to coinbase
	// rule: 3% per year, 3,000,000. 1 block per 15 seconds
	// value: 10^8 * 3% / (365*24*3600/15) * 10^18 ≈ 1.42694 * 10^18
	BlockReward, _ = util.NewUint128FromString("1426940000000000000")

	// BlockRewardV2 given to coinbase after nbre available
	// rule: 2% per year, 2,000,000. 1 block per 15 seconds
	// value: 10^8 * 2% / (365*24*3600/15) * 10^18 ≈ 0.95129 * 10^18
	BlockRewardV2, _ = util.NewUint128FromString("951290000000000000")

	// BlockRewardV3 given to coinbase
	// rule: 2.5% per year, 3,000,000. 1 block per 15 seconds
	// value: 10^8 * 2.5% / (365*24*3600/15) * 10^18 ≈ 1.18912 * 10^18
	BlockRewardV3, _ = util.NewUint128FromString("1189120000000000000")

	// GovernanceReward given to governance contract
	// rule: 0.5% per year, 3,000,000. 1 block per 15 seconds
	// value: 10^8 * 0.5% / (365*24*3600/15) * 10^18 ≈ 0.23782 * 10^18
	GovernanceReward, _ = util.NewUint128FromString("237820000000000000")

	// NebulasRewardV2 given to nebulas project address
	// rule: 1% per year, 1,000,000. 1 block per 15 seconds
	// value: 10^8 * 1% / (365*24*3600/15) * 10^18 ≈ 0.47565 * 10^18
	NebulasRewardV2, _ = util.NewUint128FromString("475650000000000000")

	// DIPRewardV2 given to dip project address
	// rule: 1% per year, 1,000,000. 1 block per 15 seconds
	// value: 10^8 * 1% / (365*24*3600/15) * 10^18 ≈ 0.47565 * 10^18
	DIPRewardV2, _ = util.NewUint128FromString("475650000000000000")

	// NebulasRewardAddress Nebulas Council Recycling address
	NebulasRewardAddress, _ = AddressParse("n1Rc66BjDF4LSoQ2uC9rbiMDnKMEV8ryG7k")

	// NebulasRewardAddress Nebulas Council Recycling address
	NebulasRewardAddressV2, _ = AddressParse("n1bMN7dssdVCv7XtnF6tmwB59pxxrwvpNwP")

	// NebulasRewardAddress Nebulas Council Recycling address
	DIPRewardAddressV2, _ = AddressParse("n1HXQWZbnCwK2QVyFuNSM47CVxEUq1GEhLc")
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

	// rand
	random *corepb.Random
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
		Random:        b.random,
	}, nil
}

// FromProto converts proto BlockHeader to domain BlockHeader
func (b *BlockHeader) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.BlockHeader); ok {
		if msg != nil {
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

			alg := keystore.Algorithm(msg.Alg)
			if err := crypto.CheckAlgorithm(alg); err != nil {
				return err
			}

			b.alg = alg
			b.sign = msg.Sign
			b.random = msg.Random
			return nil
		}
		return ErrInvalidProtoToBlockHeader
	}
	return ErrInvalidProtoToBlockHeader
}

// Block structure
type Block struct {
	header       *BlockHeader
	transactions Transactions
	dependency   *dag.Dag

	sealed bool
	height uint64

	worldState state.WorldState

	txPool       *TransactionPool
	eventEmitter *EventEmitter
	nvm          NVM
	nr           NR
	dip          Dip

	storage storage.Storage
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
				return nil, ErrInvalidProtoToTransaction
			}
		}
		dependency, err := block.dependency.ToProto()
		if err != nil {
			return nil, err
		}
		if dependency, ok := dependency.(*dagpb.Dag); ok {
			return &corepb.Block{
				Header:       header,
				Transactions: txs,
				Dependency:   dependency,
				Height:       block.height,
			}, nil
		}
		return nil, dag.ErrInvalidProtoToDag
	}
	return nil, ErrInvalidProtoToBlock
}

// FromProto converts proto Block to domain Block
func (block *Block) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.Block); ok {
		if msg != nil {
			block.header = new(BlockHeader)
			if err := block.header.FromProto(msg.Header); err != nil {
				return err
			}
			if RandomAvailableAtHeight(msg.Height) && !block.HasRandomSeed() {
				logging.VLog().WithFields(logrus.Fields{
					"blockHeight":      msg.Height,
					"compatibleHeight": NebCompatibility.RandomAvailableHeight(),
				}).Info("No random found in block header.")
				return ErrInvalidProtoToBlockHeader
			}
			block.transactions = make(Transactions, len(msg.Transactions))
			for idx, v := range msg.Transactions {
				if v != nil {
					tx := new(Transaction)
					if err := tx.FromProto(v); err != nil {
						return err
					}
					block.transactions[idx] = tx
				} else {
					return ErrInvalidProtoToTransaction
				}
			}
			block.dependency = dag.NewDag()
			if err := block.dependency.FromProto(msg.Dependency); err != nil {
				return err
			}
			block.height = msg.Height
			return nil
		}
		return ErrInvalidProtoToBlock
	}
	return ErrInvalidProtoToBlock
}

// NewBlock return new block.
func NewBlock(chainID uint32, coinbase *Address, parent *Block) (*Block, error) { // ToCheck: check args. // ToCheck: check full-functional block.
	worldState, err := parent.worldState.Clone()
	if err != nil {
		return nil, err
	}

	block := &Block{
		header: &BlockHeader{
			chainID:       chainID,
			parentHash:    parent.Hash(),
			coinbase:      coinbase,
			timestamp:     time.Now().Unix(),
			consensusRoot: &consensuspb.ConsensusRoot{},
			random:        &corepb.Random{},
		},
		transactions: make(Transactions, 0),
		dependency:   dag.NewDag(),

		worldState: worldState,
		height:     parent.height + 1,
		sealed:     false,

		txPool:       parent.txPool,
		eventEmitter: parent.eventEmitter,
		nvm:          parent.nvm,
		nr:           parent.nr,
		dip:          parent.dip,
		storage:      parent.storage,
	}

	if err := block.Begin(); err != nil {
		return nil, err
	}
	if err := block.rewardCoinbaseForMint(); err != nil {
		return nil, err
	}

	return block, nil
}

// SetSignature set the signature
func (block *Block) SetSignature(alg keystore.Algorithm, sign byteutils.Hash) {
	block.header.alg = alg
	block.header.sign = sign
}

// Sign sign transaction,sign algorithm is
func (block *Block) Sign(signature keystore.Signature) error {
	if signature == nil {
		return ErrNilArgument
	}
	sign, err := signature.Sign(block.header.hash)
	if err != nil {
		return err
	}
	block.SetSignature(keystore.Algorithm(signature.Algorithm()), sign)
	return nil
}

// SetRandomSeed set block.header.random
func (block *Block) SetRandomSeed(vrfseed, vrfproof []byte) {
	block.header.random = &corepb.Random{
		VrfSeed:  vrfseed,
		VrfProof: vrfproof,
	}
}

// HasRandomSeed check random if exists
func (block *Block) HasRandomSeed() bool {
	return block.header.random != nil && block.header.random.VrfSeed != nil && block.header.random.VrfProof != nil
}

// ChainID returns block's chainID
func (block *Block) ChainID() uint32 {
	return block.header.chainID
}

// Miner return block's miner, only block is sealed return value
func (block *Block) Miner() *Address {
	proposer := block.ConsensusRoot().Proposer
	miner, err := AddressParseFromBytes(proposer)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
			"err":   err,
		}).Warn("Failed to get block miner.")
		return nil
	}
	return miner
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

// WorldState return the world state of the block
func (block *Block) WorldState() state.WorldState {
	return block.worldState
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

// RandomSeed block random seed (VRF)
func (block *Block) RandomSeed() string {
	if RandomAvailableAtHeight(block.height) {
		return byteutils.Hex(block.header.random.VrfSeed)
	}
	return ""
}

// RandomProof block random proof (VRF)
func (block *Block) RandomProof() string {
	if RandomAvailableAtHeight(block.height) {
		return byteutils.Hex(block.header.random.VrfProof)
	}
	return ""
}

// RandomAvailable check if Math.random available in contract
func (block *Block) RandomAvailable() bool {
	return RandomAvailableAtHeight(block.height)
}

// DateAvailable check if date available in contract
func (block *Block) DateAvailable() bool {
	return DateAvailableAtHeight(block.height)
}

// DateAvailable return nr
func (block *Block) NR() NR {
	return block.nr
}

// LinkParentBlock link parent block, return true if hash is the same; false otherwise.
func (block *Block) LinkParentBlock(chain *BlockChain, parentBlock *Block) error {
	if !block.ParentHash().Equals(parentBlock.Hash()) {
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

	block.height = parentBlock.height + 1
	block.txPool = parentBlock.txPool
	block.storage = parentBlock.storage
	block.eventEmitter = parentBlock.eventEmitter
	block.nvm = parentBlock.nvm
	block.nr = parentBlock.nr
	block.dip = parentBlock.dip

	return nil
}

// Begin a batch task
func (block *Block) Begin() error {
	return block.WorldState().Begin()
}

// Commit a batch task
func (block *Block) Commit() {
	if err := block.WorldState().Commit(); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Failed to commit the block")
	}
}

// RollBack a batch task
func (block *Block) RollBack() {
	if err := block.WorldState().RollBack(); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Failed to rollback the block")
	}
}

// ReturnTransactions and giveback them to tx pool
// if a block is reverted, we should erase all changes
// made by this block on storage. use refcount.
func (block *Block) ReturnTransactions() {
	for _, tx := range block.transactions {
		block.txPool.Push(tx)
	}
}

// CollectTransactions and add them to block.
func (block *Block) CollectTransactions(deadlineInMs int64) {
	metricsBlockPackTxTime.Update(0)
	if block.sealed {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
		}).Fatal("Sealed block can't be changed.")
	}

	secondInMs := int64(1000)
	elapseInMs := deadlineInMs - time.Now().Unix()*secondInMs
	logging.VLog().WithFields(logrus.Fields{
		"elapse": elapseInMs,
	}).Info("Time to pack txs.")
	metricsBlockPackTxTime.Update(elapseInMs)
	if elapseInMs <= 0 {
		return
	}
	deadlineTimer := time.NewTimer(time.Duration(elapseInMs) * time.Millisecond)

	pool := block.txPool

	packed := int64(0)
	unpacked := int64(0)

	dag := dag.NewDag()
	transactions := []*Transaction{}
	fromBlacklist := new(sync.Map)
	toBlacklist := new(sync.Map)

	// parallelCh is used as access tokens here
	parallelCh := make(chan bool, PackedParallelNum)
	// mergeCh is used as lock here
	mergeCh := make(chan bool, 1)
	over := false

	try := 0
	fetch := 0
	failed := 0
	conflict := 0
	expired := 0
	bucket := len(block.txPool.all)
	packing := int64(0)
	prepare := int64(0)
	execute := int64(0)
	update := int64(0)
	parallel := 0
	beginAt := time.Now().UnixNano()

	go func() {
		for {
			mergeCh <- true // lock
			if over {
				<-mergeCh // unlock
				return
			}
			try++
			tx := pool.PopWithBlacklist(fromBlacklist, toBlacklist)
			if tx == nil {
				<-mergeCh // unlock
				continue
			}

			logging.VLog().WithFields(logrus.Fields{
				"tx.hash": tx.hash,
			}).Debug("Pop tx.")

			fetch++
			fromBlacklist.Store(tx.from.address.Hex(), true)
			fromBlacklist.Store(tx.to.address.Hex(), true)
			toBlacklist.Store(tx.from.address.Hex(), true)
			toBlacklist.Store(tx.to.address.Hex(), true)
			<-mergeCh // lock

			parallelCh <- true // fetch access token
			go func() {
				parallel++
				startAt := time.Now().UnixNano()
				defer func() {
					endAt := time.Now().UnixNano()
					packing += endAt - startAt
					<-parallelCh // release access token
				}()

				// step1. prepare execution environment
				mergeCh <- true // lock
				if over {
					expired++
					<-mergeCh // unlock
					if err := pool.Push(tx); err != nil {
						logging.VLog().WithFields(logrus.Fields{
							"block": block,
							"tx":    tx,
							"err":   err,
						}).Info("Failed to giveback the tx.")
					}
					return
				}

				prepareAt := time.Now().UnixNano()
				txWorldState, err := block.WorldState().Prepare(tx.Hash().String())
				preparedAt := time.Now().UnixNano()
				prepare += preparedAt - prepareAt
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"block": block,
						"tx":    tx,
						"err":   err,
					}).Info("Failed to prepare tx.")
					failed++

					if err := pool.Push(tx); err != nil {
						logging.VLog().WithFields(logrus.Fields{
							"block": block,
							"tx":    tx,
							"err":   err,
						}).Info("Failed to giveback the tx.")
					}

					fromBlacklist.Delete(tx.from.address.Hex())
					fromBlacklist.Delete(tx.to.address.Hex())
					toBlacklist.Delete(tx.from.address.Hex())
					toBlacklist.Delete(tx.to.address.Hex())
					<-mergeCh // unlock
					return
				}
				<-mergeCh // unlock

				defer func() {
					if err := txWorldState.Close(); err != nil {
						logging.VLog().WithFields(logrus.Fields{
							"block": block,
							"tx":    tx,
							"err":   err,
						}).Info("Failed to close tx.")
					}
				}()

				// step2. execute tx.
				executeAt := time.Now().UnixNano()
				giveback, err := block.ExecuteTransaction(tx, txWorldState)
				executedAt := time.Now().UnixNano()
				execute += executedAt - executeAt
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"tx":       tx,
						"err":      err,
						"giveback": giveback,
					}).Debug("invalid tx.")
					unpacked++
					failed++

					/* 					if err := txWorldState.Close(); err != nil {
						logging.VLog().WithFields(logrus.Fields{
							"block": block,
							"tx":    tx,
							"err":   err,
						}).Debug("Failed to close tx.")
					} */

					if giveback {
						if err := pool.Push(tx); err != nil {
							logging.VLog().WithFields(logrus.Fields{
								"block": block,
								"tx":    tx,
								"err":   err,
							}).Info("Failed to giveback the tx.")
						}
					}
					if err == ErrLargeTransactionNonce {
						// as for the transactions from a same account
						// we will pop them out of transaction pool order by nonce ascend
						// thus, when we find a transaction with a very large nonce
						// we won't try to pack other transactions from the same account in the block
						// the account will be in our from blacklist util the block is sealed
						if !byteutils.Equal(tx.to.address, tx.from.address) {
							fromBlacklist.Delete(tx.to.address.Hex())
						}
						toBlacklist.Delete(tx.to.address.Hex())
					} else {
						fromBlacklist.Delete(tx.from.address.Hex())
						fromBlacklist.Delete(tx.to.address.Hex())
						toBlacklist.Delete(tx.from.address.Hex())
						toBlacklist.Delete(tx.to.address.Hex())
					}
					return
				}

				// step3. check & update tx
				mergeCh <- true // lock
				if over {
					expired++
					<-mergeCh // unlock
					if err := pool.Push(tx); err != nil {
						logging.VLog().WithFields(logrus.Fields{
							"block": block,
							"tx":    tx,
							"err":   err,
						}).Info("Failed to giveback the tx.")
					}
					return
				}
				updateAt := time.Now().UnixNano()
				dependency, err := txWorldState.CheckAndUpdate()
				updatedAt := time.Now().UnixNano()
				update += updatedAt - updateAt
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"tx":         tx,
						"err":        err,
						"giveback":   giveback,
						"dependency": dependency,
					}).Info("CheckAndUpdate invalid tx.")
					unpacked++
					conflict++

					if err := pool.Push(tx); err != nil {
						logging.VLog().WithFields(logrus.Fields{
							"block": block,
							"tx":    tx,
							"err":   err,
						}).Info("Failed to giveback the tx.")
					}

					fromBlacklist.Delete(tx.from.address.Hex())
					fromBlacklist.Delete(tx.to.address.Hex())
					toBlacklist.Delete(tx.from.address.Hex())
					toBlacklist.Delete(tx.to.address.Hex())

					<-mergeCh // unlock
					return
				}
				logging.VLog().WithFields(logrus.Fields{
					"tx": tx,
				}).Debug("packed tx.")
				packed++

				transactions = append(transactions, tx)
				txid := tx.Hash().String()
				dag.AddNode(txid)
				for _, node := range dependency {
					dag.AddEdge(node, txid)
				}
				fromBlacklist.Delete(tx.from.address.Hex())
				fromBlacklist.Delete(tx.to.address.Hex())
				toBlacklist.Delete(tx.from.address.Hex())
				toBlacklist.Delete(tx.to.address.Hex())

				<-mergeCh // unlock
				return
			}()

			if over {
				return
			}
		}
	}()

	<-deadlineTimer.C
	mergeCh <- true // lock
	over = true
	block.transactions = transactions
	block.dependency = dag
	<-mergeCh // unlock

	overAt := time.Now().UnixNano()
	size := int64(len(block.transactions))
	if size == 0 {
		size = 1
	}
	averPacking := packing / size
	averPrepare := prepare / size
	averExecute := execute / size
	averUpdate := update / size

	logging.VLog().WithFields(logrus.Fields{
		"try":          try,
		"failed":       failed,
		"expired":      expired,
		"conflict":     conflict,
		"fetch":        fetch,
		"bucket":       bucket,
		"avgPacking":   averPacking,
		"avgPrepare":   averPrepare,
		"avgExecute":   averExecute,
		"avgUpdate":    averUpdate,
		"parallel":     parallel,
		"packing":      packing,
		"execute":      execute,
		"prepare":      prepare,
		"update":       update,
		"diff-all":     overAt - beginAt,
		"core-packing": execute + prepare + update,
		"packed":       len(block.transactions),
		"dag":          block.dependency,
	}).Info("CollectTransactions")
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

	defer block.RollBack()

	if err := block.rewardCoinbaseForGas(); err != nil {
		return err
	}
	if err := block.WorldState().Flush(); err != nil {
		return err
	}
	block.header.stateRoot = block.WorldState().AccountsRoot()
	block.header.txsRoot = block.WorldState().TxsRoot()
	block.header.eventsRoot = block.WorldState().EventsRoot()
	block.header.consensusRoot = block.WorldState().ConsensusRoot()

	hash, err := block.calHash()
	if err != nil {
		return err
	}
	block.header.hash = hash
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
	random := ""
	if RandomAvailableAtHeight(block.height) && block.header.random != nil {
		if block.header.random.VrfSeed != nil {
			random += "/vrf_seed/" + byteutils.Hex(block.header.random.VrfSeed)
		}
		if block.header.random.VrfProof != nil {
			random += "/vrf_proof/" + byteutils.Hex(block.header.random.VrfProof)
		}
	}
	return fmt.Sprintf(`{"height": %d, "hash": "%s", "parent_hash": "%s", "acc_root": "%s", "timestamp": %d, "tx": %d, "miner": "%s", "random": "%s"}`,
		block.height,
		block.header.hash,
		block.header.parentHash,
		block.header.stateRoot,
		block.header.timestamp,
		len(block.transactions),
		byteutils.Hash(block.header.consensusRoot.Proposer).Base58(),
		random,
	)
}

// VerifyExecution execute the block and verify the execution result.
func (block *Block) VerifyExecution() error {
	startAt := time.Now().Unix()

	if err := block.Begin(); err != nil {
		return err
	}

	beganAt := time.Now().Unix()

	if err := block.execute(); err != nil {
		block.RollBack()
		return err
	}

	executedAt := time.Now().Unix()

	if err := block.verifyState(); err != nil {
		block.RollBack()
		return err
	}

	commitAt := time.Now().Unix()

	block.Commit()

	endAt := time.Now().Unix()

	logging.VLog().WithFields(logrus.Fields{
		"start":        startAt,
		"end":          endAt,
		"commit":       commitAt,
		"diff-all":     endAt - startAt,
		"diff-commit":  endAt - commitAt,
		"diff-begin":   beganAt - startAt,
		"diff-execute": executedAt - startAt,
		"diff-verify":  commitAt - executedAt,
		"block":        block,
		"txs":          len(block.Transactions()),
	}).Info("Verify txs.")

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
		}).Info("Failed to check chainid.")
		metricsInvalidBlock.Inc(1)
		return ErrInvalidChainID
	}

	// verify transactions integrity.
	for _, tx := range block.transactions {
		if err := tx.VerifyIntegrity(block.header.chainID); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"tx":  tx,
				"err": err,
			}).Info("Failed to verify tx's integrity.")
			metricsInvalidBlock.Inc(1)
			return err
		}
	}

	// verify block hash.
	wantedHash, err := block.calHash()
	if err != nil {
		return err
	}
	if !wantedHash.Equals(block.Hash()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": wantedHash,
			"actual": block.Hash(),
			"err":    err,
		}).Info("Failed to check block's hash.")
		metricsInvalidBlock.Inc(1)
		return ErrInvalidBlockHash
	}

	// verify the block is acceptable by consensus.
	if err := consensus.VerifyBlock(block); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
			"err":   err,
		}).Info("Failed to verify block.")
		metricsInvalidBlock.Inc(1)
		return err
	}

	return nil
}

// verifyState return state verify result.
func (block *Block) verifyState() error {
	// verify state root.
	if !byteutils.Equal(block.WorldState().AccountsRoot(), block.StateRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.StateRoot(),
			"actual": block.WorldState().AccountsRoot(),
		}).Info("Failed to verify state.")
		return ErrInvalidBlockStateRoot
	}

	// verify transaction root.
	if !byteutils.Equal(block.WorldState().TxsRoot(), block.TxsRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.TxsRoot(),
			"actual": block.WorldState().TxsRoot(),
		}).Info("Failed to verify txs.")
		return ErrInvalidBlockTxsRoot
	}

	// verify events root.
	if !byteutils.Equal(block.WorldState().EventsRoot(), block.EventsRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.EventsRoot(),
			"actual": block.WorldState().EventsRoot(),
		}).Info("Failed to verify events.")
		return ErrInvalidBlockEventsRoot
	}

	// verify transaction root.
	if !reflect.DeepEqual(block.WorldState().ConsensusRoot(), block.ConsensusRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": block.ConsensusRoot(),
			"actual": block.WorldState().ConsensusRoot(),
		}).Info("Failed to verify dpos context.")
		return ErrInvalidBlockConsensusRoot
	}
	return nil
}

type verifyCtx struct {
	mergeCh chan bool
	block   *Block
}

// Execute block and return result.
func (block *Block) execute() error {
	startAt := time.Now().UnixNano()

	if err := block.rewardCoinbaseForMint(); err != nil {
		return err
	}

	context := &verifyCtx{
		mergeCh: make(chan bool, 1),
		block:   block,
	}
	parallelNum := VerifyParallelNum

	if !WsResetRecordDependencyAtHeight(block.Height()) && len(block.transactions) > 0 {
		addrs := make(map[byteutils.HexHash]bool)
		for _, tx := range block.transactions {
			if _, ok := addrs[tx.to.address.Hex()]; ok {
				parallelNum = 1
				break
			}
			addrs[tx.to.address.Hex()] = true
		}
	}

	dispatcher := dag.NewDispatcher(block.dependency, parallelNum, int64(VerifyExecutionTimeout), context, func(node *dag.Node, context interface{}) error { // TODO: if system occurs, the block won't be retried any more
		ctx := context.(*verifyCtx)
		block := ctx.block
		mergeCh := ctx.mergeCh

		idx := node.Index()
		if idx < 0 || idx > len(block.transactions)-1 {
			return ErrInvalidDagBlock
		}
		tx := block.transactions[idx]

		logging.VLog().WithFields(logrus.Fields{
			"tx.hash": tx.hash,
		}).Debug("execute tx.")

		metricsTxExecute.Mark(1)

		mergeCh <- true
		txWorldState, err := block.WorldState().Prepare(tx.Hash().String())
		if err != nil {
			<-mergeCh
			return err
		}
		<-mergeCh

		if _, err = block.ExecuteTransaction(tx, txWorldState); err != nil {
			return err
		}

		mergeCh <- true
		if _, err := txWorldState.CheckAndUpdate(); err != nil {
			return err
		}
		<-mergeCh

		return nil
	})

	start := time.Now().UnixNano()
	if err := dispatcher.Run(); err != nil {
		transactions := []string{}
		for k, tx := range block.transactions {
			txInfo := fmt.Sprintf("{Index: %d, Tx: %s}", k, tx.String())
			transactions = append(transactions, txInfo)
		}
		logging.VLog().WithFields(logrus.Fields{
			"dag": block.dependency.String(),
			"txs": transactions,
			"err": err,
		}).Info("Failed to verify txs in block.")
		return err
	}
	end := time.Now().UnixNano()

	if len(block.transactions) != 0 {
		metricsTxVerifiedTime.Update((end - start) / int64(len(block.transactions)))
	} else {
		metricsTxVerifiedTime.Update(0)
	}

	if err := block.rewardCoinbaseForGas(); err != nil {
		return err
	}
	if err := block.WorldState().Flush(); err != nil {
		return err
	}

	block.sealed = true

	endAt := time.Now().UnixNano()
	metricsBlockVerifiedTime.Update(endAt - startAt)
	metricsTxsInBlock.Update(int64(len(block.transactions)))

	return nil
}

//Dynasty return dynasty
func (block *Block) Dynasty() ([]byteutils.Hash, error) {
	ws, err := block.WorldState().Clone()
	if err != nil {
		return nil, err
	}
	return ws.Dynasty()
}

//DynastyRoot return dynasty root
func (block *Block) DynastyRoot() (byteutils.Hash, error) {
	ws, err := block.WorldState().Clone()
	if err != nil {
		return nil, err
	}
	return ws.DynastyRoot(), nil
}

// GetAccount return the account with the given address on this block.
func (block *Block) GetAccount(address byteutils.Hash) (state.Account, error) {
	worldState, err := block.WorldState().Clone()
	if err != nil {
		return nil, err
	}
	return worldState.GetOrCreateUserAccount(address)
}

// FetchEvents fetch events by txHash.
func (block *Block) FetchEvents(txHash byteutils.Hash) ([]*state.Event, error) {
	worldState, err := block.WorldState().Clone()
	if err != nil {
		return nil, err
	}
	return worldState.FetchEvents(txHash)
}

// FetchExecutionResultEvent fetch execution result event by txHash.
func (block *Block) FetchExecutionResultEvent(txHash byteutils.Hash) (*state.Event, error) {
	worldState, err := block.WorldState().Clone()
	if err != nil {
		return nil, err
	}
	events, err := worldState.FetchEvents(txHash)
	if err != nil {
		return nil, err
	}

	if events != nil && len(events) > 0 {
		idx := len(events) - 1
		event := events[idx]
		if event.Topic != TopicTransactionExecutionResult {
			logging.VLog().WithFields(logrus.Fields{
				"tx":     txHash,
				"events": events,
			}).Info("Failed to locate the result event")
			return nil, ErrInvalidTransactionResultEvent
		}
		return event, nil
	}
	return nil, ErrNotFoundTransactionResultEvent
}

func (block *Block) nebulasRewardAddress() *Address {
	if block.ChainID() == MainNetID {
		if NbreSplitAtHeight(block.height) {
			return NebulasRewardAddressV2
		} else {
			return NebulasRewardAddress
		}
	} else if block.ChainID() == TestNetID {
		return block.Coinbase()
	} else {
		return block.Coinbase()
	}
}

func (block *Block) dipRewardAddress() *Address {
	if NbreSplitAtHeight(block.height) {
		return DIPRewardAddressV2
	} else {
		return block.dip.RewardAddress()
	}
}

func (block *Block) rewardCoinbaseForMint() error {
	if NodeUpdateAtHeight(block.height) {
		coinbaseAddr := block.Coinbase().Bytes()
		coinbaseAcc, err := block.WorldState().GetOrCreateUserAccount(coinbaseAddr)
		if err != nil {
			return err
		}
		if err = coinbaseAcc.AddBalance(BlockRewardV3); err != nil {
			return err
		}

		govAcc, err := block.WorldState().GetOrCreateUserAccount(NodeGovernanceContract().Bytes())
		if err != nil {
			return err
		}
		if err = govAcc.AddBalance(GovernanceReward); err != nil {
			return err
		}
	} else if NbreAvailableHeight(block.height) {
		//after NbreAvailableHeight, reward give to coinbase, nebulas and dip address.
		// the percent is: 2% coinbase, 1% nebulas,1% dip

		// coinbase reward
		coinbaseAddr := block.Coinbase().Bytes()
		coinbaseAcc, err := block.WorldState().GetOrCreateUserAccount(coinbaseAddr)
		if err != nil {
			return err
		}
		if err = coinbaseAcc.AddBalance(BlockRewardV2); err != nil {
			return err
		}

		// nebulas reward
		nebulasAddr := block.nebulasRewardAddress().Bytes()
		nebulasAcc, err := block.WorldState().GetOrCreateUserAccount(nebulasAddr)
		if err != nil {
			return err
		}
		if err = nebulasAcc.AddBalance(NebulasRewardV2); err != nil {
			return err
		}

		// dip reward.
		dipAddr := block.dipRewardAddress().Bytes()
		dipAcc, err := block.WorldState().GetOrCreateUserAccount(dipAddr)
		if err != nil {
			return err
		}
		if err = dipAcc.AddBalance(DIPRewardV2); err != nil {
			return err
		}
	} else {
		coinbaseAddr := block.Coinbase().Bytes()
		coinbaseAcc, err := block.WorldState().GetOrCreateUserAccount(coinbaseAddr)
		if err != nil {
			return err
		}
		if err = coinbaseAcc.AddBalance(BlockReward); err != nil {
			return err
		}
	}

	return nil
}

func (block *Block) rewardCoinbaseForGas() error {
	worldState := block.WorldState()
	coinbaseAddr := (byteutils.Hash)(block.Coinbase().Bytes())

	gasConsumed := worldState.GetGas()
	for from, gas := range gasConsumed {
		fromAddr, err := AddressParse(from)
		if err != nil {
			return err
		}
		if _, err := transfer(fromAddr.Bytes(), coinbaseAddr, gas, worldState); err != nil {
			return err
		}
	}
	return nil
}

func transfer(from, to byteutils.Hash, value *util.Uint128, ws WorldState) (bool, error) {
	fromAcc, err := ws.GetOrCreateUserAccount(from)
	if err != nil {
		return true, err
	}
	toAcc, err := ws.GetOrCreateUserAccount(to)
	if err != nil {
		return true, err
	}
	if err := fromAcc.SubBalance(value); err != nil {
		// Balance is not enough to transfer the value, won't giveback the tx
		return false, err
	}
	if err := toAcc.AddBalance(value); err != nil {
		// Balance plus value result in overflow, won't giveback the tx
		return false, err
	}
	// No error, won't giveback the tx
	return false, nil
}

// ExecuteTransaction execute the transaction
// return giveback, err
// system error: giveback == true
// logic error: giveback == false, expect Bigger Nonce
func (block *Block) ExecuteTransaction(tx *Transaction, ws WorldState) (bool, error) {
	if giveback, err := CheckTransaction(tx, ws); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"tx":  tx,
			"err": err,
		}).Info("Failed to check transaction")
		return giveback, err
	}

	if err := block.dip.CheckReward(tx); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"tx":  tx,
			"err": err,
		}).Info("Failed to check transaction dip reward.")
		return false, err
	}

	if giveback, err := VerifyExecution(tx, block, ws); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"tx":  tx,
			"err": err,
		}).Info("Failed to verify transaction execution")
		return giveback, err
	}

	if giveback, err := AcceptTransaction(tx, ws); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"tx":  tx,
			"err": err,
		}).Info("Failed to accept transaction")
		return giveback, err
	}

	return false, nil
}

// CheckContract check if contract is valid
func (block *Block) CheckContract(addr *Address) (state.Account, error) {

	worldState, err := block.worldState.Clone()
	if err != nil {
		return nil, err
	}
	return CheckContract(addr, worldState)
}

// GetTransaction from txs Trie
func (block *Block) GetTransaction(hash byteutils.Hash) (*Transaction, error) {
	worldState, err := block.worldState.Clone()
	if err != nil {
		return nil, err
	}
	return GetTransaction(hash, worldState)
}

// CalHash calculate the hash of block.
func (block *Block) calHash() (byteutils.Hash, error) {
	hasher := sha3.New256()

	consensusRoot, err := proto.Marshal(block.ConsensusRoot())
	if err != nil {
		return nil, err
	}

	pbDep, err := block.dependency.ToProto()
	if err != nil {
		return nil, err
	}
	dependency, err := proto.Marshal(pbDep)
	if err != nil {
		return nil, err
	}

	hasher.Write(block.ParentHash())
	hasher.Write(block.StateRoot())
	hasher.Write(block.TxsRoot())
	hasher.Write(block.EventsRoot())
	hasher.Write(consensusRoot)
	hasher.Write(dependency)
	hasher.Write(block.header.coinbase.address)
	hasher.Write(byteutils.FromInt64(block.header.timestamp))
	hasher.Write(byteutils.FromUint32(block.header.chainID))

	for _, tx := range block.transactions {
		hasher.Write(tx.Hash())
	}

	return hasher.Sum(nil), nil
}

// HashPbBlock return the hash of pb block.
func HashPbBlock(pbBlock *corepb.Block) (byteutils.Hash, error) {
	block := new(Block)
	if err := block.FromProto(pbBlock); err != nil {
		return nil, err
	}
	return block.calHash()
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

	block.sealed = true
	block.txPool = chain.txPool
	block.eventEmitter = chain.eventEmitter
	block.nvm = chain.nvm
	block.nr = chain.nr
	block.dip = chain.dip
	block.storage = chain.storage
	return block, nil
}

// MockBlock nf/nvm/engine.CheckV8Run()  & cmd/v8/main.go
func MockBlock(header *BlockHeader, height uint64) *Block {
	return &Block{
		header: header,
		height: height,
	}
}
