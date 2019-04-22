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
#include "fs/util.h"
#include "gtest_common.h"
#include <gtest/gtest.h>

TEST(test_fs, read_blockchain_transcations) {
  std::string db_path = get_db_path_for_read();

  neb::fs::bc_storage_session::instance().init(db_path,
                                               neb::fs::storage_open_default);

  block_ptr_t block = neb::fs::blockchain::load_block_with_height(1);
  auto header = block->header();
  EXPECT_EQ(header.timestamp(), 0);

  auto txs = block->transactions();
  auto tx = txs.begin();
  EXPECT_EQ(tx->nonce(), 1);
  EXPECT_EQ(tx->chain_id(), 1003);

  std::string hash = tx->hash();
  std::string hash_hex = neb::string_to_byte(hash).to_hex();
  EXPECT_EQ(hash_hex,
            "f2a9f39bbc01d7063eaf8e703da268c043124a732ddfc69fdcbac79dd071eadd");

  std::string from = tx->from();
  std::string from_base58 = neb::string_to_byte(from).to_base58();
  EXPECT_EQ(from_base58, "n1HQb7j3Ctpj5A9sJnfhRwXS549nGPv2EkD");

  std::string value = tx->value();
  std::string value_hex = neb::string_to_byte(value).to_hex();
  EXPECT_EQ(value_hex, "00000000000000000000000000000000");

  std::string price = tx->gas_price();
  std::string price_hex = neb::string_to_byte(price).to_hex();
  EXPECT_EQ(price.size(), 16);
  EXPECT_EQ(price_hex.size(), 32);
  EXPECT_EQ(price_hex, "000000000000000000000000000f4240");

  auto data = tx->data();
  std::string type = data.type();
  EXPECT_EQ(type, "binary");

  std::string payload = data.payload();
  std::string payload_base64 = neb::string_to_byte(payload).to_base58();
  std::string want_base64(
      "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"
      "S");
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
      rs.put(neb::fs::blockchain::Block_LIB, neb::string_to_byte("xxx")),
      neb::fs::storage_exception_no_init);

  EXPECT_THROW(rs.del(neb::fs::blockchain::Block_LIB),
               neb::fs::storage_exception_no_init);

  std::string db_path_readwrite = get_db_path_for_write();
  rs.open_database(db_path_readwrite, neb::fs::storage_open_for_readwrite);

  EXPECT_THROW(rs.get("no_exist"), neb::fs::storage_general_failure);
}


TEST(test_fs, load_LIB_block) {

  std::string db_path = get_db_path_for_read();

  neb::fs::bc_storage_session::instance().init(db_path,
                                               neb::fs::storage_open_default);
  std::shared_ptr<corepb::Block> block_ptr =
      neb::fs::blockchain::load_LIB_block();

  EXPECT_EQ(block_ptr->height(), 23078);
  EXPECT_EQ(block_ptr->transactions_size(), 0);

  auto header = block_ptr->header();
  auto hash_bytes = neb::string_to_byte(header.hash());
  EXPECT_EQ(hash_bytes.to_hex(),
            "8da918837dfa46c6d689eb3be3dc2eff4dffb6dca8a1027df0eddd0d0597af8f");
  EXPECT_EQ(header.timestamp(), 1531895265);
  EXPECT_EQ(header.chain_id(), 1003);
}
