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
#include "common/ipc/shm_session.h"
#include "common/common.h"
#include "common/exception_queue.h"
#include <chrono>

namespace neb {
namespace ipc {
const static char *bookkeeper_mem_name = "io.nebulas.nbre.sessions";

void clean_shm_session_env() { clean_bookkeeper_env(bookkeeper_mem_name); }
namespace internal {
size_t max_wait_fail_times = 8;

shm_session_base::shm_session_base(const std::string &name)
    : quitable_thread(), m_name(name) {
  m_bookkeeper =
      std::unique_ptr<shm_bookkeeper>(new shm_bookkeeper(bookkeeper_mem_name));
  m_server_sema = m_bookkeeper->acquire_named_semaphore(server_sema_name());
  m_client_sema = m_bookkeeper->acquire_named_semaphore(client_sema_name());
}

shm_session_base::~shm_session_base() {
  m_bookkeeper->release_named_semaphore(server_sema_name());
  m_bookkeeper->release_named_semaphore(client_sema_name());
}

void shm_session_base::start_session() { start(); }

shm_session_server::shm_session_server(const std::string &name)
    : shm_session_base(name), m_client_started(false), m_client_alive(false) {}

void shm_session_server::wait_until_client_start() {
  if (m_client_started)
    return;
  m_client_sema->wait();
  m_client_started = true;
}

bool shm_session_server::is_client_alive() { return m_client_alive; }

void shm_session_server::thread_func() {
  uint32_t fail_counter = 0;
  while (!m_exit_flag) {
    if (!m_client_started) {
      bool ret = m_client_sema->try_wait();
      if (ret) {
        m_client_started = true;
        m_client_alive = true;
      }
    } else {
      bool ret = m_client_sema->try_wait();
      if (ret) {
        fail_counter = 0;
        m_client_alive = true;
      } else {
        fail_counter++;
        if (fail_counter >= max_wait_fail_times) {
          m_client_alive = false;
          throw shm_session_timeout();
        }
      }
    }
    std::this_thread::sleep_for(std::chrono::seconds(1));
    m_server_sema->post();
  }
}

shm_session_client::shm_session_client(const std::string &name)
    : shm_session_base(name), m_server_alive(false) {}

void shm_session_client::thread_func() {
  uint32_t fail_counter = 0;
  while (!m_exit_flag) {
    bool ret = m_server_sema->try_wait();
    if (ret) {
      fail_counter = 0;
      m_server_alive = true;
      } else {
        fail_counter++;
        if (fail_counter >= max_wait_fail_times) {
          m_server_alive = false;
          throw shm_session_timeout();
        }
      }
    std::this_thread::sleep_for(std::chrono::seconds(1));
    m_client_sema->post();
  }
}

bool shm_session_client::is_server_alive() { return m_server_alive; }
}
}
}
