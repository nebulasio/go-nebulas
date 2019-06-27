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

#include "common/configuration.h"
#include "core/command.h"
#include "fs/blockchain.h"
#include "fs/proto/block.pb.h"
#include "fs/rocksdb_storage.h"
#include "fs/util.h"
#include "gtest_common.h"
#include <gtest/gtest.h>

std::string get_db_path_for_read() {
  std::string cur_path = neb::configuration::instance().nbre_root_dir();
  return neb::fs::join_path(cur_path, "test/data/read-data.db/");
}

std::string get_db_path_for_write() {
  std::string cur_path = neb::configuration::instance().nbre_root_dir();
  return neb::fs::join_path(cur_path, "test/data/write-data.db/");
}

std::string get_blockchain_path_for_read() {
  std::string cur_path = neb::configuration::instance().nbre_root_dir();
  return neb::fs::join_path(cur_path, "../data.db/");
}

TEST(test_rocksdb_storage, open_close_simple) {
  std::string db_path = get_db_path_for_read();
  neb::fs::rocksdb_storage rs;
  EXPECT_NO_THROW(
      rs.open_database(db_path, neb::fs::storage_open_for_readonly));
  EXPECT_NO_THROW(rs.close_database());
}

TEST(test_rocksdb_storage, readonly_multi_open) {
  std::string db_path = get_db_path_for_read();
  neb::fs::rocksdb_storage rs1;
  neb::fs::rocksdb_storage rs2;

  EXPECT_NO_THROW(
      rs1.open_database(db_path, neb::fs::storage_open_for_readonly));
  EXPECT_NO_THROW(
      rs2.open_database(db_path, neb::fs::storage_open_for_readonly));
  EXPECT_NO_THROW(rs1.close_database());
  EXPECT_NO_THROW(rs2.close_database());

  EXPECT_NO_THROW(
      rs1.open_database(db_path, neb::fs::storage_open_for_readonly));
  EXPECT_NO_THROW(
      rs2.open_database(db_path, neb::fs::storage_open_for_readonly));
  EXPECT_NO_THROW(rs2.close_database());
  EXPECT_NO_THROW(rs1.close_database());
}

TEST(test_rocksdb_storage, readonly_and_readwrite_open) {
  std::string db_path = get_db_path_for_read();
  neb::fs::rocksdb_storage rs1;
  neb::fs::rocksdb_storage rs2;

  EXPECT_NO_THROW(
      rs1.open_database(db_path, neb::fs::storage_open_for_readonly));
  EXPECT_NO_THROW(
      rs2.open_database(db_path, neb::fs::storage_open_for_readwrite));
  EXPECT_NO_THROW(rs1.close_database());
  EXPECT_NO_THROW(rs2.close_database());
}

TEST(test_rocksdb_storage, readwrite_multi_open) {
  std::string db_path = get_db_path_for_read();
  neb::fs::rocksdb_storage rs1;
  neb::fs::rocksdb_storage rs2;

  EXPECT_NO_THROW(
      rs1.open_database(db_path, neb::fs::storage_open_for_readwrite));
  EXPECT_THROW(rs2.open_database(db_path, neb::fs::storage_open_for_readwrite),
               neb::fs::storage_general_failure);
}

TEST(test_rocksdb_storage, get_put_del_before_open) {
  neb::fs::rocksdb_storage rs;
  EXPECT_THROW(
      rs.put_bytes(neb::string_to_byte("key"), neb::string_to_byte("val")),
      neb::fs::storage_exception_no_init);
  EXPECT_THROW(rs.get_bytes(neb::string_to_byte("key")),
               neb::fs::storage_exception_no_init);
  EXPECT_THROW(rs.del_by_bytes(neb::string_to_byte("key")),
               neb::fs::storage_exception_no_init);
}

TEST(test_rocksdb_storage, get_put_del) {
  std::string db_path = get_db_path_for_write();
  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readwrite);
  EXPECT_NO_THROW(
      rs.put_bytes(neb::string_to_byte("key"), neb::string_to_byte("val")));
  auto ret = rs.get_bytes(neb::string_to_byte("key"));
  EXPECT_EQ(ret, neb::string_to_byte("val"));
  EXPECT_NO_THROW(rs.del_by_bytes(neb::string_to_byte("key")));
  EXPECT_THROW(rs.get_bytes(neb::string_to_byte("key")),
               neb::fs::storage_general_failure);
  rs.close_database();
}

TEST(test_rocksdb_storage, storage_batch_op) {
  std::string db_path = get_db_path_for_write();
  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readwrite);
  rs.put("123", neb::number_to_byte<neb::bytes>(static_cast<int64_t>(234)));

  auto bytes = rs.get("123");
  int64_t value = neb::byte_to_number<int64_t>(bytes);
  EXPECT_EQ(value, 234);
  rs.close_database();
}

