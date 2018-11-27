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

namespace neb {
namespace fs {
class byte_shared {
public:
  inline byte_shared() { m_data.m_data = 0; }
  inline byte_shared(byte_t data) { m_data.m_data = data; }
  inline byte_shared(byte_t bits_low, byte_t bits_high) {
    m_data.m_detail.m_bits_low = bits_low;
    m_data.m_detail.m_bits_high = bits_high;
  }

  byte_shared(const byte_shared &) = default;
  byte_shared &operator=(const byte_shared &) = default;

  inline byte_t bits_low() const { return m_data.m_detail.m_bits_low; }
  inline byte_t bits_high() const { return m_data.m_detail.m_bits_high; }

  inline byte_t data() const { return m_data.m_data; }

private:
  union _byte_shared_data_t {
    byte_t m_data;
    struct _byte_shared_detail_t {
      byte_t m_bits_low : 4;
      byte_t m_bits_high : 4;
    };
    _byte_shared_detail_t m_detail;
  };

  _byte_shared_data_t m_data;
};
} // namespace fs
} // namespace neb
