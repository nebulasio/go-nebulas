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

TEST(test_fs, read_nbre_by_name_version) {

  std::string cur_path = neb::fs::cur_dir();
  std::string db_path = neb::fs::join_path(cur_path, "test_data.db");

  std::shared_ptr<neb::fs::nbre_storage> nbre_ptr =
      std::make_shared<neb::fs::nbre_storage>(db_path, db_path);
  // nbre_ptr->write_nbre_by_height(1000);
  // std::shared_ptr<nbre::NBREIR> nbre_ir_ptr =
  // nbre_ptr->read_nbre_by_name_version("xxx", 666);
  std::shared_ptr<nbre::NBREIR> nbre_ir_ptr;

  nbre::NBREIR nbre_ir = *nbre_ir_ptr;
  EXPECT_EQ(nbre_ir.name(), "xxx");
  EXPECT_EQ(nbre_ir.version(), 666);
  EXPECT_EQ(nbre_ir.height(), 456);

  auto dep = nbre_ir.depends().begin();
  EXPECT_EQ(dep->name(), "xix");
  EXPECT_EQ(dep->version(), 789);

  EXPECT_EQ(nbre_ir.ir(), "heh");
}
