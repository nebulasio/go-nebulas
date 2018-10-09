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
#include "common/util/singleton.h"
#include <algorithm>
#include <condition_variable>
#include <exception>
#include <thread>

namespace neb {
class exception_queue : public neb::util::singleton<exception_queue> {
public:
  inline void push_back(const std::exception_ptr p) {
    std::lock_guard<std::mutex> _l(m_mutex);
    m_exceptions.push_back(p);
  }

  inline bool empty() const {
    std::lock_guard<std::mutex> _l(m_mutex);
    return m_exceptions.empty();
  }
  inline size_t size() const {
    std::lock_guard<std::mutex> _l(m_mutex);
    return m_exceptions.size();
  }

  inline void for_each(const std::function<void(std::exception_ptr p)> &func) {
    std::lock_guard<std::mutex> _l(m_mutex);
    std::for_each(m_exceptions.begin(), m_exceptions.end(), func);
  }

protected:
  std::vector<std::exception_ptr> m_exceptions;
  mutable std::mutex m_mutex;
};
}

