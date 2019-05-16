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
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the // GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//

#pragma once
#include "fs/blockchain/account/account_db.h"

namespace neb {
namespace fs {

class account_db_v2 {
public:
  account_db_v2(neb::fs::account_db *adb_ptr);

  neb::wei_t get_balance(const neb::address_t &addr,
                         neb::block_height_t height);
  neb::address_t get_contract_deployer(const neb::address_t &addr,
                                       neb::block_height_t height);

  void update_height_address_val_internal(
      neb::block_height_t start_block,
      const std::vector<neb::fs::transaction_info_t> &txs,
      std::unordered_map<neb::address_t, neb::wei_t> &addr_balance);

  neb::wei_t get_account_balance_internal(const neb::address_t &addr,
                                          neb::block_height_t height);

  static neb::floatxx_t get_normalized_value(neb::floatxx_t value);

private:
  void init_height_address_val_internal(
      neb::block_height_t start_block,
      const std::unordered_map<neb::address_t, neb::wei_t> &addr_balance);

private:
  std::unordered_map<neb::address_t, std::vector<neb::block_height_t>>
      m_addr_height_list;
  std::unordered_map<neb::block_height_t,
                     std::unordered_map<neb::address_t, neb::wei_t>>
      m_height_addr_val;
  neb::fs::account_db *m_adb_ptr;
};

} // namespace fs
} // namespace neb
