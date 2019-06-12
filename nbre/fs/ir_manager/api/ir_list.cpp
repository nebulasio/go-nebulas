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
#include "fs/ir_manager/api/ir_list.h"
#include "fs/ir_manager/api/auth_ir_list.h"
#include "fs/ir_manager/api/dip_ir_list.h"
#include "fs/ir_manager/api/nr_ir_list.h"

namespace neb {
namespace fs {
ir_list::ir_list(rocksdb_storage *storage) : m_storage(storage) {
  m_ir_item_list.insert("nr", std::make_shared<nr_ir_list>(storage));
  m_ir_item_list.insert("auth", std::make_shared<auth_ir_list>(storage));
  m_ir_item_list.insert("dip", std::make_shared<dip_ir_list>(storage));
}

void ir_list::write_ir(const nbre::NBREIR &raw_ir,
                       const nbre::NBREIR &compiled_ir) {
  std::string name = raw_ir.name();
  if (m_ir_item_list.find(name) == m_ir_item_list.end()) {
    LOG(ERROR) << "no ir list for " << name;
    throw std::runtime_error("no ir list");
  }
  m_ir_item_list[name]->write_ir(raw_ir, compiled_ir);
}

nbre::NBREIR ir_list::get_raw_ir(const std::string &ir_name, version_t v) {
  if (m_ir_item_list.find(ir_name) == m_ir_item_list.end()) {
    LOG(ERROR) << "no ir list for " << ir_name;
    throw std::runtime_error("no ir list");
  }
  return m_ir_item_list[ir_name]->get_raw_ir(v);
}
nbre::NBREIR ir_list::get_ir(const std::string &ir_name, version_t v) {
  if (m_ir_item_list.find(ir_name) == m_ir_item_list.end()) {
    LOG(ERROR) << "no ir list for " << ir_name;
    throw std::runtime_error("no ir list");
  }
  return m_ir_item_list[ir_name]->get_ir(v);
}
nbre::NBREIR ir_list::find_ir_at_height(const std::string &ir_name,
                                        block_height_t height) {
  if (m_ir_item_list.find(ir_name) == m_ir_item_list.end()) {
    LOG(ERROR) << "no ir list for " << ir_name;
    throw std::runtime_error("no ir list");
  }
  return m_ir_item_list[ir_name]->find_ir_at_height(height);
}
}
} // namespace neb
