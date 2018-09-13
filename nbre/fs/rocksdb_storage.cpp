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

namespace neb{

  namespace fs{

  rocksdb_storage::rocksdb_storage() : m_db(nullptr) {}

  rocksdb_storage::~rocksdb_storage() = default;

  rocksdb::Status
  rocksdb_storage::open_database(const rocksdb::Options &options,
                                 const std::string &db_name) {
    rocksdb::DB *db = nullptr;

    rocksdb::Status status;
    if (nullptr == m_db) {
      status = rocksdb::DB::Open(options, db_name, &db);

      if (status.ok()) {
        m_db = std::unique_ptr<rocksdb::DB>(db);
      }
    }

    return status;
    }

      rocksdb::Status rocksdb_storage::close_database() {
        rocksdb::Status status;

        if (nullptr != m_db) {
          status = m_db->Close();
        }
        
        return status;
      }

      rocksdb::Status
      rocksdb_storage::get_from_database(const rocksdb::ReadOptions &options,
                                         const rocksdb::Slice &key,
                                         std::string &value) {
        rocksdb::Status status;

        if (nullptr != m_db) {
          status = m_db->Get(options, key, &value);
        }

        return status;
      }

      rocksdb::Status
      rocksdb_storage::put_to_database(const rocksdb::WriteOptions &options,
                                       const rocksdb::Slice &key,
                                       const std::string &value) {
        rocksdb::Status status;

        if (nullptr != m_db) {
          status = m_db->Put(options, key, value);
        }

        return status;
      }

      rocksdb::Status
      rocksdb_storage::del_from_atabase(const rocksdb::WriteOptions &options,
                                        const rocksdb::Slice &key) {
        rocksdb::Status status;

        if (nullptr != m_db) {
          status = m_db->Delete(options, key);
        }

        return status;
      }

      rocksdb::Status rocksdb_storage::write_batch_to_database(
          const rocksdb::WriteOptions &options, rocksdb::WriteBatch *batch) {
        rocksdb::Status status;

        if (nullptr != m_db) {
          status = m_db->Write(options, batch);
        }

        return status;
    }
  } // namespace fs
} // namespace neb


