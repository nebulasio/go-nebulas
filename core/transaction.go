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
)

var (
	// ErrInsufficientBalance insufficient balance error.
	ErrInsufficientBalance = errors.New("insufficient balance")

	// ErrInvalidSignature the signature is not sign by from address.
	ErrInvalidSignature = errors.New("invalid transaction signature")

	// ErrInvalidTransactionHash invalid hash.
	ErrInvalidTransactionHash = errors.New("invalid transaction hash")
)

// Transaction type is used to handle all transaction data.
type Transaction struct {
	hash      Hash
	from      *Address
	to        *Address
	value     *util.Uint128
	nonce     uint64
	timestamp int64
	data      []byte
	chainID   uint32

	// Signature
	alg  uint8 // algorithm
	sign Hash  // Signature values
}

// From return from address
func (tx *Transaction) From() *Address {
	return tx.from
}

// Nonce return tx nonce
func (tx *Transaction) Nonce() uint64 {
	return tx.nonce
}

// DataLen return data length
func (tx *Transaction) DataLen() int {
	return len(tx.data)
}

// ToProto converts domain Tx to proto Tx
func (tx *Transaction) ToProto() (proto.Message, error) {
	value, err := tx.value.ToFixedSizeByteSlice()
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
func NewTransaction(chainID uint32, from, to *Address, value *util.Uint128, nonce uint64, data []byte) *Transaction {
	tx := &Transaction{
		from:      from,
		to:        to,
		value:     value,
		nonce:     nonce,
		timestamp: time.Now().Unix(),
		chainID:   chainID,
		data:      data,
	}
	return tx
}

// Hash return the hash of transaction.
func (tx *Transaction) Hash() Hash {
	return tx.hash
}

// TargetContractAddress return the target contract address.
func (tx *Transaction) TargetContractAddress() *Address {
	isContractPayload, txPayload := isContractPayload(tx.data)
	if isContractPayload == false {
		return nil
	}

	// deploy contract has different contract address rules.
	if txPayload.PayloadType == TxPayloadDeployType {
		return tx.generateContractAddress()
	}

	// tx.to is the contract address.
	return tx.to

}

// Execute transaction and return result.
func (tx *Transaction) Execute(block *Block) error {
	// check balance.
	fromAcc := block.FindAccount(tx.from)
	toAcc := block.FindAccount(tx.to)

	if fromAcc.UserBalance.Cmp(tx.value.Int) < 0 {
		return ErrInsufficientBalance
	}

	// accept the transaction
	fromAcc.SubBalance(tx.value)
	toAcc.AddBalance(tx.value)
	fromAcc.IncreNonce()

	// execute smart contract if needed.
	if tx.DataLen() > 0 {
		txPayload, err := parseTxPayload(tx.data)
		if err != nil {
			return err
		}

		if err := txPayload.Execute(tx, block); err != nil {
			return err
		}
	}

	// save account info in state trie
	block.saveAccount(tx.from, fromAcc)
	block.saveAccount(tx.to, toAcc)

	return nil
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
		return false, errors.New("recover public key not related to from address")
	}
	return signature.Verify(tx.hash, tx.sign)
}

// generateContractAddress generate and return contract address according to tx.from and tx.nonce.
func (tx *Transaction) generateContractAddress() *Address {
	address, _ := NewContractAddressFromHash(hash.Sha3256(tx.from.Bytes(), byteutils.FromUint64(tx.nonce)))
	return address
}

// HashTransaction hash the transaction.
func HashTransaction(tx *Transaction) (Hash, error) {
	bytes, err := tx.value.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	return hash.Sha3256(
		tx.from.address,
		tx.to.address,
		bytes,
		byteutils.FromUint64(tx.nonce),
		byteutils.FromInt64(tx.timestamp),
		tx.data,
		byteutils.FromUint32(tx.chainID),
	), nil
}
