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

#include "fs/ir_manager/ir_manager_helper.h"
#include "common/configuration.h"
#include "jit/jit_driver.h"
#include <boost/foreach.hpp>
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>

namespace neb {
namespace fs {

void ir_manager_helper::set_failed_flag(rocksdb_storage *rs,
                                        const std::string &flag) {
  rs->put(flag, neb::util::string_to_byte(flag));
}

bool ir_manager_helper::has_failed_flag(rocksdb_storage *rs,
                                        const std::string &flag) {
  try {
    rs->get(flag);
  } catch (const std::exception &e) {
    // LOG(INFO) << "nbre failed flag not found " << e.what();
    return false;
  }
  return true;
}

void ir_manager_helper::del_failed_flag(rocksdb_storage *rs,
                                        const std::string &flag) {
  rs->del(flag);
}

block_height_t ir_manager_helper::nbre_block_height(rocksdb_storage *rs) {

  block_height_t start_height = 0;
  try {
    start_height = neb::util::byte_to_number<block_height_t>(rs->get(
        std::string(neb::configuration::instance().nbre_max_height_name(),
                    std::allocator<char>())));
  } catch (std::exception &e) {
    LOG(INFO) << "to init nbre max height " << e.what();
    rs->put(std::string(neb::configuration::instance().nbre_max_height_name(),
                        std::allocator<char>()),
            neb::util::number_to_byte<neb::util::bytes>(start_height));
  }
  return start_height;
}

block_height_t ir_manager_helper::lib_block_height(blockchain *bc) {

  std::unique_ptr<corepb::Block> end_block = bc->load_LIB_block();
  block_height_t end_height = end_block->height();
  return end_height;
}

void ir_manager_helper::run_auth_table(
    nbre::NBREIR &nbre_ir, std::map<auth_key_t, auth_val_t> &auth_table) {

  auth_table_t rows;

  try {
    std::stringstream ss;
    ss << nbre_ir.name() << nbre_ir.version();
    std::vector<nbre::NBREIR> irs;
    irs.push_back(nbre_ir);

    jit_driver &jd = jit_driver::instance();
    rows = jd.run<auth_table_t>(
        ss.str(), irs,
        neb::configuration::instance().auth_func_mangling_name());

  } catch (const std::exception &e) {
    LOG(INFO) << "execute auth table failed " << e.what();
    return;
  }

  auth_table.clear();
  for (auto &r : rows) {
    assert(std::tuple_size<std::remove_reference<decltype(r)>::type>::value ==
           5);
    auth_key_t k =
        std::make_tuple(std::get<0>(r), std::get<1>(r), std::get<2>(r));
    auth_val_t v = std::make_tuple(std::get<3>(r), std::get<4>(r));
    auth_table.insert(std::make_pair(k, v));
  }
  return;
}

void ir_manager_helper::load_auth_table(
    rocksdb_storage *rs, std::map<auth_key_t, auth_val_t> &auth_table) {

  // auth table exists in memory
  if (!auth_table.empty()) {
    return;
  }

  std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
  neb::util::bytes payload_bytes;
  try {
    payload_bytes =
        rs->get(neb::configuration::instance().nbre_auth_table_name());
  } catch (const std::exception &e) {
    LOG(INFO) << "auth table not deploy yet " << e.what();
    return;
  }

  bool ret =
      nbre_ir->ParseFromArray(payload_bytes.value(), payload_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse payload auth table failed");
  }

  run_auth_table(*nbre_ir.get(), auth_table);
}

void ir_manager_helper::update_ir_list(const std::string &ir_list,
                                       const std::string &ir_name,
                                       rocksdb_storage *rs) {

  auto update_ir_list_to_storage =
      [&rs](const std::string &key, const boost::property_tree::ptree &val_pt) {
        std::stringstream ss;
        boost::property_tree::json_parser::write_json(ss, val_pt, false);
        rs->put(key, neb::util::string_to_byte(ss.str()));
      };

  std::string const_str_nbre_ir_list =
      neb::configuration::instance().nbre_ir_list_name();
  std::string const_str_ir_list = neb::configuration::instance().ir_list_name();

  if (ir_list.empty()) {
    boost::property_tree::ptree ele, arr, root;
    ele.put("", ir_name);
    arr.push_back(std::make_pair("", ele));
    root.add_child(const_str_ir_list, arr);
    update_ir_list_to_storage(const_str_nbre_ir_list, root);
    return;
  }

  boost::property_tree::ptree root;
  std::stringstream ss(ir_list);
  boost::property_tree::json_parser::read_json(ss, root);

  BOOST_FOREACH (boost::property_tree::ptree::value_type &v,
                 root.get_child(const_str_ir_list)) {
    boost::property_tree::ptree pt = v.second;
    if (ir_name == pt.get<std::string>(std::string())) {
      // ir name already exists
      return;
    }
  }

  boost::property_tree::ptree &arr = root.get_child(const_str_ir_list);
  boost::property_tree::ptree ele;
  ele.put("", ir_name);
  arr.push_back(std::make_pair("", ele));

  update_ir_list_to_storage(const_str_nbre_ir_list, root);
}

} // namespace fs
} // namespace neb
