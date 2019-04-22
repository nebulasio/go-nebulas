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

#include "common/address.h"
#include "common/byte.h"
#include "jit/jit_driver.h"
#include <fstream>
#include <random>
#include <sstream>

void randomly_gen_irs(const std::string &ir_path,
                      std::vector<nbre::NBREIR> &irs) {
  std::string name = std::to_string(std::rand());
  uint64_t version = std::rand();
  uint64_t height = std::rand();

  std::ifstream ifs;
  ifs.open(ir_path.c_str(), std::ios::in | std::ios::binary);
  ifs.seekg(0, ifs.end);
  std::ifstream::pos_type size = ifs.tellg();

  neb::bytes buf(size);

  ifs.seekg(0, ifs.beg);
  ifs.read((char *)buf.value(), buf.size());

  nbre::NBREIR ir;
  ir.set_name(name);
  ir.set_version(version);
  ir.set_height(height);
  ir.set_ir(neb::byte_to_string(buf));

  irs.push_back(ir);
}

int main(int argc, char *argv[]) {

  std::string ir_path = argv[1];
  std::string func_name = argv[2];

  int32_t n;
  std::cin >> n;

  while (n--) {
    std::vector<nbre::NBREIR> irs;
    randomly_gen_irs(ir_path, irs);

    auto &ele = irs.front();
    std::stringstream ss;
    ss << ele.name() << ele.version();

    auto &jd = neb::jit_driver::instance();
    jd.run<std::string>(ss.str(), irs, func_name, ss.str());

    if (!(n % 100)) {
      std::cout << n << std::endl;
    }
  }

  return 0;
}
