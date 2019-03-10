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
#include "fs/bc_storage_session.h"

namespace neb {
namespace fs {
bc_storage_session::bc_storage_session() : m_storage(), m_mutex() {
  m_storage = std::make_unique<rocksdb_storage>();
  m_init_already = false;
}

bc_storage_session::~bc_storage_session() {
  boost::unique_lock<boost::shared_mutex> _l(m_mutex);
  m_storage->close_database();
}

void bc_storage_session::init(const std::string &path,
                              enum storage_open_flag flag) {
  boost::unique_lock<boost::shared_mutex> _l(m_mutex);
  if (m_init_already)
    return;
  m_init_already = true;
  m_open_flag = flag;
  m_path = path;
  m_storage->open_database(m_path, m_open_flag);
}

bytes bc_storage_session::get_bytes(const bytes &key) {
  boost::shared_lock<boost::shared_mutex> _l(m_mutex);
  bool no_exception = true;
  bool tried_already = false;
  while (no_exception) {
    if (tried_already) {
      return m_storage->get_bytes(key);
    } else {
      try {
        return m_storage->get_bytes(key);
      } catch (...) {
        tried_already = true;
        _l.unlock();
        m_mutex.lock();
        m_storage->close_database();
        m_storage->open_database(m_path, m_open_flag);
        m_mutex.unlock();
        _l.lock();
      }
    }
  }
  return bytes();
}

void bc_storage_session::put_bytes(const bytes &key, const bytes &val) {
  boost::shared_lock<boost::shared_mutex> _l(m_mutex);
  m_storage->put_bytes(key, val);
}
} // namespace fs
} // namespace neb
