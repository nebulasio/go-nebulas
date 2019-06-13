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
#include "fs/bc_storage_session_interface.h"
#include "fs/bc_storage_session.h"

namespace neb {
namespace fs {
namespace bc_storage {

void init(const std::string &path, enum storage_open_flag flag) {
  bc_storage_session::instance().init(path, flag);
}

bytes get_bytes(const bytes &key) {
  return bc_storage_session::instance().get_bytes(key);
}
void put_bytes(const bytes &key, const bytes &value) {
  bc_storage_session::instance().put_bytes(key, value);
}
} // namespace bc_storage
} // namespace fs
} // namespace neb
