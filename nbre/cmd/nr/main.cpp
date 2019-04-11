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

#include "common/util/version.h"
#include "runtime/nr/impl/nebulas_rank.h"
#include "runtime/nr/impl/nr_impl.h"
#include "runtime/nr/api/nr_api.h"
#include <boost/program_options.hpp>

namespace po = boost::program_options;

int main(int argc, char *argv[]) {

  auto to_version_t = [](uint32_t major_version, uint16_t minor_version,
                         uint16_t patch_version) -> neb::rt::nr::version_t {
    return (0ULL + major_version) + ((0ULL + minor_version) << 32) +
           ((0ULL + patch_version) << 48);
  };

  neb::compatible_uint64_t a = 100;
  neb::compatible_uint64_t b = 2;
  neb::compatible_uint64_t c = 6;
  neb::compatible_uint64_t d = -9;
  neb::rt::nr::nr_float_t theta = 1;
  neb::rt::nr::nr_float_t mu = 1;
  neb::rt::nr::nr_float_t lambda = 2;

  po::options_description desc("Nr");
  desc.add_options()("help", "show help message")(
      "db_path", po::value<std::string>(), "Database file directory")(
      "start_block", po::value<uint64_t>(), "Start block height")(
      "end_block", po::value<uint64_t>(), "End block height");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);

  if (vm.count("help")) {
    std::cout << desc << "\n";
    return 1;
  }

  if (!vm.count("db_path")) {
    std::cout << "You must specify \"db_path\"!" << std::endl;
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

  std::string db_path = vm["db_path"].as<std::string>();
  neb::compatible_uint64_t start_block = vm["start_block"].as<uint64_t>();
  neb::compatible_uint64_t end_block = vm["end_block"].as<uint64_t>();
  auto ret = neb::rt::nr::nr_api(
      db_path,
      start_block, end_block, to_version_t(0, 0, 1), a, b, c, d,
      theta, mu, lambda);
  std::cout << ret << std::endl;
  std::cout << std::endl;

  //auto nr_infos = neb::rt::nr::nebulas_rank::json_to_nr_info(ret);
  //ret = neb::rt::nr::nebulas_rank::nr_info_to_json(*nr_infos);
  //std::cout << ret << std::endl;
  return 0;
}
