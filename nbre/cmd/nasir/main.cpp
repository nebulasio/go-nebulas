// Copyright (C) 2018 go-nebulas authors
//
//
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
#include "common/util/byte.h"
#include "fs/proto/ir.pb.h"
#include <boost/format.hpp>
#include <boost/program_options.hpp>
#include <fstream>
#include <iostream>

namespace po = boost::program_options;

int main(int argc, char *argv[]) {
  po::options_description desc("Generate IR Payload");
  desc.add_options()("help", "show help message")(
      "input", po::value<std::string>(), "IR configuration file")(
      "output", po::value<std::string>(), "output file");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);
  if (vm.count("help")) {
    std::cout << desc << "\n";
    return 1;
  }

  if (!vm.count("input")) {
    std::cout << "You must specify \"input\"!" << std::endl;
    return 1;
  }
  std::ifstream ifs;

  try {
    std::string ir_fp = vm["input"].as<std::string>();
    neb::ir_conf_reader reader(ir_fp);
    ifs.open(reader.ir_fp().c_str(), std::ios::in | std::ios::binary);
    if (!ifs.is_open()) {
      throw std::invalid_argument(
          boost::str(boost::format("can't open file %1%") % reader.ir_fp()));
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

    nbre::NBREIR ir_info;
    ir_info.set_name(reader.self_ref().name());
    ir_info.set_version(reader.self_ref().version().data());
    ir_info.set_height(reader.available_height());
    for (size_t i = 0; i < reader.depends().size(); ++i) {
      nbre::NBREIRDepend *d = ir_info.add_depends();
      d->set_name(reader.depends()[i].name());
      d->set_version(reader.depends()[i].version().data());
    }
    ir_info.set_ir(neb::util::byte_to_string(buf));

    auto bytes_long = ir_info.ByteSizeLong();
    if (bytes_long > 128 * 1024) {
      throw std::invalid_argument("bytes too long !");
    }

    std::ofstream ofs;
    ofs.open(vm["output"].as<std::string>(),
             std::ios::out | std::ios::binary | std::ios::trunc);
    if (!ofs.is_open()) {
      throw std::invalid_argument("can't open output file");
    }
    neb::util::bytes out_bytes(bytes_long);
    ir_info.SerializeToArray((void *)out_bytes.value(), out_bytes.size());

    ofs.write((const char *)out_bytes.value(), out_bytes.size());
    ofs.close();

  } catch (std::exception &e) {
    ifs.close();
    std::cout << e.what() << std::endl;
  }

  return 0;
}
