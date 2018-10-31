// Copyright (C) 2017 go-nebulas authors
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

#include "core/ir_warden.h"
#include "common/configuration.h"
#include "common/timer_loop.h"
#include "core/command.h"
#include <boost/bind.hpp>
#include <boost/date_time/posix_time/posix_time.hpp>

namespace neb {
namespace core {

const command_type_t exit_command::command_type;
ir_warden::~ir_warden() {}

std::shared_ptr<nbre::NBREIR>
ir_warden::get_ir_by_name_version(const std::string &name, uint64_t version) {
  return m_nbre_storage->read_nbre_by_name_version(name, version);
}

std::vector<std::shared_ptr<nbre::NBREIR>>
ir_warden::get_ir_by_name_height(const std::string &name, uint64_t height) {
  return m_nbre_storage->read_nbre_by_height(name, height);
}

bool ir_warden::is_sync_already() const {
  return m_nbre_storage->is_latest_irreversible_block();
}

void ir_warden::wait_until_sync() {
  LOG(INFO) << "wait until sync ...";
  std::unique_lock<std::mutex> _l(m_sync_mutex);
  if (m_is_sync_already) {
    return;
  }
  m_sync_cond_var.wait(_l);
  LOG(INFO) << "wait until sync done";
}

void ir_warden::on_timer() { m_nbre_storage->write_nbre(); }

void ir_warden::thread_func() {

  m_nbre_storage->write_nbre();
  std::unique_lock<std::mutex> _l(m_sync_mutex);
  m_is_sync_already = true;
  _l.unlock();
  m_sync_cond_var.notify_one();

  timer_loop tl(&m_io_service);
  tl.register_timer_and_callback(
      neb::configuration::instance().ir_warden_time_interval(),
      [&]() { on_timer(); });
  m_io_service.run();
}

void ir_warden::async_run() {
  if (m_thread) {
    return;
  }
  start();
}

ir_warden::ir_warden() : quitable_thread(), m_is_sync_already(false) {
  m_nbre_storage = std::unique_ptr<fs::nbre_storage>(new fs::nbre_storage(
      std::getenv("NBRE_DB"), std::getenv("BLOCKCHAIN_DB")));
}
}
} // namespace neb
