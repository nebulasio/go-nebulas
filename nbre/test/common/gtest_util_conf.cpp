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

std::string get_configuration_path() {
  return neb::fs::join_path(neb::configuration::instance().root_dir(),
                            "test/data/test_configuration.ini");
}

TEST(test_common_configuration, read_config) {
  EXPECT_EQ(neb::configuration::instance().exec_name(), "");
  EXPECT_EQ(neb::configuration::instance().runtime_library_path(), "");

  std::string conf_file = get_configuration_path();
  const char *argv[3] = {"", "--ini-file", conf_file.c_str()};

  neb::configuration::instance().init_with_args(3, argv);
  EXPECT_EQ(neb::configuration::instance().exec_name(), "bar");
  EXPECT_EQ(neb::configuration::instance().runtime_library_path(), "./lib");
}

TEST(test_common_configuration, throw_config) {
  const char *argv1[3] = {"", "--ini-file", "../test/data/test_xxxx.ini"};

  EXPECT_THROW(neb::configuration::instance().init_with_args(3, argv1),
               neb::configure_general_failure);

  const char *argv2[3] = {"", "--xxxx", "../test/data/test_configuration.ini"};

  EXPECT_THROW(neb::configuration::instance().init_with_args(3, argv2),
               neb::configure_general_failure);
}
