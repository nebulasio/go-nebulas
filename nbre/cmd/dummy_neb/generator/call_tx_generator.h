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
#pragma once
#include "cmd/dummy_neb/generator/generator_base.h"

class call_tx_generator : public generator_base {
public:
  call_tx_generator(generate_block *block, int call_tx_number);
  virtual ~call_tx_generator();

  virtual std::shared_ptr<corepb::Account> gen_account();
  virtual std::shared_ptr<corepb::Transaction> gen_tx();
  virtual checker_tasks::task_container_ptr_t gen_tasks();

private:
  std::vector<address_t> m_contract_accounts;
};
