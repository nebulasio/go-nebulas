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

#include "fs/rocksdb_session_storage.h"
#include "test/fs/gtest_common.h"
#include <gtest/gtest.h>

TEST(test_rocksdb_session_storage, get_put_del_before_init) {
  neb::fs::rocksdb_session_storage rss;
  EXPECT_THROW(
      rss.put_bytes(neb::string_to_byte("key"), neb::string_to_byte("val")),
      std::exception);
  EXPECT_THROW(rss.del_by_bytes(neb::string_to_byte("key")), std::exception);
}

TEST(test_rocksdb_session_storage, get_put_delete) {
  std::string db_path = get_db_path_for_write();
  neb::fs::rocksdb_session_storage rss;
  rss.init(db_path, neb::fs::storage_open_for_readwrite);

  EXPECT_NO_THROW(
      rss.put_bytes(neb::string_to_byte("key"), neb::string_to_byte("val")));
  auto ret = rss.get_bytes(neb::string_to_byte("key"));
  EXPECT_EQ(ret, neb::string_to_byte("val"));
  EXPECT_NO_THROW(rss.del_by_bytes(neb::string_to_byte("key")));
  EXPECT_THROW(rss.get_bytes(neb::string_to_byte("key")),
               neb::fs::storage_general_failure);
}
