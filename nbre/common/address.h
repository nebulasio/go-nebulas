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

namespace neb {
typedef std::string base58_address_t;
typedef neb::bytes address_t;

address_t base58_to_address(const base58_address_t &addr);
base58_address_t address_to_base58(const address_t &addr);

typedef std::string address_bytes_t;
// typedef std::tuple<module_t, address_bytes_t, block_height_t, block_height_t>
// auth_row_t;
// typedef std::vector<auth_row_t> auth_table_t;

inline address_t to_address(const std::string &addr) {
  return string_to_byte(addr);
}
inline std::string address_to_string(const address_t &addr) {
  return byte_to_string(addr);
}

bool is_valid_address(const address_t &addr);
bool is_contract_address(const address_t &addr);
bool is_normal_address(const address_t &addr);

#define NAS_ADDRESS_LEN 26
#define NAS_ADDRESS_MAGIC_NUM 0x19
#define NAS_ADDRESS_ACCOUNT_MAGIC_NUM 0x57
#define NAS_ADDRESS_CONTRACT_MAGIC_NUM 0x58
} // namespace neb
