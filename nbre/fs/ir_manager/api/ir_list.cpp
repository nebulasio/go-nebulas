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
ir_list::ir_list(class storage *s) : m_storage(s) {
  m_ir_item_list.insert(std::make_pair("nr", std::make_shared<nr_ir_list>(s)));
  m_ir_item_list.insert(
      std::make_pair("auth", std::make_shared<auth_ir_list>(s)));
  m_ir_item_list.insert(
      std::make_pair("dip", std::make_shared<dip_ir_list>(s)));
}

bool ir_list::ir_exist(const std::string &name, uint64_t v) {
  if (m_ir_item_list.find(name) == m_ir_item_list.end()) {
    LOG(ERROR) << "no ir list for " << name;
    return false;
  }
  return m_ir_item_list[name]->ir_exist(v);
}
void ir_list::write_ir(const nbre::NBREIR &raw_ir,
                       const nbre::NBREIR &compiled_ir) {
  std::string name = raw_ir.name();
  if (m_ir_item_list.find(name) == m_ir_item_list.end()) {
    LOG(ERROR) << "no ir list for " << name;
    throw std::runtime_error("ir_list::write_ir: no ir list");
  }
  m_ir_item_list[name]->write_ir(raw_ir, compiled_ir);
}

nbre::NBREIR ir_list::get_raw_ir(const std::string &ir_name, version_t v) {
  if (m_ir_item_list.find(ir_name) == m_ir_item_list.end()) {
    LOG(ERROR) << "no ir list for " << ir_name;
    throw std::runtime_error("ir_list::get_raw_ir: no ir list");
  }
  return m_ir_item_list[ir_name]->get_raw_ir(v);
}
nbre::NBREIR ir_list::get_ir(const std::string &ir_name, version_t v) {
  if (m_ir_item_list.find(ir_name) == m_ir_item_list.end()) {
    LOG(ERROR) << "no ir list for " << ir_name;
    throw std::runtime_error("ir_list::get_ir: no ir list");
  }
  return m_ir_item_list[ir_name]->get_ir(v);
}
nbre::NBREIR ir_list::find_ir_at_height(const std::string &ir_name,
                                        block_height_t height) {
  if (m_ir_item_list.find(ir_name) == m_ir_item_list.end()) {
    LOG(ERROR) << "no ir list for " << ir_name;
    throw std::runtime_error("ir_list::find_ir_at_height: no ir list");
  }
  return m_ir_item_list[ir_name]->find_ir_at_height(height);
}

bytes ir_list::get_ir_brief_key_with_height(const std::string &ir_name,
                                            block_height_t height) {
  if (m_ir_item_list.find(ir_name) == m_ir_item_list.end()) {
    LOG(ERROR) << "no ir list for " << ir_name;
    throw std::runtime_error(
        "ir_list::get_ir_brief_key_with_height: no ir list");
  }
  return m_ir_item_list[ir_name]->get_ir_brief_key_with_height(height);
}
std::vector<std::string> ir_list::get_ir_names() const {
  std::vector<std::string> ret;
  for (auto &p : m_ir_item_list) {
    auto t = p.second;
    auto kt = t->get_ir_names();
    ret.insert(ret.end(), kt.begin(), kt.end());
  }
  return ret;
}

std::vector<version_t>
ir_list::get_ir_versions(const std::string &ir_name) const {
  std::vector<version_t> ret;
  auto it = m_ir_item_list.find(ir_name);
  if (it == m_ir_item_list.end()) {
    return ret;
  }
  return (it->second)->get_ir_versions();
}

} // namespace fs
} // namespace neb
