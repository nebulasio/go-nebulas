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
#pragma once
#include "common/byte.h"
#include "common/common.h"
#include <cryptopp/sha3.h>

namespace neb {
namespace crypto {
typedef fix_bytes<32> sha3_256_hash_t;

sha3_256_hash_t sha3_256_hash(const bytes &b);

namespace internal {
template <typename T, bool is_arith = std::is_arithmetic<T>::value>
struct sha3_256_hash_helper {};
template <typename T> struct sha3_256_hash_helper<T, true> {
  static sha3_256_hash_t get_hash(const T &t) {
    bytes b(sizeof(t));
    number_to_byte(t, b, sizeof(t));
    return sha3_256_hash(b);
  }
};

template <> struct sha3_256_hash_helper<std::string, false> {
  static sha3_256_hash_t get_hash(const std::string &t) {
    bytes b = string_to_byte(t);
    return sha3_256_hash(b);
  }
};

} // namespace internal
template <typename T> sha3_256_hash_t sha3_256_hash(const T &v) {
  return internal::sha3_256_hash_helper<T>::get_hash(v);
};
}
typedef crypto::sha3_256_hash_t hash_t;
} // namespace neb
