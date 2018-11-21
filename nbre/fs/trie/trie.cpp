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

#include "fs/trie/trie.h"
#include "fs/trie/byte_shared.h"

namespace neb {
namespace fs {

neb::util::bytes trie::key_to_route(const neb::util::bytes &key) {

  std::unique_ptr<byte_t[]> value;
  size_t size = key.size() << 1;

  if (size > 0) {
    value = std::unique_ptr<byte_t[]>(new byte_t[size]);
    for (size_t i = 0; i < key.size(); i++) {
      byte_shared byte(key[i]);
      value.get()[i << 1] = byte.bits_high();
      value.get()[(i << 1) + 1] = byte.bits_low();
    }
  }
  return neb::util::bytes(value.get(), size);
}

neb::util::bytes trie::route_to_key(const neb::util::bytes &route) {

  std::unique_ptr<byte_t[]> value;
  size_t size = route.size() >> 1;

  if (size > 0) {
    value = std::unique_ptr<byte_t[]>(new byte_t[size]);
    for (size_t i = 0; i < size; i++) {
      byte_shared byte(route[(i << 1) + 1], route[i << 1]);
      value.get()[i] = byte.data();
    }
  }
  return neb::util::bytes(value.get(), size);
}
} // namespace fs
} // namespace neb
