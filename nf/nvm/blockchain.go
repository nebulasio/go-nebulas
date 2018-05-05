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
	//get from
	/*fromAddr, err := core.AddressParse(from)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     from,
		}).Debug("TransferFunc parse from address failed.")
		return TransferAddressParseErr
	}*/

	fromAcc, err := engine.ctx.state.GetOrCreateUserAccount(from.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"address": from.String(),
			"err":     err,
		}).Debug("GetAccountStateFunc get account state failed.")
		return TransferGetAccountErr
	}

	/*toAddr, err := core.AddressParse(to)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     to,
		}).Debug("TransferFunc parse to address failed.")
		return TransferAddressParseErr
	}*/
	toAcc, err := engine.ctx.state.GetOrCreateUserAccount(to.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"address": to.String(),
			"err":     err,
		}).Debug("GetAccountStateFunc get account state failed.")
		return TransferGetAccountErr
	}
	amount, err := util.NewUint128FromString(value)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"address": from.String(),
			"err":     err,
		}).Debug("GetAmountFunc get amount failed.")
		return TransferStringToBigIntErr
	}
	logging.CLog().Infof("amount:%v", amount)
	// update balance
	if amount.Cmp(util.NewUint128()) > 0 {
		err = fromAcc.SubBalance(amount)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"handler": uint64(uintptr(handler)),
				"key":     from.String(),
				"err":     err,
			}).Info("TransferFunc SubBalance failed.")
			return TransferSubBalance
		}

		err = toAcc.AddBalance(amount)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"account": to,
				"amount":  amount,
				"address": to.String(),
				"err":     err,
			}).Debug("failed to add balance")
			return TransferAddBalance
		}
	}
	return TransferFuncSuccess
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

	addr, err := core.AddressParse(C.GoString(to))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(to),
		}).Debug("TransferFunc parse address failed.")
		return TransferAddressParseErr
	}

	toAcc, err := engine.ctx.state.GetOrCreateUserAccount(addr.Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"address": addr,
			"err":     err,
		}).Debug("GetAccountStateFunc get account state failed.")
		return TransferGetAccountErr
	}

	amount, err := util.NewUint128FromString(C.GoString(v))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"address": addr,
			"err":     err,
		}).Debug("GetAmountFunc get amount failed.")
		return TransferStringToBigIntErr
	}
	logging.CLog().Infof("amount:%v", amount)
	// update balance
	if amount.Cmp(util.NewUint128()) > 0 {
		err = engine.ctx.contract.SubBalance(amount)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"handler": uint64(uintptr(handler)),
				"key":     C.GoString(to),
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

		event := &TransferFromContractEvent{
			Amount: amount.String(),
			From:   cAddr.String(),
			To:     addr.String(),
		}

		eData, err := json.Marshal(event)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"from":   cAddr.String(),
				"to":     addr.String(),
				"amount": amount.String(),
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
	addr, err := core.AddressParse(C.GoString(address))
	if err != nil {
		return nil
	}
	contract, err := core.CheckContract(addr, ws)
	if err != nil {
		return nil
	}

	birthTx, err := core.GetTransaction(contract.BirthPlace(), ws)
	if err != nil {
		return nil
	}
	deploy, err := core.LoadDeployPayload(birthTx.Data()) // ToConfirm: move deploy payload in ctx.
	if err != nil {
		return nil
	}
	return C.CString(string(deploy.Source))
}
func packErrInfoAndSetHead(e *V8Engine, index uint32, errType int, rerrType *C.size_t, rerr **C.char, format string, a ...interface{}) string {
	rStr := packErrInfo(errType, rerrType, rerr, format, a...)

	formatEx := rStr + ",engine index:%v"
	rStr = fmt.Sprintf(formatEx, index)

	logging.CLog().Errorf(rStr)
	if index == 0 {
		e.multiErrMsg = rStr
	} else {
		SetHeadV8ErrMsg(e.ctx.head, rStr)
	}
	return rStr
}
func packErrInfo(errType int, rerrType *C.size_t, rerr **C.char, format string, a ...interface{}) string {
	/*var rStr string
	if a == nil {
		rStr = fmt.Sprintf(format)
	} else {
		rStr = fmt.Sprintf(format, a...)
	}*/
	rStr := fmt.Sprintf(format, a...)

	// logging.CLog().Errorf(rStr)
	*rerrType = C.size_t(errType)
	*rerr = C.CString(rStr)
	return rStr
}

//SetHeadV8ErrMsg set head node err info
func SetHeadV8ErrMsg(handler unsafe.Pointer, err string) {
	if handler == nil {
		logging.CLog().Debugf("the main node")
		return
	}
	engine := getEngineByEngineHandler(handler)
	if engine == nil {
		logging.VLog().Errorf("the handler not found the v8 engine")
		return
	}
	engine.multiErrMsg = err
}

// RunMultilevelContractSourceFunc multi run contract. output[c standard]: if err return nil else return "*"
//export RunMultilevelContractSourceFunc
func RunMultilevelContractSourceFunc(handler unsafe.Pointer, address *C.char, funcName *C.char, v *C.char, args *C.char, gasCnt *C.size_t, rerrType *C.size_t, rerr **C.char) *C.char {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		packErrInfo(MultiNotFoundEngine, rerrType, rerr, "Failed to get engine.")
		return nil
	}
	index := engine.ctx.index
	if engine.ctx.index >= uint32(MultiNvmMax) {
		//rStr := packErrInfo(MultiNvmMaxLimit, rerrType, rerr, "Failed to run nvm, becase more nvm , engine index:%v", engine.ctx.index)
		//SetHeadV8ErrMsg(engine.ctx.head, rStr)
		packErrInfoAndSetHead(engine, index, MultiNvmMaxLimit, rerrType, rerr, "Failed to run nvm, becase more nvm")
		return nil
	}
	var gasSum uint64
	gasSum = uint64(RunMultilevelContractSourceFuncCost)
	ws := engine.ctx.state

	addr, err := core.AddressParse(C.GoString(address))
	if err != nil {
		//packErrInfo(MultiNotParseAddress, rerrType, rerr, "address parse err , engine index:%v", engine.ctx.index)
		packErrInfoAndSetHead(engine, index, MultiNotParseAddress, rerrType, rerr, "address parse err")
		return nil
	}
	contract, err := core.CheckContract(addr, ws)
	if err != nil {
		//packErrInfo(MultiContractIsErr, rerrType, rerr, "check contract has err , engine index:%v", engine.ctx.index)
		packErrInfoAndSetHead(engine, index, MultiContractIsErr, rerrType, rerr, "check contract has err")
		return nil
	}

	birthTx, err := core.GetTransaction(contract.BirthPlace(), ws)
	if err != nil {
		// packErrInfo(MultiGetTransErrByBirth, rerrType, rerr, "get transaction ie err by birth , engine index:%v", engine.ctx.index)
		packErrInfoAndSetHead(engine, index, MultiGetTransErrByBirth, rerrType, rerr, "get transaction ie err by birth")
		return nil
	}
	deploy, err := core.LoadDeployPayload(birthTx.Data())
	if err != nil {
		//packErrInfo(MultiLoadDeployPayLoadErr, rerrType, rerr, "LoadDeployPayload err , engine index:%v", engine.ctx.index)
		packErrInfoAndSetHead(engine, index, MultiLoadDeployPayLoadErr, rerrType, rerr, "LoadDeployPayload err")
		return nil
	}

	//run
	payloadType := core.TxPayloadCallType
	callpayload, err := core.NewCallPayload(C.GoString(funcName), C.GoString(args))
	if err != nil {
		// packErrInfo(MultiNewCallPayLoadErr, rerrType, rerr, "core.NewCallPayload err:%v, engine index:%v", err, engine.ctx.index)
		packErrInfoAndSetHead(engine, index, MultiNewCallPayLoadErr, rerrType, rerr, "core.NewCallPayload err:%v", err)
		return nil
	}
	payload, err := callpayload.ToBytes()
	if err != nil {
		// packErrInfo(MultiPayLoadToByteErr, rerrType, rerr, "callpayload.ToBytes err:%v, engine index:%v", err, engine.ctx.index)
		packErrInfoAndSetHead(engine, index, MultiPayLoadToByteErr, rerrType, rerr, "callpayload.ToBytes err:%v", err)
		return nil
	}

	oldTx := engine.ctx.tx
	from := engine.ctx.contract.Address()
	fromAddr, err := core.AddressParseFromBytes(from)
	if err != nil {
		// packErrInfo(MultiNotParseAddressFromByte, rerrType, rerr, "core.AddressParse err:%v, engine index:%v", err, engine.ctx.index)
		packErrInfoAndSetHead(engine, index, MultiNotParseAddressFromByte, rerrType, rerr, "core.AddressParse err:%v", err)
		return nil
	}
	//transfer
	var transferCoseGas uint64
	iRet := TransferByAddress(handler, fromAddr, addr, C.GoString(v), &transferCoseGas)
	if iRet != 0 {
		// packErrInfo(MultiTransferErrByAddress, rerrType, rerr, "TransferByAddress,form:%v,to:%v,value:%v,err:%v, engine index:%v", fromAddr.String(), addr.String(), C.GoString(v), err, engine.ctx.index)
		packErrInfoAndSetHead(engine, index, MultiTransferErrByAddress, rerrType, rerr, "TransferByAddress,form:%v,to:%v,value:%v,err:%v", fromAddr.String(), addr.String(), C.GoString(v), err)
		return nil
	}
	gasSum += transferCoseGas

	toValue, err := util.NewUint128FromString(C.GoString(v))
	if err != nil {
		// packErrInfo(MultiBigNumChangeErr, rerrType, rerr, "NewUint128FromString err v:%v, engine index:%v", C.GoString(v), engine.ctx.index)
		packErrInfoAndSetHead(engine, index, MultiBigNumChangeErr, rerrType, rerr, "NewUint128FromString err, v:%v", C.GoString(v))
		return nil
	}
	newTx, err := core.NewTransaction(oldTx.ChainID(), fromAddr, addr, toValue, oldTx.Nonce(), payloadType,
		payload, oldTx.GasPrice(), oldTx.GasLimit())
	if err != nil {
		//packErrInfo(MultiNewTransactionErr, rerrType, rerr, "MultiNewTransactionErr err, from:%v, to:%v, v:%v, nonce:%v, engine index:%v",
		//	fromAddr.String(), addr.String(), C.GoString(v), oldTx.Nonce(), engine.ctx.index)
		packErrInfoAndSetHead(engine, index, MultiNewTransactionErr, rerrType, rerr, "MultiNewTransactionErr err, from:%v, to:%v, v:%v, nonce:%v",
			fromAddr.String(), addr.String(), C.GoString(v), oldTx.Nonce())
		return nil
	}
	newTx.SetHash(oldTx.Hash())
	// event address need to user
	var head unsafe.Pointer
	if engine.ctx.head == nil {
		head = unsafe.Pointer(engine.v8engine)
	} else {
		head = engine.ctx.head
	}
	newCtx, err := NewChildContext(engine.ctx.block, newTx, contract, engine.ctx.state, head, engine.ctx.index+1)
	if err != nil {
		// packErrInfo(MultiNewChildContext, rerrType, rerr, "NewContext err:%v, engine index:%v", err, engine.ctx.index)
		packErrInfoAndSetHead(engine, index, MultiNewChildContext, rerrType, rerr, "NewContext err:%v", err)
		return nil
	}

	remainInstruction, remainMem := engine.GetNVMVerbResources()
	iCost := uint64(RunMultilevelContractSourceFuncCost) + transferCoseGas
	if remainInstruction <= uint64(iCost) {
		//rStr := packErrInfo(MultiNvmSystemErr, rerrType, rerr, "engine.call system failed the gas over!!!, engine index:%d", index)
		//SetHeadV8ErrMsg(engine.ctx.head, rStr)
		packErrInfoAndSetHead(engine, index, MultiNvmSystemErr, rerrType, rerr, "engine.call system failed the gas over!!!")
		return nil
	}

	remainInstruction -= uint64(RunMultilevelContractSourceFuncCost)
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
			"err":   err,
		}).Debug("failed to marshal TransferFromContractEvent")
		packErrInfoAndSetHead(engine, index, MultiTransferRecordEventFailed, rerrType, rerr, "engine.call failed to marshal TransferFromContractEvent err:%v", err)
		//packErrInfo(MultiTransferRecordEventFailed, rerrType, rerr, "engine.call failed to marshal TransferFromContractEvent err:%v, engine index:%v", err, engine.ctx.index)
		return nil
	}
	engine.ctx.state.RecordEvent(oldTx.Hash(), &state.Event{Topic: core.TopicInnerTransferContract, Data: string(eData)})
	engineNew.Dispose()
	if err != nil {
		if err == ErrExceedMemoryLimits {
			//rStr := packErrInfo(MultiSystemMemLimit, rerrType, rerr, "engine.call mem limit err:%v, engine index:%v", err, engine.ctx.index)
			//SetHeadV8ErrMsg(engine.ctx.head, rStr)
			packErrInfoAndSetHead(engine, index, MultiSystemMemLimit, rerrType, rerr, "engine.call mem limit err:%v", err)
		} else if err == ErrInsufficientGas {
			//rStr := packErrInfo(MultiSystemInsufficientLimit, rerrType, rerr, "engine.call insuff limit err:%v, engine index:%v", err, engine.ctx.index)
			//SetHeadV8ErrMsg(engine.ctx.head, rStr)
			packErrInfoAndSetHead(engine, index, MultiSystemInsufficientLimit, rerrType, rerr, "engine.call insuff limit err:%v", err)
		} else if err == core.ErrMultiExecutionFailed {
			logging.CLog().Errorf("++++++++err:%v", err)
			packErrInfo(MultiNvmSystemErr, rerrType, rerr, "engine.call system err:%v, engine index:%d", err, index)
		} else {
			packErrInfo(MultiCallErr, rerrType, rerr, "engine.call err:%v, engine index:%v", err, index)
		}
		return nil
		/*packErrInfo(MultiNvmSystemErr, rerrType, rerr, "engine.call system err:%v, engine index:%d", err, index)
		return nil*/
	}
	logging.CLog().Infof("end cal val:%v,gascount:%v,gasSum:%v, engine index:%v", val, gasCout, gasSum, index)
	*gasCnt = C.size_t(gasSum)
	return C.CString(string(val))
}
