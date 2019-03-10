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
#include "util/quitable_thread.h"
#include "common/exception_queue.h"
#include "core/command.h"

namespace neb {
namespace util {

quitable_thread::quitable_thread() : m_exit_flag(false) {}

quitable_thread::~quitable_thread() {
  if (m_thread) {
    m_thread->join();
    m_thread.reset();
  }
  neb::core::command_queue::instance().unlisten_command(this);
}

void quitable_thread::start() {
  neb::core::command_queue::instance().listen_command<neb::core::exit_command>(
      this, [this](const std::shared_ptr<neb::core::exit_command> &) {
        m_exit_flag = true;
      });

  m_thread = std::unique_ptr<std::thread>(new std::thread([this]() {
    exception_queue::catch_exception([this]() { this->thread_func(); });
  }));
}

void quitable_thread::stop() {
  std::shared_ptr<neb::core::exit_command> exit_command =
      std::make_shared<neb::core::exit_command>();
  neb::core::command_queue::instance().send_command<neb::core::exit_command>(
      exit_command);
}

wakeable_thread::wakeable_thread()
    : quitable_thread(), m_queue(), m_started(false) {}

void wakeable_thread::thread_func() {
  while (!m_exit_flag) {
    auto t = m_queue.pop_front();
    if (t.first) {
      (*t.second)();
    }
  }
}
} // namespace util

} // namespace neb
