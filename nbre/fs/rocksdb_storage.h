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

#include "common/address.h"
#include "common/common.h"
#include "fs/storage.h"
#include <rocksdb/db.h>
#include <rocksdb/write_batch.h>

namespace neb {
namespace fs {

class rocksdb_storage : public storage {
public:
  rocksdb_storage();
  virtual ~rocksdb_storage();
  rocksdb_storage(const rocksdb_storage &rs) = delete;
  rocksdb_storage &operator=(const rocksdb_storage &) = delete;

  void open_database(const std::string &db_name, storage_open_flag flag);
  void close_database();

  virtual bytes get_bytes(const bytes &key);
  virtual void put_bytes(const bytes &key, const bytes &val);
  virtual void del_by_bytes(const bytes &key);

  virtual void enable_batch();
  virtual void disable_batch();
  virtual void flush();

  virtual void display(const std::function<void(rocksdb::Iterator *)> &cb);

private:
  std::unique_ptr<rocksdb::DB> m_db;
  bool m_enable_batch;
  typedef std::function<void(rocksdb::WriteBatch &wb)> batch_operation_t;
  std::vector<batch_operation_t> m_batched_ops;
}; // end class rocksdb_storage
} // end namespace fs
} // end namespace neb

