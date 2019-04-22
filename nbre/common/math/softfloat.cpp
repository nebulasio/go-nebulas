// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify
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
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//

#include "common/math/softfloat.hpp"
#include <glog/logging.h>
#include <sstream>

/*----------------------------------------------------------------------------
| Software floating-point exception flags.
| extern THREAD_LOCAL uint_fast8_t softfloat_exceptionFlags;
| enum {
|     softfloat_flag_inexact   =  1,
|     softfloat_flag_underflow =  2,
|     softfloat_flag_overflow  =  4,
|     softfloat_flag_infinite  =  8,
|     softfloat_flag_invalid   = 16
| };
*----------------------------------------------------------------------------*/

void softfloat_raiseFlags(uint_fast8_t exception_flag) {
  std::stringstream ss;
  ss << "raise flag: " << static_cast<int>(exception_flag)
     << ", softfloat exception: ";
  uint_fast8_t flag = 0x10;
  while (flag) {
    switch (flag & exception_flag) {
    case softfloat_flag_inexact:
      ss << "inexact ";
      break;
    case softfloat_flag_underflow:
      ss << "underflow ";
      break;
    case softfloat_flag_overflow:
      ss << "overflow ";
      break;
    case softfloat_flag_infinite:
      ss << "infinite ";
      break;
    case softfloat_flag_invalid:
      ss << "invalid ";
      break;
    default:
      break;
    }
    flag >>= 1;
  }
  LOG(INFO) << ss.str();

  // ignore softfloat exception
  // throw std::runtime_error("softfloat exception");
}
