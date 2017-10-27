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

import (
	"github.com/nebulasio/go-nebulas/core"
	log "github.com/sirupsen/logrus"
)

/*
#cgo CXXFLAGS: -std=c++11
#include "engine.h"
*/
import "C"

// Engine the engine of NVM, build on top of llvm with sandboxing.
type Engine struct {
	bc *core.BlockChain
}

// NewEngine return new engine instance.
func NewEngine(bc *core.BlockChain) (*Engine, error) {
	e := &Engine{
		bc: bc,
	}

	C.Initialize()

	return e, nil
}

//export LogError
func LogError(msg *C.char) {
	log.WithFields(log.Fields{
		"func": "LogError",
	}).Error(C.GoString(msg))
}

//export LogInfo
func LogInfo(msg *C.char) {
	log.WithFields(log.Fields{
		"func": "LogInfo",
	}).Info(C.GoString(msg))
}
