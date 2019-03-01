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
#define PRECESION 1e-5

TEST(test_common_math, constants) {
  neb::floatxx_t actual_e = neb::math::constants<neb::floatxx_t>::e();
  neb::floatxx_t actual_pi = neb::math::constants<neb::floatxx_t>::pi();
  neb::floatxx_t actual_ln2 = neb::math::constants<neb::floatxx_t>::ln2();

  float expect_e = std::exp(1.0);
  float expect_pi = std::acos(-1.0);
  float expect_ln2 = std::log(2.0);

  EXPECT_TRUE(neb::math::abs(actual_e, neb::floatxx_t(expect_e)) < PRECESION);
  EXPECT_TRUE(neb::math::abs(actual_pi, neb::floatxx_t(expect_pi)) < PRECESION);
  EXPECT_TRUE(neb::math::abs(actual_ln2, neb::floatxx_t(expect_ln2)) <
              PRECESION);
}

TEST(test_common_math, exp) {
  EXPECT_EQ(neb::math::exp(neb::floatxx_t(0)), 1);

  neb::floatxx_t actual_x = neb::math::exp(neb::floatxx_t(1));
  float expect_x = std::exp(1.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) < PRECESION);

  actual_x = neb::math::exp(neb::floatxx_t(-1));
  expect_x = std::exp(-1.0);
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) < PRECESION);

  actual_x = neb::math::exp(neb::floatxx_t(80));
  expect_x = std::exp(80.0);
  LOG(INFO) << actual_x << ',' << expect_x << ',' << actual_x - expect_x;
  EXPECT_TRUE(neb::math::abs(actual_x, neb::floatxx_t(expect_x)) < PRECESION);
}
