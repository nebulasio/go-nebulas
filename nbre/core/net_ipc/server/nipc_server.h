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
#include "core/net_ipc/nipc_common.h"
#include "core/net_ipc/nipc_pkg.h"
#include "core/net_ipc/server/api_request_timer.h"
#include "core/net_ipc/server/ipc_callback_holder.h"
#include "core/net_ipc/server/ipc_client_watcher.h"
namespace neb {
namespace core {

class nipc_server {
public:
  nipc_server();
  ~nipc_server();

  void init_params(const nbre_params_t &params);
  bool start();

  template <typename PkgType>
  int send_api_pkg(void *holder, const std::shared_ptr<PkgType> &pkg) {
    if (m_request_timer == nullptr) {
      return ipc_status_fail;
    }

    m_request_timer->issue_api(
        reinterpret_cast<uint64_t>(holder),
        [this, pkg](::ff::net::tcp_connection_base_ptr conn) {
          if (conn) {
            conn->send(pkg);
          };
        },
        ipc_callback_holder::instance().get_callback(
            typename get_pkg_ack_type<PkgType>::type().type_id()));
    return ipc_status_succ;
  }

  void shutdown();

protected:
  template <typename PkgType> void to_recv_pkg() {
    m_pkg_hub->to_recv_pkg<PkgType>([this](std::shared_ptr<PkgType> pkg) {
      m_last_heart_beat_time = std::chrono::steady_clock::now();
      ipc_callback_holder::instance().call_callback<PkgType>(pkg);
      m_request_timer->remove_api(pkg->template get<p_holder>());
    });
  }

private:
  bool check_path_exists();
  void add_all_callbacks();

protected:
  std::unique_ptr<::ff::net::net_nervure> m_server;
  ::ff::net::tcp_connection_base_ptr m_conn;
  std::unique_ptr<std::thread> m_thread;

  std::unique_ptr<::ff::net::typed_pkg_hub> m_pkg_hub;
  ipc_callback_holder *m_callbacks;
  std::atomic_bool m_got_exception_when_start_ipc;
  std::unique_ptr<api_request_timer> m_request_timer;
  std::condition_variable m_start_complete_cond_var;
  std::mutex m_mutex;
  bool m_is_started;
  std::unique_ptr<ipc_client_watcher> m_client_watcher;
  std::chrono::steady_clock::time_point m_last_heart_beat_time;
  std::unique_ptr<util::timer_loop> m_heart_beat_watcher;
};

} // namespace core
} // namespace neb
