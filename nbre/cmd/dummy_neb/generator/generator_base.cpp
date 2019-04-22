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
#include "cmd/dummy_neb/generator/generator_base.h"

uint64_t checker_task_base::s_task_id = 0;

checker_task_base::checker_task_base()
    : m_task_id(s_task_id++), m_last_call_timepoint(), m_call_times(0),
      m_b_is_running(false){};

checker_task_base::~checker_task_base() {}

std::string checker_task_base::name() const { return ""; }

void checker_task_base::apply_result(const std::string &result) {
  std::unique_lock<std::mutex> _l(m_mutex);
  m_exist_results.insert(result);
  m_last_call_timepoint = std::chrono::steady_clock::now();
  m_call_times++;
  m_b_is_running = false;
}

std::string checker_task_base::status() const {
  std::unique_lock<std::mutex> _l(m_mutex);
  std::stringstream ss;
  ss << "->" << name() << " called " << m_call_times << " times, with "
     << m_exist_results.size() << " diff results. \n";
  for (auto &s : m_exist_results) {
    ss << "\t\t" << s << '\n';
  }
  return ss.str();
}
#if 0
void checker_tasks::init_from_db() {
  try {
    std::string s =
        bc_storage_session::instance().get_string(get_all_checker_info_key());
    checker_marshaler cm;
    cm.deserialize_from_string(s);
    for (auto &sub_s : cm.get<p_checkers>()) {
      checker_marshaler sub_cm;
      sub_cm.deserialize_from_string(sub_s);
      for (auto &cs : sub_cm.get<p_checkers>()) {
        auto checker = init_checker_from_string(cs);
        add_task(checker);
      }
    }
  } catch (...) {
  }
}

void checker_tasks::write_to_db() {
  checker_marshaler cm;
  for (auto &it : m_all_tasks) {
    cm.get<p_checkers>().push_back(it.first);
    checker_marshaler sub_cm;
    for (auto ik : *(it.second)) {
      sub_cm.get<p_checkers>().push_back(ik->serialize_to_string());
    }
    std::string ss = sub_cm.serialize_to_string();
    bc_storage_session::instance().put(get_checker_key_with_name(it.first), ss);
  }
  std::string s = cm.serialize_to_string();
  bc_storage_session::instance().put(get_all_checker_info_key(), s);
}
#endif

void checker_tasks::add_task(const std::shared_ptr<checker_task_base> &task) {
  if (!task)
    return;
  std::unique_lock<std::mutex> _l(m_mutex);
  m_all_tasks.insert(std::make_pair(task->task_id(), task));
}
void checker_tasks::randomly_schedule_no_running_tasks() {
  std::unique_lock<std::mutex> _l(m_mutex);
  std::vector<uint64_t> keys;
  for (auto &kv : m_all_tasks) {
    keys.push_back(kv.first);
  }
  if (keys.size() == 0)
    return;
  uint64_t k = keys[std::rand() % keys.size()];
  uint64_t max_num = 16;
  uint64_t i = 0;
  while (m_all_tasks[k]->is_running()) {
    if (i > max_num) {
      return;
    }
    i++;
    k = keys[std::rand() % keys.size()];
  }
  std::shared_ptr<checker_task_base> task = m_all_tasks[k];
  // task->check();
  task_executor::instance().schedule([task]() { task->check(); });
}

void checker_tasks::randomly_schedule_all_tasks(int num) {
  std::unique_lock<std::mutex> _l(m_mutex);
  std::vector<uint64_t> keys;
  for (auto &kv : m_all_tasks) {
    keys.push_back(kv.first);
  }
  if (keys.empty())
    return;
  for (int i = 0; i < num; ++i) {
    uint64_t k = keys[std::rand() % keys.size()];
    auto task = m_all_tasks[k];
    task_executor::instance().schedule([task]() { task->check(); });
  }
}

std::string checker_tasks::status() const {
  std::unique_lock<std::mutex> _l(m_mutex);

  std::stringstream ss;
  ss << m_all_tasks.size() << "tasks in total.\n";
  for (auto &kv : m_all_tasks) {
    ss << kv.second->status();
  }
  return ss.str();
}

std::shared_ptr<checker_task_base>
init_checker_from_string(const std::string &s) {
#define add_type(type)                                                         \
  if (type().name() == s) {                                                    \
    auto t = std::make_shared<type>();                                         \
    t->deserialize_from_string(s);                                             \
    return t;                                                                  \
  }
  return nullptr;

#undef add_type
}

generator_base::generator_base(all_accounts *accounts, generate_block *block,
                               int new_account_num, int tx_num)
    : m_all_accounts(accounts), m_block(block),
      m_new_account_num(new_account_num), m_new_tx_num(tx_num) {}

void generator_base::run() {
  for (int i = 0; i < m_new_account_num; ++i) {
    gen_account();
  }
  for (int i = 0; i < m_new_tx_num; ++i) {
    gen_tx();
  }

  auto checkers = gen_tasks();
  if (!checkers)
    return;
  for (auto &checker : *checkers) {
    checker_tasks::instance().add_task(checker);
  }
}
