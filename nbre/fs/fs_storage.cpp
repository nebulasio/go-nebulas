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

#include "fs/fs_storage.h"
#include "core/neb_ipc/server/ipc_configuration.h"

namespace neb {
namespace fs {

fs_storage::fs_storage() {
  m_storage = std::make_unique<rocksdb_storage>();
  m_storage->open_database(
      neb::core::ipc_configuration::instance().nbre_db_dir(),
      storage_open_for_readwrite);

  m_blockchain = std::make_unique<blockchain>(
      neb::core::ipc_configuration::instance().neb_db_dir(),
      storage_open_for_readonly);
}

fs_storage::~fs_storage() {
  if (m_storage) {
    m_storage->close_database();
  }
}
} // namespace fs
} // namespace neb
