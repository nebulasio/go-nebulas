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

#include "common/byte.h"
#include "common/common.h"
#include <boost/program_options.hpp>
#include <iomanip>
#include <sstream>

namespace po = boost::program_options;

int main(int argc, char *argv[]) {

  po::options_description desc("Address base58 to string bytes");
  desc.add_options()("help", "show help message")(
      "address", po::value<std::string>(), "address base58 encoding");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);

  if (vm.count("help")) {
    std::cout << desc << "\n";
    return 1;
  }

  if (!vm.count("address")) {
    std::cout << "You must specify \"address\"!" << std::endl;
    return 1;
  }

  std::string addr_base58 = vm["address"].as<std::string>();
  neb::bytes addr_bytes = neb::bytes::from_base58(addr_base58);
  std::string addr = neb::byte_to_string(addr_bytes);
  LOG(INFO) << addr.size();

  std::stringstream ss;
  ss << '{';
  for (std::size_t i = 0; i < addr.size(); i++) {
    uint8_t c = addr[i];
    ss << "0x" << std::hex << std::setw(2) << std::setfill('0')
       << static_cast<int>(c) << ',';
  }
  ss.seekp(-1, std::ios_base::end);
  ss << '}';
  LOG(INFO) << ss.str();
  return 0;
}
