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

#include "common/common.h"
#include "common/math.h"
#include <gtest/gtest.h>
#include <random>
#define PRECESION 1e-5

template <typename T> T precesion(const T &x, float pre = PRECESION) {
  return std::fabs(T(x * pre));
}

TEST(test_common_math, constants) {
  neb::floatxx_t actual_e = neb::math::constants<neb::floatxx_t>::e();
  neb::floatxx_t actual_pi = neb::math::constants<neb::floatxx_t>::pi();
  neb::floatxx_t actual_ln2 = neb::math::constants<neb::floatxx_t>::ln2();

  float expect_e = std::exp(1.0);
  float expect_pi = std::acos(-1.0);
  float expect_ln2 = std::log(2.0);

  EXPECT_TRUE(neb::math::abs(actual_e, neb::floatxx_t(expect_e)) <
              precesion(expect_e));
  EXPECT_TRUE(neb::math::abs(actual_pi, neb::floatxx_t(expect_pi)) <
              precesion(expect_pi));
  EXPECT_TRUE(neb::math::abs(actual_ln2, neb::floatxx_t(expect_ln2)) <
              precesion(expect_ln2));
}

TEST(test_common_math, min) {
  neb::floatxx_t x(0);
  neb::floatxx_t y(0);
  EXPECT_TRUE(x == neb::math::min(x, y));
  EXPECT_TRUE(y == neb::math::min(x, y));

  x = 1;
  EXPECT_TRUE(y == neb::math::min(x, y));

  y = 2;
  EXPECT_TRUE(x == neb::math::min(x, y));

  y = -1;
  EXPECT_TRUE(y == neb::math::min(x, y));

  x = -2;
  EXPECT_TRUE(x == neb::math::min(x, y));

  x = std::numeric_limits<int64_t>::min();
  EXPECT_TRUE(x == neb::math::min(x, y));

  x = std::numeric_limits<int64_t>::max();
  EXPECT_TRUE(y == neb::math::min(x, y));

  y = std::numeric_limits<int32_t>::max();
  EXPECT_TRUE(y == neb::math::min(x, y));

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(std::numeric_limits<int32_t>::min(),
                                      std::numeric_limits<int32_t>::max());
  for (auto i = 0; i < 1000; i++) {
    x = dis(mt);
    y = dis(mt);
    if (x < y) {
      EXPECT_TRUE(x == neb::math::min(x, y));
    } else if (x > y) {
      EXPECT_TRUE(y == neb::math::min(x, y));
    } else {
      EXPECT_TRUE(x == neb::math::min(x, y));
      EXPECT_TRUE(y == neb::math::min(x, y));
    }
  }
}

TEST(test_common_math, max) {
  neb::floatxx_t x(0);
  neb::floatxx_t y(0);
  EXPECT_TRUE(x == neb::math::max(x, y));
  EXPECT_TRUE(y == neb::math::max(x, y));

  x = 1;
  EXPECT_TRUE(x == neb::math::max(x, y));

  y = 2;
  EXPECT_TRUE(y == neb::math::max(x, y));

  y = -1;
  EXPECT_TRUE(x == neb::math::max(x, y));

  x = -2;
  EXPECT_TRUE(y == neb::math::max(x, y));

  x = std::numeric_limits<int64_t>::min();
  EXPECT_TRUE(y == neb::math::max(x, y));

  x = std::numeric_limits<int64_t>::max();
  EXPECT_TRUE(x == neb::math::max(x, y));

  y = std::numeric_limits<int32_t>::max();
  EXPECT_TRUE(x == neb::math::max(x, y));

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(std::numeric_limits<int32_t>::min(),
                                      std::numeric_limits<int32_t>::max());
  for (auto i = 0; i < 1000; i++) {
    x = dis(mt);
    y = dis(mt);
    if (x < y) {
      EXPECT_TRUE(y == neb::math::max(x, y));
    } else if (x > y) {
      EXPECT_TRUE(x == neb::math::max(x, y));
    } else {
      EXPECT_TRUE(x == neb::math::max(x, y));
      EXPECT_TRUE(y == neb::math::max(x, y));
    }
  }
}

TEST(test_common_math, abs) {
  neb::floatxx_t x(0);
  neb::floatxx_t y(0);
  EXPECT_TRUE((x - y) == neb::math::abs(x, y));
  EXPECT_TRUE((y - x) == neb::math::abs(x, y));

  x = 1;
  EXPECT_TRUE((x - y) == neb::math::abs(x, y));

  y = 2;
  EXPECT_TRUE((y - x) == neb::math::abs(x, y));

  y = -1;
  EXPECT_TRUE((x - y) == neb::math::abs(x, y));

  x = -2;
  EXPECT_TRUE((y - x) == neb::math::abs(x, y));

  x = std::numeric_limits<int64_t>::min();
  EXPECT_TRUE((y - x) == neb::math::abs(x, y));

  x = std::numeric_limits<int64_t>::max();
  EXPECT_TRUE((x - y) == neb::math::abs(x, y));

  y = std::numeric_limits<int32_t>::max();
  EXPECT_TRUE((x - y) == neb::math::abs(x, y));

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(std::numeric_limits<int32_t>::min(),
                                      std::numeric_limits<int32_t>::max());
  for (auto i = 0; i < 1000; i++) {
    x = dis(mt);
    y = dis(mt);
    if (x < y) {
      EXPECT_TRUE((y - x) == neb::math::abs(x, y));
    } else if (x > y) {
      EXPECT_TRUE((x - y) == neb::math::abs(x, y));
    } else {
      EXPECT_TRUE((x - y) == neb::math::abs(x, y));
      EXPECT_TRUE((y - x) == neb::math::abs(x, y));
    }
  }
}

TEST(test_common_math, to_string) {
  float x = 0;
  neb::floatxx_t y = x;
  std::ostringstream oss;
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("0"));

  x = 1;
  y = x;
  oss.str(std::string());
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("1"));

  x = -1;
  y = x;
  oss.str(std::string());
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("-1"));

  x = 123456;
  y = x;
  oss.str(std::string());
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("123456"));

  x = 1234567;
  y = x;
  oss.str(std::string());
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("1.23457e+06"));

  x = -456789;
  y = x;
  oss.str(std::string());
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("-456789"));

  x = -4567890;
  y = x;
  oss.str(std::string());
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("-4.56789e+06"));

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(std::numeric_limits<int32_t>::min(),
                                      std::numeric_limits<int32_t>::max());
  for (auto i = 0; i < 1000; i++) {
    x = dis(mt);
    y = x;
    oss.str(std::string());
    oss << x;
    EXPECT_EQ(neb::math::to_string(y), oss.str());
  }
}

TEST(test_common_math, from_string) {
  float x = 0;
  std::ostringstream oss;
  oss << x;
  std::string x_str = oss.str();
  neb::floatxx_t y = neb::math::from_string<neb::floatxx_t>(x_str);
  EXPECT_EQ(std::stof(x_str), float(y));

  x = 1;
  oss.str(std::string());
  oss << x;
  x_str = oss.str();
  y = neb::math::from_string<neb::floatxx_t>(x_str);
  EXPECT_EQ(std::stof(x_str), float(y));

  x = -1;
  oss.str(std::string());
  oss << x;
  x_str = oss.str();
  y = neb::math::from_string<neb::floatxx_t>(x_str);
  EXPECT_EQ(std::stof(x_str), float(y));

  x = 123456;
  oss.str(std::string());
  oss << x;
  x_str = oss.str();
  y = neb::math::from_string<neb::floatxx_t>(x_str);
  EXPECT_EQ(std::stof(x_str), float(y));

  x = 1234567;
  oss.str(std::string());
  oss << x;
  x_str = oss.str();
  y = neb::math::from_string<neb::floatxx_t>(x_str);
  EXPECT_EQ(std::stof(x_str), float(y));

  x = -456789;
  oss.str(std::string());
  oss << x;
  x_str = oss.str();
  y = neb::math::from_string<neb::floatxx_t>(x_str);
  EXPECT_EQ(std::stof(x_str), float(y));

  x = -4567890;
  oss.str(std::string());
  oss << x;
  x_str = oss.str();
  y = neb::math::from_string<neb::floatxx_t>(x_str);
  EXPECT_EQ(std::stof(x_str), float(y));

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(std::numeric_limits<int32_t>::min(),
                                      std::numeric_limits<int32_t>::max());
  for (auto i = 0; i < 1000; i++) {
    x = dis(mt);
    oss.str(std::string());
    oss << x;
    x_str = oss.str();
    y = neb::math::from_string<neb::floatxx_t>(x_str);
    EXPECT_EQ(std::stof(x_str), float(y));
  }
}

TEST(test_common_math, exp) {
  EXPECT_EQ(neb::math::exp(neb::floatxx_t(0)), 1);

  neb::floatxx_t actual_x = neb::math::exp(neb::floatxx_t(1));
  float expect_x = std::exp(1.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x));

  actual_x = neb::math::exp(neb::floatxx_t(-1));
  expect_x = std::exp(-1.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x));

  actual_x = neb::math::exp(neb::floatxx_t(80));
  expect_x = std::exp(80.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x));

  float delta(0.01);
  for (float x = -50.0; x <= 50.0; x += delta) {
    auto actual_x = neb::math::exp(neb::floatxx_t(x));
    auto expect_x = std::exp(x);
    EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                precesion(expect_x));
  }
}

TEST(test_common_math, arctan) {
  EXPECT_EQ(neb::math::arctan(neb::floatxx_t(0)), std::atan(0));

  neb::floatxx_t actual_x = neb::math::arctan(neb::floatxx_t(1));
  float expect_x = std::atan(1.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x));

  actual_x = neb::math::arctan(neb::floatxx_t(-1));
  expect_x = std::atan(-1.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x));

  float delta(0.01);
  for (float x = -100.0; x <= 100.0; x += delta) {
    auto actual_x = neb::math::arctan(neb::floatxx_t(x));
    auto expect_x = std::atan(x);
    if (std::fabs(x) < 0.1) {
      ASSERT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                  precesion(expect_x, 1e2 * PRECESION));
    } else if (std::fabs(x) < 1.0) {
      EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                  precesion(expect_x, 1e1 * PRECESION));
    } else {
      EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                  precesion(expect_x, 1e1 * PRECESION));
    }
  }
}

TEST(test_common_math, sin) {
  EXPECT_EQ(neb::math::sin(neb::floatxx_t(0)), std::sin(0));

  neb::floatxx_t actual_x = neb::math::sin(neb::floatxx_t(1));
  float expect_x = std::sin(1.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x));

  actual_x = neb::math::sin(neb::floatxx_t(-1));
  expect_x = std::sin(-1.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x));

  float pi = std::acos(-1.0);
  actual_x = neb::math::sin(neb::floatxx_t(pi / 2.0));
  expect_x = std::sin(pi / 2.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x));

  actual_x = neb::math::sin(neb::floatxx_t(pi / 3.0));
  expect_x = std::sin(pi / 3.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x));

  actual_x = neb::math::sin(neb::floatxx_t(-pi / 4.0));
  expect_x = std::sin(-pi / 4.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x));

  actual_x = neb::math::sin(neb::floatxx_t(-pi / 6.0));
  expect_x = std::sin(-pi / 6.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x));

  for (int32_t t = -100; t <= 100; t++) {
    auto actual_x = neb::math::sin(neb::floatxx_t(pi * t));
    auto expect_x = std::sin(pi * t);
    EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                precesion(1.0f, 1e2 * PRECESION));
  }

  float delta(0.01);
  for (float x = -100.0; x <= 100.0; x += delta) {
    auto actual_x = neb::math::sin(neb::floatxx_t(x));
    auto expect_x = std::sin(x);
    EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                precesion(1.0f, 1e2 * PRECESION));
  }
}

TEST(test_common_math, ln) {
  EXPECT_EQ(neb::math::ln(neb::floatxx_t(1)), std::log(1));

  neb::floatxx_t actual_x = neb::math::ln(neb::floatxx_t(0.1));
  float expect_x = std::log(0.1);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x, 1e1 * PRECESION));

  actual_x = neb::math::ln(neb::floatxx_t(0.5));
  expect_x = std::log(0.5);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x, 1e1 * PRECESION));

  actual_x = neb::math::ln(neb::floatxx_t(2.0));
  expect_x = std::log(2.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x, 1e1 * PRECESION));

  float delta(0.01);
  for (float x = delta; x < 1.0; x += delta) {
    auto actual_x = neb::math::ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    if (std::fabs(expect_x) < PRECESION) {
      continue;
    }

    if (std::fabs(x) < 0.01 || std::fabs(expect_x) < 0.01) {
      EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                  precesion(expect_x, 1e3 * PRECESION));
    } else if (std::fabs(x) < 0.1 || std::fabs(expect_x) < 0.1) {
      EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                  precesion(expect_x, 1e2 * PRECESION));
    } else {
      EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                  precesion(expect_x, 1e1 * PRECESION));
    }
  }

  delta = 1.0;
  for (float x = 1.0 + delta; x < 100.0; x += delta) {
    auto actual_x = neb::math::ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                precesion(expect_x, 1e2 * PRECESION));
  }

  delta = 10.0;
  for (float x = 1.0 + delta; x < 1000.0; x += delta) {
    auto actual_x = neb::math::ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                precesion(expect_x, 1e3 * PRECESION));
  }

  delta = 100.0;
  for (float x = 1.0 + delta; x < 10000.0; x += delta) {
    auto actual_x = neb::math::ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                precesion(expect_x, 1e3 * PRECESION));
  }

  delta = 1000.0;
  // for (float x = 1.0 + delta; x < 100000.0; x += delta) {
  // auto actual_x = neb::math::ln(neb::floatxx_t(x));
  // auto expect_x = std::log(x);
  // EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
  // precesion(expect_x, 1e4 * PRECESION));
  //}

  // takes time
  delta = 10000.0;
  // for (float x = 1.0 + delta; x < 1000000.0; x += delta) {
  // auto actual_x = neb::math::ln(neb::floatxx_t(x));
  // auto expect_x = std::log(x);
  // EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
  // precesion(expect_x, 1e5 * PRECESION));
  //}
}

TEST(test_common_math, fast_ln) {
  EXPECT_EQ(neb::math::fast_ln(neb::floatxx_t(1)), std::log(1));

  neb::floatxx_t actual_x = neb::math::fast_ln(neb::floatxx_t(0.1));
  float expect_x = std::log(0.1);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x, 1e1 * PRECESION));

  actual_x = neb::math::fast_ln(neb::floatxx_t(0.5));
  expect_x = std::log(0.5);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x, PRECESION));

  actual_x = neb::math::fast_ln(neb::floatxx_t(2.0));
  expect_x = std::log(2.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
              precesion(expect_x, PRECESION));

  float delta(0.01);
  for (float x = delta; x < 1.0; x += delta) {
    auto actual_x = neb::math::fast_ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    if (std::fabs(expect_x) < PRECESION) {
      continue;
    }

    if (std::fabs(x) < 0.1 || std::fabs(expect_x) < 0.1) {
      EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                  precesion(expect_x, 1e2 * PRECESION));
    } else {
      EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                  precesion(expect_x, 1e1 * PRECESION));
    }
  }

  delta = 1.0;
  for (float x = 1.0 + delta; x < 100.0; x += delta) {
    auto actual_x = neb::math::fast_ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                precesion(expect_x, 1e1 * PRECESION));
  }

  delta = 10.0;
  for (float x = 1.0 + delta; x < 1000.0; x += delta) {
    auto actual_x = neb::math::fast_ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                precesion(expect_x, 1e2 * PRECESION));
  }

  delta = 100.0;
  for (float x = 1.0 + delta; x < 10000.0; x += delta) {
    auto actual_x = neb::math::fast_ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                precesion(expect_x, 1e3 * PRECESION));
  }

  delta = 1000.0;
  // for (float x = 1.0 + delta; x < 100000.0; x += delta) {
  // auto actual_x = neb::math::fast_ln(neb::floatxx_t(x));
  // auto expect_x = std::log(x);
  // EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
  // precesion(expect_x, 1e4 * PRECESION));
  //}

  // takes time
  delta = 10000.0;
  // for (float x = 1.0 + delta; x < 1000000.0; x += delta) {
  // auto actual_x = neb::math::fast_ln(neb::floatxx_t(x));
  // auto expect_x = std::log(x);
  // EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
  // precesion(expect_x, 1e4 * PRECESION));
  //}
}

TEST(test_common_math, pow_int_y) {
  float x = 0.0;
  int64_t y = 0;
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), 1);
  EXPECT_EQ(std::pow(x, y), 1);
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), std::pow(x, y));

  x = 1.0;
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), 1);
  EXPECT_EQ(std::pow(x, y), 1);
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), std::pow(x, y));

  x = -1.0;
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), 1);
  EXPECT_EQ(std::pow(x, y), 1);
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), std::pow(x, y));

  x = 123.456;
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), 1);
  EXPECT_EQ(std::pow(x, y), 1);
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), std::pow(x, y));

  x = -456.789;
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), 1);
  EXPECT_EQ(std::pow(x, y), 1);
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), std::pow(x, y));

  x = std::numeric_limits<int32_t>::min();
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), 1);
  EXPECT_EQ(std::pow(x, y), 1);
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), std::pow(x, y));

  x = std::numeric_limits<int32_t>::max();
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), 1);
  EXPECT_EQ(std::pow(x, y), 1);
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), std::pow(x, y));

  x = std::numeric_limits<int64_t>::min();
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), 1);
  EXPECT_EQ(std::pow(x, y), 1);
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), std::pow(x, y));

  x = std::numeric_limits<int64_t>::max();
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), 1);
  EXPECT_EQ(std::pow(x, y), 1);
  EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), std::pow(x, y));

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(std::numeric_limits<int32_t>::min(),
                                      std::numeric_limits<int32_t>::max());
  for (auto i = 0; i < 1000; i++) {
    x = dis(mt);
    EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), 1);
    EXPECT_EQ(std::pow(x, y), 1);
    EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), std::pow(x, y));
  }

  x = 0.0;
  for (auto i = 0; i < 1000; i++) {
    std::uniform_int_distribution<> dis_y(1,
                                          std::numeric_limits<int32_t>::max());
    y = dis_y(mt);
    EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), 0);
    EXPECT_EQ(std::pow(x, y), 0);
    EXPECT_EQ(neb::math::pow(neb::floatxx_t(x), y), std::pow(x, y));
  }

  for (auto i = 0; i < 1000; i++) {
    std::uniform_int_distribution<> dis_x(-80, 80);
    std::uniform_int_distribution<> dis_y(-15, 15);
    x = dis_x(mt);
    y = dis_y(mt);
    if (x == 0) {
      continue;
    }

    auto actual_pow = neb::math::pow(neb::floatxx_t(x), y);
    auto expect_pow = std::pow(x, y);
    EXPECT_TRUE(neb::math::abs(actual_pow, neb::floatxx_t(expect_pow)) <
                precesion(expect_pow));
  }
}

TEST(test_common_math, sqrt) {
  float x = 0.0;
  auto actual_sqrt = neb::math::sqrt(neb::floatxx_t(x));
  auto expect_sqrt = std::sqrt(x);
  EXPECT_TRUE(neb::math::abs(actual_sqrt, neb::floatxx_t(expect_sqrt)) <
              precesion(1.0f, 1e-9 * PRECESION));

  x = 1.0;
  actual_sqrt = neb::math::sqrt(neb::floatxx_t(x));
  expect_sqrt = std::sqrt(x);
  EXPECT_TRUE(neb::math::abs(actual_sqrt, neb::floatxx_t(expect_sqrt)) <
              precesion(expect_sqrt, 1e-1 * PRECESION));

  x = 0.5;
  actual_sqrt = neb::math::sqrt(neb::floatxx_t(x));
  expect_sqrt = std::sqrt(x);
  EXPECT_TRUE(neb::math::abs(actual_sqrt, neb::floatxx_t(expect_sqrt)) <
              precesion(expect_sqrt, 1e-1 * PRECESION));

  x = 2.0;
  actual_sqrt = neb::math::sqrt(neb::floatxx_t(x));
  expect_sqrt = std::sqrt(x);
  EXPECT_TRUE(neb::math::abs(actual_sqrt, neb::floatxx_t(expect_sqrt)) <
              precesion(expect_sqrt, 1e-1 * PRECESION));

  float delta(0.001);
  for (x = delta; x < 1.0; x += delta) {
    actual_sqrt = neb::math::sqrt(neb::floatxx_t(x));
    expect_sqrt = std::sqrt(x);
    EXPECT_TRUE(neb::math::abs(actual_sqrt, neb::floatxx_t(expect_sqrt)) <
                precesion(expect_sqrt, 1e-1 * PRECESION));
  }

  delta = 0.1;
  for (x = 1.0 + delta; x < 1000.0; x += delta) {
    actual_sqrt = neb::math::sqrt(neb::floatxx_t(x));
    expect_sqrt = std::sqrt(x);
    EXPECT_TRUE(neb::math::abs(actual_sqrt, neb::floatxx_t(expect_sqrt)) <
                precesion(expect_sqrt, 1e-1 * PRECESION));
  }

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(0, std::numeric_limits<int32_t>::max());
  for (auto i = 0; i < 1000; i++) {
    x = dis(mt);
    actual_sqrt = neb::math::sqrt(neb::floatxx_t(x));
    expect_sqrt = std::sqrt(x);
    EXPECT_TRUE(neb::math::abs(actual_sqrt, neb::floatxx_t(expect_sqrt)) <
                precesion(expect_sqrt, 1e-1 * PRECESION));
  }
}
