// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library. //
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
#include "core/neb_ipc/server/ipc_server_endpoint.h"
#include "core/neb_ipc/ipc_pkg.h"
#include "fs/util.h"
#include <atomic>
#include <boost/process/child.hpp>
#include <condition_variable>

namespace neb {
namespace core {
ipc_server_endpoint::ipc_server_endpoint(const std::string &root_dir,
                                         const std::string &nbre_exe_path)
    : m_root_dir(root_dir), m_nbre_exe_name(nbre_exe_path), m_client(nullptr){};

bool ipc_server_endpoint::start() {
  if (!check_path_exists()) {
    LOG(ERROR) << "nbre path not exist";
    return false;
  }
  if (!ipc_callback_holder::instance().check_all_callbacks()) {
    LOG(ERROR) << "nbre missing callback";
    return false;
  }

  std::atomic_bool init_done(false);
  std::mutex local_mutex;
  std::condition_variable local_cond_var;
  m_got_exception_when_start_nbre = false;
  m_thread = std::unique_ptr<std::thread>(new std::thread([&, this]() {
    try {
      {
        ipc_util_t us(shm_service_name, 128, 128);
        us.reset();
      }

      this->m_ipc_server = std::unique_ptr<ipc_server_t>(
          new ipc_server_t(shm_service_name, 128, 128));

      m_ipc_server->init_local_env();
      LOG(INFO) << "nbre ipc init done!";
      add_all_callbacks();
      boost::process::child client(m_nbre_exe_name);

      local_mutex.lock();
      init_done = true;
      local_mutex.unlock();
      local_cond_var.notify_one();

      m_ipc_server->run();

      client.wait();

      LOG(INFO) << "nbre stopped!";
    } catch (const std::exception &e) {
      LOG(ERROR) << "get exception when start nbre, " << typeid(e).name()
                 << ", " << e.what();
      m_got_exception_when_start_nbre = true;
    } catch (...) {
      LOG(ERROR) << "get unknown exception when start nbre";
      m_got_exception_when_start_nbre = true;
    }
  }));

  std::unique_lock<std::mutex> _l(local_mutex);
  if (!init_done) {
    local_cond_var.wait(_l);
  }
  if (m_got_exception_when_start_nbre)
    return false;
  else
    return true;
}
bool ipc_server_endpoint::check_path_exists() {
  return neb::fs::exists(m_nbre_exe_name);
}

void ipc_server_endpoint::add_all_callbacks() {
  LOG(INFO) << "ipc server pointer: " << (void *)m_ipc_server.get();
  m_ipc_server->add_handler<ipc_pkg::nbre_version_ack>(
      [](ipc_pkg::nbre_version_ack *msg) {
        ipc_callback_holder::instance().m_nbre_version_callback(
            msg->m_holder, msg->get<ipc_pkg::major>(),
            msg->get<ipc_pkg::minor>(), msg->get<ipc_pkg::patch>());
      });
}

void ipc_server_endpoint::send_nbre_version_req(void *holder, uint64_t height) {
  ipc_pkg::nbre_version_req *req =
      m_ipc_server->construct<ipc_pkg::nbre_version_req>(
          holder, m_ipc_server->default_allocator());
  req->set<ipc_pkg::height>(height);
  m_ipc_server->push_back(req);
}

void ipc_server_endpoint::shutdown() {
  LOG(INFO) << "shutdown session";
  m_ipc_server->session()->stop();

  LOG(INFO) << "shutdown server";
  m_ipc_server->stop();

}
}// namespace core
} // namespace neb
