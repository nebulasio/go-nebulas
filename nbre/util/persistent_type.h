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
#include "fs/storage.h"

namespace neb {
namespace fs {
class storage;
}
namespace util {
template <class T> class persistent_type {
public:
  persistent_type(fs::storage *storage, const std::string &key_name)
      : m_storage(storage), m_key_name(key_name) {}

  void set(const T &v) { m_storage->put(m_key_name, v); }
  T get() const {
    T v = T();
    try {
      v = m_storage->get<T>(m_key_name);
    } catch (...) {
      m_storage->put(m_key_name, v);
    }
    return v;
  }
  void clear() {
    try {
      m_storage->del(m_key_name);
    } catch (...) {
    }
  }

protected:
  fs::storage *m_storage;
  std::string m_key_name;
};
} // namespace util
} // namespace neb
