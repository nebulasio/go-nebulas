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

namespace neb {
template <typename T, typename Lock = boost::shared_mutex>
class thread_safe_vector {
public:
  typedef std::vector<T> vector_t;
  typedef Lock lock_t;
  using guard_t = boost::shared_lock<lock_t>;

  thread_safe_vector() = default;

  void push_back(const typename vector_t::value_type &op) {
    guard_t _l(m_mutex);
    m_vector.push_back(op);
  }

  std::pair<bool, typename vector_t::value_type> try_pop_back() {
    guard_t _l(m_mutex);
    if (m_vector.empty()) {
      return std::make_pair(false, typename vector_t::value_type());
    }
    auto ret = m_vector.back();
    m_vector.pop_back();
    return std::make_pair(true, ret);
  }

  typename vector_t::value_type front() {
    guard_t _l(m_mutex);
    auto ret = m_vector.front();
    return ret;
  }

  typename vector_t::value_type back() {
    guard_t _l(m_mutex);
    auto ret = m_vector.back();
    return ret;
  }

  template <typename Func>
  std::pair<bool, typename vector_t::value_type>
  try_lower_bound(const typename vector_t::value_type &op, Func &&f) {
    guard_t _l(m_mutex);
    if (m_vector.empty()) {
      return std::make_pair(false, typename vector_t::value_type());
    }
    auto it = std::upper_bound(m_vector.begin(), m_vector.end(), op, f);
    it--;
    auto ret = *it;
    return std::make_pair(true, ret);
  }

  size_t size() const {
    guard_t _l(m_mutex);
    return m_vector.size();
  }

  bool empty() const {
    guard_t _l(m_mutex);
    return m_vector.empty();
  }

private:
  vector_t m_vector;
  mutable lock_t m_mutex;
};
} // namespace neb
