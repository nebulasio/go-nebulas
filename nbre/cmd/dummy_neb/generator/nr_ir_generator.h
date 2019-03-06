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
#include "cmd/dummy_neb/dummy_common.h"
#include "cmd/dummy_neb/generator/generator_base.h"

class nr_ir_generator : public generator_base {
public:
  nr_ir_generator(generate_block *block, const address_t &nr_admin_addr);

  virtual ~nr_ir_generator();

  int64_t m_a;
  int64_t m_b;
  int64_t m_c;
  int64_t m_d;
  float m_theta;
  float m_mu;
  float m_lambda;
  uint32_t m_major_version;
  uint32_t m_minor_version;
  uint32_t m_patch_version;

  virtual std::shared_ptr<corepb::Account> gen_account();
  virtual std::shared_ptr<corepb::Transaction> gen_tx();
  virtual checker_tasks::task_container_ptr_t gen_tasks();

protected:
  address_t m_nr_admin_addr;
};
