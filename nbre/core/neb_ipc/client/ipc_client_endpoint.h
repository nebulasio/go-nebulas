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
#include "common/ipc/shm_service.h"
#include "core/neb_ipc/ipc_common.h"

namespace neb {
namespace core {
class ipc_client_endpoint {
public:
  ipc_client_endpoint() = default;
  ~ipc_client_endpoint();

  template <typename T, typename Func> void add_handler(Func &&f) {
    m_handlers.push_back([this, f]() { m_client->add_handler<T>(f); });
  }

  bool start();

  void shutdown();

  inline ipc_client_t *ipc_connection() { return m_client.get(); }

protected:
  std::vector<std::function<void()>> m_handlers;
  std::unique_ptr<std::thread> m_thread;
  std::unique_ptr<ipc_client_t> m_client;
};
} // namespace core
} // namespace neb
