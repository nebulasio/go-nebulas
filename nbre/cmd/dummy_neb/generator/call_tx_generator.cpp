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
#include "cmd/dummy_neb/generator/call_tx_generator.h"

call_tx_generator::call_tx_generator(generate_block *block, int call_tx_number)
    : generator_base(block->get_all_accounts(), block, 0, call_tx_number) {}

call_tx_generator::~call_tx_generator() {}

std::shared_ptr<corepb::Account> call_tx_generator::gen_account() {
  return nullptr;
}

std::shared_ptr<corepb::Transaction> call_tx_generator::gen_tx() {
  auto from_addr = m_all_accounts->random_user_addr();
  address_t contract_addr;
  if (m_contract_accounts.empty()) {
    contract_addr = m_all_accounts->random_contract_addr();
  } else {
    contract_addr =
        m_contract_accounts.at(std::rand() % m_contract_accounts.size());
  }
  if (from_addr.empty() || contract_addr.empty()) {
    return nullptr;
  }
  m_contract_accounts.push_back(contract_addr);
  return m_block->add_call_transaction(from_addr, contract_addr);
}
checker_tasks::task_container_ptr_t call_tx_generator::gen_tasks() {
  return nullptr;
}
