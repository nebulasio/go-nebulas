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
#include "cmd/dummy_neb/generator/genesis_generator.h"

genesis_generator::genesis_generator(generate_block *block, int number,
                                     nas init_value)
    : generator_base(block->get_all_accounts(), block, number, 0),
      m_number(number), m_init_value(init_value) {}

genesis_generator::~genesis_generator() {}

std::shared_ptr<corepb::Account> genesis_generator::gen_account() {
  return m_block->gen_user_account(m_init_value);
}
std::shared_ptr<corepb::Transaction> genesis_generator::gen_tx() {
  return nullptr;
}
checker_tasks::task_container_ptr_t genesis_generator::gen_tasks() {
  return nullptr;
}
