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
#include "core/net_ipc/server/nipc_server.h"
#include "common/configuration.h"
#include "core/net_ipc/nipc_pkg.h"
#include "fs/util.h"
#include <boost/format.hpp>

namespace neb {
namespace core {
nipc_server::nipc_server() : m_server(nullptr), m_conn(nullptr) {}
nipc_server::~nipc_server() {
  LOG(INFO) << "to destroy nipc server";
  if (m_thread) {
    m_thread->join();
  }
}

void nipc_server::init_params(const nbre_params_t &params) {
  neb::configuration::instance().nbre_root_dir() = params.m_nbre_root_dir;
  neb::configuration::instance().nbre_exe_name() = params.m_nbre_exe_name;
  neb::configuration::instance().neb_db_dir() = params.m_neb_db_dir;
  neb::configuration::instance().nbre_db_dir() = params.m_nbre_db_dir;
  neb::configuration::instance().nbre_log_dir() = params.m_nbre_log_dir;
  neb::configuration::instance().admin_pub_addr() =
      base58_to_address(params.m_admin_pub_addr);
  neb::configuration::instance().nbre_start_height() =
      params.m_nbre_start_height;
  neb::configuration::instance().nipc_listen() = params.m_nipc_listen;
  neb::configuration::instance().nipc_port() = params.m_nipc_port;

#ifdef NDEBUG
  // supervisor start failed with reading file
#else
  // read errno_list file
  {
    std::string errno_file =
        neb::fs::join_path(neb::configuration::instance().nbre_root_dir(),
                           std::string("util/errno_list"));
    std::ifstream ifs;
    ifs.open(errno_file.c_str(), std::ios::in | std::ios::binary);
    if (!ifs.is_open()) {
      throw std::invalid_argument(
          boost::str(boost::format("can't open file %1%") % errno_file));
    }
    std::string line;
    while (std::getline(ifs, line) && !line.empty()) {
      neb::configuration::instance().set_exit_msg(line);
    }
    ifs.close();
  }
#endif
}

bool nipc_server::start() {
  m_mutex.lock();
  m_is_started = false;
  m_mutex.unlock();

  m_got_exception_when_start_ipc = false;
  m_thread = std::make_unique<std::thread>([&, this] {
    try {
      m_server = std::make_unique<::ff::net::net_nervure>();
      m_pkg_hub = std::make_unique<::ff::net::typed_pkg_hub>();
      add_all_callbacks();
      m_server->add_pkg_hub(*m_pkg_hub);
      m_server->add_tcp_server(neb::configuration::instance().nipc_listen(),
                               neb::configuration::instance().nipc_port());
      m_request_timer = std::make_unique<api_request_timer>(
          &m_server->ioservice(), &ipc_callback_holder::instance());

      m_server->get_event_handler()
          ->listen<::ff::net::event::more::tcp_server_accept_connection>(
              [this](::ff::net::tcp_connection_base_ptr conn) {
                LOG(INFO) << "got connection";
                m_conn = conn;
                m_mutex.lock();
                m_is_started = true;
                m_mutex.unlock();
                m_request_timer->reset_conn(conn);
                m_last_heart_beat_time = std::chrono::steady_clock::now();
                m_start_complete_cond_var.notify_one();
              });
      m_server->get_event_handler()
          ->listen<::ff::net::event::tcp_lost_connection>(
              [this](::ff::net::tcp_connection_base *) {
                // We may restart the client. But we can ignore this and leave
                // this to ipc_client_watcher
                LOG(INFO) << "lost connection";
                m_conn.reset();
                m_request_timer->reset_conn(nullptr);
              });

      m_server->get_event_handler()
          ->listen<::ff::net::event::more::tcp_recv_stream_succ>(
              [this](::ff::net::tcp_connection_base *, size_t) {
                m_last_heart_beat_time = std::chrono::steady_clock::now();
              });
      m_server->get_event_handler()
          ->listen<::ff::net::event::more::tcp_send_stream_succ>(
              [this](::ff::net::tcp_connection_base *, size_t) {
                m_last_heart_beat_time = std::chrono::steady_clock::now();
              });

      // We need start the client here
      m_client_watcher =
          std::unique_ptr<ipc_client_watcher>(new ipc_client_watcher(
              neb::configuration::instance().nbre_exe_name()));
      m_client_watcher->start();

      m_heart_beat_watcher =
          std::make_unique<util::timer_loop>(&m_server->ioservice());
      m_heart_beat_watcher->register_timer_and_callback(1, [this]() {
        auto now = std::chrono::steady_clock::now();
        auto duration = now - m_last_heart_beat_time;
        auto count =
            std::chrono::duration_cast<std::chrono::seconds>(duration).count();
        if (count > 60) {
          LOG(INFO) << "lost heart beat, to kill client";
          m_client_watcher->kill_client();
        }
      });
      while (true) {
        if (m_server->ioservice().stopped()) {
          LOG(INFO) << "ioservice already stopped, wait to restart";
          break;
        }
        try {
          m_server->run();
        } catch (...) {
          LOG(INFO) << "to reset ioservice";
          m_server->ioservice().reset();
        }
      }
      LOG(INFO) << "ioservice quit";
    } catch (const std::exception &e) {
      m_got_exception_when_start_ipc = true;
      LOG(ERROR) << "get exception when start ipc, " << typeid(e).name() << ", "
                 << e.what();
      m_start_complete_cond_var.notify_one();
    } catch (...) {
      m_got_exception_when_start_ipc = true;
      LOG(ERROR) << "get unknown exception when start ipc";
      m_start_complete_cond_var.notify_one();
    }
  });
  std::unique_lock<std::mutex> _l(m_mutex);
  if (!m_is_started && !m_got_exception_when_start_ipc) {
    LOG(INFO) << "wait to start complete cond var";
    m_start_complete_cond_var.wait(_l);
  }
  if (m_got_exception_when_start_ipc) {
    LOG(INFO) << "got exception when server start ipc";
    return false;
  }

  return true;
}

void nipc_server::shutdown() {
  LOG(INFO) << "to shutdown nipc server";
  if (m_conn)
    m_conn->close();
  m_server->stop();
  LOG(INFO) << "nipc server send exit command";
  neb::core::command_queue::instance().send_command(
      std::make_shared<neb::core::exit_command>());
}

void nipc_server::add_all_callbacks() {
#define define_ipc_param(type, name)
#define define_ipc_pkg(type, ...)
#define define_ipc_api(req, ack) to_recv_pkg<ack>();

#include "core/net_ipc/ipc_interface_impl.h"

#undef define_ipc_api
#undef define_ipc_pkg
#undef define_ipc_param

  m_pkg_hub->to_recv_pkg<nbre_init_req>([this](std::shared_ptr<nbre_init_req>) {
    std::shared_ptr<nbre_init_ack> ack = std::make_shared<nbre_init_ack>();
    LOG(INFO) << "recv nbre_init_req";
    configuration &conf = configuration::instance();
    ack->set<p_nbre_root_dir>(conf.nbre_root_dir());
    ack->set<p_nbre_exe_name>(conf.nbre_exe_name());
    ack->set<p_neb_db_dir>(conf.neb_db_dir());
    ack->set<p_nbre_db_dir>(conf.nbre_db_dir());
    ack->set<p_nbre_log_dir>(conf.nbre_log_dir());
    ack->set<p_admin_pub_addr>(std::to_string(conf.admin_pub_addr()));
    ack->set<p_nbre_start_height>(conf.nbre_start_height());
    m_conn->send(ack);
    LOG(INFO) << "send nbre_init_ack";
  });

  m_pkg_hub->to_recv_pkg<heart_beat_t>([this](std::shared_ptr<heart_beat_t> p) {
    LOG(INFO) << "got heart beat";
    m_last_heart_beat_time = std::chrono::steady_clock::now();
    m_conn->send(p);
  });
}
} // namespace core
} // namespace neb
