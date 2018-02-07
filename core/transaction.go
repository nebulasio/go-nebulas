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
	TransactionMaxGasPrice = util.NewUint128FromBigInt(util.NewUint128().Mul(util.NewUint128FromInt(50).Int,
		util.NewUint128().Exp(util.NewUint128FromInt(10).Int, util.NewUint128FromInt(9).Int, nil)))

	// TransactionMaxGas max gas:50 * 10 ** 9
	TransactionMaxGas = util.NewUint128FromBigInt(util.NewUint128().Mul(util.NewUint128FromInt(50).Int,
		util.NewUint128().Exp(util.NewUint128FromInt(10).Int, util.NewUint128FromInt(9).Int, nil)))

	// TransactionGasPrice default gasPrice : 10**6
	TransactionGasPrice = util.NewUint128FromBigInt(util.NewUint128().Exp(util.NewUint128FromInt(10).Int, util.NewUint128FromInt(6).Int, nil))

	// MinGasCountPerTransaction default gas for normal transaction
	MinGasCountPerTransaction = util.NewUint128FromInt(20000)

	// GasCountPerByte per byte of data attached to a transaction gas cost
	GasCountPerByte = util.NewUint128FromInt(1)

	// DelegateBaseGasCount is base gas count of delegate transaction
	DelegateBaseGasCount = util.NewUint128FromInt(20000)
	// CandidateBaseGasCount is base gas count of candidate transaction
	CandidateBaseGasCount = util.NewUint128FromInt(20000)
	// ZeroGasCount is zero gas count
	ZeroGasCount = util.NewUint128()
)

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
	if gasPrice == nil || gasPrice.Cmp(util.NewUint128FromInt(0).Int) <= 0 {
		gasPrice = TransactionGasPrice
	}
	if gasLimit == nil || gasLimit.Cmp(util.NewUint128FromInt(0).Int) <= 0 {
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
func (tx *Transaction) PayloadGasLimit(payload TxPayload) *util.Uint128 {
	// payloadGasLimit = tx.gasLimit - tx.GasCountOfTxBase
	payloadGasLimit := util.NewUint128().Sub(tx.gasLimit.Int, tx.GasCountOfTxBase().Int)
	payloadGasLimit.Sub(payloadGasLimit, payload.BaseGasCount().Int)
	return util.NewUint128FromBigInt(payloadGasLimit)
}

// MinBalanceRequired returns gasprice * gaslimit.
func (tx *Transaction) MinBalanceRequired() *util.Uint128 {
	total := util.NewUint128().Mul(tx.GasPrice().Int, tx.GasLimit().Int)
	return util.NewUint128FromBigInt(total)
}

// GasCountOfTxBase calculate the actual amount for a tx with data
func (tx *Transaction) GasCountOfTxBase() *util.Uint128 {
	txGas := util.NewUint128()
	txGas.Add(txGas.Int, MinGasCountPerTransaction.Int)
	if tx.DataLen() > 0 {
		dataGas := util.NewUint128()
		dataGas.Mul(util.NewUint128FromInt(int64(tx.DataLen())).Int, GasCountPerByte.Int)
		txGas.Add(txGas.Int, dataGas.Int)
	}
	return txGas
}

// DataLen return the length of payload
func (tx *Transaction) DataLen() int {
	return len(tx.data.Payload)
}

// LoadPayload returns tx's payload
func (tx *Transaction) LoadPayload(block *Block) (TxPayload, error) {
	// execute payload
	var (
		payload TxPayload
		err     error
	)
	switch tx.data.Type {
	case TxPayloadBinaryType:
		if block.Height() >= 280921 && block.Height() <= 297680 || block.Height() >= 300087 && block.Height() <= 302302 {
			payload, err = LoadBinaryPayloadFail(tx.data.Payload)
		} else {
			payload, err = LoadBinaryPayload(tx.data.Payload)
		}
	case TxPayloadDeployType:
		payload, err = LoadDeployPayload(tx.data.Payload)
	case TxPayloadCallType:
		payload, err = LoadCallPayload(tx.data.Payload)
	case TxPayloadCandidateType:
		payload, err = LoadCandidatePayload(tx.data.Payload)
	case TxPayloadDelegateType:
		payload, err = LoadDelegatePayload(tx.data.Payload)
	default:
		err = ErrInvalidTxPayloadType
	}
	return payload, err
}

// LocalExecution returns tx local execution
func (tx *Transaction) LocalExecution(block *Block) (*util.Uint128, string, error) {
	// update gas to max for estimate
	tx.gasLimit = TransactionMaxGas

	txBlock, err := block.Clone()
	if err != nil {
		return nil, "", err
	}

	txBlock.begin()
	defer txBlock.rollback()

	fromAcc, err := txBlock.accState.GetOrCreateUserAccount(tx.from.address)
	if err != nil {
		return nil, "", err
	}
	fromAcc.AddBalance(tx.MinBalanceRequired())
	fromAcc.AddBalance(tx.value)

	payload, err := tx.LoadPayload(txBlock)
	if err != nil {
		return util.NewUint128(), "", err
	}

	gasUsed := tx.GasCountOfTxBase()
	gasUsed.Add(gasUsed.Int, payload.BaseGasCount().Int)

	gasExecution, result, err := payload.Execute(txBlock, tx)

	gas := util.NewUint128FromBigInt(util.NewUint128().Add(gasUsed.Int, gasExecution.Int))
	return gas, result, err
}

// VerifyExecution transaction and return result.
func (tx *Transaction) VerifyExecution(block *Block) (*util.Uint128, error) {
	// check balance.
	fromAcc, err := block.accState.GetOrCreateUserAccount(tx.from.address)
	if err != nil {
		return nil, err
	}
	toAcc, err := block.accState.GetOrCreateUserAccount(tx.to.address)
	if err != nil {
		return nil, err
	}
	coinbaseAcc, err := block.accState.GetOrCreateUserAccount(block.CoinbaseHash())
	if err != nil {
		return nil, err
	}

	// balance < gasLimit*gasPric
	if fromAcc.Balance().Cmp(tx.MinBalanceRequired().Int) < 0 {
		return util.NewUint128(), ErrInsufficientBalance
	}

	// gasLimit < gasUsed
	gasUsed := tx.GasCountOfTxBase()
	if tx.gasLimit.Cmp(gasUsed.Int) < 0 {
		logging.VLog().WithFields(logrus.Fields{
			"error":       ErrOutOfGasLimit,
			"transaction": tx,
			"limit":       tx.gasLimit.String(),
			"used":        gasUsed.String(),
		}).Debug("Failed to store the payload on chain.")
		return util.NewUint128(), ErrOutOfGasLimit
	}

	payload, err := tx.LoadPayload(block)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"error":       err,
			"block":       block,
			"transaction": tx,
		}).Debug("Failed to load payload.")
		metricsTxExeFailed.Mark(1)

		tx.gasConsumption(fromAcc, coinbaseAcc, gasUsed)
		tx.triggerEvent(TopicExecuteTxFailed, block, err)
		return gasUsed, nil
	}

	gasUsed.Add(gasUsed.Int, payload.BaseGasCount().Int)
	if tx.gasLimit.Cmp(gasUsed.Int) < 0 {
		logging.VLog().WithFields(logrus.Fields{
			"err":   ErrOutOfGasLimit,
			"block": block,
			"tx":    tx,
		}).Debug("Failed to check base gas used.")
		metricsTxExeFailed.Mark(1)

		tx.gasConsumption(fromAcc, coinbaseAcc, tx.gasLimit)
		tx.triggerEvent(TopicExecuteTxFailed, block, err)
		return tx.gasLimit, nil
	}

	// block begin
	txBlock, err := block.Clone()
	if err != nil {
		return util.NewUint128(), err
	}

	// execute smart contract and sub the calcute gas.
	gasExecution, _, exeErr := payload.Execute(txBlock, tx)

	if exeErr != nil {
		txBlock.rollback()
	} else {
		block.Merge(txBlock)
	}

	fromAcc, err = block.accState.GetOrCreateUserAccount(tx.from.address)
	if err != nil {
		return nil, err
	}
	toAcc, err = block.accState.GetOrCreateUserAccount(tx.to.address)
	if err != nil {
		return nil, err
	}
	coinbaseAcc, err = block.accState.GetOrCreateUserAccount(block.CoinbaseHash())
	if err != nil {
		return nil, err
	}

	// gas = tx.GasCountOfTxBase() +  gasExecution
	gas := util.NewUint128FromBigInt(util.NewUint128().Add(gasUsed.Int, gasExecution.Int))
	tx.gasConsumption(fromAcc, coinbaseAcc, gas)

	if exeErr != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":          err,
			"block":        block,
			"tx":           tx,
			"gasUsed":      gasUsed.String(),
			"gasExecution": gasExecution.String(),
		}).Debug("Failed to execute payload.")

		metricsTxExeFailed.Mark(1)
		tx.triggerEvent(TopicExecuteTxFailed, block, err)
	} else {
		if fromAcc.Balance().Cmp(tx.value.Int) < 0 {
			logging.VLog().WithFields(logrus.Fields{
				"err":   ErrInsufficientBalance,
				"block": block,
				"tx":    tx,
			}).Debug("Failed to check balance sufficient.")

			metricsTxExeFailed.Mark(1)
			tx.triggerEvent(TopicExecuteTxFailed, block, ErrInsufficientBalance)
		} else {
			// accept the transaction
			fromAcc.SubBalance(tx.value)
			toAcc.AddBalance(tx.value)

			metricsTxExeSuccess.Mark(1)
			// record tx execution success event
			tx.triggerEvent(TopicExecuteTxSuccess, block, nil)
		}
	}

	return gas, nil
}

func (tx *Transaction) gasConsumption(from, coinbase state.Account, gas *util.Uint128) {
	gasCost := util.NewUint128().Mul(tx.GasPrice().Int, gas.Int)
	from.SubBalance(util.NewUint128FromBigInt(gasCost))
	coinbase.AddBalance(util.NewUint128FromBigInt(gasCost))
}

func (tx *Transaction) triggerEvent(topic string, block *Block, err error) {
	var txData []byte
	pbTx, _ := tx.ToProto()
	if err != nil {
		var (
			txErrEvent struct {
				Transaction proto.Message `json:"transaction"`
				Error       error         `json:"error"`
			}
		)
		txErrEvent.Transaction = pbTx
		txErrEvent.Error = err
		txData, _ = json.Marshal(txErrEvent)
	} else {
		txData, _ = json.Marshal(pbTx)
	}

	event := &Event{Topic: topic,
		Data: string(txData)}
	block.recordEvent(tx.hash, event)
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
