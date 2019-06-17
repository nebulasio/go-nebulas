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
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the // GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//
#pragma once
#include "common/common.h"
#include "util/lru_cache.h"

namespace neb {
namespace util {
//! @KT - key type
//! @DT - data type
//! @CacheType - cache type, must have following methods
//    void set(const KT & k, const DT & v);
//    bool get(const KT & k, Value & v);
//  and should be concurrency-safe
//
template <class KT, class DT, class CacheType = util::lru_cache<DT, KT>>
class one_time_calculator {
public:
  typedef KT key_type;
  typedef DT data_type;
  typedef CacheType cache_type;
  typedef std::function<data_type()> calculator_t;

  template <typename... Args>
  one_time_calculator(Args... args) : m_cached_result(args...) {}

  virtual bool get_cached_or_ignore(const key_type &key, DT &v) {
    m_running_mutex.lock();
    bool status = m_cached_result.get(key, v);
    m_running_mutex.unlock();
    if (status) {
      return true;
    }
    return false;
  }

  virtual bool get_cached_or_cal_if_not_or_ignore(const key_type &key, DT &v,
                                                  const calculator_t &func) {
    m_running_mutex.lock();
    bool status = m_cached_result.get(key, v);
    if (status) {
      m_running_mutex.unlock();
      return true;
    }
    if (m_running_functions.find(key) != m_running_functions.end()) {
      m_running_mutex.unlock();
      return false;
    }
    m_running_functions.insert(key);
    m_running_mutex.unlock();

    v = func();
    m_running_mutex.lock();
    m_cached_result.set(key, v);
    m_running_functions.erase(key);
    m_running_mutex.unlock();
    return true;
  }

protected:
  cache_type m_cached_result;
  std::mutex m_running_mutex;
  std::unordered_set<key_type> m_running_functions;
};

} // namespace util
} // namespace neb
