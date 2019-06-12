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
#include "util/persistent_flag.h"
#include "fs/storage.h"

namespace neb {
namespace util {

persistent_flag::persistent_flag(fs::storage *storage,
                                 const std::string &key_name)
    : m_storage(storage), m_key_name(key_name) {}

void persistent_flag::set() {
  m_storage->put(m_key_name, neb::string_to_byte(m_key_name));
}

bool persistent_flag::test() {
  try {
    m_storage->get(m_key_name);
  } catch (...) {
    return false;
  }
  return true;
}

void persistent_flag::clear() {
  try {
    m_storage->del(m_key_name);
  } catch (...) {
  }
}
} // namespace util
} // namespace neb
