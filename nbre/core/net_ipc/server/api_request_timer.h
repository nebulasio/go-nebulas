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
#include "core/net_ipc/nipc_common.h"
#include "core/net_ipc/server/ipc_callback_holder.h"
#include "util/timer_loop.h"
#include <boost/asio.hpp>
#include <boost/asio/io_service.hpp>

namespace neb {
namespace core {
class api_request_timer {
public:
  typedef uint64_t api_identifier_t;
  api_request_timer(boost::asio::io_service *io_service,
                    ipc_callback_holder *holder);

  ~api_request_timer();

  template <typename FT, typename Func>
  void issue_api(api_identifier_t id, FT &&f, Func &&func) {
    std::unique_lock<std::mutex> _l(m_mutex);
    m_api_requests.insert(std::make_pair(id, f));
    m_api_timeout_counter.insert(std::make_pair(id, m_timout_threshold));
    m_api_timeout_callbacks.insert(std::make_pair(id, [func, this]() {
      LOG(INFO) << "ipc api timeout";
      func(ipc_status_timeout, nullptr);
      LOG(INFO) << "ipc api timeout done";
      // issue_callback_with_error(func, ipc_status_timeout);
    }));
    std::unique_lock<std::mutex> _k(m_conn_mutex);
    if (m_conn)
      f(m_conn);
  }

  void remove_api(api_identifier_t id);

  bool is_api_alive(api_identifier_t id) const;

  inline void reset_conn(::ff::net::tcp_connection_base_ptr conn) {
    std::unique_lock<std::mutex> _l(m_conn_mutex);
    m_conn = conn;
  }

protected:
  void timer_callback();

protected:
  typedef std::unordered_map<api_identifier_t, int32_t> timeout_counter_t;
  typedef std::unordered_map<
      api_identifier_t, std::function<void(::ff::net::tcp_connection_base_ptr)>>
      api_requests_t;
  typedef std::unordered_map<api_identifier_t, std::function<void()>>
      timeout_api_requests_t;

  ::ff::net::tcp_connection_base_ptr m_conn;
  std::mutex m_conn_mutex;
  int32_t m_timout_threshold;
  mutable std::mutex m_mutex;
  timeout_counter_t m_api_timeout_counter;
  api_requests_t m_api_requests;
  timeout_api_requests_t m_api_timeout_callbacks;
  // boost::asio::io_service *m_io_service;
  ipc_callback_holder *m_cb_holder;
  std::unique_ptr<util::timer_loop> m_tl;
}; // end class api_request_timer

} // namespace core
} // namespace neb
