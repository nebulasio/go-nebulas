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
#include "crypto/hash.h"
#include <gtest/gtest.h>

TEST(test_sha3_hash, simple) {
  std::unordered_map<std::string, neb::crypto::sha3_256_hash_t> test_hashs;
  test_hashs.insert(std::make_pair(
      "", neb::crypto::sha3_256_hash_t(
              {167, 255, 198, 248, 191, 30,  215, 102, 81, 193, 71,
               86,  160, 97,  214, 98,  245, 128, 255, 77, 228, 59,
               73,  250, 130, 216, 10,  75,  128, 248, 67, 74})));
  test_hashs.insert(std::make_pair(
      "Hello, world", neb::crypto::sha3_256_hash_t(
                          {53,  80, 171, 169, 116, 146, 222, 56,  175, 48,  102,
                           240, 21, 127, 197, 50,  219, 103, 145, 179, 125, 83,
                           38,  44, 231, 104, 141, 204, 93,  70,  24,  86})));
  test_hashs.insert(
      std::make_pair("https://nebulas.io",
                     neb::crypto::sha3_256_hash_t(
                         {94,  159, 238, 157, 152, 227, 18, 248, 53,  8,  13,
                          247, 231, 21,  17,  14,  172, 34, 192, 157, 24, 175,
                          119, 254, 126, 208, 174, 17,  14, 77,  1,   55})));

  for (auto it : test_hashs) {
    auto r = neb::crypto::sha3_256_hash(it.first);
    EXPECT_TRUE(r == it.second);
  }
}
