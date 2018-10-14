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
#include "core/neb_ipc/ipc_endpoint.h"
#include "core/neb_ipc/ipc_pkg.h"
#include "fs/util.h"
#include <atomic>
#include <boost/process/child.hpp>
#include <condition_variable>

namespace neb {
namespace core {
ipc_endpoint::ipc_endpoint(const std::string &root_dir,
                           const std::string &nbre_exe_path)
    : m_root_dir(root_dir), m_nbre_exe_name(nbre_exe_path){};

bool ipc_endpoint::start() {
  if (!check_path_exists()) {
    return false;
  }
  if (!ipc_callback_holder::instance().check_all_callbacks()) {
    return false;
  }

  bool init_done = false;
  std::mutex local_mutex;
  std::condition_variable local_cond_var;
  m_thread = std::unique_ptr<std::thread>(new std::thread([&, this]() {
    try {
      this->m_ipc_server = std::unique_ptr<ipc_server_t>(
          new ipc_server_t(shm_service_name, 128, 128));

      this->m_ipc_server->reset();

      m_ipc_server->init_local_env();
      add_all_callbacks();
      boost::process::child client(m_nbre_exe_name);

      local_mutex.lock();
      init_done = true;
      local_mutex.unlock();
      local_cond_var.notify_one();
      m_ipc_server->run();
      client.wait();
    } catch (const std::exception &e) {
    }
  }));

  std::unique_lock<std::mutex> _l(local_mutex);
  if (!init_done) {
    local_cond_var.wait(_l);
  }

  return true;
}

void ipc_endpoint::add_all_callbacks() {
  LOG(INFO) << "ipc server pointer: " << (void *)m_ipc_server.get();
  m_ipc_server->add_handler<nbre_version_ack>([](nbre_version_ack *msg) {
    ipc_callback_holder::instance().m_nbre_version_callback(
        msg->m_holder, msg->m_major, msg->m_minor, msg->m_patch);
  });
}

void ipc_endpoint::send_nbre_version_req(void *holder, uint64_t height) {
  nbre_version_req *req = m_ipc_server->construct<nbre_version_req>();
  req->m_height = height;
  req->m_holder = holder;
  m_ipc_server->push_back(req);
}
bool ipc_endpoint::check_path_exists() {
  return neb::fs::exists(m_nbre_exe_name);
}
void ipc_endpoint::shutdown() {}

} // namespace core
} // namespace neb
