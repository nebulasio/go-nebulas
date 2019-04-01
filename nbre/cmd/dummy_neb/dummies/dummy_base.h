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

class dummy_base {
public:
  dummy_base(const std::string &name);
  virtual ~dummy_base();

  virtual void init_from_db();

  inline std::string name() { return m_name; }
  inline block_height_t current_height() const { return m_current_height; }

  void clean_db();

  virtual std::shared_ptr<generate_block> generate_LIB_block() = 0;

  virtual std::shared_ptr<checker_task_base> generate_checker_task() = 0;

  virtual address_t get_auth_admin_addr() = 0;

  const std::string &db_path() const;

protected:
  static void random_increase_version(neb::version &v);

protected:
  std::string m_name;
  mutable std::string m_db_path;
  block_height_t m_current_height;
};
