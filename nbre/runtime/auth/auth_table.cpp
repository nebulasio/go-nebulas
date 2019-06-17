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
#include "runtime/auth/auth_table.h"
#include "common/byte.h"
#include "common/configuration.h"
#include "runtime/param_trait.h"

namespace neb {
namespace rt {
namespace auth {
bool auth_table::is_ir_legitimate(const std::string &ir_name,
                                  const address_t &from_addr,
                                  block_height_t h) const {
  auto key = auth_key(ir_name, from_addr);
  auto it = m_auth_table.find(key);
  if (it == m_auth_table.end()) {
    return false;
  }
  auth_val_t v = it->second;
  block_height_t start = std::get<2>(v);
  block_height_t end = std::get<3>(v);
  if (h >= start && h < end) {
    return true;
  }
  return false;
}

std::string auth_table::auth_key(const std::string &ir_name,
                                 const address_t &from_addr) {
  return ir_name + std::to_string(from_addr);
}

auth_table
auth_table::generate_auth_table_from_ir(const nbre::NBREIR &compiled_ir) {
  auth_table ret;
  ret.m_version = ::neb::version(compiled_ir.version());
  ret.m_available_height = compiled_ir.height();

  std::vector<auth_items_t> items =
      param_trait::get_param<std::vector<auth_items_t>>(
          compiled_ir, configuration::instance().auth_func_name());

  for (auto &item : items) {
    ::neb::address_t addr = to_address(std::get<1>(item));
    auth_val_t av = std::make_tuple(std::get<0>(item), addr, std::get<2>(item),
                                    std::get<3>(item));
    auto key = auth_key(std::get<0>(av), std::get<1>(av));
    ret.m_auth_table.insert(std::make_pair(key, av));
  }
  return ret;
}
} // namespace auth
} // namespace rt
} // namespace neb
