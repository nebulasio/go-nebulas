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

#include "fs/blockchain/blockchain_api.h"
#include "fs/blockchain/trie/trie.h"
#include "fs/util.h"

namespace neb {
namespace fs {

blockchain_api::blockchain_api(blockchain *blockchain_ptr)
    : m_blockchain(blockchain_ptr) {}

std::shared_ptr<std::vector<transaction_info_t>>
blockchain_api::get_block_transactions_api(block_height_t height) {

  std::vector<transaction_info_t> txs;
  auto block = m_blockchain->load_block_with_height(height);

  int64_t timestamp = block->header().timestamp();

  for (auto &tx : block->transactions()) {
    transaction_info_t info;

    // TODO file status and gas_used not in proto transaction
    info.m_status = 1;
    info.m_gas_used = 0;

    info.m_from = tx.from();
    info.m_to = tx.to();
    info.m_tx_value = to_wei(neb::util::string_to_byte(tx.value()).to_hex());
    info.m_timestamp = timestamp;
    info.m_gas_price =
        to_wei(neb::util::string_to_byte(tx.gas_price()).to_hex());

    txs.push_back(info);
  }
  return std::make_shared<std::vector<transaction_info_t>>(txs);
}

std::shared_ptr<account_info_t>
blockchain_api::get_account_api(const address_t &addr, block_height_t height) {

  auto rs_ptr = m_blockchain->get_blockchain_storage();
  auto block = m_blockchain->load_block_with_height(height);

  // get block header account state
  std::string state_root_str = block->header().state_root();
  neb::util::bytes state_root_bytes = neb::util::string_to_byte(state_root_str);

  // get trie node
  trie t(rs_ptr);
  // TODO def address_t as neb::util::bytes
  auto trie_node_bytes =
      t.get_trie_node(state_root_bytes, neb::util::string_to_byte(addr));

  std::shared_ptr<corepb::Account> corepb_account_ptr =
      std::make_shared<corepb::Account>();
  bool ret = corepb_account_ptr->ParseFromArray(trie_node_bytes.value(),
                                                trie_node_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse corepb Account failed");
  }

  // get balance from pb Account
  std::string balance_str = corepb_account_ptr->balance();
  std::string hex_str = neb::util::string_to_byte(balance_str).to_hex();

  std::string address = corepb_account_ptr->address();

  return std::make_shared<account_info_t>(
      account_info_t{address, to_wei(hex_str)});
}
} // namespace fs
} // namespace neb
