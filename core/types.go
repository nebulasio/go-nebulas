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

	"github.com/nebulasio/go-nebulas/crypto/keystore"
	nebletpb "github.com/nebulasio/go-nebulas/neblet/pb"

	"github.com/nebulasio/go-nebulas/util/byteutils"

	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/net"

	"regexp"

	consensuspb "github.com/nebulasio/go-nebulas/consensus/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
)

// Payload Types
const (
	TxPayloadBinaryType   = "binary"
	TxPayloadDeployType   = "deploy"
	TxPayloadCallType     = "call"
	TxPayloadProtocolType = "protocol"
	TxPayloadDipType      = "dip"
	TxPayloadPodType      = "pod"
)

// Const.
const (
	SourceTypeJavaScript = "js"
	SourceTypeTypeScript = "ts"
)

// Const
const (
	ContractAcceptFunc = "accept"
)

const (
	NrStartHeight    = 2307000
	NrIntervalHeight = 40320
)

var (
	// PublicFuncNameChecker     in smart contract
	PublicFuncNameChecker = regexp.MustCompile("^[a-zA-Z$][A-Za-z0-9_$]*$")
)

const (
	// TxExecutionFailed failed status for transaction execute result.
	TxExecutionFailed = 0

	// TxExecutionSuccess success status for transaction execute result.
	TxExecutionSuccess = 1

	// TxExecutionPendding pendding status when transaction in transaction pool.
	TxExecutionPendding = 2
)

const (
	//InnerTransactionNonce inner tx nonce
	InnerTransactionNonce = 0
)

// Error Types
var (
	ErrInvalidBlockOnCanonicalChain                      = errors.New("invalid block, it's not on canonical chain")
	ErrNotBlockInCanonicalChain                          = errors.New("cannot find the block in canonical chain")
	ErrInvalidBlockCannotFindParentInLocal               = errors.New("invalid block received, download its parent from others")
	ErrCannotFindBlockAtGivenHeight                      = errors.New("cannot find a block at given height which is less than tail block's height")
	ErrInvalidBlockCannotFindParentInLocalAndTryDownload = errors.New("invalid block received, download its parent from others")
	ErrInvalidBlockCannotFindParentInLocalAndTrySync     = errors.New("invalid block received, sync its parent from others")
	ErrBlockNotFound                                     = errors.New("block not found in blockchain cache nor chain")

	ErrInvalidConfigChainID          = errors.New("invalid chainID, genesis chainID not equal to chainID in config")
	ErrCannotLoadGenesisConf         = errors.New("cannot load genesis conf")
	ErrGenesisNotEqualChainIDInDB    = errors.New("Failed to check. genesis chainID not equal in db")
	ErrGenesisNotEqualDynastyInDB    = errors.New("Failed to check. genesis dynasty not equal in db")
	ErrGenesisNotEqualTokenInDB      = errors.New("Failed to check. genesis TokenDistribution not equal in db")
	ErrGenesisNotEqualDynastyLenInDB = errors.New("Failed to check. genesis dynasty length not equal in db")
	ErrGenesisNotEqualTokenLenInDB   = errors.New("Failed to check. genesis TokenDistribution length not equal in db")

	ErrLinkToWrongParentBlock = errors.New("link the block to a block who is not its parent")
	ErrMissingParentBlock     = errors.New("cannot find the block's parent block in storage")
	ErrInvalidBlockHash       = errors.New("invalid block hash")
	ErrDoubleSealBlock        = errors.New("cannot seal a block twice")
	ErrDuplicatedBlock        = errors.New("duplicated block")
	ErrDoubleBlockMinted      = errors.New("double block minted")
	ErrVRFProofFailed         = errors.New("VRF proof failed")
	ErrInvalidBlockRandom     = errors.New("invalid block random")
	ErrInvalidBlockProposer   = errors.New("invalid block proposer")

	ErrInvalidChainID           = errors.New("invalid transaction chainID")
	ErrInvalidTransactionSigner = errors.New("invalid transaction signer")
	ErrInvalidTransactionHash   = errors.New("invalid transaction hash")
	ErrInvalidSignature         = errors.New("invalid transaction signature")
	ErrInvalidTxPayloadType     = errors.New("invalid transaction data payload type")
	ErrInvalidGasPrice          = errors.New("invalid gas price, should be in (0, 10^12]")
	ErrInvalidGasLimit          = errors.New("invalid gas limit, should be in (0, 5*10^10]")

	ErrNoTimeToPackTransactions       = errors.New("no time left to pack transactions in a block")
	ErrTxDataPayLoadOutOfMaxLength    = errors.New("data's payload is out of max data length")
	ErrTxDataBinPayLoadOutOfMaxLength = errors.New("data's payload is out of max data length in a binary tx")
	ErrNilArgument                    = errors.New("argument(s) is nil")
	ErrInvalidArgument                = errors.New("invalid argument(s)")

	ErrInsufficientBalance                = errors.New("insufficient balance")
	ErrBelowGasPrice                      = errors.New("below the gas price")
	ErrGasCntOverflow                     = errors.New("the count of gas used is overflow")
	ErrGasFeeOverflow                     = errors.New("the fee of gas used is overflow")
	ErrInvalidTransfer                    = errors.New("transfer error: overflow or insufficient balance")
	ErrGasLimitLessOrEqualToZero          = errors.New("gas limit less or equal to 0")
	ErrOutOfGasLimit                      = errors.New("out of gas limit")
	ErrTxExecutionFailed                  = errors.New("transaction execution failed")
	ErrZeroGasPrice                       = errors.New("gas price should be greater than zero")
	ErrZeroGasLimit                       = errors.New("gas limit should be greater than zero")
	ErrContractDeployFailed               = errors.New("contract deploy failed")
	ErrContractCheckFailed                = errors.New("contract check failed")
	ErrContractTransactionAddressNotEqual = errors.New("contract transaction from-address not equal to to-address")

	ErrDuplicatedTransaction = errors.New("duplicated transaction")
	ErrSmallTransactionNonce = errors.New("cannot accept a transaction with smaller nonce")
	ErrLargeTransactionNonce = errors.New("cannot accept a transaction with too bigger nonce")

	ErrInvalidAddress         = errors.New("address: invalid address")
	ErrInvalidAddressFormat   = errors.New("address: invalid address format")
	ErrInvalidAddressType     = errors.New("address: invalid address type")
	ErrInvalidAddressChecksum = errors.New("address: invalid address checksum")

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
	ErrInvalidProtoToBlock       = errors.New("protobuf message cannot be converted into Block")
	ErrInvalidProtoToBlockHeader = errors.New("protobuf message cannot be converted into BlockHeader")
	ErrInvalidProtoToTransaction = errors.New("protobuf message cannot be converted into Transaction")
	ErrInvalidTransactionData    = errors.New("invalid data in tx from Proto")
	ErrInvalidDagBlock           = errors.New("block's dag is incorrect")

	ErrCannotRevertLIB        = errors.New("cannot revert latest irreversible block")
	ErrCannotLoadGenesisBlock = errors.New("cannot load genesis block from storage")
	ErrCannotLoadLIBBlock     = errors.New("cannot load tail block from storage")
	ErrCannotLoadTailBlock    = errors.New("cannot load latest irreversible block from storage")
	ErrGenesisConfNotMatch    = errors.New("Failed to load genesis from storage, different with genesis conf")

	ErrInvalidDeploySource     = errors.New("invalid source of deploy payload")
	ErrInvalidDeploySourceType = errors.New("invalid source type of deploy payload")
	ErrInvalidCallFunction     = errors.New("invalid function of call payload")

	ErrInvalidTransactionResultEvent  = errors.New("invalid transaction result event, the last event in tx's events should be result event")
	ErrNotFoundTransactionResultEvent = errors.New("transaction result event is not found ")

	// nvm error
	ErrExecutionFailed = errors.New("execution failed")
	ErrUnexpected      = errors.New("Unexpected sys error")
	// multi nvm error
	ErrInnerExecutionFailed = errors.New("multi execution failed")
	ErrCreateInnerTx        = errors.New("Failed to create inner transaction")

	// access control
	ErrUnsupportedKeyword      = errors.New("transaction data has unsupported keyword")
	ErrUnsupportedFunction     = errors.New("transaction payload has unsupported function")
	ErrRestrictedFromAddress   = errors.New("transaction from address is restricted")
	ErrRestrictedToAddress     = errors.New("transaction to address is restricted")
	ErrNrc20ArgsCheckFailed    = errors.New("transaction nrc20 args check failed")
	ErrNrc20AddressCheckFailed = errors.New("transaction nrc20 address check failed")
	ErrNrc20ValueCheckFailed   = errors.New("transaction nrc20 value check failed")

	// func deprecated
	ErrFuncDeprecated = errors.New("function deprecated")

	ErrBlockStateCheckFailed = errors.New("Failed to check block state")
)

// Default gas count
var (
	DefaultPayloadGas, _ = util.NewUint128FromInt(1)

	// DefaultLimitsOfTotalMemorySize default limits of total memory size
	DefaultLimitsOfTotalMemorySize uint64 = 40 * 1000 * 1000
)

// TxPayload stored in tx
type TxPayload interface {
	ToBytes() ([]byte, error)
	BaseGasCount() *util.Uint128
	Execute(limitedGas *util.Uint128, tx *Transaction, block *Block, ws WorldState) (*util.Uint128, string, error)
}

// MessageType
const (
	MessageTypeNewBlock                   = "newblock"
	MessageTypeParentBlockDownloadRequest = "dlblock"
	MessageTypeBlockDownloadResponse      = "dlreply"
	MessageTypeNewTx                      = "newtx"
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

	Serial(timestamp int64) int64
	VerifyBlock(*Block) error
	ForkChoice() error
	UpdateLIB([]byteutils.Hash)

	NewState(*consensuspb.ConsensusRoot, storage.Storage, bool) (state.ConsensusState, error)
	GenesisConsensusState(*BlockChain, *corepb.Genesis) (state.ConsensusState, error)
	CheckTimeout(*Block) bool
	CheckDoubleMint(*Block) bool

	NumberOfBlocksInDynasty() uint64
}

// SyncService interface of sync service
type SyncService interface {
	Start()
	Stop()

	StartActiveSync() bool
	StopActiveSync()
	WaitingForFinish()
	IsActiveSyncing() bool
}

// AccountManager interface of account mananger
type AccountManager interface {
	NewAccount([]byte) (*Address, error)
	Accounts() []*Address

	Unlock(*Address, []byte, time.Duration) error
	Lock(*Address) error

	SignHash(*Address, byteutils.Hash, keystore.Algorithm) ([]byte, error)
	SignBlock(*Address, *Block) error
	GenerateRandomSeed(*Address, []byte, []byte) ([]byte, []byte, error)
	SignTransaction(*Address, *Transaction) error
	SignTransactionWithPassphrase(*Address, *Transaction, []byte) error

	Update(*Address, []byte, []byte) error
	Load([]byte, []byte) (*Address, error)
	LoadPrivate([]byte, []byte) (*Address, error)
	Import([]byte, []byte) (*Address, error)
	Remove(*Address, []byte) error
}

// NVM interface
type NVM interface {
	CreateEngine(block *Block, tx *Transaction, contract state.Account, ws WorldState) (SmartContractEngine, error)
	CheckV8Run() error
}

// SmartContractEngine interface
type SmartContractEngine interface {
	SetExecutionLimits(uint64, uint64) error
	DeployAndInit(source, sourceType, args string) (string, error)
	Call(source, sourceType, function, args string) (string, error)
	ExecutionInstructions() uint64
	Dispose()
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
	IsActiveSyncing() bool
	AccountManager() AccountManager
	Nvm() NVM
	Nr() NR
	Dip() Dip
	StartPprof(string) error
}

// WorldState needed by core
type WorldState interface {
	GetOrCreateUserAccount(addr byteutils.Hash) (state.Account, error)
	GetContractAccount(addr byteutils.Hash) (state.Account, error)
	CreateContractAccount(owner byteutils.Hash, birthPlace byteutils.Hash, contractMeta *corepb.ContractMeta) (state.Account, error)

	GetTx(txHash byteutils.Hash) ([]byte, error)
	PutTx(txHash byteutils.Hash, txBytes []byte) error

	RecordEvent(txHash byteutils.Hash, event *state.Event)
	FetchEvents(byteutils.Hash) ([]*state.Event, error)

	Dynasty() ([]byteutils.Hash, error)
	DynastyRoot() byteutils.Hash

	RecordGas(from string, gas *util.Uint128) error

	Reset(addr byteutils.Hash, isResetChangeLog bool) error
	GetBlockHashByHeight(height uint64) ([]byte, error)
	GetBlock(txHash byteutils.Hash) ([]byte, error)
}

type Data interface {
	ToBytes() ([]byte, error)
	FromBytes([]byte) error
}

// Nbre interface
type Nbre interface {
	Start() error
	Execute(command string, args ...interface{}) (interface{}, error)
	Stop()
}

type NR interface {
	GetNRListByHeight(height uint64) (Data, error)
	GetNRSummary(height uint64) (Data, error)
}

type Dip interface {
	Start()
	Stop()

	RewardAddress() *Address
	RewardValue() *util.Uint128

	GetDipList(height, version uint64) (Data, error)
	CheckReward(tx *Transaction) error
}
