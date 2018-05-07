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

import "C"

import (
	"fmt"
	"unsafe"

	"encoding/json"

	"github.com/nebulasio/go-nebulas/core"
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
	*gasCnt = C.size_t(GetTxByHashFuncCost)

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
func GetAccountStateFunc(handler unsafe.Pointer, address *C.char, gasCnt *C.size_t) *C.char {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		return nil
	}

	// calculate Gas.
	*gasCnt = C.size_t(GetAccountStateFuncCost)

	addr, err := core.AddressParse(C.GoString(address))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(address),
		}).Debug("GetAccountStateFunc parse address failed.")
		return nil
	}

	acc, err := engine.ctx.state.GetOrCreateUserAccount(addr.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"address": addr,
			"err":     err,
		}).Debug("GetAccountStateFunc get account state failed.")
		return nil
	}
	state := toSerializableAccount(acc)
	json, err := json.Marshal(state)
	if err != nil {
		return nil
	}
	return C.CString(string(json))
}

//TransferByAddress value from to
func TransferByAddress(handler unsafe.Pointer, from *core.Address, to *core.Address, value string, gasCnt *uint64) int {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Error("Failed to get engine.")
		return TransferGetEngineErr
	}
	*gasCnt = uint64(TransferFuncCost)

	iRtn := transfer(engine, from.Bytes(), to.String(), value)
	if iRtn != TransferSuccess {
		return iRtn
	}

	return TransferFuncSuccess
}
func transfer(e *V8Engine, from byteutils.Hash, to string, val string) int {
	addr, err := core.AddressParse(to)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(e.lcsHandler),
			"key":     to,
		}).Debug("TransferFunc parse address failed.")
		return TransferAddressParseErr
	}

	toAcc, err := e.ctx.state.GetOrCreateUserAccount(addr.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(e.lcsHandler),
			"address": addr,
			"err":     err,
		}).Debug("GetAccountStateFunc get account state failed.")
		return TransferGetAccountErr
	}

	fromAcc, err := e.ctx.state.GetOrCreateUserAccount(from)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(e.lcsHandler),
			"address": fromAcc.Address().String(),
			"err":     err,
		}).Debug("GetAccountStateFunc get account state failed.")
		return TransferGetAccountErr
	}

	amount, err := util.NewUint128FromString(val)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(e.lcsHandler),
			"address": addr,
			"err":     err,
		}).Debug("GetAmountFunc get amount failed.")
		return TransferStringToBigIntErr
	}
	logging.CLog().Infof("amount:%v", amount)
	// update balance
	if amount.Cmp(util.NewUint128()) > 0 {
		err = e.ctx.contract.SubBalance(amount)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"handler": uint64(e.lcsHandler),
				"key":     to,
				"err":     err,
			}).Info("TransferFunc SubBalance failed.")
			return TransferSubBalance
		}

		err = toAcc.AddBalance(amount)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"account": toAcc,
				"amount":  amount,
				"address": addr,
				"err":     err,
			}).Debug("failed to add balance")
			return TransferAddBalance
		}
	}
	return TransferSuccess
}

// TransferFunc transfer vale to address
//export TransferFunc
func TransferFunc(handler unsafe.Pointer, to *C.char, v *C.char, gasCnt *C.size_t) int {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Error("Failed to get engine.")
		return TransferGetEngineErr
	}

	// calculate Gas.
	*gasCnt = C.size_t(TransferFuncCost)
	toStr := C.GoString(to)
	val := C.GoString(v)

	iRtn := transfer(engine, engine.ctx.contract.Address(), toStr, val)
	if iRtn != TransferSuccess {
		return iRtn
	}

	if engine.ctx.block.Height() >= TransferFromContractEventCompatibilityHeight {
		cAddr, err := core.AddressParseFromBytes(engine.ctx.contract.Address())
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"txhash":  engine.ctx.tx.Hash().String(),
				"address": engine.ctx.contract.Address(),
				"err":     err,
			}).Debug("failed to parse contract address")
			return TransferRecordEventFailed
		}
		//fromStr := cAddr.String()

		event := &TransferFromContractEvent{
			Amount: val,
			From:   cAddr.String(),
			To:     toStr,
		}

		eData, err := json.Marshal(event)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"from":   cAddr.String(),
				"to":     toStr,
				"amount": val,
				"err":    err,
			}).Debug("failed to marshal TransferFromContractEvent")
			return TransferRecordEventFailed
		}

		engine.ctx.state.RecordEvent(engine.ctx.tx.Hash(), &state.Event{Topic: core.TopicTransferFromContract, Data: string(eData)})
	}

	return TransferFuncSuccess
}

// VerifyAddressFunc verify address is valid
//export VerifyAddressFunc
func VerifyAddressFunc(handler unsafe.Pointer, address *C.char, gasCnt *C.size_t) int {
	// calculate Gas.
	*gasCnt = C.size_t(VerifyAddressFuncCost)

	addr, err := core.AddressParse(C.GoString(address))
	if err != nil {
		return 0
	}
	return int(addr.Type())
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
	*gasCnt = C.size_t(GetContractSourceFuncCost)
	ws := engine.ctx.state

	deploy, err := getPayLoadByAddress(ws, C.GoString(address))
	if err != nil {
		logging.CLog().Errorf("getPayLoadByAddress err, address:%v, err:%v", address, err)
		return nil
	}

	return C.CString(string(deploy.Source))
}

//packErrInfoAndSetHead->packInner
func setHeadErrAndLog(e *V8Engine, index uint32, err string, flag bool) string {
	//rStr := packErrInfo(errType, rerrType, rerr, format, a...)

	formatEx := InnerTransactionErrPrefix + err + InnerTransactionErrEnding
	rStr := fmt.Sprintf(formatEx, index)

	if flag == true {
		logging.CLog().Errorf(rStr)
	}

	if index == 0 {
		e.innerErrMsg = rStr
	} else {
		SetHeadV8ErrMsg(e.ctx.head, rStr)
	}
	return rStr
}

//SetHeadV8ErrMsg set head node err info
func SetHeadV8ErrMsg(handler unsafe.Pointer, err string) {
	if handler == nil {
		logging.CLog().Errorf("the main node")
		return
	}
	engine := getEngineByEngineHandler(handler)
	if engine == nil {
		logging.VLog().Errorf("the handler not found the v8 engine")
		return
	}
	engine.innerErrMsg = err
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
		setHeadErrAndLog(engine, index, ErrNvmNumLimit.Error(), true)
		return nil
	}
	var gasSum uint64
	gasSum = uint64(InnerContractFuncCost)
	ws := engine.ctx.state

	addr, err := core.AddressParse(C.GoString(address))
	if err != nil {
		setHeadErrAndLog(engine, index, err.Error(), true)
		return nil
	}
	contract, err := core.CheckContract(addr, ws)
	if err != nil {
		setHeadErrAndLog(engine, index, err.Error(), true)
		return nil
	}

	deploy, err := getPayLoadByAddress(ws, C.GoString(address))
	if err != nil {
		setHeadErrAndLog(engine, index, err.Error(), true)
		return nil
	}

	//run
	payloadType := core.TxPayloadCallType
	callpayload, err := core.NewCallPayload(C.GoString(funcName), C.GoString(args))
	if err != nil {
		setHeadErrAndLog(engine, index, err.Error(), true)
		return nil
	}
	payload, err := callpayload.ToBytes()
	if err != nil {
		setHeadErrAndLog(engine, index, err.Error(), true)
		return nil
	}

	parentTx := engine.ctx.tx
	from := engine.ctx.contract.Address()
	fromAddr, err := core.AddressParseFromBytes(from)
	if err != nil {
		setHeadErrAndLog(engine, index, err.Error(), true)
		return nil
	}
	//transfer
	var transferCoseGas uint64
	iRet := TransferByAddress(handler, fromAddr, addr, C.GoString(v), &transferCoseGas)
	if iRet != 0 {
		setHeadErrAndLog(engine, index, ErrInnerTransferFailed.Error(), true)
		return nil
	}
	gasSum += transferCoseGas

	toValue, err := util.NewUint128FromString(C.GoString(v))
	if err != nil {
		setHeadErrAndLog(engine, index, err.Error(), true)
		return nil
	}
	newTx, err := core.NewTransaction(parentTx.ChainID(), fromAddr, addr, toValue, parentTx.Nonce(), payloadType,
		payload, parentTx.GasPrice(), parentTx.GasLimit())
	if err != nil {
		setHeadErrAndLog(engine, index, err.Error(), false)
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
	newCtx, err := NewChildContext(engine.ctx.block, newTx, contract, engine.ctx.state, head, engine.ctx.index+1)
	if err != nil {
		setHeadErrAndLog(engine, index, err.Error(), true)
		return nil
	}

	remainInstruction, remainMem := engine.GetNVMVerbResources()
	iCost := uint64(InnerContractFuncCost) + transferCoseGas
	if remainInstruction <= uint64(iCost) {
		logging.CLog().Infof("remainInstruction:%v, mem:%v", remainInstruction, remainMem)
		setHeadErrAndLog(engine, index, ErrInnerInsufficientGas.Error(), true)
		return nil
	}
	if remainMem <= 0 {
		setHeadErrAndLog(engine, index, ErrInnerInsufficientMem.Error(), true)
		return nil
	}
	remainInstruction -= uint64(InnerContractFuncCost)
	remainInstruction -= uint64(transferCoseGas)

	logging.CLog().Infof("begin create New V8,intance:%v, mem:%v, cost:%v", remainInstruction, remainMem, iCost)
	engineNew := NewV8Engine(newCtx)
	engineNew.SetExecutionLimits(remainInstruction, remainMem)

	val, err := engineNew.Call(string(deploy.Source), deploy.SourceType, C.GoString(funcName), C.GoString(args))
	gasCout := engineNew.ExecutionInstructions()

	gasSum += gasCout
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	event := &InnerTransferContractEvent{
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
		setHeadErrAndLog(engine, index, errMarshal.Error(), true)
		engineNew.Dispose()
		return nil
	}
	engine.ctx.state.RecordEvent(parentTx.Hash(), &state.Event{Topic: core.TopicInnerTransferContract, Data: string(eData)})
	engineNew.Dispose()
	if err != nil {
		if err == core.ErrInnerExecutionFailed || err == core.ErrExecutionFailed {
			return nil
		}
		setHeadErrAndLog(engine, index, err.Error(), true)
		return nil
	}
	logging.CLog().Infof("end cal val:%v,gascount:%v,gasSum:%v, engine index:%v", val, gasCout, gasSum, index)
	*gasCnt = C.size_t(gasSum)
	return C.CString(string(val))
}
