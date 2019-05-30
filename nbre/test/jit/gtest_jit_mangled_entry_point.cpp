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
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the // GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//

#include "common/configuration.h"
#include "fs/storage_holder.h"
#include "jit/jit_mangled_entry_point.h"
#include <gtest/gtest.h>

TEST(test_jit_mangled_entry_point, get_mangled_entry_name) {
  auto ret = std::make_unique<neb::jit::jit_mangled_entry_point>();

  auto auth_func_name = neb::configuration::instance().auth_func_name();
  auto auth_func_mangled = ret->get_mangled_entry_name(auth_func_name);
  EXPECT_EQ(auth_func_mangled, "_Z16entry_point_authB5cxx11v");

  auto nr_func_name = neb::configuration::instance().nr_func_name();
  auto nr_func_mangled = ret->get_mangled_entry_name(nr_func_name);
  EXPECT_EQ(nr_func_mangled, "_Z14entry_point_nrB5cxx11yy");

  auto dip_func_name = neb::configuration::instance().dip_func_name();
  auto dip_func_mangled = ret->get_mangled_entry_name(dip_func_name);
  EXPECT_EQ(dip_func_mangled, "_Z15entry_point_dipB5cxx11y");

  neb::fs::storage_holder::instance().release();
}

