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
#include "core/command.h"
#include <boost/asio.hpp>
#include <boost/asio/deadline_timer.hpp>

namespace neb {
namespace util {

class timer_loop {
public:
  inline timer_loop(boost::asio::io_service *service)
      : m_service(service), m_exit_flag(false) {
    neb::core::command_queue::instance()
        .listen_command<neb::core::exit_command>(
            this, [this](const std::shared_ptr<neb::core::exit_command> &) {
              m_exit_flag = true;
            });
  }

  template <typename Func>
  void register_timer_and_callback(long seconds, Func &&f) {

    auto timer = std::make_shared<boost::asio::deadline_timer>(
        *m_service, boost::posix_time::seconds(seconds));

    timer->async_wait(
        [this, timer, seconds, f](const boost::system::error_code &ec) {
          timer_callback(ec, seconds, timer, f);
        });
  }

protected:
  void timer_callback(const boost::system::error_code &ec, long seconds,
                      std::shared_ptr<boost::asio::deadline_timer> timer,
                      std::function<void()> func);

protected:
  boost::asio::io_service *m_service;
  std::atomic_bool m_exit_flag;
};
} // namespace util
} // namespace neb
