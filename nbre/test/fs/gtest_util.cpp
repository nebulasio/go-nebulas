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

TEST(test_util, cur_full_path) {
  std::string cur_path = neb::fs::cur_full_path();
  std::string suffix_cur_path = "/go-nebulas/nbre/bin";
  std::string ret = cur_path.substr(cur_path.size() - suffix_cur_path.size());
  EXPECT_EQ(ret, suffix_cur_path);
}

TEST(test_util, cur_dir) {
  auto cur_dir = neb::fs::cur_dir();
  std::string suffix_cur_dir = "/go-nebulas/nbre";
  auto ret = cur_dir.substr(cur_dir.size() - suffix_cur_dir.size());
  EXPECT_EQ(ret, suffix_cur_dir);
}

TEST(test_util, tmp_dir) { EXPECT_EQ(neb::fs::tmp_dir(), "/tmp"); }

TEST(test_util, join_path) {
  auto cur_dir = neb::fs::cur_dir();
  auto test_data_path = neb::fs::join_path(cur_dir, "test/data");
  std::string suffix_test_data_path = "/go-nebulas/nbre/test/data";
  auto ret = test_data_path.substr(test_data_path.size() -
                                   suffix_test_data_path.size());
  EXPECT_EQ(ret, suffix_test_data_path);
}

TEST(test_util, parent_dir) {
  auto cur_dir = neb::fs::cur_dir();
  auto parent_dir = neb::fs::parent_dir(cur_dir);
  std::string suffix_parent_dir = "/go-nebulas";
  auto ret = parent_dir.substr(parent_dir.size() - suffix_parent_dir.size());
  EXPECT_EQ(ret, suffix_parent_dir);
}

TEST(test_util, is_absolute_path) {
  auto cur_dir = neb::fs::cur_dir();
  EXPECT_EQ(neb::fs::is_absolute_path(cur_dir), true);

  auto nbre_dir = neb::configuration::instance().nbre_root_dir();
  EXPECT_EQ(neb::fs::is_absolute_path(nbre_dir), true);

  EXPECT_EQ(neb::fs::is_absolute_path("../"), false);
}

TEST(test_util, exists) {
  auto cur_dir = neb::fs::cur_dir();
  auto fs_path = neb::fs::join_path(cur_dir, "fs");
  EXPECT_EQ(neb::fs::exists(fs_path), true);

  auto invalid_path = neb::fs::join_path(cur_dir, "invalid");
  EXPECT_EQ(neb::fs::exists(invalid_path), false);
}

TEST(test_util, get_user_name) { EXPECT_EQ(neb::fs::get_user_name(), "usr"); }
