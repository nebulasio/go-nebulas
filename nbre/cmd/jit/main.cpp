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

#include "common/common.h"
#include "common/util/byte.h"
#include "fs/proto/ir.pb.h"
#include "jit/jit_driver.h"
#include <boost/format.hpp>
#include <boost/program_options.hpp>
#include <fstream>

namespace po = boost::program_options;

int main(int argc, char *argv[]) {

  po::options_description desc("Jit driver for auth table execute");
  desc.add_options()("help", "show help message")(
      "module", po::value<std::string>(),
      "Module name")("func", po::value<std::string>(), "function name")(
      "ir_code", po::value<std::string>(), "ir code");

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
  if (!vm.count("func")) {
    std::cout << "You must specify \"func\"!" << std::endl;
    return 1;
  }
  if (!vm.count("ir_code")) {
    std::cout << "You must specify \"ir_code\"!" << std::endl;
    return 1;
  }

  nbre::NBREIR nbre_ir;
  nbre_ir.set_name("auth");
  nbre_ir.set_version(1);
  nbre_ir.set_height(1);
  nbre_ir.clear_depends();

  std::ifstream ifs;

  try {
    std::string ir_fp = vm["ir_code"].as<std::string>();
    ifs.open(ir_fp, std::ios::in | std::ios::binary);

    if (!ifs.is_open()) {
      throw std::invalid_argument(
          boost::str(boost::format("can't open file %1%") % ir_fp));
    }

    ifs.seekg(0, ifs.end);
    std::ifstream::pos_type size = ifs.tellg();
    if (size > 128 * 1024) {
      throw std::invalid_argument("IR file too large!");
    }

    neb::util::bytes buf(size);

    ifs.seekg(0, ifs.beg);
    ifs.read((char *)buf.value(), buf.size());
    if (!ifs)
      throw std::invalid_argument(boost::str(
          boost::format("Read IR file error: only %1% could be read") %
          ifs.gcount()));

    nbre_ir.set_ir(neb::util::byte_to_string(buf));

  } catch (std::exception &e) {
    ifs.close();
    std::cout << e.what() << std::endl;
  }

  auto nbreir_ptr = std::make_shared<nbre::NBREIR>(nbre_ir);
  auto v = std::vector<std::shared_ptr<nbre::NBREIR>>({nbreir_ptr});

  neb::jit_driver jd;
  // jd.run(nullptr, v, vm["func"].as<std::string>(), nullptr);
  jd.get_auth_table(nbre_ir, vm["func"].as<std::string>());

  return 0;
}
