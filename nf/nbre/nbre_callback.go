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
	"unsafe"
)

// NbreVersionFunc returns nbre version
//export NbreVersionFunc
func NbreVersionFunc(code C.int, holder unsafe.Pointer, major C.uint32_t, minor C.uint32_t, patch C.uint32_t) {
	version := &Version{
		Major: uint64(major),
		Minor: uint64(minor),
		Patch: uint64(patch),
	}
	nbreHandled(code, holder, version, nil)
}

// NbreIrListFunc returns nbre ir list
//export NbreIrListFunc
func NbreIrListFunc(code C.int, holder unsafe.Pointer, ir_name_list *C.char) {
	result := C.GoString(ir_name_list)
	nbreHandled(code, holder, result, nil)
}

// NbreIrVersionsFunc returns nbre ir versions
//export NbreIrVersionsFunc
func NbreIrVersionsFunc(code C.int, holder unsafe.Pointer, ir_versions *C.char) {
	result := C.GoString(ir_versions)
	nbreHandled(code, holder, result, nil)
}

// NbreNrHandleFunc returns nbre nr handle
//export NbreNrHandleFunc
func NbreNrHandleFunc(code C.int, holder unsafe.Pointer, nr_handle *C.char) {
	result := C.GoString(nr_handle)
	nbreHandled(code, holder, result, nil)
}

// NbreNrResultByhandleFunc returns nbre nr list
//export NbreNrResultByhandleFunc
func NbreNrResultByhandleFunc(code C.int, holder unsafe.Pointer, nr_result *C.char) {
	result := C.GoString(nr_result)
	nbreHandled(code, holder, result, nil)
}

// NbreNrResultByHeightFunc returns nbre nr list
//export NbreNrResultByHeightFunc
func NbreNrResultByHeightFunc(code C.int, holder unsafe.Pointer, nr_result *C.char) {
	result := C.GoString(nr_result)
	nbreHandled(code, holder, result, nil)
}

// NbreNrSumFunc returns nbre nr summary data
//export NbreNrSumFunc
func NbreNrSumFunc(code C.int, holder unsafe.Pointer, nr_sum *C.char) {
	result := C.GoString(nr_sum)
	nbreHandled(code, holder, result, nil)
}

// NbreDipRewardFunc returns nbre dip list
//export NbreDipRewardFunc
func NbreDipRewardFunc(code C.int, holder unsafe.Pointer, dip_reward *C.char) {
	result := C.GoString(dip_reward)
	nbreHandled(code, holder, result, nil)
}
