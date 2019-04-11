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
#include "core/neb_ipc/client/ipc_client_endpoint.h"
#include "core/neb_ipc/ipc_pkg.h"

namespace neb {
namespace core {

ipc_client_endpoint::~ipc_client_endpoint() { shutdown(); }

bool ipc_client_endpoint::start() {
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
      auto shm_service_name_str =
          std::string("nbre.") +
          neb::shm_configuration::instance().shm_name_identity();
      auto shm_service_name = shm_service_name_str.c_str();
      this->m_client = std::unique_ptr<ipc_client_t>(
          new ipc_client_t(shm_service_name, 128, 128));

      m_client->init_local_env();
      std::for_each(m_handlers.begin(), m_handlers.end(),
                    [](const std::function<void()> &f) { f(); });

      local_mutex.lock();
      init_done = true;
      local_mutex.unlock();
      local_cond_var.notify_one();

      m_client->run();

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

void ipc_client_endpoint::shutdown() {
  if (m_thread) {
    m_thread->join();
    m_thread.reset();
  }
  m_client.reset();
}
}
}
