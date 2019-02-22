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
#include "core/net_ipc/client/nipc_client.h"
#include "core/ipc_configuration.h"
#include "core/net_ipc/nipc_pkg.h"

namespace neb {
namespace core {

nipc_client::~nipc_client() { shutdown(); }

bool nipc_client::start() {

  if (m_handlers.empty()) {
    LOG(INFO) << "No handlers here";
    return false;
  }

  bool init_done = false;
  std::mutex local_mutex;
  std::condition_variable local_cond_var;
  std::atomic_bool got_exception_when_start_ipc;

  m_thread = std::unique_ptr<std::thread>(new std::thread([&, this]() {
    try {
      got_exception_when_start_ipc = false;

      ::ff::net::net_nervure nn;
      ::ff::net::typed_pkg_hub hub;
      std::for_each(
          m_handlers.begin(), m_handlers.end(),
          [&hub](const std::function<void(::ff::net::typed_pkg_hub &)> &f) {
            f(hub);
          });
      nn.get_event_handler()->listen<::ff::net::event::tcp_get_connection>(
          [&, this](::ff::net::tcp_connection_base *) {
            m_is_connected = true;
            local_mutex.lock();
            init_done = true;
            local_mutex.unlock();
            local_cond_var.notify_one();
          });
      nn.get_event_handler()->listen<::ff::net::event::tcp_lost_connection>(
          [this](::ff::net::tcp_connection_base *) { m_is_connected = false; });
      nn.add_pkg_hub(hub);
      m_conn =
          nn.add_tcp_client("127.0.0.1", ipc_configuration::instance().port());
      nn.run();

    } catch (const std::exception &e) {
      got_exception_when_start_ipc = true;
      LOG(ERROR) << "get exception when start ipc, " << typeid(e).name() << ", "
                 << e.what();
      local_cond_var.notify_one();
    } catch (...) {
      got_exception_when_start_ipc = true;
      LOG(ERROR) << "get unknown exception when start ipc";
      local_cond_var.notify_one();
    }
  }));
  std::unique_lock<std::mutex> _l(local_mutex);
  if (!init_done) {
    local_cond_var.wait(_l);
  }
  if (got_exception_when_start_ipc) {
    return false;
  }
  return true;
}

void nipc_client::shutdown() {
  m_conn->close();
  if (m_thread) {
    m_thread->join();
    m_thread.reset();
  }
}
} // namespace core
} // namespace neb
