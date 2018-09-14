// Copyright (C) 2018 go-nebulas authors
//
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
#include "common/util/byte.h"

namespace neb {
namespace util {
bytes::bytes() : m_value(nullptr) {}
bytes::bytes(size_t len)
    : m_value(std::unique_ptr<byte_t[]>(new byte_t[len])) {}
bytes::bytes(const bytes &v) : bytes(v.size()) {
  memcpy(m_value.get(), v.m_value.get(), v.size());
}
bytes::bytes(bytes &&v) : m_value(std::move(v.m_value)) {}
bytes &bytes::operator=(const bytes &v) {
  m_value = std::unique_ptr<byte_t[]>(new byte_t[v.size()]);
  memcpy(m_value.get(), v.m_value.get(), v.size());
  return *this;
}
}
}
