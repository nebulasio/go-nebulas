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
#include "common/byte.h"
#include "common/common.h"

namespace neb {

template <int64_t ratio> class nas_currency_t {
public:
  typedef boost::multiprecision::int128_t nas_currency_value_t;

  nas_currency_t() : m_value(0) {}
  nas_currency_t(const nas_currency_value_t &v) : m_value(v) {}

  template <int64_t r>
  nas_currency_t(const nas_currency_t<r> &v) : m_value(v.wei_value()){};

  nas_currency_t(long double v) : m_value(v * ratio) {}

  nas_currency_t(const nas_currency_t<ratio> &v) : m_value(v.m_value) {}

  nas_currency_t<ratio> &operator=(const nas_currency_t<ratio> &v) {
    if (&v == this)
      return *this;
    m_value = v.m_value;
    return *this;
  }

  nas_currency_t<ratio> &operator++() {
    m_value += ratio;
    return *this;
  }
  nas_currency_t<ratio> operator++(int) {
    nas_currency_t<ratio> tmp(*this); // copy
    operator++();                     // pre-increment
    return tmp;                       // return old value
  }

  nas_currency_t<ratio> &operator--() {
    m_value -= ratio;
    return *this;
  }
  nas_currency_t<ratio> operator--(int) {
    nas_currency_t<ratio> tmp(*this); // copy
    operator--();                     // pre-increment
    return tmp;                       // return old value
  }

  template <int64_t r1, int64_t r2>
  friend bool operator<(const nas_currency_t<r1> &l,
                        const nas_currency_t<r2> &r) {
    return l.m_value < r.m_value;
  }
  template <int64_t r1, int64_t r2>
  friend bool operator>(const nas_currency_t<r1> &lhs,
                        const nas_currency_t<r2> &rhs) {
    return rhs < lhs;
  }
  template <int64_t r1, int64_t r2>
  friend bool operator<=(const nas_currency_t<r1> &lhs,
                         const nas_currency_t<r2> &rhs) {
    return !(lhs > rhs);
  }
  template <int64_t r1, int64_t r2>
  friend bool operator>=(const nas_currency_t<r1> &lhs,
                         const nas_currency_t<r2> &rhs) {
    return !(lhs < rhs);
  }
  template <int64_t r1, int64_t r2>
  friend bool operator==(const nas_currency_t<r1> &lhs,
                         const nas_currency_t<r2> &rhs) {
    return lhs.m_value == rhs.m_value;
  }
  template <int64_t r1, int64_t r2>
  friend bool operator!=(const nas_currency_t<r1> &lhs,
                         const nas_currency_t<r2> &rhs) {
    return lhs.m_value != rhs.m_value;
  }

  template <int64_t r1, int64_t r2>
  friend nas_currency_t<r1> operator+(const nas_currency_t<r1> &l,
                                      const nas_currency_t<r2> &r) {
    nas_currency_t<r1> tmp;
    tmp.m_value = l.m_value + r.m_value;
    return tmp;
  }

  template <int64_t r1, int64_t r2>
  friend nas_currency_t<r1> operator-(const nas_currency_t<r1> &l,
                                      const nas_currency_t<r2> &r) {
    nas_currency_t<r1> tmp;
    tmp.m_value = l.m_value + r.m_value;
    return tmp;
  }

  nas_currency_value_t wei_value() const { return m_value; }

  long double value() const {
    return m_value.convert_to<long double>() / ratio;
  }

protected:
  nas_currency_value_t m_value;

}; // end class nas_currency_unit

typedef nas_currency_t<1000000000000000000> nas;
typedef nas_currency_t<1> wei;

template <int64_t r1, int64_t r2>
nas_currency_t<r1> nas_cast(const nas_currency_t<r2> &v) {
  return nas_currency_t<r1>(v);
}

template <class T> struct is_nas_currency { const static bool value = false; };
template <int64_t r> struct is_nas_currency<nas_currency_t<r>> {
  const static bool value = true;
};

template <class T1, class T2>
auto nas_cast(const T2 &v) -> typename std::enable_if<
    is_nas_currency<T1>::value && is_nas_currency<T2>::value, T1>::type {
  return T1(v);
}

typedef fix_bytes<16> nas_storage_t;

template <typename T> nas_storage_t nas_to_storage(const T &v) {
  nas::nas_currency_value_t cv = v.wei_value();

  uint64_t high = static_cast<uint64_t>(cv >> 64);
  nas::nas_currency_value_t mask(0xFFFFFFFFFFFFFFFF);
  auto lm = cv & mask;
  uint64_t low = static_cast<uint64_t>(lm);
  nas_storage_t ret;
  uint64_t *low_ptr = (uint64_t *)ret.value();
  *low_ptr = boost::endian::native_to_big(high);
  uint64_t *high_ptr = (uint64_t *)(ret.value() + 8);
  *high_ptr = boost::endian::native_to_big(low);

  return ret;
}

template <typename T> T storage_to_nas(const nas_storage_t &v) {
  uint64_t low;
  uint64_t high;
  low = *(uint64_t *)(v.value());
  high = *(uint64_t *)(v.value() + 8);

  low = boost::endian::big_to_native(low);
  high = boost::endian::big_to_native(high);
  nas::nas_currency_value_t cv(low);
  cv = cv << 64;
  cv += nas::nas_currency_value_t(high);

  return nas(cv);
}

inline neb::bytes wei_to_storage(const wei_t &v) {
  return from_fix_bytes(nas_to_storage(nas(v)));
}
inline wei_t storage_to_wei(const neb::bytes &v) {
  return storage_to_nas<nas>(to_fix_bytes<nas_storage_t>(v)).wei_value();
}

inline neb::floatxx_t wei_to_nas(const neb::floatxx_t wei) {
  static uint64_t ratio = 1000000000000000000ULL;
  return wei / neb::floatxx_t(ratio);
}
inline neb::floatxx_t nas_to_wei(const neb::floatxx_t nas) {
  static uint64_t ratio = 1000000000000000000ULL;
  return nas * neb::floatxx_t(ratio);
}
} // end namespace neb

neb::nas operator"" _nas(long double x);
neb::nas operator"" _nas(const char *s);

neb::wei operator"" _wei(long double x);
neb::wei operator"" _wei(const char *s);

std::ostream &operator<<(std::ostream &os, const neb::nas &obj);

std::ostream &operator<<(std::ostream &os, const neb::wei &obj);
