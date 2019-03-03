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

package nvm

import (
	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// GetTxByHashFunc returns tx info by hash
// Return: string(value), uint64(gasCnt), bool(notnil)
func GetTxByHashFunc(handler uint64, hash string) (string, uint64, bool) {

	var gasCnt uint64 = 0
	var emptyRes string = ""

	engine, _ := getEngineByStorageHandler(handler)
	if engine == nil || engine.ctx.block == nil {
		return emptyRes, gasCnt, false
	}

	// calculate Gas.
	gasCnt = uint64(GetTxByHashGasBase)

	txHash, err := byteutils.FromHex(hash)
	if err != nil {
		return emptyRes, gasCnt, false
	}
	txBytes, err := engine.ctx.state.GetTx(txHash)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": handler,
			"key":     hash,
			"err":     err,
		}).Debug("GetTxByHashFunc get tx failed.")
		return emptyRes, gasCnt, false
	}
	sTx, err := toSerializableTransactionFromBytes(txBytes)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": handler,
			"key":     hash,
			"err":     err,
		}).Debug("GetTxByHashFunc get tx failed.")
		return emptyRes, gasCnt, false
	}
	txJSON, err := json.Marshal(sTx)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": handler,
			"key":     hash,
			"err":     err,
		}).Debug("GetTxByHashFunc get tx failed.")
		return emptyRes, gasCnt, false
	}

	return string(txJSON), gasCnt, true
}

// GetAccountStateFunc returns account info by address
// Return: execution result code, result string, exception info string, gascnt, notnil
func GetAccountStateFunc(handler uint64, address string) (int, string, string, uint64, bool) {

	var result string = ""
	var exceptionInfo string = ""
	var gasCnt uint64 = 0

	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Error("Unexpected error: failed to get engine")
		return NVM_UNEXPECTED_ERR, result, exceptionInfo, gasCnt, false
	}

	// calculate Gas.
	gasCnt = uint64(GetAccountStateGasBase)

	addr, err := core.AddressParse(address)
	if err != nil {
		exceptionInfo = "Blockchain.getAccountState(), parse address failed"
		return NVM_EXCEPTION_ERR, result, exceptionInfo, gasCnt, false
	}

	acc, err := engine.ctx.state.GetOrCreateUserAccount(addr.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": handler,
			"address": addr,
			"err":     err,
		}).Error("Unexpected error: GetAccountStateFunc get account state failed")
		return NVM_UNEXPECTED_ERR, result, exceptionInfo, gasCnt, false
	}
	state := toSerializableAccount(acc)
	json, err := json.Marshal(state)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"state": state,
			"json":  json,
			"err":   err,
		}).Error("Unexpected error: GetAccountStateFunc failed to mashal account state")
		return NVM_UNEXPECTED_ERR, result, exceptionInfo, gasCnt, false
	}

	result = string(json)
	return NVM_SUCCESS, result, exceptionInfo, gasCnt, true
}

// Record transfer failure event
func recordTransferFailureEvent(errNo int, from string, to string, value string,
	height uint64, wsState WorldState, txHash byteutils.Hash) {

	if errNo == TransferFuncSuccess && height > core.TransferFromContractEventRecordableHeight {
		event := &TransferFromContractEvent{
			Amount: value,
			From:   from,
			To:     to,
		}
		eData, err := json.Marshal(event)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"from":   from,
				"to":     to,
				"amount": value,
				"err":    err,
			}).Fatal("failed to marshal TransferFromContractEvent")
		}
		wsState.RecordEvent(txHash, &state.Event{Topic: core.TopicTransferFromContract, Data: string(eData)})

	} else if height >= core.TransferFromContractFailureEventRecordableHeight {
		var errMsg string
		switch errNo {
		case TransferFuncSuccess:
			errMsg = ""
		case TransferAddressParseErr:
			errMsg = "failed to parse to address"
		case TransferStringToBigIntErr:
			errMsg = "failed to parse transfer amount"
		case TransferSubBalance:
			errMsg = "failed to sub balace from contract address"
		default:
			logging.VLog().WithFields(logrus.Fields{
				"from":   from,
				"to":     to,
				"amount": value,
				"errNo":  errNo,
			}).Fatal("failed to marshal TransferFromContractEvent")
		}

		status := uint8(1)
		if errNo != TransferFuncSuccess {
			status = 0
		}

		event := &TransferFromContractFailureEvent{
			Amount: value,
			From:   from,
			To:     to,
			Status: status,
			Error:  errMsg,
		}

		eData, err := json.Marshal(event)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"from":   from,
				"to":     to,
				"amount": value,
				"status": event.Status,
				"error":  err,
			}).Fatal("failed to marshal TransferFromContractEvent")
		}

		wsState.RecordEvent(txHash, &state.Event{Topic: core.TopicTransferFromContract, Data: string(eData)})
	}
}

// TransferFunc transfer value to address "to"
func TransferFunc(handler uint64, to string, v string) (int, uint64) {

	engine, _ := getEngineByStorageHandler(handler)
	if engine == nil || engine.ctx == nil || engine.ctx.block == nil ||
		engine.ctx.state == nil || engine.ctx.tx == nil {
		logging.VLog().Fatal("Unexpected error: failed to get engine.")
	}

	wsState := engine.ctx.state
	height := engine.ctx.block.Height()
	txHash := engine.ctx.tx.Hash()

	cAddr, err := core.AddressParseFromBytes(engine.ctx.contract.Address())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"txhash":  engine.ctx.tx.Hash().String(),
			"address": engine.ctx.contract.Address(),
			"err":     err,
		}).Fatal("Unexpected error: failed to parse contract address")
	}

	// calculate Gas.
	var gasCnt uint64 = 0
	gasCnt = uint64(TransferGasBase)

	addr, err := core.AddressParse(to)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler":   handler,
			"toAddress": to,
		}).Debug("TransferFunc parse address failed.")
		recordTransferFailureEvent(TransferAddressParseErr, cAddr.String(), "", "", height, wsState, txHash)
		return TransferAddressParseErr, gasCnt
	}

	toAcc, err := engine.ctx.state.GetOrCreateUserAccount(addr.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": handler,
			"address": addr,
			"err":     err,
		}).Fatal("GetAccountStateFunc get account state failed.")
	}

	amount, err := util.NewUint128FromString(v)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": handler,
			"address": addr,
			"err":     err,
		}).Debug("GetAmountFunc get amount failed.")
		recordTransferFailureEvent(TransferStringToBigIntErr, cAddr.String(), addr.String(), "", height, wsState, txHash)
		return TransferStringToBigIntErr, gasCnt
	}

	// update balance
	if amount.Cmp(util.NewUint128()) > 0 {
		err = engine.ctx.contract.SubBalance(amount)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"handler": handler,
				"key":     to,
				"err":     err,
			}).Debug("TransferFunc SubBalance failed.")
			recordTransferFailureEvent(TransferSubBalance, cAddr.String(), addr.String(), amount.String(), height, wsState, txHash)
			return TransferSubBalance, gasCnt
		}

		err = toAcc.AddBalance(amount)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"account": toAcc,
				"amount":  amount,
				"address": addr,
				"err":     err,
			}).Fatal("failed to add balance")
			// recordTransferFailureEvent(TransferSubBalance, cAddr.String(), addr.String(), amount.String(), height, wsState, txHash)
			// return TransferAddBalance
		}
	}

	recordTransferFailureEvent(TransferFuncSuccess, cAddr.String(), addr.String(), amount.String(), height, wsState, txHash)
	return TransferFuncSuccess, gasCnt
}

// VerifyAddressFunc verify address is valid
//export VerifyAddressFunc
func VerifyAddressFunc(handler uint64, address string) (int, uint64) {
	// calculate Gas.
	gasCnt := uint64(VerifyAddressGasBase)

	addr, err := core.AddressParse(address)
	if err != nil {
		return 0, gasCnt
	}
	return int(addr.Type()), gasCnt
}

// GetPreBlockHashFunc returns hash of the block before current tail by n
//params: handler, offset
//return: execution result(int), result(string), exceptioninfo(string), gascnt(uint64), notNil(bool)
func GetPreBlockHashFunc(handler uint64, offset uint64) (int, string, string, uint64, bool) {

	var gasCnt uint64 = 0
	var result string = ""
	var exceptionInfo string = ""

	n := uint64(offset)
	if n > uint64(maxBlockOffset) { //31 days
		exceptionInfo = "Blockchain.GetPreBlockHash(), argument out of range"
		return NVM_EXCEPTION_ERR, result, exceptionInfo, gasCnt, false
	}

	engine, _ := getEngineByStorageHandler(handler)
	if engine == nil || engine.ctx == nil || engine.ctx.block == nil || engine.ctx.state == nil {
		logging.VLog().Error("Unexpected error: failed to get engine.")
		return NVM_UNEXPECTED_ERR, result, exceptionInfo, gasCnt, false
	}

	wsState := engine.ctx.state
	// calculate Gas.
	gasCnt = uint64(GetPreBlockHashGasBase)

	//get height
	height := engine.ctx.block.Height()
	if n >= height { // have checked it in lib js
		logging.VLog().WithFields(logrus.Fields{
			"height": height,
			"offset": n,
		}).Debug("offset is large than height")
		exceptionInfo = "Blockchain.GetPreBlockHash(), argument[offset] is large than current height"
		return NVM_EXCEPTION_ERR, result, exceptionInfo, gasCnt, false
	}
	height -= n

	blockHash, err := wsState.GetBlockHashByHeight(height)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": height,
			"err":    err,
		}).Error("Unexpected error: Failed to get block hash from wsState by height")
		return NVM_UNEXPECTED_ERR, result, exceptionInfo, gasCnt, false
	}

	result = byteutils.Hex(blockHash)
	return NVM_SUCCESS, result, exceptionInfo, gasCnt, true
}

// GetPreBlockSeedFunc returns hash of the block before current tail by n
//params: handler(uint64), offset(uint64)
//return: execution result(int), result(string), exceptionInfo(string), gasCnt(uint64), notNil(bool)
func GetPreBlockSeedFunc(handler uint64, offset uint64) (int, string, string, uint64, bool) {

	var gasCnt uint64 = 0
	var result string = ""
	var exceptionInfo string = ""

	n := uint64(offset)
	if n > uint64(maxBlockOffset) { //31 days
		exceptionInfo = "Blockchain.GetPreBlockSeed(), argument out of range"
		return NVM_EXCEPTION_ERR, result, exceptionInfo, gasCnt, false
	}

	engine, _ := getEngineByStorageHandler(handler)
	if engine == nil || engine.ctx == nil || engine.ctx.block == nil || engine.ctx.state == nil {
		logging.VLog().Error("Unexpected error: failed to get engine")
		return NVM_UNEXPECTED_ERR, result, exceptionInfo, gasCnt, false
	}
	wsState := engine.ctx.state

	// calculate Gas.
	gasCnt = uint64(GetPreBlockSeedGasBase)

	//get height
	height := engine.ctx.block.Height()
	if n >= height { // have checked it in lib js
		logging.VLog().WithFields(logrus.Fields{
			"height": height,
			"offset": n,
		}).Debug("offset is large than height")
		exceptionInfo = "Blockchain.GetPreBlockSeed(), argument[offset] is large than current height"
		return NVM_EXCEPTION_ERR, result, exceptionInfo, gasCnt, false
	}

	height -= n
	if height < core.RandomAvailableHeight {
		exceptionInfo = "Blockchain.GetPreBlockSeed(), seed is not available at this height"
		return NVM_EXCEPTION_ERR, result, exceptionInfo, gasCnt, false
	}

	blockHash, err := wsState.GetBlockHashByHeight(height)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height":    height,
			"err":       err,
			"blockHash": blockHash,
		}).Error("Unexpected error: Failed to get block hash from wsState by height")
		return NVM_UNEXPECTED_ERR, result, exceptionInfo, gasCnt, false
	}

	bytes, err := wsState.GetBlock(blockHash)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height":    height,
			"err":       err,
			"blockHash": blockHash,
		}).Error("Unexpected error: Failed to get block from wsState by hash")
		return NVM_UNEXPECTED_ERR, result, exceptionInfo, gasCnt, false
	}

	pbBlock := new(corepb.Block)
	if err = proto.Unmarshal(bytes, pbBlock); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"bytes":  bytes,
			"height": height,
			"err":    err,
		}).Error("Unexpected error: Failed to unmarshal pbBlock")
		return NVM_UNEXPECTED_ERR, result, exceptionInfo, gasCnt, false
	}

	if pbBlock.GetHeader() == nil || pbBlock.GetHeader().GetRandom() == nil ||
		pbBlock.GetHeader().GetRandom().GetVrfSeed() == nil {
		logging.VLog().WithFields(logrus.Fields{
			"pbBlock": pbBlock,
			"height":  height,
		}).Error("Unexpected error: No random found in block header")
		return NVM_UNEXPECTED_ERR, result, exceptionInfo, gasCnt, false
	}

	result = byteutils.Hex(pbBlock.GetHeader().GetRandom().GetVrfSeed())
	return NVM_SUCCESS, result, exceptionInfo, gasCnt, true
}