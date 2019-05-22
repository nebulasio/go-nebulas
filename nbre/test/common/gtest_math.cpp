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

neb::floatxx_t zero =
    softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(0);
neb::floatxx_t one =
    softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(1);
neb::floatxx_t two =
    softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(2);
neb::floatxx_t three =
    softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(3);
neb::floatxx_t four =
    softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(4);
neb::floatxx_t five =
    softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(5);
neb::floatxx_t six =
    softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(6);
neb::floatxx_t ten =
    softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(10);

template <typename T> T precesion(const T &x, float pre = PRECESION) {
  return std::fabs(T(x * pre));
}

template <typename T> std::string mem_bytes(T x) {
  auto buf = reinterpret_cast<unsigned char *>(&x);
  std::stringstream ss;
  for (auto i = 0; i < sizeof(x); i++) {
    ss << std::hex << std::setw(2) << std::setfill('0')
       << static_cast<unsigned int>(buf[i]);
  }
  return ss.str();
}

TEST(test_common_math, constants) {
  neb::floatxx_t actual_e = neb::math::constants<neb::floatxx_t>::e();
  neb::floatxx_t actual_pi = neb::math::constants<neb::floatxx_t>::pi();
  neb::floatxx_t actual_ln2 = neb::math::constants<neb::floatxx_t>::ln2();

  EXPECT_EQ(mem_bytes(actual_e), "48f82d40");
  EXPECT_EQ(mem_bytes(actual_pi), "c80f4940");
  EXPECT_EQ(mem_bytes(actual_ln2), "0572313f");

  float expect_e = std::exp(1.0);
  float expect_pi = std::acos(-1.0);
  float expect_ln2 = std::log(2.0);

  auto ret = neb::math::abs(actual_e, neb::floatxx_t(expect_e));
  EXPECT_TRUE(ret < precesion(expect_e));

  ret = neb::math::abs(actual_pi, neb::floatxx_t(expect_pi));
  EXPECT_TRUE(ret < precesion(expect_pi));

  ret = neb::math::abs(actual_ln2, neb::floatxx_t(expect_ln2));
  EXPECT_TRUE(ret < precesion(expect_ln2));
}

TEST(test_common_math, min) {
  neb::floatxx_t x(zero);
  neb::floatxx_t y(zero);
  EXPECT_TRUE(x == neb::math::min(x, y));
  EXPECT_TRUE(y == neb::math::min(x, y));
  EXPECT_EQ(mem_bytes(x), "00000000");
  EXPECT_EQ(mem_bytes(y), "00000000");
  EXPECT_EQ(mem_bytes(neb::math::min(x, y)), "00000000");

  x = one;
  EXPECT_TRUE(y == neb::math::min(x, y));
  EXPECT_EQ(mem_bytes(x), "0000803f");
  EXPECT_EQ(mem_bytes(y), "00000000");
  EXPECT_EQ(mem_bytes(neb::math::min(x, y)), "00000000");

  y = two;
  EXPECT_TRUE(x == neb::math::min(x, y));
  EXPECT_EQ(mem_bytes(x), "0000803f");
  EXPECT_EQ(mem_bytes(y), "00000040");
  EXPECT_EQ(mem_bytes(neb::math::min(x, y)), "0000803f");

  y = zero - one;
  EXPECT_TRUE(y == neb::math::min(x, y));
  EXPECT_EQ(mem_bytes(x), "0000803f");
  EXPECT_EQ(mem_bytes(y), "000080bf");
  EXPECT_EQ(mem_bytes(neb::math::min(x, y)), "000080bf");

  x = zero - two;
  EXPECT_TRUE(x == neb::math::min(x, y));
  EXPECT_EQ(mem_bytes(x), "000000c0");
  EXPECT_EQ(mem_bytes(y), "000080bf");
  EXPECT_EQ(mem_bytes(neb::math::min(x, y)), "000000c0");

  x = softfloat_cast<int64_t, typename neb::floatxx_t::value_type>(
      std::numeric_limits<int64_t>::min());
  EXPECT_TRUE(x == neb::math::min(x, y));
  EXPECT_EQ(mem_bytes(x), "000000df");
  EXPECT_EQ(mem_bytes(y), "000080bf");
  EXPECT_EQ(mem_bytes(neb::math::min(x, y)), "000000df");

  x = softfloat_cast<int64_t, typename neb::floatxx_t::value_type>(
      std::numeric_limits<int64_t>::max());
  EXPECT_TRUE(y == neb::math::min(x, y));
  EXPECT_EQ(mem_bytes(x), "0000005f");
  EXPECT_EQ(mem_bytes(y), "000080bf");
  EXPECT_EQ(mem_bytes(neb::math::min(x, y)), "000080bf");

  y = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(
      std::numeric_limits<int32_t>::max());
  EXPECT_TRUE(y == neb::math::min(x, y));
  EXPECT_EQ(mem_bytes(x), "0000005f");
  EXPECT_EQ(mem_bytes(y), "0000004f");
  EXPECT_EQ(mem_bytes(neb::math::min(x, y)), "0000004f");

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(std::numeric_limits<int32_t>::min(),
                                      std::numeric_limits<int32_t>::max());
  for (auto i = 0; i < 1000; i++) {
    x = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(dis(mt));
    y = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(dis(mt));
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
  neb::floatxx_t x(zero);
  neb::floatxx_t y(zero);
  EXPECT_TRUE(x == neb::math::max(x, y));
  EXPECT_TRUE(y == neb::math::max(x, y));
  EXPECT_EQ(mem_bytes(x), "00000000");
  EXPECT_EQ(mem_bytes(y), "00000000");
  EXPECT_EQ(mem_bytes(neb::math::max(x, y)), "00000000");

  x = one;
  EXPECT_TRUE(x == neb::math::max(x, y));
  EXPECT_EQ(mem_bytes(x), "0000803f");
  EXPECT_EQ(mem_bytes(y), "00000000");
  EXPECT_EQ(mem_bytes(neb::math::max(x, y)), "0000803f");

  y = two;
  EXPECT_TRUE(y == neb::math::max(x, y));
  EXPECT_EQ(mem_bytes(x), "0000803f");
  EXPECT_EQ(mem_bytes(y), "00000040");
  EXPECT_EQ(mem_bytes(neb::math::max(x, y)), "00000040");

  y = zero - one;
  EXPECT_TRUE(x == neb::math::max(x, y));
  EXPECT_EQ(mem_bytes(x), "0000803f");
  EXPECT_EQ(mem_bytes(y), "000080bf");
  EXPECT_EQ(mem_bytes(neb::math::max(x, y)), "0000803f");

  x = zero - two;
  EXPECT_TRUE(y == neb::math::max(x, y));
  EXPECT_EQ(mem_bytes(x), "000000c0");
  EXPECT_EQ(mem_bytes(y), "000080bf");
  EXPECT_EQ(mem_bytes(neb::math::max(x, y)), "000080bf");

  x = softfloat_cast<int64_t, typename neb::floatxx_t::value_type>(
      std::numeric_limits<int64_t>::min());
  EXPECT_TRUE(y == neb::math::max(x, y));
  EXPECT_EQ(mem_bytes(x), "000000df");
  EXPECT_EQ(mem_bytes(y), "000080bf");
  EXPECT_EQ(mem_bytes(neb::math::max(x, y)), "000080bf");

  x = softfloat_cast<int64_t, typename neb::floatxx_t::value_type>(
      std::numeric_limits<int64_t>::max());
  EXPECT_TRUE(x == neb::math::max(x, y));
  EXPECT_EQ(mem_bytes(x), "0000005f");
  EXPECT_EQ(mem_bytes(y), "000080bf");
  EXPECT_EQ(mem_bytes(neb::math::max(x, y)), "0000005f");

  y = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(
      std::numeric_limits<int32_t>::max());
  EXPECT_TRUE(x == neb::math::max(x, y));
  EXPECT_EQ(mem_bytes(x), "0000005f");
  EXPECT_EQ(mem_bytes(y), "0000004f");
  EXPECT_EQ(mem_bytes(neb::math::max(x, y)), "0000005f");

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(std::numeric_limits<int32_t>::min(),
                                      std::numeric_limits<int32_t>::max());
  for (auto i = 0; i < 1000; i++) {
    x = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(dis(mt));
    y = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(dis(mt));
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
  neb::floatxx_t x(zero);
  neb::floatxx_t y(zero);
  EXPECT_TRUE((x - y) == neb::math::abs(x, y));
  EXPECT_TRUE((y - x) == neb::math::abs(x, y));
  EXPECT_EQ(mem_bytes(x - y), "00000000");
  EXPECT_EQ(mem_bytes(y - x), "00000000");
  EXPECT_EQ(mem_bytes(neb::math::abs(x, y)), "00000000");

  x = one;
  EXPECT_TRUE((x - y) == neb::math::abs(x, y));
  EXPECT_EQ(mem_bytes(x - y), "0000803f");
  EXPECT_EQ(mem_bytes(y - x), "000080bf");
  EXPECT_EQ(mem_bytes(neb::math::abs(x, y)), "0000803f");

  y = two;
  EXPECT_TRUE((y - x) == neb::math::abs(x, y));
  EXPECT_EQ(mem_bytes(x - y), "000080bf");
  EXPECT_EQ(mem_bytes(y - x), "0000803f");
  EXPECT_EQ(mem_bytes(neb::math::abs(x, y)), "0000803f");

  y = zero - one;
  EXPECT_TRUE((x - y) == neb::math::abs(x, y));
  EXPECT_EQ(mem_bytes(x - y), "00000040");
  EXPECT_EQ(mem_bytes(y - x), "000000c0");
  EXPECT_EQ(mem_bytes(neb::math::abs(x, y)), "00000040");

  x = zero - two;
  EXPECT_TRUE((y - x) == neb::math::abs(x, y));
  EXPECT_EQ(mem_bytes(x - y), "000080bf");
  EXPECT_EQ(mem_bytes(y - x), "0000803f");
  EXPECT_EQ(mem_bytes(neb::math::abs(x, y)), "0000803f");

  x = softfloat_cast<int64_t, typename neb::floatxx_t::value_type>(
      std::numeric_limits<int64_t>::min());
  EXPECT_TRUE((y - x) == neb::math::abs(x, y));
  EXPECT_EQ(mem_bytes(x - y), "000000df");
  EXPECT_EQ(mem_bytes(y - x), "0000005f");
  EXPECT_EQ(mem_bytes(neb::math::abs(x, y)), "0000005f");

  x = softfloat_cast<int64_t, typename neb::floatxx_t::value_type>(
      std::numeric_limits<int64_t>::max());
  EXPECT_TRUE((x - y) == neb::math::abs(x, y));
  EXPECT_EQ(mem_bytes(x - y), "0000005f");
  EXPECT_EQ(mem_bytes(y - x), "000000df");
  EXPECT_EQ(mem_bytes(neb::math::abs(x, y)), "0000005f");

  y = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(
      std::numeric_limits<int32_t>::max());
  EXPECT_TRUE((x - y) == neb::math::abs(x, y));
  EXPECT_EQ(mem_bytes(x - y), "0000005f");
  EXPECT_EQ(mem_bytes(y - x), "000000df");
  EXPECT_EQ(mem_bytes(neb::math::abs(x, y)), "0000005f");

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(std::numeric_limits<int32_t>::min(),
                                      std::numeric_limits<int32_t>::max());
  for (auto i = 0; i < 1000; i++) {
    x = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(dis(mt));
    y = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(dis(mt));
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
  int32_t xx = 0;
  float x = xx;
  neb::floatxx_t y =
      softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  std::ostringstream oss;
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("0"));
  EXPECT_EQ(mem_bytes(y), "00000000");

  xx = 1;
  x = xx;
  y = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  oss.str(std::string());
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("1"));
  EXPECT_EQ(mem_bytes(y), "0000803f");

  xx = -1;
  x = xx;
  y = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  oss.str(std::string());
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("-1"));
  EXPECT_EQ(mem_bytes(y), "000080bf");

  xx = 123456;
  x = xx;
  y = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  oss.str(std::string());
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("123456"));
  EXPECT_EQ(mem_bytes(y), "0020f147");

  xx = 1234567;
  x = xx;
  y = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  oss.str(std::string());
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("1.23457e+06"));
  EXPECT_EQ(mem_bytes(y), "38b49649");

  xx = -456789;
  x = xx;
  y = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  oss.str(std::string());
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("-456789"));
  EXPECT_EQ(mem_bytes(y), "a00adfc8");

  xx = -4567890;
  x = xx;
  y = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  oss.str(std::string());
  oss << x;
  EXPECT_EQ(neb::math::to_string(y), oss.str());
  EXPECT_EQ(neb::math::to_string(x), std::string("-4.56789e+06"));
  EXPECT_EQ(mem_bytes(y), "a4668bca");

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(std::numeric_limits<int32_t>::min(),
                                      std::numeric_limits<int32_t>::max());
  for (auto i = 0; i < 1000; i++) {
    xx = dis(mt);
    x = xx;
    y = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
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
  EXPECT_EQ(neb::math::exp(zero), 1);

  int32_t xx = 1;
  neb::floatxx_t f_xx =
      softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  neb::floatxx_t actual_x = neb::math::exp(f_xx);
  float expect_x = std::exp(xx);
  auto ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x));
  EXPECT_EQ(mem_bytes(actual_x), "48f82d40");

  xx = -1;
  f_xx = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  actual_x = neb::math::exp(f_xx);
  expect_x = std::exp(xx);
  ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x));
  EXPECT_EQ(mem_bytes(actual_x), "bf5abc3e");

  xx = 80;
  f_xx = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  actual_x = neb::math::exp(f_xx);
  expect_x = std::exp(xx);
  ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x));
  EXPECT_EQ(mem_bytes(actual_x), "d9bb2a79");

  neb::floatxx_t f_delta = one / ten;
  neb::floatxx_t f_x = zero - five * ten;

  float delta(0.1);
  for (float x = -50.0; x <= 50.0; x += delta) {
    auto actual_x = neb::math::exp(f_x);
    // std::cout << '\"' << mem_bytes(actual_x) << '\"' << ',';
    auto expect_x = std::exp(x);
    ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
    EXPECT_TRUE(ret < precesion(expect_x));
    f_x += f_delta;
  }
  // std::cout << std::endl;

  std::string mem_bytes_arr[] = {
      "f82b691b", "d6d8801b", "d7658e1b", "ac5f9d1b", "b2ecad1b", "5237c01b",
      "6f6ed41b", "bec5ea1b", "4dbb011c", "1b600f1c", "3f741e1c", "5b1e2f1c",
      "2289411c", "c1e3551c", "5d626c1c", "529f821c", "1c5c901c", "c18a9f1c",
      "2252b01c", "4fddc21c", "b35bd71c", "d201ee1c", "e784031d", "d759111d",
      "29a3201d", "0988311d", "c433441d", "32d6581d", "2fa46f1d", "0f6c841d",
      "4859921d", "7cbda11d", "14c0b21d", "9d8cc51d", "5253da1d", "5c49f11d",
      "ca54051e", "815a131e", "c2d9221e", "3cfa331e", "d7e7461e", "07d35b1e",
      "6ff1721e", "223f861e", "7f5d941e", "03f8a31e", "8a36b51e", "6645c81e",
      "6b55dd1e", "6f9cf41e", "162b071f", "4362151f", "3118251f", "0d75361f",
      "68a5491f", "68da5e1f", "5a4a761f", "a618881f", "d668961f", "5a3aa61f",
      "beb5b71f", "d907cb1f", "1062e01f", "44fbf71f", "e0070920", "30711720",
      "815e2720", "9ff83820", "b06c4c20", "79ec6120", "1eaf7920", "b2f88920",
      "5f7b9820", "af84a820", "bf3dba20", "00d4cd20", "8879e320", "ef65fb20",
      "36eb0a21", "60871921", "dcac2921", "15853b21", "b93d4f21", "61096521",
      "cd1f7d21", "5ddf8b21", "31959a21", "13d7aa21", "a9cebc21", "06aad021",
      "e99be621", "b1dcfe21", "35d50c22", "e6a41b22", "5b032c22", "801a3e22",
      "b2185222", "3c316822", "4c4e8022", "becc8d22", "74b69c22", "ae31ad22",
      "a568bf22", "fc89d322", "4fc9e922", "d32f0123", "f2c50e23", "e2c91d23",
      "12622e23", "0db94023", "c7fd5423", "39646b23", "e1128223", "e3c08f23",
      "38df9e23", "9a94af23", "c60bc223", "2a74d623", "ed01ed23", "7cf70224",
      "8ebd1024", "72f61f24", "29c93024", "d0604324", "10ed5724", "84a26e24",
      "b1dd8324", "f1bb9124", "930fa124", "e5ffb124", "3ab8c424", "9068d924",
      "f145f024", "71c50425", "15bc1225", "af2a2225", "bb383325", "f6114625",
      "b4e65a25", "3fec7125", "cdae8525", "fdbd9325", "b447a325", "bd73b425",
      "206ec725", "7267dc25", "7395f325", "c5990626", "aac11426", "ab662426",
      "e4b03526", "a3cc4826", "cbea5d26", "8d417526", "59868726", "1dc79526",
      "a087a526", "41f0b626", "882dca26", "d970df26", "a4f0f626", "87740827",
      "5dce1627", "8faa2627", "c7313827", "e3904b27", "94f96027", "a8a27827",
      "5b648927", "6dd79727", "80cfa727", "8175b927", "b4f6cc27", "f984e227",
      "a557fa27", "d9550a28", "48e21828", "72f62828", "76bb3a28", "ea5e4e28",
      "26136428", "c90f7c28", "20498b28", "3aef9928", "bc1faa28", "2004bc28",
      "3ccacf28", "c1a4e528", "b1cbfd28", "6b3e0c29", "52fe1a29", "584b2b29",
      "384f3d29", "30385129", "34396729", "968a7f29", "66358d29", "440f9c29",
      "fa78ac29", "9d9cbe29", "a1a8d229", "60d0e829", "51a6002a", "0f2e0e2a",
      "15221d2a", "b5a82d2a", "4bec3f2a", "9c1b542a", "5c6a6a2a", "e588812a",
      "71288f2a", "cb369e2a", "89daae2a", "433ec12a", "2791d52a", "3607ec2a",
      "fa6c022b", "8a24102b", "6c4d1f2b", "740e302b", "9192422b", "3e09572b",
      "d3a66d2b", "aa52832b", "6622912b", "f465a02b", "8044b12b", "38e9c32b",
      "e783d82b", "5a49ef2b", "f239042c", "f421122c", "6d80212c", "af7c322c",
      "3a42452c", "34015a2c", "b8ee702c", "c622852c", "5123932c", "dc9ca22c",
      "01b7b32c", "a49dc62c", "2181db2c", "0497f22c", "3d0d062d", "7026142d",
      "32bb232d", "78f3342d", "69fb472d", "a9035d2d", "3842742d", "51f9862d",
      "532b952d", "8edba42d", "2932b62d", "8d5bc92d", "e288de2d", "60f0f52d",
      "02e7072e", "0632162e", "e0fd252e", "0473372e", "2abe4a2e", "c510602e",
      "81a1772e", "55d6882e", "843a972e", "2b22a72e", "13b6b82e", "3723cc2e",
      "609be12e", "9655f92e", "4dc7092f", "d844182f", "8548282f", "59fb392f",
      "b18a4d2f", "a728632f", "a40c7b2f", "f3b98a2f", "0051992f", "dc70a92f",
      "dc42bb2f", "acf4ce2f", "afb8e42f", "c2c6fc2f", "3eae0b30", "fa5e1a30",
      "429b2a30", "a58c3c30", "1d615030", "7d4b6630", "ec837e30", "3ba48c30",
      "db6e9b30", "b1c7ab30", "acd8bd30", "17d0d130", "0fe1e730", "11220031",
      "e69b0d31", "8e801c31", "36f62c31", "02273f31", "8c415331", "64796931",
      "b6038131", "45958e31", "28949d31", "d126ae31", "a377c031", "9bb5d431",
      "9014eb31", "ebe60132", "62900f32", "aaa91e32", "7e592f32", "93ca4132",
      "332c5632", "84b26c32", "a9cb8232", "338d9032", "15c19f32", "488eb032",
      "de1fc332", "56a5d732", "5b53ee32", "03b20333", "c18b1133", "69da2033",
      "37c53133", "74774433", "1f215933", "11f76f33", "ee998433", "138c9233",
      "b0f5a133", "44feb233", "74d1c533", "819fda33", "a59df133", "6f830534",
      "268e1334", "e6122334", "7d393434", "d12d4734", "7e205c34", "23477334",
      "916e8634", "01929434", "1432a434", "dd76b534", "978cc834", "23a4dd34",
      "95f3f434", "4f5b0735", "9e971535", "40532535", "6ab63635", "c0ed4935",
      "782a5f35", "eda27635", "ac498835", "0f9f9635", "6276a635", "2ff8b735",
      "5a51cb35", "79b3e035", "4355f835", "b0390936", "4ea81736", "849b2736",
      "273c3936", "65b74c36", "243f6236", "9c0a7a36", "572b8a36", "62b39836",
      "b0c2a836", "5e82ba36", "ea1fce36", "95cde336", "edc2fb36", "aa1e0b37",
      "4dc01937", "e1eb2937", "cfca3b37", "e78a4f37", "c05e6537", "4a7e7d37",
      "a8138c37", "0dcf9a37", "1f17ab37", "8715bd37", "65f8d037", "b2f2e637",
      "af3cff37", "550a0d38", "b1df1b38", "67442c38", "82623e38", "68685238",
      "65896838", "187f8038", "b5028e38", "2bf29c38", "c473ad38", "c9b1bf38",
      "efdad338", "e822ea38", "62610139", "ccfc0e39", "91061e39", "3ca52e39",
      "5e034139", "07505539", "39bf6b39", "38458239", "98f88f39", "db1c9f39",
      "c9d8af39", "3c57c239", "9dc7d639", "495eed39", "8f2a033a", "0cf6103a",
      "f034203a", "4e0e313a", "50ad433a", "a841583a", "12006f3a", "7111843a",
      "34f5913a", "f14ea13a", "f645b23a", "bc05c53a", "50bed93a", "c7a4f03a",
      "e7f9043b", "1ef6123b", "df6a223b", "c17f333b", "8960463b", "953d5b3b",
      "5e4c723b", "f8e3853b", "ccf8933b", "bf88a33b", "b4bbb43b", "b5bdc73b",
      "7cbfdc3b", "dbf6f33b", "a4cf063c", "43fd143c", "9aa8243c", "d3f9353c",
      "4c1d493c", "0b445e3c", "44a4753c", "efbc873c", "8103963c", "70caa53c",
      "243ab73c", "4c7fca3c", "47cbdf3c", "a954f73c", "daab083d", "920b173d",
      "43ee263d", "a87c383d", "bae34b3d", "3555613d", "0008793d", "6b9c893d",
      "7115983d", "1814a83d", "61c1b93d", "a74acd3d", "dae1e23d", "5cbefa3d",
      "af8e0a3e", "2421193e", "f43b293e", "6e083b3e", "04b44e3e", "3a71643e",
      "de777c3e", "97828b3e", "b52e9a3e", "f465aa3e", "a851bc3e", "df1fd03e",
      "7703e63e", "3634fe3e", "43780c3f", "253e1b3f", "fe912b3f", "279d3d3f",
      "598e513f", "7b98673f", "a7f37f3f", "476f8d3f", "4a4f9c3f", "cdbfac3f",
      "b3eabe3f", "0fffd23f", "be2fe93f", "14db0040", "51680e40", "87621d40",
      "e5ef2d40", "dc3a4040", "83725440", "5dca6a40", "debd8140", "10638f40",
      "8f779e40", "0d22af40", "5c8dc140", "7ee8d540", "aa67ec40", "53a20241",
      "795f1041", "848e1f41", "68563041", "12e24241", "08615741", "dc076e41",
      "45888341", "9a5d9141", "67a7a041", "cb8cb141", "1839c441", "25dcd841",
      "d8aaef41", "ca6f0442", "7b5d1242", "2cc22142", "51c53242", "7e924542",
      "e1595a42", "b8507142", "eb588542", "205f9342", "ebdea242", "0100b442",
      "45eec642", "39dadb42", "78f9f242", "a1430643", "88621443", "99fd2343",
      "d43c3543", "6b4c4843", "365d5d43", "28a57443", "f52f8743", "b6679543",
      "3f1ea543", "da7bb643", "fcacc943", "dbe2de43", "c853f643", "e81d0844",
      "af6e1644", "e0402644", "0bbd3744", "f70f4b44", "2c6b6044", "58057844",
      "800d8944", "79779744", "8665a744", "7800b944", "6875cc44", "2ef6e144",
      "e9b9f944", "bdfe0945", "12821845", "2b8c2845", "18463a45", "4cdd4d45",
      "f7836345", "8f717b45", "b4f18a45", "9d8e9945", "f5b4a945", "248ebb45",
      "d547cf45", "a314e545", "612cfd45", "63e60b46", "079d1a46", "d2df2a46",
      "6cd83c46", "e0b45046", "0ea86646", "35ea7e46", "bedc8c46", "51ad9b46",
      "bd0cac46", "f924be46", "6724d246", "403ee846", "93550047", "d1d40d47",
      "78bf1c47", "b93b2d47", "d4733f47", "79965347", "3cd76947", "90378147",
      "99ce8e47", "7fd39d47", "ce6cae47", "fec4c047", "160bd547", "0b73eb47",
      "201b0248", "13ca0f48", "71e91e48", "f79f2f48", "76184248", "43825648",
      "ad116d48", "40008348", "4bc79048", "4901a048", "3fd5b048", "476ec348",
      "07fcd748", "2bb3ee48", "f0e60349", "41c61149", "131b2149", "a60c3249",
      "71c64449", "5d785949", "84577049", "3dcf8449", "f8c69249", "c636a249",
      "3a46b349", "f520c649", "5bf7da49", "c6fef149", "1ab9054a", "70c9134a",
      "7254234a", "e981344a", "dd7d474a", "f8785c4a", "eca8734a", "98a4864a",
      "b6cd944a", "1574a44a", "c4bfb54a", "31ddc84a", "38fddd4a", "0556f54a",
      "b791074b", "b8d3154b", "a295254b", "d2ff364b", "de3e4a4b", "17845f4b",
      "0206774b", "6a80884b", "90db964b", "40b9a64b", "1442b84b", "08a3cb4b",
      "b90de14b", "03b9f84b", "cf70094c", "3be5174c", "dade274c", "9386394c",
      "a5094d4c", "089a624c", "0b6f7a4c", "d6628a4c", "b9f0984c", "7906a94c",
      "45cdba4c", "b672ce4c", "1b29e44c", "0d28fc4c", "8a560b4d", "0efe194d",
      "23302a4d", "3f163c4d", "47de4f4d", "e4ba654d", "1be47d4d", "ed4b8c4d",
      "3f0d9b4d", "d85bab4d", "7b61bd4d", "564cd14d", "734fe74d", "3da3ff4d",
      "fc420d4e", "4d1e1c4e", "9c892c4e", "01af3e4e", "ebbc524e", "cfe6684e",
      "b2b2804e", "bd3b8e4e", "3d319d4e", "77b9ad4e", "ccfebf4e", "0c30d44e",
      "f280ea4e", "5995014f", "40360f4f", "0a461e4f", "60eb2e4f", "e550414f",
      "bba5554f", "ec1d6c4f", "8d79824f", "6f32904f", "c75c9f4f", "6f1fb04f",
      "54a5c24f", "ef1dd74f", "c0bded4f", "4e5f0350", "60301150", "6a752050",
      "95553150", "1cfc4350", "c3985850", "6a606f50", "a7468450", "0d309250",
      "fe8fa150", "df8db250", "4055c550", "3916da50", "f705f150", "9a2f0551",
      "80311351", "81ac2251", "4ec83351", "c2b04651", "47965b51", "62ae7251",
      "2b1a8651", "b2349451", "f9caa351", "ed04b551", "ab0ec851", "fd18dd51",
      "c359f451", "53060752", "b4391552", "6deb2452", "b4433652", "f56e4952",
      "559e5e52", "18087652", "17f48752", "7f409652", "dd0da652", "ab84b752",
      "afd1ca52", "5c26e052", "56b9f752", "84e30853", "15491753", "46322753",
      "d8c73853", "df364c53", "17b16153", "9a6d7953", "94d48953", "7f539853",
      "bc58a853", "410dba53", "7a9ecd53", "873ee353", "d624fb53", "4dc70a54",
      "c25f1954", "31812954", "e7543b54", "98084f54", "c4ce6454", "19df7c54",
      "b1bb8b54", "db6d9a54", "acabaa54", "cd9ebc54", "3175d054", "ab61e654",
      "6e9cfe54", "c8b10c55", "cf7d1b55", "3bd82b55", "faea3d55", "41e45155",
      "62f76755", "6a2e8055", "88a98d55", "a68f9c55", "e006ad55", "6439bf55",
      "e855d355", "ed8fe955", "22100156", "05a30e56", "58a31d56", "92372e56",
      "288a4056", "13ca5456", "322b6b56", "6bf38156", "369e8f56", "e6b89e56",
      "366aaf56", "f9dcc156", "6640d656", "c4c8ec56", "e5d70257", "a39a1057",
      "ddcf1f57", "859e3057", "b6314357", "f7b85757", "ef686e57", "debd8357",
      "cc989157", "bee8a057", "efd4b157", "c288c457", "2134d957", "f70bf057",
      "69a50458", "b2981258", "8b032258", "820d3358", "32e24558", "eab15a58",
      "ddb17158", "8e8e8558", "599a9358", "5020a358", "3548b458", "fa3dc758",
      "3d32dc58", "ab5af358", "49790659", "c59d1459", "033f2459", "0e853559",
      "289c4859", "41b55d59", "61067559", "9e658759", "faa29559", "b55fa559",
      "17c4b659", "c7fcc959", "ee3adf59", "06b5f659", "9c53085a", "fea9165a",
      "6082265a", "5605385a", "c65f4b5a", "43c3605a", "a166785a", "3b43895a",
      "ccb2975a", "fda6a75a", "c748b95a", "3cc5cc5a", "534ee25a", "441bfa5a",
      "78340a5b", "6abd185b", "a9cd285b", "6d8e3a5b", "1a2d4e5b", "15dc635b",
      "d9d27b5b", "61278b5b", "dec9995b", "59f6a95b", "4ed6bb5b", "8197cf5b",
      "8f6ce55b", "678dfd5b", "f91b0c5c", "26d81a5c", "12212b5c", "73203d5c",
      "5f04515c", "ccff665c", "164b7f5c", "30128d5c", "56e89b5c", "e04dac5c",
      "d56cbe5c", "b973d25c", "ce95e85c", "e185005d", "250a0e5d", "58fa1c5d",
      "b77c2d5d", "88bb3f5d", "a2e5535d", "962e6a5d", "c267815d", "ce038f5d",
      "390e9e5d", "a1adae5d", "880cc15d", "0f5ad55d", "2dcaeb5d", "3a4b025e",
      "2dff0f5e", "07241f5e", "a0e02f5e", "da5f425e", "09d1565e", "9a686d5e",
      "3530835e", "3dfc905e", "b83ba05e", "bf15b15e", "7bb5c35e", "9b4ad85e",
      "db09ef5e", "cb16045f", "15fb115f", "6155215f", "fe4c325f", "740d455f",
      "c1c6595f", "09ae705f", "f5fe845f", "abfb925f", "f170a25f", "6286b35f",
      "c167c65f", "8345db5f", "0a55f25f", "b4e80560", "f4fd1360", "6c8e2360",
      "edc13460", "77c44760", "e5c65c60", "f2fe7360", "0fd48660", "1a029560",
      "eaada460", "a0ffb560", "a023c960", "f24ade60", "d2abf560", "0ac10761",
      "f7071661", "51cf2561", "803f3761", "22854a61", "a9d15f61", "9a5b7761",
      "a5af8861", "a30f9761", "c1f2a661", "8e81b861", "12e9cb61", "015be161",
      "580ef961", "e59f0962", "34191862", "2b182862", "dec53962", "724f4d62",
      "1ce76262", "14c47a62", "c3918a62", "83249962", "9b3fa962", "5a0cbb62",
      "50b8ce62", "e275e462", "d27cfc62", "51850b63", "ac311a63", "1a692a63",
      "14553c63", "a0235063", "71076663", "93387e63", "867a8c63"
  };
  neb::floatxx_t y = zero - five * ten;
  neb::floatxx_t y_delta = one / ten;
  for (size_t i = 0; i < sizeof(mem_bytes_arr) / sizeof(mem_bytes_arr[0]);
       i++) {
    EXPECT_EQ(mem_bytes(neb::math::exp(y)), mem_bytes_arr[i]);
    y += y_delta;
  }
}

TEST(test_common_math, arctan) {
  EXPECT_EQ(neb::math::arctan(zero), std::atan(0));

  int32_t xx = 1;
  neb::floatxx_t f_xx =
      softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  neb::floatxx_t actual_x = neb::math::arctan(f_xx);
  float expect_x = std::atan(xx);
  auto ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x));
  EXPECT_EQ(mem_bytes(actual_x), "3c10493f");

  xx = -1;
  f_xx = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  actual_x = neb::math::arctan(f_xx);
  expect_x = std::atan(xx);
  ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x));
  EXPECT_EQ(mem_bytes(actual_x), "3c1049bf");

  neb::floatxx_t f_delta = one / ten;
  neb::floatxx_t f_x = zero - ten * ten;

  float delta(0.1);
  for (float x = -100.0; x <= 100.0; x += delta) {
    auto actual_x = neb::math::arctan(f_x);
    // std::cout << '\"' << mem_bytes(actual_x) << '\"' << ',';
    auto expect_x = std::atan(x);
    ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
    if (std::fabs(x) < 0.1) {
      EXPECT_TRUE(ret < precesion(expect_x, 1e2 * PRECESION));
    } else {
      EXPECT_TRUE(ret < precesion(expect_x, 1e1 * PRECESION));
    }
    f_x += f_delta;
  }
  // std::cout << std::endl;

  std::string mem_bytes_arr[] = {
      "1ac8c7bf", "c6c7c7bf", "72c7c7bf", "1ec7c7bf", "c9c6c7bf", "74c6c7bf",
      "20c6c7bf", "cbc5c7bf", "75c5c7bf", "20c5c7bf", "cbc4c7bf", "75c4c7bf",
      "1fc4c7bf", "c9c3c7bf", "73c3c7bf", "1cc3c7bf", "c6c2c7bf", "6fc2c7bf",
      "18c2c7bf", "c1c1c7bf", "6ac1c7bf", "13c1c7bf", "bbc0c7bf", "63c0c7bf",
      "0bc0c7bf", "b3bfc7bf", "5bbfc7bf", "02bfc7bf", "a9bec7bf", "51bec7bf",
      "f8bdc7bf", "9ebdc7bf", "45bdc7bf", "ebbcc7bf", "91bcc7bf", "37bcc7bf",
      "ddbbc7bf", "83bbc7bf", "28bbc7bf", "cebac7bf", "73bac7bf", "18bac7bf",
      "bcb9c7bf", "61b9c7bf", "05b9c7bf", "a9b8c7bf", "4db8c7bf", "f1b7c7bf",
      "94b7c7bf", "38b7c7bf", "dbb6c7bf", "7eb6c7bf", "21b6c7bf", "c3b5c7bf",
      "66b5c7bf", "08b5c7bf", "aab4c7bf", "4bb4c7bf", "edb3c7bf", "8eb3c7bf",
      "30b3c7bf", "d1b2c7bf", "71b2c7bf", "12b2c7bf", "b2b1c7bf", "52b1c7bf",
      "f2b0c7bf", "92b0c7bf", "32b0c7bf", "d1afc7bf", "70afc7bf", "0fafc7bf",
      "aeaec7bf", "4caec7bf", "eaadc7bf", "88adc7bf", "26adc7bf", "c4acc7bf",
      "61acc7bf", "ffabc7bf", "9cabc7bf", "38abc7bf", "d5aac7bf", "71aac7bf",
      "0daac7bf", "a9a9c7bf", "45a9c7bf", "e1a8c7bf", "7ca8c7bf", "17a8c7bf",
      "b2a7c7bf", "4ca7c7bf", "e7a6c7bf", "81a6c7bf", "1ba6c7bf", "b4a5c7bf",
      "4ea5c7bf", "e7a4c7bf", "80a4c7bf", "19a4c7bf", "b1a3c7bf", "4aa3c7bf",
      "e2a2c7bf", "7aa2c7bf", "11a2c7bf", "a9a1c7bf", "40a1c7bf", "d7a0c7bf",
      "6da0c7bf", "04a0c7bf", "9a9fc7bf", "309fc7bf", "c69ec7bf", "5b9ec7bf",
      "f19dc7bf", "869dc7bf", "1a9dc7bf", "af9cc7bf", "439cc7bf", "d79bc7bf",
      "6b9bc7bf", "ff9ac7bf", "929ac7bf", "259ac7bf", "b899c7bf", "4a99c7bf",
      "dd98c7bf", "6f98c7bf", "0198c7bf", "9297c7bf", "2397c7bf", "b496c7bf",
      "4596c7bf", "d695c7bf", "6695c7bf", "f694c7bf", "8694c7bf", "1594c7bf",
      "a593c7bf", "3493c7bf", "c292c7bf", "5192c7bf", "df91c7bf", "6d91c7bf",
      "fa90c7bf", "8890c7bf", "1590c7bf", "a28fc7bf", "2e8fc7bf", "bb8ec7bf",
      "478ec7bf", "d38dc7bf", "5e8dc7bf", "e98cc7bf", "748cc7bf", "ff8bc7bf",
      "898bc7bf", "138bc7bf", "9d8ac7bf", "278ac7bf", "b089c7bf", "3989c7bf",
      "c288c7bf", "4a88c7bf", "d287c7bf", "5a87c7bf", "e186c7bf", "6986c7bf",
      "f085c7bf", "7685c7bf", "fd84c7bf", "8384c7bf", "0984c7bf", "8e83c7bf",
      "1383c7bf", "9882c7bf", "1d82c7bf", "a181c7bf", "2581c7bf", "a980c7bf",
      "2c80c7bf", "af7fc7bf", "327fc7bf", "b57ec7bf", "377ec7bf", "b97dc7bf",
      "3a7dc7bf", "bb7cc7bf", "3c7cc7bf", "bd7bc7bf", "3d7bc7bf", "bd7ac7bf",
      "3d7ac7bf", "bc79c7bf", "3b79c7bf", "ba78c7bf", "3878c7bf", "b777c7bf",
      "3477c7bf", "b276c7bf", "2f76c7bf", "ac75c7bf", "2875c7bf", "a474c7bf",
      "2074c7bf", "9b73c7bf", "1673c7bf", "9172c7bf", "0c72c7bf", "8671c7bf",
      "0071c7bf", "7970c7bf", "f26fc7bf", "6b6fc7bf", "e36ec7bf", "5b6ec7bf",
      "d36dc7bf", "4a6dc7bf", "c16cc7bf", "386cc7bf", "ae6bc7bf", "246bc7bf",
      "9a6ac7bf", "0f6ac7bf", "8469c7bf", "f868c7bf", "6d68c7bf", "e067c7bf",
      "5467c7bf", "c766c7bf", "3a66c7bf", "ac65c7bf", "1e65c7bf", "8f64c7bf",
      "0164c7bf", "7263c7bf", "e262c7bf", "5262c7bf", "c261c7bf", "3161c7bf",
      "a060c7bf", "0f60c7bf", "7d5fc7bf", "eb5ec7bf", "585ec7bf", "c55dc7bf",
      "325dc7bf", "9e5cc7bf", "0a5cc7bf", "755bc7bf", "e05ac7bf", "4b5ac7bf",
      "b559c7bf", "1f59c7bf", "8958c7bf", "f257c7bf", "5a57c7bf", "c356c7bf",
      "2b56c7bf", "9255c7bf", "f954c7bf", "6054c7bf", "c653c7bf", "2c53c7bf",
      "9152c7bf", "f651c7bf", "5a51c7bf", "bf50c7bf", "2250c7bf", "854fc7bf",
      "e84ec7bf", "4b4ec7bf", "ad4dc7bf", "0e4dc7bf", "6f4cc7bf", "d04bc7bf",
      "304bc7bf", "904ac7bf", "ef49c7bf", "4e49c7bf", "ac48c7bf", "0a48c7bf",
      "6847c7bf", "c546c7bf", "2146c7bf", "7e45c7bf", "d944c7bf", "3444c7bf",
      "8f43c7bf", "ea42c7bf", "4342c7bf", "9d41c7bf", "f640c7bf", "4e40c7bf",
      "a63fc7bf", "fd3ec7bf", "543ec7bf", "ab3dc7bf", "013dc7bf", "563cc7bf",
      "ac3bc7bf", "003bc7bf", "543ac7bf", "a839c7bf", "fb38c7bf", "4d38c7bf",
      "a037c7bf", "f136c7bf", "4236c7bf", "9335c7bf", "e334c7bf", "3234c7bf",
      "8133c7bf", "d032c7bf", "1e32c7bf", "6b31c7bf", "b830c7bf", "0530c7bf",
      "512fc7bf", "9c2ec7bf", "e72dc7bf", "312dc7bf", "7b2cc7bf", "c42bc7bf",
      "0d2bc7bf", "552ac7bf", "9d29c7bf", "e428c7bf", "2a28c7bf", "7027c7bf",
      "b626c7bf", "fb25c7bf", "3f25c7bf", "8324c7bf", "c623c7bf", "0823c7bf",
      "4b22c7bf", "8c21c7bf", "cd20c7bf", "0d20c7bf", "4d1fc7bf", "8c1ec7bf",
      "cb1dc7bf", "091dc7bf", "461cc7bf", "831bc7bf", "bf1ac7bf", "fa19c7bf",
      "3519c7bf", "7018c7bf", "aa17c7bf", "e316c7bf", "1b16c7bf", "5315c7bf",
      "8a14c7bf", "c113c7bf", "f712c7bf", "2d12c7bf", "6111c7bf", "9610c7bf",
      "c90fc7bf", "fc0ec7bf", "2e0ec7bf", "600dc7bf", "910cc7bf", "c10bc7bf",
      "f10ac7bf", "200ac7bf", "4e09c7bf", "7c08c7bf", "a907c7bf", "d506c7bf",
      "0106c7bf", "2c05c7bf", "5604c7bf", "8003c7bf", "a802c7bf", "d101c7bf",
      "f800c7bf", "1f00c7bf", "45ffc6bf", "6bfec6bf", "8ffdc6bf", "b3fcc6bf",
      "d7fbc6bf", "f9fac6bf", "1bfac6bf", "3cf9c6bf", "5df8c6bf", "7cf7c6bf",
      "9bf6c6bf", "b9f5c6bf", "d7f4c6bf", "f4f3c6bf", "10f3c6bf", "2bf2c6bf",
      "45f1c6bf", "5ff0c6bf", "78efc6bf", "90eec6bf", "a7edc6bf", "beecc6bf",
      "d4ebc6bf", "e9eac6bf", "fde9c6bf", "10e9c6bf", "23e8c6bf", "35e7c6bf",
      "46e6c6bf", "56e5c6bf", "66e4c6bf", "74e3c6bf", "82e2c6bf", "8fe1c6bf",
      "9be0c6bf", "a7dfc6bf", "b1dec6bf", "bbddc6bf", "c3dcc6bf", "cbdbc6bf",
      "d2dac6bf", "d9d9c6bf", "ded8c6bf", "e2d7c6bf", "e6d6c6bf", "e9d5c6bf",
      "ebd4c6bf", "ecd3c6bf", "ecd2c6bf", "ebd1c6bf", "e9d0c6bf", "e6cfc6bf",
      "e3cec6bf", "decdc6bf", "d9ccc6bf", "d3cbc6bf", "cccac6bf", "c3c9c6bf",
      "bac8c6bf", "b0c7c6bf", "a5c6c6bf", "99c5c6bf", "8cc4c6bf", "7ec3c6bf",
      "70c2c6bf", "60c1c6bf", "4fc0c6bf", "3dbfc6bf", "2abec6bf", "17bdc6bf",
      "02bcc6bf", "ecbac6bf", "d5b9c6bf", "bdb8c6bf", "a4b7c6bf", "8bb6c6bf",
      "70b5c6bf", "54b4c6bf", "37b3c6bf", "19b2c6bf", "f9b0c6bf", "d9afc6bf",
      "b8aec6bf", "96adc6bf", "72acc6bf", "4eabc6bf", "28aac6bf", "01a9c6bf",
      "daa7c6bf", "b1a6c6bf", "87a5c6bf", "5ba4c6bf", "2fa3c6bf", "02a2c6bf",
      "d3a0c6bf", "a39fc6bf", "729ec6bf", "409dc6bf", "0d9cc6bf", "d99ac6bf",
      "a399c6bf", "6c98c6bf", "3497c6bf", "fb95c6bf", "c094c6bf", "8593c6bf",
      "4892c6bf", "0a91c6bf", "ca8fc6bf", "8a8ec6bf", "488dc6bf", "058cc6bf",
      "c08ac6bf", "7b89c6bf", "3488c6bf", "eb86c6bf", "a285c6bf", "5784c6bf",
      "0b83c6bf", "bd81c6bf", "6e80c6bf", "1e7fc6bf", "cd7dc6bf", "7a7cc6bf",
      "257bc6bf", "d079c6bf", "7978c6bf", "2077c6bf", "c775c6bf", "6b74c6bf",
      "0f73c6bf", "b171c6bf", "5170c6bf", "f06ec6bf", "8e6dc6bf", "2a6cc6bf",
      "c46ac6bf", "5e69c6bf", "f567c6bf", "8c66c6bf", "2065c6bf", "b363c6bf",
      "4562c6bf", "d560c6bf", "645fc6bf", "f15dc6bf", "7c5cc6bf", "065bc6bf",
      "8e59c6bf", "1558c6bf", "9a56c6bf", "1e55c6bf", "9f53c6bf", "2052c6bf",
      "9e50c6bf", "1b4fc6bf", "964dc6bf", "104cc6bf", "884ac6bf", "fe48c6bf",
      "7247c6bf", "e545c6bf", "5644c6bf", "c542c6bf", "3341c6bf", "9e3fc6bf",
      "083ec6bf", "703cc6bf", "d73ac6bf", "3b39c6bf", "9e37c6bf", "ff35c6bf",
      "5e34c6bf", "bb32c6bf", "1631c6bf", "702fc6bf", "c72dc6bf", "1d2cc6bf",
      "702ac6bf", "c228c6bf", "1227c6bf", "5f25c6bf", "ab23c6bf", "f521c6bf",
      "3d20c6bf", "821ec6bf", "c61cc6bf", "081bc6bf", "4719c6bf", "8517c6bf",
      "c015c6bf", "f913c6bf", "3012c6bf", "6610c6bf", "980ec6bf", "c90cc6bf",
      "f80ac6bf", "2409c6bf", "4e07c6bf", "7605c6bf", "9b03c6bf", "bf01c6bf",
      "e0ffc5bf", "fffdc5bf", "1bfcc5bf", "35fac5bf", "4df8c5bf", "62f6c5bf",
      "75f4c5bf", "86f2c5bf", "94f0c5bf", "a0eec5bf", "a9ecc5bf", "b0eac5bf",
      "b5e8c5bf", "b7e6c5bf", "b6e4c5bf", "b3e2c5bf", "ade0c5bf", "a5dec5bf",
      "9adcc5bf", "8cdac5bf", "7cd8c5bf", "69d6c5bf", "53d4c5bf", "3bd2c5bf",
      "20d0c5bf", "02cec5bf", "e2cbc5bf", "bfc9c5bf", "99c7c5bf", "70c5c5bf",
      "44c3c5bf", "15c1c5bf", "e4bec5bf", "afbcc5bf", "78bac5bf", "3eb8c5bf",
      "00b6c5bf", "c0b3c5bf", "7db1c5bf", "36afc5bf", "edacc5bf", "a0aac5bf",
      "50a8c5bf", "fda5c5bf", "a7a3c5bf", "4ea1c5bf", "f29ec5bf", "929cc5bf",
      "2f9ac5bf", "c897c5bf", "5f95c5bf", "f292c5bf", "8190c5bf", "0d8ec5bf",
      "968bc5bf", "1b89c5bf", "9d86c5bf", "1b84c5bf", "9581c5bf", "0c7fc5bf",
      "807cc5bf", "ef79c5bf", "5b77c5bf", "c474c5bf", "2872c5bf", "896fc5bf",
      "e66cc5bf", "3f6ac5bf", "9467c5bf", "e664c5bf", "3362c5bf", "7c5fc5bf",
      "c25cc5bf", "035ac5bf", "4057c5bf", "7954c5bf", "ae51c5bf", "df4ec5bf",
      "0c4cc5bf", "3449c5bf", "5846c5bf", "7743c5bf", "9340c5bf", "a93dc5bf",
      "bc3ac5bf", "ca37c5bf", "d334c5bf", "d831c5bf", "d82ec5bf", "d32bc5bf",
      "ca28c5bf", "bc25c5bf", "a922c5bf", "911fc5bf", "751cc5bf", "5319c5bf",
      "2d16c5bf", "5613c5bf", "2610c5bf", "f10cc5bf", "b709c5bf", "7806c5bf",
      "3303c5bf", "e9ffc4bf", "9afcc4bf", "45f9c4bf", "ebf5c4bf", "8cf2c4bf",
      "27efc4bf", "bcebc4bf", "4be8c4bf", "d5e4c4bf", "59e1c4bf", "d8ddc4bf",
      "50dac4bf", "c2d6c4bf", "2ed3c4bf", "95cfc4bf", "f5cbc4bf", "4fc8c4bf",
      "a2c4c4bf", "f0c0c4bf", "37bdc4bf", "77b9c4bf", "b1b5c4bf", "e5b1c4bf",
      "11aec4bf", "37aac4bf", "56a6c4bf", "6fa2c4bf", "809ec4bf", "8b9ac4bf",
      "8e96c4bf", "8a92c4bf", "7f8ec4bf", "6d8ac4bf", "5386c4bf", "3282c4bf",
      "097ec4bf", "d979c4bf", "a175c4bf", "6171c4bf", "196dc4bf", "c968c4bf",
      "7264c4bf", "1260c4bf", "aa5bc4bf", "3957c4bf", "c152c4bf", "3f4ec4bf",
      "b549c4bf", "2345c4bf", "8740c4bf", "e33bc4bf", "3637c4bf", "8032c4bf",
      "c02dc4bf", "f728c4bf", "2524c4bf", "491fc4bf", "641ac4bf", "7515c4bf",
      "7c10c4bf", "790bc4bf", "6c06c4bf", "5401c4bf", "33fcc3bf", "07f7c3bf",
      "d0f1c3bf", "8fecc3bf", "43e7c3bf", "ece1c3bf", "89dcc3bf", "1cd7c3bf",
      "a3d1c3bf", "1fccc3bf", "8fc6c3bf", "f3c0c3bf", "4cbbc3bf", "98b5c3bf",
      "d8afc3bf", "0baac3bf", "32a4c3bf", "4d9ec3bf", "5a98c3bf", "5a92c3bf",
      "4e8cc3bf", "3386c3bf", "0c80c3bf", "d679c3bf", "9373c3bf", "416dc3bf",
      "e166c3bf", "7360c3bf", "f659c3bf", "6a53c3bf", "cf4cc3bf", "2546c3bf",
      "6c3fc3bf", "a238c3bf", "c931c3bf", "e02ac3bf", "e623c3bf", "dc1cc3bf",
      "c115c3bf", "940ec3bf", "5707c3bf", "0800c3bf", "a7f8c2bf", "35f1c2bf",
      "b0e9c2bf", "18e2c2bf", "6ddac2bf", "b0d2c2bf", "dfcac2bf", "fac2c2bf",
      "02bbc2bf", "f5b2c2bf", "d4aac2bf", "9ea2c2bf", "529ac2bf", "f191c2bf",
      "7b89c2bf", "ee80c2bf", "4a78c2bf", "906fc2bf", "bf66c2bf", "d65dc2bf",
      "d554c2bf", "bc4bc2bf", "8a42c2bf", "3f39c2bf", "da2fc2bf", "5b26c2bf",
      "c31cc2bf", "0f13c2bf", "4009c2bf", "56ffc1bf", "4ff5c1bf", "2cebc1bf",
      "ece0c1bf", "8ed6c1bf", "12ccc1bf", "78c1c1bf", "beb6c1bf", "e5abc1bf",
      "eca0c1bf", "d195c1bf", "968ac1bf", "397fc1bf", "b973c1bf", "1668c1bf",
      "4f5cc1bf", "6450c1bf", "5444c1bf", "1e38c1bf", "c22bc1bf", "3e1fc1bf",
      "9312c1bf", "be05c1bf", "c1f8c0bf", "99ebc0bf", "47dec0bf", "c8d0c0bf",
      "1dc3c0bf", "44b5c0bf", "3da7c0bf", "0799c0bf", "a08ac0bf", "087cc0bf",
      "3e6dc0bf", "405ec0bf", "0e4fc0bf", "a73fc0bf", "0930c0bf", "3420c0bf",
      "2610c0bf", "deffbfbf", "5aefbfbf", "9adebfbf", "9ccdbfbf", "5fbcbfbf",
      "e2aabfbf", "2299bfbf", "2087bfbf", "d874bfbf", "4962bfbf", "734fbfbf",
      "533cbfbf", "e728bfbf", "2e15bfbf", "2701bfbf", "ceecbebf", "22d8bebf",
      "22c3bebf", "cbadbebf", "1b98bebf", "1082bebf", "a76bbebf", "df54bebf",
      "b43dbebf", "2526bebf", "2f0ebebf", "cff5bdbf", "02ddbdbf", "c6c3bdbf",
      "17aabdbf", "f38fbdbf", "5675bdbf", "3d5abdbf", "a53ebdbf", "8a22bdbf",
      "e805bdbf", "bce8bcbf", "02cbbcbf", "b6acbcbf", "d38dbcbf", "546ebcbf",
      "374ebcbf", "742dbcbf", "090cbcbf", "efe9bbbf", "21c7bbbf", "99a3bbbf",
      "537fbbbf", "475abbbf", "6f34bbbf", "c50dbbbf", "41e6babf", "ddbdbabf",
      "9294babf", "566ababf", "233fbabf", "ef12babf", "b1e5b9bf", "60b7b9bf",
      "f387b9bf", "5e57b9bf", "9725b9bf", "93f2b8bf", "46beb8bf", "a388b8bf",
      "9d51b8bf", "2619b8bf", "2fdfb7bf", "aaa3b7bf", "2f66b7bf", "5427b7bf",
      "b7e6b6bf", "44a4b6bf", "e75fb6bf", "8b19b6bf", "19d1b5bf", "7a86b5bf",
      "9439b5bf", "4ceab4bf", "8698b4bf", "2344b4bf", "03edb3bf", "0493b3bf",
      "0236b3bf", "d4d5b2bf", "5272b2bf", "4e0bb2bf", "9aa0b1bf", "0232b1bf",
      "4fbfb0bf", "4748b0bf", "a9ccafbf", "334cafbf", "9ac6aebf", "903baebf",
      "beaaadbf", "c813adbf", "4a76acbf", "d6d1abbf", "f625abbf", "2872aabf",
      "e1b5a9bf", "ddf0a8bf", "d621a8bf", "5d48a7bf", "a963a6bf", "da72a5bf",
      "fb74a4bf", "fe68a3bf", "b74da2bf", "d921a1bf", "f2e39fbf", "65929ebf",
      "102b9dbf", "7fac9bbf", "2c149abf", "895f98bf", "b58b96bf", "6a9594bf",
      "777992bf", "f33290bf", "76bd8dbf", "0e138bbf", "462e88bf", "700885bf",
      "309881bf", "d8ab7bbf", "586c73bf", "ae5a6abf", "655e60bf", "f75955bf",
      "442f49bf", "86bc3bbf", "89e12cbf", "1d821cbf", "ed860abf", "e7c5edbe",
      "a73dc3be", "eeab95be", "51124bbe", "d30bcebd", "006979ba", "2930ca3d",
      "a732493e", "21c7943e", "9f66c23e", "6dfeec3e", "952a0a3f", "702e1c3f",
      "82952c3f", "f5763b3f", "13f1483f", "3422553f", "422b603f", "4c2c6a3f",
      "3a42733f", "7b857b3f", "ac86813f", "68f8843f", "911f883f", "87058b3f",
      "ffb08d3f", "6d27903f", "ca6e923f", "808b943f", "7b82963f", "f056983f",
      "230c9a3f", "faa49b3f", "02249d3f", "c58b9e3f", "b6dd9f3f", "f81ba13f",
      "2b48a23f", "c063a33f", "0470a43f", "256ea53f", "315fa63f", "1e44a73f",
      "cc1da83f", "04eda83f", "36b2a93f", "a86eaa3f", "9d22ab3f", "a3ceab3f",
      "3a73ac3f", "d910ad3f", "eda7ad3f", "dc38ae3f", "02c4ae3f", "b549af3f",
      "43caaf3f", "f845b03f", "16bdb03f", "de2fb13f", "899eb13f", "5009b23f",
      "6470b23f", "f7d3b23f", "3534b33f", "4791b33f", "54ebb33f", "8142b43f",
      "f196b43f", "c4e8b43f", "1838b53f", "0985b53f", "b3cfb53f", "2f18b63f",
      "955eb63f", "fba2b63f", "78e5b63f", "1e26b73f", "0165b73f", "84a2b73f",
      "11deb73f", "0f18b83f", "8d50b83f", "9a87b83f", "44bdb83f", "98f1b83f",
      "a224b93f", "6f56b93f", "0987b93f", "7cb6b93f", "d2e4b93f", "1512ba3f",
      "4e3eba3f", "8669ba3f", "c693ba3f", "16bdba3f", "7fe5ba3f", "060dbb3f",
      "b433bb3f", "9059bb3f", "a07ebb3f", "eba2bb3f", "76c6bb3f", "47e9bb3f",
      "640bbc3f", "d32cbc3f", "994dbc3f", "ba6dbc3f", "3b8dbc3f", "21acbc3f",
      "70cabc3f", "2de8bc3f", "5c05bd3f", "0022bd3f", "1d3ebd3f", "b859bd3f",
      "d374bd3f", "728fbd3f", "99a9bd3f", "4ac3bd3f", "88dcbd3f", "57f5bd3f",
      "b90dbe3f", "b225be3f", "423dbe3f", "6f54be3f", "396bbe3f", "a381be3f",
      "b097be3f", "62adbe3f", "bbc2be3f", "bdd7be3f", "6aecbe3f", "c400bf3f",
      "ce14bf3f", "8828bf3f", "f53bbf3f", "164fbf3f", "ee61bf3f", "7e74bf3f",
      "c786bf3f", "cb98bf3f", "8caabf3f", "0bbcbf3f", "49cdbf3f", "48debf3f",
      "09efbf3f", "8effbf3f", "d70fc03f", "e61fc03f", "bd2fc03f", "5c3fc03f",
      "c44ec03f", "f75dc03f", "f56cc03f", "c07bc03f", "5a8ac03f", "c198c03f",
      "f9a6c03f", "01b5c03f", "dac2c03f", "86d0c03f", "05dec03f", "59ebc03f",
      "81f8c03f", "8005c13f", "5412c13f", "011fc13f", "852bc13f", "e237c13f",
      "1944c13f", "2a50c13f", "155cc13f", "dd67c13f", "8073c13f", "017fc13f",
      "5f8ac13f", "9b95c13f", "b6a0c13f", "b0abc13f", "89b6c13f", "44c1c13f",
      "dfcbc13f", "5bd6c13f", "b9e0c13f", "faeac13f", "1ef5c13f", "25ffc13f",
      "1009c23f", "df12c23f", "941cc23f", "2d26c23f", "ac2fc23f", "1139c23f",
      "5d42c23f", "8f4bc23f", "a954c23f", "aa5dc23f", "9466c23f", "666fc23f",
      "2078c23f", "c480c23f", "5189c23f", "c891c23f", "2a9ac23f", "75a2c23f",
      "acaac23f", "ceb2c23f", "dbbac23f", "d4c2c23f", "b9cac23f", "8ad2c23f",
      "48dac23f", "f3e1c23f", "8be9c23f", "10f1c23f", "83f8c23f", "e4ffc23f",
      "3407c33f", "710ec33f", "9e15c33f", "b91cc33f", "c423c33f", "be2ac33f",
      "a731c33f", "8138c33f", "4b3fc33f", "0546c33f", "af4cc33f", "4a53c33f",
      "d659c33f", "5460c33f", "c266c33f", "226dc33f", "7473c33f", "b879c33f",
      "ed7fc33f", "1686c33f", "308cc33f", "3d92c33f", "3d98c33f", "309ec33f",
      "16a4c33f", "efa9c33f", "bcafc33f", "7cb5c33f", "30bbc33f", "d8c0c33f",
      "74c6c33f", "04ccc33f", "88d1c33f", "02d7c33f", "6fdcc33f", "d2e1c33f",
      "29e7c33f", "75ecc33f", "b7f1c33f", "edf6c33f", "1afcc33f", "3b01c43f",
      "5306c43f", "600bc43f", "6310c43f", "5c15c43f", "4c1ac43f", "311fc43f",
      "0d24c43f", "e028c43f", "a92dc43f", "6832c43f", "1f37c43f", "cc3bc43f",
      "7140c43f", "0c45c43f", "9f49c43f", "294ec43f", "ab52c43f", "2457c43f",
      "945bc43f", "fc5fc43f", "5c64c43f", "b468c43f", "046dc43f", "4c71c43f",
      "8c75c43f", "c479c43f", "f57dc43f", "1e82c43f", "3f86c43f", "598ac43f",
      "6b8ec43f", "7792c43f", "7b96c43f", "779ac43f", "6d9ec43f", "5ca2c43f",
      "44a6c43f", "24aac43f", "ffadc43f", "d2b1c43f", "9fb5c43f", "65b9c43f",
      "25bdc43f", "dec0c43f", "90c4c43f", "3dc8c43f", "e3cbc43f", "83cfc43f",
      "1dd3c43f", "b1d6c43f", "3fdac43f", "c6ddc43f", "48e1c43f", "c4e4c43f",
      "3be8c43f", "abebc43f", "16efc43f", "7bf2c43f", "dbf5c43f", "35f9c43f",
      "8afcc43f", "d9ffc43f", "2303c53f", "6806c53f", "a709c53f", "e10cc53f",
      "1610c53f", "4613c53f", "1d16c53f", "4419c53f", "651cc53f", "821fc53f",
      "9a22c53f", "ad25c53f", "bb28c53f", "c42bc53f", "c92ec53f", "c931c53f",
      "c434c53f", "bb37c53f", "ad3ac53f", "9b3dc53f", "8440c53f", "6943c53f",
      "4a46c53f", "2649c53f", "fe4bc53f", "d14ec53f", "a151c53f", "6c54c53f",
      "3357c53f", "f659c53f", "b45cc53f", "6f5fc53f", "2662c53f", "d864c53f",
      "8767c53f", "326ac53f", "d96cc53f", "7c6fc53f", "1c72c53f", "b774c53f",
      "4f77c53f", "e379c53f", "737cc53f", "007fc53f", "8981c53f", "0f84c53f",
      "9186c53f", "0f89c53f", "8a8bc53f", "018ec53f", "7590c53f", "e692c53f",
      "5395c53f", "bd97c53f", "239ac53f", "869cc53f", "e69ec53f", "43a1c53f",
      "9ca3c53f", "f2a5c53f", "45a8c53f", "95aac53f", "e2acc53f", "2bafc53f",
      "72b1c53f", "b5b3c53f", "f5b5c53f", "33b8c53f", "6dbac53f", "a5bcc53f",
      "d9bec53f", "0bc1c53f", "39c3c53f", "65c5c53f", "8ec7c53f", "b4c9c53f",
      "d8cbc53f", "f8cdc53f", "16d0c53f", "31d2c53f", "49d4c53f", "5fd6c53f",
      "72d8c53f", "82dac53f", "90dcc53f", "9bdec53f", "a3e0c53f", "a9e2c53f",
      "ace4c53f", "ade6c53f", "abe8c53f", "a7eac53f", "a0ecc53f", "97eec53f",
      "8bf0c53f", "7df2c53f", "6cf4c53f", "59f6c53f", "44f8c53f", "2cfac53f",
      "12fcc53f", "f5fdc53f", "d7ffc53f", "b601c63f", "9203c63f", "6d05c63f",
      "4507c63f", "1b09c63f", "ef0ac63f", "c00cc63f", "900ec63f", "5d10c63f",
      "2812c63f", "f113c63f", "b715c63f", "7c17c63f", "3f19c63f", "ff1ac63f",
      "bd1cc63f", "7a1ec63f", "3420c63f", "ec21c63f", "a323c63f", "5725c63f",
      "0927c63f", "ba28c63f", "682ac63f", "142cc63f", "bf2dc63f", "672fc63f",
      "0e31c63f", "b332c63f", "5634c63f", "f735c63f", "9637c63f", "3339c63f",
      "cf3ac63f", "693cc63f", "013ec63f", "973fc63f", "2b41c63f", "be42c63f",
      "4e44c63f", "dd45c63f", "6b47c63f", "f648c63f", "804ac63f", "084cc63f",
      "8f4dc63f", "144fc63f", "9750c63f", "1852c63f", "9853c63f", "1655c63f",
      "9356c63f", "0e58c63f", "8759c63f", "ff5ac63f", "755cc63f", "ea5dc63f",
      "5d5fc63f", "ce60c63f", "3e62c63f", "ac63c63f", "1965c63f", "8566c63f",
      "ee67c63f", "5769c63f", "be6ac63f", "236cc63f", "876dc63f", "e96ec63f",
      "4a70c63f", "aa71c63f", "0873c63f", "6574c63f", "c075c63f", "1a77c63f",
      "7278c63f", "c979c63f", "1f7bc63f", "737cc63f", "c67dc63f", "187fc63f",
      "6880c63f", "b781c63f", "0483c63f", "5184c63f", "9c85c63f", "e586c63f",
      "2d88c63f", "7489c63f", "ba8ac63f", "ff8bc63f", "428dc63f", "848ec63f",
      "c48fc63f", "0491c63f", "4292c63f", "7f93c63f", "ba94c63f", "f595c63f",
      "2e97c63f", "6698c63f", "9d99c63f", "d39ac63f", "079cc63f", "3a9dc63f",
      "6c9ec63f", "9d9fc63f", "cda0c63f", "fca1c63f", "29a3c63f", "56a4c63f",
      "81a5c63f", "aba6c63f", "d4a7c63f", "fca8c63f", "22aac63f", "48abc63f",
      "6dacc63f", "90adc63f", "b2aec63f", "d4afc63f", "f4b0c63f", "13b2c63f",
      "31b3c63f", "4eb4c63f", "6ab5c63f", "85b6c63f", "9fb7c63f", "b8b8c63f",
      "d0b9c63f", "e7bac63f", "fcbbc63f", "11bdc63f", "25bec63f", "38bfc63f",
      "4ac0c63f", "5bc1c63f", "6ac2c63f", "79c3c63f", "87c4c63f", "94c5c63f",
      "a0c6c63f", "abc7c63f", "b5c8c63f", "bec9c63f", "c7cac63f", "cecbc63f",
      "d4ccc63f", "dacdc63f", "decec63f", "e2cfc63f", "e4d0c63f", "e6d1c63f",
      "e7d2c63f", "e7d3c63f", "e6d4c63f", "e4d5c63f", "e1d6c63f", "ded7c63f",
      "d9d8c63f", "d4d9c63f", "cedac63f", "c7dbc63f", "bfdcc63f", "b6ddc63f",
      "acdec63f", "a2dfc63f", "97e0c63f", "8ae1c63f", "7ee2c63f", "70e3c63f",
      "61e4c63f", "52e5c63f", "41e6c63f", "30e7c63f", "1fe8c63f", "0ce9c63f",
      "f8e9c63f", "e4eac63f", "cfebc63f", "b9ecc63f", "a3edc63f", "8beec63f",
      "73efc63f", "5af0c63f", "41f1c63f", "26f2c63f", "0bf3c63f", "eff3c63f",
      "d3f4c63f", "b5f5c63f", "97f6c63f", "78f7c63f", "58f8c63f", "38f9c63f",
      "17fac63f", "f5fac63f", "d2fbc63f", "affcc63f", "8bfdc63f", "66fec63f",
      "41ffc63f", "1b00c73f", "f400c73f", "cd01c73f", "a402c73f", "7b03c73f",
      "5204c73f", "2805c73f", "fd05c73f", "d106c73f", "a507c73f", "7808c73f",
      "4a09c73f", "1c0ac73f", "ed0ac73f", "bd0bc73f", "8d0cc73f", "5c0dc73f",
      "2a0ec73f", "f80ec73f", "c50fc73f", "9210c73f", "5e11c73f", "2912c73f",
      "f312c73f", "bd13c73f", "8714c73f", "4f15c73f", "1716c73f", "df16c73f",
      "a617c73f", "6c18c73f", "3219c73f", "f719c73f", "bb1ac73f", "7f1bc73f",
      "421cc73f", "051dc73f", "c71dc73f", "881ec73f", "491fc73f", "0a20c73f",
      "c920c73f", "8821c73f", "4722c73f", "0523c73f", "c223c73f", "7f24c73f",
      "3b25c73f", "f725c73f", "b226c73f", "6d27c73f", "2728c73f", "e028c73f",
      "9929c73f", "522ac73f", "0a2bc73f", "c12bc73f", "782cc73f", "2e2dc73f",
      "e42dc73f", "992ec73f", "4d2fc73f", "0130c73f", "b530c73f", "6831c73f",
      "1b32c73f", "cd32c73f", "7e33c73f", "2f34c73f", "df34c73f", "8f35c73f",
      "3f36c73f", "ee36c73f", "9c37c73f", "4a38c73f", "f838c73f", "a439c73f",
      "513ac73f", "fd3ac73f", "a83bc73f", "533cc73f", "fe3cc73f", "a83dc73f",
      "513ec73f", "fa3ec73f", "a33fc73f", "4b40c73f", "f240c73f", "9a41c73f",
      "4042c73f", "e642c73f", "8c43c73f", "3144c73f", "d644c73f", "7a45c73f",
      "1e46c73f", "c246c73f", "6547c73f", "0748c73f", "a948c73f", "4b49c73f",
      "ec49c73f", "8d4ac73f", "2d4bc73f", "cd4bc73f", "6c4cc73f", "0b4dc73f",
      "aa4dc73f", "484ec73f", "e54ec73f", "824fc73f", "1f50c73f", "bc50c73f",
      "5751c73f", "f351c73f", "8e52c73f", "2953c73f", "c353c73f", "5d54c73f",
      "f654c73f", "8f55c73f", "2856c73f", "c056c73f", "5857c73f", "ef57c73f",
      "8658c73f", "1c59c73f", "b359c73f", "485ac73f", "de5ac73f", "735bc73f",
      "075cc73f", "9b5cc73f", "2f5dc73f", "c25dc73f", "555ec73f", "e85ec73f",
      "7a5fc73f", "0c60c73f", "9d60c73f", "2e61c73f", "bf61c73f", "4f62c73f",
      "df62c73f", "6f63c73f", "fe63c73f", "8d64c73f", "1b65c73f", "a965c73f",
      "3766c73f", "c466c73f", "5167c73f", "de67c73f", "6a68c73f", "f668c73f",
      "8169c73f", "0c6ac73f", "976ac73f", "226bc73f", "ac6bc73f", "356cc73f",
      "bf6cc73f", "486dc73f", "d06dc73f", "596ec73f", "e16ec73f", "686fc73f",
      "ef6fc73f", "7670c73f", "fd70c73f", "8371c73f", "0972c73f", "8f72c73f",
      "1473c73f", "9973c73f", "1d74c73f", "a274c73f", "2675c73f", "a975c73f",
      "2c76c73f", "af76c73f", "3277c73f", "b477c73f", "3678c73f", "b878c73f",
      "3979c73f", "ba79c73f", "3b7ac73f", "bb7ac73f", "3b7bc73f", "bb7bc73f",
      "3a7cc73f", "b97cc73f", "387dc73f", "b67dc73f", "347ec73f", "b27ec73f",
      "307fc73f", "ad7fc73f", "2a80c73f", "a680c73f", "2381c73f", "9f81c73f",
      "1a82c73f", "9682c73f", "1183c73f", "8c83c73f", "0684c73f", "8184c73f",
      "fa84c73f", "7485c73f", "ed85c73f", "6686c73f", "df86c73f", "5887c73f",
      "d087c73f", "4888c73f", "bf88c73f", "3789c73f", "ae89c73f", "248ac73f",
      "9b8ac73f", "118bc73f", "878bc73f", "fd8bc73f", "728cc73f", "e78cc73f",
      "5c8dc73f", "d08dc73f", "458ec73f", "b98ec73f", "2c8fc73f", "a08fc73f",
      "1390c73f", "8690c73f", "f890c73f", "6b91c73f", "dd91c73f", "4f92c73f",
      "c092c73f", "3193c73f", "a293c73f", "1394c73f", "8494c73f", "f494c73f",
      "6495c73f", "d495c73f", "4396c73f", "b296c73f", "2197c73f", "9097c73f",
      "fe97c73f", "6d98c73f", "db98c73f", "4899c73f", "b699c73f", "239ac73f",
      "909ac73f", "fd9ac73f", "699bc73f", "d59bc73f", "419cc73f", "ad9cc73f",
      "189dc73f", "849dc73f", "ef9dc73f", "599ec73f", "c49ec73f", "2e9fc73f",
      "989fc73f", "02a0c73f", "6ba0c73f", "d5a0c73f", "3ea1c73f", "a7a1c73f",
      "0fa2c73f", "78a2c73f", "e0a2c73f", "48a3c73f", "afa3c73f", "17a4c73f",
      "7ea4c73f", "e5a4c73f", "4ca5c73f", "b2a5c73f", "19a6c73f", "7fa6c73f",
      "e5a6c73f", "4aa7c73f", "b0a7c73f", "15a8c73f", "7aa8c73f", "dfa8c73f",
      "43a9c73f", "a7a9c73f", "0caac73f", "6faac73f", "d3aac73f", "36abc73f",
      "9aabc73f", "fdabc73f", "60acc73f", "c2acc73f", "24adc73f", "87adc73f",
      "e9adc73f", "4aaec73f", "acaec73f", "0dafc73f", "6eafc73f", "cfafc73f",
      "30b0c73f", "90b0c73f", "f0b0c73f", "51b1c73f", "b0b1c73f", "10b2c73f",
      "6fb2c73f", "cfb2c73f", "2eb3c73f", "8db3c73f", "ebb3c73f", "4ab4c73f",
      "a8b4c73f", "06b5c73f", "64b5c73f", "c1b5c73f", "1fb6c73f", "7cb6c73f",
      "d9b6c73f", "36b7c73f", "93b7c73f", "efb7c73f", "4bb8c73f", "a7b8c73f",
      "03b9c73f", "5fb9c73f", "bbb9c73f", "16bac73f", "71bac73f", "ccbac73f",
      "27bbc73f", "81bbc73f", "dcbbc73f", "36bcc73f", "90bcc73f", "eabcc73f",
      "43bdc73f", "9dbdc73f", "f6bdc73f", "4fbec73f", "a8bec73f", "00bfc73f",
      "59bfc73f", "b1bfc73f", "0ac0c73f", "61c0c73f", "b9c0c73f", "11c1c73f",
      "68c1c73f", "c0c1c73f", "17c2c73f", "6ec2c73f", "c4c2c73f", "1bc3c73f",
      "71c3c73f", "c7c3c73f", "1dc4c73f", "73c4c73f", "c9c4c73f", "1ec5c73f",
      "74c5c73f", "c9c5c73f", "1ec6c73f", "73c6c73f", "c7c6c73f", "1cc7c73f",
      "70c7c73f", "c4c7c73f", "18c8c73f"
  };
  neb::floatxx_t y = zero - ten * ten;
  neb::floatxx_t y_delta = one / ten;
  for (size_t i = 0; i < sizeof(mem_bytes_arr) / sizeof(mem_bytes_arr[0]);
       i++) {
    EXPECT_EQ(mem_bytes(neb::math::arctan(y)), mem_bytes_arr[i]);
    y += y_delta;
  }
}

TEST(test_common_math, sin) {
  EXPECT_EQ(neb::math::sin(zero), std::sin(0));

  int32_t xx = 1;
  neb::floatxx_t f_xx =
      softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  neb::floatxx_t actual_x = neb::math::sin(f_xx);
  float expect_x = std::sin(xx);
  auto ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x));
  EXPECT_EQ(mem_bytes(actual_x), "766a573f");

  xx = -1;
  f_xx = softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(xx);
  actual_x = neb::math::sin(f_xx);
  expect_x = std::sin(xx);
  ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x));
  EXPECT_EQ(mem_bytes(actual_x), "766a57bf");

  float pi = std::acos(-1.0);
  neb::floatxx_t f_pi = neb::math::constants<neb::floatxx_t>::pi();

  actual_x = neb::math::sin(f_pi / two);
  expect_x = std::sin(pi / 2.0);
  ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x));
  EXPECT_EQ(mem_bytes(actual_x), "1e00803f");

  actual_x = neb::math::sin(f_pi / three);
  expect_x = std::sin(pi / 3.0);
  ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x));
  EXPECT_EQ(mem_bytes(actual_x), "85b35d3f");

  actual_x = neb::math::sin(zero - f_pi / four);
  expect_x = std::sin(-pi / 4.0);
  ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x));
  EXPECT_EQ(mem_bytes(actual_x), "e10435bf");

  actual_x = neb::math::sin(zero - f_pi / six);
  expect_x = std::sin(-pi / 6.0);
  ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x));
  EXPECT_EQ(mem_bytes(actual_x), "190000bf");

  for (int32_t t = -100; t <= 100; t++) {
    neb::floatxx_t tt =
        softfloat_cast<int32_t, typename neb::floatxx_t::value_type>(t);
    auto actual_x = neb::math::sin(f_pi * tt);
    // std::cout << '\"' << mem_bytes(actual_x) << '\"' << ',';
    auto expect_x = std::sin(pi * t);
    EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                precesion(1.0f, 1e2 * PRECESION));
  }
  // std::cout << std::endl;
  std::string mem_bytes_arr[] = {
      "00000000", "f60a2237", "00000000", "f60a2237", "00000000", "0f7cb0b7",
      "00000000", "0f7cb0b7", "00000000", "0f7cb0b7", "00000000", "0f7cb0b7",
      "00000000", "f60a2237", "00000000", "f60a2237", "00000000", "f60a2237",
      "00000000", "dc3dbdb6", "00000000", "dc3dbdb6", "00000000", "dc3dbdb6",
      "00000000", "dc3dbdb6", "00000000", "dc3dbdb6", "00000000", "dc3dbdb6",
      "00000000", "f60a2237", "00000000", "dc3dbdb6", "00000000", "dc3dbdb6",
      "00000000", "dc3dbdb6", "00000000", "dc3dbdb6", "00000000", "dc3dbdb6",
      "00000000", "dc3dbdb6", "00000000", "dc3dbdb6", "00000000", "f60a2237",
      "b87b9737", "dc3dbdb6", "00000000", "dc3dbdb6", "00000000", "dc3dbdb6",
      "00000000", "dc3dbdb6", "00000000", "dc3dbdb6", "00000000", "dc3dbdb6",
      "00000000", "dc3dbdb6", "00000000", "dc3dbdb6", "00000000", "c0080f36",
      "00000000", "dc3dbdb6", "00000000", "dc3dbdb6", "00000000", "dc3dbdb6",
      "00000000", "c0080f36", "b87b9737", "dc3dbdb6", "00000000", "dc3dbdb6",
      "00000000", "dc3dbdb6", "00000000", "1022e6b5", "00000000", "dc3dbdb6",
      "00000000", "1022e6b5", "00000000", "dc3dbdb6", "00000000", "1022e6b5",
      "00000000", "8c9d81b6", "00000000", "8c9d81b6", "00000000", "8c9d81b6",
      "00000000", "8c9d81b6", "882f3a37", "8c9d81b6", "00000000", "8c9d8136",
      "882f3ab7", "8c9d8136", "00000000", "8c9d8136", "00000000", "8c9d8136",
      "00000000", "8c9d8136", "00000000", "1022e635", "00000000", "dc3dbd36",
      "00000000", "1022e635", "00000000", "dc3dbd36", "00000000", "1022e635",
      "00000000", "dc3dbd36", "00000000", "dc3dbd36", "00000000", "dc3dbd36",
      "b87b97b7", "c0080fb6", "00000000", "dc3dbd36", "00000000", "dc3dbd36",
      "00000000", "dc3dbd36", "00000000", "c0080fb6", "00000000", "dc3dbd36",
      "00000000", "dc3dbd36", "00000000", "dc3dbd36", "00000000", "dc3dbd36",
      "00000000", "dc3dbd36", "00000000", "dc3dbd36", "00000000", "dc3dbd36",
      "00000000", "dc3dbd36", "b87b97b7", "f60a22b7", "00000000", "dc3dbd36",
      "00000000", "dc3dbd36", "00000000", "dc3dbd36", "00000000", "dc3dbd36",
      "00000000", "dc3dbd36", "00000000", "dc3dbd36", "00000000", "dc3dbd36",
      "00000000", "f60a22b7", "00000000", "dc3dbd36", "00000000", "dc3dbd36",
      "00000000", "dc3dbd36", "00000000", "dc3dbd36", "00000000", "dc3dbd36",
      "00000000", "dc3dbd36", "00000000", "f60a22b7", "00000000", "f60a22b7",
      "00000000", "f60a22b7", "00000000", "0f7cb037", "00000000", "0f7cb037",
      "00000000", "0f7cb037", "00000000", "0f7cb037", "00000000", "f60a22b7",
      "00000000", "f60a22b7", "00000000"
  };
  neb::floatxx_t y = zero - ten * ten;
  for (size_t i = 0; i < sizeof(mem_bytes_arr) / sizeof(mem_bytes_arr[0]);
       i++) {
    EXPECT_EQ(mem_bytes(neb::math::sin(f_pi * y)), mem_bytes_arr[i]);
    y += one;
  }

  y = zero - ten * ten;
  neb::floatxx_t y_delta = one / ten;

  float delta(0.1);
  for (float x = -100.0; x <= 100.0; x += delta) {
    auto actual_x = neb::math::sin(y);
    // std::cout << '\"' << mem_bytes(actual_x) << '\"' << ',';
    auto expect_x = std::sin(x);
    EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) <
                precesion(1.0f, 1e2 * PRECESION));
    y += y_delta;
  }
  // std::cout << std::endl;
  std::string another_mem_bytes_arr[] = {
      "009a013f", "d4fd163f", "27e02a3f", "780d3d3f", "53574d3f", "c5935b3f",
      "8f9e673f", "f258713f", "82aa783f", "5b7f7d3f", "dbcb7f3f", "1c8a7f3f",
      "ceba7c3f", "0765773f", "90966f3f", "bd62653f", "e9e4583f", "303c4a3f",
      "378e393f", "9105273f", "b3d1123f", "914cfa3e", "6376cc3e", "03949c3e",
      "9842563e", "ae72e23d", "bce7303c", "3f71b6bd", "7f9840be", "4f0492be",
      "1147c2be", "0d99f0be", "d6410ebf", "4dcb22bf", "66b435bf", "54cc46bf",
      "5ce855bf", "2ee162bf", "a8956dbf", "67ea75bf", "25ca7bbf", "40267fbf",
      "21f57fbf", "6d357ebf", "8eeb79bf", "7a2273bf", "deea69bf", "f65d5ebf",
      "199850bf", "98bc40bf", "12f42ebf", "076c1bbf", "a75606bf", "9ed2dfbe",
      "a8bbb0be", "33bf7fbe", "3f7d1bbe", "cdae56bd", "35cc423d", "f48f163e",
      "f5ec7a3e", "3364ae3e", "bf93dd3e", "8546053f", "816d1a3f", "4d0a2e3f",
      "d9e93f3f", "74de4f3f", "4ebf5d3f", "0a69693f", "bebd723f", "01a6793f",
      "2a0f7e3f", "71ee7f3f", "303f7f3f", "20037c3f", "7242763f", "e10b6e3f",
      "0174633f", "fc96563f", "0695473f", "8994363f", "0ec1233f", "a54a0f3f",
      "b9ccf23e", "6d95c43e", "8467943e", "297c453e", "a360c03d", "612fc2bb",
      "4387d8bd", "f36251be", "e5339abe", "0d2ccabe", "3a1ff8be", "e6cb11bf",
      "491326bf", "3fb138bf", "be7749bf", "cc3a58bf", "b9d464bf", "4e256fbf",
      "2c1277bf", "13877cbf", "6d767fbf", "ccd77fbf", "dfaa7dbf", "2af578bf",
      "b5c271bf", "512568bf", "08375cbf", "51154ebf", "64e43dbf", "bccd2bbf",
      "290018bf", "e0ac02bf", "3817d8be", "26aca8be", "d5216fbe", "cc8a0abe",
      "004812bd", "a795833d", "f779273e", "72bf853e", "f06ab63e", "db43e53e",
      "5de9083f", "1fd21d3f", "ca27313f", "43b8423f", "8c56523f", "efda5f3f",
      "f4226b3f", "3c11743f", "b48f7a3f", "ad8c7e3f", "d6fe7f3f", "0ee27e3f",
      "67397b3f", "2c0e753f", "856f6c3f", "fb74613f", "b739543f", "8ddf443f",
      "d18d333f", "bb70203f", "58b90b3f", "393aeb3e", "80a6bc3e", "57308c3e",
      "74a7343e", "0a419e3d", "1588b9bc", "078efabd", "811e62be", "7b58a2be",
      "9302d2be", "a293ffbe", "854b15bf", "5f4f29bf", "95a13bbf", "b9144cbf",
      "c27d5abf", "e1b766bf", "d5a370bf", "402878bf", "90327dbf", "4fb47fbf",
      "27a87fbf", "2a0e7dbf", "f5ec77bf", "a15170bf", "214f66bf", "54005abf",
      "c6834bbf", "98fe3abf", "179b28bf", "c48814bf", "81f3fdbe", "594cd0be",
      "9190a0be", "2f735ebe", "cb1cf3bd", "00909bbc", "e5c1a53d", "be5a383e",
      "effe8d3e", "1365be3e", "2ee4ec3e", "5a820c3f", "ac2b213f", "e738343f",
      "1179453f", "2dc0543f", "05e7613f", "00cc6c3f", "e853753f", "a6677b3f",
      "3bf87e3f", "aefc7f3f", "9a727e3f", "8a5d7a3f", "36c8733f", "c1c26a3f",
      "8d655f3f", "facc513f", "b01b423f", "d479303f", "a1141d3f", "961d083f",
      "e695e33e", "02a9b43e", "2aee833e", "08c4233e", "9423783d", "1b4321bd",
      "52430ebe", "c8cb72be", "5772aabe", "f9cad9be", "4f7b03bf", "dbc018bf",
      "b57f2cbf", "d2843ebf", "64a34ebf", "55b15cbf", "bc8a68bf", "4b1172bf",
      "af2c79bf", "4ccb7dbf", "e7df7fbf", "2b667fbf", "415f7cbf", "ddd276bf",
      "2bcf6ebf", "396864bf", "bfb957bf", "5ae348bf", "100b38bf", "005c25bf",
      "570611bf", "347af6be", "a771c8be", "896898be", "84b24dbe", "910ed1bd",
      "000014bb", "98e2c73d", "252f493e", "b934963e", "d451c63e", "4974f43e",
      "c811103f", "e979243f", "1e3d373f", "aa2b483f", "571a573f", "a5e2633f",
      "34646e3f", "6d84763f", "662d7c3f", "67517f3f", "59e87f3f", "e1f07d3f",
      "cb6f793f", "d270723f", "1b05693f", "32465d3f", "23514f3f", "e3493f3f",
      "45592d3f", "38ad193f", "1c78043f", "55e1db3e", "a19eac3e", "2945773e",
      "c4d4123e", "a6b3333d", "1cb765bd", "92351fbe", "d4b381be", "ff7fb2be",
      "be83e1be", "5f2307bf", "3d2b1cbf", "aea32fbf", "635a41bf", "402351bf",
      "16d55ebf", "e84c6abf", "686d73bf", "401f7abf", "d7517ebf", "2cf97fbf",
      "e5117fbf", "439e7bbf", "16a775bf", "983b6dbf", "f07062bf", "b26355bf",
      "883446bf", "560a35bf", "0d1122bf", "89790dbf", "3fefeebe", "9888c0be",
      "963590be", "03e33cbe", "56f1aebd", "103c6d3c", "d2f1e93d", "19f2593e",
      "e75e9e3e", "852fce3e", "48f1fb3e", "7796133f", "d9bb273f", "fd333a3f",
      "cecf4a3f", "e564593f", "d9cd653f", "07eb6f3f", "23a3773f", "10e17c3f",
      "40987f3f", "adc17f3f", "085d7d3f", "4c70783f", "2108713f", "0337673f",
      "1f175b3f", "d6c64c3f", "b56a3c3f", "9e2c2a3f", "3b3b163f", "8fc9003f",
      "e81dd43e", "e188a43e", "329e663e", "0bdd013e", "9c7dde3c", "5e0995bd",
      "d51c30be", "88f789be", "f47fbabe", "7d2be9be", "58c10abf", "0c8a1fbf",
      "b8ba32bf", "c72144bf", "d39353bf", "a1e860bf", "1bfe6bbf", "edb774bf",
      "ceff7abf", "1ec67ebf", "0c0080bf", "62ab7ebf", "57cb7abf", "d86974bf",
      "36976bbf", "ab6960bf", "a4fe52bf", "d37743bf", "fffc31bf", "e4ba1ebf",
      "06e309bf", "1254e7be", "ad92b8be", "43f987be", "dc072cbe", "8acb8cbd",
      "d5abff3c", "35f9053e", "e6a56a3e", "f17da63e", "10ffd53e", "8aae013f",
      "ff10173f", "f3f12a3f", "b11d3d3f", "71654d3f", "f89f5b3f", "bea8673f",
      "e560713f", "3cb0783f", "a0827d3f", "bccc7f3f", "a7887f3f", "fbb67c3f",
      "f35e773f", "378e6f3f", "3858653f", "49d8583f", "a52d4a3f", "d47d393f",
      "92f3263f", "3dbe123f", "1d23fa3e", "d74acc3e", "da669c3e", "c2e5553e",
      "d1b5e13d", "45f82a3c", "6e2eb7bd", "c8f540be", "d73192be", "0373c2be",
      "f7c2f0be", "92550ebf", "a1dd22bf", "1dc535bf", "49db46bf", "66f555bf",
      "2fec62bf", "7f9e6dbf", "fff075bf", "6bce7bbf", "2f287fbf", "b0f47fbf",
      "9e327ebf", "67e679bf", "081b73bf", "38e169bf", "31525ebf", "548a50bf",
      "f7ac40bf", "bbe22ebf", "cf591bbf", "6f4206bf", "e4a7dfbe", "138fb0be",
      "39637fbe", "591f1bbe", "523355bd", "7349443d", "efee163e", "eb497b3e",
      "6391ae3e", "41bfdd3e", "e15a053f", "c3801a3f", "d11b2e3f", "9ff93f3f",
      "4eec4f3f", "45cb5d3f", "f072693f", "5ac5723f", "54ab793f", "22127e3f",
      "07ef7f3f", "6c3d7f3f", "e8fe7b3f", "e43b763f", "08036e3f", "0669633f",
      "e289563f", "ff85473f", "b783363f", "99ae233f", "c3360f3f", "81a2f23e",
      "1d69c43e", "8e39943e", "f71d453e", "77a1bf3d", "f52fcebb", "3146d9bd",
      "fdc051be", "a9619abe", "2b58cabe", "2f49f8be", "9fdf11bf", "8e2526bf",
      "e0c138bf", "8b8649bf", "a54758bf", "7adf64bf", "dd2d6fbf", "741877bf",
      "018b7cbf", "fa777fbf", "f3d67fbf", "a3a77dbf", "93ef78bf", "cfba71bf",
      "321b68bf", "ca2a5cbf", "12074ebf", "4cd43dbf", "f0bb2bbf", "d8ec17bf",
      "3c9802bf", "b2ebd7be", "d47ea8be", "7fc46ebe", "ad2b0abe", "00c810bd",
      "6558843d", "58d9273e", "bfed853e", "fb97b63e", "276fe53e", "f2fd083f",
      "26e51d3f", "3a39313f", "12c8423f", "9b64523f", "d5e65f3f", "632c6b3f",
      "9518743f", "b5947a3f", "5d8f7e3f", "e3fe7f3f", "badf7e3f", "b2347b3f",
      "2c07753f", "40666c3f", "7e69613f", "282c543f", "15d0443f", "8b7c333f",
      "e45d203f", "07a50b3f", "1f0feb3e", "5979bc3e", "b2018c3e", "1948343e",
      "697f9d3d", "e78fbcbc", "784efbbd", "247d62be", "7a86a2be", "d22ed2be",
      "a8bdffbe", "375f15bf", "8e6129bf", "13b23bbf", "5c234cbf", "648a5abf",
      "61c266bf", "19ac70bf", "352e78bf", "24367dbf", "78b57fbf", "e5a67fbf",
      "7e0a7dbf", "e9e677bf", "454970bf", "8a4466bf", "9df359bf", "0f754bbf",
      "08ee3abf", "d88828bf", "047514bf", "61c9fdbe", "0a20d0be", "8362a0be",
      "81145ebe", "295cf2bd", "008898bc", "5481a63d", "51b9383e", "a62c8e3e",
      "7291be3e", "130fed3e", "70960c3f", "5b3e213f", "004a343f", "5788453f",
      "76cd543f", "42f2613f", "24d56c3f", "b95a753f", "186c7b3f", "52fa7e3f",
      "82fc7f3f", "e86f7e3f", "9d587a3f", "e4c0733f", "25b96a3f", "e4595f3f",
      "38bf513f", "020c423f", "7868303f", "aa011d3f", "4709083f", "ee6ae33e",
      "217cb43e", "d6bf833e", "4365233e", "24a4763d", "2fc322bd", "6ba20ebe",
      "082973be", "949faabe", "6cf6d9be", "e78f03bf", "1bd418bf", "71912cbf",
      "d8943ebf", "8fb14ebf", "7dbd5cbf", "c39468bf", "181972bf", "2f3279bf",
      "6fce7dbf", "a7e07fbf", "85647fbf", "3a5b7cbf", "7ecc76bf", "84c66ebf",
      "625d64bf", "d2ac57bf", "78d448bf", "60fa37bf", "ac4925bf", "8ff210bf",
      "2050f6be", "7b45c8be", "b63a98be", "7a544dbe", "924fd0bd", "0000f8ba",
      "c2a0c83d", "dc8a493e", "aa61963e", "087ec63e", "9c9df43e", "8525103f",
      "218c243f", "b44d373f", "773a483f", "2327573f", "94ed633f", "ce6c6e3f",
      "c58a763f", "84317c3f", "20537f3f", "b1e77f3f", "e6ed7d3f", "6e6a793f",
      "2169723f", "3efb683f", "443a5d3f", "42434f3f", "133a3f3f", "c5472d3f",
      "3d9a193f", "c463043f", "7ab6db3e", "ec71ac3e", "f4e8763e", "b676123e",
      "d537323d", "a93267bd", "6f931fbe", "cae181be", "89acb2be", "62aee1be",
      "8d3707bf", "103e1cbf", "f4b42fbf", "f46941bf", "ee3051bf", "c6e05ebf",
      "78566abf", "c27473bf", "4e247abf", "8f547ebf", "84f97fbf", "dd0f7fbf",
      "e2997bbf", "65a075bf", "a9326dbf", "dc6562bf", "935655bf", "7f2546bf",
      "8bf934bf", "abfe21bf", "bd650dbf", "03c5eebe", "1d5cc0be", "490790be",
      "ad833cbe", "0c2faebd", "2e71733c", "21b6ea3d", "96545a3e", "518e9e3e",
      "2f5ece3e", "071dfc3e", "7cab133f", "46cf273f", "91453a3f", "9edf4a3f",
      "a272593f", "62d9653f", "3af46f3f", "cfa9773f", "1ee57c3f", "b2997f3f",
      "71c07f3f", "29597d3f", "c169783f", "0bff703f", "482b673f", "02095b3f",
      "4bb64c3f", "fd573c3f", "e3172a3f", "a424163f", "57b1003f", "a4ead33e",
      "5b53a43e", "912f663e", "f26b013e", "20eada3c", "b0ee95bd", "918e30be",
      "612f8abe", "33b6babe", "8c5fe9be", "06da0abf", "20a11fbf", "eacf32bf",
      "e83444bf", "a6a453bf", "f3f660bf", "c6096cbf", "cdc074bf", "c5057bbf",
      "19c97ebf", "030080bf", "3ea87ebf", "23c57abf", "9f6074bf", "068b6bbf",
      "9f5a60bf", "deec52bf", "7b6343bf", "43e631bf", "f9a11ebf", "29c809bf",
      "f81ae7be", "bd56b8be", "14bb87be", "33882bbe", "27c88bbd", "c2db013d",
      "507a063e", "e7256b3e", "63bca63e", "433bd63e", "17cb013f", "032c173f",
      "de0a2b3f", "34343d3f", "a3794d3f", "57b15b3f", "1eb7673f", "396c713f",
      "49b8783f", "6b877d3f", "1bce7f3f", "99867f3f", "7cb17c3f", "f555773f",
      "d9816f3f", "a848653f", "8cc5583f", "f7174a3f", "6065393f", "8dd8263f",
      "eaa0123f", "88e4f93e", "ca08cc3e", "f7219c3e", "f257553e", "8694e03d",
      "50d8213c", "3752b8bd", "238641be", "847892be", "77b7c2be", "7b04f1be",
      "8a740ebf", "75fa22bf", "88df35bf", "f8f246bf", "1e0a56bf", "b3fd62bf",
      "9fac6dbf", "90fb75bf", "4ed57bbf", "4a2b7fbf", "f5f37fbf", "062e7ebf",
      "fbdd79bf", "d60e73bf", "5ad169bf", "c93e5ebf", "8d7350bf", "089340bf",
      "e0c52ebf", "4d3a1bbf", "942006bf", "2460dfbe", "f943b0be", "cfc77ebe",
      "33801abe", "30ae52bd", "e0ca463d", "998f173e", "d4e67b3e", "4ddeae3e",
      "0d09de3e", "017e053f", "67a11a3f", "293a2e3f", "3415403f", "a604503f",
      "3de05d3f", "1684693f", "c9d2723f", "a3b4793f", "50177e3f", "ffef7f3f",
      "123a7f3f", "55f77b3f", "2730763f", "33f36d3f", "3a55633f", "5972563f",
      "ea6a473f", "4465363f", "248d233f", "97120f3f", "6855f23e", "1e18c43e",
      "59e5933e", "bf70443e", "ea40be3d", "2a5de4bb", "40a8dabd", "a76f52be",
      "fbb69abe", "92aacabe", "e697f8be", "b00412bf", "f64726bf", "41e138bf",
      "97a249bf", "0c6058bf", "f7f364bf", "323e6fbf", "772477bf", "8c927cbf",
      "ef7a7fbf", "4cd57fbf", "57a17dbf", "aee478bf", "67ab71bf", "680768bf",
      "cc125cbf", "16eb4dbf", "95b43dbf", "cb982bbf", "99c617bf", "446f02bf",
      "1795d7be", "6624a8be", "cc096ebe", "ee6c09be", "00c40dbd", "03d9853d",
      "5799283e", "af4b863e", "58f3b63e", "b7c6e53e", "3827093f", "d40b1e3f",
      "d55c313f", "15e8423f", "a880523f", "dafe5f3f", "25406b3f", "8c27743f",
      "029f7a3f", "a7947e3f", "47ff7f3f", "11db7e3f", "f12a7b3f", "77f8743f",
      "c7526c3f", "6651613f", "980f543f", "57af443f", "f057333f", "cf35203f",
      "d8790b3f", "50b3ea3e", "1519bc3e", "e89d8b3e", "587b333e", "b9df9b3d",
      "6e17c3bc", "82eefcbd", "de4963be", "3feaa2be", "008fd2be", "9f0c00bf",
      "3d8a15bf", "5f8929bf", "47d63bbf", "8f434cbf", "3ca65abf", "95d966bf",
      "66be70bf", "663b78bf", "113e7dbf", "07b87fbf", "0da47fbf", "3d027dbf",
      "52d977bf", "753670bf", "af2c66bf", "edd659bf", "d0534bbf", "8bc83abf",
      "795f28bf", "294814bf", "9a69fdbe", "1ebbcfbe", "71f99fbe", "273c5dbe",
      "43a3f0bd", "009591bc", "153da83d", "dd94393e", "fd978e3e", "e0f9be3e",
      "4772ed3e", "3ec50c3f", "0a6a213f", "0672343f", "41ac453f", "e1ec543f",
      "e60c623f", "9bea6c3f", "f56a753f", "b0767b3f", "7cff7e3f", "f2fb7f3f",
      "94697e3f", "894c7a3f", "5eaf733f", "38a26a3f", "c13d5f3f", "299e513f",
      "5ce6413f", "873e303f", "ecd31c3f", "15d8073f", "bf02e33e", "0d0fb43e",
      "134f833e", "957e223e", "e3fd723d", "e66b26bd", "e58a0fbe", "730d74be",
      "9d0eabbe", "1461dabe", "8ec203bf", "8b0319bf", "2ebd2cbf", "74bc3ebf",
      "9ad44ebf", "9ddb5cbf", "a5ad68bf", "7c2c72bf", "dc3f79bf", "3cd67dbf",
      "7ee27fbf", "5b607fbf", "12517cbf", "6fbc76bf", "b2b06ebf", "014264bf",
      "228c57bf", "c8ae48bf", "0cd037bf", "1d1b25bf", "38c010bf", "e1e4f5be",
      "b9d4c7be", "8cc597be", "ac634cbe", "b965cebd", "006979ba", "1285ca3d",
      "d47a4a3e", "47d7963e", "ffeec63e", "fe09f53e", "4a59103f", "bcbb243f",
      "2a79373f", "4461483f", "c048573f", "7009643f", "81836e3f", "139b763f",
      "de3b7c3f", "87577f3f", "20e67f3f", "4ce67d3f", "305c793f", "ce54723f",
      "8ae1683f", "951a5d3f", "fe1d4f3f", "890f3f3f", "23192d3f", "7367193f",
      "662d043f", "9242db3e", "9bf9ab3e", "b5f1753e", "c278113e", "20352e3d",
      "10366bbd", "029220be", "fa5d82be", "9225b3be", "1d23e2be", "586e07bf",
      "40711cbf", "06e42fbf", "d69441bf", "765651bf", "9c005fbf", "41706abf",
      "098973bf", "3b327abf", "985b7ebf", "43fa7fbf", "500a7fbf", "118e7bbf",
      "5f8e75bf", "c2196dbf", "1b4762bf", "003255bf", "63fb45bf", "51ca34bf",
      "b0ca21bf", "752d0dbf", "384feebe", "c7e0bfbe", "98878fbe", "b27f3bbe",
      "c21facbd", "00ff813c", "92c4ec3d", "21555b3e", "7a0c9f3e", "18d7ce3e",
      "cb90fc3e", "71e2133f", "7901283f", "30733a3f", "f6074b3f", "5595593f",
      "02f6653f", "3b0b703f", "feb9773f", "1def7c3f", "519d7f3f", "c2bd7f3f",
      "33507d3f", "c459783f", "b1e8703f", "610f673f", "fee65a3f", "a08e4c3f",
      "eb2a3c3f", "b9e6293f", "61ef153f", "8a78003f", "3372d33e", "8ed6a33e",
      "0c30653e", "0b66003e", "99abd23c", "ecfc97bd", "2c9031be", "e2ad8abe",
      "0c31bbbe", "9fd5e9be", "36110bbf", "82d41fbf", "e8fe32bf", "6f5f44bf",
      "aec953bf", "0d1661bf", "b2226cbf", "26d474bf", "bc127bbf", "17cf7ebf",
      "d7ff7fbf", "a0a17ebf", "39b87abf", "9c4d74bf", "29716bbf", "0a3b60bf",
      "b1c752bf", "bf3843bf", "9bb631bf", "d96d1ebf", "669009bf", "82a5e6be",
      "4edcb7be", "873c87be", "52852abe", "bcba89bd", "00f9053d", "a37e073e",
      "a9246c3e", "c738a73e", "ccb2d63e", "e903023f", "1b62173f", "2f3c2b3f",
      "c6603d3f", "eba04d3f", "00d35b3f", "a2d2673f", "2882713f", "63c7783f",
      "4a907d3f", "aad07f3f", "c4827f3f", "c5a67c3f", "de44773f", "766a6f3f",
      "a22b653f", "95a2583f", "54ef493f", "6d37393f", "aea6263f", "086b123f",
      "c771f93e", "878fcb3e", "a3a49b3e", "a857543e", "fd87de3d", "1a5b113c",
      "ea5fbabd", "398742be", "88f692be", "8d31c3be", "8679f1be", "17ab0ebf",
      "162d23bf", "a80d36bf", "911c47bf", "262e56bf", "c91b63bf", "23c56dbf",
      "d00d76bf", "2be17bbf", "34307fbf", "9bf27fbf", "61267ebf", "10d079bf",
      "0cfa72bf", "26b769bf", "4f1e5ebf", "4e4d50bf", "8b6740bf", "98952ebf",
      "ed041bbf", "67e805bf", "a4e9debe", "33c8afbe", "69cb7dbe", "687c19be",
      "21974ebd", "bce44a3d", "da92183e", "78e47c3e", "a459af3e", "887fde3e",
      "42b6053f", "04d71a3f", "a46a2e3f", "be40403f", "0a2b503f", "c6005e3f",
      "769e693f", "90e7723f", "b9c2793f", "0e1f7e3f", "6df17f3f", "36357f3f",
      "9aeb7b3f", "031e763f", "ddda6d3f", "4637633f", "774e563f", "6541473f",
      "3b37363f", "965a233f", "1ddc0e3f", "81e0f13e", "229ec33e", "8267933e",
      "1270433e", "ff33bc3d", "afa702bc", "94b4dcbd", "c66f53be", "4d349bbe",
      "e623cbbe", "a90af9be", "a03a12bf", "e27926bf", "450f39bf", "4ccb49bf",
      "098358bf", "021165bf", "b2556fbf", "a53577bf", "599d7cbf", "e47e7fbf",
      "e1d27fbf", "a3987dbf", "b7d578bf", "ac9571bf", "11ec67bf", "45f15bbf",
      "f2c34dbf", "2b883dbf", "ba672bbf", "a79017bf", "693602bf", "c51dd7be",
      "66a8a7be", "c90a6dbe", "136808be", "faa709bd", "6ae4873d", "309c293e",
      "70ca863e", "e36db73e", "383ce63e", "f35e093f", "fd3f1e3f", "798c313f",
      "ca12433f", "13a6523f", "711e603f", "0d5a6b3f", "583b743f", "efab7a3f",
      "5d9b7e3f", "97ff7f3f", "1dd57e3f", "221e7b3f", "42e5743f", "f6396c3f",
      "6132613f", "bdea533f", "f184443f", "1e29333f", "9102203f", "c4420b3f",
      "863dea3e", "6d9ebb3e", "941f8b3e", "097a323e", "d8d1993d", "3d53cbbc",
      "91f9febd", "3d4964be", "f266a3be", "7707d3be", "754500bf", "8cbf15bf",
      "8bba29bf", "61033cbf", "4c6b4cbf", "53c85abf", "97f566bf", "ced470bf",
      "8b4b78bf", "16477dbf", "e5ba7fbf", "96a07fbf", "66f87cbf", "3bc977bf",
      "9a1f70bf", "2c1066bf", "5fb459bf", "9e2b4bbf", "2d9b3abf", "422d28bf",
      "5b1114bf", "7ef6fcbe", "fc41cfbe", "a27b9fbe", "783b5cbe", "e497eebd",
      "425189bc", "954caa3d", "24993a3e", "cc178f3e", "ce74bf3e", "2fe8ed3e",
      "acfd0c3f", "ff9d213f", "39a1343f", "6bd6453f", "8711553f", "872b623f",
      "9f036d3f", "217d753f", "bc827b3f", "12057f3f", "2efb7f3f", "aa627e3f",
      "ba3e7a3f", "369b733f", "95886a3f", "0c1e5f3f", "c178513f", "98bb413f",
      "9a0f303f", "d0a01c3f", "6aa1073f", "3d8ee23e", "3e96b33e", "0bd3823e",
      "4280213e", "6dfb6e3d", "ba6d2abd", "c68810be", "9c0475be", "ed86abbe",
      "f0d4dabe", "f6f803bf", "5c3619bf", "dbeb2cbf", "17e73ebf", "e4f94ebf",
      "5afb5cbf", "7dc768bf", "e94072bf", "2e4e79bf", "e1dd7dbf", "2be47fbf",
      "135c7fbf", "d6467cbf", "55ac76bf", "2c9a6ebf", "452664bf", "cf6a57bf",
      "538848bf", "dda437bf", "8deb24bf", "908c10bf", "b978f5be", "b363c7be",
      "865097be", "74744bbe", "ef83ccbd", "48a9c5b7", "a76ecc3d", "8d6b4b3e",
      "6c4c973e", "c45fc73e", "4575f53e", "a88b103f", "56ea243f", "8ba3373f",
      "0487483f", "8569573f", "e724643f", "6c996e3f", "3bab763f", "22467c3f",
      "d05b7f3f", "6ae47f3f", "9ede7d3f", "a24e793f", "8941723f", "c8c8683f",
      "92fc5c3f", "11fb4e3f", "0de83e3f", "80ed2c3f", "1d38193f", "ddfa033f",
      "13d8da3e", "ba8aab3e", "a60d753e", "8090103e", "668d2a3d", "afdb6ebd",
      "8b7821be", "a2ce82be", "8592b3be", "468be2be", "939f07bf", "169f1cbf",
      "f90d30bf", "92ba41bf", "987751bf", "cb1c5fbf", "43876abf", "a99a73bf",
      "643e7abf", "1a627ebf", "06fb7fbf", "5e057fbf", "6a837bbf", "487e75bf",
      "69046dbf", "c82c62bf", "be1255bf", "99d745bf", "5ca234bf", "289f21bf",
      "72fe0cbf", "1decedbe", "5079bfbe", "9f1c8fbe", "9da43abe", "e864aabd",
      "00f0883c", "127dee3d", "532d5c3e", "7f759f3e", "fc3bcf3e", "90f0fc3e",
      "4d0f143f", "dd2a283f", "b6983a3f", "41294b3f", "13b2593f", "ef0d663f",
      "1f1e703f", "aac7773f", "76f77c3f", "43a07f3f", "4ebb7f3f", "63487d3f",
      "af4c783f", "83d6703f", "4bf8663f", "44cb5a3f", "8c6e4c3f", "d4063c3f",
      "03bf293f", "76c4153f", "d64a003f", "3c12d33e", "f872a33e", "8f63643e",
      "0d2dff3d", "3e26cc3c", "489c99bd", "bd5c32be", "a2118bbe", "4c91bbbe",
      "6631eabe", "5c3c0bbf", "8ffc1fbf", "822333bf", "338044bf", "3fe653bf",
      "3e2e61bf", "3f366cbf", "ede274bf", "961c7bbf", "04d47ebf", "9cff7fbf",
      "719c7ebf", "18ae7abf", "fe3d74bf", "a65d6bbf", "312360bf", "7fab52bf",
      "f91843bf", "2e9331bf", "2d471ebf", "276709bf", "364ee6be", "4c81b7be",
      "ffde86be", "35c729be", "603b88bd", "00fc083d", "2f3d083e", "34df6c3e",
      "2693a73e", "5e09d73e", "e02c023f", "5b88173f", "585f2b3f", "83803d3f",
      "efbc4d3f", "0beb5b3f", "7ae6673f", "9f91713f", "59d2783f", "ac967d3f",
      "67d27f3f", "e57f7f3f", "549f7c3f", "f438773f", "3a5a6f3f", "4417653f",
      "4c8a583f", "63d3493f", "2318393f", "5e84263f", "1246123f", "4723f93e",
      "433dcb3e", "7d4f9b3e", "3fa9533e", "7526dd3d", "6c46063c", "c4bfbbbd",
      "2f3443be", "ab4a93be", "9b82c3be", "adc6f1be", "4ecf0ebf", "9b4e23bf",
      "2f2c36bf", "ba3747bf", "c54556bf", "9a2f63bf", "00d56dbf", "a61976bf",
      "cde87bbf", "93337fbf", "b3f17fbf", "3e217ebf", "e2c679bf", "e1ec72bf",
      "0ea669bf", "86095ebf", "343550bf", "0b4c40bf", "38772ebf", "24e41abf",
      "56c505bf", "e69fdebe", "917bafbe", "032e7dbe", "77dd18be", "540b4cbd",
      "ee694d3d", "0b32193e", "f57f7d3e", "cca4af3e", "59c7de3e", "26d8053f",
      "91f61a3f", "8c872e3f", "bb5a403f", "e141503f", "3e145e3f", "66ae693f",
      "d3f3723f", "37cb793f", "b7237e3f", "3cf27f3f", "2e327f3f", "cce47b3f",
      "8413763f", "cccc6d3f", "d325633f", "cd39563f", "c329473f", "e01c363f",
      "cd3d233f", "33bd0e3f", "0e9ff13e", "b859c33e", "de20933e", "bddf423e",
      "3b10bb3d", "85c80bbc", "cdd5ddbd", "8afd53be", "1a799bbe", "f165cbbe",
      "5849f9be", "f85712bf", "ef9426bf", "c42739bf", "0be149bf", "e09558bf",
      "a82065bf", "1d626fbf", "c33e77bf", "00a37cbf", "06817fbf", "93d17fbf",
      "e0937dbf", "87cd78bf", "518a71bf", "a4dd67bf", "ebdf5bbf", "e4af4dbf",
      "7d713dbf", "b94e2bbf", "ba7517bf", "e71902bf", "51e1d6be", "196aa7be",
      "d38b6cbe", "1de707be", "fe9a07bd", "d7e7883d", "e41b2a3e", "a908873e",
      "dea9b73e", "6175e63e", "d879093f", "f1581e3f", "40a3313f", "2d27433f",
      "e6b7523f", "8a2d603f", "4a666b3f", "b444743f", "3eb27a3f", "959e7e3f",
      "baff7f3f", "2cd27e3f", "2c187b3f", "5edc743f", "3e2e6c3f", "fe23613f",
      "d3d9533f", "b671443f", "c313333f", "5aeb1f3f", "e8290b3f", "0e09ea3e",
      "c467bb3e", "3fe78a3e", "4107323e", "97ea983d", "ffeecebc", "ffddffbd",
      "d6b864be", "f99ca3be", "2d3bd3be", "ec5d00bf", "5bd615bf", "87cf29bf",
      "4d163cbf", "0d7c4cbf", "acd65abf", "7d0167bf", "1cde70bf", "245278bf",
      "184b7dbf", "0fbc7fbf", "279f7fbf", "63f47cbf", "7bc277bf", "621670bf",
      "8a0466bf", "88a659bf", "981b4bbf", "67893abf", "c81928bf", "4cfc13bf",
      "f4c9fcbe", "9313cfbe", "b24b9fbe", "2dd85bbe", "4bd0edbd", "a63186bc",
      "e610ab3d", "7ef93a3e", "99468f3e", "c8a1bf3e", "e912ee3e", "b3110d3f",
      "9bb0213f", "3bb2343f", "a5e5453f", "d41e553f", "c536623f", "b10c6d3f",
      "ee83753f", "36877b3f", "2b077f3f", "e1fa7f3f", "f85f7e3f", "ab397a3f",
      "d493733f", "f47e6a3f", "45125f3f", "f46a513f", "e7ab413f", "2cfe2f3f",
      "d28d1c3f", "0b8d073f", "2d63e23e", "4a69b33e", "aaa4823e", "6521213e",
      "d17b6d3d", "46ed2bbd", "dfe710be", "cb6175be", "29b4abbe", "5a00dbbe",
      "840d04bf", "8c4919bf", "89fd2cbf", "0cf73ebf", "09084fbf", "7c075dbf",
      "7cd168bf", "ae4872bf", "af5379bf", "fae07dbf", "e4e47fbf", "455a7fbf",
      "af427cbf", "d4a576bf", "5d916ebf", "511b64bf", "d35d57bf", "3d7948bf",
      "fe9337bf", "3cd924bf", "897810bf", "904ef5be", "b137c7be", "c02297be",
      "f5164bbe", "6ac0cbbd", "0000bc39", "ae2fcd3d", "9bca4b3e", "c17a973e",
      "6f8cc73e", "d59ff53e", "ab9f103f", "e1fc243f", "70b4373f", "1796483f",
      "9d76573f", "e72f643f", "34a26e3f", "b6b1763f", "404a7c3f", "865d7f3f",
      "b3e37f3f", "7cdb7d3f", "1f49793f", "b139723f", "b0be683f", "53f05c3f",
      "cbec4e3f", "e3d73e3f", "a0db2c3f", "b124193f", "15e6033f", "38acda3e",
      "145dab3e", "81af743e", "7e30103e", "b709293d", "f35e70bd", "57d821be",
      "89fd82be", "f7bfb3be", "bdb6e2be", "14b407bf", "3db21cbf", "8d1f30bf",
      "6bca41bf", "8d8551bf", "ae285fbf", "f3906abf", "11a273bf", "78437abf",
      "cc647ebf", "53fb7fbf", "44037fbf", "ea7e7bbf", "6a7775bf", "2dfb6cbf",
      "6b2162bf", "4c0555bf", "3dc845bf", "379134bf", "618c21bf", "15ea0cbf",
      "16c1edbe", "144cbfbe", "59ee8ebe", "cd453abe", "14a2a9bd", "00008c3c",
      "bb3fef3d", "058d5c3e", "0ea49f3e", "ca68cf3e", "2a1bfd3e", "4a23143f",
      "543d283f", "7ba93a3f", "28384b3f", "f7be593f", "ac18663f", "9e26703f",
      "d1cd773f", "37fb7c3f", "97a17f3f", "2dba7f3f", "d2447d3f", "b846783f",
      "31ce703f", "baed663f", "8cbe5a3f", "cc5f4c3f", "34f63b3f", "acac293f",
      "94b0153f", "a035003f", "95e5d23e", "8844a33e", "0d04643e", "726afe3d",
      "fb16c93c", "e05f9abd", "3cbd32be", "d1408bbe", "edbebbbe", "ea5ceabe",
      "f2500bbf", "a80f20bf", "023533bf", "e88f44bf", "04f453bf", "f13961bf",
      "b53f6cbf", "06ea74bf", "5b217bbf", "4fd67ebf", "88ff7fbf", "db997ebf",
      "1aa97abf", "a53674bf", "fc536bbf", "621760bf", "a29d52bf", "ea0843bf",
      "888131bf", "02341ebf", "8b5209bf", "a522e6be", "3a53b7be", "40af86be",
      "8d6529be", "217787bd", "00800a3d", "519d083e", "8e3d6d3e", "f9c0a73e",
      "6135d73e", "c141023f", "e69b173f", "5a712b3f", "cf903d3f", "5dcb4d3f",
      "73f75b3f", "bef0673f", "a699713f", "0dd8783f", "fc997d3f", "50d37f3f",
      "5f7e7f3f", "659b7c3f", "a932773f", "9e516f3f", "710c653f", "597d583f",
      "77c4493f", "5f07393f", "f571263f", "2932123f", "ebf8f83e", "c210cb3e",
      "45219b3e", "644a533e", "8c65dc3d", "ce36003c", "f280bcbd", "579343be",
      "227993be", "61afc3be", "6ff1f1be", "69e30ebf", "426123bf", "2c3d36bf",
      "f14647bf", "085356bf", "d33a63bf", "f7dd6dbf", "502076bf", "16ed7bbf",
      "8b357fbf", "3df17fbf", "5a1e7ebf", "6ac179bf", "33e572bf", "219c69bf",
      "8cfd5dbf", "1b2750bf", "213c40bf", "74652ebf", "c8d01abf", "acb005bf",
      "f173debe", "654eafbe", "53d07cbe", "cd7b18be", "238d4abd", "71e94e3d",
      "f590193e", "f3dc7d3e", "e1d1af3e", "90f2de3e", "9bec053f", "ac091b3f",
      "19992e3f", "916a403f", "d44f503f", "2e205e3f", "30b8693f", "64fb723f",
      "77d0793f", "99267e3f", "b8f27f3f", "47307f3f", "83e07b3f", "e40c763f",
      "e8c36d3f", "be1a633f", "a82c563f", "ae1a473f", "030c363f", "502b233f",
      "45a90e3f", "b774f13e", "5c2dc33e", "e2f2923e", "7581423e", "1f51ba3d",
      "18c811bc", "9594debd", "995b54be", "e8a69bbe", "0192cbbe", "3473f9be",
      "a56b12bf", "25a726bf", "5b3839bf", "cdef49bf", "a7a258bf", "5e2b65bf",
      "986a6fbf", "f24477bf", "cfa67cbf", "8f827fbf", "afd07fbf", "a4907dbf",
      "f6c778bf", "5f8271bf", "8bd367bf", "aad35bbf", "a0a14dbf", "81613dbf",
      "f43c2bbf", "696217bf", "190502bf"};
  y = zero - ten * ten;
  y_delta = one / ten;
  for (size_t i = 0;
       i < sizeof(another_mem_bytes_arr) / sizeof(another_mem_bytes_arr[0]);
       i++) {
    EXPECT_EQ(mem_bytes(neb::math::sin(y)), another_mem_bytes_arr[i]);
    y += y_delta;
  }
}

TEST(test_common_math, ln) {
  EXPECT_EQ(neb::math::ln(one), std::log(1));

  neb::floatxx_t actual_x = neb::math::ln(one / ten);
  float expect_x = std::log(0.1);
  auto ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x, 1e1 * PRECESION));

  actual_x = neb::math::ln(one / two);
  expect_x = std::log(0.5);
  ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x, 1e1 * PRECESION));

  actual_x = neb::math::ln(two);
  expect_x = std::log(2.0);
  ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x, 1e1 * PRECESION));

  float delta(0.01);
  for (float x = delta; x < 1.0; x += delta) {
    auto actual_x = neb::math::ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    if (std::fabs(expect_x) < PRECESION) {
      continue;
    }

    ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
    if (std::fabs(x) < 0.01 || std::fabs(expect_x) < 0.01) {
      EXPECT_TRUE(ret < precesion(expect_x, 1e3 * PRECESION));
    } else if (std::fabs(x) < 0.1 || std::fabs(expect_x) < 0.1) {
      EXPECT_TRUE(ret < precesion(expect_x, 1e2 * PRECESION));
    } else {
      EXPECT_TRUE(ret < precesion(expect_x, 1e1 * PRECESION));
    }
  }

  delta = 1.0;
  for (float x = 1.0 + delta; x < 100.0; x += delta) {
    auto actual_x = neb::math::ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
    EXPECT_TRUE(ret < precesion(expect_x, 1e2 * PRECESION));
  }

  delta = 10.0;
  for (float x = 1.0 + delta; x < 1000.0; x += delta) {
    auto actual_x = neb::math::ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
    EXPECT_TRUE(ret < precesion(expect_x, 1e3 * PRECESION));
  }

  delta = 100.0;
  for (float x = 1.0 + delta; x < 10000.0; x += delta) {
    auto actual_x = neb::math::ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
    EXPECT_TRUE(ret < precesion(expect_x, 1e3 * PRECESION));
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
  EXPECT_EQ(neb::math::fast_ln(one), std::log(1));

  neb::floatxx_t actual_x = neb::math::fast_ln(one / ten);
  float expect_x = std::log(0.1);
  auto ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x, 1e1 * PRECESION));

  actual_x = neb::math::fast_ln(one / two);
  expect_x = std::log(0.5);
  ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x, PRECESION));

  actual_x = neb::math::fast_ln(two);
  expect_x = std::log(2.0);
  ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
  EXPECT_TRUE(ret < precesion(expect_x, PRECESION));

  float delta(0.01);
  for (float x = delta; x < 1.0; x += delta) {
    auto actual_x = neb::math::fast_ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    if (std::fabs(expect_x) < PRECESION) {
      continue;
    }

    ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
    if (std::fabs(x) < 0.1 || std::fabs(expect_x) < 0.1) {
      EXPECT_TRUE(ret < precesion(expect_x, 1e2 * PRECESION));
    } else {
      EXPECT_TRUE(ret < precesion(expect_x, 1e1 * PRECESION));
    }
  }

  delta = 1.0;
  for (float x = 1.0 + delta; x < 100.0; x += delta) {
    auto actual_x = neb::math::fast_ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
    EXPECT_TRUE(ret < precesion(expect_x, 1e1 * PRECESION));
  }

  delta = 10.0;
  for (float x = 1.0 + delta; x < 1000.0; x += delta) {
    auto actual_x = neb::math::fast_ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
    EXPECT_TRUE(ret < precesion(expect_x, 1e2 * PRECESION));
  }

  delta = 100.0;
  for (float x = 1.0 + delta; x < 10000.0; x += delta) {
    auto actual_x = neb::math::fast_ln(neb::floatxx_t(x));
    auto expect_x = std::log(x);
    ret = neb::math::abs(actual_x, neb::floatxx_t(expect_x));
    EXPECT_TRUE(ret < precesion(expect_x, 1e3 * PRECESION));
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
