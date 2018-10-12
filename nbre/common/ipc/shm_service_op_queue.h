// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or // (at your
// option) any later version.
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
#include <condition_variable>
#include <mutex>
#include <queue>
#include <thread>

namespace neb {
namespace ipc {
namespace internal {

class shm_service_op_base {
public:
  enum shm_service_op_id {
    op_allocate_obj,
    op_recv_obj,
    op_push_back,
    op_destroy,
  };
  inline shm_service_op_base(shm_service_op_id op_id) : m_op_id(op_id) {}

  inline shm_service_op_id op_id() const { return m_op_id; }

protected:
  shm_service_op_id m_op_id;
};

class shm_service_op_allocate : public shm_service_op_base {
public:
  inline shm_service_op_allocate(uint64_t counter,
                                 const std::function<void *()> &func)
      : shm_service_op_base(op_allocate_obj), m_counter(counter), m_func(func) {
  }

  uint64_t m_counter;
  std::function<void *()> m_func;
  void *m_ret;
};

class shm_service_op_recv : public shm_service_op_base {
public:
  inline shm_service_op_recv() : shm_service_op_base(op_recv_obj) {}
  void *m_pointer;
  shm_type_id_t m_type_id;
};

class shm_service_op_push_back : public shm_service_op_base {
public:
  inline shm_service_op_push_back() : shm_service_op_base(op_push_back) {}
  void *m_pointer;
  shm_type_id_t m_type_id;
};
class shm_service_op_destroy : public shm_service_op_base {
public:
  inline shm_service_op_destroy() : shm_service_op_base(op_destroy) {}
  void *m_pointer;
};

class shm_service_op_queue {
public:
  typedef std::queue<std::shared_ptr<shm_service_op_base>> queue_t;
  shm_service_op_queue() = default;

  void push_back(const queue_t::value_type &op);

  std::pair<bool, queue_t::value_type> pop_front();

  std::pair<bool, queue_t::value_type> try_pop_front();

  size_t size() const;

  bool empty() const;

  void wake_up_if_empty();

protected:
  queue_t m_queue;
  mutable std::mutex m_mutex;
  std::condition_variable m_cond_var;
};
} // namespace internal
} // namespace ipc
} // namespace neb
