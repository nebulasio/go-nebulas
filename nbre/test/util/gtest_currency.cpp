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
#include "common/nebulas_currency.h"
#include <gtest/gtest.h>

TEST(test_currency, simple) {
  neb::nas v3(1089238_nas);
  neb::nas_storage_t vs = neb::nas_to_storage(v3);
  neb::nas v4 = neb::storage_to_nas<neb::nas>(vs);
  EXPECT_EQ(v3, v4);
}

TEST(test_currency, neg) {
  neb::nas v0;
  neb::nas v1(1089238_nas);
  neb::nas v3 = v0 - v1;
  neb::nas_storage_t vs = neb::nas_to_storage(v3);
  neb::nas v4 = neb::storage_to_nas<neb::nas>(vs);
  EXPECT_EQ(v3, v4);
}

TEST(test_currency, to_storage_and_from_storage) {
  neb::nas v0 = 1_nas;
  std::cout << v0 << std::endl;
  auto tmp = neb::wei_to_storage(v0.wei_value());
  neb::wei v1 = neb::storage_to_wei(tmp);
  std::cout << v1 << std::endl;
}
