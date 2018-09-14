// Copyright (C) 2018 go-nebulas authors
//
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

#pragma once
#include "common/common.h"
#include <boost/endian/conversion.hpp>

namespace neb {
namespace util {
template <typename T>
auto byte_to_number(byte_t *bytes, size_t len) ->
    typename std::enable_if<std::is_arithmetic<T>::value, T>::type {
  if (len < sizeof(T))
    return T();

  T *val = (T *)bytes;
  T ret = boost::endian::big_to_native(*val);
  return ret;
}

template <typename T>
auto number_to_byte(T val, byte_t *bytes, size_t len) ->
    typename std::enable_if<std::is_arithmetic<T>::value, void>::type {
  if (len < sizeof(T))
    return;

  T v = boost::endian::native_to_big(val);
  T *p = (T *)bytes;
  *p = v;
  return;
}
}
}
