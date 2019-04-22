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

#include "rocksdb_storage.h"
#include <rocksdb/advanced_options.h>
#include <rocksdb/cache.h>
#include <rocksdb/filter_policy.h>
#include <rocksdb/options.h>
#include <rocksdb/slice.h>
#include <rocksdb/table.h>

namespace neb{
namespace fs {

rocksdb_storage::rocksdb_storage() : m_db(nullptr), m_enable_batch(false) {}

rocksdb_storage::~rocksdb_storage() { close_database(); }

void rocksdb_storage::open_database(const std::string &db_name,
                                    storage_open_flag flag) {
  rocksdb::DB *db = nullptr;
  rocksdb::Status status;

  if (nullptr == m_db) {
    if (flag == storage_open_for_readonly) {
      rocksdb::Options options;
      options.keep_log_file_num = 1;
      options.max_open_files = 500;
      status = rocksdb::DB::OpenForReadOnly(options, db_name, &db, false);
      m_enable_batch = true;
    } else {
      rocksdb::Options options;
      options.keep_log_file_num = 1;

      //! TODO setup bloomfilter, LRUCache, writer buffer size
      rocksdb::BlockBasedTableOptions table_options;
      table_options.filter_policy.reset(rocksdb::NewBloomFilterPolicy(10));
      table_options.block_cache = rocksdb::NewLRUCache(512 << 20);
      options.table_factory.reset(
          rocksdb::NewBlockBasedTableFactory(table_options));

      options.create_if_missing = true;
      options.max_open_files = 500;
      options.write_buffer_size = 64 * 1024 * 1024;
      options.IncreaseParallelism(4);
      status = rocksdb::DB::Open(options, db_name, &db);
      m_enable_batch = false;
    }

    if (status.ok()) {
      m_db = std::unique_ptr<rocksdb::DB>(db);
    } else {
      LOG(ERROR) << "open db error: " << status.ToString();
      throw storage_general_failure(status.ToString());
    }
  } else {
    throw std::runtime_error("database already open");
  }
}

void rocksdb_storage::close_database() {
  if (!m_db) {
    return;
  }
  rocksdb::Status status = m_db->Close();
  if (!status.ok()) {
    throw std::runtime_error("close database failed");
  }
  m_db.reset(nullptr);
}

bytes rocksdb_storage::get_bytes(const bytes &key) {
  if (!m_db) {
    throw storage_exception_no_init();
  }
  rocksdb::Slice s(reinterpret_cast<const char *>(key.value()), key.size());
  std::string value;
  auto status = m_db->Get(rocksdb::ReadOptions(), s, &value);
  if (!status.ok()) {
    throw storage_general_failure(status.ToString());
  }
  return string_to_byte(value);
}

void rocksdb_storage::put_bytes(const bytes &key, const bytes &val) {
  if (!m_db) {
    throw storage_exception_no_init();
  }
  if (m_enable_batch) {
    m_batched_ops.emplace_back([key, val](rocksdb::WriteBatch &wb) {
      std::string str_value = byte_to_string(val);
      rocksdb::Slice s(reinterpret_cast<const char *>(key.value()), key.size());
      auto status = wb.Put(s, str_value);
      if (!status.ok()) {
        throw storage_general_failure(status.ToString());
      }
    });
    return;
  }
  std::string str_value = byte_to_string(val);
  rocksdb::Slice s(reinterpret_cast<const char *>(key.value()), key.size());
  auto status = m_db->Put(rocksdb::WriteOptions(), s, str_value);
  if (!status.ok()) {
    throw storage_general_failure(status.ToString());
  }
}

void rocksdb_storage::del_by_bytes(const bytes &key) {
  if (!m_db) {
    throw storage_exception_no_init();
  }
  if (m_enable_batch) {
    m_batched_ops.emplace_back([key](rocksdb::WriteBatch &wb) {
      rocksdb::Slice s(reinterpret_cast<const char *>(key.value()), key.size());
      auto status = wb.Delete(s);
      if (!status.ok()) {
        throw storage_general_failure(status.ToString());
      }
    });
    return;
  }
  rocksdb::Slice s(reinterpret_cast<const char *>(key.value()), key.size());
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
  rocksdb::WriteBatch wb(0, 0);
  std::for_each(m_batched_ops.begin(), m_batched_ops.end(),
                [&wb](const batch_operation_t &f) { f(wb); });
  auto status = m_db->Write(rocksdb::WriteOptions(), &wb);
  if (!status.ok()) {
    throw storage_general_failure(status.ToString());
  }
}

void rocksdb_storage::display(
    const std::function<void(rocksdb::Iterator *)> &cb) {
  rocksdb::Iterator *it = m_db->NewIterator(rocksdb::ReadOptions());
  cb(it);
}

} // end namespace fs
} // end namespace neb

