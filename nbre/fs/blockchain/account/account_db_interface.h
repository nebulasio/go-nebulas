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
#include "common/address.h"
#include "common/common.h"
#include "fs/blockchain/data_type.h"

namespace neb {
namespace fs {

class account_db_interface {
public:
  virtual neb::wei_t get_balance(const neb::address_t &addr,
                                 const neb::block_height_t &height) = 0;

  virtual neb::address_t get_contract_deployer(const neb::address_t &addr) = 0;

  virtual void update_height_address_val_internal(
      const neb::block_height_t &start_block,
      const std::vector<neb::fs::transaction_info_t> &txs,
      std::unordered_map<neb::address_t, neb::wei_t> &addr_balance) = 0;

  virtual neb::wei_t
  get_account_balance_internal(const neb::address_t &addr,
                               const neb::block_height_t &height) = 0;
};

} // namespace fs
} // namespace neb
