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
#include "cmd/dummy_neb/dummy_driver.h"
#include "cmd/dummy_neb/dummies/dummies.h"
#include "cmd/dummy_neb/dummy_common.h"
#include "cmd/dummy_neb/generator/generators.h"

dummy_driver::dummy_driver() : m_io_service(), m_block_interval_seconds(0) {}

void dummy_driver::add_dummy(const std::shared_ptr<dummy_base> &dummy) {
  m_all_dummies.insert(std::make_pair(dummy->name(), dummy));
}
dummy_driver::~dummy_driver() { LOG(INFO) << "To quit dummy driver"; }

std::vector<std::string> dummy_driver::get_all_dummy_names() const {
  std::vector<std::string> ret;
  for (auto it : m_all_dummies) {
    ret.push_back(it.first);
  }
  return ret;
}

void dummy_driver::run(const std::string &dummy_name, uint64_t block_interval) {
  if (m_block_interval_seconds != 0)
    throw std::invalid_argument("already run");

  auto it = m_all_dummies.find(dummy_name);
  if (it == m_all_dummies.end())
    throw std::invalid_argument("can't find dummy name");

  std::shared_ptr<dummy_base> dummy = it->second;

  dummy->init_from_db();

  // This for generating block with interval.
  m_block_interval_seconds = block_interval;
  m_block_gen_timer = std::make_unique<neb::util::timer_loop>(&m_io_service);
  LOG(INFO) << "start block gen timer with interval: "
            << m_block_interval_seconds << ", for dummy: " << dummy_name;
  m_block_gen_timer->register_timer_and_callback(
      m_block_interval_seconds, [dummy]() {
        auto block = dummy->generate_LIB_block();
        block->write_to_blockchain_db();
        block_height_t height = dummy->current_height();

        ipc_nbre_ir_transactions_create(nullptr, height);
        for (auto &tx : block->all_transactions()) {
          corepb::Data d = tx->data();
          if (d.type() == "protocol") {
            std::string s = tx->SerializeAsString();
            ipc_nbre_ir_transactions_append(nullptr, height, s.data(),
                                            s.size());
            LOG(INFO) << "append protocol tx";
          }
        }
        ipc_nbre_ir_transactions_send(nullptr, height);
        LOG(INFO) << "gen block " << height;
      });

  m_checker_gen_timer = std::make_unique<neb::util::timer_loop>(&m_io_service);
  m_checker_gen_timer->register_timer_and_callback(1, [dummy]() {
    // dummy->generate_checker_task();
    // checker_tasks::instance().randomly_schedule_no_running_tasks();
    // checker_tasks::instance().randomly_schedule_all_tasks(20);
  });

  m_io_service.run();
}

void dummy_driver::shutdown() { m_io_service.stop(); }

void dummy_driver::reset_dummy(const std::string &dummy_name) {
  auto it = m_all_dummies.find(dummy_name);
  if (it == m_all_dummies.end())
    throw std::invalid_argument("can't find dummy name");

  std::shared_ptr<dummy_base> dummy = it->second;
  dummy->clean_db();
}

std::shared_ptr<dummy_base>
dummy_driver::get_dummy_with_name(const std::string &dummy_name) const {
  auto it = m_all_dummies.find(dummy_name);
  if (it == m_all_dummies.end())
    throw std::invalid_argument("can't find dummy name");

  std::shared_ptr<dummy_base> dummy = it->second;
  return dummy;
}
