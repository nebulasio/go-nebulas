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
template <typename Key, typename Val, typename Lock = boost::shared_mutex>
class thread_safe_map {
public:
  typedef std::map<Key, Val> map_t;
  typedef Lock lock_t;
  using r_guard_t = boost::shared_lock<lock_t>;
  using w_guard_t = std::unique_lock<lock_t>;

  thread_safe_map() = default;

  void insert(const typename map_t::key_type &k,
              const typename map_t::mapped_type &v) {
    w_guard_t _l(m_mutex);
    if (m_map.find(k) != m_map.end()) {
      m_map[k] = v;
      return;
    }
    m_map.insert(std::make_pair(k, v));
  }

  bool insert_if_not_exist(const typename map_t::key_type &k,
                           const typename map_t::mapped_type &v) {
    w_guard_t _l(m_mutex);
    if (m_map.find(k) != m_map.end()) {
      m_map[k] = v;
      return false;
    }
    m_map.insert(std::make_pair(k, v));
    return true;
  }

  void erase(const typename map_t::key_type &k) {
    w_guard_t _l(m_mutex);
    auto it = m_map.find(k);
    if (it != m_map.end()) {
      m_map.erase(it);
    }
  }

  std::pair<bool, typename map_t::mapped_type>
  try_get_and_update_val(const typename map_t::key_type &k,
                         const typename map_t::mapped_type &v) {
    w_guard_t _l(m_mutex);
    auto ret = try_get_val(k);
    insert(k, v);
    return ret;
  }

  std::pair<bool, typename map_t::mapped_type>
  try_get_val(const typename map_t::key_type &k) {
    r_guard_t _l(m_mutex);
    if (m_map.find(k) == m_map.end()) {
      return std::make_pair(false, typename map_t::mapped_type());
    }
    auto ret = m_map.find(k)->second;
    return std::make_pair(true, ret);
  }

  std::pair<bool, typename map_t::value_type>
  try_lower_than(const typename map_t::key_type &k) {
    r_guard_t _l(m_mutex);
    if (m_map.empty() || k < m_map.begin()->first) {
      return std::make_pair(false, typename map_t::value_type());
    }
    auto it = m_map.upper_bound(k);
    it--;
    auto ret = *it;
    return std::make_pair(true, ret);
  }

  typename map_t::value_type begin() {
    r_guard_t _l(m_mutex);
    auto ret = m_map.begin();
    return *ret;
  }

  typename map_t::value_type end() {
    r_guard_t _l(m_mutex);
    auto ret = m_map.end();
    return *ret;
  }

  bool exist(const typename map_t::key_type &k) {
    r_guard_t _l(m_mutex);
    return m_map.find(k) != m_map.end();
  }

  size_t size() const {
    r_guard_t _l(m_mutex);
    return m_map.size();
  }

  bool empty() const {
    r_guard_t _l(m_mutex);
    return m_map.empty();
  }

  void clear() {
    w_guard_t _l(m_mutex);
    m_map.clear();
  }

private:
  map_t m_map;
  mutable lock_t m_mutex;
};
} // namespace neb
