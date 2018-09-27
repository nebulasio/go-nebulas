// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
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
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

#include "rocksdb_storage.h"
#include <rocksdb/advanced_options.h>
#include <rocksdb/options.h>
#include <rocksdb/slice.h>

namespace neb{
namespace fs {

rocksdb_storage::rocksdb_storage() : m_db(nullptr) {}

rocksdb_storage::~rocksdb_storage() {
  if (m_db) {
    close_database();
  }
}

void rocksdb_storage::open_database(const std::string &db_name,
                                    storage_open_flag flag) {
  rocksdb::DB *db = nullptr;
  rocksdb::Status status;

  if (nullptr == m_db) {
    if (flag == storage_open_for_readonly) {
      rocksdb::Options options;
      status = rocksdb::DB::OpenForReadOnly(options, db_name, &db);
      m_enable_batch = true;
    } else {
      rocksdb::Options options;
      //! TODO setup bloomfilter, LRUCache, writer buffer size
      options.create_if_missing = true;
      status = rocksdb::DB::Open(options, db_name, &db);
      m_enable_batch = false;
    }

    if (status.ok()) {
      m_db = std::unique_ptr<rocksdb::DB>(db);
    } else {
      throw storage_general_failure(status.ToString());
    }
  } else {
    throw std::runtime_error("database already open");
  }
}

void rocksdb_storage::close_database() {
  if (!m_db)
    return;
  rocksdb::Status status = m_db->Close();
  if (!status.ok()) {
    throw std::runtime_error("close database failed");
  }
  m_db.reset(nullptr);
}

util::bytes rocksdb_storage::get_bytes(const util::bytes &key) {
  if (!m_db) {
    throw storage_exception_no_init();
  }
  rocksdb::Slice s((const char *)key.value(), key.size());
  std::string value;
  auto status = m_db->Get(rocksdb::ReadOptions(), s, &value);
  if (!status.ok()) {
    throw storage_general_failure(status.ToString());
  }
  return util::string_to_byte(value);
}

void rocksdb_storage::put_bytes(const util::bytes &key,
                                const util::bytes &val) {
  if (!m_db) {
    throw storage_exception_no_init();
  }
  if (m_enable_batch) {
    m_batched_ops.push_back([key, val](rocksdb::WriteBatch &wb) {
      std::string str_value = util::byte_to_string(val);
      rocksdb::Slice s((const char *)key.value(), key.size());
      auto status = wb.Put(s, str_value);
      if (!status.ok()) {
        throw storage_general_failure(status.ToString());
      }
    });
    return;
  }
  std::string str_value = util::byte_to_string(val);
  rocksdb::Slice s((const char *)key.value(), key.size());
  auto status = m_db->Put(rocksdb::WriteOptions(), s, str_value);
  if (!status.ok()) {
    throw storage_general_failure(status.ToString());
  }
  return;
}

void rocksdb_storage::del_by_bytes(const util::bytes &key) {
  if (!m_db) {
    throw storage_exception_no_init();
  }
  if (m_enable_batch) {
    m_batched_ops.push_back([key](rocksdb::WriteBatch &wb) {
      rocksdb::Slice s((const char *)key.value(), key.size());
      auto status = wb.Delete(s);
      if (!status.ok()) {
        throw storage_general_failure(status.ToString());
      }
    });
    return;
  }
  rocksdb::Slice s((const char *)key.value(), key.size());
  auto status = m_db->Delete(rocksdb::WriteOptions(), s);
  if (!status.ok()) {
    throw storage_general_failure(status.ToString());
  }
}

void rocksdb_storage::enable_batch() { m_enable_batch = true; }
void rocksdb_storage::disable_batch() {
  if (m_enable_batch) {
    flush();
  }
  m_enable_batch = false;
}
void rocksdb_storage::flush() {
  if (!m_enable_batch) {
    return;
  }
  if (!m_db) {
    return;
  }
  rocksdb::WriteBatch wb;
  std::for_each(m_batched_ops.begin(), m_batched_ops.end(),
                [&wb](const batch_operation_t &f) { f(wb); });
  auto status = m_db->Write(rocksdb::WriteOptions(), &wb);
  if (!status.ok()) {
    throw storage_general_failure(status.ToString());
  }
}

void rocksdb_storage::show_all(
    const std::function<void(rocksdb::Iterator *)> &cb) {
  rocksdb::Iterator *it = m_db->NewIterator(rocksdb::ReadOptions());
  cb(it);
}

} // end namespace fs
} // end namespace neb

