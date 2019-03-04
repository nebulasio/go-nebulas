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

#include "common/configuration.h"
#include "fs/bc_storage_session.h"
#include "fs/blockchain.h"
#include "fs/blockchain/account/account_db.h"
#include "fs/blockchain/blockchain_api.h"
#include "fs/blockchain/trie/trie.h"
#include "fs/util.h"
#include "util/nebulas_currency.h"

#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>
#include <random>

void json_parse_event(const std::string &json) {
  boost::property_tree::ptree pt;
  std::stringstream ss(json);
  try {
    boost::property_tree::read_json(ss, pt);
  } catch (boost::property_tree::ptree_error &e) {
    return;
  }

  std::string topic = pt.get<std::string>("Topic");
  assert(topic.compare("chain.transactionResult") == 0);

  std::string data_json = pt.get<std::string>("Data");
  ss = std::stringstream(data_json);
  try {
    boost::property_tree::read_json(ss, pt);
  } catch (boost::property_tree::ptree_error &e) {
    return;
  }

  int32_t status = pt.get<int32_t>("status");
  std::string gas_used = pt.get<std::string>("gas_used");
  LOG(INFO) << status << ',' << gas_used;
}

void trie_event(const neb::block_height_t start_block,
                const neb::block_height_t end_block) {

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(start_block, end_block);

  std::string neb_db = neb::configuration::instance().neb_db_dir();
  neb::fs::bc_storage_session::instance().init(
      neb_db, neb::fs::storage_open_for_readonly);
  neb::fs::trie t;

  std::vector<neb::block_height_t> v{end_block};

  while (true) {
    // for (auto h : v) {
    // get block
    neb::block_height_t h = dis(mt);
    neb::block_height_t height = h;
    std::shared_ptr<corepb::Block> block = std::make_shared<corepb::Block>();
    neb::util::bytes height_hash =
        neb::fs::bc_storage_session::instance().get_bytes(
            neb::util::number_to_byte<neb::util::bytes>(height));
    neb::util::bytes block_bytes =
        neb::fs::bc_storage_session::instance().get_bytes(height_hash);

    bool ret = block->ParseFromArray(block_bytes.value(), block_bytes.size());
    if (!ret) {
      throw std::runtime_error("parse block failed");
    }

    for (auto &tx : block->transactions()) {

      std::string tx_hash_str = tx.hash();
      neb::util::bytes tx_hash_bytes = neb::util::string_to_byte(tx_hash_str);

      // get block header event hash
      std::string events_root_str = block->header().events_root();
      neb::util::bytes events_root_bytes =
          neb::util::string_to_byte(events_root_str);

      // get block header txs hash
      std::string txs_root_str = block->header().txs_root();
      neb::util::bytes txs_root_bytes = neb::util::string_to_byte(txs_root_str);

      neb::util::bytes trie_node_bytes;

      // traversing event list for specific transaction
      // and get topic chain.transactionResult

      neb::util::bytes txs_result;
      for (int64_t id = 1;; id++) {
        neb::util::bytes id_bytes =
            neb::util::number_to_byte<neb::util::bytes>(id);
        // LOG(INFO) << "tx hash hex: " << tx_hash_bytes.to_hex();
        neb::util::bytes events_tx_hash_bytes = tx_hash_bytes;
        events_tx_hash_bytes.append_bytes(id_bytes.value(), id_bytes.size());
        // LOG(INFO) << "tx hash append id: " << tx_hash_bytes.to_hex();

        // ret = t.get_trie_node(txs_root_bytes, tx_hash_bytes,
        // trie_node_bytes);
        ret = t.get_trie_node(events_root_bytes, events_tx_hash_bytes,
                              trie_node_bytes);
        if (!ret) {
          break;
        }
        txs_result = trie_node_bytes;
      }
      // LOG(INFO) << neb::util::byte_to_string(txs_result);
      json_parse_event(neb::util::byte_to_string(txs_result));

      // assert(ret == true);
    }
  }
}

std::string get_stdout_from_command(std::string &cmd) {
  std::string data;
  FILE *stream;
  const int max_buffer = 256;
  char buffer[max_buffer];
  cmd.append(" 2>&1");

  stream = popen(cmd.c_str(), "r");
  if (stream) {
    while (!feof(stream)) {
      if (fgets(buffer, max_buffer, stream) != NULL) {
        data.append(buffer);
      }
    }
    pclose(stream);
  }
  return data;
}

std::vector<std::string> split_by_comma(const std::string &str, char comma) {
  std::vector<std::string> v;
  std::stringstream ss(str);
  std::string token;

  while (getline(ss, token, comma)) {
    v.push_back(token);
  }
  return v;
}

std::pair<std::string, int> json_parse_account_state(const std::string &json) {
  boost::property_tree::ptree pt;
  std::stringstream ss(json);

  try {
    boost::property_tree::read_json(ss, pt);
  } catch (boost::property_tree::ptree_error &e) {
    return std::make_pair("", -1);
  }

  boost::property_tree::ptree result = pt.get_child("result");
  std::string balance = result.get<std::string>("balance");
  int type = result.get<int>("type");
  return std::make_pair(balance, type);
}

std::pair<std::string, int> get_account_state(const std::string &address,
                                              neb::block_height_t height) {
  std::string cmd =
      "curl -s -H 'Content-Type: application/json' -X POST "
      "http://localhost:8685/v1/user/accountstate -d '{\"height\": " +
      std::to_string(height) + ", \"address\": \"" + address + "\"}'";
  if (height == 0) {
    cmd = "curl -s -H 'Content-Type: application/json' -X POST "
          "http://localhost:8685/v1/user/accountstate -d '{\"address\": \"" +
          address + "\"}'";
  }
  // std::cout << cmd << std::endl;
  std::string ret = get_stdout_from_command(cmd);
  std::vector<std::string> v = split_by_comma(ret, '\n');

  if (v.empty()) {
    return std::make_pair("", -1);
  }
  std::string resp = v[v.size() - 1];

  return json_parse_account_state(resp);
}

void trie_balance(const neb::block_height_t start_block,
                  const neb::block_height_t end_block) {

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(start_block, end_block);

  std::string neb_db = neb::configuration::instance().neb_db_dir();
  neb::fs::blockchain_api ba;

  std::vector<std::string> v{"n1KxWR8ycXg7Kb9CPTtNjTTEpvka269PniB"};

  while (true) {
    neb::block_height_t h = dis(mt);
    for (auto addr : v) {
      std::string balance_expect_str = get_account_state(addr, h).first;
      neb::wei_t balance_expect =
          boost::lexical_cast<neb::wei_t>(balance_expect_str);

      auto addr_str = neb::base58_to_address(addr);
      auto corepb_account_ptr = ba.get_account_api(addr_str, h);
      std::string balance_str = corepb_account_ptr->balance();
      neb::wei_t balance_actual =
          neb::storage_to_wei(neb::util::string_to_byte(balance_str));

      LOG(INFO) << addr << ',' << h << " expect:" << balance_expect
                << " actual:" << balance_actual;
      assert(balance_expect == balance_actual);
    }
  }
}

void trie_contract_deployer() {

  std::string neb_db = std::getenv("NEB_DB_DIR");
  neb::fs::blockchain_api ba;
  neb::fs::account_db ad(&ba);

  std::vector<std::string> v{"n1g6JZsQS1uRUySdwvuFJ7FYT4dFoyoSN5q"};
  for (auto &addr : v) {
    auto contract_bytes = neb::base58_to_address(addr);
    auto addr_str = ad.get_contract_deployer(contract_bytes, 1000000);
    LOG(INFO) << std::to_string(addr_str);
  }
}

int main(int argc, char *argv[]) {
  // trie_balance(std::stoll(argv[1]), std::stoll(argv[2]));
  // trie_event(std::stoll(argv[1]), std::stoll(argv[2]));
  trie_contract_deployer();
  return 0;
}
