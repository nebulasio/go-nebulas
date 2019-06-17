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
#include "core/command.h"
#include "core/execution_context.h"
#include "fs/blockchain.h"
#include "fs/ir_manager/ir_processor.h"
#include "fs/storage.h"
#include <boost/bind.hpp>
#include <boost/date_time/posix_time/posix_time.hpp>

namespace neb {
namespace core {

const command_type_t exit_command::command_type;
ir_warden::~ir_warden() {}


bool ir_warden::is_sync_already() const {
  std::unique_lock<std::mutex> _l(m_sync_mutex);
  return m_is_sync_already;
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

void ir_warden::on_timer() {
  context->ir_processor()->parse_irs(m_queue);
  std::unique_lock<std::mutex> _l(m_sync_mutex);
  if (!m_is_sync_already) {
    m_is_sync_already = true;
    _l.unlock();
    m_sync_cond_var.notify_one();
  }
}

void ir_warden::on_receive_ir_transactions(
    const std::shared_ptr<nbre_ir_transactions_req> &txs_ptr) {
  m_queue.push_back(txs_ptr);
}

ir_warden::ir_warden() : m_is_sync_already(false) {}

}
} // namespace neb
