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
#include "cmd/dummy_neb/generator/cli_generator.h"
#include "auth_table_generator.h"
#include "cmd/dummy_neb/cli/pkg.h"

cli_generator::cli_generator() : generator_base(nullptr, nullptr, 0, 0) {
}
cli_generator::~cli_generator() {}

void cli_generator::update_info(generate_block *block) {
  m_block = block;
  m_all_accounts = block->get_all_accounts();
  m_new_tx_num = m_pkgs.size();
}

std::shared_ptr<corepb::Account> cli_generator::gen_account() {
  return nullptr;
}
std::shared_ptr<corepb::Transaction> cli_generator::gen_tx() {
  while (!m_pkgs.empty()) {
    auto ret = m_pkgs.try_pop_front();
    if (!ret.first)
      continue;
    auto pkg = ret.second;
    address_t to_addr;
    if (pkg->type_id() == cli_submit_ir_pkg) {
      cli_submit_ir_t *req = (cli_submit_ir_t *)pkg.get();
      std::string payload_base64 = req->get<p_payload>();
      auto payload_bytes = neb::bytes::from_base64(payload_base64);

      if (req->get<p_type>() == "nr") {
        if (m_nr_admin_addr.empty()) {
          m_nr_admin_addr = m_all_accounts->random_user_addr();
        }
        to_addr = m_nr_admin_addr;
        LOG(INFO) << "submit nr ir";
        return m_block->add_protocol_transaction(to_addr, payload_bytes);
      } else if (req->get<p_type>() == "dip") {
        if (m_dip_admin_addr.empty()) {
          m_dip_admin_addr = m_all_accounts->random_user_addr();
        }
        to_addr = m_dip_admin_addr;
        LOG(INFO) << "submit dip ir";
        return m_block->add_protocol_transaction(to_addr, payload_bytes);

      } else if (req->get<p_type>() == "auth") {
        if (m_auth_admin_addr.empty()) {
          m_auth_admin_addr = m_all_accounts->random_user_addr();
        }
        if (m_nr_admin_addr.empty()) {
          m_nr_admin_addr = m_all_accounts->random_user_addr();
        }
        if (m_dip_admin_addr.empty()) {
          m_dip_admin_addr = m_all_accounts->random_user_addr();
        }
        to_addr = m_auth_admin_addr;
        LOG(INFO) << "submit auth table";
        return m_block->add_protocol_transaction(
            to_addr, gen_auth_table_payload(m_nr_admin_addr, m_dip_admin_addr));
      }
    }
  }
  return nullptr;
}
checker_tasks::task_container_ptr_t cli_generator::gen_tasks() {
  return nullptr;
}

