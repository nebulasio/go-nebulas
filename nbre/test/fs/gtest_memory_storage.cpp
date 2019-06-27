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

#include "fs/memory_storage.h"
#include <gtest/gtest.h>

TEST(test_memory_storage, get_put_del) {
  neb::fs::memory_storage ms;
  EXPECT_NO_THROW(
      ms.put_bytes(neb::string_to_byte("key"), neb::string_to_byte("val")));
  auto ret = ms.get_bytes(neb::string_to_byte("key"));
  EXPECT_EQ(ret, neb::string_to_byte("val"));
  EXPECT_NO_THROW(ms.del_by_bytes(neb::string_to_byte("key")));
  EXPECT_THROW(ms.get_bytes(neb::string_to_byte("key")),
               neb::fs::storage_general_failure);
}
