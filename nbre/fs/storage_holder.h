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

#pragma once

#include "common/byte.h"
#include "fs/blockchain.h"
#include "fs/rocksdb_storage.h"
#include "util/singleton.h"

namespace neb {
namespace fs {
class storage_holder : public util::singleton<storage_holder> {
public:
  storage_holder();
  ~storage_holder();

  inline rocksdb_storage *nbre_db_ptr() { return m_storage.get(); }
  // inline blockchain *neb_db_ptr() { return m_blockchain.get(); }

private:
  std::unique_ptr<rocksdb_storage> m_storage;
  // std::unique_ptr<blockchain> m_blockchain;
};
} // namespace fs
} // namespace neb
