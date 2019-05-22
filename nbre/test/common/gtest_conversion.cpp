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

#include "common/int128_conversion.h"
#include <gtest/gtest.h>
#include <random>

#define PRECESION 1e-5
template <typename T> T precesion(const T &x, float pre = PRECESION) {
  return std::fabs(T(x * pre));
}

TEST(test_common_conversion, basic) {
  neb::internal::int128_conversion_helper cv;
  EXPECT_EQ(cv.data(), 0);
  EXPECT_EQ(cv.high(), 0);
  EXPECT_EQ(cv.low(), 0);

  neb::int128_t data = boost::lexical_cast<neb::int128_t>("0");
  neb::internal::int128_conversion_helper cv_0(data);
  EXPECT_EQ(cv_0.data(), 0);
  EXPECT_EQ(cv_0.high(), 0);
  EXPECT_EQ(cv_0.low(), 0);

  data = boost::lexical_cast<neb::int128_t>("1");
  neb::internal::int128_conversion_helper cv_1(data);
  EXPECT_EQ(cv_1.data(), 1);
  EXPECT_EQ(cv_1.high(), 0);
  EXPECT_EQ(cv_1.low(), 1);

  int64_t high = 1;
  uint64_t low = 1;
  neb::internal::int128_conversion_helper tmp;
  tmp.high() = high;
  tmp.low() = low;
  EXPECT_EQ(tmp.data(),
            boost::lexical_cast<neb::int128_t>("18446744073709551617"));
  neb::internal::int128_conversion_helper cv_3(tmp.data());
  EXPECT_EQ(cv_3.data(), tmp.data());
  EXPECT_EQ(cv_3.high(), tmp.high());
  EXPECT_EQ(cv_3.low(), tmp.low());

  high = 123456;
  low = 456789;
  tmp.high() = high;
  tmp.low() = low;
  EXPECT_EQ(tmp.data(),
            boost::lexical_cast<neb::int128_t>("2277361236363886404761685"));
  neb::internal::int128_conversion_helper cv_4(tmp.data());
  EXPECT_EQ(cv_4.data(), tmp.data());
  EXPECT_EQ(cv_4.high(), tmp.high());
  EXPECT_EQ(cv_4.low(), tmp.low());
}

TEST(test_common_conversion, to_float) {
  neb::internal::int128_conversion_helper cv;
  auto f_data = cv.to_float<neb::floatxx_t>();
  EXPECT_EQ(f_data, 0);

  auto data = boost::lexical_cast<neb::int128_t>("1");
  neb::internal::int128_conversion_helper cv_1(data);
  f_data = cv_1.to_float<neb::floatxx_t>();
  EXPECT_EQ(f_data, 1);

  int64_t high = 1;
  uint64_t low = 1;
  neb::internal::int128_conversion_helper tmp;
  tmp.high() = high;
  tmp.low() = low;
  neb::internal::int128_conversion_helper cv_3(tmp.data());
  f_data = cv_3.to_float<neb::floatxx_t>();
  std::ostringstream oss;
  oss << boost::lexical_cast<float>(tmp.data());
  EXPECT_EQ(oss.str(), std::string("1.84467e+19"));
  EXPECT_EQ(neb::math::to_string(f_data), oss.str());

  high = 123456;
  low = 456789;
  tmp.high() = high;
  tmp.low() = low;
  neb::internal::int128_conversion_helper cv_4(tmp.data());
  f_data = cv_4.to_float<neb::floatxx_t>();
  oss.str(std::string());
  oss << boost::lexical_cast<float>(tmp.data());
  EXPECT_EQ(oss.str(), std::string("2.27736e+24"));
  EXPECT_EQ(neb::math::to_string(f_data), oss.str());

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(0, std::numeric_limits<int32_t>::max() %
                                             (1 << 20));
  for (auto i = 0; i < 1000; i++) {
    high = dis(mt);
    low = dis(mt);
    tmp.high() = high;
    tmp.low() = low;
    neb::internal::int128_conversion_helper cv(tmp.data());
    f_data = cv.to_float<neb::floatxx_t>();
    oss.str(std::string());
    oss << boost::lexical_cast<float>(tmp.data());
    EXPECT_EQ(neb::math::to_string(f_data), oss.str());
  }
}

TEST(test_common_conversion, from_float) {
  uint64_t x = 0;
  float y = x;
  neb::internal::int128_conversion_helper cv;
  EXPECT_EQ(cv.from_float(neb::floatxx_t(y)), 0);

  x = 1;
  y = x;
  EXPECT_EQ(cv.from_float(neb::floatxx_t(y)), 1);

  x = 10000000000ULL;
  y = x;
  y = y * x;
  auto actual_y = cv.from_float(neb::floatxx_t(y));
  auto expect_y = boost::lexical_cast<neb::int128_t>("100000000000000000000");
  EXPECT_TRUE(boost::lexical_cast<float>(neb::math::abs(actual_y, expect_y)) <
              precesion(y, 1e-6));

  x = 12345600000ULL;
  y = x;
  y = y * 1e15;
  actual_y = cv.from_float(neb::floatxx_t(y));
  expect_y = boost::lexical_cast<neb::int128_t>("12345600000000000000000000");
  EXPECT_TRUE(boost::lexical_cast<float>(neb::math::abs(actual_y, expect_y)) <
              precesion(y, 1e-6));

  x = 45678900000ULL;
  y = x;
  y = y * 1e20;
  actual_y = cv.from_float(neb::floatxx_t(y));
  expect_y =
      boost::lexical_cast<neb::int128_t>("4567890000000000000000000000000");
  EXPECT_TRUE(boost::lexical_cast<float>(neb::math::abs(actual_y, expect_y)) <
              precesion(y, 1e-6));

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(0, std::numeric_limits<int32_t>::max());
  for (auto i = 0; i < 1000; i++) {
    x = dis(mt);
    y = x;
    expect_y = x;
    x = dis(mt);
    y = y * x;
    expect_y = expect_y * x;
    actual_y = cv.from_float(neb::floatxx_t(y));
    EXPECT_TRUE(boost::lexical_cast<float>(neb::math::abs(actual_y, expect_y)) <
                precesion(y, 1e-6));
  }
}
