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
#include "fs/blockchain/blockchain_api.h"
#include "util/lru_cache.h"

namespace neb {
namespace fs {
class blockchain_api_test : public blockchain_api_base {
public:
  blockchain_api_test();
  virtual ~blockchain_api_test();
  virtual std::vector<transaction_info_t>
  get_block_transactions_api(block_height_t height);

  virtual std::unique_ptr<corepb::Account>
  get_account_api(const address_t &addr, block_height_t height = 0);

  virtual std::unique_ptr<corepb::Transaction>
  get_transaction_api(const bytes &tx_hash);

protected:
  std::shared_ptr<corepb::Block> get_block_with_height(block_height_t height);
  std::shared_ptr<corepb::Block> get_LIB_block();

protected:
  util::lru_cache<block_height_t, std::shared_ptr<corepb::Block>> m_block_cache;
};
} // namespace fs
} // namespace neb
