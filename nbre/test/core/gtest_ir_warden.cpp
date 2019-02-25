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
#include "core/ir_warden.h"
#include "fs/util.h"
#include <gtest/gtest.h>

TEST(test_core, ir_warden_instance_init) {
  auto &instance = neb::core::ir_warden::instance();
}

TEST(test_core, is_sync_already) {
  auto &instance = neb::core::ir_warden::instance();
  instance.on_timer();
  instance.wait_until_sync();
  EXPECT_EQ(instance.is_sync_already(), true);
}

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

TEST(test_core, wait_until_sync) {
  auto &instance = neb::core::ir_warden::instance();
  instance.on_timer();
  instance.wait_until_sync();

  std::string db_read = get_blockchain_path_for_read();
  std::string db_write = get_db_path_for_write();

  neb::fs::rocksdb_storage rs_write;
  rs_write.open_database(db_write, neb::fs::storage_open_for_readonly);
  auto version_bytes = rs_write.get("dip");
  std::vector<uint64_t> versions({1LL << 48});
  size_t gap = sizeof(uint64_t) / sizeof(uint8_t);
  for (size_t i = 0; i < version_bytes.size(); i += gap) {
    neb::byte_t *bytes = version_bytes.value() + i;
    uint64_t version = neb::util::byte_to_number<uint64_t>(bytes, gap);
    EXPECT_EQ(version, versions[i / gap]);
  }

  std::vector<std::pair<uint64_t, neb::block_height_t>> version_and_height(
      {{1LL << 48, 121523}});
  for (auto &it : version_and_height) {
    uint64_t version = it.first;
    neb::block_height_t height = it.second;

    auto payload_bytes = rs_write.get("dip" + std::to_string(version));

    neb::fs::rocksdb_storage rs_read;
    rs_read.open_database(db_read, neb::fs::storage_open_for_readonly);

    neb::util::bytes height_hash =
        rs_read.get_bytes(neb::util::number_to_byte<neb::util::bytes>(height));
    neb::util::bytes block_bytes = rs_read.get_bytes(height_hash);

    std::shared_ptr<corepb::Block> block = std::make_shared<corepb::Block>();
    bool ret = block->ParseFromArray(block_bytes.value(), block_bytes.size());
    if (!ret) {
      throw std::runtime_error("parse block failed");
    }
    auto it_tx = block->transactions().begin();
    auto payload = it_tx->data().payload();
    EXPECT_TRUE(payload_bytes == neb::util::string_to_byte(payload));

    rs_read.close_database();
  }
  rs_write.close_database();
}

TEST(test_core, get_ir_by_name_version) {

  auto &instance = neb::core::ir_warden::instance();

  std::string name = "nr";
  uint64_t version = (1LL << 48);
  auto nbreir_ptr = instance.get_ir_by_name_version(name, version);

  EXPECT_EQ(nbreir_ptr->name(), "nr");
  EXPECT_EQ(nbreir_ptr->version(), 1LL << 48);
  EXPECT_EQ(nbreir_ptr->height(), 90000);
  EXPECT_EQ(nbreir_ptr->depends().size(), 0);
}

TEST(test_core, get_ir_by_name_height) {

  auto &instance = neb::core::ir_warden::instance();

  auto ret_ptr = instance.get_ir_by_name_height("dip", 90000);
  auto ret = *ret_ptr;
  EXPECT_EQ(ret.size(), 1);
  auto it = ret.begin();
  auto nbre_ir_ptr = it;
  EXPECT_EQ(nbre_ir_ptr->name(), "dip");
  EXPECT_EQ(nbre_ir_ptr->version(), 1LL << 48);
  EXPECT_EQ(nbre_ir_ptr->height(), 90000);
  EXPECT_EQ(nbre_ir_ptr->depends_size(), 0);
}

TEST(test_core, ir_warden_instance_dealloc) {
  auto &instance = neb::core::ir_warden::instance();
  instance.release();
}
