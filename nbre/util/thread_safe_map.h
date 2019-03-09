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
template <typename Key, typename Val> class thread_safe_map {
public:
  typedef std::unordered_map<Key, Val> map_t;
  thread_safe_map() = default;

  void insert(const typename map_t::value_type &op) {
    std::unique_lock<std::mutex> _l(m_mutex);
    m_map.insert(op);
  }

  std::pair<bool, typename map_t::mapped_type>
  try_get_val(const typename map_t::key_type &k) {
    std::unique_lock<std::mutex> _l(m_mutex);
    if (m_map.find(k) == m_map.end()) {
      return std::make_pair(false, typename map_t::mapped_type());
    }
    auto ret = m_map.find(k)->second;
    return std::make_pair(true, ret);
  }

  bool exist(const typename map_t::key_type &k) {
    std::unique_lock<std::mutex> _l(m_mutex);
    return m_map.find(k) != m_map.end();
  }

  size_t size() const {
    std::unique_lock<std::mutex> _l(m_mutex);
    return m_map.size();
  }

  bool empty() const {
    std::unique_lock<std::mutex> _l(m_mutex);
    return m_map.empty();
  }

private:
  map_t m_map;
  mutable std::mutex m_mutex;
};
} // namespace neb
