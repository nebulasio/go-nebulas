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

import "errors"

// Payload Types
const (
	TxPayloadBinaryType = "binary"
	TxPayloadDeployType = "deploy"
	TxPayloadCallType   = "call"
	TxPayloadVoteType   = "vote"
)

// Error Types
var (
	ErrInvalidTxPayloadType   = errors.New("invalid transaction data payload type")
	ErrInvalidContractAddress = errors.New("invalid contract address")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrOutofGasLimit          = errors.New("out of gas limit")
	ErrInvalidSignature       = errors.New("invalid transaction signature")
	ErrInvalidTransactionHash = errors.New("invalid transaction hash")
)

// TxPayload stored in tx
type TxPayload interface {
	ToBytes() ([]byte, error)
	Execute(tx *Transaction, block *Block) error
}

// MessageType
const (
	MessageTypeNewBlock = "newblock"
	MessageTypeNewTx    = "newtx"
)

// Consensus interface
type Consensus interface {
	VerifyBlock(*Block) error
}
