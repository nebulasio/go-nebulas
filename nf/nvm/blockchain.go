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
	"fmt"
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
	*gasCnt = C.size_t(GetTxByHashGasBase)

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
		logging.VLog().Error("Unexpected error: failed to get engine")
		return C.NVM_UNEXPECTED_ERR
	}

	// calculate Gas.
	*gasCnt = C.size_t(GetAccountStateGasBase)

	addr, err := core.AddressParse(C.GoString(address))
	if err != nil {
		*exceptionInfo = C.CString("Blockchain.getAccountState(), parse address failed")
		return C.NVM_EXCEPTION_ERR
	}

	acc, err := engine.ctx.state.GetOrCreateUserAccount(addr.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"address": addr,
			"err":     err,
		}).Error("Unexpected error: GetAccountStateFunc get account state failed")
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

	if errNo == TransferFuncSuccess && core.TransferFromContractEventRecordableAtHeight(height) {
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

	} else if core.TransferFromContractFailureEventRecordableAtHeight(height) {
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

//TransferByAddress value from to
func TransferByAddress(handler unsafe.Pointer, from *core.Address, to *core.Address, value string, gasCnt *uint64) int {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx == nil || engine.ctx.block == nil ||
		engine.ctx.state == nil || engine.ctx.tx == nil {
		logging.VLog().Fatal("Unexpected error: failed to get engine.")
	}

	*gasCnt = uint64(TransferGasBase)

	iRtn := transfer(engine, from, to, value)
	if iRtn != TransferSuccess {
		return iRtn
	}

	return TransferFuncSuccess
}

func transfer(e *V8Engine, from *core.Address, to *core.Address, val string) int {
	toAcc, err := e.ctx.state.GetOrCreateUserAccount(to.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(e.lcsHandler),
			"address": to,
			"err":     err,
		}).Error("GetAccountStateFunc get account state failed.")
		return TransferGetAccountErr
	}

	fromAcc, err := e.ctx.state.GetOrCreateUserAccount(from.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(e.lcsHandler),
			"address": fromAcc.Address().String(),
			"err":     err,
		}).Error("GetAccountStateFunc get account state failed.")
		return TransferGetAccountErr
	}

	amount, err := util.NewUint128FromString(val)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(e.lcsHandler),
			"address": to,
			"err":     err,
		}).Error("GetAmountFunc get amount failed.")
		return TransferStringToBigIntErr
	}
	logging.CLog().Infof("amount:%v", amount)
	// update balance
	if amount.Cmp(util.NewUint128()) > 0 {
		err = fromAcc.SubBalance(amount)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"handler": uint64(e.lcsHandler),
				"account": fromAcc,
				"from":    from,
				"amount":  amount,
				"err":     err,
			}).Error("TransferFunc SubBalance failed.")
			return TransferSubBalance
		}

		err = toAcc.AddBalance(amount)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"account": toAcc,
				"amount":  amount,
				"address": to,
				"err":     err,
			}).Error("failed to add balance")
			return TransferAddBalance
		}
	}
	return TransferSuccess
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
		}).Fatal("Unexpected error: failed to parse contract address")
	}

	// calculate Gas.
	*gasCnt = C.size_t(TransferGasBase)

	addr, err := core.AddressParse(C.GoString(to))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler":   uint64(uintptr(handler)),
			"toAddress": C.GoString(to),
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
		}).Fatal("GetAccountStateFunc get account state failed.")
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
			}).Debug("TransferFunc SubBalance failed.")
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
			}).Fatal("failed to add balance")
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
	*gasCnt = C.size_t(VerifyAddressGasBase)

	addr, err := core.AddressParse(C.GoString(address))
	if err != nil {
		return 0
	}
	return int(addr.Type())
}

// GetPreBlockHashFunc returns hash of the block before current tail by n
//export GetPreBlockHashFunc
func GetPreBlockHashFunc(handler unsafe.Pointer, offset C.ulonglong,
	gasCnt *C.size_t, result **C.char, exceptionInfo **C.char) int {
	*result = nil
	*exceptionInfo = nil
	n := uint64(offset)
	if n > uint64(maxBlockOffset) { //31 days
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
	*gasCnt = C.size_t(GetPreBlockHashGasBase)

	//get height
	height := engine.ctx.block.Height()
	if n >= height { // have checked it in lib js
		logging.VLog().WithFields(logrus.Fields{
			"height": height,
			"offset": n,
		}).Debug("offset is large than height")
		*exceptionInfo = C.CString("Blockchain.GetPreBlockHash(), argument[offset] is large than current height")
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
func GetPreBlockSeedFunc(handler unsafe.Pointer, offset C.ulonglong,
	gasCnt *C.size_t, result **C.char, exceptionInfo **C.char) int {
	*result = nil
	*exceptionInfo = nil

	n := uint64(offset)
	if n > uint64(maxBlockOffset) { //31 days
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
	*gasCnt = C.size_t(GetPreBlockSeedGasBase)

	//get height
	height := engine.ctx.block.Height()
	if n >= height { // have checked it in lib js
		logging.VLog().WithFields(logrus.Fields{
			"height": height,
			"offset": n,
		}).Debug("offset is large than height")
		*exceptionInfo = C.CString("Blockchain.GetPreBlockSeed(), argument[offset] is large than current height")
		return C.NVM_EXCEPTION_ERR
	}

	height -= n
	if !core.RandomAvailableAtHeight(height) {
		*exceptionInfo = C.CString("Blockchain.GetPreBlockSeed(), seed is not available at this height")
		return C.NVM_EXCEPTION_ERR
	}

	blockHash, err := wsState.GetBlockHashByHeight(height)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height":    height,
			"err":       err,
			"blockHash": blockHash,
		}).Error("Unexpected error: Failed to get block hash from wsState by height")
		return C.NVM_UNEXPECTED_ERR
	}

	bytes, err := wsState.GetBlock(blockHash)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height":    height,
			"err":       err,
			"blockHash": blockHash,
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

//getPayLoadByAddress
func getPayLoadByAddress(ws WorldState, address string) (*core.DeployPayload, error) {
	addr, err := core.AddressParse(address)
	if err != nil {
		return nil, err
	}
	contract, err := core.CheckContract(addr, ws)
	if err != nil {
		return nil, err
	}

	birthTx, err := core.GetTransaction(contract.BirthPlace(), ws)
	if err != nil {
		return nil, err
	}

	deploy, err := core.LoadDeployPayload(birthTx.Data()) // ToConfirm: move deploy payload in ctx.
	if err != nil {
		return nil, err
	}
	return deploy, nil
}

// GetContractSourceFunc get contract code by address
//export GetContractSourceFunc
func GetContractSourceFunc(handler unsafe.Pointer, address *C.char, gasCnt *C.size_t) *C.char {
	// calculate Gas.
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Error("Failed to get engine.")
		return nil
	}
	*gasCnt = C.size_t(GetContractSourceGasBase)
	ws := engine.ctx.state

	deploy, err := getPayLoadByAddress(ws, C.GoString(address))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"address": address,
			"err":     err,
		}).Error("getPayLoadByAddress err")

		return nil
	}

	return C.CString(string(deploy.Source))
}

//packErrInfoAndSetHead->packInner
func setHeadErrAndLog(e *V8Engine, index uint32, err error, result string, flag bool) string {
	//rStr := packErrInfo(errType, rerrType, rerr, format, a...)
	formatEx := InnerTransactionErrPrefix + err.Error() + InnerTransactionResult + result + InnerTransactionErrEnding
	rStr := fmt.Sprintf(formatEx, index)

	if flag == true {
		logging.CLog().Errorf(rStr)
	}
	// logging.CLog().Errorf("setHeadErrAndLog err:%v, result:%v", err, result)
	if index == 0 {
		e.innerErrMsg = result
		e.innerErr = err
	} else {
		setHeadV8ErrMsg(e.ctx.head, err, result)
	}
	// logging.CLog().Errorf("setHeadErrAndLog1111111 err:%v, result:%v", e.innerErr, result)
	return rStr
}

//setHeadV8ErrMsg set head node err info
func setHeadV8ErrMsg(handler unsafe.Pointer, err error, result string) {
	if handler == nil {
		logging.CLog().Errorf("the main node")
		return
	}
	engine := getEngineByEngineHandler(handler)
	if engine == nil {
		logging.VLog().Errorf("the handler not found the v8 engine")
		return
	}
	// logging.CLog().Errorf("setHeadErrAndLogsssssss err:%v, result:%v", err, result)
	engine.innerErr = err
	engine.innerErrMsg = result
}

// InnerContractFunc multi run contract. output[c standard]: if err return nil else return "*"
//export InnerContractFunc
func InnerContractFunc(handler unsafe.Pointer, address *C.char, funcName *C.char, v *C.char, args *C.char, gasCnt *C.size_t) *C.char {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		logging.CLog().Errorf(ErrEngineNotFound.Error())
		return nil
	}
	index := engine.ctx.index
	if engine.ctx.index >= uint32(MultiNvmMax) {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, ErrNvmNumLimit.Error(), true)
		return nil
	}
	var gasSum uint64
	gasSum = uint64(InnerContractGasBase)
	*gasCnt = C.size_t(gasSum)
	ws := engine.ctx.state

	addr, err := core.AddressParse(C.GoString(address))
	if err != nil {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
		return nil
	}
	contract, err := core.CheckContract(addr, ws)
	if err != nil {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
		return nil
	}
	logging.CLog().Infof("inner contract:%v", contract.ContractMeta())

	deploy, err := getPayLoadByAddress(ws, C.GoString(address))
	if err != nil {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
		return nil
	}

	//run
	payloadType := core.TxPayloadCallType
	callpayload, err := core.NewCallPayload(C.GoString(funcName), C.GoString(args))
	if err != nil {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
		return nil
	}
	payload, err := callpayload.ToBytes()
	if err != nil {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
		return nil
	}

	parentTx := engine.ctx.tx
	from := engine.ctx.contract.Address()
	fromAddr, err := core.AddressParseFromBytes(from)
	if err != nil {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
		return nil
	}
	//transfer
	var transferCostGas uint64
	iRet := TransferByAddress(handler, fromAddr, addr, C.GoString(v), &transferCostGas) //TODO: gas cost?
	if iRet != 0 {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, ErrInnerTransferFailed.Error(), true)
		return nil
	}
	gasSum += transferCostGas
	*gasCnt = C.size_t(gasSum)

	toValue, err := util.NewUint128FromString(C.GoString(v))
	if err != nil {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
		return nil
	}
	newTx, err := core.NewTransaction(parentTx.ChainID(), fromAddr, addr, toValue, parentTx.Nonce(), payloadType,
		payload, parentTx.GasPrice(), parentTx.GasLimit())
	if err != nil {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), false)
		logging.VLog().WithFields(logrus.Fields{
			"from":  fromAddr.String(),
			"to":    addr.String(),
			"value": C.GoString(v),
			"err":   err,
		}).Error("failed to marshal TransferFromContractEvent")
		return nil
	}
	newTx.SetHash(parentTx.Hash())
	// event address need to user
	var head unsafe.Pointer
	if engine.ctx.head == nil {
		head = unsafe.Pointer(engine.v8engine)
	} else {
		head = engine.ctx.head
	}
	//TODO: 确定world reset 是否需要
	newCtx, err := NewChildContext(engine.ctx.block, newTx, contract, engine.ctx.state, head, engine.ctx.index+1, engine.ctx.contextRand)
	if err != nil {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
		return nil
	}

	remainInstruction, remainMem := engine.GetNVMVerbResources()
	iCost := uint64(InnerContractGasBase) + transferCostGas
	if remainInstruction <= uint64(iCost) {
		logging.CLog().Errorf("remainInstruction:%v, mem:%v, err:%v", remainInstruction, remainMem, ErrInnerInsufficientGas.Error())
		setHeadErrAndLog(engine, index, ErrInsufficientGas, "null", false)
		return nil
	}
	if remainMem <= 0 {
		logging.CLog().Errorf("remainInstruction:%v, mem:%v, err:%v", remainInstruction, remainMem, ErrInnerInsufficientMem.Error())
		setHeadErrAndLog(engine, index, ErrExceedMemoryLimits, "null", false)
		return nil
	}
	remainInstruction -= uint64(InnerContractGasBase)
	remainInstruction -= uint64(transferCostGas)

	logging.CLog().Infof("begin create New V8,intance:%v, mem:%v, cost:%v", remainInstruction, remainMem, iCost)
	engineNew := NewV8Engine(newCtx)
	defer engineNew.Dispose()
	engineNew.SetExecutionLimits(remainInstruction, remainMem)

	val, err := engineNew.Call(string(deploy.Source), deploy.SourceType, C.GoString(funcName), C.GoString(args))
	gasCout := engineNew.ExecutionInstructions()
	gasSum += gasCout
	*gasCnt = C.size_t(gasSum)
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	event := &InnerContractEvent{
		From:  fromAddr.String(),
		To:    addr.String(),
		Value: toValue.String(),
		Err:   errStr,
	}

	eData, errMarshal := json.Marshal(event)
	if errMarshal != nil {
		logging.CLog().WithFields(logrus.Fields{
			"from":  fromAddr.String(),
			"to":    addr.String(),
			"value": toValue.String(),
			"err":   errMarshal.Error(),
		}).Error("failed to marshal TransferFromContractEvent")
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, errMarshal.Error(), true)

		return nil
	}
	engine.ctx.state.RecordEvent(parentTx.Hash(), &state.Event{Topic: core.TopicInnerContract, Data: string(eData)})
	if err != nil {
		if err == core.ErrInnerExecutionFailed {
			logging.CLog().Errorf("check inner err, engine index:%v", index)
			// return nil
		} else {
			errLog := setHeadErrAndLog(engine, index, err, val, false)
			logging.CLog().Errorf(errLog)
		}

		return nil
	}
	logging.CLog().Infof("end cal val:%v,gascount:%v,gasSum:%v, engine index:%v", val, gasCout, gasSum, index)
	// *gasCnt = C.size_t(gasSum)
	return C.CString(string(val))
}
