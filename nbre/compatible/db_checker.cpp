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
#include "compatible/db_checker.h"
#include "fs/storage_holder.h"
#include "runtime/version.h"

namespace neb {
namespace compatible {
static std::string key_nbre_version = "key_nbre_version";
void db_checker::update_db_if_needed() {
  fs::rocksdb_storage *db = fs::storage_holder::instance().nbre_db_ptr();

  version cur = rt::get_version();
  db->put(key_nbre_version, cur.data());
}
} // namespace compatible
} // namespace neb
