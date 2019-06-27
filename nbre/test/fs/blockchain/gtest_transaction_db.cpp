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
#include "fs/blockchain.h"
#include "fs/blockchain/blockchain_api.h"
#include "fs/blockchain/transaction/transaction_algo.h"
#include "fs/blockchain/transaction/transaction_db.h"
#include "fs/rocksdb_session_storage.h"
#include "test/fs/gtest_common.h"
#include <gtest/gtest.h>

TEST(test_transaction_db, read_transactions_from_db_with_duration_normal) {
  std::string db_path = get_db_path_for_read();

  auto rss_ptr = std::make_unique<neb::fs::rocksdb_session_storage>();
  rss_ptr->init(db_path, neb::fs::storage_open_default);

  auto bc_ptr = std::make_unique<neb::fs::blockchain>(rss_ptr.get());
  auto bab_ptr = std::make_unique<neb::fs::blockchain_api>(bc_ptr.get());
  auto tdb_ptr = std::make_unique<neb::fs::transaction_db>(bab_ptr.get());

  auto txs = tdb_ptr->read_transactions_from_db_with_duration(12785, 12800);

  // clang-format off
  std::string transactions_str[] = {
    "12785,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,deploy,0,1531709985,23553,1000000",
    "12785,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,deploy,0,1531709985,22567,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1JKY3kFRuStzoUxMr2qpPfFJ3U3GkY3kNd,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1YWWM3z1p3CTh4yJ8d95iriP9PJkh44hZZ,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1R9YdiHz8mpfKod1gXGrhyC2YutratH4qK,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1WkTyCqxWDPoEM96eoNakKGjc4wazGwB3A,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1dkuk9d8X9JiDT5rS6KS7CTiGeZJp3oi8g,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1aWaeMwtxVTsrxWwSuogvj3FugJNp99vhZ,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1P3P3F4MQZsFRteaCGwY3bryCTjnmKzU4c,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1aPy22TQo9Ac9NbSbceQv9bLhhHwrnLpY4,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1Mxf1fpXAq4aM9juZy6w5JuqxZRyEwyQBD,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1FmrbppsF5xyahWMUw7jJtqCPbPUdff9sA,binary,1000000000000000,1531710000,20000,1000000",
    "12787,1,n1JKY3kFRuStzoUxMr2qpPfFJ3U3GkY3kNd,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1YWWM3z1p3CTh4yJ8d95iriP9PJkh44hZZ,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1R9YdiHz8mpfKod1gXGrhyC2YutratH4qK,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1WkTyCqxWDPoEM96eoNakKGjc4wazGwB3A,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1dkuk9d8X9JiDT5rS6KS7CTiGeZJp3oi8g,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1aWaeMwtxVTsrxWwSuogvj3FugJNp99vhZ,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1aPy22TQo9Ac9NbSbceQv9bLhhHwrnLpY4,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1Mxf1fpXAq4aM9juZy6w5JuqxZRyEwyQBD,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1FmrbppsF5xyahWMUw7jJtqCPbPUdff9sA,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1P3P3F4MQZsFRteaCGwY3bryCTjnmKzU4c,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000"
  };
  // clang-format on

  for (size_t i = 0; i < txs.size(); i++) {
    auto &tx = txs[i];
    std::stringstream ss;
    ss << tx.m_height << ',' << tx.m_status << ',' << tx.m_from.to_base58()
       << ',' << tx.m_to.to_base58() << ',' << tx.m_tx_type << ','
       << tx.m_tx_value << ',' << tx.m_timestamp << ',' << tx.m_gas_used << ','
       << tx.m_gas_price;
    EXPECT_EQ(ss.str(), transactions_str[i]);
  }
}

TEST(test_transaction_db, read_transaction_from_db_with_duration_invalid) {

  std::string db_path = get_db_path_for_read();

  auto rss_ptr = std::make_unique<neb::fs::rocksdb_session_storage>();
  rss_ptr->init(db_path, neb::fs::storage_open_default);

  auto bc_ptr = std::make_unique<neb::fs::blockchain>(rss_ptr.get());
  auto bab_ptr = std::make_unique<neb::fs::blockchain_api>(bc_ptr.get());
  auto tdb_ptr = std::make_unique<neb::fs::transaction_db>(bab_ptr.get());

  EXPECT_THROW(tdb_ptr->read_transactions_from_db_with_duration(-2, -1),
               neb::fs::storage_general_failure);

  auto txs = tdb_ptr->read_transactions_from_db_with_duration(-2, -2);
  EXPECT_TRUE(txs.empty());

  txs = tdb_ptr->read_transactions_from_db_with_duration(-1, -1);
  EXPECT_TRUE(txs.empty());

  txs = tdb_ptr->read_transactions_from_db_with_duration(0, 0);
  EXPECT_TRUE(txs.empty());

  txs = tdb_ptr->read_transactions_from_db_with_duration(1, 1);
  EXPECT_TRUE(txs.empty());

  txs = tdb_ptr->read_transactions_from_db_with_duration(10, 10);
  EXPECT_TRUE(txs.empty());

  txs = tdb_ptr->read_transactions_from_db_with_duration(-1, 0);
  EXPECT_TRUE(txs.empty());

  txs = tdb_ptr->read_transactions_from_db_with_duration(-1, 100);
  EXPECT_TRUE(txs.empty());

  txs = tdb_ptr->read_transactions_from_db_with_duration(-1, -123);
  EXPECT_TRUE(txs.empty());

  EXPECT_THROW(tdb_ptr->read_transactions_from_db_with_duration(10, -123),
               neb::fs::storage_general_failure);

  txs = tdb_ptr->read_transactions_from_db_with_duration(10, 5);
  EXPECT_TRUE(txs.empty());

  txs = tdb_ptr->read_transactions_from_db_with_duration(0, 1);
  EXPECT_TRUE(txs.empty());

  txs = tdb_ptr->read_transactions_from_db_with_duration(0, 2);
  EXPECT_TRUE(txs.empty());
}

TEST(test_transaction_db, read_transactions_with_address_type) {

  std::string db_path = get_db_path_for_read();

  auto rss_ptr = std::make_unique<neb::fs::rocksdb_session_storage>();
  rss_ptr->init(db_path, neb::fs::storage_open_default);

  auto bc_ptr = std::make_unique<neb::fs::blockchain>(rss_ptr.get());
  auto bab_ptr = std::make_unique<neb::fs::blockchain_api>(bc_ptr.get());
  auto tdb_ptr = std::make_unique<neb::fs::transaction_db>(bab_ptr.get());

  auto txs = tdb_ptr->read_transactions_from_db_with_duration(12785, 12800);

  // clang-format off
  std::string acc_to_acc_transactions_str[] = {
    "12785,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,deploy,0,1531709985,23553,1000000",
    "12785,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,deploy,0,1531709985,22567,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1JKY3kFRuStzoUxMr2qpPfFJ3U3GkY3kNd,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1YWWM3z1p3CTh4yJ8d95iriP9PJkh44hZZ,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1R9YdiHz8mpfKod1gXGrhyC2YutratH4qK,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1WkTyCqxWDPoEM96eoNakKGjc4wazGwB3A,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1dkuk9d8X9JiDT5rS6KS7CTiGeZJp3oi8g,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1aWaeMwtxVTsrxWwSuogvj3FugJNp99vhZ,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1P3P3F4MQZsFRteaCGwY3bryCTjnmKzU4c,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1aPy22TQo9Ac9NbSbceQv9bLhhHwrnLpY4,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1Mxf1fpXAq4aM9juZy6w5JuqxZRyEwyQBD,binary,1000000000000000,1531710000,20000,1000000",
    "12786,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1FmrbppsF5xyahWMUw7jJtqCPbPUdff9sA,binary,1000000000000000,1531710000,20000,1000000"
  };
  std::string acc_to_contract_transactions_str[] = {
    "12787,1,n1JKY3kFRuStzoUxMr2qpPfFJ3U3GkY3kNd,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1YWWM3z1p3CTh4yJ8d95iriP9PJkh44hZZ,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1R9YdiHz8mpfKod1gXGrhyC2YutratH4qK,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1WkTyCqxWDPoEM96eoNakKGjc4wazGwB3A,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1dkuk9d8X9JiDT5rS6KS7CTiGeZJp3oi8g,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1aWaeMwtxVTsrxWwSuogvj3FugJNp99vhZ,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1aPy22TQo9Ac9NbSbceQv9bLhhHwrnLpY4,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1Mxf1fpXAq4aM9juZy6w5JuqxZRyEwyQBD,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1FmrbppsF5xyahWMUw7jJtqCPbPUdff9sA,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000",
    "12787,1,n1P3P3F4MQZsFRteaCGwY3bryCTjnmKzU4c,n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S,call,1,1531710015,20121,1000000"
  };
  // clang-format on

  auto ret = neb::fs::algo::read_transactions_with_address_type(
      txs, NAS_ADDRESS_ACCOUNT_MAGIC_NUM, NAS_ADDRESS_ACCOUNT_MAGIC_NUM);
  for (size_t i = 0; i < ret.size(); i++) {
    auto &tx = ret[i];
    std::stringstream ss;
    ss << tx.m_height << ',' << tx.m_status << ',' << tx.m_from.to_base58()
       << ',' << tx.m_to.to_base58() << ',' << tx.m_tx_type << ','
       << tx.m_tx_value << ',' << tx.m_timestamp << ',' << tx.m_gas_used << ','
       << tx.m_gas_price;
    EXPECT_EQ(ss.str(), acc_to_acc_transactions_str[i]);
  }

  ret = neb::fs::algo::read_transactions_with_address_type(
      txs, NAS_ADDRESS_ACCOUNT_MAGIC_NUM, NAS_ADDRESS_CONTRACT_MAGIC_NUM);
  for (size_t i = 0; i < ret.size(); i++) {
    auto &tx = ret[i];
    std::stringstream ss;
    ss << tx.m_height << ',' << tx.m_status << ',' << tx.m_from.to_base58()
       << ',' << tx.m_to.to_base58() << ',' << tx.m_tx_type << ','
       << tx.m_tx_value << ',' << tx.m_timestamp << ',' << tx.m_gas_used << ','
       << tx.m_gas_price;
    EXPECT_EQ(ss.str(), acc_to_contract_transactions_str[i]);
  }

#define NAS_ADDRESS_INVALID_MAGIC_NUM 0xff
  ret = neb::fs::algo::read_transactions_with_address_type(
      txs, NAS_ADDRESS_INVALID_MAGIC_NUM, NAS_ADDRESS_ACCOUNT_MAGIC_NUM);
  EXPECT_TRUE(ret.empty());

  ret = neb::fs::algo::read_transactions_with_address_type(
      txs, NAS_ADDRESS_INVALID_MAGIC_NUM, NAS_ADDRESS_CONTRACT_MAGIC_NUM);
  EXPECT_TRUE(ret.empty());

  ret = neb::fs::algo::read_transactions_with_address_type(
      txs, NAS_ADDRESS_ACCOUNT_MAGIC_NUM, NAS_ADDRESS_INVALID_MAGIC_NUM);
  EXPECT_TRUE(ret.empty());

  ret = neb::fs::algo::read_transactions_with_address_type(
      txs, NAS_ADDRESS_CONTRACT_MAGIC_NUM, NAS_ADDRESS_INVALID_MAGIC_NUM);
  EXPECT_TRUE(ret.empty());
}

TEST(test_transaction_db, read_transactions_with_succ) {

  std::string db_path = get_db_path_for_read();

  auto rss_ptr = std::make_unique<neb::fs::rocksdb_session_storage>();
  rss_ptr->init(db_path, neb::fs::storage_open_default);

  auto bc_ptr = std::make_unique<neb::fs::blockchain>(rss_ptr.get());
  auto bab_ptr = std::make_unique<neb::fs::blockchain_api>(bc_ptr.get());
  auto tdb_ptr = std::make_unique<neb::fs::transaction_db>(bab_ptr.get());

  auto txs = tdb_ptr->read_transactions_from_db_with_duration(13690, 13692);

  // clang-format off
  std::string succ_transactions_str[] = {
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1NQRCT51iRNdHrtUCf2LdwncHmhFSXcHMf,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1GwxMufp9yK26Jttzb2t4nFjmur3o7metA,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1dDsPV3NEgxL2ARM9uJ3MYhsyc41NFSTMk,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1ZRrr7KcypiieJCAFgxSy6xP7XG1ug4CkS,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1YFj55m8uoQk3xqsSVxxyRbUMo6zsjpNDE,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1EdcEeJakeqpTLMRHGbt3Mkbz6Kbwy6Ymw,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1bp6LSrhaEhSXw9mEXbKsormF4Ff2yfnkS,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1cjAvpSiCdbSU1hDnepSoDj99FXj5aiQFE,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1ZQ5VDioNTj98JZ4nGyNBnpWSyvGRRmgdK,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1Kmg4YxqE7HFkJedMFwqpDhmvJf5Et43g5,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1EyWJDYA38i7Xqc7mtvpSQpyT3kLXcr57j,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1K9Y8ofh8R6HpRjX8mkoRj42GGeBDt2Gck,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1MqvAvVaAsk4Ky51Bh5oAZ21CFrXx5zt6v,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1KU1uGSZpWcHcoYaFAcSqJsm4AoYCH5LkK,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1T7fvSeUuU8xbmuZzbAyFY7hxb7WnFrkfs,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1EeF8L7oirH95sZpf6fPKuVJseNPmhdhm3,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1cVPpEcvoxX3qyvXqRBso3UdhM8NJxs8vd,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1XEY4fLnSvHsr6nitPecRPDagZSxyt6jjc,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1KSo7ekbZpm6gEVvHcgtQ9tJU5st4F9dXK,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1dpMTzFXXRkEBE37nzXvhSjQ7KXGqL5Y3N,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1R8x8CSMsBLXs6uxm8KVKywE6S8aD88BkD,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1bqWwAYrdxuViBCeA4LcnMe4xYUWXYtZp5,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1EnQEzxkqge8XsH7uFaVY5ahsM3srDhPhP,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1cUrg627XmKPccVak3gF1aakHqAdcB1PF5,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1HnVjj7mDZDx9G3ybMfScgSUscwV6V8KL3,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1cYsED3wzLbqeNnde7cPrt9EwVqu4BuSfY,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1HJ61dccMLZyWisy3v9b9zvLi7xzq1YN42,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1bUH87StyKnQnzBKti6x5LJwRpQ5WEBkgR,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1EzCQzD9hg4U5rNdExeKMKbY1r6ZRaqhsQ,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1WwPXtRnateipBKAzqS6HDZacgpRLW3eKY,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1FM45GD5Xq5L4CJjJhKf1SG5551asHrBWC,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1P4YQmHP2FLGNrzoehuFYp9gbPaRLNbNMY,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1QvRhgDtx1bFAFEN39MaQZp89o4sh1BAG9,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1F15rj8JhTeN2bPLPckh6EHYeLjkyz1aBB,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1Q7Yk1SPjqeSZKoNqwKg8WPqgj2etDCrnK,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1H8PZBuuW9vqQE66xm7hFmV8Go8EYN5g16,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1SzPNtdTuKhAfqyo1Cdy9VhYWYz4Z7fW8N,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1Q4uHGmFvxeqt3iCzZnzmSHnnKkW6inK2d,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1RCbuNRbVuxNhH2JZhwnBGE3Gzwd6HHXvD,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1MxBZwT8kJLsXtkboMrWGJgSuDWXntsVq3,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1QLMmX8JEncCR2jj4TYN3iso8kJgcRnTs8,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1bvDYkFcxV8gkgriYqyHqGVa5Z5vpXAfsp,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1NXKcSXsJouLcj9ycyLdkJiT3EP6dkGFST,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1Gx1hR3YfLP5Fc8kKJXXEs561cjvVaFGWv,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1KWEapwVkVx5cqPrV8f9WtPdHvZTGQYK5G,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1N25HeoB1Uci1VVLWUyXYEqTSVAL5vFjEV,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1KrzDxim8jjtBdGGJU8F3HMk79fVmQWtmi,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1STTN5DndkMWops1MWKgY3Wf1YHyrXEYMW,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1Emg4q43xrjySKtSU8AkW1y6WGT2KcfC9x,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1L2wARA12CppMu6RdgeczTSAUsneYwe1QQ,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1PZ3zgP8yHxGt5qsGR7N4G6SLoD2rBF1C2,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1T94a4ko9FBv3LCxGnJLbxBabDyiX2FXFu,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1TWLPisUy6UUsxTKPevyQEgq3Jew9rv9ww,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1ZKNtQZKCgdJfWmvzeR1vDuU5JErZXivYC,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1TScMSkMacEFUX2d37Cvxbs3hMSQzqLow9,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1cosmXCR1VycUu8HV9V6sHbTfcurEAzYCZ,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1dYJ5DPv9acCKS6wfEhFJtN6gdXpJoymeY,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1LTLN7VGToigJLRxfvti3v1ssBiDd3Uee8,binary,1000000000000000,1531726275,20000,1000000",
    "13690,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1Fw763xUpPqkiWU1zsqUVReG1ETzhJp9eU,binary,1000000000000000,1531726275,20000,1000000"
  };
  // clang-format on

  auto ret = neb::fs::algo::read_transactions_with_succ(txs);

  for (size_t i = 0; i < ret.size(); i++) {
    auto &tx = ret[i];
    std::stringstream ss;
    ss << tx.m_height << ',' << tx.m_status << ',' << tx.m_from.to_base58()
       << ',' << tx.m_to.to_base58() << ',' << tx.m_tx_type << ','
       << tx.m_tx_value << ',' << tx.m_timestamp << ',' << tx.m_gas_used << ','
       << tx.m_gas_price;
    EXPECT_EQ(ss.str(), succ_transactions_str[i]);
  }

  txs = tdb_ptr->read_transactions_from_db_with_duration(13691, 13692);
  ret = neb::fs::algo::read_transactions_with_succ(txs);
  EXPECT_TRUE(ret.empty());

  txs = tdb_ptr->read_transactions_from_db_with_duration(13691, 13691);
  ret = neb::fs::algo::read_transactions_with_succ(txs);
  EXPECT_TRUE(ret.empty());
}

