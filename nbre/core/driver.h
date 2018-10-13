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
#include "core/neb_ipc/ipc_client.h"
#include "core/neb_ipc/ipc_pkg.h"
#include "fs/util.h"
#include <atomic>

namespace neb {
namespace core {
class driver {
public:
  driver();

  bool init();

  void run();

  ipc_client_t *ipc_conn() { return m_ipc_conn; }

private:
  void add_handlers();

  void handle_exception(const std::shared_ptr<neb::neb_exception> &p);

protected:
  std::unique_ptr<ipc_client> m_client;
  ipc_client_t *m_ipc_conn;
  std::unique_ptr<std::thread> m_client_thread;
  std::atomic_bool m_exit_flag;
};
}
}
