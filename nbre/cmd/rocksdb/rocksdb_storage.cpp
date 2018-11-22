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
#include "common/util/byte.h"
#include "fs/manager/nbre_storage.h"
#include "fs/proto/block.pb.h"
#include "fs/util.h"

#include <boost/foreach.hpp>
#include <boost/program_options.hpp>
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>
#include <sstream>

namespace po = boost::program_options;

int main(int argc, char *argv[]) {

  po::options_description desc("Rocksdb read account balance");
  desc.add_options()("help", "show help message")(
      "db_path", po::value<std::string>(), "Database file directory")(
      "max_height", po::value<neb::block_height_t>(), "nbre max height");

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
  if (!vm.count("max_height")) {
    std::cout << "You must specify \"max_height\"!" << std::endl;
    return 1;
  }

  std::string db_path = vm["db_path"].as<std::string>();
  neb::fs::rocksdb_storage rs;
  rs.open_database(db_path, neb::fs::storage_open_for_readwrite);

  neb::block_height_t max_height = vm["max_height"].as<neb::block_height_t>();
  auto f_set_nbre_max_height = [&]() {
    rs.put("nbre_max_height",
           neb::util::number_to_byte<neb::util::bytes>(max_height));
  };
  f_set_nbre_max_height();

  auto f_set_ir_list = [&rs]() {
    auto f_build_json = []() -> std::string {
      boost::property_tree::ptree pt;
      boost::property_tree::ptree children;
      boost::property_tree::ptree child1, child2;

      child1.put("", "nr");
      child2.put("", "dip");
      children.push_back(std::make_pair("", child1));
      children.push_back(std::make_pair("", child2));

      pt.add_child("ir_list", children);

      std::stringstream ss;
      boost::property_tree::json_parser::write_json(ss, pt);
      return ss.str();
    };

    auto json_str = f_build_json();
    rs.put(neb::configuration::instance().nbre_ir_list_name(),
           neb::util::string_to_byte(json_str));
  };
  // f_set_ir_list();

  rs.close_database();
  return 0;
}
