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
#include "common/math/internal/math_template.h"

namespace neb {
namespace math {

template <typename T> T min(const T &x, const T &y) { return x < y ? x : y; }
template <typename T> T max(const T &x, const T &y) { return x > y ? x : y; }

template <typename T> T abs(const T &x, const T &y) {
  T delta = x - y;
  return delta > 0 ? delta : 0 - delta;
}

template <typename T> std::string to_string(const T &val) {
  std::stringstream ss;
  ss << val;
  return ss.str();
}
template <typename T> T from_string(const std::string &str) {
  std::stringstream ss(str);
  T val;
  ss >> val;
  return val;
}

template <typename T> bool exit_cond(const T &x, const T &y) {
  if (x - y < MATH_MIN && y - x < MATH_MIN) {
    return true;
  }
  if (to_string(x) == std::string("inf") &&
      to_string(y) == std::string("inf")) {
    return true;
  }
  if (to_string(x) == std::string("-nan") &&
      to_string(y) == std::string("-nan")) {
    return true;
  }
  return false;
}

template <typename T> T exp(const T &x) {
  T zero = softfloat_cast<uint32_t, typename T::value_type>(0);
  T one = softfloat_cast<uint32_t, typename T::value_type>(1);

  if (x < zero) {
    return one / exp(-x);
  }

  T ret = one;
  T i = one;
  T tail = x;

  while (true) {
    T tmp;

    tmp = ret + tail;
    if (exit_cond(tmp, ret)) {
      break;
    }

    ret = tmp;
    i += one;
    tail *= x / i;
  }

  return ret;
}

template <typename T> T arctan(const T &x) {
  T zero = softfloat_cast<uint32_t, typename T::value_type>(0);
  T one = softfloat_cast<uint32_t, typename T::value_type>(1);
  T two = softfloat_cast<uint32_t, typename T::value_type>(2);
  T half_pi = constants<T>::pi() / two;

  if (x > one) {
    return half_pi - arctan(one / x);
  } else if (x < zero - one) {
    return zero - half_pi - arctan(one / x);
  }

  T x2 = x * x;
  T ret = zero;
  T i = one;
  T s = x;
  bool odd = false;

  while (true) {
    T tmp;
    if (odd) {
      tmp = ret - s / i;
    } else {
      tmp = ret + s / i;
    }
    if (exit_cond(tmp, ret)) {
      break;
    }
    ret = tmp;
    odd = !odd;
    i += two;
    s = s * x2;
  }
  return ret;
}

template <typename T> T sin(const T &x) {
  T zero = softfloat_cast<uint32_t, typename T::value_type>(0);
  if (x < zero) {
    return zero - sin(zero - x);
  }

  T one = softfloat_cast<uint32_t, typename T::value_type>(1);
  T two = softfloat_cast<uint32_t, typename T::value_type>(2);
  T double_pi = two * constants<T>::pi();
  if (x > double_pi) {
    T tmp = (x / double_pi).integer_val();
    return sin(x - tmp * double_pi);
  }

  T x2 = x * x;
  T ret = zero;
  T i = one;
  T tail = x;
  bool odd = false;

  while (true) {
    T tmp;
    if (odd) {
      tmp = ret - tail;
    } else {
      tmp = ret + tail;
    }
    if (exit_cond(tmp, ret)) {
      break;
    }
    ret = tmp;
    odd = !odd;
    tail *= (x2 / ((i + one) * (i + two)));
    i += two;
  }
  return ret;
}

template <typename T> T ln(const T &x) {
  T zero = softfloat_cast<uint32_t, typename T::value_type>(0);
  T one = softfloat_cast<uint32_t, typename T::value_type>(1);
  T two = softfloat_cast<uint32_t, typename T::value_type>(2);

  auto func = [&](T x) {
    T ret = zero;
    bool odd = true;

    T s = x;
    T i = one;

    while (true) {
      T tmp;

      if (odd) {
        tmp = ret + s / i;
      } else {
        tmp = ret - s / i;
      }
      if (exit_cond(tmp, ret)) {
        break;
      }

      ret = tmp;
      odd = !odd;
      i += one;
      s = s * x;
    }
    return ret;
  };

  if (x > two) {
    return zero - func(one / x - one);
  }
  return func(x - one);
}

template <typename T> T fast_ln(const T &x) {
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
      if (exit_cond(tmp, ret)) {
        break;
      }

      ret = tmp;
      i += two;
      s = s * x2;
    }
    return ret;
  };

  return func((x - one) / (x + one));
}

namespace internal {
union float16_detail_t {
  float16_t v;
  struct {
    uint64_t ey : 10;
    uint16_t exponent : 5;
    uint8_t sign : 1;
  } detail;
};
union float32_detail_t {
  float32_t v;
  struct {
    uint64_t fraction : 23;
    uint16_t exponent : 8;
    uint8_t sign : 1;
  } detail;
};
union float64_detail_t {
  float64_t v;
  struct {
    uint64_t fraction : 52;
    uint16_t exponent : 11;
    uint8_t sign : 1;
  } detail;
};
union float128_detail_t {
  float128_t v;
  struct {
    uint64_t fraction0 : 48;
    uint64_t fraction1 : 64;
    uint16_t exponent : 15;
    uint8_t sign : 1;
  } detail;
};

template <typename T> struct float_detail {};
template <> struct float_detail<float16> {
  typedef float16_detail_t value_type;
  constexpr static uint16_t bias = 15;
};
template <> struct float_detail<float32> {
  typedef float32_detail_t value_type;
  constexpr static uint16_t bias = 127;
};
template <> struct float_detail<float64> {
  typedef float64_detail_t value_type;
  constexpr static uint16_t bias = 1023;
};
template <> struct float_detail<float128> {
  typedef float128_detail_t value_type;
  constexpr static uint16_t bias = 16383;
};

template <typename T> struct float_math_helper {
  typedef typename T::value_type value_type;
  typedef typename float_detail<T>::value_type detail_type;

  static T get_exponent(const T &x) {
    detail_type dt;
    dt.v = (T)x;
    detail_type ret;
    ret.detail.sign = dt.detail.sign;
    ret.detail.exponent = dt.detail.exponent;
    return T(ret.v);
  }
  static T get_one_plus_fraction(const T &x) {
    detail_type dt;
    dt.v = (T)x;
    detail_type ret;
    ret.v = dt.v;
    ret.detail.exponent = detail_type::bias;
    return T(ret.v);
  }
};
} // namespace internal

template <typename T> T log2(const T &x) { return ln(x) / constants<T>::ln2(); }

//! return x^y
template <typename T> T pow(const T &x, const T &y) {
  return exp(y * fast_ln(x));
}

template <typename T> T pow(const T &x, const int64_t &y) {
  T one = softfloat_cast<uint32_t, typename T::value_type>(1);
  if (y == 0) {
    return one;
  }

  T ret = one;
  T tmp_x = x;
  uint64_t tmp_y = y > 0 ? y : -y;

  while (tmp_y) {
    if ((tmp_y & 0x1) == 1) {
      ret *= tmp_x;
    }
    tmp_x *= tmp_x;
    tmp_y >>= 1;
  }
  return y > 0 ? ret : one / ret;
}

// only applicable for float16 and float32
template <typename T> T sqrt(const T &x) {
  T one = softfloat_cast<uint32_t, typename T::value_type>(1);
  T two = softfloat_cast<uint32_t, typename T::value_type>(2);
  T one_and_half = one + one / two;
  T half_x = x / two;

  typename T::value_type tmp_x = typename T::value_type(x);
  uint32_t i = *reinterpret_cast<uint32_t *>(&tmp_x);
  i = 0x5f3759df - (i >> 1);
  tmp_x = *reinterpret_cast<typename T::value_type *>(&i);
  T sqrt_x(tmp_x);
  sqrt_x = sqrt_x * (one_and_half - half_x * sqrt_x * sqrt_x);
  sqrt_x = sqrt_x * (one_and_half - half_x * sqrt_x * sqrt_x);
  sqrt_x = sqrt_x * (one_and_half - half_x * sqrt_x * sqrt_x);
  return one / sqrt_x;
}

template <typename T> T ssqrt(const T &x) {
  typename T::value_type tmp_x = typename T::value_type(x);
  return softfloat_sqrt(tmp_x);
}
} // namespace math

} // namespace neb
