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

#include "common/configuration.h"
#include "jit/jit_driver.h"
#include "jit/jit_mangled_entry_point.h"

int main(int argc, char *argv[]) {

  auto n1 = neb::jit_driver::instance().get_mangled_entry_point(
      neb::configuration::instance().nr_func_name());
  LOG(INFO) << "nr entry point: " << n1;
  auto n2 = neb::jit_driver::instance().get_mangled_entry_point(
      neb::configuration::instance().dip_func_name());
  LOG(INFO) << "dip entry point: " << n2;
  auto n3 = neb::jit_driver::instance().get_mangled_entry_point(
      neb::configuration::instance().auth_func_name());
  LOG(INFO) << "auth entry point: " << n3;
  return 0;
}
