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

#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>

namespace neb {
namespace fs {

blockchain_api::blockchain_api(blockchain *blockchain_ptr)
    : m_blockchain(blockchain_ptr) {}

std::unique_ptr<std::vector<transaction_info_t>>
blockchain_api::get_block_transactions_api(block_height_t height) {

  // special for  block height 1
  if (height <= 1) {
    return std::make_unique<std::vector<transaction_info_t>>();
  }

  std::vector<transaction_info_t> txs;
  auto block = m_blockchain->load_block_with_height(height);

  int64_t timestamp = block->header().timestamp();

  std::string events_root_str = block->header().events_root();
  neb::util::bytes events_root_bytes =
      neb::util::string_to_byte(events_root_str);

  for (auto &tx : block->transactions()) {
    transaction_info_t info;

    info.m_height = height;
    info.m_timestamp = timestamp;

    info.m_from = tx.from();
    info.m_to = tx.to();
    info.m_tx_value = to_wei(neb::util::string_to_byte(tx.value()).to_hex());
    info.m_gas_price =
        to_wei(neb::util::string_to_byte(tx.gas_price()).to_hex());

    // get topic chain.transactionResult
    std::string tx_hash_str = tx.hash();
    neb::util::bytes tx_hash_bytes = neb::util::string_to_byte(tx_hash_str);
    auto txs_result_ptr =
        get_transaction_result_api(events_root_bytes, tx_hash_bytes);

    info.m_status = txs_result_ptr->m_status;
    info.m_gas_used = txs_result_ptr->m_gas_used;

    txs.push_back(info);
  }
  return std::make_unique<std::vector<transaction_info_t>>(txs);
}

std::unique_ptr<event_info_t>
blockchain_api::get_transaction_result_api(const neb::util::bytes &events_root,
                                           const neb::util::bytes &tx_hash) {

  auto rs_ptr = m_blockchain->get_blockchain_storage();
  trie t(rs_ptr);
  neb::util::bytes txs_result;

  for (int64_t id = 1;; id++) {
    neb::util::bytes id_bytes = neb::util::number_to_byte<neb::util::bytes>(id);
    neb::util::bytes events_tx_hash = tx_hash;
    events_tx_hash.append_bytes(id_bytes.value(), id_bytes.size());

    neb::util::bytes trie_node_bytes;
    bool ret = t.get_trie_node(events_root, events_tx_hash, trie_node_bytes);
    if (!ret) {
      break;
    }
    txs_result = trie_node_bytes;
  }
  assert(!txs_result.empty());

  std::string json_str = neb::util::byte_to_string(txs_result);

  return json_parse_event(json_str);
}

std::unique_ptr<event_info_t>
blockchain_api::json_parse_event(const std::string &json) {
  boost::property_tree::ptree pt;
  std::stringstream ss(json);
  boost::property_tree::read_json(ss, pt);

  std::string topic = pt.get<std::string>("Topic");
  assert(topic.compare("chain.transactionResult") == 0);

  std::string data_json = pt.get<std::string>("Data");
  ss = std::stringstream(data_json);
  boost::property_tree::read_json(ss, pt);

  int32_t status = pt.get<int32_t>("status");
  wei_t gas_used = boost::lexical_cast<wei_t>(pt.get<std::string>("gas_used"));

  return std::make_unique<event_info_t>(event_info_t{status, gas_used});
}

std::unique_ptr<corepb::Account>
blockchain_api::get_account_api(const address_t &addr, block_height_t height) {

  auto rs_ptr = m_blockchain->get_blockchain_storage();
  auto block = m_blockchain->load_block_with_height(height);

  // get block header account state
  std::string state_root_str = block->header().state_root();
  neb::util::bytes state_root_bytes = neb::util::string_to_byte(state_root_str);

  // get trie node
  trie t(rs_ptr);
  neb::util::bytes addr_bytes = neb::util::string_to_byte(addr);
  // TODO def address_t as neb::util::bytes
  neb::util::bytes trie_node_bytes;
  bool is_found =
      t.get_trie_node(state_root_bytes, addr_bytes, trie_node_bytes);
  if (!is_found) {
    std::make_unique<account_info_t>(account_info_t{addr, 0});
  }

  std::unique_ptr<corepb::Account> corepb_account_ptr =
      std::make_unique<corepb::Account>();
  bool ret = corepb_account_ptr->ParseFromArray(trie_node_bytes.value(),
                                                trie_node_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse corepb Account failed");
  }
  return std::move(corepb_account_ptr);
}

std::unique_ptr<corepb::Transaction>
blockchain_api::get_transaction_api(const std::string &tx_hash,
                                    block_height_t height) {
  std::unique_ptr<corepb::Transaction> corepb_txs_ptr =
      std::make_unique<corepb::Transaction>();

  auto rs_ptr = m_blockchain->get_blockchain_storage();
  // suppose height is the latest block height
  auto block = m_blockchain->load_block_with_height(height);

  // get block header transaction root
  std::string txs_root_str = block->header().txs_root();
  neb::util::bytes txs_root_bytes = neb::util::string_to_byte(txs_root_str);

  // get trie node
  trie t(rs_ptr);
  neb::util::bytes tx_hash_bytes = neb::util::string_to_byte(tx_hash);
  neb::util::bytes trie_node_bytes;
  bool ret = t.get_trie_node(txs_root_bytes, tx_hash_bytes, trie_node_bytes);
  if (!ret) {
    return std::move(corepb_txs_ptr);
  }

  ret = corepb_txs_ptr->ParseFromArray(trie_node_bytes.value(),
                                       trie_node_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse corepb Transaction failed");
  }
  return std::move(corepb_txs_ptr);
}
} // namespace fs
} // namespace neb
