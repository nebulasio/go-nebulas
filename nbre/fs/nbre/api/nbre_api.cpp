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

#include "fs/nbre/api/nbre_api.h"
#include "common/configuration.h"
#include <boost/foreach.hpp>
#include <boost/property_tree/json_parser.hpp>

namespace neb {
namespace fs {

nbre_api::nbre_api(const std::string &db_path,
                   enum storage_open_flag open_flag) {
  m_storage = std::make_unique<rocksdb_storage>();
  m_storage->open_database(db_path, open_flag);
}

nbre_api::~nbre_api() {
  if (m_storage) {
    m_storage->close_database();
  }
}

std::shared_ptr<std::vector<std::string>> nbre_api::get_irs() {

  neb::util::bytes bytes_ir_name_list_json =
      m_storage->get(neb::configuration::instance().nbre_ir_list_name());
  std::string ir_name_list_json =
      neb::util::byte_to_string(bytes_ir_name_list_json);

  boost::property_tree::ptree root;
  std::stringstream ss(ir_name_list_json);
  boost::property_tree::json_parser::read_json(ss, root);

  std::vector<std::string> v;

  BOOST_FOREACH (
      boost::property_tree::ptree::value_type &ir_name,
      root.get_child(neb::configuration::instance().ir_list_name())) {
    boost::property_tree::ptree pt = ir_name.second;
    v.push_back(pt.get<std::string>(std::string()));
  }
  return std::make_shared<std::vector<std::string>>(v);
}

std::shared_ptr<std::vector<version_t>>
nbre_api::get_ir_versions(const std::string &ir_name) {

  std::vector<version_t> v;
  try {
    neb::util::bytes bytes_versions = m_storage->get(ir_name);
    size_t gap = sizeof(uint64_t) / sizeof(uint8_t);

    for (size_t i = 0; i < bytes_versions.size(); i += gap) {
      byte_t *bytes_version =
          bytes_versions.value() + (bytes_versions.size() - gap - i);
      if (bytes_version != nullptr) {
        uint64_t version =
            neb::util::byte_to_number<uint64_t>(bytes_version, gap);
        v.push_back(version);
      }
    }
  } catch (const std::exception &e) {
    LOG(INFO) << e.what();
  }
  return std::make_shared<std::vector<version_t>>(v);
}
} // namespace fs
} // namespace neb
