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
#include "common/timer_loop.h"
#include "core/neb_ipc/ipc_common.h"
#include "core/neb_ipc/server/ipc_callback_holder.h"
#include <boost/asio.hpp>
#include <boost/asio/io_service.hpp>

namespace neb {
namespace core {
class api_request_timer {
public:
  typedef void *api_identifier_t;
  api_request_timer(ipc_server_t *ipc, ipc_callback_holder *holder);

  ~api_request_timer();

  template <typename FT, typename Func>
  void issue_api(api_identifier_t id, FT &&f, Func &&func) {
    std::unique_lock<std::mutex> _l(m_mutex);
    m_api_requests.insert(std::make_pair(id, f));
    m_api_timeout_counter.insert(std::make_pair(id, m_timout_threshold));
    m_api_timeout_callbacks.insert(std::make_pair(id, [func, this]() {
      m_ipc->schedule_task_in_service_thread(
          [func]() { issue_callback_with_error(func, ipc_status_timeout); });
    }));
    f();
  }

  void remove_api(api_identifier_t id);

  bool is_api_alive(api_identifier_t id) const;

protected:
  void timer_callback();

protected:
  typedef std::unordered_map<api_identifier_t, int32_t> timeout_counter_t;
  typedef std::unordered_map<api_identifier_t, std::function<void()>>
      api_requests_t;

  int32_t m_timout_threshold;
  mutable std::mutex m_mutex;
  timeout_counter_t m_api_timeout_counter;
  api_requests_t m_api_requests;
  api_requests_t m_api_timeout_callbacks;
  ipc_server_t *m_ipc;
  ipc_callback_holder *m_cb_holder;
  std::unique_ptr<std::thread> m_thread;
}; // end class api_request_timer

} // namespace core
} // namespace neb
