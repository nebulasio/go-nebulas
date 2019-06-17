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
#include "common/byte.h"
#include "common/common.h"
#include "fs/storage.h"
#include "util/lru_cache.h"

namespace neb {
namespace util {
template <class KT, class DT> class db_mem_cache {
public:
  typedef KT key_type;
  typedef DT data_type;

  db_mem_cache(fs::storage *storage) : m_storage(storage) {}

  fs::storage *storage() const { return m_storage; }

  void set(const key_type &k, const data_type &v) {
    m_mem_data.set(k, v);
    m_storage->put_bytes(get_key_bytes(k), serialize_data_to_bytes(v));
  }

  bool get(const key_type &k, data_type &v) {
    bool status = m_mem_data.get(k, v);
    if (status) {
      return true;
    }
    try {
      bytes data = m_storage->get_bytes(get_key_bytes(k));
      v = deserialize_data_from_bytes(data);
      return true;
    } catch (...) {
      return false;
    }
  }

  virtual bytes get_key_bytes(const key_type &k) {
    return string_to_byte(std::to_string(k));
  }

  virtual bytes serialize_data_to_bytes(const data_type &v) = 0;
  virtual data_type deserialize_data_from_bytes(const bytes &data) = 0;

protected:
  fs::storage *m_storage;
  lru_cache<key_type, data_type> m_mem_data;
};
} // namespace util
} // namespace neb
