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

#include "common/address.h"
#include "common/common.h"
#include "common/math.h"
#include "fs/blockchain/transaction/transaction_db.h"

namespace neb {
namespace fs {

class account_db {
public:
  account_db(blockchain_api_base *blockchain_ptr);

  wei_t get_balance(const address_t &addr, block_height_t height);
  address_t get_contract_deployer(const address_t &address_t,
                                  block_height_t height);

  void set_height_address_val_internal(
      const std::vector<transaction_info_t> &txs,
      std::unordered_map<address_t, wei_t> &addr_balance);

  wei_t get_account_balance_internal(const address_t &addr,
                                     block_height_t height);

  static floatxx_t get_normalized_value(floatxx_t value);

private:
  std::unordered_map<address_t, std::vector<block_height_t>> m_addr_height_list;
  std::unordered_map<block_height_t, std::unordered_map<address_t, wei_t>>
      m_height_addr_val;
  blockchain_api_base *m_blockchain;
};
} // namespace fs
} // namespace neb
