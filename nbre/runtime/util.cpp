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

#include "runtime/util.h"
#include "util/json_parser.h"

namespace neb {
namespace rt {

std::string meta_info_to_json(
    const std::vector<std::pair<std::string, std::string>> &meta_info) {
  boost::property_tree::ptree pt;
  for (auto &ele : meta_info) {
    pt.put(ele.first, ele.second);
  }
  std::string ret;
  neb::util::json_parser::write_json(ret, pt);
  return ret;
}

std::vector<std::pair<std::string, std::string>>
json_to_meta_info(const std::string &json) {
  boost::property_tree::ptree pt;
  neb::util::json_parser::read_json(json, pt);

  std::vector<std::pair<std::string, std::string>> meta_info;
  std::string start_height = pt.get<std::string>("start_height");
  std::string end_height = pt.get<std::string>("end_height");
  std::string version = pt.get<std::string>("version");

  meta_info.push_back(std::make_pair("start_height", start_height));
  meta_info.push_back(std::make_pair("end_height", end_height));
  meta_info.push_back(std::make_pair("version", version));
  return meta_info;
}

} // namespace rt
} // namespace neb
