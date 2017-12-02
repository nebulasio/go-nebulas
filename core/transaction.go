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

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

var (
	// TransactionGasPrice default gasPrice
	TransactionGasPrice = util.NewUint128FromInt(1)

	// TransactionGas default gasLimt
	TransactionGas = util.NewUint128FromInt(20000)

	// TransactionDataGas per byte of data attached to a transaction gas cost
	TransactionDataGas = util.NewUint128FromInt(50)
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

// Nonce return tx nonce
func (tx *Transaction) Nonce() uint64 {
	return tx.nonce
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
	return errors.New("Pb Message cannot be converted into Transaction")
}

func (tx *Transaction) String() string {
	return fmt.Sprintf("Tx {from:%s; to:%s; nonce:%d, value: %d}",
		byteutils.Hex(tx.from.address),
		byteutils.Hex(tx.to.address),
		tx.nonce,
		tx.value.Int64(),
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
		gasLimit = TransactionGas
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
	// if gasPrice <= 0 , returns default gasPrice
	if tx.gasPrice.Cmp(util.NewUint128FromInt(0).Int) <= 0 {
		// TODO:the default gasPrice needs to be computed dynamically
		return TransactionGasPrice
	}
	return tx.gasPrice
}

// GasLimit returns gasLimit
func (tx *Transaction) GasLimit() *util.Uint128 {
	// if gasLimit <= 0 , returns default gasLimit
	if tx.gasPrice.Cmp(util.NewUint128FromInt(0).Int) <= 0 {
		return TransactionGas
	}
	return tx.gasLimit
}

// Cost returns value + gasprice * gaslimit.
func (tx *Transaction) Cost() *util.Uint128 {
	total := util.NewUint128().Mul(tx.GasPrice().Int, tx.GasLimit().Int)
	total.Add(total, tx.value.Int)
	return util.NewUint128FromBigInt(total)
}

// CalculateGas calculate the actual amount for a tx with data
func (tx *Transaction) CalculateGas() *util.Uint128 {
	txGas := util.NewUint128()
	txGas.Add(txGas.Int, TransactionGas.Int)
	if tx.DataLen() > 0 {
		dataGas := util.NewUint128()
		dataGas.Mul(util.NewUint128FromInt(int64(tx.DataLen())).Int, TransactionDataGas.Int)
		txGas.Add(txGas.Int, dataGas.Int)
	}
	return txGas
}

// DataLen return the length of payload
func (tx *Transaction) DataLen() int {
	return len(tx.data.Payload)
}

// Execute transaction and return result.
func (tx *Transaction) Execute(block *Block) error {
	// check balance.
	fromAcc := block.accState.GetOrCreateUserAccount(tx.from.address)
	toAcc := block.accState.GetOrCreateUserAccount(tx.to.address)

	if fromAcc.Balance().Cmp(tx.Cost().Int) < 0 {
		return ErrInsufficientBalance
	}
	if tx.gasLimit.Cmp(tx.CalculateGas().Int) < 0 {
		return ErrOutofGasLimit
	}

	// accept the transaction
	fromAcc.SubBalance(tx.value)
	toAcc.AddBalance(tx.value)
	fromAcc.IncreNonce()

	gas := util.NewUint128().Mul(tx.GasPrice().Int, tx.CalculateGas().Int)
	fromAcc.SubBalance(util.NewUint128FromBigInt(gas))

	// execute payload
	var payload TxPayload
	var err error
	switch tx.data.Type {
	case TxPayloadBinaryType:
		payload, err = LoadBinaryPayload(tx.data.Payload)
	case TxPayloadDeployType:
		payload, err = LoadDeployPayload(tx.data.Payload)
	case TxPayloadCallType:
		payload, err = LoadCallPayload(tx.data.Payload)
	default:
		return ErrInvalidTxPayloadType
	}

	if err != nil {
		return err
	}

	// execute smart contract and sub the calcute gas.
	return payload.Execute(tx, block)
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

// Verify return transaction verify result, including Hash and Signature.
func (tx *Transaction) Verify(chainID uint32) error {
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
	signVerify, err := tx.verifySign()
	if err != nil {
		return err
	}
	if !signVerify {
		return ErrInvalidSignature
	}

	return nil
}

func (tx *Transaction) verifySign() (bool, error) {
	signature, err := crypto.NewSignature(keystore.Algorithm(tx.alg))
	if err != nil {
		return false, err
	}
	pub, err := signature.RecoverPublic(tx.hash, tx.sign)
	if err != nil {
		return false, err
	}
	pubdata, err := pub.Encoded()
	if err != nil {
		return false, err
	}
	addr, err := NewAddressFromPublicKey(pubdata)
	if err != nil {
		return false, err
	}
	if !tx.from.Equals(addr) {
		log.WithFields(log.Fields{
			"recover address": addr.String(),
			"tx":              tx,
		}).Error("Transaction verifySign.")
		return false, errors.New("Transaction recover public key address not equal to from. ")
	}
	return true, nil
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
