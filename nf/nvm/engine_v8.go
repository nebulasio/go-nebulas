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
#cgo LDFLAGS: -L${SRCDIR}/native-lib -lnebulasv8

#include "v8/engine.h"

// Forward declaration.
void V8Log_cgo(int level, const char *msg);

char *RequireDelegateFunc_cgo(void *handler, const char *filename, size_t *lineOffset);

char *StorageGetFunc_cgo(void *handler, const char *key);
int StoragePutFunc_cgo(void *handler, const char *key, const char *value);
int StorageDelFunc_cgo(void *handler, const char *key);

char *GetTxByHashFunc_cgo(void *handler, const char *hash);
char *GetAccountStateFunc_cgo(void *handler, const char *address);
int TransferFunc_cgo(void *handler, const char *to, const char *value);
int VerifyAddressFunc_cgo(void *handler, const char *address);

void EventTriggerFunc_cgo(void *handler, const char *topic, const char *data);

*/
import "C"
import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Const.
const (
	SourceTypeJavaScript = "js"
	SourceTypeTypeScript = "ts"
)

// Errors
var (
	ErrExecutionFailed                = errors.New("execution failed")
	ErrDisallowCallPrivateFunction    = errors.New("disallow call private function")
	ErrExecutionTimeout               = errors.New("execution timeout")
	ErrInsufficientGas                = errors.New("insufficient gas")
	ErrExceedMemoryLimits             = errors.New("exceed memory limits")
	ErrInjectTracingInstructionFailed = errors.New("inject tracing instructions failed")
	ErrTranspileTypeScriptFailed      = errors.New("transpile TypeScript failed")
	ErrUnsupportedSourceType          = errors.New("unsupported source type")
)

var (
	v8engineOnce          = sync.Once{}
	storages              = make(map[uint64]*V8Engine, 256)
	storagesIdx           = uint64(0)
	storagesLock          = sync.RWMutex{}
	engines               = make(map[*C.V8Engine]*V8Engine, 256)
	enginesLock           = sync.RWMutex{}
	publicFuncNameChecker = regexp.MustCompile("^[a-zA-Z$][A-Za-z0-9_$]*$")
)

// V8Engine v8 engine.
type V8Engine struct {
	ctx                                *Context
	modules                            Modules
	v8engine                           *C.V8Engine
	enableLimits                       bool
	limitsOfExecutionInstructions      uint64
	limitsOfTotalMemorySize            uint64
	actualCountOfExecutionInstructions uint64
	actualTotalMemorySize              uint64
	lcsHandler                         uint64
	gcsHandler                         uint64
}

// InitV8Engine initialize the v8 engine.
func InitV8Engine() {
	C.Initialize()

	// Logger.
	C.InitializeLogger((C.LogFunc)(unsafe.Pointer(C.V8Log_cgo)))

	// Require.
	C.InitializeRequireDelegate((C.RequireDelegate)(unsafe.Pointer(C.RequireDelegateFunc_cgo)))

	// Storage.
	C.InitializeStorage((C.StorageGetFunc)(unsafe.Pointer(C.StorageGetFunc_cgo)), (C.StoragePutFunc)(unsafe.Pointer(C.StoragePutFunc_cgo)), (C.StorageDelFunc)(unsafe.Pointer(C.StorageDelFunc_cgo)))

	// Blockchain.
	C.InitializeBlockchain((C.GetTxByHashFunc)(unsafe.Pointer(C.GetTxByHashFunc_cgo)), (C.GetAccountStateFunc)(unsafe.Pointer(C.GetAccountStateFunc_cgo)), (C.TransferFunc)(unsafe.Pointer(C.TransferFunc_cgo)), (C.VerifyAddressFunc)(unsafe.Pointer(C.VerifyAddressFunc_cgo)))

	// Event.
	C.InitializeEvent((C.EventTriggerFunc)(unsafe.Pointer(C.EventTriggerFunc_cgo)))
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
		ctx:                                ctx,
		modules:                            NewModules(),
		v8engine:                           C.CreateEngine(),
		enableLimits:                       false,
		limitsOfExecutionInstructions:      0,
		limitsOfTotalMemorySize:            0,
		actualCountOfExecutionInstructions: 0,
		actualTotalMemorySize:              0,
	}

	(func() {
		enginesLock.Lock()
		defer enginesLock.Unlock()
		engines[engine.v8engine] = engine
	})()

	(func() {
		storagesLock.Lock()
		defer storagesLock.Unlock()

		storagesIdx++
		engine.lcsHandler = storagesIdx
		storagesIdx++
		engine.gcsHandler = storagesIdx

		storages[engine.lcsHandler] = engine
		storages[engine.gcsHandler] = engine
	})()
	return engine
}

// Dispose dispose all resources.
func (e *V8Engine) Dispose() {
	storagesLock.Lock()
	delete(storages, e.lcsHandler)
	delete(storages, e.gcsHandler)
	storagesLock.Unlock()

	enginesLock.Lock()
	delete(engines, e.v8engine)
	enginesLock.Unlock()

	C.DeleteEngine(e.v8engine)
}

// Context returns engine context
func (e *V8Engine) Context() *Context {
	return e.ctx
}

// SetTestingFlag set testing flag, default is False.
func (e *V8Engine) SetTestingFlag(flag bool) {
	if flag {
		e.v8engine.testing = C.int(1)
	} else {
		e.v8engine.testing = C.int(0)
	}
}

// SetExecutionLimits set execution limits of V8 Engine, prevent Halting Problem.
func (e *V8Engine) SetExecutionLimits(limitsOfExecutionInstructions, limitsOfTotalMemorySize uint64) {
	e.v8engine.limits_of_executed_instructions = C.size_t(limitsOfExecutionInstructions)
	e.v8engine.limits_of_total_memory_size = C.size_t(limitsOfTotalMemorySize)

	logging.VLog().WithFields(logrus.Fields{
		"limits_of_executed_instructions": e.v8engine.limits_of_executed_instructions,
		"limits_of_total_memory_size":     e.v8engine.limits_of_total_memory_size,
	}).Info("set execution limits.")

	e.limitsOfExecutionInstructions = limitsOfExecutionInstructions
	e.limitsOfTotalMemorySize = limitsOfTotalMemorySize
	e.enableLimits = limitsOfExecutionInstructions != 0 || limitsOfTotalMemorySize != 0

	// V8 needs at least 6M heap memory.
	if limitsOfTotalMemorySize > 0 && limitsOfTotalMemorySize < 6000000 {
		logging.VLog().Warnf("V8 needs at least 6M (6000000) heap memory, your limitsOfTotalMemorySize (%d) is too low.", limitsOfTotalMemorySize)
	}
}

// ExecutionInstructions returns the execution instructions
func (e *V8Engine) ExecutionInstructions() uint64 {
	return e.actualCountOfExecutionInstructions
}

// TranspileTypeScript transpile typescript to javascript and return it.
func (e *V8Engine) TranspileTypeScript(source string) (string, int, error) {
	cSource := C.CString(source)
	defer C.free(unsafe.Pointer(cSource))

	lineOffset := C.int(0)
	jsSource := C.TranspileTypeScriptModule(e.v8engine, cSource, &lineOffset)
	if jsSource == nil {
		return "", 0, ErrTranspileTypeScriptFailed
	}
	defer C.free(unsafe.Pointer(jsSource))

	return C.GoString(jsSource), int(lineOffset), nil

}

// InjectTracingInstructions process the source to inject tracing instructions.
func (e *V8Engine) InjectTracingInstructions(source string) (string, int, error) {
	cSource := C.CString(source)
	defer C.free(unsafe.Pointer(cSource))

	lineOffset := C.int(0)
	traceableCSource := C.InjectTracingInstructions(e.v8engine, cSource, &lineOffset)
	if traceableCSource == nil {
		return "", 0, ErrInjectTracingInstructionFailed
	}
	defer C.free(unsafe.Pointer(traceableCSource))

	return C.GoString(traceableCSource), int(lineOffset), nil
}

// CollectTracingStats collect tracing data from v8 engine.
func (e *V8Engine) CollectTracingStats() {
	// read memory stats.
	C.ReadMemoryStatistics(e.v8engine)

	e.actualCountOfExecutionInstructions = uint64(e.v8engine.stats.count_of_executed_instructions)
	e.actualTotalMemorySize = uint64(e.v8engine.stats.total_memory_size)
}

// RunScriptSource run js source.
func (e *V8Engine) RunScriptSource(source string, sourceLineOffset int) (err error) {
	if e.enableLimits {
		traceableSource, traceableSourceLineOffset, err := e.InjectTracingInstructions(source)
		if err != nil {
			return err
		}
		source = traceableSource
		sourceLineOffset += traceableSourceLineOffset
	}

	cSource := C.CString(source)
	defer C.free(unsafe.Pointer(cSource))
	var ret C.int

	done := make(chan bool, 1)
	go func() {
		ret = C.RunScriptSource(e.v8engine, cSource, C.int(sourceLineOffset), C.uintptr_t(e.lcsHandler),
			C.uintptr_t(e.gcsHandler))
		done <- true
	}()

	select {
	case <-done:
		if ret != 0 {
			err = ErrExecutionFailed
		}
	case <-time.After(10 * time.Second):
		C.TerminateExecution(e.v8engine)
		err = ErrExecutionTimeout

		// wait for C.RunScriptSource() returns.
		select {
		case <-done:
		}
	}

	// collect tracing stats.
	e.CollectTracingStats()

	if e.enableLimits {
		// check limits.
		ret = C.IsEngineLimitsExceeded(e.v8engine)
		if ret == 1 {
			err = ErrInsufficientGas
		} else if ret == 2 {
			err = ErrExceedMemoryLimits
		}

		if e.actualCountOfExecutionInstructions > e.limitsOfExecutionInstructions || err == ErrExceedMemoryLimits {
			e.actualCountOfExecutionInstructions = e.limitsOfExecutionInstructions
		}
	}

	return
}

// Call function in a script
func (e *V8Engine) Call(source, sourceType, function, args string) error {
	if publicFuncNameChecker.MatchString(function) == false || strings.EqualFold("init", function) == true {
		return ErrDisallowCallPrivateFunction
	}
	return e.RunContractScript(source, sourceType, function, args)
}

// DeployAndInit a contract
func (e *V8Engine) DeployAndInit(source, sourceType, args string) error {
	return e.RunContractScript(source, sourceType, "init", args)
}

// RunContractScript execute script in Smart Contract's way.
func (e *V8Engine) RunContractScript(source, sourceType, function, args string) error {
	var runnableSource string
	var sourceLineOffset int
	var err error

	switch sourceType {
	case SourceTypeJavaScript:
		runnableSource, sourceLineOffset, err = e.prepareRunnableContractScript(source, function, args)
	case SourceTypeTypeScript:
		// transpile to javascript.
		jsSource, _, err := e.TranspileTypeScript(source)
		if err != nil {
			return err
		}
		runnableSource, sourceLineOffset, err = e.prepareRunnableContractScript(jsSource, function, args)
	default:
		return ErrUnsupportedSourceType
	}

	if err != nil {
		return err
	}

	return e.RunScriptSource(runnableSource, sourceLineOffset)
}

// AddModule add module.
func (e *V8Engine) AddModule(id, source string, sourceLineOffset int) error {
	// inject tracing instruction when enable limits.
	if e.enableLimits {
		traceableSource, lineOffset, err := e.InjectTracingInstructions(source)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Error("inject tracing instruction failed.")
			return err
		}
		source = traceableSource
		sourceLineOffset = lineOffset
	}

	e.modules.Add(NewModule(id, source, sourceLineOffset))
	return nil
}

func (e *V8Engine) prepareRunnableContractScript(source, function, args string) (string, int, error) {
	sourceLineOffset := 0

	// add module.
	const ModuleID string = "contract.js"
	if err := e.AddModule(ModuleID, source, sourceLineOffset); err != nil {
		return "", 0, err
	}

	// prepare for execute.
	blockJSON, _ := e.ctx.SerializeContextBlock()
	txJSON, _ := e.ctx.SerializeContextTx()
	var runnableSource string

	if len(args) > 0 {
		runnableSource = fmt.Sprintf("var __contract = require(\"%s\");\n var __instance = new __contract();\n Blockchain.blockParse(\"%s\");\n Blockchain.transactionParse(\"%s\");\n __instance[\"%s\"].apply(__instance, JSON.parse(\"%s\"));\n", ModuleID, formatArgs(string(blockJSON)), formatArgs(string(txJSON)), function, formatArgs(args))
	} else {
		runnableSource = fmt.Sprintf("var __contract = require(\"%s\");\n var __instance = new __contract();\n Blockchain.blockParse(\"%s\");\n Blockchain.transactionParse(\"%s\");\n __instance[\"%s\"].apply(__instance);\n", ModuleID, formatArgs(string(blockJSON)), formatArgs(string(txJSON)), function)
	}
	return runnableSource, 0, nil
}

func getEngineByStorageHandler(handler uint64) (*V8Engine, state.Account) {
	storagesLock.RLock()
	engine := storages[handler]
	storagesLock.RUnlock()

	if engine == nil {
		logging.VLog().WithFields(logrus.Fields{
			"func":          "nvm.getEngineByStorageHandler",
			"wantedHandler": handler,
		}).Error("wantedHandler is not found.")
		return nil, nil
	}

	if engine.lcsHandler == handler {
		return engine, engine.ctx.contract
	} else if engine.gcsHandler == handler {
		// disable gcs according to issue https://github.com/nebulasio/go-nebulas/issues/23.
		return nil, nil
		// return engine, engine.ctx.owner
	} else {
		logging.VLog().WithFields(logrus.Fields{
			"func":          "nvm.getEngineByStorageHandler",
			"lcsHandler":    engine.lcsHandler,
			"gcsHandler":    engine.gcsHandler,
			"wantedHandler": handler,
		}).Error("in-consistent storage handler.")
		return nil, nil
	}
}

func getEngineByEngineHandler(handler unsafe.Pointer) *V8Engine {
	v8engine := (*C.V8Engine)(handler)
	enginesLock.RLock()
	defer enginesLock.RUnlock()

	return engines[v8engine]
}

func formatArgs(s string) string {
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "\n", "\\n", -1)
	s = strings.Replace(s, "\r", "\\r", -1)
	s = strings.Replace(s, "\"", "\\\"", -1)
	return s
}
