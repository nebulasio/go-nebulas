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
#include "fs/blockchain/account/account_db_interface.h"

namespace neb {
namespace fs {

class blockchain_api_base;
class account_db : public account_db_interface {
public:
  account_db(neb::fs::blockchain_api_base *blockchain_ptr);

  virtual neb::wei_t get_balance(const neb::address_t &addr,
                                 const neb::block_height_t &height);
  virtual neb::address_t get_contract_deployer(const neb::address_t &addr);

  virtual void update_height_address_val_internal(
      const neb::block_height_t &start_block,
      const std::vector<neb::fs::transaction_info_t> &txs,
      std::unordered_map<neb::address_t, neb::wei_t> &addr_balance);

  virtual neb::wei_t
  get_account_balance_internal(const neb::address_t &addr,
                               const neb::block_height_t &height);

protected:
  void init_height_address_val_internal(
      const neb::block_height_t &start_block,
      const std::unordered_map<neb::address_t, neb::wei_t> &addr_balance);

protected:
  std::unordered_map<neb::address_t, std::vector<neb::block_height_t>>
      m_addr_height_list;
  std::unordered_map<neb::block_height_t,
                     std::unordered_map<neb::address_t, neb::wei_t>>
      m_height_addr_val;
  blockchain_api_base *m_blockchain;
};

} // namespace fs
} // namespace neb
