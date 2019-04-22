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
#include "core/net_ipc/server/ipc_client_watcher.h"
#include "common/configuration.h"
#include "fs/util.h"
#include <boost/process/args.hpp>
#include <boost/process/io.hpp>
#include <thread>

namespace neb {
namespace core {
ipc_client_watcher::ipc_client_watcher(const std::string &path)
    : m_path(path), m_restart_times(0), m_last_start_time(),
      m_b_client_alive(false) {}

void ipc_client_watcher::thread_func() {
  std::chrono::seconds threshold(8);
  while (!m_exit_flag) {
    std::chrono::system_clock::time_point now =
        std::chrono::system_clock::now();

    std::chrono::seconds duration =
        std::chrono::duration_cast<std::chrono::seconds>(now -
                                                         m_last_start_time);

    std::chrono::seconds delta = threshold - duration;
    if (delta >= std::chrono::seconds(0)) {
      std::this_thread::sleep_for(delta + std::chrono::seconds(1));
      if (m_exit_flag) {
        LOG(INFO) << "recv exit flag, to exit thread func";
        return;
      }
    }

    LOG(INFO) << "to start nbre : " << m_path;
    m_last_start_time = now;
    boost::process::ipstream stream;
    std::vector<std::string> v(
        {"--ipc-ip", configuration::instance().nipc_listen(), "--ipc-port",
         std::to_string(configuration::instance().nipc_port()), "--log-dir",
         configuration::instance().nbre_log_dir()});

    LOG(INFO) << "log-to-stderr: " << neb::glog_log_to_stderr;
    if (glog_log_to_stderr) {
      v.push_back("--log-to-stderr");
    }
    if (use_test_blockchain) {
      v.push_back("--use-test-blockchain");
    }

    {
      std::unique_lock<std::mutex> _l(m_mutex);
      m_client = std::make_unique<boost::process::child>(
          m_path, boost::process::args(v));
      if (m_client->valid()) {
        LOG(INFO) << "child process is valid";
        m_b_client_alive = true;
        m_killed_already = false;
      }
    }

    std::error_code ec;
    m_client->wait(ec);
    if (ec) {
      LOG(ERROR) << ec.message();
    }
    if (m_client->exit_code() != 0) {
      uint32_t exit_code = m_client->exit_code();
      LOG(ERROR) << "nbre abnormal quit, exit code: " << exit_code;
    }
    m_b_client_alive = false;
  }
}

ipc_client_watcher::~ipc_client_watcher() {
  LOG(INFO) << "to destroy ipc client watcher";
  if (m_thread) {
    m_thread->join();
    m_thread.reset();
  }
}
void ipc_client_watcher::kill_client() {
  std::unique_lock<std::mutex> _l(m_mutex);
  LOG(INFO) << "to kill client";
  if (m_killed_already) {
    LOG(INFO) << "already killed";
    return;
  }

  if (m_client) {
    LOG(INFO) << "to terminate client";
    m_killed_already = true;
    m_client->terminate();
    LOG(INFO) << "client is down";
  } else {
    LOG(WARNING) << "no client to kill";
  }
}
} // namespace core
} // namespace neb
