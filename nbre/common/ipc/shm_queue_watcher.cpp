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
#include "common/ipc/shm_queue_watcher.h"
#include "core/command.h"

namespace neb {
namespace ipc {
namespace internal {
shm_queue_watcher::shm_queue_watcher(shm_queue *queue,
                                     shm_service_op_queue *op_queue)
    : m_queue(queue), m_op_queue(op_queue) {

  neb::core::command_queue::instance().listen_command<neb::core::exit_command>(
      this, [this](const std::shared_ptr<neb::core::exit_command> &) {
        m_queue->wake_up_if_empty();
      });
}

void shm_queue_watcher::thread_func() {
  while (!m_exit_flag) {
    std::tuple<void *, shm_type_id_t, shm_queue::element_op_tag> r =
        m_queue->pop_front();
    if (std::get<0>(r)) {
      std::shared_ptr<shm_service_op_recv> p =
          std::make_shared<shm_service_op_recv>();
      p->m_pointer = std::get<0>(r);
      p->m_type_id = std::get<1>(r);
      m_op_queue->push_back(p);
    }
  }
}
} // namespace internal
} // namespace ipc
} // namespace neb
