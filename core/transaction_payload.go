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
	"encoding/json"
	"errors"

	"github.com/nebulasio/go-nebulas/core/pb"
	log "github.com/sirupsen/logrus"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/nf/nvm"
)

const (
	TxPayloadBinaryType = "binary"
	TxPayloadDeployType = "deploy"
	TxPayloadCallType   = "call"
	TxPayloadVoteType   = "vote"
)

var (
	ErrInvalidTxPayloadType   = errors.New("invalid transaction data payload type")
	ErrInvalidContractAddress = errors.New("invalid contract address")
)

type txPayload struct {
	PayloadType string
	Source      string
	Function    string
	Args        string
	binaryData  []byte
}

func (payload *txPayload) Execute(tx *Transaction, block *Block) error {
	if payload.PayloadType == TxPayloadBinaryType {
		return nil
	}

	contractAddress := tx.TargetContractAddress()
	contractAccount, created := block.FindOrCreateAccount(contractAddress)
	if created == true {
		contractAccount.SetContractTransactionHash(tx.Hash())
	}

	stateTrie := block.stateTrie
	gcsTrie := block.FindAccount(contractAccount.ContractOwner).UserGlobalStorage
	lcsTrie := contractAccount.ContractLocalStorage

	// balance trie.
	stateTrie.BeginBatch()
	gcsTrie.BeginBatch()
	lcsTrie.BeginBatch()

	shouldCommit := false

	defer (func() {
		if shouldCommit {
			stateTrie.Commit()
			gcsTrie.Commit()
			lcsTrie.Commit()
		} else {
			stateTrie.RollBack()
			gcsTrie.RollBack()
			lcsTrie.RollBack()
		}
	})()

	engine := nvm.NewV8Engine(stateTrie, lcsTrie, gcsTrie)
	defer engine.Dispose()

	var err error
	switch payload.PayloadType {
	case TxPayloadBinaryType:
		err = nil
	case TxPayloadDeployType:
		err = engine.DeployAndInit(payload.Source, payload.Args)
	case TxPayloadCallType:
		source, err := findContractSource(block, contractAddress, contractAccount)
		if err == nil {
			err = engine.Call(source, payload.Function, payload.Args)
		}
	default:
		err = ErrInvalidTxPayloadType
	}

	if err == nil {
		shouldCommit = true
	}

	return err
}

func NewDeploySCPayload(source, args string) ([]byte, error) {
	payload := &txPayload{
		PayloadType: TxPayloadDeployType,
		Source:      source,
		Args:        args,
	}
	return json.Marshal(payload)
}

func NewCallSCPayload(function, args string) ([]byte, error) {
	payload := &txPayload{
		PayloadType: TxPayloadCallType,
		Function:    function,
		Args:        args,
	}
	return json.Marshal(payload)
}

func parseTxPayload(data []byte) (*txPayload, error) {
	payload := &txPayload{}
	if err := json.Unmarshal(data, &payload); err != nil {
		payload.PayloadType = TxPayloadBinaryType
		payload.binaryData = data
	}
	return payload, nil
}

func isContractPayload(data []byte) (bool, *txPayload) {
	if data == nil || len(data) == 0 {
		return false, nil
	}

	txPayload, err := parseTxPayload(data)
	if err != nil {
		return false, nil
	}

	return true, txPayload
}

func findContractSource(block *Block, contractAddress *Address, contractAccount *Account) (string, error) {
	logFields := log.Fields{
		"contractAddress": contractAddress.ToHex(),
		"contractAccount": contractAccount.ContractTransactionHash.Hex(),
		"err":             nil,
	}

	txBytes, err := block.txsTrie.Get(contractAccount.ContractTransactionHash)
	if err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Error("get contract transaction from txsTrie failed.")
		return "", err
	}

	pbTx := new(corepb.Transaction)
	if err = proto.Unmarshal(txBytes, pbTx); err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Error("unmarshal contract transaction to corepb.Transaction failed.")

		return "", err
	}

	tx := new(Transaction)
	if err = tx.FromProto(pbTx); err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Error("convert corepb.Trnsaction to Core.Transaction failed.")
		return "", err
	}

	payload, err := parseTxPayload(tx.data)
	if err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Error("parse transaction payload failed.")
		return "", err
	}

	if payload.PayloadType != TxPayloadDeployType {
		err = ErrInvalidContractAddress
		logFields["err"] = err
		log.WithFields(logFields).Error("transaction must be the type of smart contract deployment.")
		return "", err
	}

	return payload.Source, nil
}
