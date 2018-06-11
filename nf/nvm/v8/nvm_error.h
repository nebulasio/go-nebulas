// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//

#ifndef _NEBULAS_NF_NVM_V8_ENGINE_ERROR_H_
#define _NEBULAS_NF_NVM_V8_ENGINE_ERROR_H_

enum nvmErrno {
  NVM_SUCCESS = 0,
  NVM_EXCEPTION_ERR = -1,
  NVM_MEM_LIMIT_ERR = -2,
  NVM_GAS_LIMIT_ERR = -3,
  NVM_UNEXPECTED_ERR = -4,
};

#endif //_NEBULAS_NF_NVM_V8_ENGINE_ERROR_H_