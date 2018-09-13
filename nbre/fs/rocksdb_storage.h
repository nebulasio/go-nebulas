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

#pragma once

#include <cstdio>
#include <string>
#include <memory>

#include <rocksdb/db.h>
#include <rocksdb/slice.h>
#include <rocksdb/options.h>

namespace neb{
  namespace fs{
    class rocksdb_storage{
      public:
        rocksdb_storage();
        ~rocksdb_storage();
        rocksdb_storage(const rocksdb_storage& rs) = delete;
        rocksdb_storage& operate=(const rocksdb_storage&) = delete;

        rocksdb::Status open_database(const rocksdb::Options& options, const std::string& db_name);
        rocksdb::Status close_database();

        rocksdb::Status get_from_database(const rocksdb::Options& options, const rocksdb::Slice& key, std::string& value);
        rocksdb::Status put_to_database(const rocksdb::Options& options, const rocksdb::Slice& key, const std::string& value);
        rocksdb::Status del_from_atabase(const rocksdb::Options& options, const rocksdb::Slice& key);

        rocksdb::Status write_batch_to_database(const rocksdb::Options& options, const rocksdb::WriteBatch& batch);
      private:
        std::unique_ptr<rocksdb::DB> m_db;

    };
  };
};

