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

import "C"

import (
	"github.com/nebulasio/go-nebulas/util/logging"
)

// V8Log export V8Log
//export V8Log
func V8Log(level int, msg *C.char) {
	s := C.GoString(msg)

	switch level {
	case 1:
		logging.CLog().Debug(s)
	case 2:
		logging.CLog().Warn(s)
	case 3:
		logging.CLog().Info(s)
	case 4:
		logging.CLog().Error(s)
	default:
		logging.CLog().Error(s)
	}
}
