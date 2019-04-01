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

#include "common/address.h"
#include "common/base58.h"
namespace neb {
address_t base58_to_address(const base58_address_t &addr) {
  return bytes::from_base58(addr);
}
base58_address_t address_to_base58(const address_t &addr) {
  return addr.to_base58();
}

bool is_valid_address(const address_t &addr) {
  if (addr.size() != NAS_ADDRESS_LEN)
    return false;
  if (addr[0] != NAS_ADDRESS_MAGIC_NUM)
    return false;
  if (addr[1] != NAS_ADDRESS_CONTRACT_MAGIC_NUM &&
      addr[1] != NAS_ADDRESS_ACCOUNT_MAGIC_NUM)
    return false;
  return true;
}
bool is_contract_address(const address_t &addr) {
  if (is_valid_address(addr) && addr[1] == NAS_ADDRESS_CONTRACT_MAGIC_NUM)
    return true;
  return false;
}
bool is_normal_address(const address_t &addr) {
  if (is_valid_address(addr) && addr[1] == NAS_ADDRESS_ACCOUNT_MAGIC_NUM)
    return true;
  return false;
}
} // namespace neb
