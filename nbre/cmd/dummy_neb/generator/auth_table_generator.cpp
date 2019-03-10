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
#include "cmd/dummy_neb/generator/auth_table_generator.h"
#include "fs/proto/ir.pb.h"
#include <sstream>

static std::string ir_slice1 =
    "#include <string>  \n"
    "#include <tuple> \n"
    "#include <vector> \n"
    "typedef std::string name_t; \n"
    "typedef uint64_t version_t; \n"
    "typedef std::string address_t; \n"
    "typedef uint64_t height_t; \n"
    "typedef std::tuple<name_t, address_t, height_t, height_t> "
    "row_t; \n"
    "std::vector<row_t> entry_point_auth() { \n"
    " auto to_version_t = [](uint32_t major_version, uint16_t minor_version, "
    "\n"
    "                         uint16_t patch_version) -> version_t { \n"
    "    return (0ULL + major_version) + ((0ULL + minor_version) << 32) + \n"
    "           ((0ULL + patch_version) << 48); \n"
    "  };";

static std::string ir_slice_2 =
    "std::vector<row_t> auth_table = {NR_ITEM, \n DIP_ITEM}; \n";
static std::string ir_slice_nr =
    "std::make_tuple(\"nr\", std::string(VAR.begin(), VAR.end()), 1ULL, "
    "0xFFFFFFFFFFFFFFFFULL)";
static std::string ir_slice_dip =
    "std::make_tuple(\"dip\", std::string(VAR.begin(), VAR.end()), 1ULL, "
    "0xFFFFFFFFFFFFFFFFULL)";
static std::string ir_slice_end = "return auth_table;}";

static std::string gen_admin_var(const address_t &addr,
                                 const std::string &var_name) {
  std::stringstream ss;
  ss << "\n\tauto " << var_name << " = {";
  for (size_t i = 0; i < addr.size(); ++i) {
    ss << "0x" << std::hex << static_cast<uint32_t>(addr[i]);
    if (i != addr.size() - 1) {
      ss << ", ";
    }
  }
  ss << "};\n";
  return ss.str();
}
std::string gen_auth_table_ir(const address_t &nr_admin,
                              const address_t &dip_admin) {
  std::string nr_var = gen_admin_var(nr_admin, "nr_admin_addr");
  std::string dip_var = gen_admin_var(dip_admin, "dip_admin_addr");

  std::stringstream ss;
  ss << ir_slice1;
  ss << nr_var;
  ss << dip_var;
  std::string nr_item = ir_slice_nr;
  boost::replace_all(nr_item, "VAR", "nr_admin_addr");
  std::string dip_item = ir_slice_dip;
  boost::replace_all(dip_item, "VAR", "dip_admin_addr");

  std::string ir_slice_items = ir_slice_2;
  boost::replace_all(ir_slice_items, "NR_ITEM", nr_item);
  boost::replace_all(ir_slice_items, "DIP_ITEM", dip_item);
  ss << ir_slice_items;
  ss << ir_slice_end;
  return ss.str();
}

neb::bytes gen_auth_table_payload(const address_t &nr_admin,
                                  const address_t &dip_admin) {
  std::string payload = gen_auth_table_ir(nr_admin, dip_admin);
  nbre::NBREIR ir;
  ir.set_name("auth");
  ir.set_version(0);
  ir.set_height(0);
  ir.set_ir(payload);
  ir.set_ir_type(neb::ir_type::cpp);

  std::string ir_str = ir.SerializeAsString();
  return neb::string_to_byte(ir_str);
}

auth_table_generator::auth_table_generator(all_accounts *accounts,
                                           generate_block *block)
    : generator_base(accounts, block, 0, 1) {}

auth_table_generator::~auth_table_generator() {}

std::shared_ptr<corepb::Account> auth_table_generator::gen_account() {
  return nullptr;
}

std::shared_ptr<corepb::Transaction> auth_table_generator::gen_tx() {
  auto ir = gen_auth_table_payload(m_nr_admin_addr, m_dip_admin_addr);
  return m_block->add_protocol_transaction(m_auth_admin_addr, ir);
}

checker_tasks::task_container_ptr_t auth_table_generator::gen_tasks() {
  return nullptr;
}

