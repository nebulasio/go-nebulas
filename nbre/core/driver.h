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
#include "common/timer_loop.h"
#include "core/neb_ipc/client/ipc_client_endpoint.h"
#include "core/neb_ipc/ipc_pkg.h"
#include "fs/util.h"
#include <atomic>

namespace neb {
namespace core {
namespace internal {
class driver_base {

public:
  driver_base();

  virtual bool init();

  virtual void run();

  ipc_client_t *ipc_conn() { return m_ipc_conn; }

protected:
  virtual void add_handlers() = 0;

  void handle_exception(const std::shared_ptr<neb::neb_exception> &p);

  void init_timer_thread();

  void init_nbre();

protected:
  std::unique_ptr<ipc_client_endpoint> m_client;
  ipc_client_t *m_ipc_conn;
  std::unique_ptr<std::thread> m_client_thread;
  std::atomic_bool m_exit_flag;
  std::unique_ptr<std::thread> m_timer_thread;
  std::unique_ptr<timer_loop> m_timer_loop;
};
} // end namespace internal

class driver : public internal::driver_base {
public:
  driver();

protected:
  virtual void add_handlers();
};
}
}
