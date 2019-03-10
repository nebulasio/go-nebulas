// Copyright (C) 2017 go-nebulas authors
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
#include "fs/rocksdb_storage.h"
#include "fs/storage.h"
#include "util/singleton.h"

namespace neb {
namespace fs {
class bc_storage_session : public util::singleton<bc_storage_session> {
public:
  bc_storage_session();
  ~bc_storage_session();

  void init(const std::string &path, enum storage_open_flag flag);

  bytes get_bytes(const bytes &key);

  template <typename FixBytes> bytes get_bytes(const FixBytes &key) {
    return get_bytes(from_fix_bytes(key));
  }

  inline bytes get_bytes(const std::string &key) {
    return get_bytes(string_to_byte(key));
  }

  inline std::string get_string(const bytes &key) {
    return byte_to_string(get_bytes(key));
  }
  inline std::string get_string(const std::string &key) {
    return byte_to_string(get_bytes(key));
  }

  void put_bytes(const bytes &key, const bytes &value);
  inline void put(const bytes &key, const bytes &value) {
    return put_bytes(key, value);
  }
  inline void put(const std::string &key, const bytes &value) {
    return put_bytes(string_to_byte(key), value);
  }
  inline void put(const bytes &key, const std::string &value) {
    return put_bytes(key, string_to_byte(value));
  }
  inline void put(const std::string &key, const std::string &value) {
    return put_bytes(string_to_byte(key), string_to_byte(value));
  }

protected:
  std::unique_ptr<rocksdb_storage> m_storage;
  boost::shared_mutex m_mutex;
  std::string m_path;
  bool m_init_already;
  enum storage_open_flag m_open_flag;
};
} // namespace fs
} // namespace neb
