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
#include "cmd/dummy_neb/generator/contract_generator.h"

contract_generator::contract_generator(generate_block *block,
                                       int contract_number)
    : generator_base(block->get_all_accounts(), block, contract_number, 0) {}

contract_generator::~contract_generator() {}

std::shared_ptr<corepb::Account> contract_generator::gen_account() {
  auto from_addr = m_all_accounts->random_user_addr();
  if (from_addr.empty()) {
    return nullptr;
  }
  auto ret = m_block->add_deploy_transaction(from_addr, neb::bytes());
  return ret;
}

std::shared_ptr<corepb::Transaction> contract_generator::gen_tx() {
  return nullptr;
}

checker_tasks::task_container_ptr_t contract_generator::gen_tasks() {
  return nullptr;
}
