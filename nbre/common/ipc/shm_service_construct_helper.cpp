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
#include "common/ipc/shm_service_construct_helper.h"

namespace neb {
namespace ipc {
namespace internal {
shm_service_construct_helper::shm_service_construct_helper(
    boost::interprocess::managed_shared_memory *shmem,
    shm_service_op_queue *op_queue)
    : m_op_queue(op_queue), m_shmem(shmem), m_next_alloc_op_counter(0) {
  m_local_thread_id = std::this_thread::get_id();
}

void shm_service_construct_helper::handle_construct_op(
    const std::shared_ptr<shm_service_op_base> &op) {
  assert(op->op_id() == shm_service_op_base::op_allocate_obj);
  shm_service_op_allocate *alloc_op =
      static_cast<shm_service_op_allocate *>(op.get());
  alloc_op->m_ret = alloc_op->m_func();
  std::unique_lock<std::mutex> _l(m_mutex);
  m_finished_alloc_ops.push_back(alloc_op->m_counter);
  m_cond_var.notify_all();
}

void shm_service_construct_helper::handle_destroy_op(
    const std::shared_ptr<shm_service_op_base> &op) {
  assert(op->op_id() == shm_service_op_base::op_destroy);

  shm_service_op_destroy *dry_op =
      static_cast<shm_service_op_destroy *>(op.get());
  dry_op->m_func();
}
} // namespace internal
} // namespace ipc
} // namespace neb
