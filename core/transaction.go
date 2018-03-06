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

	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

var (
	// TransactionMaxGasPrice max gasPrice:50 * 10 ** 9
	TransactionMaxGasPrice, _ = util.NewUint128FromString("50000000000")

	// TransactionMaxGas max gas:50 * 10 ** 9
	TransactionMaxGas, _ = util.NewUint128FromString("50000000000")

	// TransactionGasPrice default gasPrice : 10**6
	TransactionGasPrice, _ = util.NewUint128FromInt(1000000)

	// MinGasCountPerTransaction default gas for normal transaction
	MinGasCountPerTransaction, _ = util.NewUint128FromInt(20000)

	// GasCountPerByte per byte of data attached to a transaction gas cost
	GasCountPerByte, _ = util.NewUint128FromInt(1)

	// DelegateBaseGasCount is base gas count of delegate transaction
	DelegateBaseGasCount, _ = util.NewUint128FromInt(20000)

	// CandidateBaseGasCount is base gas count of candidate transaction
	CandidateBaseGasCount, _ = util.NewUint128FromInt(20000)

	// ZeroGasCount is zero gas count
	ZeroGasCount = util.NewUint128()
)

// TransactionEvent transaction event
type TransactionEvent struct {
	Hash    string `json:"hash"`
	Status  int8   `json:"status"`
	GasUsed string `json:"gas_used"`
	Error   string `json:"error"`
}

// Transaction type is used to handle all transaction data.
type Transaction struct {
	hash      byteutils.Hash
	from      *Address
	to        *Address
	value     *util.Uint128
	nonce     uint64
	timestamp int64
	data      *corepb.Data
	chainID   uint32
	gasPrice  *util.Uint128
	gasLimit  *util.Uint128

	// Signature
	alg  uint8          // algorithm
	sign byteutils.Hash // Signature values
}

// From return from address
func (tx *Transaction) From() *Address {
	return tx.from
}

// Timestamp return timestamp
func (tx *Transaction) Timestamp() int64 {
	return tx.timestamp
}

// To return to address
func (tx *Transaction) To() *Address {
	return tx.to
}

// ChainID return chainID
func (tx *Transaction) ChainID() uint32 {
	return tx.chainID
}

// Value return tx value
func (tx *Transaction) Value() *util.Uint128 {
	return tx.value
}

// Nonce return tx nonce
func (tx *Transaction) Nonce() uint64 {
	return tx.nonce
}

// Type return tx type
func (tx *Transaction) Type() string {
	return tx.data.Type
}

// Data return tx data
func (tx *Transaction) Data() []byte {
	return tx.data.Payload
}

// ToProto converts domain Tx to proto Tx
func (tx *Transaction) ToProto() (proto.Message, error) {
	value, err := tx.value.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	gasPrice, err := tx.gasPrice.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	gasLimit, err := tx.gasLimit.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	return &corepb.Transaction{
		Hash:      tx.hash,
		From:      tx.from.address,
		To:        tx.to.address,
		Value:     value,
		Nonce:     tx.nonce,
		Timestamp: tx.timestamp,
		Data:      tx.data,
		ChainId:   tx.chainID,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Alg:       uint32(tx.alg),
		Sign:      tx.sign,
	}, nil
}

// FromProto converts proto Tx into domain Tx
func (tx *Transaction) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.Transaction); ok {
		tx.hash = msg.Hash
		tx.from = &Address{msg.From}
		tx.to = &Address{msg.To}
		value, err := util.NewUint128FromFixedSizeByteSlice(msg.Value)
		if err != nil {
			return err
		}
		tx.value = value
		tx.nonce = msg.Nonce
		tx.timestamp = msg.Timestamp
		tx.data = msg.Data
		tx.chainID = msg.ChainId
		gasPrice, err := util.NewUint128FromFixedSizeByteSlice(msg.GasPrice)
		if err != nil {
			return err
		}
		tx.gasPrice = gasPrice
		gasLimit, err := util.NewUint128FromFixedSizeByteSlice(msg.GasLimit)
		if err != nil {
			return err
		}
		tx.gasLimit = gasLimit
		tx.alg = uint8(msg.Alg)
		tx.sign = msg.Sign
		return nil
	}
	return errors.New("Protobug Message cannot be converted into Transaction")
}

func (tx *Transaction) String() string {
	return fmt.Sprintf(`{"chainID":%d, "hash":"%s", "from":"%s", "to":"%s", "nonce":%d, "value":"%s", "timestamp":%d, "gasprice": "%s", "gaslimit":"%s", "type":"%s"}`,
		tx.chainID,
		tx.hash.String(),
		tx.from.String(),
		tx.to.String(),
		tx.nonce,
		tx.value.String(),
		tx.timestamp,
		tx.gasPrice.String(),
		tx.gasLimit.String(),
		tx.Type(),
	)
}

// Transactions is an alias of Transaction array.
type Transactions []*Transaction

// NewTransaction create #Transaction instance.
func NewTransaction(chainID uint32, from, to *Address, value *util.Uint128, nonce uint64, payloadType string, payload []byte, gasPrice *util.Uint128, gasLimit *util.Uint128) *Transaction {
	//if gasPrice is not specified, use the default gasPrice
	if gasPrice == nil || gasPrice.Cmp(util.NewUint128()) <= 0 {
		gasPrice = TransactionGasPrice
	}
	if gasLimit == nil || gasLimit.Cmp(util.NewUint128()) <= 0 {
		gasLimit = MinGasCountPerTransaction
	}

	tx := &Transaction{
		from:      from,
		to:        to,
		value:     value,
		nonce:     nonce,
		timestamp: time.Now().Unix(),
		chainID:   chainID,
		data:      &corepb.Data{Type: payloadType, Payload: payload},
		gasPrice:  gasPrice,
		gasLimit:  gasLimit,
	}
	return tx
}

// Hash return the hash of transaction.
func (tx *Transaction) Hash() byteutils.Hash {
	return tx.hash
}

// GasPrice returns gasPrice
func (tx *Transaction) GasPrice() *util.Uint128 {
	return tx.gasPrice
}

// GasLimit returns gasLimit
func (tx *Transaction) GasLimit() *util.Uint128 {
	return tx.gasLimit
}

// PayloadGasLimit returns payload gasLimit
func (tx *Transaction) PayloadGasLimit(payload TxPayload) (*util.Uint128, error) {
	// payloadGasLimit = tx.gasLimit - tx.GasCountOfTxBase
	gasCountOfTxBase, err := tx.GasCountOfTxBase()
	if err != nil {
		return nil, err
	}
	payloadGasLimit, err := tx.gasLimit.Sub(gasCountOfTxBase)
	if err != nil {
		return nil, ErrOutOfGasLimit
	}
	payloadGasLimit, err = payloadGasLimit.Sub(payload.BaseGasCount())
	if err != nil {
		return nil, ErrOutOfGasLimit
	}
	return payloadGasLimit, nil
}

// MinBalanceRequired returns gasprice * gaslimit.
func (tx *Transaction) MinBalanceRequired() (*util.Uint128, error) {
	total, err := tx.GasPrice().Mul(tx.GasLimit())
	if err != nil {
		return nil, err
	}
	total, err = total.Add(tx.value)
	if err != nil {
		return nil, err
	}
	return total, nil
}

// GasCountOfTxBase calculate the actual amount for a tx with data
func (tx *Transaction) GasCountOfTxBase() (*util.Uint128, error) {
	txGas := MinGasCountPerTransaction.DeepCopy()
	if tx.DataLen() > 0 {
		dataLen, err := util.NewUint128FromInt(int64(tx.DataLen()))
		if err != nil {
			return nil, err
		}
		dataGas, err := dataLen.Mul(GasCountPerByte)
		if err != nil {
			return nil, err
		}
		txGas, err = txGas.Add(dataGas)
		if err != nil {
			return nil, err
		}
	}
	return txGas, nil
}

// DataLen return the length of payload
func (tx *Transaction) DataLen() int {
	return len(tx.data.Payload)
}

// LoadPayload returns tx's payload
func (tx *Transaction) LoadPayload() (TxPayload, error) {
	// execute payload
	var (
		payload TxPayload
		err     error
	)
	switch tx.data.Type {
	case TxPayloadBinaryType:
		payload, err = LoadBinaryPayload(tx.data.Payload)
	case TxPayloadDeployType:
		payload, err = LoadDeployPayload(tx.data.Payload)
	case TxPayloadCallType:
		payload, err = LoadCallPayload(tx.data.Payload)
	default:
		err = ErrInvalidTxPayloadType
	}
	return payload, err
}

// LocalExecution returns tx local execution
func (tx *Transaction) LocalExecution(block *Block) (*util.Uint128, string, error) {
	hash, err := HashTransaction(tx)
	if err != nil {
		return nil, "", err
	}
	tx.hash = hash

	txWorldState, err := block.Prepare(tx)
	if err != nil {
		return nil, "", err
	}

	payload, err := tx.LoadPayload()
	if err != nil {
		return util.NewUint128(), "", err
	}

	gasUsed, err := tx.GasCountOfTxBase()
	if err != nil {
		return util.NewUint128(), "", err
	}

	gasUsed, err = gasUsed.Add(payload.BaseGasCount())
	if err != nil {
		return util.NewUint128(), "", err
	}

	gasExecution, result, exeErr := payload.Execute(tx, block, txWorldState)
	gas, err := gasUsed.Add(gasExecution)
	if err != nil {
		return gasUsed, result, err
	}

	if err := block.Reset(tx); err != nil {
		return util.NewUint128(), "", err
	}

	return gas, result, exeErr
}

// VerifyExecution transaction and return result.
func VerifyExecution(tx *Transaction, block *Block, txWorldState state.TxWorldState) error {
	coinbase := block.CoinbaseHash()

	if err := tx.checkBalance(block, txWorldState); err != nil {
		return ErrInsufficientBalance
	}

	// gasLimit < gasUsed
	gasUsed, err := tx.GasCountOfTxBase()
	if err != nil {
		return err
	}
	if tx.gasLimit.Cmp(gasUsed) < 0 {
		logging.VLog().WithFields(logrus.Fields{
			"error":       ErrOutOfGasLimit,
			"transaction": tx,
			"limit":       tx.gasLimit.String(),
			"used":        gasUsed.String(),
		}).Debug("Failed to store the payload on chain.")
		return ErrOutOfGasLimit
	}

	payload, err := tx.LoadPayload()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"error":       err,
			"block":       block,
			"transaction": tx,
		}).Debug("Failed to load payload.")
		metricsTxExeFailed.Mark(1)

		if err := tx.recordGas(coinbase, gasUsed, block); err != nil {
			return err
		}
		tx.triggerEvent(txWorldState, block, TopicExecuteTxFailed, gasUsed, err)
		return nil
	}

	gasUsed, err = gasUsed.Add(payload.BaseGasCount())
	if err != nil {
		return err
	}
	if tx.gasLimit.Cmp(gasUsed) < 0 {
		logging.VLog().WithFields(logrus.Fields{
			"err":   ErrOutOfGasLimit,
			"block": block,
			"tx":    tx,
		}).Debug("Failed to check base gas used.")
		metricsTxExeFailed.Mark(1)

		if err := tx.recordGas(coinbase, tx.gasLimit, block); err != nil {
			return err
		}
		tx.triggerEvent(txWorldState, block, TopicExecuteTxFailed, tx.gasLimit, ErrOutOfGasLimit)
		return nil
	}

	// execute smart contract and sub the calcute gas.
	gasExecution, _, exeErr := payload.Execute(tx, block, txWorldState)
	if exeErr != nil {
		if err := block.Reset(tx); err != nil {
			return err
		}
	}

	allGas, gasErr := gasUsed.Add(gasExecution)
	if gasErr != nil {
		return gasErr
	}

	if tx.gasLimit.Cmp(allGas) < 0 {
		logging.VLog().WithFields(logrus.Fields{
			"err":   ErrOutOfGasLimit,
			"block": block,
			"tx":    tx,
		}).Debug("Failed to check gas executed.")
		metricsTxExeFailed.Mark(1)

		if err := block.Reset(tx); err != nil {
			return err
		}
		if err := tx.recordGas(coinbase, tx.gasLimit, block); err != nil {
			return err
		}
		tx.triggerEvent(txWorldState, block, TopicExecuteTxFailed, tx.gasLimit, ErrOutOfGasLimit)
		return nil
	}
	tx.recordGas(coinbase, allGas, block)

	if exeErr != nil {
		logging.VLog().WithFields(logrus.Fields{
			"exeErr":       exeErr,
			"block":        block,
			"tx":           tx,
			"gasUsed":      gasUsed.String(),
			"gasExecution": gasExecution.String(),
		}).Debug("Failed to execute payload.")

		metricsTxExeFailed.Mark(1)
		tx.triggerEvent(txWorldState, block, TopicExecuteTxFailed, allGas, err)
	} else {
		// accept the transaction
		if err := transfer(tx.from.address, tx.to.address, tx.value, txWorldState); err != nil {
			return err
		}

		metricsTxExeSuccess.Mark(1)
		tx.triggerEvent(txWorldState, block, TopicExecuteTxSuccess, allGas, nil)
	}

	return nil
}

func (tx *Transaction) checkBalance(block *Block, txWorldState state.TxWorldState) error {
	fromAcc, err := txWorldState.GetOrCreateUserAccount(tx.from.address)
	if err != nil {
		return err
	}
	minBalanceRequired, err := tx.MinBalanceRequired()
	if err != nil {
		return err
	}
	if fromAcc.Balance().Cmp(minBalanceRequired) < 0 {
		return ErrInsufficientBalance
	}
	return nil
}

func (tx *Transaction) recordGas(coinbase byteutils.Hash, gasCnt *util.Uint128, block *Block) error {
	gasCost, err := tx.GasPrice().Mul(gasCnt)
	if err != nil {
		return err
	}
	return block.recordGas(tx.from.String(), gasCost)
}

func transfer(from, to byteutils.Hash, value *util.Uint128, txWorldState state.TxWorldState) error {
	fromAcc, err := txWorldState.GetOrCreateUserAccount(from)
	if err != nil {
		return err
	}
	toAcc, err := txWorldState.GetOrCreateUserAccount(to)
	if err != nil {
		return err
	}
	fromAcc.SubBalance(value)
	toAcc.AddBalance(value)
	return nil
}

func (tx *Transaction) triggerEvent(txWorldState state.TxWorldState, block *Block, topic string, gasUsed *util.Uint128, err error) {
	txEvent := &TransactionEvent{
		Hash:    tx.hash.String(),
		GasUsed: gasUsed.String(),
	}
	if err != nil {
		txEvent.Status = TxExecutionFailed
		txEvent.Error = err.Error()
	} else {
		txEvent.Status = TxExecutionSuccess
	}

	txData, _ := json.Marshal(txEvent)
	//logging.VLog().WithFields(logrus.Fields{
	//	"topic": TopicTransactionExecutionResult,
	//	"event": string(txData),
	//}).Debug("record event.")
	event := &state.Event{
		Topic: TopicTransactionExecutionResult,
		Data:  string(txData),
	}
	txWorldState.RecordEvent(tx.hash, event)
}

// Sign sign transaction,sign algorithm is
func (tx *Transaction) Sign(signature keystore.Signature) error {
	hash, err := HashTransaction(tx)
	if err != nil {
		return err
	}
	sign, err := signature.Sign(hash)
	if err != nil {
		return err
	}
	tx.hash = hash
	tx.alg = uint8(signature.Algorithm())
	tx.sign = sign
	return nil
}

// VerifyIntegrity return transaction verify result, including Hash and Signature.
func (tx *Transaction) VerifyIntegrity(chainID uint32) error {
	// check ChainID.
	if tx.chainID != chainID {
		return ErrInvalidChainID
	}

	// check Hash.
	wantedHash, err := HashTransaction(tx)
	if err != nil {
		return err
	}
	if wantedHash.Equals(tx.hash) == false {
		return ErrInvalidTransactionHash
	}

	// check Signature.
	if err := tx.verifySign(); err != nil {
		return err
	}

	return nil
}

func (tx *Transaction) verifySign() error {
	signature, err := crypto.NewSignature(keystore.Algorithm(tx.alg))
	if err != nil {
		return err
	}
	pub, err := signature.RecoverPublic(tx.hash, tx.sign)
	if err != nil {
		return err
	}
	pubdata, err := pub.Encoded()
	if err != nil {
		return err
	}
	addr, err := NewAddressFromPublicKey(pubdata)
	if err != nil {
		return err
	}
	if !tx.from.Equals(addr) {
		logging.VLog().WithFields(logrus.Fields{
			"recover address": addr.String(),
			"tx":              tx,
		}).Debug("Failed to verify tx's sign.")
		return ErrInvalidTransactionSigner
	}
	return nil
}

// GenerateContractAddress according to tx.from and tx.nonce.
func (tx *Transaction) GenerateContractAddress() (*Address, error) {
	return NewContractAddressFromHash(hash.Sha3256(tx.from.Bytes(), byteutils.FromUint64(tx.nonce)))
}

// CheckContract check if contract is valid
func CheckContract(addr *Address, txWorldState state.TxWorldState) error {
	contract, err := txWorldState.GetContractAccount(addr.Bytes())
	if err != nil {
		return err
	}

	if len(contract.BirthPlace()) == 0 {
		return state.ErrAccountNotFound
	}

	birthEvents, err := txWorldState.FetchEvents(contract.BirthPlace())
	if err != nil {
		return err
	}

	result := false
	for _, v := range birthEvents {

		if v.Topic == TopicTransactionExecutionResult {
			txEvent := TransactionEvent{}
			json.Unmarshal([]byte(v.Data), &txEvent)
			if txEvent.Status == TxExecutionSuccess {
				result = true
				break
			}
		} else if v.Topic == TopicExecuteTxSuccess {
			result = true
			break
		}
	}
	if !result {
		return state.ErrAccountNotFound
	}

	return nil
}

// CheckTransaction in a tx world state
func CheckTransaction(tx *Transaction, txWorldState state.TxWorldState) (bool, error) {
	// check nonce
	fromAcc, err := txWorldState.GetOrCreateUserAccount(tx.from.address)
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

// AcceptTransaction in a tx world state
func AcceptTransaction(tx *Transaction, txWorldState state.TxWorldState) error {
	// record tx
	pbTx, err := tx.ToProto()
	if err != nil {
		return err
	}
	txBytes, err := proto.Marshal(pbTx)
	if err != nil {
		return err
	}
	if err := txWorldState.PutTx(tx.hash, txBytes); err != nil {
		return err
	}
	// incre nonce
	fromAcc, err := txWorldState.GetOrCreateUserAccount(tx.from.address)
	if err != nil {
		return err
	}
	fromAcc.IncrNonce()
	return nil
}

// GetTransaction from txs Trie
func GetTransaction(hash byteutils.Hash, txWorldState state.TxWorldState) (*Transaction, error) {
	bytes, err := txWorldState.GetTx(hash)
	if err != nil {
		return nil, err
	}
	pbTx := new(corepb.Transaction)
	if err := proto.Unmarshal(bytes, pbTx); err != nil {
		return nil, err
	}
	tx := new(Transaction)
	if err = tx.FromProto(pbTx); err != nil {
		return nil, err
	}
	return tx, nil
}

// HashTransaction hash the transaction.
func HashTransaction(tx *Transaction) (byteutils.Hash, error) {
	value, err := tx.value.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	data, err := proto.Marshal(tx.data)
	if err != nil {
		return nil, err
	}
	gasPrice, err := tx.gasPrice.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	gasLimit, err := tx.gasLimit.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	return hash.Sha3256(
		tx.from.address,
		tx.to.address,
		value,
		byteutils.FromUint64(tx.nonce),
		byteutils.FromInt64(tx.timestamp),
		data,
		byteutils.FromUint32(tx.chainID),
		gasPrice,
		gasLimit,
	), nil
}
