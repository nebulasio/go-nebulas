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

#include "fs/blockchain/trie/byte_shared.h"
#include <gtest/gtest.h>

TEST(test_byte_shared, constructor_default) {
  neb::fs::byte_shared bs;
  EXPECT_EQ(bs.bits_low(), 0x0);
  EXPECT_EQ(bs.bits_high(), 0x0);
  EXPECT_EQ(bs.data(), 0x0);
}

TEST(test_byte_shared, constructor_data) {
  neb::byte_t data = 0xff;
  neb::fs::byte_shared bs(data);
  EXPECT_EQ(bs.bits_low(), 0xf);
  EXPECT_EQ(bs.bits_high(), 0xf);
  EXPECT_EQ(bs.data(), 0xff);

  data = 0x1;
  bs = neb::fs::byte_shared(data);
  EXPECT_EQ(bs.bits_low(), 0x1);
  EXPECT_EQ(bs.bits_high(), 0x0);
  EXPECT_EQ(bs.data(), 0x1);

  data = 0x10;
  bs = neb::fs::byte_shared(data);
  EXPECT_EQ(bs.bits_low(), 0x0);
  EXPECT_EQ(bs.bits_high(), 0x1);
  EXPECT_EQ(bs.data(), 0x10);
}

TEST(test_byte_shared, constructor_lh) {
  neb::fs::byte_shared bs(0x0, 0x1);
  EXPECT_EQ(bs.bits_low(), 0x0);
  EXPECT_EQ(bs.bits_high(), 0x1);
  EXPECT_EQ(bs.data(), 0x10);

  bs = neb::fs::byte_shared(0x1, 0x0);
  EXPECT_EQ(bs.bits_low(), 0x1);
  EXPECT_EQ(bs.bits_high(), 0x0);
  EXPECT_EQ(bs.data(), 0x1);

  bs = neb::fs::byte_shared(0xa, 0x8);
  EXPECT_EQ(bs.bits_low(), 0xa);
  EXPECT_EQ(bs.bits_high(), 0x8);
  EXPECT_EQ(bs.data(), 0x8a);
}
