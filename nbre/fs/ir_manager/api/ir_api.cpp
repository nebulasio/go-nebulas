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

#include "fs/ir_manager/api/ir_api.h"
#include "common/configuration.h"
#include "common/util/version.h"
#include <boost/foreach.hpp>
#include <boost/property_tree/json_parser.hpp>

namespace neb {
namespace fs {

std::unique_ptr<std::vector<std::string>>
ir_api::get_ir_list(rocksdb_storage *rs) {
  auto ret = std::make_unique<std::vector<std::string>>();

  std::string ir_list = neb::configuration::instance().ir_list_name();
  neb::util::bytes ir_list_bytes;
  try {
    ir_list_bytes = rs->get(ir_list);
  } catch (const std::exception &e) {
    LOG(INFO) << "ir list empty, get ir list failed " << e.what();
    return ret;
  }

  boost::property_tree::ptree root;
  std::stringstream ss(neb::util::byte_to_string(ir_list_bytes));
  boost::property_tree::json_parser::read_json(ss, root);

  BOOST_FOREACH (boost::property_tree::ptree::value_type &name,
                 root.get_child(ir_list)) {
    boost::property_tree::ptree pt = name.second;
    ret->push_back(pt.get<std::string>(std::string()));
  }
  return ret;
}

std::unique_ptr<std::vector<version_t>>
ir_api::get_ir_versions(const std::string &name, rocksdb_storage *rs) {
  auto ret = std::make_unique<std::vector<version_t>>();

  neb::util::bytes ir_versions_bytes;
  try {
    ir_versions_bytes = rs->get(name);
  } catch (const std::exception &e) {
    LOG(INFO) << "ir with name " << name << " versions empty, get " << name
              << " versions failed " << e.what();
    return ret;
  }

  boost::property_tree::ptree root;
  std::stringstream ss(neb::util::byte_to_string(ir_versions_bytes));
  boost::property_tree::json_parser::read_json(ss, root);

  BOOST_FOREACH (boost::property_tree::ptree::value_type &version,
                 root.get_child(name)) {
    boost::property_tree::ptree pt = version.second;
    ret->push_back(pt.get<version_t>(std::string()));
  }

  sort(ret->begin(), ret->end(), [](const version_t &v1, const version_t &v2) {
    neb::util::version obj_v1(v1);
    neb::util::version obj_v2(v2);
    return obj_v1 > obj_v2;
  });
  return ret;
}

bool ir_api::ir_version_exist(const std::string &name, version_t version,
                              rocksdb_storage *rs) {
  auto versions = get_ir_versions(name, rs);
  return std::find(versions->begin(), versions->end(), version) !=
         versions->end();
}
} // namespace fs
} // namespace neb
