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
#include "cmd/dummy_neb/dummies/dummy_base.h"
#include "cmd/dummy_neb/generator/generators.h"

class random_dummy : public dummy_base {
public:
  random_dummy(const std::string &name, int initial_account_num,
               nas initial_nas, double account_increase_ratio,
               double auth_ratio);

  virtual ~random_dummy();
  virtual std::shared_ptr<generate_block> generate_LIB_block();

  virtual std::shared_ptr<checker_task_base> generate_checker_task();

protected:
  all_accounts m_all_accounts;
  std::unique_ptr<transaction_generator> m_tx_gen;
  std::unique_ptr<genesis_generator> m_genesis_gen;
  int m_initial_account_num;
  nas m_initial_nas;
  double m_account_increase_ratio;
  double m_auth_ratio;
};
