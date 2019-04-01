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
#include "common/math/softfloat.hpp"
#include <functional>

#define MATH_MIN 1e-5

namespace neb {
namespace math {

namespace internal {

template <typename T> T __get_pi() {
  T zero = softfloat_cast<uint32_t, typename T::value_type>(0);
  T one = softfloat_cast<uint32_t, typename T::value_type>(1);
  T two = softfloat_cast<uint32_t, typename T::value_type>(2);
  T four = softfloat_cast<uint32_t, typename T::value_type>(4);

  T ret = zero;
  T i = one;
  bool odd = true;

  while (true) {
    T tmp;
    auto x = four / i;
    if (odd) {
      tmp = ret + x;
    } else {
      tmp = ret - x;
    }
    if (tmp - ret < MATH_MIN && ret - tmp < MATH_MIN) {
      break;
    }
    ret = tmp;
    i += two;
    odd = !odd;
  }
  return ret;
}

template <typename T> T __fast_get_pi() {
  T one = softfloat_cast<uint32_t, typename T::value_type>(1);
  T two = softfloat_cast<uint32_t, typename T::value_type>(2);
  T four = softfloat_cast<uint32_t, typename T::value_type>(4);

  T ret = one + two;
  T i = two;
  bool odd = true;
  while (true) {
    T tmp;
    auto tail = four / (i * (i + one) * (i + two));
    if (odd) {
      tmp = ret + tail;
    } else {
      tmp = ret - tail;
    }
    if (tmp - ret < MATH_MIN && ret - tmp < MATH_MIN) {
      break;
    }
    ret = tmp;
    i += two;
    odd = !odd;
  }
  return ret;
}

template <typename T> T __get_e() {
  T one = softfloat_cast<uint32_t, typename T::value_type>(1);
  T ret = one;

  T i = one;
  T prev = one;

  while (true) {
    T tmp;

    tmp = ret + one / prev;
    if (tmp - ret < MATH_MIN && ret - tmp < MATH_MIN) {
      break;
    }
    ret = tmp;
    i += one;
    prev = prev * i;
  }

  return ret;
}

template <typename T> T __get_ln2() {
  T zero = softfloat_cast<uint32_t, typename T::value_type>(0);
  T one = softfloat_cast<uint32_t, typename T::value_type>(1);

  T ret = zero;
  T i = one;
  bool odd = true;

  while (true) {
    T tmp;
    if (odd) {
      tmp = ret + one / i;
    } else {
      tmp = ret - one / i;
    }

    if (tmp - ret < MATH_MIN && ret - tmp < MATH_MIN) {
      break;
    }
    ret = tmp;
    i += one;
    odd = !odd;
  }
  return ret;
}

template <typename T> T __fast_get_ln2() {
  T zero = softfloat_cast<uint32_t, typename T::value_type>(0);
  T one = softfloat_cast<uint32_t, typename T::value_type>(1);
  T two = softfloat_cast<uint32_t, typename T::value_type>(2);

  auto func = [&](T x) {
    T ret = zero;
    T s = two * x;
    T i = one;
    T x2 = x * x;

    while (true) {
      T tmp;

      tmp = ret + s / i;
      if (tmp - ret < MATH_MIN && ret - tmp < MATH_MIN) {
        break;
      }

      ret = tmp;
      i += two;
      s = s * x2;
    }
    return ret;
  };

  return func((two - one) / (two + one));
}

} // namespace internal

template <typename T> class constants {
public:
  static T pi() {
    std::call_once(s_init_once, std::bind(constants<T>::init));
    return s_pi;
  }
  static T ln2() {
    std::call_once(s_init_once, std::bind(constants<T>::init));
    return s_ln2;
  }
  static T e() {
    std::call_once(s_init_once, std::bind(constants<T>::init));
    return s_e;
  }

protected:
  static void init() {
    s_pi = internal::__fast_get_pi<T>();
    s_ln2 = internal::__fast_get_ln2<T>();
    s_e = internal::__get_e<T>();
  }
  static T s_pi;
  static T s_ln2;
  static T s_e;
  static std::once_flag s_init_once;
};

template <typename T> T constants<T>::s_pi;
template <typename T> T constants<T>::s_ln2;
template <typename T> T constants<T>::s_e;
template <typename T> std::once_flag constants<T>::s_init_once;
} // namespace math
} // namespace neb
