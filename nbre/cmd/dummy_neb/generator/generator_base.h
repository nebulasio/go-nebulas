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
#include "util/quitable_thread.h"
#include <ff/network.h>

define_nt(p_checkers, std::vector<std::string>);
using checker_marshaler = ff::net::ntpackage<1, p_checkers>;

class checker_task_base {
public:
  checker_task_base();

  virtual ~checker_task_base();
  virtual void check() = 0;
  virtual std::string name() const;
  inline uint64_t &task_id() { return m_task_id; };
  inline uint64_t task_id() const { return m_task_id; }
  bool is_running() const { return m_b_is_running; }

  std::string status() const;

protected:
  void apply_result(const std::string &result);

protected:
  mutable std::mutex m_mutex;
  uint64_t m_task_id;
  std::chrono::steady_clock::time_point m_last_call_timepoint;
  std::unordered_set<std::string> m_exist_results;
  uint64_t m_call_times;
  uint16_t m_diff_result_num;
  bool m_b_is_running;

  static uint64_t s_task_id;
};

std::shared_ptr<checker_task_base>
init_checker_from_string(const std::string &s);

class task_executor : public neb::util::wakeable_thread,
                      public neb::util::singleton<task_executor> {};

class checker_tasks : public neb::util::singleton<checker_tasks> {
public:
  typedef std::vector<std::shared_ptr<checker_task_base>> task_container_t;
  typedef std::shared_ptr<task_container_t> task_container_ptr_t;

  // void init_from_db();
  // void write_to_db();

  void add_task(const std::shared_ptr<checker_task_base> &task);
  task_container_ptr_t get_tasks_with_name(const std::string &name);
  inline size_t size() const { return m_all_tasks.size(); }

  void randomly_schedule_no_running_tasks();
  void randomly_schedule_all_tasks(int num = 1);

  std::string status() const;

protected:
  inline static std::string get_all_checker_info_key() {
    return std::string("all_checker_names");
  }
  inline static std::string get_checker_key_with_name(const std::string &name) {
    return std::string("checker_with_name") + name;
  }

protected:
  typedef std::unordered_map<std::string, task_container_ptr_t>
      task_name_container_t;
  mutable std::mutex m_mutex;
  std::unordered_map<uint64_t, std::shared_ptr<checker_task_base>> m_all_tasks;
};


class generator_base {
public:
  generator_base(all_accounts *accounts, generate_block *block,
                 int new_account_num, int tx_num);
  inline virtual ~generator_base(){};

  virtual void run();
  virtual std::shared_ptr<corepb::Account> gen_account() = 0;
  virtual std::shared_ptr<corepb::Transaction> gen_tx() = 0;
  virtual checker_tasks::task_container_ptr_t gen_tasks() = 0;

protected:
  all_accounts *m_all_accounts;
  generate_block *m_block;
  int m_new_account_num;
  int m_new_tx_num;
};

