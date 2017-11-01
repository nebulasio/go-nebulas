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

// #cgo CFLAGS:
// #cgo LDFLAGS: -L${SRCDIR}/libs -lv8engine
// #include "v8/engine.h"
// void GoLogFunc_cgo(int level, const char *msg); // Forward declaration.
import "C"
import (
	"errors"
	"unsafe"
)

var (
	ErrFailed = errors.New("Failed.")
)

// V8Engine v8 engine.
type V8Engine struct {
	engine                *C.V8Engine
	balanceStorage        Storage
	localContractStorage  Storage
	globalContractStorage Storage
}

// InitV8Engine initialize the v8 engine.
func InitV8Engine() {
	C.Initialize((C.LogFunc)(unsafe.Pointer(C.GoLogFunc_cgo)))
}

// DisposeV8Engine dispose the v8 engine.
func DisposeV8Engine() {
	C.Dispose()
}

// NewV8Engine return new V8Engine instance.
func NewV8Engine(balanceStorage, localContractStorage, globalContractStorage Storage) *V8Engine {
	engine := &V8Engine{
		engine:                C.CreateEngine(),
		balanceStorage:        balanceStorage,
		localContractStorage:  localContractStorage,
		globalContractStorage: globalContractStorage,
	}
	return engine
}

// Delete delete engine.
func (e *V8Engine) Delete() {
	C.DeleteEngine(e.engine)
	e.engine = nil
}

// RunScript
func (e *V8Engine) RunScriptSource(content string) error {
	ret := C.RunScriptSource(e.engine, C.CString(content))
	if ret != 0 {
		return ErrFailed
	}
	return nil
}

func (e *V8Engine) Call(contractAddress string, function, args string) error {
	return nil
}

func (e *V8Engine) DeployAndInit(source, args string) error {

	return nil
}
