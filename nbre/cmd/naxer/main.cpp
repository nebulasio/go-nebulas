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

#include "common/ir_conf_reader.h"
#include "common/configuration.h"
#include "core/ir_warden.h"
#include "jit/jit_driver.h"

int main(int argc, char *argv[]) {

  // naxer --module nr --height 1000
  std::string name = "nr";
  neb::block_height_t height = 23083;

  // neb::core::ir_warden::instance().async_run();
//
  // neb::core::ir_warden::instance().wait_until_sync();

  auto irs =
      neb::core::ir_warden::instance().get_ir_by_name_height(name, height);


  const char *argv_jit[3] = {"", "--ini-file",
                         "../test/data/jit_configuration.ini"};

  neb::configuration::instance().init_with_args(3, argv_jit);

  neb::jit_driver jd;
  jd.run(irs);
}
