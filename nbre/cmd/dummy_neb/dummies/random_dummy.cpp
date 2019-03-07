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
#include "cmd/dummy_neb/dummies/random_dummy.h"

random_dummy::random_dummy(const std::string &name, int initial_account_num,
                           nas initial_nas, double account_increase_ratio)
    : dummy_base(name), m_initial_account_num(initial_account_num),
      m_initial_nas(initial_nas),
      m_account_increase_ratio(account_increase_ratio), m_auth_ratio(0) {}

random_dummy::~random_dummy() {}

std::shared_ptr<generate_block> random_dummy::generate_LIB_block() {
  std::shared_ptr<generate_block> ret =
      std::make_shared<generate_block>(&m_all_accounts, m_current_height);

  if (m_current_height == 0) {
    genesis_generator g(ret.get(), m_initial_account_num, m_initial_nas);
    g.run();
  } else {
    int account_num = m_account_increase_ratio * m_initial_account_num;
    int tx_num = account_num + std::rand() % m_initial_account_num;

    m_tx_gen = std::make_unique<transaction_generator>(
        &m_all_accounts, ret.get(),
        m_account_increase_ratio * m_initial_account_num, tx_num);
    m_tx_gen->run();

    if (m_auth_ratio != 0 && std::rand() % 1000 < m_auth_ratio * 1000) {
      m_auth_gen =
          std::make_unique<auth_table_generator>(&m_all_accounts, ret.get());
      address_t nr_admin_addr =
          neb::to_address(m_all_accounts.random_user_account()->address());
      address_t dip_admin_addr =
          neb::to_address(m_all_accounts.random_user_account()->address());
      m_nr_admin_addr = nr_admin_addr;
      m_dip_admin_addr = dip_admin_addr;
      m_auth_gen->set_auth_admin_addr(m_auth_admin_addr);
      m_auth_gen->set_nr_admin_addr(nr_admin_addr);
      m_auth_gen->set_dip_admin_addr(dip_admin_addr);
      m_auth_gen->run();
    }

    if (m_nr_ratio != 0 && std::rand() % 1000 < m_nr_ratio * 1000) {
      if (m_nr_admin_addr.empty()) {
        m_nr_admin_addr =
            neb::to_address(m_all_accounts.random_user_account()->address());
      }
      m_nr_gen = std::make_unique<nr_ir_generator>(ret.get(), m_nr_admin_addr);
      random_increase_version(m_nr_version);
      m_nr_gen->m_major_version = m_nr_version.major_version();
      m_nr_gen->m_minor_version = m_nr_version.minor_version();
      m_nr_gen->m_patch_version = m_nr_version.patch_version();

      m_nr_gen->run();
    }

    if (m_contract_ratio != 0 && std::rand() % 1000 < m_contract_ratio * 1000) {
      m_contract_gen = std::make_unique<contract_generator>(ret.get(), 1);
      m_contract_gen->run();
    }
    if (m_call_ratio != 0 && std::rand() % 1000 < m_call_ratio * 1000) {
      m_call_gen = std::make_unique<call_tx_generator>(
          ret.get(), std::rand() % (m_all_accounts.size() / 5));
      m_call_gen->run();
    }
  }

  m_current_height++;
  return ret;
}

address_t random_dummy::enable_auth_gen_with_ratio(double auth_ratio) {
  if (m_current_height == 0)
    generate_LIB_block();
  corepb::Account *admin_account = m_all_accounts.random_user_account();
  address_t admin_addr = neb::to_address(admin_account->address());
  m_auth_admin_addr = admin_addr;
  m_auth_ratio = auth_ratio;
  return admin_addr;
}

void random_dummy::enable_nr_ir_with_ratio(double nr_ratio) {
  if (m_current_height == 0)
    generate_LIB_block();
  m_nr_ratio = nr_ratio;
}

void random_dummy::enable_call_tx_with_ratio(double contract_ratio,
                                             double call_ratio) {
  if (m_current_height == 0)
    generate_LIB_block();
  m_contract_ratio = contract_ratio;
  m_call_ratio = call_ratio;
}

std::shared_ptr<checker_task_base> random_dummy::generate_checker_task() {
  return nullptr;
}

address_t random_dummy::get_auth_admin_addr() { return m_auth_admin_addr; }
