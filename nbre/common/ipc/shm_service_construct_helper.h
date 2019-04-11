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
#include <atomic>
#include <condition_variable>
#include <mutex>
#include <thread>

namespace neb {
namespace ipc {

namespace internal {

class shm_service_construct_helper {
public:
  shm_service_construct_helper(
      boost::interprocess::managed_shared_memory *shmem,
      shm_service_op_queue *op_queue);

  template <typename T, typename... ARGS>
  T *construct(ARGS... args){
    if (!m_shmem) {
      LOG(ERROR) << "no shared memory";
      throw shm_service_failure("no shared memory");
    }
    auto thrd_id = std::this_thread::get_id();
    if (thrd_id == m_local_thread_id) {
      return m_shmem->construct<T>(boost::interprocess::anonymous_instance)(
          args...);
    }

    uint64_t counter = m_next_alloc_op_counter.fetch_add(1);
    std::shared_ptr<shm_service_op_allocate> p =
        std::make_shared<shm_service_op_allocate>(counter, [this, args...]() {
          return m_shmem->construct<T>(boost::interprocess::anonymous_instance)(
              args...);
        });
    m_op_queue->push_back(p);
    while (true) {
      std::unique_lock<std::mutex> _l(m_mutex);
      for (auto it = m_finished_alloc_ops.begin();
           it != m_finished_alloc_ops.end(); ++it) {
        auto lp = *it;
        if (lp == p->m_counter) {
          m_finished_alloc_ops.erase(it);
          return (T *)p->m_ret;
        }
      }
      m_cond_var.wait(_l);
    }
  };

  template <typename T> void destroy(T *ptr) {
    std::shared_ptr<shm_service_op_destroy> p =
        std::make_shared<shm_service_op_destroy>(m_shmem, ptr);
    m_op_queue->push_back(p);
  }

  template <typename T> void push_back(T *ptr) {
    std::shared_ptr<shm_service_op_push_back> p =
        std::make_shared<shm_service_op_push_back>();
    p->m_pointer = ptr;
    p->m_type_id = T::pkg_identifier;
    m_op_queue->push_back(p);
  }

  void handle_construct_op(const std::shared_ptr<shm_service_op_base> &op);
  void handle_destroy_op(const std::shared_ptr<shm_service_op_base> &op);

protected:
  std::mutex m_mutex;
  std::condition_variable m_cond_var;
  shm_service_op_queue *m_op_queue;
  boost::interprocess::managed_shared_memory *m_shmem;
  std::vector<uint64_t> m_finished_alloc_ops;
  std::atomic_uint_fast64_t m_next_alloc_op_counter;
  std::thread::id m_local_thread_id;
};
} // namespace internal
} // namespace ipc
} // namespace neb
