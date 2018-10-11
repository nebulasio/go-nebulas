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
#include "core/exception_handler.h"
#include "common/exception_queue.h"
#include "core/command.h"

namespace neb {
namespace core {

exception_handler::~exception_handler() {
  if (m_thread) {
    LOG(INFO) << "to wait exception handler done!";
    m_thread->join();
    LOG(INFO) << "exception handler done!";
  }
  m_thread.reset();
}

void exception_handler::kill() {
  neb::exception_queue::instance().push_back(std::logic_error("to quit"));
}

void exception_handler::run() {
  m_thread = std::unique_ptr<std::thread>(new std::thread([]() {
    neb::exception_queue &eq = neb::exception_queue::instance();
    auto ep = eq.pop_front();
    if (ep) {
      // TODO we just send exit_command for now
      command_queue::instance().send_command(std::make_shared<exit_command>());
    }
  }));
}
} // namespace core
} // namespace neb
