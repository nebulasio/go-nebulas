// Copyright (C) 2018 go-nebulas authors
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

package nbre

/*
#include <stdint.h>
*/
import "C"
import (
	"encoding/json"
	"unsafe"

	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// NbreVersionFunc returns nbre version
//export NbreVersionFunc
func NbreVersionFunc(code C.int, holder unsafe.Pointer, major C.uint32_t, minor C.uint32_t, patch C.uint32_t) {
	handlerId := uint64(uintptr(holder))
	handler, _ := getNbreHandler(handlerId)
	logging.VLog().WithFields(logrus.Fields{
		"handlerId": handlerId,
		"handler":   handler,
	}).Debug("nbre versions callback")
	if handler != nil {
		version := &Version{
			Major: uint64(major),
			Minor: uint64(minor),
			Patch: uint64(patch),
		}
		result, err := json.Marshal(version)
		nbreHandled(handler, result, err)
	}
}

// NbreIrListFunc returns nbre ir list
//export NbreIrListFunc
func NbreIrListFunc(code C.int, holder unsafe.Pointer, ir_name_list *C.char) {
	handlerId := uint64(uintptr(holder))
	handler, _ := getNbreHandler(handlerId)
	if handler != nil {
		result := C.GoString(ir_name_list)

		logging.VLog().WithFields(logrus.Fields{
			"handlerId": handlerId,
			"handler":   handler,
			"ir":        result,
		}).Debug("nbre ir list callback")
		nbreHandled(handler, []byte(result), nil)
	}
}

// NbreIrVersionsFunc returns nbre ir versions
//export NbreIrVersionsFunc
func NbreIrVersionsFunc(code C.int, holder unsafe.Pointer, ir_versions *C.char) {
	handlerId := uint64(uintptr(holder))
	handler, _ := getNbreHandler(handlerId)
	if handler != nil {
		result := C.GoString(ir_versions)

		logging.VLog().WithFields(logrus.Fields{
			"handlerId": handlerId,
			"handler":   handler,
			"versions":  result,
		}).Debug("nbre ir version callback")
		nbreHandled(handler, []byte(result), nil)
	}
}

// NbreNrHandlerFunc returns nbre nr handler
//export NbreNrHandlerFunc
func NbreNrHandlerFunc(code C.int, holder unsafe.Pointer, nr_handler *C.char) {
	handlerId := uint64(uintptr(holder))
	handler, _ := getNbreHandler(handlerId)
	if handler != nil {
		result := C.GoString(nr_handler)

		logging.VLog().WithFields(logrus.Fields{
			"handlerId":  handlerId,
			"handler":    handler,
			"nr_handler": result,
		}).Debug("nbre nr callback")
		nbreHandled(handler, []byte(result), nil)
	}
}

// NbreNrResultFunc returns nbre nr list
//export NbreNrResultFunc
func NbreNrResultFunc(code C.int, holder unsafe.Pointer, nr_result *C.char) {
	handlerId := uint64(uintptr(holder))
	handler, _ := getNbreHandler(handlerId)
	if handler != nil {
		result := C.GoString(nr_result)

		logging.VLog().WithFields(logrus.Fields{
			"handlerId": handlerId,
			"handler":   handler,
			"nr":        result,
		}).Debug("nbre nr callback")
		nbreHandled(handler, []byte(result), nil)
	}
}
