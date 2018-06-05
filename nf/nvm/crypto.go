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

package nvm

import "C"

import (
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Sha256Func ..
//export Sha256Func
func Sha256Func(data *C.char, gasCnt *C.size_t) *C.char {
	s := C.GoString(data)
	*gasCnt = C.size_t(1000 + len(s))

	r := hash.Sha256([]byte(s))
	return C.CString(byteutils.Hex(r))
}

// Sha3256Func ..
//export Sha3256Func
func Sha3256Func(data *C.char, gasCnt *C.size_t) *C.char {
	s := C.GoString(data)
	*gasCnt = C.size_t(1000 + len(s))

	r := hash.Sha3256([]byte(s))
	return C.CString(byteutils.Hex(r))
}

// Ripemd160Func ..
//export Ripemd160Func
func Ripemd160Func(data *C.char, gasCnt *C.size_t) *C.char {
	s := C.GoString(data)
	*gasCnt = C.size_t(1000 + len(s))

	r := hash.Ripemd160([]byte(s))
	return C.CString(byteutils.Hex(r))
}

// RecoverAddressFunc ..
//export RecoverAddressFunc
func RecoverAddressFunc(alg int, data, sign *C.char, gasCnt *C.size_t) *C.char {
	d := C.GoString(data)
	s := C.GoString(sign)

	*gasCnt = C.size_t(0)

	plain, err := byteutils.FromHex(d)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"hash": d,
			"sign": s,
			"alg":  alg,
			"err":  err,
		}).Debug("convert hash to byte array error.")
		return nil
	}
	cipher, err := byteutils.FromHex(s)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"data": d,
			"sign": s,
			"alg":  alg,
			"err":  err,
		}).Debug("convert sign to byte array error.")
		return nil
	}
	addr, err := core.RecoverSignerFromSignature(keystore.Algorithm(alg), plain, cipher)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"data": d,
			"sign": s,
			"alg":  alg,
			"err":  err,
		}).Debug("recover address error.")
		return nil
	}

	return C.CString(addr.String())
}
