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
#include "core/net_ipc/nipc_pkg.h"
#include <boost/algorithm/string/replace.hpp>
#include <boost/foreach.hpp>
#include <boost/property_tree/json_parser.hpp>

namespace neb {
namespace core {

std::string pkg_type_id_to_name(uint64_t type) {
  std::string ret;
  switch (type) {
  case heart_beat_pkg:
    ret = "heart_beat_pkg";
    break;
  case nipc_last_pkg_id:
    ret = "nipc_last_pkg_id";
    break;

#define define_ipc_param(type, name)

#define define_ipc_pkg(type, ...)                                              \
  case JOIN(type, _pkg):                                                       \
    ret = #type;                                                               \
    break;

#define define_ipc_api(req, ack)

#include "core/net_ipc/ipc_interface_impl.h"

#undef define_ipc_api
#undef define_ipc_pkg
#undef define_ipc_param
  default:
    ret = "not_def_in_nipc";
  }

  return ret;
}

bool is_pkg_type_has_callback(uint64_t type) {

#define define_ipc_param(type, name)
#define define_ipc_pkg(type, ...)
#define define_ipc_api(req, ack)                                               \
  if (type == JOIN(req, _pkg))                                                 \
    return true;

#include "core/net_ipc/ipc_interface_impl.h"

#undef define_ipc_api
#undef define_ipc_pkg
#undef define_ipc_param
  return false;
}
std::string convert_nr_result_to_json(const nr_result &nr) {
  boost::property_tree::ptree root;
  boost::property_tree::ptree arr;

  root.put("start_height", std::to_string(nr.get<p_start_block>()));
  root.put("end_height", std::to_string(nr.get<p_end_block>()));
  root.put("version", std::to_string(nr.get<p_nr_version>()));
  std::vector<nr_item> nrs = nr.get<p_nr_items>();
  if (nrs.empty()) {
    boost::property_tree::ptree p;
    arr.push_back(std::make_pair(std::string(), p));
  }
  for (auto &it : nrs) {
    boost::property_tree::ptree p;
    auto addr = to_address(it.get<p_nr_item_addr>());
    p.put(std::string("address"), address_to_base58(addr));
    p.put(std::string("in_outs"), std::to_string(it.get<p_nr_item_in_outs>()));
    p.put(std::string("median"), std::to_string(it.get<p_nr_item_median>()));
    p.put(std::string("weight"), std::to_string(it.get<p_nr_item_weight>()));
    p.put(std::string("score"), std::to_string(it.get<p_nr_item_score>()));
    arr.push_back(std::make_pair(std::string(), p));
  }
  root.add_child("nrs", arr);
  std::stringstream ss;
  boost::property_tree::json_parser::write_json(ss, root, false);
  std::string tmp_ss;
  boost::replace_all(tmp_ss, "[\"\"]", "[]");
  return tmp_ss;
}

std::string result_status_to_string(uint32_t status) {
  std::string result;
  switch (status) {
  case result_status::succ:
    result = "success";
    break;
  case result_status::is_running:
    result = "the same computation is already running, ignore this one";
    break;
  case result_status::no_cached:
    result = "there is no cached data for this on disk or in mem!";
    break;
  default:
    result = "unknown status";
    break;
  }
  return result;
}
} // namespace core
} // namespace neb
