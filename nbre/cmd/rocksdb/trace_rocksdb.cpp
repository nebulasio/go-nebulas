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
#include "common/util/version.h"
#include "fs/ir_manager/ir_manager.h"
#include "fs/proto/block.pb.h"
#include "fs/proto/ir.pb.h"
#include "fs/rocksdb_storage.h"
#include "fs/util.h"
#include <boost/foreach.hpp>
#include <boost/format.hpp>
#include <boost/program_options.hpp>
#include <boost/property_tree/json_parser.hpp>

namespace po = boost::program_options;

void display_ir_versions(neb::fs::rocksdb_storage &rs) {

  neb::util::bytes bytes_ir_name_list =
      rs.get(neb::configuration::instance().nbre_ir_list_name());
  std::string json_ir_name_list = neb::util::byte_to_string(bytes_ir_name_list);

  std::string ir_list_name = neb::configuration::instance().ir_list_name();

  boost::property_tree::ptree root;
  std::stringstream ss(json_ir_name_list);
  boost::property_tree::json_parser::read_json(ss, root);

  BOOST_FOREACH (boost::property_tree::ptree::value_type &v,
                 root.get_child(ir_list_name)) {
    boost::property_tree::ptree pt = v.second;
    std::string module_name = pt.get<std::string>(std::string());

    auto bytes_versions = rs.get(module_name);
    size_t gap = sizeof(uint64_t) / sizeof(uint8_t);

    std::cout << "for module " << module_name << " versions:\n";
    for (size_t i = 0; i < bytes_versions.size(); i += gap) {
      neb::byte_t *bytes_version =
          bytes_versions.value() + (bytes_versions.size() - gap - i);

      if (bytes_version != nullptr) {
        uint64_t version =
            neb::util::byte_to_number<uint64_t>(bytes_version, gap);
        std::cout << version << ' ';
      }
    }
    std::cout << std::endl;
  }
}

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
  rs.open_database(db_path, neb::fs::storage_open_for_readonly);

  auto f_keys = [](rocksdb::Iterator *it) {
    for (it->SeekToFirst(); it->Valid(); it->Next()) {
      LOG(INFO) << it->key().ToString();
    }
  };
  rs.display(f_keys);

  display_ir_versions(rs);

  rs.close_database();
  return 0;
}
