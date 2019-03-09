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

class nbre_version_checker : public checker_task_base {
public:
  nbre_version_checker();
  virtual ~nbre_version_checker();
  virtual void check();
  virtual std::string name() const;
};

class nbre_nr_result_check : public checker_task_base {
public:
  nbre_nr_result_check(const std::string &nr_handle);

  virtual ~nbre_nr_result_check();
  virtual void check();
  virtual std::string name() const;

protected:
  std::string m_nr_handle;
};

class nbre_nr_handle_check : public checker_task_base {
public:
  nbre_nr_handle_check(uint64_t start_block, uint64_t end_block);
  virtual ~nbre_nr_handle_check();
  virtual void check();
  virtual std::string name() const;

protected:
  uint64_t m_start_block;
  uint64_t m_end_block;
  std::shared_ptr<nbre_nr_result_check> m_nr_result_checker;
};

class nbre_dip_reward_check : public checker_task_base {
public:
  nbre_dip_reward_check(uint64_t height);
  virtual ~nbre_dip_reward_check();
  virtual void check();
  virtual std::string name() const;

protected:
  uint64_t m_height;
};
