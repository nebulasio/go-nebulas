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
#include "common/quitable_thread.h"
#include "common/exception_queue.h"
#include "core/command.h"

namespace neb {
quitable_thread::quitable_thread() : m_exit_flag(false) {}

quitable_thread::~quitable_thread() {
  LOG(INFO) << "~quitable_thread : " << (void *)m_thread.get();
  if (m_thread) {
    m_thread->join();
    LOG(INFO) << "thread quit";
  }
  neb::core::command_queue::instance().unlisten_command(this);
  LOG(INFO) << "~quitable_thread done";
}

void quitable_thread::start() {
  LOG(INFO) << "quitable_thread start enter";
  neb::core::command_queue::instance().listen_command<neb::core::exit_command>(
      this, [this](const std::shared_ptr<neb::core::exit_command> &) {
        m_exit_flag = true;
      });

  m_thread = std::unique_ptr<std::thread>(new std::thread([this]() {
    try {
      this->thread_func();
    } catch (const std::exception &e) {
      LOG(INFO) << "got exception :" << e.what();
      exception_queue::instance().push_back(std::current_exception());
    }
  }));
  LOG(INFO) << "quitable_thread start done";
}
}
