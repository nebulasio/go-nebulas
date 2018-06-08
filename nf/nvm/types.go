package nvm

import (
	"errors"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// Error Types
var (
	ErrEngineRepeatedStart      = errors.New("engine repeated start")
	ErrEngineNotStart           = errors.New("engine not start")
	ErrContextConstructArrEmpty = errors.New("context construct err by args empty")
	ErrEngineNotFound           = errors.New("Failed to get engine")

	ErrDisallowCallPrivateFunction     = errors.New("disallow call private function")
	ErrExecutionTimeout                = errors.New("execution timeout")
	ErrInsufficientGas                 = errors.New("insufficient gas")
	ErrExceedMemoryLimits              = errors.New("exceed memory limits")
	ErrInjectTracingInstructionFailed  = errors.New("inject tracing instructions failed")
	ErrTranspileTypeScriptFailed       = errors.New("transpile TypeScript failed")
	ErrUnsupportedSourceType           = errors.New("unsupported source type")
	ErrArgumentsFormat                 = errors.New("arguments format error")
	ErrLimitHasEmpty                   = errors.New("limit args has empty")
	ErrSetMemorySmall                  = errors.New("set memory small than v8 limit")
	ErrDisallowCallNotStandardFunction = errors.New("disallow call not standard function")

	ErrNvmNumLimit          = errors.New("out of limit nvm count")
	ErrInnerTransferFailed  = errors.New("inner transfer failed")
	ErrInnerInsufficientGas = errors.New("preparation inner nvm insufficient gas")
	ErrInnerInsufficientMem = errors.New("preparation inner nvm insufficient mem")
)

//define
var (
	EventNameSpaceContract    = "chain.contract" //ToRefine: move to core
	InnerTransactionErrPrefix = "inner transation err ["
	InnerTransactionErrEnding = "] engine index:%v"
)

//common err
var (
	ErrKeyNotFound = storage.ErrKeyNotFound
)

//transfer err code enum
const (
	TransferFuncSuccess = iota
	TransferSuccess
	TransferGetEngineErr
	TransferAddressParseErr
	TransferGetAccountErr
	TransferStringToBigIntErr
	TransferSubBalance
	TransferAddBalance
	TransferRecordEventFailed
	TransferAddressFailed
)

//MultiV8error err info, err only in InnerContractFunc .so not to deine #
type MultiV8error struct {
	errCode uint32
	index   uint32
	errStr  string
}

//nvm args define //TODO: 确定所有值的大小
var (
	MultiNvmMax               = 3
	GetTxByHashFuncCost       = 1000
	GetAccountStateFuncCost   = 1000
	TransferFuncCost          = 2000
	VerifyAddressFuncCost     = 100
	GetContractSourceFuncCost = 100
	InnerContractFuncCost     = 100
)

// Block interface breaks cycle import dependency and hides unused services.
type Block interface {
	Hash() byteutils.Hash
	Height() uint64 // ToAdd: timestamp interface
	Timestamp() int64
	RandomSeed() string
	RandomAvailable() bool
	DateAvailable() bool
}

// Transaction interface breaks cycle import dependency and hides unused services.
type Transaction interface {
	ChainID() uint32
	Hash() byteutils.Hash
	From() *core.Address
	To() *core.Address
	Value() *util.Uint128
	Nonce() uint64
	Timestamp() int64
	GasPrice() *util.Uint128
	GasLimit() *util.Uint128
}

// Account interface breaks cycle import dependency and hides unused services.
type Account interface {
	Address() byteutils.Hash
	Balance() *util.Uint128
	Nonce() uint64
	AddBalance(value *util.Uint128) error
	SubBalance(value *util.Uint128) error
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Del(key []byte) error
}

// WorldState interface breaks cycle import dependency and hides unused services.
type WorldState interface {
	GetOrCreateUserAccount(addr byteutils.Hash) (state.Account, error)
	GetTx(txHash byteutils.Hash) ([]byte, error)
	RecordEvent(txHash byteutils.Hash, event *state.Event)
	CreateContractAccount(owner byteutils.Hash, birthPlace byteutils.Hash) (state.Account, error)
	Dynasty() ([]byteutils.Hash, error)
	DynastyRoot() byteutils.Hash
	FetchEvents(byteutils.Hash) ([]*state.Event, error)
	GetContractAccount(addr byteutils.Hash) (state.Account, error)
	PutTx(txHash byteutils.Hash, txBytes []byte) error
	RecordGas(from string, gas *util.Uint128) error
	Reset(addr byteutils.Hash) error //Need to consider risk
}
