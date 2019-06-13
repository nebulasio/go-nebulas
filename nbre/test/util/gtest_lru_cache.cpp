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
#include "util/lru_cache.h"
#include <gtest/gtest.h>

TEST(test_lru_cache, construct) {
  neb::util::lru_cache<int32_t, int32_t> lc;
  EXPECT_EQ(lc.size(), 0);
  EXPECT_TRUE(lc.empty());
}

TEST(test_lru_cache, size) {
  neb::util::lru_cache<int32_t, int32_t, std::mutex, 1, 2> lc;
  EXPECT_EQ(lc.size(), 0);
  lc.set(1, 1);
  EXPECT_EQ(lc.size(), 1);
  std::this_thread::sleep_for(std::chrono::seconds(2));
  EXPECT_EQ(lc.size(), 0);
}

TEST(test_lru_cache, empty) {
  neb::util::lru_cache<int32_t, int32_t, std::mutex, 1, 2> lc;
  EXPECT_TRUE(lc.empty());
  lc.set(1, 1);
  EXPECT_TRUE(!lc.empty());
  std::this_thread::sleep_for(std::chrono::seconds(2));
  EXPECT_TRUE(lc.empty());
}

TEST(test_lru_cache, clear) {
  neb::util::lru_cache<int32_t, int32_t, std::mutex, 1, 2> lc;
  EXPECT_TRUE(lc.empty());
  lc.set(1, 1);
  EXPECT_TRUE(!lc.empty());
  lc.clear();
  EXPECT_TRUE(lc.empty());
}

TEST(test_lru_cache, set) {
  neb::util::lru_cache<int32_t, int32_t, std::mutex, 1, 2> lc;
  EXPECT_TRUE(lc.empty());
  lc.set(1, 1);
  EXPECT_EQ(lc.size(), 1);
  lc.set(1, 2);
  EXPECT_EQ(lc.size(), 1);
  lc.set(2, 2);
  EXPECT_EQ(lc.size(), 2);
}

TEST(test_lru_cache, get) {
  neb::util::lru_cache<int32_t, int32_t, std::mutex, 1, 2> lc;
  EXPECT_TRUE(lc.empty());
  lc.set(1, 1);

  int val;
  auto ret = lc.get(1, val);
  EXPECT_TRUE(ret);
  EXPECT_EQ(val, 1);

  ret = lc.get(2, val);
  EXPECT_TRUE(!ret);

  lc.set(2, 3);
  ret = lc.get(2, val);
  EXPECT_TRUE(ret);
  EXPECT_EQ(val, 3);
}

TEST(test_lru_cache, exists) {
  neb::util::lru_cache<int32_t, int32_t, std::mutex, 1, 2> lc;
  EXPECT_TRUE(!lc.exists(0));
  EXPECT_TRUE(!lc.exists(1));

  lc.set(1, 1);
  EXPECT_TRUE(!lc.exists(0));
  EXPECT_TRUE(lc.exists(1));

  std::this_thread::sleep_for(std::chrono::seconds(2));
  lc.set(2, 2);
  EXPECT_TRUE(!lc.exists(0));
  EXPECT_TRUE(!lc.exists(1));
  EXPECT_TRUE(lc.exists(2));
}
