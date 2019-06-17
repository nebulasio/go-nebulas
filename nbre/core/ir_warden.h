// Copyright (C) 2017 go-nebulas authors
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
#include "core/net_ipc/nipc_pkg.h"
#include "util/quitable_thread.h"
#include "util/singleton.h"
#include <boost/asio.hpp>
#include <boost/asio/deadline_timer.hpp>
#include <condition_variable>
#include <thread>

namespace neb {
class execution_context;
namespace core {
class ir_warden : public util::singleton<ir_warden> {
public:
  ir_warden();
  virtual ~ir_warden();

  void init_with_context(execution_context *c);

  bool is_sync_already() const;
  void wait_until_sync();
  void on_timer();

  void on_receive_ir_transactions(
      const std::shared_ptr<nbre_ir_transactions_req> &txs_ptr);

private:
  bool m_is_sync_already;
  mutable std::mutex m_sync_mutex;
  std::condition_variable m_sync_cond_var;
  util::wakeable_queue<std::shared_ptr<nbre_ir_transactions_req>> m_queue;
};
}
}
