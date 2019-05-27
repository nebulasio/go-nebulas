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

#include "runtime/dip/dip_reward_v2.h"
#include "common/configuration.h"
#include "common/int128_conversion.h"
#include "runtime/dip/dip_handler.h"

namespace neb {
namespace rt {
namespace dip {

std::vector<std::shared_ptr<dip_info_t>> dip_reward_v2::get_dip_reward(
    neb::block_height_t start_block, neb::block_height_t end_block,
    neb::block_height_t height,
    const std::vector<std::shared_ptr<nr_info_t>> &nr_result,
    const neb::rt::nr::transaction_db_v2_ptr_t &tdb_ptr,
    const neb::rt::nr::account_db_v2_ptr_t &adb_ptr, floatxx_t alpha,
    floatxx_t beta) {

  auto it_txs =
      tdb_ptr->read_transactions_from_db_with_duration(start_block, end_block);
  auto txs = *it_txs;
  LOG(INFO) << "transaction size " << txs.size();

  auto it_nr_infos = nr_result;
  auto it_acc_to_contract_txs =
      fs::transaction_db::read_transactions_with_address_type(
          txs, NAS_ADDRESS_ACCOUNT_MAGIC_NUM, NAS_ADDRESS_CONTRACT_MAGIC_NUM);
  ignore_account_transfer_contract(*it_acc_to_contract_txs, "binary");
  LOG(INFO) << "account to contract size " << it_acc_to_contract_txs->size();
  // dapp total votes
  auto it_acc_to_contract_votes =
      account_to_contract_votes(*it_acc_to_contract_txs, it_nr_infos);
  LOG(INFO) << "account to contract votes " << it_acc_to_contract_votes->size();
  auto it_dapp_votes = dapp_votes(*it_acc_to_contract_votes);
  LOG(INFO) << "dapp votes size " << it_dapp_votes->size();

  // bonus pool in total
  auto dip_params_ptr = dip_handler::instance().get_dip_params(end_block + 1);
  auto &dip_params = *dip_params_ptr;
  address_t dip_reward_addr = to_address(dip_params.get<reward_addr>());
  address_t dip_coinbase_addr = to_address(dip_params.get<coinbase_addr>());

  wei_t balance = adb_ptr->get_balance(dip_reward_addr, end_block);
  floatxx_t bonus_total = to_float<floatxx_t>(balance);
  LOG(INFO) << "bonus total " << bonus_total;
  // bonus_total = adb_ptr->get_normalized_value(bonus_total);

  floatxx_t sum_votes(0);
  for (auto &v : *it_dapp_votes) {
    sum_votes += v.second * v.second;
  }
  LOG(INFO) << "sum votes " << sum_votes;

  floatxx_t reward_sum(0);
  std::vector<std::shared_ptr<dip_info_t>> dip_infos;
  for (auto &v : *it_dapp_votes) {
    std::shared_ptr<dip_info_t> info_ptr = std::make_shared<dip_info_t>();
    dip_info_t &info = *info_ptr;
    info.m_contract = v.first;
    info.m_deployer = adb_ptr->get_contract_deployer(v.first, end_block);

    floatxx_t reward_in_wei =
        v.second * v.second *
        participate_lambda(alpha, beta, *it_acc_to_contract_txs, it_nr_infos) *
        bonus_total / sum_votes;
    reward_sum += reward_in_wei;

    info.m_reward = neb::math::to_string(from_float(reward_in_wei));
    dip_infos.push_back(info_ptr);
  }
  LOG(INFO) << "reward sum " << reward_sum << ", bonus total " << bonus_total;
  // assert(reward_sum <= bonus_total);
  back_to_coinbase(dip_infos, bonus_total - reward_sum, dip_coinbase_addr);
  LOG(INFO) << "back to coinbase";
  return dip_infos;
}

void dip_reward_v2::back_to_coinbase(
    std::vector<std::shared_ptr<dip_info_t>> &dip_infos, floatxx_t reward_left,
    const address_t &dip_coinbase_addr) {

  if (!dip_coinbase_addr.empty() && reward_left > 0) {
    std::shared_ptr<dip_info_t> info = std::make_shared<dip_info_t>();
    info->m_deployer = dip_coinbase_addr;
    info->m_reward = neb::math::to_string(from_float(reward_left));
    dip_infos.push_back(info);
  }
}

void dip_reward_v2::full_fill_meta_info(
    const std::vector<std::pair<std::string, std::string>> &meta,
    boost::property_tree::ptree &root) {

  assert(meta.size() == 3);

  for (auto &ele : meta) {
    root.put(ele.first, ele.second);
  }
}

std::unique_ptr<
    std::unordered_map<address_t, std::unordered_map<address_t, uint32_t>>>
dip_reward_v2::account_call_contract_count(
    const std::vector<neb::fs::transaction_info_t> &txs) {

  auto cnt = std::make_unique<
      std::unordered_map<address_t, std::unordered_map<address_t, uint32_t>>>();

  for (auto &tx : txs) {
    address_t acc_addr = tx.m_from;
    address_t contract_addr = tx.m_to;
    auto it = cnt->find(acc_addr);

    if (it != cnt->end()) {
      std::unordered_map<address_t, uint32_t> &tmp = it->second;
      if (tmp.find(contract_addr) != tmp.end()) {
        tmp[contract_addr]++;
      } else {
        tmp.insert(std::make_pair(contract_addr, 1));
      }
    } else {
      std::unordered_map<address_t, uint32_t> tmp;
      tmp.insert(std::make_pair(contract_addr, 1));
      cnt->insert(std::make_pair(acc_addr, tmp));
    }
  }
  return cnt;
}

std::unique_ptr<std::map<address_t, std::map<address_t, floatxx_t>>>
dip_reward_v2::account_to_contract_votes(
    const std::vector<neb::fs::transaction_info_t> &txs,
    const std::vector<std::shared_ptr<neb::rt::nr::nr_info_t>> &nr_infos) {

  auto ret =
      std::make_unique<std::map<address_t, std::map<address_t, floatxx_t>>>();

  auto it_cnt = account_call_contract_count(txs);
  auto cnt = *it_cnt;

  for (auto &info : nr_infos) {
    address_t addr = info->m_address;
    floatxx_t score = info->m_nr_score;

    auto it_acc = cnt.find(addr);
    if (it_acc == cnt.end()) {
      continue;
    }

    floatxx_t sum_votes(0);
    for (auto &e : it_acc->second) {
      sum_votes += e.second;
    }

    for (auto &e : it_acc->second) {
      std::map<address_t, floatxx_t> tmp;
      tmp.insert(std::make_pair(e.first, e.second * score / sum_votes));
      ret->insert(std::make_pair(addr, tmp));
    }
  }
  return ret;
}

std::unique_ptr<std::map<address_t, floatxx_t>> dip_reward_v2::dapp_votes(
    const std::map<address_t, std::map<address_t, floatxx_t>>
        &acc_contract_votes) {
  auto ret = std::make_unique<std::map<address_t, floatxx_t>>();

  for (auto &it : acc_contract_votes) {
    for (auto &ite : it.second) {
      auto iter = ret->find(ite.first);
      if (iter != ret->end()) {
        floatxx_t &tmp = iter->second;
        tmp += neb::math::sqrt(ite.second);
      } else {
        ret->insert(std::make_pair(ite.first, neb::math::sqrt(ite.second)));
      }
    }
  }
  return ret;
}

floatxx_t dip_reward_v2::participate_lambda(
    floatxx_t alpha, floatxx_t beta,
    const std::vector<neb::fs::transaction_info_t> &txs,
    const std::vector<std::shared_ptr<neb::rt::nr::nr_info_t>> &nr_infos) {

  std::unordered_set<address_t> addr_set;
  for (auto &tx : txs) {
    addr_set.insert(tx.m_from);
  }

  std::vector<floatxx_t> participate_nr;
  for (auto &info : nr_infos) {
    address_t addr = info->m_address;
    if (addr_set.find(addr) != addr_set.end()) {
      participate_nr.push_back(info->m_nr_score);
    }
  }

  floatxx_t gamma_p(0);
  for (auto &nr : participate_nr) {
    gamma_p += nr;
  }

  floatxx_t gamma_s(0);
  for (auto &info : nr_infos) {
    gamma_s += info->m_nr_score;
  }

  floatxx_t zero = softfloat_cast<uint32_t, typename floatxx_t::value_type>(0);
  floatxx_t one = softfloat_cast<uint32_t, typename floatxx_t::value_type>(1);

  if (gamma_s == zero) {
    return zero;
  }
  if (gamma_p == beta * gamma_s) {
    return one;
  }

  return math::min(one, alpha * gamma_s / (beta * gamma_s - gamma_p));
}

void dip_reward_v2::ignore_account_transfer_contract(
    std::vector<neb::fs::transaction_info_t> &txs, const std::string &tx_type) {
  for (auto it = txs.begin(); it != txs.end();) {
    it = (it->m_tx_type == tx_type ? txs.erase(it) : it + 1);
  }
}
} // namespace dip
} // namespace rt
} // namespace neb
