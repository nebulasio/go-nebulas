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

template <int64_t ratio> class nas_currency_t {
public:
  typedef boost::multiprecision::int128_t nas_currency_value_t;

  nas_currency_t() : m_value(0) {}
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
} // namespace fs
} // end namespace neb

neb::fs::nas operator"" _nas(long double x);
neb::fs::nas operator"" _nas(const char *s);

neb::fs::wei operator"" _wei(long double x);
neb::fs::wei operator"" _wei(const char *s);

std::ostream &operator<<(std::ostream &os, const neb::fs::nas &obj);

std::ostream &operator<<(std::ostream &os, const neb::fs::wei &obj);
