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
#include "runtime/dip/dip_algo.h"
#include "common/math.h"

namespace neb {
namespace rt {
namespace dip {

void dip_algo::back_to_coinbase(std::vector<dip_item> &dip_infos,
                                floatxx_t reward_left,
                                const address_t &dip_coinbase_addr) {

  if (!dip_coinbase_addr.empty() && reward_left > 0) {
    dip_item di;
    di.set<p_dip_deployer>(address_to_string(dip_coinbase_addr));
    di.set<p_dip_reward_value>(reward_left);
    dip_infos.push_back(di);
  }
}

std::unordered_map<address_t, std::unordered_map<address_t, uint32_t>>
dip_algo::account_call_contract_count(
    const std::vector<neb::fs::transaction_info_t> &txs) {

  std::unordered_map<address_t, std::unordered_map<address_t, uint32_t>> cnt;

  for (auto &tx : txs) {
    address_t acc_addr = tx.m_from;
    address_t contract_addr = tx.m_to;
    auto it = cnt.find(acc_addr);

    if (it != cnt.end()) {
      std::unordered_map<address_t, uint32_t> &tmp = it->second;
      if (tmp.find(contract_addr) != tmp.end()) {
        tmp[contract_addr]++;
      } else {
        tmp.insert(std::make_pair(contract_addr, 1));
      }
    } else {
      std::unordered_map<address_t, uint32_t> tmp;
      tmp.insert(std::make_pair(contract_addr, 1));
      cnt.insert(std::make_pair(acc_addr, tmp));
    }
  }
  return cnt;
}

std::unordered_map<address_t, std::unordered_map<address_t, floatxx_t>>
dip_algo::account_to_contract_votes(
    const std::vector<neb::fs::transaction_info_t> &txs,
    const std::vector<nr_item> &nr_infos) {

  std::unordered_map<address_t, std::unordered_map<address_t, floatxx_t>> ret;

  auto cnt = account_call_contract_count(txs);

  for (auto &info : nr_infos) {
    address_t addr = to_address(info.get<p_nr_item_addr>());
    floatxx_t score = info.get<p_nr_item_score>();

    auto it_acc = cnt.find(addr);
    if (it_acc == cnt.end()) {
      continue;
    }

    floatxx_t sum_votes(0);
    for (auto &e : it_acc->second) {
      sum_votes += e.second;
    }

    for (auto &e : it_acc->second) {
      std::unordered_map<address_t, floatxx_t> tmp;
      tmp.insert(std::make_pair(e.first, e.second * score / sum_votes));
      ret.insert(std::make_pair(addr, tmp));
    }
  }
  return ret;
}

std::unordered_map<address_t, floatxx_t> dip_algo::dapp_votes(
    const std::unordered_map<address_t,
                             std::unordered_map<address_t, floatxx_t>>
        &acc_contract_votes) {
  std::unordered_map<address_t, floatxx_t> ret;

  for (auto &it : acc_contract_votes) {
    for (auto &ite : it.second) {
      auto iter = ret.find(ite.first);
      if (iter != ret.end()) {
        floatxx_t &tmp = iter->second;
        tmp += neb::math::sqrt(ite.second);
      } else {
        ret.insert(std::make_pair(ite.first, neb::math::sqrt(ite.second)));
      }
    }
  }
  return ret;
}

floatxx_t dip_algo::participate_lambda(
    floatxx_t alpha, floatxx_t beta,
    const std::vector<neb::fs::transaction_info_t> &txs,
    const std::vector<nr_item> &nr_infos) {

  std::unordered_set<address_t> addr_set;
  for (auto &tx : txs) {
    addr_set.insert(tx.m_from);
  }

  std::vector<floatxx_t> participate_nr;
  for (auto &info : nr_infos) {
    address_t addr = to_address(info.get<p_nr_item_addr>());
    if (addr_set.find(addr) != addr_set.end()) {
      participate_nr.push_back(info.get<p_nr_item_score>());
    }
  }

  floatxx_t gamma_p(0);
  for (auto &nr : participate_nr) {
    gamma_p += nr;
  }

  floatxx_t gamma_s(0);
  for (auto &info : nr_infos) {
    gamma_s += info.get<p_nr_item_score>();
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

void dip_algo::ignore_account_transfer_contract(
    std::vector<neb::fs::transaction_info_t> &txs, const std::string &tx_type) {
  for (auto it = txs.begin(); it != txs.end();) {
    it = (it->m_tx_type == tx_type ? txs.erase(it) : it + 1);
  }
}
} // namespace dip
} // namespace rt
} // namespace neb
