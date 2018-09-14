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
#include "common/util/base58.h"
#include <iomanip>
#include <sstream>

namespace neb {
namespace util {

namespace internal {
std::string convert_byte_to_hex(const byte_t *buf, size_t len) {
  std::stringstream s;
  for (size_t i = 0; i < len; i++) {
    s << std::hex << std::setfill('0') << static_cast<uint8_t>(buf[i]);
  }
  return s.str();
}
std::string convert_byte_to_base58(const byte_t *buf, size_t len) {
  return ::neb::encode_base58(buf, buf + len);
}

bool convert_hex_to_bytes(const std::string &s, byte_t *buf, size_t &len) {
  auto char2int = [](char input) {
    if (input >= '0' && input <= '9')
      return input - '0';
    if (input >= 'A' && input <= 'F')
      return input - 'A' + 10;
    if (input >= 'a' && input <= 'f')
      return input - 'a' + 10;
    throw std::invalid_argument("Invalid input string");
  };

  try {
    size_t i = 0;
    while (i * 2 < s.size() && i * 2 + 1 < s.size()) {
      if (buf) {
        buf[i] = (char2int(s[i * 2]) << 4) + char2int(s[i * 2 + 1]);
      }
      i++;
    }
    len = i;
  } catch (std::exception &e) {
    return false;
  }
  return true;
}

bool convert_base58_to_bytes(const std::string &s, byte_t *buf, size_t &len) {
  std::vector<unsigned char> ret;
  bool rv = ::neb::decode_base58(s, ret);
  if (rv) {
    len = ret.size();
    if (!buf) {
      memcpy(buf, &ret[0], len);
    }
  }

  return rv;
}
}
bytes::bytes() : m_value(nullptr) {}

bytes::bytes(size_t len)
    : m_value(std::unique_ptr<byte_t[]>(new byte_t[len])) {}

bytes::bytes(const bytes &v) : bytes(v.size()) {
  memcpy(m_value.get(), v.m_value.get(), v.size());
}

bytes::bytes(bytes &&v) : m_value(std::move(v.m_value)) {}

bytes::bytes(const byte_t *buf, size_t buf_len) {
  if (buf_len > 0) {
    m_value = std::unique_ptr<byte_t[]>(new byte_t[buf_len]);
    memcpy(m_value.get(), buf, buf_len);
  }
}
bytes &bytes::operator=(const bytes &v) {
  m_value = std::unique_ptr<byte_t[]>(new byte_t[v.size()]);
  memcpy(m_value.get(), v.m_value.get(), v.size());
  return *this;
}

bytes &bytes::operator=(bytes &&v) {
  m_value = std::move(v.m_value);
  return *this;
}

bool bytes::operator==(const bytes &v) {
  if (v.size() != size())
    return false;
  return memcmp(v.value(), value(), size()) == 0;
}
bool bytes::operator!=(const bytes &v) { return !operator==(v); }

bytes bytes::from_base58(const std::string &t) {
  size_t len = 0;
  bool succ = internal::convert_base58_to_bytes(t, nullptr, len);
  if (!succ)
    throw std::invalid_argument("invalid base58 string for from_base58");
  bytes ret(len);
  internal::convert_base58_to_bytes(t, ret.value(), len);
  return ret;
}
bytes bytes::from_hex(const std::string &t) {
  size_t len = 0;
  bool succ = internal::convert_hex_to_bytes(t, nullptr, len);
  if (!succ) {
    throw std::invalid_argument("invalid hex string for from_hex");
  }
  bytes ret(len);
  internal::convert_hex_to_bytes(t, ret.value(), len);
  return ret;
}
}
}
