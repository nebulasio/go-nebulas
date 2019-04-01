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

class auth_table_check_task : public checker_task_base {
public:
  virtual void check() = 0;
  virtual std::string name() = 0;
  virtual std::string serialize_to_string() = 0;
  virtual void deserialize_from_string(const std::string &s);
};

class auth_table_generator : public generator_base {
public:
  auth_table_generator(all_accounts *accounts, generate_block *block);

  virtual ~auth_table_generator();

  inline void reset() {
    m_nr_admin_addr = address_t();
    m_dip_admin_addr = address_t();
  }

  inline void set_nr_admin_addr(const address_t &addr) {
    m_nr_admin_addr = addr;
  }
  inline void set_dip_admin_addr(const address_t &addr) {
    m_dip_admin_addr = addr;
  }
  inline void set_auth_admin_addr(const address_t &addr) {
    m_auth_admin_addr = addr;
  }

  virtual std::shared_ptr<corepb::Account> gen_account();
  virtual std::shared_ptr<corepb::Transaction> gen_tx();
  virtual checker_tasks::task_container_ptr_t gen_tasks();

protected:
  address_t m_nr_admin_addr;
  address_t m_dip_admin_addr;
  address_t m_auth_admin_addr;
};

neb::bytes gen_auth_table_payload(const address_t &nr_admin,
                                  const address_t &dip_admin);
