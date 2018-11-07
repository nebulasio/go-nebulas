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

#include "fs/nbre_storage.h"
#include "fs/util.h"
#include "gtest_common.h"
#include <gtest/gtest.h>

TEST(test_fs, write_nbre) {

  std::string db_read = get_db_path_for_read();
  std::string db_write = get_db_path_for_write();

  std::shared_ptr<neb::fs::nbre_storage> nbre_ptr =
      std::make_shared<neb::fs::nbre_storage>(db_write, db_read);

  nbre_ptr->write_nbre();
  nbre_ptr.reset();

  neb::fs::rocksdb_storage rs_write;
  rs_write.open_database(db_write, neb::fs::storage_open_for_readwrite);
  auto max_height_bytes = rs_write.get("nbre_max_height");
  EXPECT_EQ(neb::util::byte_to_number<neb::block_height_t>(max_height_bytes),
            23082);

  auto version_bytes = rs_write.get("nr");
  std::vector<uint64_t> versions({(1LL << 48) + (0LL << 8) + 0LL,
                                  (2LL << 48) + (0LL << 8) + 0LL,
                                  (3LL << 48) + (0LL << 8) + 0LL});
  size_t gap = sizeof(uint64_t) / sizeof(uint8_t);
  for (size_t i = 0; i < version_bytes.size(); i += gap) {
    neb::byte_t *bytes = version_bytes.value() + i;
    uint64_t version = neb::util::byte_to_number<uint64_t>(bytes, gap);
    EXPECT_EQ(version, versions[i / gap]);
  }

  std::vector<std::pair<uint64_t, neb::block_height_t>> version_and_height(
      {{(1LL << 48) + (0LL << 8) + 0LL, 23079},
       {(2LL << 48) + (0LL << 8) + 0LL, 23080},
       {(3LL << 48) + (0LL << 8) + 0LL, 23081}});
  for (auto &it : version_and_height) {
    uint64_t version = it.first;
    neb::block_height_t height = it.second;

    auto payload_bytes = rs_write.get("nr" + std::to_string(version));

    neb::fs::rocksdb_storage rs_read;
    rs_read.open_database(db_read, neb::fs::storage_open_for_readonly);

    neb::util::bytes height_bytes =
        neb::util::number_to_byte<neb::util::bytes>(height);
    neb::util::bytes block_hash_bytes =
        neb::util::string_to_byte(height_bytes.to_hex());
    auto block_bytes = rs_read.get_bytes(block_hash_bytes);

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

TEST(test_fs, read_nbre_by_height_simple) {

  std::string db_read = get_db_path_for_read();
  std::string db_write = get_db_path_for_write();

  std::shared_ptr<neb::fs::nbre_storage> nbre_ptr =
      std::make_shared<neb::fs::nbre_storage>(db_write, db_read);

  auto ret = nbre_ptr->read_nbre_by_height("nr", 1000, true);
  EXPECT_EQ(ret.size(), 1);
  auto it = ret.begin();
  auto nbre_ir_ptr = *it;
  EXPECT_EQ(nbre_ir_ptr->name(), "nr");
  EXPECT_EQ(nbre_ir_ptr->version(), 3LL << 48);
  EXPECT_EQ(nbre_ir_ptr->height(), 150);
  EXPECT_EQ(nbre_ir_ptr->depends_size(), 0);

  ret = nbre_ptr->read_nbre_by_height("nr", 150, true);
  EXPECT_EQ(ret.size(), 1);
  it = ret.begin();
  nbre_ir_ptr = *it;
  EXPECT_EQ(nbre_ir_ptr->name(), "nr");
  EXPECT_EQ(nbre_ir_ptr->version(), 3LL << 48);
  EXPECT_EQ(nbre_ir_ptr->height(), 150);
  EXPECT_EQ(nbre_ir_ptr->depends_size(), 0);

  ret = nbre_ptr->read_nbre_by_height("nr", 149, true);
  EXPECT_EQ(ret.size(), 1);
  it = ret.begin();
  nbre_ir_ptr = *it;
  EXPECT_EQ(nbre_ir_ptr->name(), "nr");
  EXPECT_EQ(nbre_ir_ptr->version(), 1LL << 48);
  EXPECT_EQ(nbre_ir_ptr->height(), 100);
  EXPECT_EQ(nbre_ir_ptr->depends_size(), 0);

  ret = nbre_ptr->read_nbre_by_height("nr", 100, true);
  EXPECT_EQ(ret.size(), 1);
  it = ret.begin();
  nbre_ir_ptr = *it;
  EXPECT_EQ(nbre_ir_ptr->name(), "nr");
  EXPECT_EQ(nbre_ir_ptr->version(), 1LL << 48);
  EXPECT_EQ(nbre_ir_ptr->height(), 100);
  EXPECT_EQ(nbre_ir_ptr->depends_size(), 0);

}

TEST(test_fs, read_nbre_by_height) {

  std::string db_read = get_db_path_for_read();
  std::string db_write = get_db_path_for_write();

  std::shared_ptr<neb::fs::nbre_storage> nbre_ptr =
      std::make_shared<neb::fs::nbre_storage>(db_write, db_read);

  auto ret = nbre_ptr->read_nbre_by_height("dip", 1000, true);
  EXPECT_EQ(ret.size(), 2);
  auto it = ret.begin();
  auto nbre_ir_ptr = *it;
  EXPECT_EQ(nbre_ir_ptr->name(), "nr");
  EXPECT_EQ(nbre_ir_ptr->version(), 2LL << 48);
  EXPECT_EQ(nbre_ir_ptr->height(), 200);
  EXPECT_EQ(nbre_ir_ptr->depends_size(), 0);

  it++;
  nbre_ir_ptr = *it;
  EXPECT_EQ(nbre_ir_ptr->name(), "dip");
  EXPECT_EQ(nbre_ir_ptr->version(), 1);
  EXPECT_EQ(nbre_ir_ptr->height(), 180);
  EXPECT_EQ(nbre_ir_ptr->depends_size(), 1);
}

TEST(test_fs, is_latest_irreversible_block) {

  std::string db_read = get_db_path_for_read();
  std::string db_write = get_db_path_for_write();

  std::shared_ptr<neb::fs::nbre_storage> nbre_ptr =
      std::make_shared<neb::fs::nbre_storage>(db_write, db_read);

  EXPECT_EQ(nbre_ptr->is_latest_irreversible_block(), true);
}

TEST(test_fs, read_nbre_by_name_version) {

  std::string db_read = get_db_path_for_read();
  std::string db_write = get_db_path_for_write();

  std::shared_ptr<neb::fs::nbre_storage> nbre_ptr =
      std::make_shared<neb::fs::nbre_storage>(db_write, db_read);

  std::string name = "dip";
  uint64_t version = 1;
  auto nbreir_ptr = nbre_ptr->read_nbre_by_name_version(name, version);

  EXPECT_EQ(nbreir_ptr->name(), "dip");
  EXPECT_EQ(nbreir_ptr->version(), 1);
  EXPECT_EQ(nbreir_ptr->height(), 180);

  auto it = nbreir_ptr->depends().begin();
  EXPECT_EQ(it->name(), "nr");
  EXPECT_EQ(it->version(), 2LL << 48);
}

TEST(test_fs, get_auth_table) {

  std::string db_read = get_db_path_for_read();
  std::string db_write = get_db_path_for_write();

  std::shared_ptr<neb::fs::nbre_storage> nbre_ptr =
      std::make_shared<neb::fs::nbre_storage>(db_write, db_read);
}
