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

/*
#include "v8/lib/nvm_error.h"
*/
import "C"

import (
	"unsafe"

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
//export GetTxByHashFunc
func GetTxByHashFunc(handler unsafe.Pointer, hash *C.char, gasCnt *C.size_t) *C.char {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		return nil
	}

	// calculate Gas.
	*gasCnt = C.size_t(1000)

	txHash, err := byteutils.FromHex(C.GoString(hash))
	if err != nil {
		return nil
	}
	txBytes, err := engine.ctx.state.GetTx(txHash)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(hash),
			"err":     err,
		}).Debug("GetTxByHashFunc get tx failed.")
		return nil
	}
	sTx, err := toSerializableTransactionFromBytes(txBytes)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(hash),
			"err":     err,
		}).Debug("GetTxByHashFunc get tx failed.")
		return nil
	}
	txJSON, err := json.Marshal(sTx)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(hash),
			"err":     err,
		}).Debug("GetTxByHashFunc get tx failed.")
		return nil
	}

	return C.CString(string(txJSON))
}

// GetAccountStateFunc returns account info by address
//export GetAccountStateFunc
func GetAccountStateFunc(handler unsafe.Pointer, address *C.char, gasCnt *C.size_t,
	result **C.char, exceptionInfo **C.char) int {
	*result = nil
	*exceptionInfo = nil
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Error("Unexpected error: failed to get engine.")
		return C.NVM_UNEXPECTED_ERR
	}

	// calculate Gas.
	*gasCnt = C.size_t(1000)

	addr, err := core.AddressParse(C.GoString(address))
	if err != nil {
		*exceptionInfo = C.CString("GetAccountState parse address failed.")
		return C.NVM_EXCEPTION_ERR
	}

	acc, err := engine.ctx.state.GetOrCreateUserAccount(addr.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"address": addr,
			"err":     err,
		}).Error("Unexpected error: GetAccountStateFunc get account state failed") //TODO: to confirm if sys err
		return C.NVM_UNEXPECTED_ERR
	}
	state := toSerializableAccount(acc)
	json, err := json.Marshal(state)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"state": state,
			"json":  json,
			"err":   err,
		}).Error("Unexpected error: GetAccountStateFunc failed to mashal account state")
		return C.NVM_UNEXPECTED_ERR
	}

	*result = C.CString(string(json))
	return C.NVM_SUCCESS
}

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
			}).Fatal("failed to marshal TransferFromContractEvent") // TODO: to confirm
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
			}).Fatal("failed to marshal TransferFromContractEvent") // TODO: to confirm
		}

		wsState.RecordEvent(txHash, &state.Event{Topic: core.TopicTransferFromContract, Data: string(eData)})

	}
}

// TransferFunc transfer vale to address
//export TransferFunc
func TransferFunc(handler unsafe.Pointer, to *C.char, v *C.char, gasCnt *C.size_t) int {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
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
		}).Fatal("Unexpected error: failed to parse contract address") //TODO: sys err ,crash
	}

	// calculate Gas.
	*gasCnt = C.size_t(2000)

	addr, err := core.AddressParse(C.GoString(to))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(to),
		}).Debug("TransferFunc parse address failed.")
		recordTransferFailureEvent(TransferAddressParseErr, cAddr.String(), "", "", height, wsState, txHash)
		return TransferAddressParseErr
	}

	toAcc, err := engine.ctx.state.GetOrCreateUserAccount(addr.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"address": addr,
			"err":     err,
		}).Fatal("GetAccountStateFunc get account state failed.") // TODO: sys err, crash
	}

	amount, err := util.NewUint128FromString(C.GoString(v))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"address": addr,
			"err":     err,
		}).Debug("GetAmountFunc get amount failed.")
		recordTransferFailureEvent(TransferStringToBigIntErr, cAddr.String(), addr.String(), "", height, wsState, txHash)
		return TransferStringToBigIntErr
	}
	// update balance
	if amount.Cmp(util.NewUint128()) > 0 {
		err = engine.ctx.contract.SubBalance(amount)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"handler": uint64(uintptr(handler)),
				"key":     C.GoString(to),
				"err":     err,
			}).Debug("TransferFunc SubBalance failed.") //TODO: to confirm
			recordTransferFailureEvent(TransferSubBalance, cAddr.String(), addr.String(), amount.String(), height, wsState, txHash)
			return TransferSubBalance
		}

		err = toAcc.AddBalance(amount)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"account": toAcc,
				"amount":  amount,
				"address": addr,
				"err":     err,
			}).Fatal("failed to add balance") //TODO: to confirm
			//			recordTransferFailureEvent(TransferSubBalance, cAddr.String(), addr.String(), amount.String(), height, wsState, txHash)
			// return TransferAddBalance
		}
	}

	recordTransferFailureEvent(TransferFuncSuccess, cAddr.String(), addr.String(), amount.String(), height, wsState, txHash)
	return TransferFuncSuccess
}

// VerifyAddressFunc verify address is valid
//export VerifyAddressFunc
func VerifyAddressFunc(handler unsafe.Pointer, address *C.char, gasCnt *C.size_t) int {
	// calculate Gas.
	*gasCnt = C.size_t(100)

	addr, err := core.AddressParse(C.GoString(address))
	if err != nil {
		return 0
	}
	return int(addr.Type())
}

// GetPreBlockHashFunc returns hash of the block before current tail by n
//export GetPreBlockHashFunc
func GetPreBlockHashFunc(handler unsafe.Pointer, distance C.ulonglong,
	gasCnt *C.size_t, result **C.char, exceptionInfo **C.char) int {
	*result = nil
	*exceptionInfo = nil
	n := uint64(distance)
	if n > uint64(maxBlockDistance) { //31 days
		*exceptionInfo = C.CString("Blockchain.GetPreBlockHash(), argument out of range")
		return C.NVM_EXCEPTION_ERR
	}

	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx == nil || engine.ctx.block == nil || engine.ctx.state == nil {
		logging.VLog().Error("Unexpected error: failed to get engine.")
		return C.NVM_UNEXPECTED_ERR
	}
	wsState := engine.ctx.state
	// calculate Gas.
	*gasCnt = C.size_t(1000) //TODO: to confirm

	//get height
	height := engine.ctx.block.Height()
	if n >= height { // have checked it in lib js
		logging.VLog().WithFields(logrus.Fields{
			"height":   height,
			"distance": n,
		}).Debug("distance is large than height")
		*exceptionInfo = C.CString("Blockchain.GetPreBlockHash(), argument[distance] is large than current height")
		return C.NVM_EXCEPTION_ERR
	}
	height -= n

	blockHash, err := wsState.GetBlockHashByHeight(height)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": height,
			"err":    err,
		}).Error("Unexpected error: Failed to get block hash from wsState by height")
		return C.NVM_UNEXPECTED_ERR
	}

	*result = C.CString(byteutils.Hex(blockHash))
	return C.NVM_SUCCESS
}

// GetPreBlockSeedFunc returns hash of the block before current tail by n
//export GetPreBlockSeedFunc
func GetPreBlockSeedFunc(handler unsafe.Pointer, distance C.ulonglong,
	gasCnt *C.size_t, result **C.char, exceptionInfo **C.char) int {
	*result = nil
	*exceptionInfo = nil

	n := uint64(distance)
	if n > uint64(maxBlockDistance) { //31 days
		*exceptionInfo = C.CString("Blockchain.GetPreBlockSeed(), argument out of range")
		return C.NVM_EXCEPTION_ERR
	}

	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx == nil || engine.ctx.block == nil || engine.ctx.state == nil {
		logging.VLog().Error("Unexpected error: failed to get engine")
		return C.NVM_UNEXPECTED_ERR
	}
	wsState := engine.ctx.state
	// calculate Gas.
	*gasCnt = C.size_t(1000) //TODO: to confirm

	//get height
	height := engine.ctx.block.Height()
	if n >= height { // have checked it in lib js
		logging.VLog().WithFields(logrus.Fields{
			"height":   height,
			"distance": n,
		}).Debug("distance is large than height")
		*exceptionInfo = C.CString("Blockchain.GetPreBlockSeed(), argument[distance] is large than current height")
		return C.NVM_EXCEPTION_ERR
	}

	height -= n
	if height < core.RandomAvailableHeight {
		*exceptionInfo = C.CString("Blockchain.GetPreBlockSeed(), seed is not available at this height")
		return C.NVM_EXCEPTION_ERR
	}

	blockHash, err := wsState.GetBlockHashByHeight(height)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": height,
			"err":    err,
		}).Error("Unexpected error: Failed to get block hash from wsState by height")
		return C.NVM_UNEXPECTED_ERR
	}

	bytes, err := wsState.GetBlock(blockHash)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": height,
			"err":    err,
		}).Error("Unexpected error: Failed to get block from wsState by hash")
		return C.NVM_UNEXPECTED_ERR
	}

	pbBlock := new(corepb.Block)
	if err = proto.Unmarshal(bytes, pbBlock); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"bytes":  bytes,
			"height": height,
			"err":    err,
		}).Error("Unexpected error: Failed to unmarshal pbBlock")
		return C.NVM_UNEXPECTED_ERR
	}

	if pbBlock.GetHeader() == nil || pbBlock.GetHeader().GetRandom() == nil ||
		pbBlock.GetHeader().GetRandom().GetVrfSeed() == nil {
		logging.VLog().WithFields(logrus.Fields{
			"pbBlock": pbBlock,
			"height":  height,
		}).Error("Unexpected error: No random found in block header")
		return C.NVM_UNEXPECTED_ERR
	}

	*result = C.CString(byteutils.Hex(pbBlock.GetHeader().GetRandom().GetVrfSeed()))
	return C.NVM_SUCCESS
}
