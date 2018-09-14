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
#include "common/common.h"

namespace neb {

namespace util {
namespace internal {
template <size_t ByteLength = 32> class sha_hash_impl {
public:
  sha_hash_impl() : m_value{0} {};
  sha_hash_impl(const sha_hash_impl<ByteLength> &v) : m_value{0} {
    memcpy(m_value, v.m_value, ByteLength);
  }
  bool operator==(const sha_hash_impl<ByteLength> &v) const {
    return memcmp(m_value, v.m_value, ByteLength) == 0;
  }

  std::string to_base58() const { return ""; }

  std::string to_hex() const { return ""; }

  static sha_hash_impl<ByteLength> from_base58(const std::string &t) {}
  static sha_hash_impl<ByteLength> from_hex(const std::string &t) {}

protected:
  uint8_t m_value[ByteLength];
}; // end class sha_hash_impl
} // end namespace internal
} // end namespace util
typedef util::internal::sha_hash_impl<32> sha256_hash;
}
