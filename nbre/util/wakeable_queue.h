// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or // (at your
// option) any later version.
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
#include <condition_variable>
#include <queue>

namespace neb {
namespace util {

template <typename T> class wakeable_queue {
public:
  typedef std::queue<T> queue_t;
  wakeable_queue() = default;

  void push_back(const typename queue_t::value_type &op) {
    std::unique_lock<std::mutex> _l(m_mutex);
    bool was_empty = m_queue.empty();
    m_queue.push(op);
    _l.unlock();
    if (was_empty) {
      m_cond_var.notify_all();
    }
  }

  std::pair<bool, typename queue_t::value_type> pop_front() {
    std::unique_lock<std::mutex> _l(m_mutex);
    bool is_empty = m_queue.empty();
    if (is_empty) {
      m_cond_var.wait(_l);
    }
    if (m_queue.empty()) {
      return std::make_pair(false, typename queue_t::value_type());
    }
    auto ret = m_queue.front();
    m_queue.pop();
    return std::make_pair(true, ret);
  }

  std::pair<bool, typename queue_t::value_type> try_pop_front() {
    std::unique_lock<std::mutex> _l(m_mutex);
    if (m_queue.empty()) {
      return std::make_pair(false, typename queue_t::value_type());
    }
    auto ret = m_queue.front();
    m_queue.pop();
    return std::make_pair(true, ret);
  }

  size_t size() const {
    std::unique_lock<std::mutex> _l(m_mutex);
    return m_queue.size();
  }

  bool empty() const {
    std::unique_lock<std::mutex> _l(m_mutex);
    return m_queue.empty();
  }

  void wake_up_if_empty() {
    std::unique_lock<std::mutex> _l(m_mutex);
    if (m_queue.empty()) {
      _l.unlock();
      m_cond_var.notify_one();
    }
  }

protected:
  queue_t m_queue;
  mutable std::mutex m_mutex;
  std::condition_variable m_cond_var;
};
} // namespace util

} // namespace neb
