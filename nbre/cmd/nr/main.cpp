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

#include "runtime/nr/impl/nebulas_rank.h"
#include "runtime/nr/impl/nr_impl.h"
#include <boost/program_options.hpp>

namespace po = boost::program_options;

int main(int argc, char *argv[]) {

  neb::rt::nr::nr_float_t a = 2000.0;
  neb::rt::nr::nr_float_t b = 200000.0;
  neb::rt::nr::nr_float_t c = 100.0;
  neb::rt::nr::nr_float_t d = 1000.0;
  int64_t mu = 1;
  int64_t lambda = 3;

  po::options_description desc("Nr");
  desc.add_options()("help", "show help message")(
      "start_block", po::value<uint64_t>(), "Start block height")(
      "end_block", po::value<uint64_t>(), "End block height");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);

  if (vm.count("help")) {
    std::cout << desc << "\n";
    return 1;
  }

  if (!vm.count("start_block")) {
    std::cout << "You must specify \"start_block\"!" << std::endl;
    return 1;
  }
  if (!vm.count("end_block")) {
    std::cout << "You must specify \"end_block\"!" << std::endl;
    return 1;
  }

  uint64_t start_block = vm["start_block"].as<uint64_t>();
  uint64_t end_block = vm["end_block"].as<uint64_t>();
  auto ret = neb::rt::nr::entry_point_nr_impl(start_block, end_block, a, b, c,
                                              d, mu, lambda);
  std::cout << ret << std::endl;
  std::cout << std::endl;

  auto nr_infos = neb::rt::nr::nebulas_rank::json_to_nr_info(ret);
  ret = neb::rt::nr::nebulas_rank::nr_info_to_json(*nr_infos);
  std::cout << ret << std::endl;
  return 0;
}
