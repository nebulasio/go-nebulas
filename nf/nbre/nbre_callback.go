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
func NbreVersionFunc(holder unsafe.Pointer, major C.uint32_t, minor C.uint32_t, patch C.uint32_t) {
	handlerId := uint64(uintptr(holder))
	handler, _ := getNbreHandler(handlerId)
	logging.CLog().WithFields(logrus.Fields{
		"handlerId": handlerId,
		"handler":   handler,
	}).Debug("nbre callback")
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
