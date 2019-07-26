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

	"github.com/nebulasio/go-nebulas/nr"

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

func recordTransferEvent(errNo int, from string, to string, value string,
	height uint64, wsState WorldState, txHash byteutils.Hash) {

	if errNo == SuccessTransferFunc && core.TransferFromContractEventRecordableAtHeight(height) {
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
		case SuccessTransferFunc:
			errMsg = ""
		case ErrTransferAddressParse:
			errMsg = "failed to parse to address"
		case ErrTransferStringToUint128:
			errMsg = "failed to parse transfer amount"
			if !core.TransferFromContractFailureEventRecordableAtHeight2(height) {
				value = ""
			}
		case ErrTransferSubBalance:
			errMsg = "failed to sub balace from contract address"
		default:
			logging.VLog().WithFields(logrus.Fields{
				"from":   from,
				"to":     to,
				"amount": value,
				"errNo":  errNo,
			}).Error("unexpected error to handle")
			return
		}

		status := uint8(1)
		if errNo != SuccessTransferFunc {
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
func TransferByAddress(handler unsafe.Pointer, from *core.Address, to *core.Address, value *util.Uint128) int {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx == nil || engine.ctx.block == nil ||
		engine.ctx.state == nil || engine.ctx.tx == nil {
		logging.VLog().Fatal("Unexpected error: failed to get engine.")
	}

	// *gasCnt = uint64(TransferGasBase)
	iRtn := transfer(engine, from, to, value)
	if iRtn != SuccessTransfer {
		return iRtn
	}

	return SuccessTransferFunc
}

func transfer(e *V8Engine, from *core.Address, to *core.Address, amount *util.Uint128) int {
	toAcc, err := e.ctx.state.GetOrCreateUserAccount(to.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(e.lcsHandler),
			"address": to,
			"err":     err,
		}).Error("Failed to get to account state")
		return ErrTransferGetAccount
	}

	fromAcc, err := e.ctx.state.GetOrCreateUserAccount(from.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(e.lcsHandler),
			"address": from,
			"err":     err,
		}).Error("Failed to get from account state")
		return ErrTransferGetAccount
	}
	// TestNet sync adjust
	if amount == nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(e.lcsHandler),
			"address": from,
			"err":     err,
		}).Error("Failed to get amount failed.")
		return ErrTransferStringToUint128
	}

	// update balance
	if amount.Cmp(util.NewUint128()) > 0 {
		err = fromAcc.SubBalance(amount) //TODO: add unit amount不足，超大, NaN
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"handler": uint64(e.lcsHandler),
				"account": fromAcc,
				"from":    from,
				"amount":  amount,
				"err":     err,
			}).Error("Failed to sub balance")
			return ErrTransferSubBalance
		}

		err = toAcc.AddBalance(amount)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"account": toAcc,
				"amount":  amount,
				"address": to,
				"err":     err,
			}).Error("Failed to add balance")
			//TODO: Failed to / Successed to	/ Unexpected error
			return ErrTransferAddBalance
		}
	}
	return SuccessTransfer
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
		recordTransferEvent(ErrTransferAddressParse, cAddr.String(), "", "", height, wsState, txHash)
		return ErrTransferAddressParse
	}

	// in old world state, toAcc accountstate create before amount check
	if !core.TransferFromContractFailureEventRecordableAtHeight2(height) {
		_, err := engine.ctx.state.GetOrCreateUserAccount(addr.Bytes())
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"handler": uint64(uintptr(handler)),
				"address": addr,
				"err":     err,
			}).Error("GetAccountStateFunc get account state failed.")
			recordTransferEvent(ErrTransferGetAccount, cAddr.String(), addr.String(), "", height, wsState, txHash)
			return ErrTransferGetAccount
		}
	}

	transferValueStr := C.GoString(v)
	recordValue := transferValueStr
	amount, err := util.NewUint128FromString(transferValueStr)
	if core.NvmValueCheckUpdateHeight(height) {
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"handler": uint64(uintptr(handler)),
				"address": addr.String(),
				"err":     err,
				"val":     transferValueStr,
			}).Error("Failed to get amount failed.")
			recordTransferEvent(ErrTransferStringToUint128, cAddr.String(), addr.String(), transferValueStr, height, wsState, txHash)
			return ErrTransferStringToUint128
		}
	} else {
		// in old version, record value is empty when it cannot convert to uint128
		if err != nil {
			recordValue = ""
		}
	}
	ret := TransferByAddress(handler, cAddr, addr, amount)

	if ret != ErrTransferStringToUint128 && ret != ErrTransferSubBalance && ret != SuccessTransferFunc { // Unepected to happen, should not to be on chain
		logging.VLog().WithFields(logrus.Fields{
			"height":      engine.ctx.block.Height(),
			"txhash":      engine.ctx.tx.Hash().String(),
			"fromAddress": cAddr.String(),
			"toAddress":   addr.String(),
			"value":       transferValueStr,
			"ret":         ret,
		}).Error("Unexpected error")
	}

	recordTransferEvent(ret, cAddr.String(), addr.String(), recordValue, height, wsState, txHash)
	return ret
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

//getPayloadByAddress
func getPayloadByAddress(ws WorldState, address string) (*Payload, error) {
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
	return &Payload{deploy, contract}, nil
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

	payload, err := getPayloadByAddress(ws, C.GoString(address))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"address": address,
			"err":     err,
		}).Error("getPayLoadByAddress err")

		return nil
	}

	return C.CString(string(payload.deploy.Source))
}

//packErrInfoAndSetHead->packInner
func setHeadErrAndLog(e *V8Engine, index uint32, err error, result string, flag bool) string {
	//rStr := packErrInfo(errType, rerrType, rerr, format, a...)
	formatEx := InnerTransactionErrPrefix + err.Error() + InnerTransactionResult + result + InnerTransactionErrEnding
	rStr := fmt.Sprintf(formatEx, index)

	if flag == true {
		logging.VLog().Errorf(rStr)
	}
	if index == 0 {
		e.innerErrMsg = result
		e.innerErr = err
	} else {
		setHeadV8ErrMsg(e.ctx.head, err, result)
	}
	return rStr
}

//setHeadV8ErrMsg set head node err info
func setHeadV8ErrMsg(handler unsafe.Pointer, err error, result string) {
	if handler == nil {
		logging.VLog().Errorf("invalid handler is nil")
		return
	}
	engine := getEngineByEngineHandler(handler)
	if engine == nil {
		logging.VLog().Errorf("not found the v8 engine")
		return
	}
	engine.innerErr = err
	engine.innerErrMsg = result
}

//createInnerContext is private func only in InnerContractFunc
func createInnerContext(engine *V8Engine, fromAddr *core.Address, toAddr *core.Address, value *util.Uint128, funcName string, args string) (innerCtx *Context, err error) {
	ws := engine.ctx.state
	contract, err := core.CheckContract(toAddr, ws)
	if err != nil {
		return nil, err
	}
	logging.VLog().Infof("inner contract:%v", contract.ContractMeta()) //FIXME: ver limit
	payloadType := core.TxPayloadCallType
	callpayload, err := core.NewCallPayload(funcName, args)
	if err != nil {
		return nil, err
	}
	newPayloadHex, err := callpayload.ToBytes()
	if err != nil {
		return nil, err
	}

	// test sync adaptation
	// In Testnet, before nbre available height, inner to address is fromAddr that is a bug.
	innerToAddr := toAddr
	if engine.ctx.tx.ChainID() == core.TestNetID &&
		!core.NbreAvailableHeight(engine.ctx.block.Height()) {
		innerToAddr = fromAddr
	}
	parentTx := engine.ctx.tx
	newTx, err := parentTx.NewInnerTransaction(parentTx.To(), innerToAddr, value, payloadType, newPayloadHex)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"from":  fromAddr.String(),
			"to":    toAddr.String(),
			"value": value.String(),
			"err":   err,
		}).Error("failed to create new tx")
		return nil, err
	}
	var head unsafe.Pointer
	if engine.ctx.head == nil {
		head = unsafe.Pointer(engine.v8engine)
	} else {
		head = engine.ctx.head
	}
	newCtx, err := NewInnerContext(engine.ctx.block, newTx, contract, engine.ctx.state, head, engine.ctx.index+1, engine.ctx.contextRand)
	if err != nil {
		return nil, err
	}
	return newCtx, nil
}

//recordInnerContractEvent private func only in InnerContractFunc
func recordInnerContractEvent(e *V8Engine, err error, from string, to string, value string, innerFunc string, innerArgs string, wsState WorldState, txHash byteutils.Hash) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	event := &InnerContractEvent{
		From:  from,
		To:    to,
		Value: value,
		Err:   errStr,
	}

	if core.NbreSplitAtHeight(e.ctx.block.Height()) {
		event.Function = innerFunc
		event.Args = innerArgs
	}

	eData, errMarshal := json.Marshal(event)
	if errMarshal != nil {
		logging.VLog().WithFields(logrus.Fields{
			"from":  from,
			"to":    to,
			"value": value,
			"err":   errStr,
		}).Fatal("failed to marshal TransferFromContractEvent")
	}
	wsState.RecordEvent(txHash, &state.Event{Topic: core.TopicInnerContract, Data: string(eData)})

}

// In earlier versions of inter-contract invocation, inconsistent logic resulted in inconsistent data on the chain, requiring adaptation.
func earlierTestnetInnerTxCompatibility(engine *V8Engine) bool {
	testnetUpdateHeight := uint64(845750)
	return engine.ctx.tx.ChainID() == core.TestNetID && engine.ctx.block.Height() < testnetUpdateHeight
}

// InnerContractFunc multi run contract. output[c standard]: if err return nil else return "*"
//export InnerContractFunc
func InnerContractFunc(handler unsafe.Pointer, address *C.char, funcName *C.char, v *C.char, args *C.char, gasCnt *C.size_t) *C.char {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Errorf(ErrEngineNotFound.Error())
		return nil
	}
	index := engine.ctx.index
	if engine.ctx.index >= uint32(MaxInnerContractLevel) {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, ErrMaxInnerContractLevelLimit.Error(), true)
		return nil
	}
	gasSum := uint64(InnerContractGasBase)
	*gasCnt = C.size_t(gasSum)
	ws := engine.ctx.state

	addr, err := core.AddressParse(C.GoString(address))
	if err != nil {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
		return nil
	}

	var (
		newCtx   *Context
		deploy   *core.DeployPayload
		fromAddr *core.Address
		toValue  *util.Uint128
	)

	parentTx := engine.ctx.tx
	innerTxValueStr := C.GoString(v)
	if earlierTestnetInnerTxCompatibility(engine) {
		contract, err := core.CheckContract(addr, ws)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return nil
		}
		logging.VLog().Infof("inner contract:%v", contract.ContractMeta()) //FIXME: ver limit

		payload, err := getPayloadByAddress(ws, C.GoString(address))
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return nil
		}
		deploy = payload.deploy
		//run
		payloadType := core.TxPayloadCallType
		callpayload, err := core.NewCallPayload(C.GoString(funcName), C.GoString(args))
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return nil
		}
		newPayloadHex, err := callpayload.ToBytes()
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return nil
		}

		from := engine.ctx.contract.Address()
		fromAddr, err = core.AddressParseFromBytes(from)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return nil
		}
		//transfer
		// var transferCostGas uint64
		toValue, err = util.NewUint128FromString(innerTxValueStr)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return nil
		}
		iRet := TransferByAddress(handler, fromAddr, addr, toValue)
		if iRet != 0 {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, ErrInnerTransferFailed.Error(), true)
			return nil
		}

		newTx, err := parentTx.NewInnerTransaction(parentTx.To(), addr, toValue, payloadType, newPayloadHex)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), false)
			logging.VLog().WithFields(logrus.Fields{
				"from":  fromAddr.String(),
				"to":    addr.String(),
				"value": innerTxValueStr,
				"err":   err,
			}).Error("failed to create new tx")
			return nil
		}
		// event address need to user
		var head unsafe.Pointer
		if engine.ctx.head == nil {
			head = unsafe.Pointer(engine.v8engine)
		} else {
			head = engine.ctx.head
		}
		newCtx, err = NewInnerContext(engine.ctx.block, newTx, contract, engine.ctx.state, head, engine.ctx.index+1, engine.ctx.contextRand)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return nil
		}
	} else {
		payload, err := getPayloadByAddress(ws, C.GoString(address))
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return nil
		}
		deploy = payload.deploy

		from := engine.ctx.contract.Address()
		fromAddr, err = core.AddressParseFromBytes(from)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return nil
		}
		//transfer
		toValue, err = util.NewUint128FromString(innerTxValueStr)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return nil
		}
		iRet := TransferByAddress(handler, fromAddr, addr, toValue)
		if iRet != 0 {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, ErrInnerTransferFailed.Error(), true)
			return nil
		}

		newCtx, err = createInnerContext(engine, fromAddr, addr, toValue, C.GoString(funcName), C.GoString(args))
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return nil
		}
	}

	remainInstruction, remainMem := engine.GetNVMLeftResources()
	if remainInstruction <= uint64(InnerContractGasBase) {
		logging.VLog().WithFields(logrus.Fields{
			"remainInstruction": remainInstruction,
			"mem":               remainMem,
			"err":               ErrInnerInsufficientGas.Error(),
		}).Error("failed to prepare create nvm")
		setHeadErrAndLog(engine, index, ErrInsufficientGas, "null", false)
		return nil
	} else {
		remainInstruction -= InnerContractGasBase
	}
	if remainMem <= 0 {
		logging.VLog().WithFields(logrus.Fields{
			"remainInstruction": remainInstruction,
			"mem":               remainMem,
			"err":               ErrInnerInsufficientMem.Error(),
		}).Error("failed to prepare create nvm")
		setHeadErrAndLog(engine, index, ErrExceedMemoryLimits, "null", false)
		return nil
	}

	logging.VLog().Debugf("begin create New V8,intance:%v, mem:%v", remainInstruction, remainMem)
	engineNew := NewV8Engine(newCtx)
	defer engineNew.Dispose()
	engineNew.SetExecutionLimits(remainInstruction, remainMem)

	innerFunc := C.GoString(funcName)
	innerArgs := C.GoString(args)
	val, err := engineNew.Call(string(deploy.Source), deploy.SourceType, innerFunc, innerArgs)
	gasCout := engineNew.ExecutionInstructions()
	gasSum += gasCout
	*gasCnt = C.size_t(gasSum)
	recordInnerContractEvent(engine, err, fromAddr.String(), addr.String(), toValue.String(), innerFunc, innerArgs, ws, parentTx.Hash())
	if err != nil {
		if err == core.ErrInnerExecutionFailed {
			logging.VLog().Errorf("check inner err, engine index:%v", index)
		} else {
			errLog := setHeadErrAndLog(engine, index, err, val, false)
			logging.VLog().Errorf(errLog)
		}
		return nil
	}

	logging.VLog().Infof("end cal val:%v,gascount:%v,gasSum:%v, engine index:%v", val, gasCout, gasSum, index)
	return C.CString(string(val))
}

// GetLatestNebulasRankFunc returns nebulas rank value of given account address
//export GetLatestNebulasRankFunc
func GetLatestNebulasRankFunc(handler unsafe.Pointer, address *C.char,
	gasCnt *C.size_t, result **C.char, exceptionInfo **C.char) int {

	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Error("Unexpected error: failed to get engine")
		return C.NVM_UNEXPECTED_ERR
	}

	*result = nil
	*exceptionInfo = nil
	*gasCnt = C.size_t(GetLatestNebulasRankGasBase)

	addr, err := core.AddressParse(C.GoString(address))
	if err != nil {
		*exceptionInfo = C.CString("Address is invalid")
		return C.NVM_EXCEPTION_ERR
	}

	data, err := engine.ctx.block.NR().GetNRListByHeight(engine.ctx.block.Height())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": engine.ctx.block.Height(),
			"addr":   addr,
			"err":    err,
		}).Debug("Failed to get nr list")
		*exceptionInfo = C.CString("Failed to get nr list")
		return C.NVM_EXCEPTION_ERR
	}

	nrData := data.(*nr.NRData)
	nr, err := func(list *nr.NRData, addr *core.Address) (*nr.NRItem, error) {
		for _, nr := range nrData.Nrs {
			if nr.Address == addr.String() {
				return nr, nil
			}
		}
		return nil, nr.ErrNRNotFound
	}(nrData, addr)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": engine.ctx.block.Height(),
			"addr":   addr,
			"err":    err,
		}).Debug("Failed to find nr value")
		*exceptionInfo = C.CString("Failed to find nr value")
		return C.NVM_EXCEPTION_ERR
	}

	*result = C.CString(nr.Score)
	return C.NVM_SUCCESS
}

// GetLatestNebulasRankSummaryFunc returns nebulas rank summary info.
//export GetLatestNebulasRankSummaryFunc
func GetLatestNebulasRankSummaryFunc(handler unsafe.Pointer,
	gasCnt *C.size_t, result **C.char, exceptionInfo **C.char) int {

	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Error("Unexpected error: failed to get engine")
		return C.NVM_UNEXPECTED_ERR
	}

	*result = nil
	*exceptionInfo = nil
	*gasCnt = C.size_t(GetLatestNebulasRankSummaryGasBase)

	data, err := engine.ctx.block.NR().GetNRSummary(engine.ctx.block.Height())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": engine.ctx.block.Height(),
			"err":    err,
		}).Debug("Failed to get nr summary info")
		*exceptionInfo = C.CString("Failed to get nr summary")
		return C.NVM_EXCEPTION_ERR
	}

	bytes, err := data.ToBytes()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": engine.ctx.block.Height(),
			"err":    err,
		}).Debug("Failed to serialize nr summary info")
		*exceptionInfo = C.CString("Failed to serialize nr summary")
		return C.NVM_EXCEPTION_ERR
	}
	*result = C.CString(string(bytes))
	return C.NVM_SUCCESS
}
