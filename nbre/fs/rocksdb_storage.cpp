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
#include <memory>

using namespace std;
using namespace rocksdb;

namespace neb{

  namespace fs{

    rocksdb_storage::rocksdb_storage()
      : m_db(make_unique<DB>(nullptr))
    {
    }

    rocksdb_storage::~rocksdb_storage(){}

    Status rocksdb_storage::open_database(const Options& options, const string& db_name) {
      return DB::Open(&options, db_name, *m_db);
    }

    Status rocksdb_storage::close_database() {

    }

    Status rocksdb_storage::get_from_database(const Options& options, const Slice& key, string& value) {
      return m_db->Get(&options, key, &value);
    }

    Status rocksdb_storage::put_to_database(const Options& options, const Slice& key, const string& value) {
      return m_db->Put(&options, key, value);
    }

    Status rocksdb_storage::del_from_atabase(const Options& options, const Slice& key) {
      return m_db->Delete(&options, key);
    }

    Status rocksdb_storage::write_batch_to_database(const Options& options, const WriteBatch& batch) {
      return m_db->Write(&options, batch);
    }

  }

}


