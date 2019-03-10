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

#include "core/net_ipc/nipc_pkg.h"
#include "fs/ir_manager/ir_manager.h"
#include "fs/util.h"
#include "test/fs/gtest_common.h"
#include <gtest/gtest.h>

TEST(test_fs, write_nbre_until_sync) {

  std::string db_read = get_blockchain_path_for_read();
  std::string db_write = get_db_path_for_write();

  std::shared_ptr<neb::fs::ir_manager> nbre_ptr =
      std::make_shared<neb::fs::ir_manager>();

  neb::util::wakeable_queue<std::shared_ptr<nbre_ir_transactions_req>> q;
  nbre_ptr->parse_irs(q);
  nbre_ptr.reset();

  neb::fs::rocksdb_storage rs_write;
  rs_write.open_database(db_write, neb::fs::storage_open_for_readonly);
  auto max_height_bytes = rs_write.get("nbre_max_height");

  auto version_bytes = rs_write.get("nr");
  std::vector<uint64_t> versions({(1LL << 48) + (0LL << 32) + 0LL,
                                  (1LL << 48) + (0LL << 32) + 0LL,
                                  (2LL << 48) + (0LL << 32) + 0LL});
  size_t gap = sizeof(uint64_t) / sizeof(uint8_t);
  for (size_t i = 0; i < version_bytes.size(); i += gap) {
    neb::byte_t *bytes = version_bytes.value() + i;
    uint64_t version = neb::byte_to_number<uint64_t>(bytes, gap);
    EXPECT_EQ(version, versions[i / gap]);
  }

  std::vector<std::pair<uint64_t, neb::block_height_t>> version_and_height(
      {{(1LL << 48) + (0LL << 32) + 0LL, 121279},
       {(2LL << 48) + (0LL << 32) + 0LL, 121462}});
  for (auto &it : version_and_height) {
    uint64_t version = it.first;
    neb::block_height_t height = it.second;

    auto payload_bytes = rs_write.get("nr" + std::to_string(version));

    neb::fs::rocksdb_storage rs_read;
    rs_read.open_database(db_read, neb::fs::storage_open_for_readonly);

    neb::bytes height_hash =
        rs_read.get_bytes(neb::number_to_byte<neb::bytes>(height));
    neb::bytes block_bytes = rs_read.get_bytes(height_hash);

    std::shared_ptr<corepb::Block> block = std::make_shared<corepb::Block>();
    bool ret = block->ParseFromArray(block_bytes.value(), block_bytes.size());
    if (!ret) {
      throw std::runtime_error("parse block failed");
    }
    auto it_tx = block->transactions().begin();
    auto payload = it_tx->data().payload();
    EXPECT_EQ(neb::string_to_byte(payload).to_hex(), payload_bytes.to_hex());

    rs_read.close_database();
  }
  rs_write.close_database();
}

TEST(test_fs, read_nbre_by_height_simple) {

  std::string db_read = get_blockchain_path_for_read();
  std::string db_write = get_db_path_for_write();

  std::shared_ptr<neb::fs::ir_manager> nbre_ptr =
      std::make_shared<neb::fs::ir_manager>();

  auto ret_ptr = nbre_ptr->read_irs("nr", 90001, true);
  auto ret = *ret_ptr;
  EXPECT_EQ(ret.size(), 1);
  auto it = ret.begin();
  auto nbre_ir_ptr = it;
  EXPECT_EQ(nbre_ir_ptr->name(), "nr");
  EXPECT_EQ(nbre_ir_ptr->version(), 2LL << 48);
  EXPECT_EQ(nbre_ir_ptr->height(), 90000);
  EXPECT_EQ(nbre_ir_ptr->depends_size(), 0);

  ret_ptr = nbre_ptr->read_irs("nr", 90000, true);
  ret = *ret_ptr;
  EXPECT_EQ(ret.size(), 1);
  it = ret.begin();
  nbre_ir_ptr = it;
  EXPECT_EQ(nbre_ir_ptr->name(), "nr");
  EXPECT_EQ(nbre_ir_ptr->version(), 2LL << 48);
  EXPECT_EQ(nbre_ir_ptr->height(), 90000);
  EXPECT_EQ(nbre_ir_ptr->depends_size(), 0);

  ret_ptr = nbre_ptr->read_irs("nr", 89999, true);
  ret = *ret_ptr;
  EXPECT_EQ(ret.size(), 0);
}

TEST(test_fs, read_nbre_by_height) {

  std::string db_read = get_blockchain_path_for_read();
  std::string db_write = get_db_path_for_write();

  std::shared_ptr<neb::fs::ir_manager> nbre_ptr =
      std::make_shared<neb::fs::ir_manager>();

  auto ret_ptr = nbre_ptr->read_irs("dip", 90000, true);
  auto ret = *ret_ptr;
  EXPECT_EQ(ret.size(), 1);
  auto it = ret.begin();
  auto nbre_ir_ptr = it;
  EXPECT_EQ(nbre_ir_ptr->name(), "dip");
  EXPECT_EQ(nbre_ir_ptr->version(), 1LL << 48);
  EXPECT_EQ(nbre_ir_ptr->height(), 90000);
  EXPECT_EQ(nbre_ir_ptr->depends_size(), 0);
}

TEST(test_fs, read_nbre_by_name_version) {

  std::string db_read = get_blockchain_path_for_read();
  std::string db_write = get_db_path_for_write();

  std::shared_ptr<neb::fs::ir_manager> nbre_ptr =
      std::make_shared<neb::fs::ir_manager>();

  std::string name = "dip";
  uint64_t version = 1LL << 48;
  auto nbreir_ptr = nbre_ptr->read_ir(name, version);

  EXPECT_EQ(nbreir_ptr->name(), "dip");
  EXPECT_EQ(nbreir_ptr->version(), 1LL << 48);
  EXPECT_EQ(nbreir_ptr->height(), 90000);
  EXPECT_EQ(nbreir_ptr->depends_size(), 0);
}

TEST(test_fs, get_auth_table) {

  std::string db_read = get_blockchain_path_for_read();
  std::string db_write = get_db_path_for_write();

  std::shared_ptr<neb::fs::ir_manager> nbre_ptr =
      std::make_shared<neb::fs::ir_manager>();
}
