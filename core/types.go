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
	"hash/fnv"
	"strconv"

	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
)

// Payload Types
const (
	TxPayloadBinaryType    = "binary"
	TxPayloadDeployType    = "deploy"
	TxPayloadCallType      = "call"
	TxPayloadDelegateType  = "delegate"
	TxPayloadCandidateType = "candidate"
)

// Error Types
var (
	ErrInvalidTxPayloadType                = errors.New("invalid transaction data payload type")
	ErrInvalidBlockCannotFindParentInLocal = errors.New("invalid block received, download its parent from others")
	ErrLinkToWrongParentBlock              = errors.New("link the block to a block who is not its parent")
	ErrInvalidContractAddress              = errors.New("invalid contract address")
	ErrInsufficientBalance                 = errors.New("insufficient balance")
	ErrBelowGasPrice                       = errors.New("below the gas price")
	ErrOutOfGasLimit                       = errors.New("out of gas limit")
	ErrTxExecutionFailed                   = errors.New("transaction execution failed")
	ErrInvalidSignature                    = errors.New("invalid transaction signature")
	ErrInvalidTransactionHash              = errors.New("invalid transaction hash")
	ErrMissingParentBlock                  = errors.New("cannot find the block's parent block in storage")
	ErrTooFewCandidates                    = errors.New("the size of candidates in consensus is un-safe, should be greater than or equal " + strconv.Itoa(SafeSize))
	ErrNotBlockForgTime                    = errors.New("now is not time to forg block")
	ErrInvalidBlockHash                    = errors.New("invalid block hash")
	ErrInvalidBlockStateRoot               = errors.New("invalid block state root hash")
	ErrInvalidBlockTxsRoot                 = errors.New("invalid block txs root hash")
	ErrInvalidBlockEventsRoot              = errors.New("invalid block events root hash")
	ErrInvalidBlockDposContextRoot         = errors.New("invalid block dpos context root hash")
	ErrInvalidChainID                      = errors.New("invalid transaction chainID")
	ErrDuplicatedTransaction               = errors.New("duplicated transaction")
	ErrSmallTransactionNonce               = errors.New("cannot accept a transaction with smaller nonce")
	ErrLargeTransactionNonce               = errors.New("cannot accept a transaction with too bigger nonce")
	ErrDuplicatedBlock                     = errors.New("duplicated block")
	ErrDoubleBlockMinted                   = errors.New("double block minted")
	ErrInvalidAddress                      = errors.New("address: invalid address")
	ErrInvalidAddressDataLength            = errors.New("address: invalid address data length")
	ErrDoubleSealBlock                     = errors.New("cannot seal a block twice")
	ErrInvalidCandidatePayloadAction       = errors.New("invalid transaction candidate payload action")
	ErrInvalidDelegatePayloadAction        = errors.New("invalid transaction vote payload action")
	ErrInvalidDelegateToNonCandidate       = errors.New("cannot delegate to non-candidate")
	ErrInvalidUnDelegateFromNonDelegatee   = errors.New("cannot un-delegate from non-delegatee")
	ErrInvalidBaseAndNextDynastyID         = errors.New("cannot kickout from baseDynastyID to nextDynastyID if nextDynastyID <= baseDynastyID")
	ErrInitialDynastyNotEnough             = errors.New("the size of initial dynasty in genesis block is un-safe, should be greater than or equal " + strconv.Itoa(SafeSize))
	ErrInvalidTransactionSigner            = errors.New("transaction recover public key address not equal to from")
	ErrNotBlockInCanonicalChain            = errors.New("cannot find the block in canonical chain")
	ErrCloneAccountState                   = errors.New("Failed to clone account state")
	ErrCloneTxsState                       = errors.New("Failed to clone txs state")
	ErrCloneDynastyTrie                    = errors.New("Failed to clone dynasty trie")
	ErrCloneNextDynastyTrie                = errors.New("Failed to clone next dynasty trie")
	ErrCloneDelegateTrie                   = errors.New("Failed to clone delegate trie")
	ErrCloneCandidatesTrie                 = errors.New("Failed to clone candidates trie")
	ErrCloneVoteTrie                       = errors.New("Failed to clone vote trie")
	ErrCloneMintCntTrie                    = errors.New("Failed to clone mint count trie")
	ErrCloneEventsState                    = errors.New("Failed to clone events state")
	ErrGenerateNextDynastyContext          = errors.New("Failed to generate next dynasty context")
	ErrLoadNextDynastyContext              = errors.New("Failed to load next dynasty context")
)

// Default gas count
var (
	DefaultPayloadGas = util.NewUint128FromInt(1)
)

// TxPayload stored in tx
type TxPayload interface {
	ToBytes() ([]byte, error)
	BaseGasCount() *util.Uint128
	Execute(ctx *PayloadContext) (*util.Uint128, error)
}

// MessageType
const (
	MessageTypeNewBlock             = "newblock"
	MessageTypeDownloadedBlock      = "dlblock"
	MessageTypeDownloadedBlockReply = "dlreply"
	MessageTypeNewTx                = "newtx"
)

// Consensus interface
type Consensus interface {
	VerifyBlock(block *Block, parent *Block) error
	FastVerifyBlock(block *Block) error
}

// Less return if a < b
func Less(a *Block, b *Block) bool {
	hasherA := fnv.New32a()
	hasherA.Write(a.Hash())
	scoreA := hasherA.Sum32()
	hasherB := fnv.New32a()
	hasherB.Write(b.Hash())
	scoreB := hasherB.Sum32()
	return scoreA < scoreB
}

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Genesis() *corepb.Genesis
	Storage() storage.Storage
	EventEmitter() *EventEmitter
	StartSync()
}
