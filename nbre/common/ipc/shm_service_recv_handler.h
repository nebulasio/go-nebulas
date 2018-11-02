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
#include "common/common.h"
#include "common/ipc/shm_base.h"
#include "common/ipc/shm_service_op_queue.h"
#include "common/quitable_thread.h"

namespace neb {
namespace ipc {
namespace internal {
class shm_service_recv_handler {
public:
  shm_service_recv_handler(boost::interprocess::managed_shared_memory *shmem,
                           shm_service_op_queue *op_queue);

  template <typename T, typename Func> void add_handler(Func &&f) {
    std::lock_guard<std::mutex> _l(m_handlers_mutex);
    m_all_handlers.insert(std::make_pair(T::pkg_identifier, [this, f](void *p) {
      T *r = (T *)p;
      f(r);
      std::shared_ptr<shm_service_op_destroy> tp =
          std::make_shared<shm_service_op_destroy>(m_shmem, r);
      m_op_queue->push_back(tp);
    }));
  }

  void handle_recv_op(const std::shared_ptr<shm_service_op_base> &op);

protected:
  typedef std::function<void(void *)> pkg_handler_t;
  boost::interprocess::managed_shared_memory *m_shmem;
  std::mutex m_handlers_mutex;
  std::unordered_map<shm_type_id_t, pkg_handler_t> m_all_handlers;
  shm_service_op_queue *m_op_queue;
  wakeable_thread m_handler_thread;
};
} // namespace internal
} // namespace ipc
} // namespace neb
