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

#include "util/timer_loop.h"

namespace neb {
namespace util {

void timer_loop::timer_callback(const boost::system::error_code &ec,
                                long seconds,
                                boost::asio::deadline_timer *timer,
                                std::function<void()> func) {
  if (m_exit_flag)
    return;
  if (ec) {
    LOG(ERROR) << ec;
    return;
  }
  func();
  timer->expires_at(timer->expires_at() + boost::posix_time::seconds(seconds));
  timer->async_wait(
      [this, timer, seconds, func](const boost::system::error_code &ec) {
        timer_callback(ec, seconds, timer, func);
      });
}
} // namespace util
} // namespace neb
