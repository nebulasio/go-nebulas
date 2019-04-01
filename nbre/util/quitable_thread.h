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
#include "core/command.h"
#include "util/wakeable_queue.h"
#include <atomic>
#include <thread>

namespace neb {
namespace util {
class quitable_thread {
public:
  quitable_thread();
  virtual ~quitable_thread();

  virtual void start();
  virtual void stop();
  virtual void thread_func() = 0;

protected:
  std::unique_ptr<std::thread> m_thread;
  std::atomic_bool m_exit_flag;
};

class wakeable_thread : public quitable_thread {
public:
  wakeable_thread();
  template <typename Func> void schedule(Func &&f) {
    m_queue.push_back(std::make_shared<std::function<void()>>(f));
    if (!m_started) {
      m_started = true;
      neb::core::command_queue::instance()
          .listen_command<neb::core::exit_command>(
              this, [this](const std::shared_ptr<neb::core::exit_command> &) {
                m_queue.wake_up_if_empty();
              });
      start();
    }
  }
  virtual void thread_func();
  inline size_t size() const { return m_queue.size(); }

protected:
  typedef std::shared_ptr<std::function<void()>> func_ptr;
  using queue_t = wakeable_queue<func_ptr>;
  queue_t m_queue;
  bool m_started;
};
} // namespace util
} // namespace neb
