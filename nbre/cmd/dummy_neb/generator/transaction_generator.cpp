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

#include "cmd/dummy_neb/generator/transaction_generator.h"

transaction_generator::transaction_generator(all_accounts *accounts,
                                             generate_block *block,
                                             int new_account_num, int tx_num)
    : generator_base(accounts, block, new_account_num, tx_num),
      m_new_addresses(), m_last_used_address_index(0) {}

transaction_generator::~transaction_generator() {}

std::shared_ptr<corepb::Account> transaction_generator::gen_account() {
  auto ret = m_block->gen_user_account(100_nas);
  m_new_addresses.push_back(neb::to_address(ret->address()));
  return ret;
}

std::shared_ptr<corepb::Transaction> transaction_generator::gen_tx() {
  auto from_addr =
      neb::to_address(m_all_accounts->random_user_account()->address());
  address_t to_addr;
  if (m_last_used_address_index < m_new_addresses.size()) {
    to_addr = m_new_addresses[m_last_used_address_index];
    m_last_used_address_index++;
  } else {
    to_addr = neb::to_address(m_all_accounts->random_user_account()->address());
  }
  return m_block->add_binary_transaction(from_addr, to_addr, 1_nas);
}

checker_tasks::task_container_ptr_t transaction_generator::gen_tasks() {
  return nullptr;
}
