package nvm

import "C"

import (
	"unsafe"

	"github.com/nebulasio/go-nebulas/util/logging"
)

//GetTxRandomFunc return random
//export GetTxRandomFunc
func GetTxRandomFunc(handler unsafe.Pointer) *C.char {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		return nil
	}
	logging.CLog().Infof("Hello GetTxRandomFunc")
	// calculate Gas.
	// *gasCnt = C.size_t(GetTxByHashFuncCost)

	return C.CString(string(""))
}
