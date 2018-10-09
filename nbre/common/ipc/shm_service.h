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

struct shm_service_failure : public std::exception {
  inline shm_service_failure(const std::string &msg) : m_msg(msg) {}
  inline const char *what() const throw() { return m_msg.c_str(); }
protected:
  std::string m_msg;
};
namespace internal {
template <size_t S, typename Role>
class shm_service_base : public quitable_thread {
public:
  shm_service_base(const std::string &shm_name, const std::string &shm_in_name,
                   const std::string &shm_out_name, size_t shm_in_capacity,
                   size_t shm_out_capacity)
      : m_shm_name(shm_name), m_shm_in_name(shm_in_name),
        m_shm_out_name(shm_out_name) {

    neb::util::enable_func_if<std::is_same<Role, shm_server>::value>(
        [this, shm_name]() {
          m_session = std::unique_ptr<shm_session_base>(
              new shm_session_server(shm_name));
        });

    neb::util::enable_func_if<std::is_same<Role, shm_client>::value>(
        [this, shm_name]() {
          m_session = std::unique_ptr<shm_session_base>(
              new shm_session_client(shm_name));
        });

    m_session->bookkeeper()->acquire(shm_name, [this, shm_name]() {
      m_shmem = new boost::interprocess::managed_shared_memory(
          boost::interprocess::open_or_create, shm_name.c_str(), S);
    });

    m_in_buffer = new shm_queue(shm_in_name.c_str(), m_shmem, shm_in_capacity);
    m_out_buffer =
        new shm_queue(shm_out_name.c_str(), m_shmem, shm_out_capacity);

  }
  virtual ~shm_service_base() {
    delete m_in_buffer;
    delete m_out_buffer;

    delete m_shmem;
    m_session->bookkeeper()->release(m_shm_name, [this]() {
      boost::interprocess::shared_memory_object::remove(m_shm_name.c_str());
    });
  }

  template <typename T, typename... ARGS> T *construct(ARGS... args) {
    if (!m_shmem) {
      throw shm_service_failure("no shared memory");
    }
    return m_shmem->construct<boost::interprocess::anonymous_instance>(args...);
  }

  template <typename T> void destroy(T *ptr) {
    if (!m_shmem) {
      throw shm_service_failure("no shared memory");
    }
    m_shmem->destroy_ptr<T>(ptr);
  }

  template <typename T> void push_back(T *ptr) {
    if (!ptr) {
      return;
    }
    m_out_buffer->push_back(ptr);
  }

  template <typename T, typename Func> void add_handler(Func &&f) {
    std::lock_guard<std::mutex> _l(m_handlers_mutex);
    m_all_handlers.insert(std::make_pair(T::pkg_identifier, [&f](void *p) {
      T *r = (T *)p;
      f(r);
    }));
  }

  template <typename T, typename Func> void add_def_handler(Func &&f) {
    std::lock_guard<std::mutex> _l(m_handlers_mutex);
    m_all_def_handlers.insert(std::make_pair(T::pkg_identifier, [&f](void *p) {
      T *r = (T *)p;
      f(r);
    }));
  }

  void run() {
    //! We can ignore *unlisten* in ~shm_service_base, since we already have it
    //! in the base class *quitable_thread*
    neb::core::command_queue::instance()
        .listen_command<neb::core::exit_command>([this]() {
          m_in_buffer->wake_up_if_empty();
        });
    start();
  }

  virtual void thread_func() {
    while (!m_exit_flag) {
      std::pair<void *, shm_type_id_t> r = m_in_buffer->pop_front();
      if (r.first) {
        typename decltype(m_all_handlers)::const_iterator fr =
            m_all_handlers.find(r.second);
        if (fr != m_all_handlers.end()) {
          fr->second(r.first);
        }
        m_shmem->destroy_ptr(r.first);
      }
    }
  }

  shm_session_base *session() { return m_session.get(); }

  // void wait_till_finish() {
  // if (!m_thread)
  // return;
  // m_thread->join();
  //}

private:
  std::string semaphore_name() const {
    return m_shm_name + std::string(".quit_semaphore");
  }

  std::string mutex_name() const { return m_shm_name + std::string(".mutex"); }

protected:
  typedef std::function<void(void *)> pkg_handler_t;

  std::string m_shm_name;
  std::string m_shm_in_name;
  std::string m_shm_out_name;
  boost::interprocess::managed_shared_memory *m_shmem;
  std::unordered_map<shm_type_id_t, pkg_handler_t> m_all_handlers;
  std::unordered_map<shm_type_id_t, pkg_handler_t> m_all_def_handlers;
  std::mutex m_handlers_mutex;
  shm_queue *m_in_buffer;
  shm_queue *m_out_buffer;
  std::unique_ptr<shm_session_base> m_session;
}; // end class shm_service_base
}

template <size_t S>
class shm_service_server : public internal::shm_service_base<S, shm_server> {
public:
  shm_service_server(const std::string &name, size_t in_obj_max_count,
                     size_t out_obj_max_count)
      : internal::shm_service_base<S, shm_server>(
            name,
            internal::shm_other_side_role<shm_server>::type::role_name(name),
            shm_server::role_name(name), in_obj_max_count, out_obj_max_count) {}

  void wait_until_client_start() {
    internal::shm_session_server *ss =
        (internal::shm_session_server *)
            internal::shm_service_base<S, shm_server>::session()
                .get();
    ss->wait_until_client_start();
  }
};

template <size_t S>
class shm_service_client : public internal::shm_service_base<S, shm_client> {
public:
  shm_service_client(const std::string &name, size_t in_obj_max_count,
                     size_t out_obj_max_count)
      : internal::shm_service_base<S, shm_client>(
            name,
            internal::shm_other_side_role<shm_client>::type::role_name(name),
            shm_client::role_name(name), in_obj_max_count, out_obj_max_count) {}
};
}
}
