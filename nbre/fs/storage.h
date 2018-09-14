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
#include "common/util/byte.h"

namespace neb {
namespace fs {
class storage {
public:
  template <typename T, typename KT> T get(const KT &key) {
    // TODO this is only for number types
    return util::byte_to_number<T>(get_bytes(util::number_to_byte(key)));
  }

  template <typename T, typename KT> void put(const KT &key, const T &val) {
    //! TODO this is only for number types
    put_bytes(util::number_to_byte(key), util::number_to_byte(val));
  }
  template <typename KT> void del(const KT &key) {
    //! TODO this is only for number types
    del_by_bytes(util::number_to_byte(key));
  }

  virtual util::bytes get_bytes(const util::bytes &key) = 0;
  virtual void put_bytes(const util::bytes &key, const util::bytes &val) = 0;
  virtual void del_by_bytes(const util::bytes &key) = 0;

  virtual void enable_batch() = 0;
  virtual void disable_batch() = 0;
  virtual void flush() = 0;
};
}
}
