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
#include "common/wakeable_queue.h"

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
    op_general,
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
  template <typename T>
  shm_service_op_destroy(boost::interprocess::managed_shared_memory *shmem,
                         T *pointer)
      : shm_service_op_base(op_destroy) {
    m_func = [shmem, pointer]() { shmem->destroy_ptr(pointer); };
  }
  std::function<void()> m_func;
};

class shm_service_op_general : public shm_service_op_base {
public:
  template <typename T>
  shm_service_op_general(T &&f) : shm_service_op_base(op_general), m_func(f) {}
  std::function<void()> m_func;
};

using shm_service_op_queue =
    wakeable_queue<std::shared_ptr<shm_service_op_base>>;
} // namespace internal
} // namespace ipc
} // namespace neb
