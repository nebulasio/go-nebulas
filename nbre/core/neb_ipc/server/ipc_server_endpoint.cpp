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
#include "core/neb_ipc/server/ipc_server_endpoint.h"
#include "core/neb_ipc/ipc_pkg.h"
#include "fs/util.h"
#include <atomic>
#include <condition_variable>

namespace neb {
namespace core {
ipc_server_endpoint::ipc_server_endpoint(const std::string &root_dir,
                                         const std::string &nbre_exe_path)
    : m_root_dir(root_dir), m_nbre_exe_name(nbre_exe_path), m_client(nullptr){};

ipc_server_endpoint::~ipc_server_endpoint() {
  if (m_thread) {
    m_thread->join();
    m_thread.reset();
  }
}
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

      LOG(INFO) << "start ipc server ";
      m_ipc_server = std::unique_ptr<ipc_server_t>(
          new ipc_server_t(shm_service_name, 128, 128));

      m_ipc_server->init_local_env();
      m_request_timer =
          std::unique_ptr<api_request_timer>(new api_request_timer(
              m_ipc_server.get(), &ipc_callback_holder::instance()));

      LOG(INFO) << "nbre ipc init done!";
      add_all_callbacks();

      m_client_watcher = std::unique_ptr<ipc_client_watcher>(
          new ipc_client_watcher(m_nbre_exe_name));
      m_client_watcher->start();

      local_mutex.lock();
      init_done = true;
      local_mutex.unlock();
      local_cond_var.notify_one();

      m_ipc_server->run();

      LOG(INFO) << "ipc server stopped!";
    } catch (const std::exception &e) {
      LOG(ERROR) << "get exception when start nbre, " << typeid(e).name()
                 << ", " << e.what();
      m_got_exception_when_start_nbre = true;
      local_cond_var.notify_one();
    } catch (...) {
      LOG(ERROR) << "get unknown exception when start nbre";
      m_got_exception_when_start_nbre = true;
      local_cond_var.notify_one();
    }
  }));

  std::unique_lock<std::mutex> _l(local_mutex);
  if (!init_done) {
    local_cond_var.wait(_l);
  }
  if (m_got_exception_when_start_nbre)
    return false;
  try {
    m_ipc_server->wait_until_client_start();
  } catch (const std::exception &e) {
    LOG(ERROR) << "get exception when wait client start, " << typeid(e).name()
               << e.what();
    return false;
  } catch (...) {
    LOG(ERROR) << "get unknown exception when wait nbre";
    return false;
  }
  return true;
}

bool ipc_server_endpoint::check_path_exists() {
  return neb::fs::exists(m_nbre_exe_name);
}

void ipc_server_endpoint::init_params(const char *admin_pub_addr) {
  m_admin_pub_addr = admin_pub_addr;
}
void ipc_server_endpoint::add_all_callbacks() {
  LOG(INFO) << "ipc server pointer: " << (void *)m_ipc_server.get();
  ipc_server_t *p = m_ipc_server.get();

  m_callbacks = &(ipc_callback_holder::instance());
  m_ipc_server->add_handler<ipc_pkg::nbre_version_ack>(
      [p, this](ipc_pkg::nbre_version_ack *msg) {
        LOG(INFO) << "alloc: " << (void *)p;
        m_request_timer->remove_api(msg->m_holder);
        ipc_callback_holder::instance().m_nbre_version_callback(
            ipc_status_succ, msg->m_holder, msg->get<ipc_pkg::major>(),
            msg->get<ipc_pkg::minor>(), msg->get<ipc_pkg::patch>());
      });

  m_ipc_server->add_handler<ipc_pkg::nbre_init_req>(
      [p, this](ipc_pkg::nbre_init_req *) {
        LOG(INFO) << "get init req ";
        ipc_pkg::nbre_init_ack *ack = p->construct<ipc_pkg::nbre_init_ack>(
            nullptr, p->default_allocator());
        ack->set<ipc_pkg::admin_pub_addr>(m_admin_pub_addr.c_str());
        ack->set<ipc_pkg::nbre_root_dir>(m_root_dir.c_str());
        p->push_back(ack);
      });
}

int ipc_server_endpoint::send_nbre_version_req(void *holder, uint64_t height) {
  CHECK_NBRE_STATUS(m_callbacks->m_nbre_version_callback);

  m_request_timer->issue_api(
      holder,
      [holder, height, this]() {
        ipc_pkg::nbre_version_req *req =
            m_ipc_server->construct<ipc_pkg::nbre_version_req>(
                holder, m_ipc_server->default_allocator());
        if (req == nullptr) {
          return; // will call timeout later
        }
        req->set<ipc_pkg::height>(height);
        m_ipc_server->push_back(req);
      },
      m_callbacks->m_nbre_version_callback);
  return ipc_status_succ;
}

void ipc_server_endpoint::shutdown() {
  neb::core::command_queue::instance().send_command(
      std::make_shared<neb::core::exit_command>());
}
}// namespace core
} // namespace neb
