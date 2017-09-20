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

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/ecdsa"
	"github.com/nebulasio/go-nebulas/crypto/keystore/key"
	"github.com/nebulasio/go-nebulas/utils/byteutils"
	log "github.com/sirupsen/logrus"
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
	from      Address
	to        Address
	value     uint64
	nonce     uint64
	timestamp time.Time
	data      []byte

	// Signature values
	sign Hash
}

type TxStream struct {
	Hash  []byte
	From  []byte
	To    []byte
	Value uint64
	Nonce uint64
	Time  int64
	Data  []byte
}

// Serialize a transaction
func (tx *Transaction) Serialize() ([]byte, error) {
	serializer := &byteutils.JSONSerializer{}
	data := TxStream{
		tx.hash,
		tx.from.address,
		tx.to.address,
		tx.value,
		tx.nonce,
		tx.timestamp.UnixNano(),
		tx.data,
	}
	return serializer.Serialize(data)
}

// Deserialize a transaction
func (tx *Transaction) Deserialize(blob []byte) error {
	serializer := &byteutils.JSONSerializer{}
	var data TxStream
	if err := serializer.Deserialize(blob, &data); err != nil {
		return err
	}
	tx.hash = data.Hash
	tx.from = Address{data.From}
	tx.to = Address{data.To}
	tx.value = data.Value
	tx.nonce = data.Nonce
	tx.timestamp = time.Unix(0, data.Time)
	tx.data = data.Data
	return nil
}

// Transactions is an alias of Transaction array.
type Transactions []*Transaction

// Serialize txs
func (txs *Transactions) Serialize() ([]byte, error) {
	var data [][]byte
	serializer := &byteutils.JSONSerializer{}
	for _, v := range *txs {
		ir, err := v.Serialize()
		if err != nil {
			return nil, err
		}
		data = append(data, ir)
	}
	return serializer.Serialize(data)
}

// Deserialize txs
func (txs *Transactions) Deserialize(blob []byte) error {
	var data [][]byte
	serializer := &byteutils.JSONSerializer{}
	if err := serializer.Deserialize(blob, &data); err != nil {
		return err
	}
	for _, v := range data {
		tx := &Transaction{}
		if err := tx.Deserialize(v); err != nil {
			return err
		}
		*txs = append(*txs, tx)
	}
	return nil
}

// NewTransaction create #Transaction instance.
func NewTransaction(from, to Address, value uint64, nonce uint64, data []byte) *Transaction {
	tx := &Transaction{
		from:      from,
		to:        to,
		value:     value,
		nonce:     nonce,
		timestamp: time.Now(),
		data:      data,
	}
	return tx
}

// Hash return the hash of transaction.
func (tx *Transaction) Hash() Hash {
	return tx.hash
}

// Sign sign transaction.
func (tx *Transaction) Sign(ks *keystore.Keystore, passphrase []byte) error {
	// TODO(larry.wang):impl sign ,change input to context ,keystore in context
	tx.hash = HashTransaction(tx)
	passparam, err := key.NewPassphrase(passphrase)
	if err != nil {
		return err
	}
	alias := key.Alias(string(tx.from.address))
	key, err := ks.GetKey(alias, passparam)
	if err != nil {
		return err
	}
	signer := &ecdsa.Signature{}
	priv, err := key.Encoded()
	if err != nil {
		return err
	}
	signer.InitSign(priv)
	signature, err := signer.Sign(tx.hash)
	if err != nil {
		return err
	}
	tx.sign = signature
	return nil
}

// Verify return transaction verify result, including Hash and Signature.
func (tx *Transaction) Verify() error {
	wantedHash := HashTransaction(tx)
	if wantedHash.Equals(tx.hash) == false {
		log.WithFields(log.Fields{
			"func": "Transaction.Verify",
			"err":  ErrInvalidTransactionHash,
			"tx":   tx,
		}).Error("invalid transaction hash")
		return ErrInvalidTransactionHash
	}

	// TODO: verify signature and from address.

	return nil
}

// VerifySign tx
func (tx *Transaction) VerifySign() (bool, error) {
	if len(tx.sign) == 0 {
		return false, errors.New("Transaction: VerifySign need sign hash")
	}
	pub, err := ecdsa.RecoverPublicKey(tx.hash, tx.sign)
	if err != nil {
		return false, err
	}
	return ecdsa.Verify(tx.hash, tx.sign, pub), nil
}

// Execute execute transaction, eg. transfer Nas, call smart contract.
func (tx *Transaction) Execute(stateTrie *trie.Trie) error {
	fromOrigBalance := uint64(0)
	toOriginBalance := uint64(0)

	if v, _ := stateTrie.Get(tx.from.address); v != nil {
		fromOrigBalance = byteutils.Uint64(v)
	}

	if v, _ := stateTrie.Get(tx.to.address); v != nil {
		toOriginBalance = byteutils.Uint64(v)
	}

	if fromOrigBalance < tx.value {
		return ErrInsufficientBalance
	}

	fromBalance := fromOrigBalance - tx.value
	toBalance := toOriginBalance + tx.value

	stateTrie.Put(tx.from.address, byteutils.FromUint64(fromBalance))
	stateTrie.Put(tx.to.address, byteutils.FromUint64(toBalance))

	return nil
}

// HashTransaction hash the transaction.
func HashTransaction(tx *Transaction) Hash {
	return hash.Sha3256(
		tx.from.address,
		tx.to.address,
		byteutils.FromUint64(tx.value),
		byteutils.FromUint64(tx.nonce),
		byteutils.FromInt64(tx.timestamp.UnixNano()),
		tx.data,
	)
}
