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
#include "common/ipc/shm_queue.h"
#include "common/ipc/shm_queue_watcher.h"
#include "common/ipc/shm_service_construct_helper.h"
#include "common/ipc/shm_service_op_queue.h"
#include "common/ipc/shm_service_recv_handler.h"
#include "common/ipc/shm_session.h"
#include "common/quitable_thread.h"
#include "common/util/enable_func_if.h"
#include "core/command.h"
#include <mutex>
#include <thread>
#include <type_traits>
#include <unordered_map>

namespace neb {
namespace ipc {

namespace internal {
class shm_service_base : public quitable_thread {
public:
  shm_service_base(shm_role role, const std::string &shm_name,
                   const std::string &shm_in_name,
                   const std::string &shm_out_name, size_t mem_size,
                   size_t shm_in_capacity, size_t shm_out_capacity);
  virtual ~shm_service_base();

  template <typename T, typename... ARGS> T *construct(ARGS... args) {
    if (!m_shmem) {
      throw shm_service_failure("no shared memory");
    }
    return m_constructer->construct<T>(args...);
  }

  template <typename T> void destroy(T *ptr) { m_constructer->destroy(ptr); }

  template <typename T> void push_back(T *ptr) {
    m_constructer->push_back(ptr);
  }

  template <typename T, typename Func> void add_handler(Func &&f) {
    m_recv_handler->add_handler<T>(f);
  }

  void init_local_env();

  //! this will block the current thread.
  void run();

  void reset();

  inline shm_session_base *session() { return m_session.get(); }
  inline char_allocator_t &char_allocator() { return *m_char_allocator; }
  inline default_allocator_t &default_allocator() {
    return *m_default_allocator;
  }

  template <typename T> void schedule_task_in_service_thread(T &&func) {
    auto task =
        std::shared_ptr<shm_service_op_base>(new shm_service_op_general(func));
    m_op_queue->push_back(task);
  }

private:
  void init_local_interprocess_var();

  virtual void thread_func();

  inline std::string semaphore_name() const {
    return m_shm_name + std::string(".quit_semaphore");
  }

  inline std::string mutex_name() const {
    return m_shm_name + std::string(".mutex");
  }

protected:
  shm_role m_role;
  size_t m_mem_size;
  size_t m_shm_in_capacity;
  size_t m_shm_out_capacity;
  std::string m_shm_name;
  std::string m_shm_in_name;
  std::string m_shm_out_name;
  boost::interprocess::managed_shared_memory *m_shmem;
  shm_queue *m_in_buffer;
  shm_queue *m_out_buffer;
  std::unique_ptr<shm_session_base> m_session;
  std::unique_ptr<shm_service_op_queue> m_op_queue;
  std::unique_ptr<shm_service_construct_helper> m_constructer;
  std::unique_ptr<shm_service_recv_handler> m_recv_handler;
  std::unique_ptr<shm_queue_watcher> m_queue_watcher;
  std::unique_ptr<char_allocator_t> m_char_allocator;
  std::unique_ptr<default_allocator_t> m_default_allocator;
}; // end class shm_service_base
}

template <size_t S>
class shm_service_server : public internal::shm_service_base {
public:
  shm_service_server(const std::string &name, size_t in_obj_max_count,
                     size_t out_obj_max_count)
      : internal::shm_service_base(
            role_server, name,
            internal::shm_other_side_role<shm_server>::type::role_name(name),
            shm_server::role_name(name), S, in_obj_max_count,
            out_obj_max_count) {}

  void wait_until_client_start() {
    internal::shm_session_server *ss =
        (internal::shm_session_server *)internal::shm_service_base::session();
    ss->wait_until_client_start();
  }
};

template <size_t S>
class shm_service_client : public internal::shm_service_base {
public:
  shm_service_client(const std::string &name, size_t in_obj_max_count,
                     size_t out_obj_max_count)
      : internal::shm_service_base(
            role_client, name,
            internal::shm_other_side_role<shm_client>::type::role_name(name),
            shm_client::role_name(name), S, in_obj_max_count,
            out_obj_max_count) {}
};

template <size_t S> class shm_service_util : public internal::shm_service_base {
public:
  shm_service_util(const std::string &name, size_t in_obj_max_count,
                   size_t out_obj_max_count)
      : internal::shm_service_base(
            role_util, name,
            internal::shm_other_side_role<shm_server>::type::role_name(name),
            shm_server::role_name(name), S, in_obj_max_count,
            out_obj_max_count) {}
};
}
}
