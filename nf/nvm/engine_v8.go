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
#include <stdlib.h>
#cgo CFLAGS:
#cgo LDFLAGS: -lv8engine

#include "v8/engine.h"

// Forward declaration.
void V8Log_cgo(int level, const char *msg);
char *StorageGetFunc_cgo(void *handler, const char *key);
int StoragePutFunc_cgo(void *handler, const char *key, const char *value);
int StorageDelFunc_cgo(void *handler, const char *key);
*/
import "C"
import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unsafe"

	"github.com/nebulasio/go-nebulas/core/state"
	log "github.com/sirupsen/logrus"
)

// Errors
var (
	ErrExecutionFailed     = errors.New("execute source failed")
	ErrInvalidFunctionName = errors.New("invalid function name")
)

var (
	v8engineOnce = sync.Once{}
	storages     = make(map[uint64]*V8Engine, 256)
	storagesIdx  = uint64(0)
	storagesLock = sync.RWMutex{}

	functionNameRe = regexp.MustCompile("^[a-zA-Z_]+$")
)

// V8Engine v8 engine.
type V8Engine struct {
	ctx *Context

	v8engine *C.V8Engine

	lcsHandler uint64
	gcsHandler uint64
}

// InitV8Engine initialize the v8 engine.
func InitV8Engine() {
	C.Initialize()
	C.InitializeLogger((C.LogFunc)(unsafe.Pointer(C.V8Log_cgo)))
	C.InitializeStorage((C.StorageGetFunc)(unsafe.Pointer(C.StorageGetFunc_cgo)), (C.StoragePutFunc)(unsafe.Pointer(C.StoragePutFunc_cgo)), (C.StorageDelFunc)(unsafe.Pointer(C.StorageDelFunc_cgo)))
}

// DisposeV8Engine dispose the v8 engine.
func DisposeV8Engine() {
	C.Dispose()
}

// NewV8Engine return new V8Engine instance.
func NewV8Engine(ctx *Context) *V8Engine {
	v8engineOnce.Do(func() {
		InitV8Engine()
	})

	engine := &V8Engine{
		ctx:      ctx,
		v8engine: C.CreateEngine(),
	}

	storagesLock.Lock()
	defer storagesLock.Unlock()

	storagesIdx++
	engine.lcsHandler = storagesIdx
	storagesIdx++
	engine.gcsHandler = storagesIdx

	storages[engine.lcsHandler] = engine
	storages[engine.gcsHandler] = engine

	return engine
}

// Dispose dispose all resources.
func (e *V8Engine) Dispose() {
	C.DeleteEngine(e.v8engine)
	storagesLock.Lock()
	delete(storages, e.lcsHandler)
	delete(storages, e.gcsHandler)
	storagesLock.Unlock()
}

// RunScriptSource run js source.
func (e *V8Engine) RunScriptSource(content string) error {
	data := C.CString(content)
	defer C.free(unsafe.Pointer(data))
	// log.Errorf("[--------------] RunScriptSource, lcsHandler = %d, gcsHadnler = %d", e.lcsHandler, e.gcsHandler)
	ret := C.RunScriptSource(e.v8engine, data, C.uintptr_t(e.lcsHandler),
		C.uintptr_t(e.gcsHandler))

	if ret != 0 {
		return ErrExecutionFailed
	}
	return nil
}

// Call function in a script
func (e *V8Engine) Call(source, function, args string) error {
	if functionNameRe.MatchString(function) == false || strings.Compare("init", function) == 0 {
		return ErrInvalidFunctionName
	}
	return e.executeScript(source, function, args)
}

// DeployAndInit a contract
func (e *V8Engine) DeployAndInit(source, args string) error {
	return e.executeScript(source, "init", args)
}

// Execute execute the script and return error.
func (e *V8Engine) executeScript(source, function, args string) error {
	executablesource, err := e.prepareExecutableSource(source, function, args)
	if err != nil {
		return err
	}

	// log.WithFields(log.Fields{
	// 	"source":           source,
	// 	"args":             args,
	// 	"function":         function,
	// 	"executablesource": executablesource,
	// }).Info("executeScript")

	return e.RunScriptSource(executablesource)
}

func (e *V8Engine) prepareExecutableSource(source, function, args string) (string, error) {
	// inject tracing instructions.
	cSource := C.CString(source)
	defer C.free(unsafe.Pointer(cSource))
	traceableCSource := C.InjectTracingInstructions(e.v8engine, cSource)
	if traceableCSource == nil {
		return "", errors.New("inject tracing instructions failed")
	}
	defer C.free(unsafe.Pointer(traceableCSource))

	// encapsulate to module style.
	cmSource := C.EncapsulateSourceToModuleStyle(traceableCSource)
	defer C.free(unsafe.Pointer(cmSource))

	// prepare for execute.
	contextJSON := e.ctx.getParamsJSON()
	var executablesource string

	if len(args) > 0 {
		executablesource = fmt.Sprintf("var __contract = %s;\n var __instance = new __contract();\n __instance.context = JSON.parse(\"%s\");\n __instance[\"%s\"].apply(__instance, JSON.parse(\"%s\"));\n", C.GoString(cmSource), formatArgs(contextJSON), function, formatArgs(args))
	} else {
		executablesource = fmt.Sprintf("var __contract = %s;\n var __instance = new __contract();\n __instance.context = JSON.parse(\"%s\");\n __instance[\"%s\"].apply(__instance);\n", C.GoString(cmSource), formatArgs(contextJSON), function)
	}

	return executablesource, nil
}

func getEngineAndStorage(handler uint64) (*V8Engine, state.Account) {
	// log.Errorf("[--------------] getEngineAndStorage, handler = %d", handler)

	storagesLock.RLock()
	engine := storages[handler]
	storagesLock.RUnlock()

	if engine == nil {
		log.WithFields(log.Fields{
			"func":          "nvm.getEngineAndStorage",
			"wantedHandler": handler,
		}).Error("wantedHandler is not found.")
		return nil, nil
	}

	if engine.lcsHandler == handler {
		return engine, engine.ctx.contract
	} else if engine.gcsHandler == handler {
		return engine, engine.ctx.owner
	} else {
		log.WithFields(log.Fields{
			"func":          "nvm.getEngineAndStorage",
			"lcsHandler":    engine.lcsHandler,
			"gcsHandler":    engine.gcsHandler,
			"wantedHandler": handler,
		}).Error("in-consistent storage handler.")
		return nil, nil
	}
}

func formatArgs(args string) string {
	return strings.Replace(args, "\"", "\\\"", -1)
}
