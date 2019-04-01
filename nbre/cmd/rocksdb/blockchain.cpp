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
#include "fs/ir_manager/ir_manager.h"
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
      "db_path", po::value<std::string>(), "Database file directory")(
      "height", po::value<neb::block_height_t>(), "block height");

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
  if (!vm.count("height")) {
    std::cout << "You must specify \"height\"!" << std::endl;
    return 1;
  }

  std::string db_path = vm["db_path"].as<std::string>();
  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readonly);

  auto f_lib_height = [&]() {
    auto block_hash_bytes = rs.get("blockchain_lib");
    auto block_bytes = rs.get_bytes(block_hash_bytes);

    std::shared_ptr<corepb::Block> block = std::make_shared<corepb::Block>();
    bool ret = block->ParseFromArray(block_bytes.value(), block_bytes.size());
    LOG(INFO) << "blockchain lib height: " << block->height();
  };
  f_lib_height();

  neb::block_height_t height = vm["height"].as<neb::block_height_t>();
  auto f_block_hash = [&](neb::block_height_t height) {
    auto block_hash_bytes =
        rs.get_bytes(neb::number_to_byte<neb::bytes>(height));
    LOG(INFO) << neb::byte_to_number<neb::block_height_t>(block_hash_bytes);
  };
  f_block_hash(height);

  return 0;
}
