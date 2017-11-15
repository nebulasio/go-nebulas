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
	log "github.com/sirupsen/logrus"
)

type accountState struct {
	nonce   uint64 `json:"nonce"`
	balance string `json:"balance"`
}

// GetBlockByHashFunc returns the block info by hash
//export GetBlockByHashFunc
func GetBlockByHashFunc(handler unsafe.Pointer, hash *C.char) *C.char {
	engine, _ := getEngineAndStorage(uint64(uintptr(handler)))
	if engine == nil {
		return nil
	}
	block, err := engine.ctx.chain.SerializeBlockByHash([]byte(C.GoString(hash)))
	if err != nil {
		log.WithFields(log.Fields{
			"func":    "nvm.GetBlockByHashFunc",
			"handler": uint64(uintptr(handler)),
			"hash":    C.GoString(hash),
			"err":     err,
		}).Info("GetBlockByHashFunc get block failed.")
		return nil
	}
	json, _ := json.Marshal(block)
	return C.CString(string(json))
}

// GetTxByHashFunc returns tx info by hash
//export GetTxByHashFunc
func GetTxByHashFunc(handler unsafe.Pointer, hash *C.char) *C.char {
	engine, _ := getEngineAndStorage(uint64(uintptr(handler)))
	if engine == nil {
		return nil
	}
	tx, err := engine.ctx.chain.SerializeTxByHash([]byte(C.GoString(hash)))
	if err != nil {
		log.WithFields(log.Fields{
			"func":    "nvm.GetTxByHashFunc",
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(hash),
			"err":     err,
		}).Error("GetTxByHashFunc get tx failed.")
		return nil
	}
	json, _ := json.Marshal(tx)
	return C.CString(string(json))
}

// GetAccountStateFunc returns account info by address
//export GetAccountStateFunc
func GetAccountStateFunc(handler unsafe.Pointer, address *C.char) *C.char {
	engine, _ := getEngineAndStorage(uint64(uintptr(handler)))
	if engine == nil {
		return nil
	}
	addr := C.GoString(address)
	valid := engine.ctx.chain.VerifyAddress(addr)
	if !valid {
		log.WithFields(log.Fields{
			"func":    "nvm.GetAccountStateFunc",
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(address),
		}).Error("GetAccountStateFunc parse address failed.")
		return nil
	}

	// TODO: handle specific block number.
	acc := engine.ctx.state.GetOrCreateUserAccount([]byte(addr))
	state := &accountState{
		nonce:   acc.Nonce(),
		balance: acc.Balance().String(),
	}
	json, _ := json.Marshal(state)
	return C.CString(string(json))
}

// TransferFunc transfer vale to address
//export TransferFunc
func TransferFunc(handler unsafe.Pointer, to *C.char, v *C.char) int {
	engine, _ := getEngineAndStorage(uint64(uintptr(handler)))
	if engine == nil {
		return 0
	}

	addr := C.GoString(to)
	valid := engine.ctx.chain.VerifyAddress(addr)
	if !valid {
		log.WithFields(log.Fields{
			"func":    "nvm.TransferFunc",
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(to),
		}).Error("TransferFunc parse address failed.")
		return 0
	}
	toAcc := engine.ctx.state.GetOrCreateUserAccount([]byte(addr))

	var (
		amount *util.Uint128
		err    error
	)
	value := []byte(C.GoString(v))
	if len(value) > 0 {
		amount, err = util.NewUint128FromFixedSizeByteSlice(value)
		if err != nil {
			log.WithFields(log.Fields{
				"func":    "nvm.TransferFunc",
				"handler": uint64(uintptr(handler)),
				"key":     C.GoString(to),
				"err":     err,
			}).Error("TransferFunc parse balance failed.")
			return 0
		}
	} else {
		amount = util.NewUint128()
	}

	// update balance
	err = engine.ctx.contract.SubBalance(amount)
	if err != nil {
		log.WithFields(log.Fields{
			"func":    "nvm.TransferFunc",
			"handler": uint64(uintptr(handler)),
			"key":     C.GoString(to),
			"err":     err,
		}).Error("TransferFunc SubBalance failed.")
		return 0
	}
	toAcc.AddBalance(amount)
	return 1
}

// VerifyAddressFunc verify address is valid
//export VerifyAddressFunc
func VerifyAddressFunc(handler unsafe.Pointer, address *C.char) int {
	engine, _ := getEngineAndStorage(uint64(uintptr(handler)))
	if engine == nil {
		return 0
	}

	if engine.ctx.chain.VerifyAddress(C.GoString(address)) {
		return 1
	}
	return 0
}
