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
#include <atomic>
#include <thread>

namespace neb {
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
}
