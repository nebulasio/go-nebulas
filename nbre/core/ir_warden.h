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
#include "common/util/singleton.h"
#include "fs/proto/ir.pb.h"
#include "fs/storage.h"
#include <boost/asio.hpp>
#include <boost/asio/deadline_timer.hpp>
#include <thread>

namespace neb {
namespace core {
class ir_warden : public util::singleton<ir_warden> {
public:
  ir_warden();
  virtual ~ir_warden();

  void async_run();

  std::vector<std::shared_ptr<nbre::NBREIR>>
  get_ir_from_height(const std::string &name, block_height_t height);

  bool is_sync_already() const;

  void wait_until_sync();

protected:
  void timer_callback(const boost::system::error_code &ec);

  void on_timer();

private:
  std::unique_ptr<std::thread> m_thread;
  boost::asio::io_service m_io_service;
  std::mutex m_mutex;
  std::atomic_int m_exit_flag;
  std::unique_ptr<boost::asio::deadline_timer> m_timer;
};
}
}
