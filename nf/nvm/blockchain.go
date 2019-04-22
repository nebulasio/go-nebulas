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
	"fmt"
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
func TransferByAddress(handler uint64, from *core.Address, to *core.Address, value *util.Uint128) int {
	engine, _ := getEngineByStorageHandler(handler)
	if engine == nil || engine.ctx == nil || engine.ctx.block == nil ||
		engine.ctx.state == nil || engine.ctx.tx == nil {
		logging.VLog().Fatal("Unexpected error: failed to get engine.")
	}

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
		err = fromAcc.SubBalance(amount) 		//TODO: add unit amount不足，超大, NaN
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
			//TODO: Failed to / Successed to / Unexpected error
			return ErrTransferAddBalance
		}
	}
	return SuccessTransfer
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
		recordTransferEvent(ErrTransferAddressParse, cAddr.String(), "", "", height, wsState, txHash)
		return ErrTransferAddressParse, gasCnt
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
			return ErrTransferGetAccount, gasCnt
		}
	}

	transferValueStr := v
	amount, err := util.NewUint128FromString(transferValueStr)
	// in old world state, accountstate create before amount check
	if core.NvmValueCheckUpdateHeight(height) {
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"handler": uint64(uintptr(handler)),
				"address": addr.String(),
				"err":     err,
				"val":     transferValueStr,
			}).Error("Failed to get amount failed.")
			recordTransferEvent(ErrTransferStringToUint128, cAddr.String(), addr.String(), transferValueStr, height, wsState, txHash)
			return ErrTransferStringToUint128, gasCnt
		}
	}
	ret := TransferByAddress(handler, cAddr, addr, amount)

	if ret != ErrTransferStringToUint128 && ret != ErrTransferSubBalance && ret != SuccessTransferFunc { // Unepected to happen, should not to be on chain
		logging.VLog().WithFields(logrus.Fields{
			"height":      engine.ctx.block.Height(),
			"txhash":      engine.ctx.tx.Hash().String(),
			"fromAddress": cAddr.String(),
			"toAddress":   addr.String(),
			"value":       v,
			"ret":         ret,
		}).Error("Unexpected error")
	}

	recordTransferEvent(ret, cAddr.String(), addr.String(), v, height, wsState, txHash)
	return ret, gasCnt
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
	if !core.RandomAvailableAtHeight(height) {
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

//getPayloadByAddress: get payload by given contract address
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
//return result(string), gasCnt(string), notnil(bool)
func GetContractSourceFunc(handler uint64, address string) (string, uint64, bool) {

	var gasCnt uint64 = GetContractSourceGasBase
	var res string = ""

	engine, _ := getEngineByStorageHandler(handler)
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Error("GetContractSourceFunc: Failed to get engine.")
		return res, gasCnt, false
	}

	ws := engine.ctx.state
	payload, err := getPayloadByAddress(ws, address)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"address": address,
			"err":     err,
		}).Error("GetContractSourceFunc: getPayLoadByAddress failed")

		return res, gasCnt, false
	}

	return string(payload.deploy.Source), gasCnt, true
}

//packErrInfoAndSetHead->packInner
func setHeadErrAndLog(e *V8Engine, index uint32, err error, result string, flag bool) string {
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
func setHeadV8ErrMsg(handler uint64, err error, result string) {
	if handler == 0 {
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
	var head uint64
	if engine.ctx.head == 0 {
		head = engine.lcsHandler
		// replace this with an unint64 value
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
func recordInnerContractEvent(err error, from string, to string, value string, wsState WorldState, txHash byteutils.Hash) {
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

// In earlier versions of inner-contract call, inconsistent logic resulted in inconsistent data on the chain, requiring adaptation.
func earlierTestnetInnerTxCompatibility(engine *V8Engine) bool {
	testnetUpdateHeight := uint64(845750)
	return engine.ctx.tx.ChainID() == core.TestNetID && engine.ctx.block.Height() < testnetUpdateHeight
}

//Return: engineNew(*V8Engine), nvmConfigNew(*NVMConfig), gasCnt(uint64), err(error)
func InnerContractFunc(handler uint64, address string, funcName string, 
	innerTxValueStr string, args string) (*V8Engine, *core.NVMConfig, uint64, error) {

	var gasCnt uint64 = uint64(InnerContractGasBase)
	var engineNew *V8Engine = nil
	var nvmConfigNew *core.NVMConfig = nil

	engine, _ := getEngineByStorageHandler(handler)
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Errorf(ErrEngineNotFound.Error())
		return engineNew, nvmConfigNew, gasCnt, ErrEngineNotFound
	}

	index := engine.ctx.index
	if engine.ctx.index >= uint32(MaxInnerContractLevel) {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, ErrMaxInnerContractLevelLimit.Error(), true)
		return engineNew, nvmConfigNew, gasCnt, ErrMaxInnerContractLevelLimit
	}

	ws := engine.ctx.state
	addr, err := core.AddressParse(address)
	if err != nil {
		setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
		return engineNew, nvmConfigNew, gasCnt, err
	}

	var (
		newCtx   *Context
		deploy   *core.DeployPayload
		fromAddr *core.Address
		toValue  *util.Uint128
	)

	parentTx := engine.ctx.tx
	if earlierTestnetInnerTxCompatibility(engine) {

		contract, err := core.CheckContract(addr, ws)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return engineNew, nvmConfigNew, gasCnt, err
		}
		logging.VLog().Infof("inner contract:%v", contract.ContractMeta()) //FIXME: ver limit

		payload, err := getPayloadByAddress(ws, address)					//TODO: the payload/source can be read from the cache
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return engineNew, nvmConfigNew, gasCnt, err
		}

		deploy = payload.deploy
		//run
		payloadType := core.TxPayloadCallType
		callpayload, err := core.NewCallPayload(funcName, args)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return engineNew, nvmConfigNew, gasCnt, err
		}
		newPayloadHex, err := callpayload.ToBytes()
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return engineNew, nvmConfigNew, gasCnt, err
		}

		from := engine.ctx.contract.Address()
		fromAddr, err = core.AddressParseFromBytes(from)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return engineNew, nvmConfigNew, gasCnt, err
		}
		//transfer
		// var transferCostGas uint64
		toValue, err = util.NewUint128FromString(innerTxValueStr)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return engineNew, nvmConfigNew, gasCnt, err
		}
		iRet := TransferByAddress(handler, fromAddr, addr, toValue)
		if iRet != 0 {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, ErrInnerTransferFailed.Error(), true)
			return engineNew, nvmConfigNew, gasCnt, err
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
			return engineNew, nvmConfigNew, gasCnt, err
		}

		// event address need to user
		var head uint64
		if engine.ctx.head == 0 {
			head = handler
		} else {
			head = engine.ctx.head
		}
		newCtx, err = NewInnerContext(engine.ctx.block, newTx, contract, engine.ctx.state, head, engine.ctx.index+1, engine.ctx.contextRand)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return engineNew, nvmConfigNew, gasCnt, err
		}


	} else {

		payload, err := getPayloadByAddress(ws, address)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return engineNew, nvmConfigNew, gasCnt, err
		}
		deploy = payload.deploy

		from := engine.ctx.contract.Address()
		fromAddr, err = core.AddressParseFromBytes(from)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return engineNew, nvmConfigNew, gasCnt, err
		}
		//transfer
		toValue, err = util.NewUint128FromString(innerTxValueStr)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return engineNew, nvmConfigNew, gasCnt, err
		}
		iRet := TransferByAddress(handler, fromAddr, addr, toValue)
		if iRet != 0 {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, ErrInnerTransferFailed.Error(), true)
			return engineNew, nvmConfigNew, gasCnt, ErrInnerTransferFailed
		}

		newCtx, err = createInnerContext(engine, fromAddr, addr, toValue, funcName, args)
		if err != nil {
			setHeadErrAndLog(engine, index, core.ErrExecutionFailed, err.Error(), true)
			return engineNew, nvmConfigNew, gasCnt, err
		}

	}

	// The function should be moved into V8 process
	engineNew = NewV8Engine(newCtx)

	//engineNew.SetExecutionLimits(remainInstruction, remainMem)

	nvmConf := &core.NVMConfig{
		PayloadSource: deploy.Source,
		PayloadSourceType: deploy.SourceType,
		FunctionName: funcName,
		ContractArgs: args,
		ListenAddr: engine.serverListenAddr,
		ChainID: engine.chainID,
	}

	return engineNew, nvmConf, gasCnt, nil
}

// GetLatestNebulasRankFunc returns nebulas rank value of given account address
// Return: result_code(int), result(string), exceptionInfo(string), gasCnt(uint64), notNil(bool)
func GetLatestNebulasRankFunc(handler uint64, address string) (int, string, string, uint64, bool) {

	var result string = ""
	var exceptionInfo string = ""
	var gasCnt uint64 = 0

	engine, _ := getEngineByStorageHandler(handler)
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Error("Unexpected error: failed to get engine")
		return NVM_UNEXPECTED_ERR, result, exceptionInfo, gasCnt, false
	}

	addr, err := core.AddressParse(address)
	if err != nil {
		exceptionInfo = "Address is invalid"
		return NVM_EXCEPTION_ERR, result, exceptionInfo, gasCnt, false
	}

	gasCnt = uint64(GetLatestNebulasRankGasBase)

	data, err := engine.ctx.block.NR().GetNRListByHeight(engine.ctx.block.Height())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": engine.ctx.block.Height(),
			"addr":   addr,
			"err":    err,
		}).Debug("Failed to get nr list")
		exceptionInfo = "Failed to get nr list"
		return NVM_EXCEPTION_ERR, result, exceptionInfo, gasCnt, false
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
		exceptionInfo = "Failed to find nr value"
		return NVM_EXCEPTION_ERR, result, exceptionInfo, gasCnt, false
	}

	result = nr.Score
	return NVM_SUCCESS, result, exceptionInfo, gasCnt, true
}

// GetLatestNebulasRankSummaryFunc returns nebulas rank summary info.
// return: result_code(int), result(string), exceptionInfo(string), gasCnt(uint64), notNil(bool)
func GetLatestNebulasRankSummaryFunc(handler uint64) (int, string, string, uint64, bool){

	var result string = ""
	var exceptionInfo string = ""
	var gasCnt uint64 = 0

	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Error("Unexpected error: failed to get engine")
		return NVM_UNEXPECTED_ERR, result, exceptionInfo, gasCnt, false
	}

	gasCnt = GetLatestNebulasRankSummaryGasBase

	data, err := engine.ctx.block.NR().GetNRSummary(engine.ctx.block.Height())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": engine.ctx.block.Height(),
			"err":    err,
		}).Debug("Failed to get nr summary info")
		exceptionInfo = "Failed to get nr summary"
		return NVM_EXCEPTION_ERR, result, exceptionInfo, gasCnt, false
	}

	bytes, err := data.ToBytes()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": engine.ctx.block.Height(),
			"err":    err,
		}).Debug("Failed to serialize nr summary info")
		exceptionInfo = "Failed to serialize nr summary"
		return NVM_EXCEPTION_ERR, result, exceptionInfo, gasCnt, false
	}

	result = string(bytes)
	return NVM_SUCCESS, result, exceptionInfo, gasCnt, true
}