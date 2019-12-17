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

	"github.com/nebulasio/go-nebulas/crypto/sha3"

	"github.com/nebulasio/go-nebulas/core/state"

	"encoding/json"

	"github.com/gogo/protobuf/proto"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

const (
	// TxHashByteLength invalid tx hash length(len of []byte)
	TxHashByteLength = 32
)

var (
	// TransactionMaxGasPrice max gasPrice:1 * 10 ** 12
	TransactionMaxGasPrice, _ = util.NewUint128FromString("1000000000000")

	// TransactionMaxGas max gas:50 * 10 ** 9
	TransactionMaxGas, _ = util.NewUint128FromString("50000000000")

	// TransactionGasPrice default gasPrice : 2*10**10
	TransactionGasPrice, _ = util.NewUint128FromString("20000000000")

	// GenesisGasPrice default gasPrice : 1*10**6
	GenesisGasPrice, _ = util.NewUint128FromInt(1000000)

	// MinGasCountPerTransaction default gas for normal transaction
	MinGasCountPerTransaction, _ = util.NewUint128FromInt(20000)

	// GasCountPerByte per byte of data attached to a transaction gas cost
	GasCountPerByte, _ = util.NewUint128FromInt(1)

	// MaxDataPayLoadLength Max data length in transaction
	MaxDataPayLoadLength = 128 * 1024
	// MaxDataBinPayloadLength Max data length in binary transaction
	MaxDataBinPayloadLength = 64

	// MaxEventErrLength Max error length in event
	MaxEventErrLength = 256

	// MaxResultLength max execution result length
	MaxResultLength = 256
)

// TransactionEvent transaction event
type TransactionEvent struct {
	Hash    string `json:"hash"`
	Status  int8   `json:"status"`
	GasUsed string `json:"gas_used"`
	Error   string `json:"error"`
}

// TransactionEventV2 add execution result
type TransactionEventV2 struct {
	Hash          string `json:"hash"`
	Status        int8   `json:"status"`
	GasUsed       string `json:"gas_used"`
	Error         string `json:"error"`
	ExecuteResult string `json:"execute_result"`
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
	alg  keystore.Algorithm
	sign byteutils.Hash // Signature values
}

// SetTimestamp update the timestamp.
func (tx *Transaction) SetTimestamp(timestamp int64) {
	tx.timestamp = timestamp
}

// SetSignature update tx sign
func (tx *Transaction) SetSignature(alg keystore.Algorithm, sign byteutils.Hash) {
	tx.alg = alg
	tx.sign = sign
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

// SetNonce update th nonce
func (tx *Transaction) SetNonce(newNonce uint64) {
	tx.nonce = newNonce
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
		if msg != nil {

			tx.hash = msg.Hash
			from, err := AddressParseFromBytes(msg.From)
			if err != nil {
				return err
			}
			tx.from = from

			to, err := AddressParseFromBytes(msg.To)
			if err != nil {
				return err
			}
			tx.to = to

			value, err := util.NewUint128FromFixedSizeByteSlice(msg.Value)
			if err != nil {
				return err
			}
			tx.value = value

			tx.nonce = msg.Nonce
			tx.timestamp = msg.Timestamp
			tx.chainID = msg.ChainId

			if msg.Data == nil {
				return ErrInvalidTransactionData
			}
			if len(msg.Data.Payload) > MaxDataPayLoadLength {
				return ErrTxDataPayLoadOutOfMaxLength
			}
			if CheckGenesisTransaction(tx) == false &&
				msg.Data.Type == TxPayloadBinaryType &&
				len(msg.Data.Payload) > MaxDataBinPayloadLength {
				return ErrTxDataBinPayLoadOutOfMaxLength
			}
			tx.data = msg.Data

			gasPrice, err := util.NewUint128FromFixedSizeByteSlice(msg.GasPrice)
			if err != nil {
				return err
			}
			if gasPrice.Cmp(util.Uint128Zero()) <= 0 || gasPrice.Cmp(TransactionMaxGasPrice) > 0 {
				return ErrInvalidGasPrice
			}
			tx.gasPrice = gasPrice

			gasLimit, err := util.NewUint128FromFixedSizeByteSlice(msg.GasLimit)
			if err != nil {
				return err
			}
			if gasLimit.Cmp(util.Uint128Zero()) <= 0 || gasLimit.Cmp(TransactionMaxGas) > 0 {
				return ErrInvalidGasLimit
			}
			tx.gasLimit = gasLimit

			alg := keystore.Algorithm(msg.Alg)
			if err := crypto.CheckAlgorithm(alg); err != nil {
				return err
			}

			tx.alg = alg
			tx.sign = msg.Sign
			return nil
		}
		return ErrInvalidProtoToTransaction
	}
	return ErrInvalidProtoToTransaction
}

func (tx *Transaction) String() string {
	return fmt.Sprintf(`{"chainID":%d, "hash":"%s", "from":"%s", "to":"%s", "nonce":%d, "value":"%s", "timestamp":%d, "gasprice": "%s", "gaslimit":"%s", "data": "%s", "type":"%s"}`,
		tx.chainID,
		tx.hash.String(),
		tx.from.String(),
		tx.to.String(),
		tx.nonce,
		tx.value.String(),
		tx.timestamp,
		tx.gasPrice.String(),
		tx.gasLimit.String(),
		tx.Data(),
		tx.Type(),
	)
}

func (tx *Transaction) StringWithoutData() string {
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

// JSONString of transaction
func (tx *Transaction) JSONString() string {
	txJSONObj := make(map[string]interface{})
	txJSONObj["chainID"] = tx.chainID
	txJSONObj["hash"] = tx.hash.String()
	txJSONObj["from"] = tx.from.String()
	txJSONObj["to"] = tx.to.String()
	txJSONObj["nonce"] = tx.nonce
	txJSONObj["value"] = tx.value.String()
	txJSONObj["timestamp"] = tx.timestamp
	txJSONObj["gasprice"] = tx.gasPrice.String()
	txJSONObj["gaslimit"] = tx.gasLimit.String()
	txJSONObj["data"] = string(tx.Data())
	txJSONObj["type"] = tx.Type()
	txJSON, err := json.Marshal(txJSONObj)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"tx":  tx,
		}).Error("Failed to get transaction json string")
	}
	return string(txJSON)
}

// Transactions is an alias of Transaction array.
type Transactions []*Transaction

// NewTransaction create #Transaction instance.
func NewTransaction(chainID uint32, from, to *Address, value *util.Uint128, nonce uint64, payloadType string, payload []byte, gasPrice *util.Uint128, gasLimit *util.Uint128) (*Transaction, error) {
	if gasPrice == nil || gasPrice.Cmp(util.NewUint128()) <= 0 || gasPrice.Cmp(TransactionMaxGasPrice) > 0 {
		return nil, ErrInvalidGasPrice
	}
	if gasLimit == nil || gasLimit.Cmp(util.NewUint128()) <= 0 || gasLimit.Cmp(TransactionMaxGas) > 0 {
		return nil, ErrInvalidGasLimit
	}

	if nil == from || nil == to || nil == value {
		return nil, ErrInvalidArgument
	}

	if len(payload) > MaxDataPayLoadLength {
		return nil, ErrTxDataPayLoadOutOfMaxLength
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
	return tx, nil
}

// NewChildTransaction return child tx to inner nvm
func (tx *Transaction) NewInnerTransaction(from, to *Address, value *util.Uint128, payloadType string, payload []byte) (*Transaction, error) {
	innerTx, err := NewTransaction(tx.chainID, from, to, value, InnerTransactionNonce, payloadType, payload, tx.GasPrice(), tx.GasLimit())
	if err != nil {
		return nil, ErrCreateInnerTx
	}
	innerTx.SetHash(tx.hash)
	return innerTx, nil
}

// Hash return the hash of transaction.
func (tx *Transaction) Hash() byteutils.Hash {
	return tx.hash
}

// SetHash set hash to in args
func (tx *Transaction) SetHash(in byteutils.Hash) {
	tx.hash = in
}

// GasPrice returns gasPrice
func (tx *Transaction) GasPrice() *util.Uint128 {
	return tx.gasPrice
}

// GasLimit returns gasLimit
func (tx *Transaction) GasLimit() *util.Uint128 {
	return tx.gasLimit
}

// GasCountOfTxBase calculate the actual amount for a tx with data
func (tx *Transaction) GasCountOfTxBase() (*util.Uint128, error) {
	txGas := MinGasCountPerTransaction
	if tx.DataLen() > 0 {
		dataLen, err := util.NewUint128FromInt(int64(tx.DataLen()))
		if err != nil {
			return nil, err
		}
		dataGas, err := dataLen.Mul(GasCountPerByte)
		if err != nil {
			return nil, err
		}
		baseGas, err := txGas.Add(dataGas)
		if err != nil {
			return nil, err
		}
		txGas = baseGas
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
	case TxPayloadProtocolType:
		payload, err = LoadProtocolPayload(tx.data.Payload)
	case TxPayloadDipType:
		payload, err = LoadDipPayload(tx.data.Payload)
	case TxPayloadPodType:
		payload, err = LoadPodPayload(tx.data.Payload)
	default:
		err = ErrInvalidTxPayloadType
	}
	return payload, err
}

func submitTx(tx *Transaction, block *Block, ws WorldState,
	gas *util.Uint128, exeErr error, exeErrTy string, exeResult string) (bool, error) {
	if exeErr != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":         exeErr,
			"block":       block,
			"transaction": tx,
		}).Info(exeErrTy)
		metricsTxExeFailed.Mark(1)
	} else {
		metricsTxExeSuccess.Mark(1)
	}

	if exeErr != nil {
		// if execution failed, the previous changes on world state should be reset
		// record dependency

		if WsResetRecordDependencyAtHeight2(block.Height()) {
			if err := ws.Reset(nil, false); err != nil {
				// if reset failed, the tx should be given back
				return true, err
			}
		} else {
			addr := tx.to.address
			if !WsResetRecordDependencyAtHeight(block.Height()) {
				addr = tx.from.address
			}
			if err := ws.Reset(addr, true); err != nil {
				// if reset failed, the tx should be given back
				return true, err
			}
		}
	}

	if err := tx.recordGas(gas, ws); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":   err,
			"tx":    tx,
			"gas":   gas,
			"block": block,
		}).Error("Failed to record gas, unexpected error")
		metricsUnexpectedBehavior.Update(1)
		return true, err
	}

	if err := tx.recordResultEvent(gas, exeErr, ws, block, exeResult); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":   err,
			"tx":    tx,
			"gas":   gas,
			"block": block,
		}).Error("Failed to record result event, unexpected error")
		metricsUnexpectedBehavior.Update(1)
		return true, err
	}
	// No error, won't giveback the tx
	return false, nil
}

// VerifyExecution transaction and return result.
func VerifyExecution(tx *Transaction, block *Block, ws WorldState) (bool, error) {
	// step0. perpare accounts.
	fromAcc, err := ws.GetOrCreateUserAccount(tx.from.address)
	if err != nil {
		return true, err
	}
	toAcc, err := ws.GetOrCreateUserAccount(tx.to.address)
	if err != nil {
		return true, err
	}

	// step1. check balance >= gasLimit * gasPrice
	limitedFee, err := tx.gasLimit.Mul(tx.gasPrice)
	if err != nil {
		// Gas overflow, won't giveback the tx
		return false, ErrGasFeeOverflow
	}

	if tx.Type() != TxPayloadPodType {
		if fromAcc.Balance().Cmp(limitedFee) < 0 {
			// Balance is smaller than limitedFee, won't giveback the tx
			return false, ErrInsufficientBalance
		}
	}

	// step2. check gasLimit >= txBaseGas.
	baseGas, err := tx.GasCountOfTxBase()
	if err != nil {
		// Gas overflow, won't giveback the tx
		return false, ErrGasCntOverflow
	}
	gasUsed := baseGas
	if tx.gasLimit.Cmp(gasUsed) < 0 {
		logging.VLog().WithFields(logrus.Fields{
			"error":       ErrOutOfGasLimit,
			"transaction": tx,
			"limit":       tx.gasLimit,
			"acceptedGas": gasUsed,
		}).Debug("Failed to check gasLimit >= txBaseGas.")
		// GasLimit is smaller than based tx gas, won't giveback the tx
		return false, ErrOutOfGasLimit
	}

	// !!!!!!Attention: all txs passed here will be on chain.

	// step3. check payload vaild.
	payload, payloadErr := tx.LoadPayload()
	if payloadErr != nil {
		return submitTx(tx, block, ws, gasUsed, payloadErr, "Failed to load payload.", "")
	}

	// step4. calculate base gas of payload
	payloadGas, err := gasUsed.Add(payload.BaseGasCount())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":            err,
			"tx":             tx,
			"gasUsed":        gasUsed,
			"payloadBaseGas": payload.BaseGasCount(),
			"block":          block,
		}).Error("Failed to add payload base gas, unexpected error")
		metricsUnexpectedBehavior.Update(1)
		return submitTx(tx, block, ws, gasUsed, ErrGasCntOverflow, "Failed to add the count of base payload gas", "")
	}
	gasUsed = payloadGas
	if tx.gasLimit.Cmp(gasUsed) < 0 {
		return submitTx(tx, block, ws, tx.gasLimit, ErrOutOfGasLimit, "Failed to check gasLimit >= txBaseGas + payloasBaseGas.", "")
	}

	// step5. check balance >= limitedFee + value. and transfer
	minBalanceRequired, balanceErr := limitedFee.Add(tx.value)
	if balanceErr != nil {
		return submitTx(tx, block, ws, gasUsed, ErrGasFeeOverflow, "Failed to add tx.value", "")
	}
	if tx.Type() != TxPayloadPodType {
		if fromAcc.Balance().Cmp(minBalanceRequired) < 0 {
			return submitTx(tx, block, ws, gasUsed, ErrInsufficientBalance, "Failed to check balance >= gasLimit * gasPrice + value", "")
		}
	}
	var transferSubErr, transferAddErr error
	transferSubErr = fromAcc.SubBalance(tx.value)
	if transferSubErr == nil {
		transferAddErr = toAcc.AddBalance(tx.value)
	}
	if transferSubErr != nil || transferAddErr != nil {
		logging.VLog().WithFields(logrus.Fields{
			"subErr":      transferSubErr,
			"addErr":      transferAddErr,
			"tx":          tx,
			"fromBalance": fromAcc.Balance(),
			"toBalance":   toAcc.Balance(),
			"block":       block,
		}).Error("Failed to transfer value, unexpected error")
		metricsUnexpectedBehavior.Update(1)
		return submitTx(tx, block, ws, gasUsed, ErrInvalidTransfer, "Failed to transfer tx.value", "")
	}

	// step6. calculate contract's limited gas
	contractLimitedGas, err := tx.gasLimit.Sub(gasUsed)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":     err,
			"tx":      tx,
			"gasUsed": gasUsed,
			"block":   block,
		}).Error("Failed to calculate payload's limit gas, unexpected error")
		metricsUnexpectedBehavior.Update(1)
		return submitTx(tx, block, ws, tx.gasLimit, ErrOutOfGasLimit, "Failed to calculate payload's limit gas", "")
	}

	// step7. execute contract.
	gasExecution, exeResult, exeErr := payload.Execute(contractLimitedGas, tx, block, ws)
	if exeErr == ErrUnexpected {
		return false, exeErr
	}

	// step8. calculate final gas.
	allGas, gasErr := gasUsed.Add(gasExecution)
	if gasErr != nil {
		return submitTx(tx, block, ws, gasUsed, ErrGasCntOverflow, "Failed to add the fee of execution gas", "")
	}
	if tx.gasLimit.Cmp(allGas) < 0 {
		return submitTx(tx, block, ws, tx.gasLimit, ErrOutOfGasLimit, "Failed to check gasLimit >= allGas", "")
	}

	// step9. over
	return submitTx(tx, block, ws, allGas, exeErr, "Failed to execute payload", exeResult)
}

// simulateExecution simulate execution and return gasUsed, executionResult and executionErr, sysErr if occurred.
func (tx *Transaction) simulateExecution(block *Block) (*SimulateResult, error) {
	// hash is necessary in nvm
	hash, err := tx.HashTransaction()
	if err != nil {
		return nil, err
	}
	tx.hash = hash

	// check dip reward
	if err := block.dip.CheckReward(tx); err != nil {
		return nil, err
	}

	// Generate world state
	ws := block.WorldState()

	// Get from account
	fromAcc, err := ws.GetOrCreateUserAccount(tx.from.address)
	if err != nil {
		return nil, err
	}

	// calculate min gas.
	gasUsed, err := tx.GasCountOfTxBase()
	if err != nil {
		return &SimulateResult{util.NewUint128(), "GasCountOfTxBase error", err}, nil
	}

	payload, err := tx.LoadPayload()
	if err != nil {
		return &SimulateResult{gasUsed, "Invalid payload", err}, nil
	}

	payloasGas, err := gasUsed.Add(payload.BaseGasCount())
	if err != nil {
		return &SimulateResult{gasUsed, "GasCountOfTxBase + GasCountOfPayloadBase error", err}, nil
	}
	gasUsed = payloasGas

	var (
		result string
		exeErr error
	)

	// try run smart contract if payload is.
	if tx.data.Type == TxPayloadCallType || tx.data.Type == TxPayloadDeployType ||
		(tx.data.Type == TxPayloadBinaryType && tx.to.Type() == ContractAddress && AcceptAvailableAtHeight(block.height)) {

		// transfer value to smart contract.
		toAcc, err := ws.GetOrCreateUserAccount(tx.to.address)
		if err != nil {
			return nil, err
		}
		err = toAcc.AddBalance(tx.value)
		if err != nil {
			return &SimulateResult{gasUsed, "Too big value", err}, nil
		}

		// execute.
		gasExecution := util.NewUint128()
		gasExecution, result, exeErr = payload.Execute(TransactionMaxGas, tx, block, ws)

		// add gas.
		executedGas, err := gasUsed.Add(gasExecution)
		if err != nil {
			return &SimulateResult{gasUsed, "CalFinalGasCount error", err}, nil
		}
		gasUsed = executedGas

		if exeErr != nil {
			return &SimulateResult{gasUsed, result, exeErr}, nil
		}
	}

	if tx.Type() != TxPayloadPodType {
		// check balance.
		err = checkBalanceForGasUsedAndValue(ws, fromAcc, tx.value, gasUsed, tx.gasPrice)
	}
	return &SimulateResult{gasUsed, result, err}, nil
}

// checkBalanceForGasUsedAndValue check balance >= gasUsed * gasPrice + value.
func checkBalanceForGasUsedAndValue(ws WorldState, fromAcc state.Account, value, gasUsed, gasPrice *util.Uint128) error {
	gasFee, err := gasPrice.Mul(gasUsed)
	if err != nil {
		return err
	}
	balanceRequired, err := gasFee.Add(value)
	if err != nil {
		return err
	}
	if fromAcc.Balance().Cmp(balanceRequired) < 0 {
		return ErrInsufficientBalance
	}
	return nil

}

func (tx *Transaction) recordGas(gasCnt *util.Uint128, ws WorldState) error {
	gasCost, err := tx.GasPrice().Mul(gasCnt)
	if err != nil {
		return err
	}

	// There is no gas fee for the consensus transaction
	if tx.Type() == TxPayloadPodType {
		gasCost = util.NewUint128()
	}

	return ws.RecordGas(tx.from.String(), gasCost)
}

func (tx *Transaction) recordResultEvent(gasUsed *util.Uint128, err error, ws WorldState, block *Block, exeResult string) error {

	var txData []byte
	if RecordCallContractResultAtHeight(block.height) {

		if len(exeResult) > MaxResultLength {
			exeResult = exeResult[:MaxResultLength]
		}
		txEvent := &TransactionEventV2{
			Hash:          tx.hash.String(),
			GasUsed:       gasUsed.String(),
			Status:        TxExecutionSuccess,
			ExecuteResult: exeResult,
		}

		if err != nil {
			txEvent.Status = TxExecutionFailed
			txEvent.Error = err.Error()
			if len(txEvent.Error) > MaxEventErrLength {
				txEvent.Error = txEvent.Error[:MaxEventErrLength]
			}
		}
		txData, err = json.Marshal(txEvent)
	} else {
		txEvent := &TransactionEvent{
			Hash:    tx.hash.String(),
			GasUsed: gasUsed.String(),
			Status:  TxExecutionSuccess,
		}

		if err != nil {
			txEvent.Status = TxExecutionFailed
			txEvent.Error = err.Error()
			if len(txEvent.Error) > MaxEventErrLength {
				txEvent.Error = txEvent.Error[:MaxEventErrLength]
			}
		}
		txData, err = json.Marshal(txEvent)
	}

	if err != nil {
		return err
	}

	event := &state.Event{
		Topic: TopicTransactionExecutionResult,
		Data:  string(txData),
	}
	ws.RecordEvent(tx.hash, event)
	return nil
}

// Sign sign transaction,sign algorithm is
func (tx *Transaction) Sign(signature keystore.Signature) error {
	if signature == nil {
		return ErrNilArgument
	}
	hash, err := tx.HashTransaction()
	if err != nil {
		return err
	}
	sign, err := signature.Sign(hash)
	if err != nil {
		return err
	}
	tx.hash = hash
	tx.alg = signature.Algorithm()
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
	wantedHash, err := tx.HashTransaction()
	if err != nil {
		return err
	}
	if wantedHash.Equals(tx.hash) == false {
		return ErrInvalidTransactionHash
	}

	// check Signature.
	return tx.verifySign()

}

func (tx *Transaction) verifySign() error {
	signer, err := RecoverSignerFromSignature(tx.alg, tx.hash, tx.sign)
	if err != nil {
		return err
	}
	if !tx.from.Equals(signer) {
		logging.VLog().WithFields(logrus.Fields{
			"signer":  signer.String(),
			"tx.from": tx.from,
		}).Debug("Failed to verify tx's sign.")
		return ErrInvalidTransactionSigner
	}
	return nil
}

// GenerateContractAddress according to tx.from and tx.nonce.
func (tx *Transaction) GenerateContractAddress() (*Address, error) {
	if TxPayloadDeployType != tx.Type() {
		return nil, errors.New("playload type err")
	}
	return NewContractAddressFromData(tx.from.Bytes(), byteutils.FromUint64(tx.nonce))
}

// CheckContract check if contract is valid
func CheckContract(addr *Address, ws WorldState) (state.Account, error) {
	if addr == nil || ws == nil {
		return nil, ErrNilArgument
	}

	if addr.Type() != ContractAddress {
		return nil, ErrContractCheckFailed
	}

	contract, err := ws.GetContractAccount(addr.Bytes())
	if err != nil {
		return nil, err
	}

	birthEvents, err := ws.FetchEvents(contract.BirthPlace())
	if err != nil {
		return nil, err
	}

	result := false
	if birthEvents != nil && len(birthEvents) > 0 {
		event := birthEvents[len(birthEvents)-1]
		if event.Topic == TopicTransactionExecutionResult {
			txEvent := TransactionEvent{}
			if err := json.Unmarshal([]byte(event.Data), &txEvent); err != nil {
				return nil, err
			}
			if txEvent.Status == TxExecutionSuccess {
				result = true
			}
		}
	}
	if !result {
		return nil, ErrContractCheckFailed
	}

	return contract, nil
}

// CheckTransaction in a tx world state
func CheckTransaction(tx *Transaction, ws WorldState) (bool, error) {
	// check nonce
	fromAcc, err := ws.GetOrCreateUserAccount(tx.from.address)
	if err != nil {
		return true, err
	}

	// pass current Nonce.
	currentNonce := fromAcc.Nonce()

	if tx.nonce < currentNonce+1 {
		// Nonce is too small, won't giveback the tx
		return false, ErrSmallTransactionNonce
	} else if tx.nonce > currentNonce+1 {
		return true, ErrLargeTransactionNonce
	}

	return false, nil
}

// AcceptTransaction in a tx world state
func AcceptTransaction(tx *Transaction, ws WorldState) (bool, error) {
	// record tx
	pbTx, err := tx.ToProto()
	if err != nil {
		return true, err
	}
	txBytes, err := proto.Marshal(pbTx)
	if err != nil {
		return true, err
	}
	if err := ws.PutTx(tx.hash, txBytes); err != nil {
		return true, err
	}
	// incre nonce
	fromAcc, err := ws.GetOrCreateUserAccount(tx.from.address)
	if err != nil {
		return true, err
	}
	fromAcc.IncrNonce()
	// No error, won't giveback the tx
	return false, nil
}

// GetTransaction from txs Trie
func GetTransaction(hash byteutils.Hash, ws WorldState) (*Transaction, error) {
	if len(hash) != TxHashByteLength {
		return nil, ErrInvalidArgument
	}
	bytes, err := ws.GetTx(hash)
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
func (tx *Transaction) HashTransaction() (byteutils.Hash, error) {
	hasher := sha3.New256()

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

	hasher.Write(tx.from.address)
	hasher.Write(tx.to.address)
	hasher.Write(value)
	hasher.Write(byteutils.FromUint64(tx.nonce))
	hasher.Write(byteutils.FromInt64(tx.timestamp))
	hasher.Write(data)
	hasher.Write(byteutils.FromUint32(tx.chainID))
	hasher.Write(gasPrice)
	hasher.Write(gasLimit)

	return hasher.Sum(nil), nil
}
