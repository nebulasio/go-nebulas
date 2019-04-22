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
#include <list>
#include <mutex>
#include <thread>

namespace neb {
namespace util {

template <class Key, class Value, class Lock = std::mutex,
          int32_t CacheCleanPeriod = 1, int32_t CacheCleanCounter = 64>
class lru_cache {
public:
  typedef std::unordered_map<Key, Value> map_t;
  typedef Lock lock_t;
  using guard_t = std::lock_guard<lock_t>;

  lru_cache() : m_thread_exit_flag(0) {
    m_thread = std::make_unique<std::thread>([&]() {
      while (!m_thread_exit_flag) {
        decrease_counter_and_clean_cache();
        std::this_thread::sleep_for(std::chrono::seconds(CacheCleanPeriod));
      }
    });
  }
  virtual ~lru_cache() {
    m_thread_exit_flag = 1;
    m_thread->join();
  };

  size_t size() const {
    guard_t __l(m_lock);
    return m_cache_map.size();
  }

  bool empty() const {
    guard_t __l(m_lock);
    return m_cache_map.empty();
  }

  void clear() {
    guard_t __l(m_lock);
    m_cache_map.clear();
    m_counter.clear();
  }

  void set(const Key &k, const Value &v) {
    guard_t __l(m_lock);
    const auto iter = m_cache_map.find(k);
    if (iter != m_cache_map.end()) {
      return;
    }

    m_cache_map[k] = v;
    m_counter[k] = CacheCleanCounter;
  }

  bool get(const Key &k, Value &v) {
    guard_t __l(m_lock);
    const auto iter = m_cache_map.find(k);
    if (iter == m_cache_map.end()) {
      return false;
    }
    v = iter->second;
    m_counter[k] = CacheCleanCounter;

    return true;
  }

  bool exists(const Key &k) const {
    guard_t __l(m_lock);
    return m_cache_map.find(k) != m_cache_map.end();
  }

  template <typename F> void watch(F &&f) const {
    guard_t __l(m_lock);
    std::for_each(
        m_cache_map.begin(), m_cache_map.end(),
        [&f](const std::pair<Key, Value> &it) { f(it.first, it.second); });
  }

protected:
  void decrease_counter_and_clean_cache() {
    guard_t __l(m_lock);
    std::unordered_set<Key> to_remove_sets;
    for (auto it = m_counter.begin(); it != m_counter.end(); ++it) {
      it->second--;
      if (it->second <= 0) {
        to_remove_sets.insert(it->first);
      }
    }
    for (auto key : to_remove_sets) {
      m_cache_map.erase(key);
      m_counter.erase(key);
    }
  }

private:
  mutable Lock m_lock;
  map_t m_cache_map;
  std::unordered_map<Key, int32_t> m_counter;
  std::unique_ptr<std::thread> m_thread;
  std::atomic_int m_thread_exit_flag;
};
} // namespace util
} // namespace neb
