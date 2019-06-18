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
#include "util/npr.h"
#include "common/configuration.h"
#include "fs/proto/block.pb.h"
#include "util/json_parser.h"
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>

namespace neb {
namespace util {
bool is_npr_tx(const corepb::Transaction &tx) {
  std::string dst_type = neb::configuration::instance().ir_tx_payload_type();

  std::string type = tx.data().type();
  return type == dst_type;
}

nbre::NBREIR extract_npr(const corepb::Transaction &tx) {
  nbre::NBREIR npr;

  auto &data = tx.data();
  boost::property_tree::ptree pt;
  neb::util::json_parser::read_json(data.payload(), pt);
  neb::bytes payload_bytes =
      neb::bytes::from_base64(pt.get<std::string>("Data"));
  bool status = npr.ParseFromArray(payload_bytes.value(), payload_bytes.size());
  if (!status) {
    LOG(ERROR) << "parse transaction payload failed "
               << pt.get<std::string>("Data");
    throw std::runtime_error("extract_npr: parse transaction payload failed");
  }
  return npr;
}

bool is_auth_npr(const address_t &from, const std::string &module_name) {
  if (neb::configuration::instance().auth_module_name() == module_name &&
      neb::configuration::instance().admin_pub_addr() == from) {
    return true;
  }
  return false;
}

} // namespace util
} // namespace neb
