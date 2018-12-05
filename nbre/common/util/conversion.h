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

#pragma once
#include "common/common.h"
#include "common/math.h"

namespace neb {

class int128_conversion {
public:
  inline int128_conversion() { m_data.m_data = 0; }
  inline int128_conversion(int128_t data) { m_data.m_data = data; }

  template <typename T> inline T to_float() {
    uint64_t tmp = 1ULL << 63;
    return T(low()) + T(high()) * T(tmp) * 2;
  }

  inline uint64_t low() { return m_data.m_detail.m_low; }
  inline int64_t high() { return m_data.m_detail.m_high; }

private:
  union _int128_t {
    _int128_t() { m_data = 0; }
    int128_t m_data;
    struct _detail_t {
      uint64_t m_low;
      int64_t m_high;
    };
    _detail_t m_detail;
  };

  _int128_t m_data;
};

} // namespace neb
