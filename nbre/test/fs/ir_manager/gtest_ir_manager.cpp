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

#include "common/version.h"
#include "core/net_ipc/nipc_pkg.h"
#include "fs/ir_manager/ir_manager.h"
#include "fs/storage_holder.h"
#include "fs/util.h"
#include "test/fs/gtest_common.h"
#include <gtest/gtest.h>

typedef std::pair<std::string, neb::version> depend_t;
void gen_ir(const std::string &name, const neb::version &v,
            neb::block_height_t height, const std::vector<depend_t> &depends,
            neb::fs::rocksdb_storage *rs) {

  nbre::NBREIR ir;
  ir.set_name(name);
  ir.set_version(v.data());
  ir.set_height(height);
  for (auto &dep : depends) {
    auto deps = ir.add_depends();
    deps->set_name(dep.first);
    deps->set_version(dep.second.data());
  }

  auto size = ir.ByteSizeLong();
  neb::bytes buf(size);
  ir.SerializeToArray((void *)buf.value(), buf.size());

  neb::fs::ir_manager_helper::update_ir_list(name, rs);
  neb::fs::ir_manager_helper::update_ir_versions(name, v.data(), rs);
  neb::fs::ir_manager_helper::deploy_ir(name, v.data(), buf, rs);
}

std::shared_ptr<neb::fs::ir_manager> nbre_ptr;
TEST(test_fs, init) { nbre_ptr = std::make_shared<neb::fs::ir_manager>(); }

TEST(test_fs, read_nbre_by_height_simple) {

  std::string name = "nr";
  neb::version v(0, 0, 1);
  neb::block_height_t height = 123;
  std::vector<depend_t> depends;
  auto rs = neb::fs::storage_holder::instance().nbre_db_ptr();
  gen_ir(name, v, height, depends, rs);

  auto ret_ptr = nbre_ptr->read_irs(name, height + 1, true);
  auto ret = *ret_ptr;
  EXPECT_EQ(ret.size(), 1);
  auto it = ret.begin();
  auto nbre_ir_ptr = it;
  EXPECT_EQ(nbre_ir_ptr->name(), name);
  EXPECT_EQ(nbre_ir_ptr->version(), v.data());
  EXPECT_EQ(nbre_ir_ptr->height(), height);
  EXPECT_EQ(nbre_ir_ptr->depends_size(), 0);

  ret_ptr = nbre_ptr->read_irs(name, height, true);
  ret = *ret_ptr;
  EXPECT_EQ(ret.size(), 1);
  it = ret.begin();
  nbre_ir_ptr = it;
  EXPECT_EQ(nbre_ir_ptr->name(), name);
  EXPECT_EQ(nbre_ir_ptr->version(), v.data());
  EXPECT_EQ(nbre_ir_ptr->height(), height);
  EXPECT_EQ(nbre_ir_ptr->depends_size(), 0);

  ret_ptr = nbre_ptr->read_irs(name, height - 1, true);
  ret = *ret_ptr;
  EXPECT_EQ(ret.size(), 0);
}

TEST(test_fs, read_nbre_by_height) {

  std::string name = "dip";
  neb::version v(1, 2, 3);
  neb::block_height_t height = 456;
  std::vector<depend_t> depends;
  depends.push_back(std::make_pair("nr", neb::version(0, 0, 1)));
  auto rs = neb::fs::storage_holder::instance().nbre_db_ptr();
  gen_ir(name, v, height, depends, rs);

  auto ret_ptr = nbre_ptr->read_irs(name, height, true);
  auto ret = *ret_ptr;
  EXPECT_EQ(ret.size(), 2);
  auto it = ret.begin();
  auto nbre_ir_ptr = it;
  EXPECT_EQ(nbre_ir_ptr->name(), name);
  EXPECT_EQ(nbre_ir_ptr->version(), v.data());
  EXPECT_EQ(nbre_ir_ptr->height(), height);
  EXPECT_EQ(nbre_ir_ptr->depends_size(), 1);
}

TEST(test_fs, read_nbre_by_name_version) {

  std::string name = "dip";
  neb::version v(1, 2, 3);
  neb::block_height_t height = 456;

  std::shared_ptr<neb::fs::ir_manager> nbre_ptr =
      std::make_shared<neb::fs::ir_manager>();

  auto nbreir_ptr = nbre_ptr->read_ir(name, v.data());
  EXPECT_EQ(nbreir_ptr->name(), name);
  EXPECT_EQ(nbreir_ptr->version(), v.data());
  EXPECT_EQ(nbreir_ptr->height(), height);
  EXPECT_EQ(nbreir_ptr->depends_size(), 1);
}

TEST(test_fs, get_auth_table) {}
