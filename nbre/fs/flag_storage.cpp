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

#include "fs/flag_storage.h"
#include "common/common.h"

namespace neb {
namespace fs {

flag_storage::flag_storage(rocksdb_storage *db_ptr) : m_storage(db_ptr) {}

void flag_storage::set_flag(const std::string &flag) {
  m_storage->put(flag, neb::util::string_to_byte(flag));
}

bool flag_storage::has_flag(const std::string &flag) const {
  try {
    m_storage->get(flag);
  } catch (const std::exception &e) {
    return false;
  }
  return true;
}

void flag_storage::del_flag(const std::string &flag) { m_storage->del(flag); }

} // namespace fs
} // namespace neb
