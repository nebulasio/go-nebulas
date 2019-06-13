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
#include "fs/memory_storage.h"

namespace neb {
namespace fs {
memory_storage::memory_storage() = default;
memory_storage::~memory_storage() = default;

bytes memory_storage::get_bytes(const bytes &key) {
  auto ret = m_memory.try_get_val(key);
  if (!ret.first) {
    throw storage_general_failure("get memory storage failed");
  }
  return ret.second;
}

void memory_storage::put_bytes(const bytes &key, const bytes &val) {
  m_memory.insert(key, val);
}

void memory_storage::del_by_bytes(const bytes &key) { m_memory.erase(key); }
}
} // namespace neb
