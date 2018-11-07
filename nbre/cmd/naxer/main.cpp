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
#include "core/ir_warden.h"
#include "jit/jit_driver.h"
#include <boost/program_options.hpp>

namespace po = boost::program_options;

int main(int argc, char *argv[]) {
  po::options_description desc("Rocksdb read and write");
  desc.add_options()("help", "show help message")(
      "module", po::value<std::string>(),
      "Module name")("func", po::value<std::string>(), "function name")(
      "height", po::value<neb::block_height_t>(), "block height");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);

  if (vm.count("help")) {
    std::cout << desc << "\n";
    return 1;
  }

  if (!vm.count("module")) {
    std::cout << "You must specify \"module\"!" << std::endl;
    return 1;
  }
  if (!vm.count("height")) {
    std::cout << "You must specify \"height\"!" << std::endl;
    return 1;
  }

  // naxer --module nr --height 1000

  neb::core::ir_warden::instance().on_timer();
  neb::core::ir_warden::instance().wait_until_sync();

  std::string module = vm["module"].as<std::string>();
  neb::block_height_t height = vm["height"].as<neb::block_height_t>();
  auto irs =
      neb::core::ir_warden::instance().get_ir_by_name_height(module, height);

  neb::jit_driver &jd = neb::jit_driver::instance();
  jd.run_ir<int, neb::core::driver *, void *>(
      module, height, vm["func"].as<std::string>(), nullptr, nullptr);
  // jd.run(nullptr, irs, vm["func"].as<std::string>(), nullptr);
}
