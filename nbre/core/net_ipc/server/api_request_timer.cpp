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
#include "core/net_ipc/server/api_request_timer.h"
#include <unordered_set>

namespace neb {
namespace core {
api_request_timer::api_request_timer(boost::asio::io_service *io_service,
                                     ipc_callback_holder *holder)
    : m_timout_threshold(30), m_cb_holder(holder) {
  m_tl = std::make_unique<util::timer_loop>(io_service);
  m_tl->register_timer_and_callback(1, [this]() {
    timer_callback();
  });
}

api_request_timer::~api_request_timer() {}

void api_request_timer::remove_api(api_identifier_t id) {
  std::lock_guard<std::mutex> _l(m_mutex);
  m_api_requests.erase(id);
  m_api_timeout_counter.erase(id);
  m_api_timeout_callbacks.erase(id);
}

bool api_request_timer::is_api_alive(api_identifier_t id) const {
  std::lock_guard<std::mutex> _l(m_mutex);
  return m_api_requests.find(id) != m_api_requests.end();
}
void api_request_timer::timer_callback() {

  std::vector<api_identifier_t> timeout_apis;
  std::vector<api_identifier_t> to_retry_apis;
  std::unique_lock<std::mutex> _l(m_mutex);
  for (auto &p : m_api_timeout_counter) {
    p.second--;
    if (p.second == 0) {
      timeout_apis.push_back(p.first);
    } else if (p.second % 10 == 0) {
      to_retry_apis.push_back(p.first);
    }
  }
  _l.unlock();

  for (auto &id : timeout_apis) {
    auto it = m_api_timeout_callbacks.find(id);
    if (it != m_api_timeout_callbacks.end()) {
      std::unique_lock<std::mutex> _k(m_conn_mutex);
      it->second();
    }
  }

  _l.lock();
  for (auto &id : timeout_apis) {
    m_api_timeout_counter.erase(id);
    m_api_timeout_callbacks.erase(id);
    m_api_requests.erase(id);
  }
  _l.unlock();

  for (auto id : to_retry_apis) {
    _l.lock();
    auto ret = m_api_requests.find(id);
    if (ret == m_api_requests.end()) {
      _l.unlock();
      continue;
    } else {
      _l.unlock();
      std::unique_lock<std::mutex> _k(m_conn_mutex);
      if (m_conn)
        ret->second(m_conn);
    }
  }
}
} // namespace core
} // namespace neb
