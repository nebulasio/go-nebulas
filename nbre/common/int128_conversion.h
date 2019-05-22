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

namespace internal {
class int128_conversion_helper {
public:
  inline int128_conversion_helper() { m_data.m_data = 0; }
  inline int128_conversion_helper(const int128_t &data) {
    m_data.m_data = data;
  }

  template <typename T> inline T to_float() {
    T one = softfloat_cast<uint32_t, typename T::value_type>(1);
    T limits = softfloat_cast<uint64_t, typename T::value_type>(
        std::numeric_limits<uint64_t>::max());

    T tl = T(m_data.m_detail.m_low);
    T th = T(m_data.m_detail.m_high) * (limits + one);
    return tl + th;
  }

  template <typename T> inline int128_t from_float(const T &x) {
    T one = softfloat_cast<uint32_t, typename T::value_type>(1);
    T limits = softfloat_cast<uint64_t, typename T::value_type>(
        std::numeric_limits<uint64_t>::max());

    if (x < limits) {
      m_data.m_detail.m_low =
          softfloat_cast<typename T::value_type, uint64_t>(x);
      m_data.m_detail.m_high = 0;
      return m_data.m_data;
    }

    m_data.m_detail.m_high =
        softfloat_cast<typename T::value_type, int64_t>(x / (limits + one));
    m_data.m_detail.m_low = softfloat_cast<typename T::value_type, uint64_t>(
        x - m_data.m_detail.m_high * (limits + one));
    return m_data.m_data;
  }

  inline const int128_t &data() const { return m_data.m_data; }
  inline int128_t &data() { return m_data.m_data; }

  inline const uint64_t &low() const { return m_data.m_detail.m_low; }
  inline uint64_t &low() { return m_data.m_detail.m_low; }

  inline const int64_t &high() const { return m_data.m_detail.m_high; }
  inline int64_t &high() { return m_data.m_detail.m_high; }

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
} // namespace internal

template <typename T> inline T to_float(const int128_t &data) {
  internal::int128_conversion_helper ch(data);
  return ch.to_float<T>();
}
template <typename T> inline int128_t from_float(const T &x) {
  internal::int128_conversion_helper ch;
  return ch.from_float(x);
}

} // namespace neb
