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
#include "fs/util.h"
#include <gtest/gtest.h>
#include <iostream>

TEST(test_fs_util, simple) {
  std::string cur_path = neb::fs::cur_full_path();
  std::string cur_dir = neb::configuration::instance().nbre_root_dir();
  std::string tmp_dir = neb::fs::tmp_dir();

  EXPECT_TRUE(cur_path.size() > 0);
  EXPECT_TRUE(cur_dir.size() > 0);
  EXPECT_TRUE(tmp_dir.size() > 0);

  std::string rocksdb_data_path =
      neb::fs::join_path(cur_dir, "test/data/read-data.db/CURRENT");

  EXPECT_TRUE(neb::fs::exists(rocksdb_data_path));
}
