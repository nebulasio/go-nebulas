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
#include "common/configuration.h"
#include "common/exception_queue.h"
#include "core/net_ipc/client/nipc_client.h"
#include "core/net_ipc/nipc_pkg.h"
#include "fs/util.h"
#include "util/timer_loop.h"

namespace neb {
namespace core {
class client_context;
namespace internal {
class client_driver_base {

public:
  client_driver_base();
  virtual ~client_driver_base();

  virtual bool init();

  virtual void run();

protected:
  virtual void add_handlers() = 0;

  void handle_exception(const std::shared_ptr<neb::neb_exception> &p);

  void init_timer_thread();

  void init_nbre();

protected:
  std::unique_ptr<nipc_client> m_client;
  ::ff::net::tcp_connection_base_ptr m_ipc_conn;
  std::atomic_bool m_exit_flag;
  std::unique_ptr<std::thread> m_timer_thread;
  std::unique_ptr<util::timer_loop> m_timer_loop;
  client_context *m_context;
};
} // end namespace internal
} // namespace core
} // namespace neb
