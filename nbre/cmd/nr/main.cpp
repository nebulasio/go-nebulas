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

#include "common/version.h"
#include "runtime/nr/impl/nebulas_rank.h"
#include "runtime/nr/impl/nr_impl.h"
#include <boost/program_options.hpp>

namespace po = boost::program_options;

int main(int argc, char *argv[]) {

  int64_t a = 3118;
  int64_t b = 3792;
  int64_t c = 6034;
  int64_t d = 4158;
  neb::rt::nr::nr_float_t theta = 2.2;
  neb::rt::nr::nr_float_t mu = 0.1;
  neb::rt::nr::nr_float_t lambda = 0.3;

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
  auto nr_ret = neb::rt::nr::entry_point_nr_impl(start_block, end_block,
                                                 neb::version(0, 1, 0).data(),
                                                 a, b, c, d, theta, mu, lambda);
  // std::cout << ret << std::endl;
  // std::cout << std::endl;

  if (std::get<0>(nr_ret)) {
    auto ret = neb::rt::nr::nebulas_rank::nr_info_to_json(nr_ret);
    std::cout << *ret << std::endl;
  }
  return 0;
}
