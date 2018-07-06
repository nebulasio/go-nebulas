package nvm

/*
#include "v8/lib/nvm_error.h"
*/
import "C"

import (
	"crypto/md5"
	"encoding/binary"
	"io"
	"strconv"
	"unsafe"

	"math/rand"

	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

//GetTxRandomFunc return random
//export GetTxRandomFunc
func GetTxRandomFunc(handler unsafe.Pointer, gasCnt *C.size_t, result **C.char, exceptionInfo **C.char) int {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		logging.VLog().Error("random.GetTxRandomFunc Unexpected error: failed to get engine")
		return C.NVM_UNEXPECTED_ERR
	}
	// calculate Gas.
	*gasCnt = C.size_t(GetTxRandomGasBase)

	if engine.ctx.contextRand == nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": engine.ctx.block.Height(),
		}).Error("ContextRand is nil")
		*exceptionInfo = C.CString("random.GetTxRandomFunc(), contextRand is nil")
		return C.NVM_EXCEPTION_ERR
	}

	if engine.ctx.contextRand.rand == nil {
		bs := engine.ctx.block.RandomSeed()
		if len(bs) == 0 {
			logging.VLog().WithFields(logrus.Fields{
				"height": engine.ctx.block.Height(),
			}).Error("block seed is nil")
			*exceptionInfo = C.CString("random.GetTxRandomFunc(), randomSeed len is zero")
			return C.NVM_EXCEPTION_ERR
		}

		txhash := engine.ctx.tx.Hash().String()
		if len(txhash) == 0 {
			logging.VLog().WithFields(logrus.Fields{
				"height": engine.ctx.block.Height(),
			}).Error("transaction hash is nil")
			*exceptionInfo = C.CString("random.GetTxRandomFunc(), randomSeed len is zero")
			return C.NVM_EXCEPTION_ERR
		}

		m := md5.New()
		io.WriteString(m, bs)
		io.WriteString(m, txhash)
		seed := int64(binary.BigEndian.Uint64(m.Sum(nil)))
		engine.ctx.contextRand.rand = rand.New(rand.NewSource(seed))
	}
	*result = C.CString(strconv.FormatFloat(engine.ctx.contextRand.rand.Float64(), 'f', -1, 64))
	return C.NVM_SUCCESS
}
