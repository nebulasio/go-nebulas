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

// RunMultilevelContractSourceFunc multi run contract. output[c standard]: if err return nil else return "*"
//export RunMultilevelContractSourceFunc
func RunMultilevelContractSourceFunc(handler unsafe.Pointer, address *C.char, funcName *C.char, v *C.char, args *C.char, gasCnt *C.size_t) *C.char {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Error("Failed to get engine.")
		return nil
	}
	if engine.ctx.index >= uint32(MultiNvmMax) {
		logging.VLog().Errorf("Failed to run nvm, becase more nvm ,current nvm:%v", engine.ctx.index)
		return nil
	}
	var gasSum uint64
	//*gasCnt = C.size_t(RunMultilevelContractSourceFuncCost)
	gasSum = uint64(RunMultilevelContractSourceFuncCost)
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
	deploy, err := core.LoadDeployPayload(birthTx.Data())
	if err != nil {
		return nil
	}

	//run
	payloadType := core.TxPayloadCallType
	callpayload, err := core.NewCallPayload(C.GoString(funcName), C.GoString(args))
	if err != nil {
		logging.VLog().Errorf("core.NewCallPayload err:", err)
		return nil
	}
	payload, err := callpayload.ToBytes()
	if err != nil {
		logging.VLog().Errorf("callpayload.ToBytes err:", err)
		return nil
	}

	oldTx := engine.ctx.tx
	// zeroVal := util.NewUint128()
	from := engine.ctx.contract.Address()
	fromAddr, err := core.AddressParseFromBytes(from)
	if err != nil {
		logging.CLog().Errorf("core.AddressParse err:", err)
		return nil
	}
	//transfer
	var transferCoseGas uint64
	//var iRet int
	iRet := TransferByAddress(handler, fromAddr, addr, C.GoString(v), &transferCoseGas)
	if iRet != 0 {
		return nil
	}
	// logging.CLog().Errorf("end TransferFunc:form:%v, to:%v, v:%v,transferCoseGas:%v", engine.ctx.tx.From().Bytes(), addr.Bytes(), C.GoString(v), transferCoseGas)
	toValue, err := util.NewUint128FromString(C.GoString(v))
	if err != nil {
		return nil
	}
	newTx, err := core.NewTransaction(oldTx.ChainID(), fromAddr, addr, toValue, oldTx.Nonce(), payloadType,
		payload, oldTx.GasPrice(), oldTx.GasLimit())
	if err != nil {
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
		logging.CLog().Errorf("NewContext err:%v", err)
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
		return nil
	}
	logging.CLog().Infof("end cal val:%v,gascount:%V,gasSum:%v", val, gasCout, gasSum)
	*gasCnt = C.size_t(gasSum)
	return C.CString(string(val))
	//return C.CString("")
}
