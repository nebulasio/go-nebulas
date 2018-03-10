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

	"github.com/nebulasio/go-nebulas/consensus/pb"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
)

// Payload Types
const (
	TxPayloadBinaryType = "binary"
	TxPayloadDeployType = "deploy"
	TxPayloadCallType   = "call"
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
	ErrInvalidConfigChainID                              = errors.New("invalid chainID, genesis chainID not equal to chainID in config")
	ErrCannotLoadGenesisConf                             = errors.New("cannot load genesis conf")
	ErrGenesisNotEqualChainIDInDB                        = errors.New("Failed to check. genesis chainID not equal in db")
	ErrGenesisNotEqualDynastyInDB                        = errors.New("Failed to check. genesis dynasty not equal in db")
	ErrGenesisNotEqualTokenInDB                          = errors.New("Failed to check. genesis TokenDistribution not equal in db")
	ErrGenesisNotEqualDynastyLenInDB                     = errors.New("Failed to check. genesis dynasty length not equal in db")
	ErrGenesisNotEqualTokenLenInDB                       = errors.New("Failed to check. genesis TokenDistribution length not equal in db")

	ErrLinkToWrongParentBlock = errors.New("link the block to a block who is not its parent")
	ErrMissingParentBlock     = errors.New("cannot find the block's parent block in storage")
	ErrInvalidBlockHash       = errors.New("invalid block hash")
	ErrDuplicatedBlock        = errors.New("duplicated block")
	ErrDoubleBlockMinted      = errors.New("double block minted")

	ErrInvalidChainID           = errors.New("invalid transaction chainID")
	ErrInvalidTransactionSigner = errors.New("transaction recover public key address not equal to from")
	ErrInvalidTransactionHash   = errors.New("invalid transaction hash")
	ErrInvalidTxPayloadType     = errors.New("invalid transaction data payload type")

	ErrInsufficientBalance                = errors.New("insufficient balance")
	ErrBelowGasPrice                      = errors.New("below the gas price")
	ErrGasLimitLessOrEqualToZero          = errors.New("gas limit less or equal to 0")
	ErrOutOfGasLimit                      = errors.New("out of gas limit")
	ErrContractCheckFailed                = errors.New("contract check failed")
	ErrContractTransactionAddressNotEqual = errors.New("contract transaction from-address not equal to to-address")

	ErrDuplicatedTransaction = errors.New("duplicated transaction")
	ErrSmallTransactionNonce = errors.New("cannot accept a transaction with smaller nonce")
	ErrLargeTransactionNonce = errors.New("cannot accept a transaction with too bigger nonce")

	ErrInvalidAddress           = errors.New("address: invalid address")
	ErrInvalidAddressDataLength = errors.New("address: invalid address data length")

	ErrCloneWorldState           = errors.New("Failed to clone world state")
	ErrCloneAccountState         = errors.New("Failed to clone account state")
	ErrCloneTxsState             = errors.New("Failed to clone txs state")
	ErrCloneEventsState          = errors.New("Failed to clone events state")
	ErrInvalidBlockStateRoot     = errors.New("invalid block state root hash")
	ErrInvalidBlockTxsRoot       = errors.New("invalid block txs root hash")
	ErrInvalidBlockEventsRoot    = errors.New("invalid block events root hash")
	ErrInvalidBlockConsensusRoot = errors.New("invalid block consensus root hash")

	ErrCannotRevertLIB     = errors.New("cannot revert latest irreversible block")
	ErrCannotLoadTailBlock = errors.New("cannot load latest irreversible block from storage")

	ErrNoTimeToPackTransactions    = errors.New("no time left to pack transactions in a block")
	ErrTxDataPayLoadOutOfMaxLength = errors.New("data's payload is out of max data length")
	ErrNilArgument                 = errors.New("argument(s) is nil")
	ErrIllegalArgument             = errors.New("illegal argument(s)")

	ErrInvalidTransactionData   = errors.New("invalid data in tx from Proto")
	ErrCannotConvertTransaction = errors.New("proto message cannot be converted into Transaction")
)

// TxPayload stored in tx
type TxPayload interface {
	ToBytes() ([]byte, error)
	BaseGasCount() *util.Uint128
	Execute(block *Block, tx *Transaction) (*util.Uint128, string, error)
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

	VerifyBlock(*Block) error
	ForkChoice() error
	UpdateLIB()

	NewState(*consensuspb.ConsensusRoot, storage.Storage) (state.ConsensusState, error)
	GenesisState(*BlockChain, *corepb.Genesis) (state.ConsensusState, error)
	CheckTimeout(*Block) bool
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

// AccountManager interface of account mananger
type AccountManager interface {
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

// Engine interface breaks cycle import dependency and hides unused services.
type Engine interface {
	StartEngine(block *Block, tx *Transaction, owner, contract state.Account, state state.AccountState) error
	SetEngineExecutionLimits(limitsOfExecutionInstructions uint64) error
	DeployAndInitEngine(source, sourceType, args string) (string, error)
	CallEngine(source, sourceType, function, args string) (string, error)
	ExecutionInstructions() (uint64, error)
	DisposeEngine()
}

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Genesis() *corepb.Genesis
	SetGenesis(*corepb.Genesis)
	Config() *nebletpb.Config
	Storage() storage.Storage
	EventEmitter() *EventEmitter
	Consensus() Consensus
	BlockChain() *BlockChain
	NetService() net.Service
	AccountManager() AccountManager
	Nvm() Engine
	StartPprof(string) error
}
