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

#include "common/math.h"
#include <cmath>
#include <gtest/gtest.h>

TEST(test_common_math, simple) {
  float64 e = neb::math::constants<float64>::e();
  std::cout << "e: " << e << std::endl;
  float64 pi = neb::math::constants<float64>::pi();
  std::cout << "pi: " << pi << std::endl;
  float64 ln2 = neb::math::constants<float64>::ln2();
  std::cout << "ln2: " << ln2 << std::endl;

  auto t = neb::math::exp(float64(2));
  std::cout << "e^2: " << t << std::endl;

  float64 ie = e.integer_val();
  float64 ipi = pi.integer_val();
  float64 iln2 = ln2.integer_val();
  std::cout << "ie: " << ie << std::endl;
  std::cout << "ipi: " << ipi << std::endl;
  std::cout << "iln2: " << iln2 << std::endl;
  std::cout << "xxxxxxxxxxxxx" << std::endl;

  auto epi = neb::math::exp(pi);
  auto pi_4 = neb::math::arctan(float64(1));
  auto sin = neb::math::sin(pi / 4);

  auto lne = neb::math::ln(e);
  std::cout << "e^pi: " << epi << std::endl;
  std::cout << "pi/4: " << pi_4 << std::endl;
  std::cout << "sin(pi/4): " << sin << std::endl;
  std::cout << "ln(e): " << lne << std::endl;
}
bool equal(float64 a, float64 b) {
  if (a - b < 1e-6 && b - a < 1e-6) {
    return true;
  }
  return false;
}
TEST(test_common_math, math_functions) {
  std::cout.precision(10);
  // exp
  for (int i = -9; i < 9; ++i) {
    auto e = neb::math::exp(float64(i));
    std::cout << "neb::exp(" << i << "): " << e << std::endl;
    std::cout << "diff from std: " << e - float64(std::exp(i)) << std::endl;
    EXPECT_TRUE(equal(e, float64(std::exp(i))));
  }
  // arctan
  for (int i = -9; i < 9; ++i) {
    auto pi = neb::math::arctan(float64(1));
    std::cout << "neb::arctan(" << 1 << "): " << pi << std::endl;
    std::cout << "diff from std: " << pi - float64(std::atan(1)) << std::endl;
    EXPECT_TRUE(equal(pi, float64(std::atan(1))));
    std::cout << "atan(2): " << std::atan(2) << std::endl;
  }
  // sin
  for (int i = -9; i < 9; ++i) {
    auto s = neb::math::sin(float64(i));
    std::cout << "neb::sin(" << i << "): " << s << std::endl;
    std::cout << "diff from std: " << s - float64(std::sin(i)) << std::endl;
    EXPECT_TRUE(equal(s, float64(std::sin(i))));
  }
  // ln
  for (int i = 1; i < 10; ++i) {
    auto l = neb::math::ln(float64(i));
    std::cout << "neb::ln(" << i << "): " << l << std::endl;
    std::cout << "diff from std: " << l - float64(std::log(i)) << std::endl;
    // EXPECT_TRUE(equal(l, float64(std::log(i))));
  }
  // log2
  for (int i = 1; i < 10; ++i) {
    auto l2 = neb::math::log2(float64(i));
    std::cout << "neb::log2(" << i << "): " << l2 << std::endl;
    std::cout << "diff from std: " << l2 - float64(std::log2(i)) << std::endl;
    // EXPECT_TRUE(equal(l2, float64(std::log2(i))));
  }
  // exp
  for (int i = -10; i < 10; ++i) {
    float64 e = neb::math::constants<float64>::e();
    auto p = neb::math::pow(e, i);
    std::cout << "neb::pow(e, " << i << "): " << p << std::endl;
    std::cout << "diff form std: " << p - std::exp(i) << std::endl;
    // EXPECT_TRUE(equal(p, float64(std::exp(i))));
  }
}
