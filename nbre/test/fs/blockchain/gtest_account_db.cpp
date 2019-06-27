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

#include "common/address.h"
#include "common/configuration.h"
#include "fs/blockchain.h"
#include "fs/blockchain/account/account_db.h"
#include "fs/blockchain/blockchain_api.h"
#include "fs/blockchain/transaction/transaction_db.h"
#include "fs/blockchain/trie/trie.h"
#include "fs/rocksdb_session_storage.h"
#include "test/fs/gtest_common.h"
#include <gtest/gtest.h>
#include <random>

TEST(test_account_db, get_balance) {

  std::string db_path = get_db_path_for_read();

  auto rss_ptr = std::make_unique<neb::fs::rocksdb_session_storage>();
  rss_ptr->init(db_path, neb::fs::storage_open_default);

  auto bc_ptr = std::make_unique<neb::fs::blockchain>(rss_ptr.get());
  auto bab_ptr = std::make_unique<neb::fs::blockchain_api>(bc_ptr.get());
  auto adb_ptr = std::make_unique<neb::fs::account_db>(bab_ptr.get());

  auto block_ptr = bc_ptr->load_LIB_block();

  typedef std::tuple<std::string, neb::block_height_t, std::string> item_t;
  // std::vector<item_t> v;

  // clang-format off
  std::vector<item_t> vv{
    item_t{"n1HQb7j3Ctpj5A9sJnfhRwXS549nGPv2EkD",1,"0"},
    item_t{"n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m",12785,"9999999999999953880000000"},
    item_t{"n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m",12785,"9999999999999953880000000"},
    item_t{"n1JKY3kFRuStzoUxMr2qpPfFJ3U3GkY3kNd",12786,"1000000000000000"},
    item_t{"n1YWWM3z1p3CTh4yJ8d95iriP9PJkh44hZZ",12786,"1000000000000000"},
    item_t{"n1R9YdiHz8mpfKod1gXGrhyC2YutratH4qK",12786,"1000000000000000"},
    item_t{"n1WkTyCqxWDPoEM96eoNakKGjc4wazGwB3A",12786,"1000000000000000"},
    item_t{"n1dkuk9d8X9JiDT5rS6KS7CTiGeZJp3oi8g",12786,"1000000000000000"},
    item_t{"n1aWaeMwtxVTsrxWwSuogvj3FugJNp99vhZ",12786,"1000000000000000"},
    item_t{"n1P3P3F4MQZsFRteaCGwY3bryCTjnmKzU4c",12786,"1000000000000000"},
    item_t{"n1aPy22TQo9Ac9NbSbceQv9bLhhHwrnLpY4",12786,"1000000000000000"},
    item_t{"n1Mxf1fpXAq4aM9juZy6w5JuqxZRyEwyQBD",12786,"1000000000000000"},
    item_t{"n1FmrbppsF5xyahWMUw7jJtqCPbPUdff9sA",12786,"1000000000000000"},
    item_t{"n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S",12787,"10"},
    item_t{"n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S",12787,"10"},
    item_t{"n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S",12787,"10"},
    item_t{"n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S",12787,"10"},
    item_t{"n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S",12787,"10"},
    item_t{"n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S",12787,"10"},
    item_t{"n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S",12787,"10"},
    item_t{"n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S",12787,"10"},
    item_t{"n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S",12787,"10"},
    item_t{"n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S",12787,"10"}
  };
  // clang-format on
  auto it = vv.begin();
  // auto LIB_block_height = block_ptr->height();
  // LOG(INFO) << "LIB block height " << LIB_block_height;
  neb::block_height_t LIB_block_height = 12787;
  for (neb::block_height_t h = 1; h <= LIB_block_height; h++) {
    block_ptr = bc_ptr->load_block_with_height(h);
    for (auto &tx : block_ptr->transactions()) {
      auto to_bytes = neb::to_address(tx.to());

      item_t item;
      std::get<0>(item) = to_bytes.to_base58();
      std::get<1>(item) = h;
      std::get<2>(item) =
          boost::lexical_cast<std::string>(adb_ptr->get_balance(to_bytes, h));

      EXPECT_EQ(std::get<0>(item), std::get<0>(*it));
      EXPECT_EQ(std::get<1>(item), std::get<1>(*it));
      EXPECT_EQ(std::get<2>(item), std::get<2>(*it));
      it++;
      // v.push_back(item);
    }
  }

  // for (auto &e : v) {
  // std::cout << "item_t{\"" << std::get<0>(e) << "\"," << std::get<1>(e)
  //<< ",\"" << std::get<2>(e) << "\"}," << std::endl;
  //}

}

TEST(test_account_db, get_balance_invalid) {

  std::string db_path = get_db_path_for_read();

  auto rss_ptr = std::make_unique<neb::fs::rocksdb_session_storage>();
  rss_ptr->init(db_path, neb::fs::storage_open_default);

  auto bc_ptr = std::make_unique<neb::fs::blockchain>(rss_ptr.get());
  auto bab_ptr = std::make_unique<neb::fs::blockchain_api>(bc_ptr.get());
  auto adb_ptr = std::make_unique<neb::fs::account_db>(bab_ptr.get());

  EXPECT_EQ(adb_ptr->get_balance(neb::string_to_byte("123"), 11), 0);
  EXPECT_THROW(
      adb_ptr->get_balance(
          neb::string_to_byte("n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S"), 90000),
      neb::fs::storage_general_failure);
}

TEST(test_account_db, get_contract_deployer) {

  std::string db_path = get_db_path_for_read();

  auto rss_ptr = std::make_unique<neb::fs::rocksdb_session_storage>();
  rss_ptr->init(db_path, neb::fs::storage_open_default);

  auto bc_ptr = std::make_unique<neb::fs::blockchain>(rss_ptr.get());
  auto bab_ptr = std::make_unique<neb::fs::blockchain_api>(bc_ptr.get());
  auto tdb_ptr = std::make_unique<neb::fs::transaction_db>(bab_ptr.get());
  auto adb_ptr = std::make_unique<neb::fs::account_db>(bab_ptr.get());

  auto block = bc_ptr->load_LIB_block();
  auto LIB_height = block->height();

  auto txs = tdb_ptr->read_transactions_from_db_with_duration(1, LIB_height);
  std::unordered_set<std::string> deployer_s;
  std::unordered_set<std::string> contract_s;
  for (auto &tx : txs) {
    if (tx.m_tx_type == "deploy") {
      deployer_s.insert(tx.m_from.to_base58());
    } else if (tx.m_tx_type == "call") {
      contract_s.insert(tx.m_to.to_base58());
    }
  }

  for (auto &c : contract_s) {
    auto addr = adb_ptr->get_contract_deployer(neb::bytes::from_base58(c));
    EXPECT_TRUE(deployer_s.find(addr.to_base58()) != deployer_s.end());
    // std::cout << "std::make_pair(\"" << c << "\",\"" << addr.to_base58()
    //<< "\")," << std::endl;
  }

  // clang-format off
  std::vector<std::pair<std::string, std::string>> contract_deployer({
    std::make_pair("n1mWX7JAbzZUuDjKB61NF5aPCADrY3cvkst","n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m"),
    std::make_pair("n1r6XsVSD7eKrNWqpBsqXAVq3ZhVu8pWQ2m","n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m"),
    std::make_pair("n1nT22NaMhH9GfQCauKdXQ9CruQ72uuosmR","n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m"),
    std::make_pair("n1xaW17cS9DTzPksmPHzhaoaC9yGfCFQy4C","n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m"),
    std::make_pair("n1pohg1UWjFFEKb9XnqGRdscabcBrT7WbaF","n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m"),
    std::make_pair("n1xfspHUmwzvx5xZJQpQ2xC52NzRhxekkFr","n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m"),
    std::make_pair("n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S","n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m"),
    std::make_pair("n1mLSCnaEu2EVykhzGvSug9vf35di213KY2","n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m"),
    std::make_pair("n1sNXeyJhsnknTfYiRQtFsS2epwUp3Ls9KH","n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m"),
    std::make_pair("n1mg8VTQJZ4kGDnDrggQnJzn4v5bDpo8bzy","n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m"),
    std::make_pair("n1koYepMB2DD6eFu6xfZvXQ7NstQ6FXeb69","n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m")
  });
  // clang-format on

  for (auto &item : contract_deployer) {
    auto contract_addr = neb::bytes::from_base58(item.first);
    auto deployer_addr = neb::bytes::from_base58(item.second);
    auto addr = adb_ptr->get_contract_deployer(contract_addr);
    EXPECT_EQ(addr, deployer_addr);
  }
}

TEST(test_account_db, get_balance_internal) {

  std::string db_path = get_db_path_for_read();

  auto rss_ptr = std::make_unique<neb::fs::rocksdb_session_storage>();
  rss_ptr->init(db_path, neb::fs::storage_open_default);

  auto bc_ptr = std::make_unique<neb::fs::blockchain>(rss_ptr.get());
  auto bab_ptr = std::make_unique<neb::fs::blockchain_api>(bc_ptr.get());
  auto tdb_ptr = std::make_unique<neb::fs::transaction_db>(bab_ptr.get());
  auto adb_ptr = std::make_unique<neb::fs::account_db>(bab_ptr.get());

  auto block_ptr = bc_ptr->load_LIB_block();
  std::unordered_set<neb::address_t> addr_s;
  std::vector<neb::address_t> addr_v;

  for (neb::block_height_t h = 1; h <= block_ptr->height(); h++) {
    auto block = bc_ptr->load_block_with_height(h);
    for (auto &tx : block->transactions()) {
      addr_s.insert(neb::string_to_byte(tx.from()));
      addr_s.insert(neb::string_to_byte(tx.to()));
    }
  }

  neb::block_height_t start_block = 12785;
  std::unordered_map<neb::address_t, neb::wei_t> addr_balance;
  for (auto &addr : addr_s) {
    auto balance = adb_ptr->get_balance(addr, start_block);
    addr_balance.insert(std::make_pair(addr, balance));
    addr_v.push_back(addr);
  }

  auto txs = tdb_ptr->read_transactions_from_db_with_duration(
      start_block, block_ptr->height());
  adb_ptr->update_height_address_val_internal(start_block, txs, addr_balance);

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(start_block, block_ptr->height());

  std::srand(time(0));
  int32_t addr_cnt = 10;

  while (addr_cnt--) {
    size_t index = random() % addr_v.size();
    auto addr = addr_v[index];

    int32_t height_cnt = 10;
    while (height_cnt--) {
      auto h = dis(mt);
      auto balance_expect = adb_ptr->get_balance(addr, h);
      auto balance_actual = adb_ptr->get_account_balance_internal(addr, h);
      // std::cout << h << ',' << addr.to_base58() << ',' << balance_actual <<
      // ','
      //<< balance_expect << std::endl;
      EXPECT_EQ(balance_actual, balance_expect);
    }
  }
}
