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
	"time"

	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
)

// const definition
const (
	// OptimizeHeight after this height,
	// update transaction execution result event,
	// update binary transaction payload.
	OptimizeHeight = 480000

	// update deploy execution, from & to must equal
	NewOptimizeHeight = 750000
)

// Payload Types
const (
	TxPayloadBinaryType    = "binary"
	TxPayloadDeployType    = "deploy"
	TxPayloadCallType      = "call"
	TxPayloadDelegateType  = "delegate"
	TxPayloadCandidateType = "candidate"
)

const (
	// TxExecutionFailed failed status for transaction execute result.
	TxExecutionFailed = 0

	// TxExecutionSuccess success status for transaction execute result.
	TxExecutionSuccess = 1

	// TxExecutionPendding pendding status when transaction in transaction pool.
	TxExecutionPendding = 2
)

// Error Types
var (
	ErrInvalidBlockOnCanonicalChain                      = errors.New("invalid block, it's not on canonical chain")
	ErrNotBlockInCanonicalChain                          = errors.New("cannot find the block in canonical chain")
	ErrInvalidBlockCannotFindParentInLocal               = errors.New("invalid block received, download its parent from others")
	ErrCannotFindBlockAtGivenHeight                      = errors.New("cannot find a block at given height which is less than tail block's height")
	ErrInvalidBlockCannotFindParentInLocalAndTryDownload = errors.New("invalid block received, download its parent from others")
	ErrInvalidBlockCannotFindParentInLocalAndTrySync     = errors.New("invalid block received, sync its parent from others")

	ErrLinkToWrongParentBlock = errors.New("link the block to a block who is not its parent")
	ErrMissingParentBlock     = errors.New("cannot find the block's parent block in storage")
	ErrInvalidBlockHash       = errors.New("invalid block hash")
	ErrDoubleSealBlock        = errors.New("cannot seal a block twice")
	ErrDuplicatedBlock        = errors.New("duplicated block")
	ErrDoubleBlockMinted      = errors.New("double block minted")

	ErrInvalidChainID           = errors.New("invalid transaction chainID")
	ErrInvalidTransactionSigner = errors.New("transaction recover public key address not equal to from")
	ErrInvalidTransactionHash   = errors.New("invalid transaction hash")
	ErrInvalidSignature         = errors.New("invalid transaction signature")
	ErrInvalidTxPayloadType     = errors.New("invalid transaction data payload type")

	ErrInsufficientBalance                = errors.New("insufficient balance")
	ErrBelowGasPrice                      = errors.New("below the gas price")
	ErrOutOfGasLimit                      = errors.New("out of gas limit")
	ErrTxExecutionFailed                  = errors.New("transaction execution failed")
	ErrContractDeployFailed               = errors.New("contract deploy failed")
	ErrContractNotFound                   = errors.New("contract not found")
	ErrContractTransactionAddressNotEqual = errors.New("contract transaction from-address not equal to to-address")

	ErrDuplicatedTransaction = errors.New("duplicated transaction")
	ErrSmallTransactionNonce = errors.New("cannot accept a transaction with smaller nonce")
	ErrLargeTransactionNonce = errors.New("cannot accept a transaction with too bigger nonce")

	ErrInvalidAddress           = errors.New("address: invalid address")
	ErrInvalidAddressDataLength = errors.New("address: invalid address data length")

	ErrInvalidCandidatePayloadAction     = errors.New("invalid transaction candidate payload action")
	ErrInvalidDelegatePayloadAction      = errors.New("invalid transaction vote payload action")
	ErrInvalidDelegateToNonCandidate     = errors.New("cannot delegate to non-candidate")
	ErrInvalidUnDelegateFromNonDelegatee = errors.New("cannot un-delegate from non-delegatee")

	ErrCloneWorldState           = errors.New("Failed to clone world state")
	ErrCloneAccountState         = errors.New("Failed to clone account state")
	ErrCloneTxsState             = errors.New("Failed to clone txs state")
	ErrCloneEventsState          = errors.New("Failed to clone events state")
	ErrInvalidBlockStateRoot     = errors.New("invalid block state root hash")
	ErrInvalidBlockTxsRoot       = errors.New("invalid block txs root hash")
	ErrInvalidBlockEventsRoot    = errors.New("invalid block events root hash")
	ErrInvalidBlockConsensusRoot = errors.New("invalid block consensus root hash")

	ErrCannotRevertLIB        = errors.New("cannot revert latest irreversible block")
	ErrCannotLoadGenesisBlock = errors.New("cannot load genesis block from storage")
	ErrCannotLoadLIBBlock     = errors.New("cannot load tail block from storage")
	ErrCannotLoadTailBlock    = errors.New("cannot load latest irreversible block from storage")
	ErrGenesisConfNotMatch    = errors.New("Failed to load genesis from storage, different with genesis conf")
	ErrInvalidConfigChainID   = errors.New("invalid chainID, genesis chainID not equal to chainID in config")
)

// Default gas count
var (
	DefaultPayloadGas = util.NewUint128FromInt(1)
)

// TxPayload stored in tx
type TxPayload interface {
	ToBytes() ([]byte, error)
	BaseGasCount() *util.Uint128
	Execute(tx *Transaction, block *Block, txWorldState state.TxWorldState) (*util.Uint128, string, error)
}

// MessageType
const (
	MessageTypeNewBlock             = "newblock"
	MessageTypeDownloadedBlock      = "dlblock"
	MessageTypeDownloadedBlockReply = "dlreply"
	MessageTypeNewTx                = "newtx"
)

// Consensus interface of consensus algorithm.
type Consensus interface {
	Setup(Neblet) error
	Start()
	Stop()

	EnableMining(string) error
	DisableMining() error
	Enable() bool

	ResumeMining()
	SuspendMining()
	Pending() bool

	VerifyBlock(*Block, *Block) error
	FastVerifyBlock(*Block) error
	ForkChoice() error
	UpdateLIB()

	NewState(byteutils.Hash, storage.Storage) (state.ConsensusState, error)
	CheckTimeout(*Block) bool

	GenesisConsensusState(*BlockChain, *corepb.Genesis) (state.ConsensusState, error)
}

// SyncService interface of sync service
type SyncService interface {
	Start()
	Stop()

	StartActiveSync() bool
	StopActiveSync()
	WaitingForFinish() error
	IsActiveSyncing() bool
}

// Manager interface of account mananger
type Manager interface {
	NewAccount([]byte) (*Address, error)
	Accounts() []*Address

	Unlock(*Address, []byte, time.Duration) error
	Lock(*Address) error

	SignBlock(*Address, *Block) error
	SignTransaction(*Address, *Transaction) error
	SignTransactionWithPassphrase(*Address, *Transaction, []byte) error

	Update(*Address, []byte, []byte) error
	Load([]byte, []byte) (*Address, error)
	Import([]byte, []byte) (*Address, error)
	Delete(*Address, []byte) error
}

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Genesis() *corepb.Genesis
	Config() *nebletpb.Config
	Storage() storage.Storage
	EventEmitter() *EventEmitter
	Consensus() Consensus
	BlockChain() *BlockChain
	NetService() net.Service
	AccountManager() Manager
}
