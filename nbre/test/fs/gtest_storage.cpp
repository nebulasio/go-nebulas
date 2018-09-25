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

#include "fs/blockchain.h"
#include "fs/proto/block.pb.h"
#include "fs/rocksdb_storage.h"
#include "fs/util.h"
#include "core/command.h"
#include <gtest/gtest.h>

typedef std::shared_ptr<corepb::Block> block_shared_ptr;

std::string get_db_path_for_read() {
  std::string cur_path = neb::fs::cur_dir();
  return neb::fs::join_path(cur_path, "test/data/data.db/");
}

std::string get_db_path_for_write() {
  std::string cur_path = neb::fs::cur_dir();
  return neb::fs::join_path(cur_path, "test/data/data_test_write.db/");
}

TEST(test_fs, positive_storage_read_bc) {
  std::string db_path = get_db_path_for_read(); 

  neb::fs::rocksdb_storage rs;
  EXPECT_THROW(rs.get(neb::fs::blockchain::Block_LIB),
               neb::fs::storage_exception_no_init);
  EXPECT_THROW(
      rs.put(neb::fs::blockchain::Block_LIB, neb::util::string_to_byte("xxx")),
      neb::fs::storage_exception_no_init);
  EXPECT_THROW(rs.del(neb::fs::blockchain::Block_LIB),
               neb::fs::storage_exception_no_init);

  rs.open_database(db_path, neb::fs::storage_open_for_readonly);
  neb::fs::rocksdb_storage rs2;
  rs2.open_database(db_path, neb::fs::storage_open_for_readonly);

  auto tail_block_hash = rs.get(neb::fs::blockchain::Block_LIB);

  auto tail_bytes = rs.get_bytes(tail_block_hash);

  corepb::Block block;
  block.ParseFromArray(tail_bytes.value(), tail_bytes.size());
  rs.close_database();
}

TEST(test_fs, storage_read_write) {
  std::string db_path = get_db_path_for_read(); 

  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readonly);
  neb::fs::rocksdb_storage rs2;
  rs2.open_database(db_path, neb::fs::storage_open_for_readwrite);
}

TEST(test_fs, storage_write_write) {
  std::string db_path = get_db_path_for_read();

  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readwrite);
  neb::fs::rocksdb_storage rs2;
  EXPECT_THROW(rs2.open_database(db_path, neb::fs::storage_open_for_readwrite),
               neb::fs::storage_general_failure);
}

TEST(test_fs, storage_batch_op) {
  std::string db_path = get_db_path_for_write();
  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readwrite);
  // rs.enable_batch();
  rs.put("123", neb::util::number_to_byte<neb::util::bytes>(
                    static_cast<int64_t>(234)));
  // rs.put(static_cast<int64_t>(124), static_cast<int64_t>(256));
  // rs.put(static_cast<int64_t>(125), static_cast<int64_t>(267));
  // rs.del(static_cast<int64_t>(124));
  // rs.flush();

  auto bytes = rs.get("123");
  int64_t value = neb::util::byte_to_number<int64_t>(bytes);
  EXPECT_EQ(value, 234);
}

TEST(test_fs, read_blockchain_transcations) {
  std::string db_path = get_db_path_for_read();

  neb::fs::blockchain block_chain(db_path);
  block_shared_ptr block =  block_chain.load_block_with_height(1);
  auto header = block->header();
  EXPECT_EQ(header.timestamp(), 0);

  auto txs = block->transactions();
  auto tx = txs.begin();
  EXPECT_EQ(tx->nonce(), 1);
  EXPECT_EQ(tx->chain_id(), 1003);

  std::string hash = tx->hash();
  std::string hash_hex = neb::util::string_to_byte(hash).to_hex();
  EXPECT_EQ(hash_hex, "f2a9f39bbc01d7063eaf8e703da268c043124a732ddfc69fdcbac79dd071eadd");

  std::string from = tx->from();
  std::string from_base58 = neb::util::string_to_byte(from).to_base58();
  EXPECT_EQ(from_base58, "n1HQb7j3Ctpj5A9sJnfhRwXS549nGPv2EkD");

  std::string value = tx->value();
  std::string value_hex = neb::util::string_to_byte(value).to_hex();
  EXPECT_EQ(value_hex, "00000000000000000000000000000000");

  std::string price = tx->gas_price();
  std::string price_hex = neb::util::string_to_byte(price).to_hex();
  EXPECT_EQ(price.size(), 16);
  EXPECT_EQ(price_hex.size(), 32);
  EXPECT_EQ(price_hex, "000000000000000000000000000f4240");

  auto data = tx->data();
  std::string type = data.type();
  EXPECT_EQ(type, "binary");

  std::string payload = data.payload();
  std::string payload_base64 = neb::util::string_to_byte(payload).to_base58();
  std::string want_base64("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");
  EXPECT_EQ(payload_base64, want_base64);
}

TEST(test_fs, throw_operation_db) {
  std::string db_path = get_db_path_for_read(); 

  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readonly);

  EXPECT_THROW(rs.open_database(db_path, neb::fs::storage_open_for_readonly),
      std::runtime_error);

  rs.close_database();
  EXPECT_THROW(rs.open_database("", neb::fs::storage_open_for_readonly),
      neb::fs::storage_general_failure);

  EXPECT_THROW(rs.get(neb::fs::blockchain::Block_LIB),
      neb::fs::storage_exception_no_init);

  EXPECT_THROW(
      rs.put(neb::fs::blockchain::Block_LIB, neb::util::string_to_byte("xxx")),
      neb::fs::storage_exception_no_init);

  EXPECT_THROW(rs.del(neb::fs::blockchain::Block_LIB),
      neb::fs::storage_exception_no_init);

  std::string db_path_readwrite = get_db_path_for_write(); 
  rs.open_database(db_path_readwrite, neb::fs::storage_open_for_readwrite);

  EXPECT_THROW(rs.get("no_exist"),
      neb::fs::storage_general_failure);
}
