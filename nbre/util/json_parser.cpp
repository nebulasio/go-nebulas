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

#include "util/json_parser.h"
#include <boost/property_tree/json_parser.hpp>

namespace neb {
namespace util {
void json_parser::read_json(const std::string &json_str,
                            boost::property_tree::ptree &pt) {
  std::stringstream ss(json_str);
  boost::property_tree::json_parser::read_json(ss, pt);
  return;
}

void json_parser::write_json(std::string &json_str,
                             const boost::property_tree::ptree &pt) {
  std::stringstream ss;
  boost::property_tree::json_parser::write_json(ss, pt, false);
  json_str = ss.str();
  return;
}
} // namespace util
} // namespace neb
