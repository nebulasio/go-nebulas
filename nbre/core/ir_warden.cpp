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

#include "core/ir_warden.h"
#include <boost/date_time/posix_time/posix_time.hpp>

namespace neb {
namespace core {
ir_warden::~ir_warden() {}

std::vector<std::shared_ptr<nbre::NBREIR>>
ir_warden::get_ir_from_height(const std::string &name, block_height_t height) {
  // TODO add impl here
  return std::vector<std::shared_ptr<nbre::NBREIR>>();
}

bool ir_warden::is_sync_already() const {
  // TODO add impl here
  throw std::invalid_argument("no impl");
}

void ir_warden::wait_until_sync() {
  // TODO add impl here
  throw std::invalid_argument("no impl");
}

void ir_warden::timer_callback() {
  boost::asio::deadline_timer pt(m_io_service, boost::posix_time::seconds(15));
  pt.async_wait(std::bind(&ir_warden::timer_callback, this,
                          boost::asio::placeholders::error));
  m_io_service.run();
}
ir_warden::ir_warden() {
  m_thread = std::unique_ptr<std::thread>(
      new std::thread([this]() { timer_callback(); }));
}
}
}
