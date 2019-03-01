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

import (
	"crypto/md5"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Sha256Func ..
//export Sha256Func
//return: sha result(string), gasCnt(uint64)
func Sha256Func(s string) (string, uint64) {
	var gasCnt uint64 = 0
	gasCnt = uint64(len(s) + CryptoSha256GasBase)

	r := hash.Sha256([]byte(s))
	return byteutils.Hex(r), gasCnt
}

// Sha3256Func ..
//export Sha3256Func
//return: sha result(string), gasCnt(uint64)
func Sha3256Func(s string) (string, uint64) {
	var gasCnt uint64 = 0
	gasCnt = uint64(len(s) + CryptoSha3256GasBase)

	r := hash.Sha3256([]byte(s))
	return byteutils.Hex(r), gasCnt
}

// Ripemd160Func ..
//export Ripemd160Func
//return execution result(string), gasCnt(uint64)
func Ripemd160Func(s string) (string, uint64) {
	var gasCnt uint64 = 0
	gasCnt = uint64(len(s) + CryptoRipemd160GasBase)

	r := hash.Ripemd160([]byte(s))
	return byteutils.Hex(r), gasCnt
}

// RecoverAddressFunc ..
//export RecoverAddressFunc
//params: alg(int), d(string), s(string)
//return: execution result(string), gasCnt(uint64)
func RecoverAddressFunc(alg int, d, s string) (string, uint64) {

	gasCnt := uint64(CryptoRecoverAddressGasBase)
	result := ""

	plain, err := byteutils.FromHex(d)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"hash": d,
			"sign": s,
			"alg":  alg,
			"err":  err,
		}).Debug("convert hash to byte array error.")
		return result, gasCnt
	}
	cipher, err := byteutils.FromHex(s)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"data": d,
			"sign": s,
			"alg":  alg,
			"err":  err,
		}).Debug("convert sign to byte array error.")
		return result, gasCnt
	}
	addr, err := core.RecoverSignerFromSignature(keystore.Algorithm(alg), plain, cipher)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"data": d,
			"sign": s,
			"alg":  alg,
			"err":  err,
		}).Debug("recover address error.")
		return result, gasCnt
	}

	return addr.String(),gasCnt
}

// Md5Func ..
//export Md5Func
func Md5Func(s string) (string, uint64) {
	gasCnt := uint64(len(s) + CryptoMd5GasBase)

	r := md5.Sum([]byte(s))
	return byteutils.Hex(r[:]), gasCnt
}

// Base64Func ..
//export Base64Func
func Base64Func(s string) (string, uint64) {
	gasCnt := uint64(len(s) + CryptoBase64GasBase)

	r := hash.Base64Encode([]byte(s))
	return string(r), gasCnt
}
