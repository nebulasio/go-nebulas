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
#include "runtime/auth/auth_handler.h"
#include "common/configuration.h"
#include "fs/ir_manager/api/ir_list.h"
#include "runtime/auth/auth_table.h"

namespace neb {
namespace rt {

namespace auth {

auth_handler::auth_handler(fs::ir_list *ir_list) : m_ir_list(ir_list) {}

bool auth_handler::is_ir_legitimate(const std::string ir_name,
                                    const address_t &from_addr,
                                    block_height_t height) {
  nbre::NBREIR cir;
  try {
    cir = m_ir_list->find_ir_at_height(
        configuration::instance().auth_module_name(), height);
  } catch (...) {
    return false;
  }
  version_t v = cir.version();
  if (m_auth_tables.find(v) == m_auth_tables.end()) {
    auth_table t = auth_table::generate_auth_table_from_ir(cir);
    std::unique_ptr<auth_table> table =
        std::unique_ptr<auth_table>(new auth_table(t));

    m_auth_tables.insert(std::make_pair(v, std::move(table)));
  }
  const std::unique_ptr<auth_table> &table = m_auth_tables[v];
  return table->is_ir_legitimate(ir_name, from_addr, height);
}
} // namespace auth
} // namespace rt
} // namespace neb
