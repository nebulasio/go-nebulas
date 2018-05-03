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
func packErrInfo(errType int, rerrType *C.size_t, rerr **C.char, format string, a ...interface{}) string {
	var rStr string
	if a == nil {
		rStr = fmt.Sprintf(format)
	} else {
		rStr = fmt.Sprintf(format, a)
	}

	logging.CLog().Errorf(rStr)
	*rerrType = C.size_t(errType)
	*rerr = C.CString(rStr)
	return rStr
}

// RunMultilevelContractSourceFunc multi run contract. output[c standard]: if err return nil else return "*"
//export RunMultilevelContractSourceFunc
func RunMultilevelContractSourceFunc(handler unsafe.Pointer, address *C.char, funcName *C.char, v *C.char, args *C.char, gasCnt *C.size_t, rerrType *C.size_t, rerr **C.char) *C.char {
	/*logging.VLog().Error("Failed to get engine.")
	*rerrType = MultiNotFoundEngine
	a := packV8Err(MultiNotFoundEngine, "Failed to get engine.", 0)
	logging.CLog().Errorf("err:%v", a)
	//*rerr = C.CString(packV8Err(MultiNotFoundEngine, "Failed to get engine.", 0))
	*rerr = C.CString("Failed to get engine.")
	return nil*/
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		/*rStr := fmt.Sprintf("Failed to get engine.")
		logging.VLog().Errorf(rStr)
		*rerrType = MultiNotFoundEngine
		*rerr = C.CString(rStr)*/
		packErrInfo(MultiNotFoundEngine, rerrType, rerr, "Failed to get engine.")
		//logging.VLog().Error("Failed to get engine.")
		//*rerr = C.CString(packV8Err(MultiNotFoundEngine, "Failed to get engine.", engine.ctx.index))
		return nil
	}
	if engine.ctx.index >= uint32(MultiNvmMax) {
		/*rStr := fmt.Sprintf("Failed to run nvm, becase more nvm , engine index:%v", engine.ctx.index)
		logging.VLog().Errorf(rStr)
		*rerrType = MultiNvmMaxLimit
		*rerr = C.CString(rStr)*/
		packErrInfo(MultiNvmMaxLimit, rerrType, rerr, "Failed to run nvm, becase more nvm , ")
		//logging.VLog().Errorf(rStr)
		//packErrInfo(MultiNotFoundEngine, rerrType, rerr, "Failed to run nvm, becase more nvm , engine index:%v", engine.ctx.index)
		//logging.VLog().Errorf("Failed to run nvm, becase more nvm ,current nvm:%v", engine.ctx.index)
		//*rerr = C.CString(packV8Err(MultiNvmMaxLimit, "Failed to run nvm, becase more nvm", engine.ctx.index))
		return nil
	}
	var gasSum uint64
	//*gasCnt = C.size_t(RunMultilevelContractSourceFuncCost)
	gasSum = uint64(RunMultilevelContractSourceFuncCost)
	ws := engine.ctx.state

	addr, err := core.AddressParse(C.GoString(address))
	if err != nil {
		/*rStr := fmt.Sprintf("address parse err , engine index:%v", engine.ctx.index)
		logging.VLog().Errorf(rStr)
		*rerrType = MultiNotParseAddress
		*rerr = C.CString(rStr)*/
		packErrInfo(MultiNotParseAddress, rerrType, rerr, "address parse err , engine index:%v", engine.ctx.index)
		//*rerr = C.CString(packV8Err(MultiNotParseAddress, "address parse err", engine.ctx.index))
		return nil
	}
	contract, err := core.CheckContract(addr, ws)
	if err != nil {
		/*rStr := fmt.Sprintf("check contract has err , engine index:%v", engine.ctx.index)
		logging.VLog().Errorf(rStr)
		*rerrType = MultiContractIsErr
		*rerr = C.CString(rStr)*/
		packErrInfo(MultiContractIsErr, rerrType, rerr, "check contract has err , engine index:%v", engine.ctx.index)
		// *rerr = C.CString(packV8Err(MultiContractIsErr, "check contract has err", engine.ctx.index))
		return nil
	}

	birthTx, err := core.GetTransaction(contract.BirthPlace(), ws)
	if err != nil {
		/*rStr := fmt.Sprintf("get transaction ie err by birth , engine index:%v", engine.ctx.index)
		logging.VLog().Errorf(rStr)
		*rerrType = MultiGetTransErrByBirth
		*rerr = C.CString(rStr)*/
		packErrInfo(MultiGetTransErrByBirth, rerrType, rerr, "get transaction ie err by birth , engine index:%v", engine.ctx.index)
		//*rerr = C.CString(packV8Err(MultiGetTransErrByBirth, "get transaction ie err by birth", engine.ctx.index))
		return nil
	}
	deploy, err := core.LoadDeployPayload(birthTx.Data())
	if err != nil {
		/*rStr := fmt.Sprintf("LoadDeployPayload err , engine index:%v", engine.ctx.index)
		logging.VLog().Errorf(rStr)
		*rerrType = MultiLoadDeployPayLoadErr
		*rerr = C.CString(rStr)*/
		packErrInfo(MultiLoadDeployPayLoadErr, rerrType, rerr, "LoadDeployPayload err , engine index:%v", engine.ctx.index)
		// *rerr = C.CString(packV8Err(MultiLoadDeployPayLoadErr, "LoadDeployPayload err", engine.ctx.index))
		return nil
	}

	//run
	payloadType := core.TxPayloadCallType
	callpayload, err := core.NewCallPayload(C.GoString(funcName), C.GoString(args))
	if err != nil {
		/*rStr := fmt.Sprintf("core.NewCallPayload err:%v", err)
		logging.VLog().Errorf(rStr)
		*rerrType = MultiNewCallPayLoadErr
		*rerr = C.CString(rStr)*/
		packErrInfo(MultiNewCallPayLoadErr, rerrType, rerr, "core.NewCallPayload err:%v, engine index:%v", err, engine.ctx.index)
		// *rerr = C.CString(packV8Err(MultiNewCallPayLoadErr, rStr, engine.ctx.index))
		return nil
	}
	payload, err := callpayload.ToBytes()
	if err != nil {
		/*rStr := fmt.Sprintf("callpayload.ToBytes err:%v", err)
		logging.VLog().Errorf(rStr)
		*rerrType = MultiPayLoadToByteErr*/
		packErrInfo(MultiPayLoadToByteErr, rerrType, rerr, "callpayload.ToBytes err:%v, engine index:%v", err, engine.ctx.index)
		// *rerr = C.CString(packV8Err(MultiPayLoadToByteErr, rStr, engine.ctx.index))
		//*rerr = C.CString(rStr)
		return nil
	}

	oldTx := engine.ctx.tx
	// zeroVal := util.NewUint128()
	from := engine.ctx.contract.Address()
	fromAddr, err := core.AddressParseFromBytes(from)
	if err != nil {
		/*rStr := fmt.Sprintf("core.AddressParse err:%v", err)
		logging.CLog().Errorf(rStr)
		*rerrType = MultiNotParseAddressFromByte*/
		packErrInfo(MultiNotParseAddressFromByte, rerrType, rerr, "core.AddressParse err:%v, engine index:%v", err, engine.ctx.index)
		// *rerr = C.CString(packV8Err(MultiNotParseAddressFromByte, rStr, engine.ctx.index))
		//*rerr = C.CString(rStr)
		return nil
	}
	//transfer
	var transferCoseGas uint64
	//var iRet int
	iRet := TransferByAddress(handler, fromAddr, addr, C.GoString(v), &transferCoseGas)
	if iRet != 0 {
		/*rStr := fmt.Sprintf("TransferByAddress,form:%v,to:%v,value:%v", fromAddr.String(), addr.String(), C.GoString(v))
		logging.CLog().Errorf(rStr)
		*rerrType = MultiTransferErrByAddress
		*rerr = C.CString(rStr)*/
		packErrInfo(MultiTransferErrByAddress, rerrType, rerr, "TransferByAddress,form:%v,to:%v,value:%v,engine index:%v", fromAddr.String(), addr.String(), C.GoString(v), err, engine.ctx.index)
		// *rerr = C.CString(packV8Err(MultiTransferErrByAddress, rStr, engine.ctx.index))
		return nil
	}
	// logging.CLog().Errorf("end TransferFunc:form:%v, to:%v, v:%v,transferCoseGas:%v", engine.ctx.tx.From().Bytes(), addr.Bytes(), C.GoString(v), transferCoseGas)
	toValue, err := util.NewUint128FromString(C.GoString(v))
	if err != nil {
		/*rStr := fmt.Sprintf("NewUint128FromString err v:%v", C.GoString(v))
		logging.CLog().Errorf(rStr)
		*rerrType = MultiBigNumChangeErr
		*rerr = C.CString(rStr)*/
		packErrInfo(MultiBigNumChangeErr, rerrType, rerr, "NewUint128FromString err v:%v, engine index:%v", C.GoString(v), engine.ctx.index)
		//*rerr = C.CString(packV8Err(MultiBigNumChangeErr, rStr, engine.ctx.index))
		return nil
	}
	newTx, err := core.NewTransaction(oldTx.ChainID(), fromAddr, addr, toValue, oldTx.Nonce(), payloadType,
		payload, oldTx.GasPrice(), oldTx.GasLimit())
	if err != nil {
		/*rStr := fmt.Sprintf("MultiNewTransactionErr err, from:%v, to:%v, v:%v, nonce:%v",
			fromAddr.String(), addr.String(), C.GoString(v), oldTx.Nonce())
		logging.CLog().Errorf(rStr)
		*rerrType = MultiNewTransactionErr
		*rerr = C.CString(rStr)*/
		packErrInfo(MultiNewTransactionErr, rerrType, rerr, "MultiNewTransactionErr err, from:%v, to:%v, v:%v, nonce:%v, engine index:%v",
			fromAddr.String(), addr.String(), C.GoString(v), oldTx.Nonce(), engine.ctx.index)
		// *rerr = C.CString(packV8Err(MultiNewTransactionErr, rStr, engine.ctx.index))
		return nil
	}

	// event address need to user
	var head unsafe.Pointer
	if engine.ctx.head == nil {
		head = unsafe.Pointer(engine.v8engine)
	} else {
		head = engine.ctx.head
	}
	newCtx, err := NewChildContext(engine.ctx.block, newTx, contract, engine.ctx.state, head, engine.ctx.index+1)
	if err != nil {
		/*rStr := fmt.Sprintf("NewContext err:%v", err)
		logging.CLog().Errorf(rStr)
		*rerrType = MultiNewChildContext
		*rerr = C.CString(rStr)*/
		packErrInfo(MultiNewChildContext, rerrType, rerr, "NewContext err:%v, engine index:%v", err, engine.ctx.index)
		// *rerr = C.CString(packV8Err(MultiNewTransactionErr, rStr, engine.ctx.index))

		// logging.CLog().Errorf("NewContext err:%v", err)
		return nil
	}

	verbInstruction, verbMem := engine.GetNVMVerbResources()
	verbInstruction -= uint64(RunMultilevelContractSourceFuncCost)
	verbInstruction -= uint64(TransferFuncCost)

	logging.CLog().Infof("begin create New V8,intance:%v, mem:%v", verbInstruction, verbMem)
	engineNew := NewV8Engine(newCtx)
	engineNew.SetExecutionLimits(verbInstruction, verbMem)
	// logging.CLog().Errorf("begin Call,source:%v, sourceType:%v", deploy.Source, deploy.SourceType)
	val, err := engineNew.Call(string(deploy.Source), deploy.SourceType, C.GoString(funcName), C.GoString(args))
	gasCout := engineNew.ExecutionInstructions()
	gasSum += gasCout
	/*instructions, err := util.NewUint128FromInt(int64(gasCout))
	if err != nil {
		return util.NewUint128(), "", err
	}*/

	engineNew.Dispose()
	if err != nil {
		if err == ErrExceedMemoryLimits {
			/*rStr := fmt.Sprintf("NewContext err:%v", err)
			logging.CLog().Errorf(rStr)
			*rerrType = MUltiSystemMemLimit
			*rerr = C.CString(rStr)*/
			packErrInfo(MultiSystemMemLimit, rerrType, rerr, "engine.call mem limit err:%v, engine index:%v", err, engine.ctx.index)
			// *rerr = C.CString(packV8Err(MUltiSystemMemLimit, rStr, engine.ctx.index))
		} else if err == ErrInsufficientGas {
			/*rStr := fmt.Sprintf("NewContext err:%v", err)
			logging.CLog().Errorf(rStr)
			*rerrType = MultiSystemInsufficientLimit
			*rerr = C.CString(rStr)*/
			packErrInfo(MultiSystemMemLimit, rerrType, rerr, "engine.call insuff limit err:%v, engine index:%v", err, engine.ctx.index)
			// *rerr = C.CString(packV8Err(MultiNewTransactionErr, rStr, engine.ctx.index))
		} else {
			packErrInfo(MultiCallErr, rerrType, rerr, "engine.call err:%v, engine index:%v", err, engine.ctx.index)
		}

		//logging.CLog().Errorf()
		return nil
	}
	logging.CLog().Infof("end cal val:%v,gascount:%V,gasSum:%v", val, gasCout, gasSum)
	*gasCnt = C.size_t(gasSum)
	return C.CString(string(val))
	//return C.CString("")
}
