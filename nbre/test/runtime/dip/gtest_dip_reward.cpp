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

#include "common/common.h"
#include "runtime/dip/dip_reward.h"
#include "runtime/util.h"
#include <gtest/gtest.h>
#include <random>
#define PRECESION 1e-5

template <typename T> T precesion(const T &x, float pre = PRECESION) {
  return std::fabs(T(x * pre));
}

std::vector<std::shared_ptr<neb::rt::dip::dip_info_t>>
gen_dip_infos(std::vector<std::pair<std::string, std::string>> &meta) {

  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(0, std::numeric_limits<int16_t>::max());

  std::vector<std::shared_ptr<neb::rt::dip::dip_info_t>> ret;
  meta.push_back(std::make_pair("start_height", std::to_string(dis(mt))));
  meta.push_back(std::make_pair("end_height", std::to_string(dis(mt))));
  meta.push_back(std::make_pair("version", std::to_string(dis(mt))));
  int32_t infos_size = std::sqrt(dis(mt));
  for (int32_t i = 0; i < infos_size; i++) {
    auto info_ptr =
        std::shared_ptr<neb::rt::dip::dip_info_t>(new neb::rt::dip::dip_info_t{
            neb::to_address(std::to_string(uint32_t(std::pow(dis(mt), 0.3)))),
            neb::to_address(std::to_string(uint32_t(std::pow(dis(mt), 0.3)))),
            std::to_string(dis(mt))});
    ret.push_back(info_ptr);
  }
  return ret;
}

TEST(test_runtime_dip_reward, json_seri_deseri) {
  neb::rt::dip::dip_ret_type dip_ret;
  std::get<0>(dip_ret) = 1;
  std::vector<std::pair<std::string, std::string>> meta;
  auto &ret = std::get<2>(dip_ret);
  ret = gen_dip_infos(meta);
  std::get<1>(dip_ret) = neb::rt::meta_info_to_json(meta);
  auto str_ptr = neb::rt::dip::dip_reward::dip_info_to_json(dip_ret);
  dip_ret = neb::rt::dip::dip_reward::json_to_dip_info(*str_ptr);
  auto &info_v = std::get<2>(dip_ret);
  EXPECT_EQ(ret.size(), info_v.size());

  for (size_t i = 0; i < ret.size(); i++) {
    EXPECT_EQ(ret[i]->m_deployer, info_v[i]->m_deployer);
    EXPECT_EQ(ret[i]->m_contract, info_v[i]->m_contract);
    EXPECT_EQ(ret[i]->m_reward, info_v[i]->m_reward);
  }
}

TEST(test_runtime_dip_reward, back_to_coinbase) {
  std::vector<std::shared_ptr<neb::rt::dip::dip_info_t>> infos;
  neb::floatxx_t reward_left(0);
  neb::address_t coinbase_addr = neb::to_address(std::string());
  neb::rt::dip::dip_reward::back_to_coinbase(infos, reward_left, coinbase_addr);
  EXPECT_TRUE(infos.empty());

  infos.clear();
  reward_left = 1;
  neb::rt::dip::dip_reward::back_to_coinbase(infos, reward_left, coinbase_addr);
  EXPECT_TRUE(infos.empty());

  infos.clear();
  coinbase_addr = neb::to_address(std::string("a"));
  neb::rt::dip::dip_reward::back_to_coinbase(infos, reward_left, coinbase_addr);
  EXPECT_TRUE(!infos.empty());

  infos.clear();
  reward_left = 0;
  neb::rt::dip::dip_reward::back_to_coinbase(infos, reward_left, coinbase_addr);
  EXPECT_TRUE(infos.empty());

  std::vector<std::pair<std::string, std::string>> meta;
  auto ret = gen_dip_infos(meta);
  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(1, std::numeric_limits<int16_t>::max());

  infos.clear();
  infos = ret;
  size_t infos_size = infos.size();
  reward_left = dis(mt);
  coinbase_addr = std::pow(dis(mt), 0.3);
  neb::rt::dip::dip_reward::back_to_coinbase(infos, reward_left, coinbase_addr);
  EXPECT_TRUE(!infos.empty());
  EXPECT_EQ(infos_size + 1, infos.size());
}

TEST(test_runtime_dip_reward, ignore_account_transfer_contract) {
  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(1, std::numeric_limits<int16_t>::max());
  size_t txs_size = std::sqrt(dis(mt));

  std::vector<std::string> tx_type_v({"binary", "call", "deploy", "protocol"});
  std::unordered_map<std::string, int32_t> tx_type_cnt;
  std::vector<neb::fs::transaction_info_t> txs;

  for (size_t i = 0; i < txs_size; i++) {
    neb::fs::transaction_info_t info;
    info.m_tx_type = tx_type_v[dis(mt) % tx_type_v.size()];
    txs.push_back(info);
    tx_type_cnt[info.m_tx_type]++;
  }

  std::string type = tx_type_v[dis(mt) % tx_type_v.size()];
  neb::rt::dip::dip_reward::ignore_account_transfer_contract(txs, type);
  EXPECT_EQ(txs.size(), txs_size - tx_type_cnt[type]);
}

TEST(test_runtime_dip_reward, account_call_contract_count) {
  size_t ch_size = 'z' - 'a';
  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(0, ch_size);
  size_t txs_size = dis(mt) * dis(mt);
  txs_size++;

  std::vector<std::vector<int32_t>> s2t = std::vector<std::vector<int32_t>>(
      ch_size + 1, std::vector<int32_t>(ch_size + 1, 0));

  std::vector<neb::fs::transaction_info_t> txs;
  for (size_t i = 0; i < txs_size; i++) {
    char ch_s = 'a' + dis(mt);
    char ch_t = 'A' + dis(mt);

    s2t[ch_s - 'a'][ch_t - 'A']++;

    neb::fs::transaction_info_t info;
    info.m_from = neb::to_address(std::string(1, ch_s));
    info.m_to = neb::to_address(std::string(1, ch_t));
    txs.push_back(info);
  }

  auto ret = neb::rt::dip::dip_reward::account_call_contract_count(txs);
  for (size_t i = 0; i < s2t.size(); i++) {
    int32_t row_not_empty = 0;
    for (size_t j = 0; j < s2t[0].size(); j++) {
      if (s2t[i][j]) {
        row_not_empty++;
      }
    }
    auto tmp = ret->find(neb::to_address(std::string(1, 'a' + i)));
    if (row_not_empty) {
      EXPECT_TRUE(tmp != ret->end());
      EXPECT_EQ(row_not_empty, tmp->second.size());
    } else {
      EXPECT_TRUE(tmp == ret->end());
    }
  }
  for (size_t i = 0; i < s2t.size(); i++) {
    for (size_t j = 0; j < s2t[0].size(); j++) {
      if (s2t[i][j]) {
        auto tmp = ret->find(neb::to_address(std::string(1, 'a' + i)));
        EXPECT_TRUE(tmp != ret->end());
        auto temp = tmp->second.find(neb::to_address(std::string(1, 'A' + j)));
        EXPECT_TRUE(temp != tmp->second.end());
        EXPECT_EQ(s2t[i][j], temp->second);
      }
    }
  }
}

TEST(test_runtime_dip_reward, account_to_contract_votes) {
  std::vector<neb::fs::transaction_info_t> txs;
  neb::fs::transaction_info_t tx;
  tx.m_from = neb::to_address("a");
  tx.m_to = neb::to_address("A");
  txs.push_back(tx);

  std::vector<std::shared_ptr<neb::rt::nr::nr_info_t>> nr_infos;
  auto info_ptr = std::make_shared<neb::rt::nr::nr_info_t>();
  info_ptr->m_address = neb::to_address("a");
  info_ptr->m_nr_score = 1;
  nr_infos.push_back(info_ptr);

  auto ret = neb::rt::dip::dip_reward::account_to_contract_votes(txs, nr_infos);
  EXPECT_EQ(ret->size(), 1);
  auto tmp = ret->find(neb::to_address("a"));
  EXPECT_TRUE(tmp != ret->end());
  auto temp = tmp->second.find(neb::to_address("A"));
  EXPECT_TRUE(temp != tmp->second.end());
  EXPECT_EQ(temp->second, 1);

  // add account
  tx.m_from = neb::to_address("b");
  tx.m_to = neb::to_address("A");
  txs.push_back(tx);
  info_ptr = std::make_shared<neb::rt::nr::nr_info_t>();
  info_ptr->m_address = neb::to_address("b");
  info_ptr->m_nr_score = 1;
  nr_infos.push_back(info_ptr);

  ret = neb::rt::dip::dip_reward::account_to_contract_votes(txs, nr_infos);
  EXPECT_EQ(ret->size(), 2);
  tmp = ret->find(neb::to_address("a"));
  EXPECT_TRUE(tmp != ret->end());
  temp = tmp->second.find(neb::to_address("A"));
  EXPECT_TRUE(temp != tmp->second.end());
  EXPECT_EQ(temp->second, 1);

  tmp = ret->find(neb::to_address("b"));
  EXPECT_TRUE(tmp != ret->end());
  temp = tmp->second.find(neb::to_address("A"));
  EXPECT_TRUE(temp != tmp->second.end());
  EXPECT_EQ(temp->second, 1);
}

TEST(test_runtime_dip_reward, dapp_votes) {
  std::vector<neb::fs::transaction_info_t> txs;
  neb::fs::transaction_info_t tx;
  tx.m_from = neb::to_address("a");
  tx.m_to = neb::to_address("A");
  txs.push_back(tx);

  std::vector<std::shared_ptr<neb::rt::nr::nr_info_t>> nr_infos;
  auto info_ptr = std::make_shared<neb::rt::nr::nr_info_t>();
  info_ptr->m_address = neb::to_address("a");
  info_ptr->m_nr_score = 1;
  nr_infos.push_back(info_ptr);

  auto tmp = neb::rt::dip::dip_reward::account_to_contract_votes(txs, nr_infos);
  auto ret = neb::rt::dip::dip_reward::dapp_votes(*tmp);
  EXPECT_EQ(ret->size(), 1);
  EXPECT_EQ(ret->begin()->first, neb::to_address("A"));
  neb::floatxx_t val = ret->begin()->second;
  EXPECT_TRUE(neb::math::abs(val, neb::floatxx_t(1)) < precesion(1e-1));

  tx.m_from = neb::to_address("b");
  tx.m_to = neb::to_address("A");
  txs.push_back(tx);
  info_ptr = std::make_shared<neb::rt::nr::nr_info_t>();
  info_ptr->m_address = neb::to_address("b");
  info_ptr->m_nr_score = 1;
  nr_infos.push_back(info_ptr);

  tmp = neb::rt::dip::dip_reward::account_to_contract_votes(txs, nr_infos);
  ret = neb::rt::dip::dip_reward::dapp_votes(*tmp);
  EXPECT_EQ(ret->size(), 1);
  EXPECT_EQ(ret->begin()->first, neb::to_address("A"));
  val = ret->begin()->second;
  EXPECT_TRUE(neb::math::abs(val, neb::floatxx_t(2)) < precesion(1e-1));
}

TEST(test_runtime_dip_reward, participate_lambda) {
  std::vector<neb::fs::transaction_info_t> txs;
  neb::fs::transaction_info_t tx;
  tx.m_from = neb::to_address("a");
  tx.m_to = neb::to_address("A");
  txs.push_back(tx);

  std::vector<std::shared_ptr<neb::rt::nr::nr_info_t>> nr_infos;
  auto info_ptr = std::make_shared<neb::rt::nr::nr_info_t>();
  neb::rt::nr::nr_info_t &info = *info_ptr;
  info.m_address = neb::to_address("a");
  info.m_nr_score = 1;
  nr_infos.push_back(info_ptr);

  neb::floatxx_t alpha = 1;
  neb::floatxx_t beta = 1;
  neb::floatxx_t lambda =
      neb::rt::dip::dip_reward::participate_lambda(alpha, beta, txs, nr_infos);
  EXPECT_TRUE(neb::math::abs(lambda, neb::floatxx_t(1)) < precesion(1e-1));
}
