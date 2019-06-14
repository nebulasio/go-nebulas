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
class rocksdb_storage;
//! this is to handle rocksdb reopen issue
class rocksdb_session_storage : public storage {
public:
  rocksdb_session_storage();
  virtual ~rocksdb_session_storage();

  void init(const std::string &path, enum storage_open_flag flag);

  virtual bytes get_bytes(const bytes &key);
  virtual void put_bytes(const bytes &key, const bytes &val);
  virtual void del_by_bytes(const bytes &key);

  virtual void enable_batch();
  virtual void disable_batch();
  virtual void flush();

protected:
  std::unique_ptr<rocksdb_storage> m_storage;
  boost::shared_mutex m_mutex;
  std::string m_path;
  bool m_init_already;
  enum storage_open_flag m_open_flag;
};
} // namespace fs

} // namespace neb
