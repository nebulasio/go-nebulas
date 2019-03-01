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
#include "common/ir_conf_reader.h"
#include "fs/util.h"
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>
#include <gtest/gtest.h>

TEST(test_boost_read_json, simple) {
  std::string fp =
      neb::fs::join_path(neb::configuration::instance().nbre_root_dir(),
                         "test/data/common/nasir.json");

  boost::property_tree::ptree json_root;
  boost::property_tree::read_json(fp, json_root);
}

TEST(test_common_json_util, read_json) {
  std::string fp =
      neb::fs::join_path(neb::configuration::instance().nbre_root_dir(),
                         "test/data/common/json_util.json");
  neb::ir_conf_reader json_reader(fp);
  EXPECT_EQ(json_reader.self_ref().name(), "nr");
  EXPECT_EQ(json_reader.self_ref().version().major_version(), 0);
  EXPECT_EQ(json_reader.self_ref().version().minor_version(), 0);
  EXPECT_EQ(json_reader.self_ref().version().patch_version(), 1);
  EXPECT_EQ(json_reader.available_height(), 100);
  EXPECT_EQ(json_reader.depends()[0].name(), "yyy");
  EXPECT_EQ(json_reader.depends()[0].version().major_version(), 11);
  EXPECT_EQ(json_reader.depends()[0].version().minor_version(), 12);
  EXPECT_EQ(json_reader.depends()[0].version().patch_version(), 13);
  EXPECT_EQ(json_reader.depends()[1].name(), "zzz");
  EXPECT_EQ(json_reader.depends()[1].version().major_version(), 21);
  EXPECT_EQ(json_reader.depends()[1].version().minor_version(), 22);
  EXPECT_EQ(json_reader.depends()[1].version().patch_version(), 23);
}

TEST(test_common_json_util, throw_json) {
  EXPECT_THROW(neb::ir_conf_reader json_reader("xxx"),
               neb::json_general_failure);

  std::string fp =
      neb::fs::join_path(neb::configuration::instance().nbre_root_dir(),
                         "test/data/common/test_throw_exceptions.json");
  EXPECT_THROW(neb::ir_conf_reader json_reader(fp), neb::json_general_failure);
}
