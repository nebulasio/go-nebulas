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
	"encoding/json"
	"unsafe"

	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// GetTxByHashFunc returns tx info by hash
//export GetTxByHashFunc
func GetTxByHashFunc(handler unsafe.Pointer, hash *C.char) *C.char {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		return nil
	}
	tx, err := engine.ctx.SerializeTxByHash([]byte(C.GoString(hash)))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(hash),
			"err":     err,
		}).Debug("GetTxByHashFunc get tx failed.")
		return nil
	}
	return C.CString(string(tx))
}

// GetAccountStateFunc returns account info by address
//export GetAccountStateFunc
func GetAccountStateFunc(handler unsafe.Pointer, address *C.char) *C.char {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		return nil
	}
	addr := C.GoString(address)
	valid := engine.ctx.block.VerifyAddress(addr)
	if !valid {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(address),
		}).Debug("GetAccountStateFunc parse address failed.")
		return nil
	}

	acc, err := engine.ctx.state.GetOrCreateUserAccount([]byte(addr))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"address": addr,
			"err":     err,
		}).Debug("GetAccountStateFunc get account state failed.")
		return nil
	}
	state := &AccountState{
		Nonce:   acc.Nonce(),
		Balance: acc.Balance().String(),
	}
	json, _ := json.Marshal(state)
	return C.CString(string(json))
}

// TransferFunc transfer vale to address
//export TransferFunc
func TransferFunc(handler unsafe.Pointer, to *C.char, v *C.char) int {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		return 1
	}

	addr := C.GoString(to)
	valid := engine.ctx.block.VerifyAddress(addr)
	if !valid {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(to),
		}).Debug("TransferFunc parse address failed.")
		return 1
	}

	toAcc, err := engine.ctx.state.GetOrCreateUserAccount([]byte(addr))
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"address": addr,
			"err":     err,
		}).Debug("GetAccountStateFunc get account state failed.")
		return 1
	}

	amount := util.NewUint128FromString(C.GoString(v))

	// update balance
	err = engine.ctx.contract.SubBalance(amount)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(to),
			"err":     err,
		}).Debug("TransferFunc SubBalance failed.")
		return 1
	}

	toAcc.AddBalance(amount)
	return 0
}

// VerifyAddressFunc verify address is valid
//export VerifyAddressFunc
func VerifyAddressFunc(handler unsafe.Pointer, address *C.char) int {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		return 0
	}

	if engine.ctx.block.VerifyAddress(C.GoString(address)) {
		return 1
	}
	return 0
}
