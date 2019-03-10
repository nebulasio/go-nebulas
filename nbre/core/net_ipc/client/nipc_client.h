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
#include "core/net_ipc/nipc_common.h"
#include "util/timer_loop.h"
#include <ff/functionflow.h>
#include <ff/network.h>

namespace neb {
namespace core {
class nipc_client {
public:
  nipc_client() = default;
  ~nipc_client();

  //! The handler f will run in a thread pool.
  template <typename T, typename Func> void add_handler(Func &&f) {
    m_handlers.push_back([this, f](::ff::net::typed_pkg_hub &hub) {
      hub.to_recv_pkg<T>([f](std::shared_ptr<T> pkg) {
        ff::para<> p;
        p([pkg, f]() { f(pkg); });
      });
    });
  }

  bool start();

  void shutdown();

  inline ::ff::net::tcp_connection_base_ptr connection() { return m_conn; }

protected:
  std::vector<std::function<void(::ff::net::typed_pkg_hub &)>> m_handlers;
  std::unique_ptr<std::thread> m_thread;
  std::unique_ptr<util::timer_loop> m_heart_bear_timer;
  ::ff::net::tcp_connection_base_ptr m_conn;
  std::atomic_bool m_is_connected;
  int32_t m_to_recv_heart_beat_msg;
  // std::unique_ptr<ipc_client_t> m_client;
};
} // namespace core
} // namespace neb
