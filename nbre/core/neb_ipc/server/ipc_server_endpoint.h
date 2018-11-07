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
#include "core/neb_ipc/server/api_request_timer.h"
#include "core/neb_ipc/server/ipc_callback_holder.h"
#include "core/neb_ipc/server/ipc_client_watcher.h"

namespace neb {
namespace core {

class ipc_server_endpoint {
public:
  ipc_server_endpoint(const std::string &root_dir,
                      const std::string &nbre_exe_path);

  ~ipc_server_endpoint();

  void init_params(const char *admin_pub_addr);
  bool start();

  int send_nbre_version_req(void *holder, uint64_t height);
  int send_nbre_ir_list_req(void *holder);

  void shutdown();

private:
  bool check_path_exists();
  void add_all_callbacks();

protected:
  std::string m_root_dir;
  std::string m_nbre_exe_name;
  std::unique_ptr<std::thread> m_thread;
  std::unique_ptr<ipc_server_t> m_ipc_server;
  boost::process::child *m_client;
  std::atomic_bool m_got_exception_when_start_nbre;

  std::string m_admin_pub_addr;
  std::unique_ptr<std::thread> m_timer_thread;
  std::unique_ptr<api_request_timer> m_request_timer;
  std::unique_ptr<ipc_client_watcher> m_client_watcher;
  ipc_callback_holder *m_callbacks;
};

} // namespace core
} // namespace neb
