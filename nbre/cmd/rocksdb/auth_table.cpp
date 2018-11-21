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
#include "fs/manager/nbre_storage.h"
#include "fs/proto/block.pb.h"
#include "fs/proto/ir.pb.h"
#include "fs/rocksdb_storage.h"
#include "fs/util.h"
#include <boost/format.hpp>
#include <boost/program_options.hpp>

namespace po = boost::program_options;

int main(int argc, char *argv[]) {

  po::options_description desc("Rocksdb read and write");
  desc.add_options()("help", "show help message")(
      "db_path", po::value<std::string>(), "Database file directory");

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

  std::string db_path = vm["db_path"].as<std::string>();
  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readwrite);

  auto f_del_nbre_auth_table = [&]() { rs.del("nbre_auth_table"); };
  f_del_nbre_auth_table();

  auto f_set_nbre_auth_table = [&]() {
    std::string addr1_base58 = neb::util::string_to_byte("addr1").to_base58();
    std::string addr2_base58 = neb::util::string_to_byte("addr2").to_base58();

    rs.put("nbre_auth_table",
           neb::util::string_to_byte(boost::str(
               boost::format(
                   "nr,1,%1%,100,200\nnr,2,%2%,150,250\ndip,1,%1%,200,300\n") %
               addr1_base58 % addr2_base58)));
  };
  // f_set_nbre_auth_table();

  return 0;
}
