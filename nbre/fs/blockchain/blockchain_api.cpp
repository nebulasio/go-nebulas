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
#include "common/nebulas_currency.h"
#include "fs/blockchain.h"
#include "fs/blockchain/trie/trie.h"
#include "fs/util.h"
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>

namespace neb {
namespace fs {
blockchain_api_base::blockchain_api_base(blockchain *bc) : m_blockchain(bc) {}

blockchain_api_base::~blockchain_api_base() {}

blockchain_api::blockchain_api(blockchain *bc) : blockchain_api_base(bc) {}

blockchain_api::~blockchain_api() {}

std::vector<transaction_info_t>
blockchain_api::get_block_transactions_api(block_height_t height) {

  std::vector<transaction_info_t> ret;
  // special for block height 1
  if (height <= 1) {
    return ret;
  }

  auto block = m_blockchain->load_block_with_height(height);
  int64_t timestamp = block->header().timestamp();

  std::string events_root_str = block->header().events_root();
  neb::bytes events_root_bytes = neb::string_to_byte(events_root_str);

  for (auto &tx : block->transactions()) {
    transaction_info_t info;

    info.m_height = height;
    info.m_timestamp = timestamp;

    info.m_from = to_address(tx.from());
    info.m_to = to_address(tx.to());
    info.m_tx_value = storage_to_wei(neb::string_to_byte(tx.value()));
    info.m_gas_price = storage_to_wei(neb::string_to_byte(tx.gas_price()));
    info.m_tx_type = tx.data().type();

    // get topic chain.transactionResult
    std::string tx_hash_str = tx.hash();
    neb::bytes tx_hash_bytes = neb::string_to_byte(tx_hash_str);
    std::vector<transaction_info_t> events;
    get_transfer_event(events_root_bytes, tx_hash_bytes, events, info);

    ret.push_back(info);
    if (info.m_status) {
      for (auto &e : events) {
        e.m_status = info.m_status;
        ret.push_back(e);
      }
    }
  }
  return ret;
}

std::unique_ptr<corepb::Account>
blockchain_api::get_account_api(const address_t &addr, block_height_t height) {

  std::unique_ptr<corepb::Block> block = std::make_unique<corepb::Block>();

  if (!height) {
    block = m_blockchain->load_LIB_block();
  } else {
    block = m_blockchain->load_block_with_height(height);
  }

  // get block header account state
  std::string state_root_str = block->header().state_root();
  neb::bytes state_root_bytes = neb::string_to_byte(state_root_str);

  // get trie node
  trie t(m_blockchain->storage());
  neb::bytes trie_node_bytes;
  bool is_found = t.get_trie_node(state_root_bytes, addr, trie_node_bytes);
  auto corepb_account_ptr = std::make_unique<corepb::Account>();
  if (!is_found) {
    corepb_account_ptr->set_address(std::to_string(addr));
    corepb_account_ptr->set_balance(std::to_string(neb::wei_to_storage(0)));
    return corepb_account_ptr;
  }

  bool ret = corepb_account_ptr->ParseFromArray(trie_node_bytes.value(),
                                                trie_node_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse corepb Account failed");
  }
  return corepb_account_ptr;
}

std::unique_ptr<corepb::Transaction>
blockchain_api::get_transaction_api(const bytes &tx_hash) {
  auto corepb_txs_ptr = std::make_unique<corepb::Transaction>();

  // suppose height is the latest irreversible block height
  auto block = m_blockchain->load_LIB_block();

  // get block header transaction root
  std::string txs_root_str = block->header().txs_root();
  neb::bytes txs_root_bytes = neb::string_to_byte(txs_root_str);

  // get trie node
  trie t(m_blockchain->storage());
  neb::bytes trie_node_bytes;
  bool ret = t.get_trie_node(txs_root_bytes, tx_hash, trie_node_bytes);
  if (!ret) {
    return corepb_txs_ptr;
  }

  ret = corepb_txs_ptr->ParseFromArray(trie_node_bytes.value(),
                                       trie_node_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse corepb Transaction failed");
  }
  return corepb_txs_ptr;
}

void blockchain_api::get_transfer_event(const neb::bytes &events_root,
                                        const neb::bytes &tx_hash,
                                        std::vector<transaction_info_t> &events,
                                        transaction_info_t &info) {

  for (int64_t id = 1;; id++) {
    neb::bytes id_bytes = neb::number_to_byte<neb::bytes>(id);
    neb::bytes events_tx_hash = tx_hash;
    events_tx_hash.append_bytes(id_bytes.value(), id_bytes.size());

    trie t(m_blockchain->storage());
    neb::bytes trie_node_bytes;
    bool ret = t.get_trie_node(events_root, events_tx_hash, trie_node_bytes);
    if (!ret) {
      break;
    }

    std::string json_str = neb::byte_to_string(trie_node_bytes);
    json_parse_event(json_str, events, info);
  }
}

void blockchain_api::json_parse_event(const std::string &json,
                                      std::vector<transaction_info_t> &events,
                                      transaction_info_t &info) {

  boost::property_tree::ptree pt;
  std::stringstream ss(json);
  boost::property_tree::read_json(ss, pt);

  std::string topic = pt.get<std::string>("Topic");

  if (topic.compare("chain.transactionResult") == 0) {
    std::string data_json = pt.get<std::string>("Data");
    ss = std::stringstream(data_json);
    boost::property_tree::read_json(ss, pt);

    int32_t status = pt.get<int32_t>("status");
    wei_t gas_used =
        boost::lexical_cast<wei_t>(pt.get<std::string>("gas_used"));
    info.m_status = status;
    info.m_gas_used = gas_used;
  }

  if (topic.compare("chain.transferFromContract") == 0) {
    std::string data_json = pt.get<std::string>("Data");
    ss = std::stringstream(data_json);
    boost::property_tree::read_json(ss, pt);

    boost::property_tree::ptree::const_assoc_iterator it = pt.find("error");
    if (it != pt.not_found()) {
      return;
    }

    std::string from = pt.get<std::string>("from");
    std::string to = pt.get<std::string>("to");
    std::string amount = pt.get<std::string>("amount");

    transaction_info_t event;
    event.m_height = info.m_height;
    event.m_timestamp = info.m_timestamp;
    event.m_from = neb::base58_to_address(from);
    event.m_to = neb::base58_to_address(to);
    event.m_tx_value = boost::lexical_cast<wei_t>(amount);
    event.m_gas_price = info.m_gas_price;
    event.m_tx_type = std::string("event");
    event.m_status = 1;
    event.m_gas_used = 0;

    events.push_back(event);
  }
}

} // namespace fs
} // namespace neb
