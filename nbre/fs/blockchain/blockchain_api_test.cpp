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
#include "fs/blockchain/blockchain_api_test.h"
#include "common/nebulas_currency.h"
#include "fs/blockchain.h"
#include "util/bc_generator.h"

namespace neb {
namespace fs {

blockchain_api_test::blockchain_api_test() : blockchain_api_base(nullptr) {}

blockchain_api_test::~blockchain_api_test() {}

std::vector<transaction_info_t>
blockchain_api_test::get_block_transactions_api(block_height_t height) {
  std::vector<transaction_info_t> ret;

  auto block = get_block_with_height(height);

  for (int i = 0; i < block->transactions_size(); ++i) {
    corepb::Transaction tx = block->transactions(i);

    transaction_info_t info;
    info.m_height = height;
    info.m_status = 1; // All generated tx are succ
    info.m_from = to_address(tx.from());
    info.m_to = to_address(tx.to());
    info.m_tx_value = storage_to_wei(string_to_byte(tx.value()));
    info.m_timestamp = tx.timestamp();
    ret.push_back(info);
  }

  return ret;
}

std::unique_ptr<corepb::Account>
blockchain_api_test::get_account_api(const address_t &addr,
                                     block_height_t height) {
  auto ret = std::make_unique<corepb::Account>();

  std::vector<std::shared_ptr<corepb::Account>> accounts;

  if (!height) {
    accounts = util::generate_block::read_accounts_in_LIB();
  } else {
    accounts = util::generate_block::read_accounts_in_height(height);
  }

  bool found_flag = false;
  for (auto &ap : accounts) {
    address_t ap_addr = to_address(ap->address());
    if (ap_addr == addr) {
      *ret = *ap;
      found_flag = true;
      break;
    }
  }
  if (!found_flag) {
    ret->set_address(std::to_string(addr));
    ret->set_balance(std::to_string(neb::wei_to_storage(0)));
  }

  return ret;
}

std::unique_ptr<corepb::Transaction>
blockchain_api_test::get_transaction_api(const bytes &tx_hash) {
  auto ret = std::make_unique<corepb::Transaction>();

  auto block = get_LIB_block();

  for (int i = 0; i < block->transactions_size(); ++i) {
    corepb::Transaction tx = block->transactions(i);
    if (tx.hash() == neb::byte_to_string(tx_hash)) {
      *ret = tx;
      break;
    }
  }

  return ret;
}

std::shared_ptr<corepb::Block> blockchain_api_test::get_LIB_block() {
  std::shared_ptr<corepb::Block> ret = util::generate_block::read_LIB_block();
  auto height = ret->height();
  m_block_cache.set(height, ret);
  return ret;
}

std::shared_ptr<corepb::Block>
blockchain_api_test::get_block_with_height(block_height_t height) {
  std::shared_ptr<corepb::Block> ret;
  if (m_block_cache.get(height, ret)) {
    return ret;
  } else {
    ret = util::generate_block::read_block_with_height(height);
    m_block_cache.set(height, ret);
    return ret;
  }
}
} // namespace fs
} // namespace neb
