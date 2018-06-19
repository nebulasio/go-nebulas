package nvm

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
func GetTxRandomFunc(handler unsafe.Pointer) *C.char {
	engine, _ := getEngineByStorageHandler(uint64(uintptr(handler)))
	if engine == nil || engine.ctx.block == nil {
		return nil
	}
	logging.CLog().Infof("Hello GetTxRandomFunc")
	// calculate Gas.
	// *gasCnt = C.size_t(GetTxByHashFuncCost)

	if engine.ctx.contextRand == nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": engine.ctx.block.Height(),
		}).Error("ContextRand is nil")
		return nil
	}

	if engine.ctx.contextRand.rand == nil {
		bs := engine.ctx.block.RandomSeed()
		if len(bs) == 0 {
			logging.VLog().WithFields(logrus.Fields{
				"height": engine.ctx.block.Height(),
			}).Error("block seed is nil")
			return nil
		}

		txhash := engine.ctx.tx.Hash().String()
		if len(txhash) == 0 {
			logging.VLog().WithFields(logrus.Fields{
				"height": engine.ctx.block.Height(),
			}).Error("transaction hash is nil")
			return nil
		}

		m := md5.New()
		io.WriteString(m, bs)
		io.WriteString(m, txhash)
		seed := int64(binary.BigEndian.Uint64(m.Sum(nil)))
		engine.ctx.contextRand.rand = rand.New(rand.NewSource(seed))
	}

	return C.CString(strconv.FormatFloat(engine.ctx.contextRand.rand.Float64(), 'f', -1, 64))
}
