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

#include "runtime/dip/dip_reward.h"
#include "common/configuration.h"
#include "common/util/conversion.h"
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>

namespace neb {
namespace rt {
namespace dip {

std::unique_ptr<std::vector<dip_info_t>> dip_reward::get_dip_reward(
    neb::block_height_t start_block, neb::block_height_t end_block,
    neb::block_height_t height, const std::string &nr_result,
    const neb::rt::nr::transaction_db_ptr_t &tdb_ptr,
    const neb::rt::nr::account_db_ptr_t &adb_ptr, floatxx_t alpha,
    floatxx_t beta) {

  auto it_txs =
      tdb_ptr->read_transactions_from_db_with_duration(start_block, end_block);
  auto txs = *it_txs;

  auto it_nr_infos = neb::rt::nr::nebulas_rank::json_to_nr_info(nr_result);
  auto it_acc_to_contract_txs =
      neb::fs::transaction_db::read_transactions_with_address_type(txs, 0x57,
                                                                   0x58);
  // dapp total votes
  auto it_acc_to_contract_votes =
      account_to_contract_votes(*it_acc_to_contract_txs, *it_nr_infos);
  auto it_dapp_votes = dapp_votes(*it_acc_to_contract_votes);

  // bonus pool in total
  std::string admin_addr = neb::configuration::instance().admin_pub_addr();
  neb::util::bytes addr_bytes = neb::util::bytes::from_base58(admin_addr);
  wei_t balance =
      adb_ptr->get_balance(neb::util::byte_to_string(addr_bytes), height);
  floatxx_t bonus_total = int128_conversion(balance).to_float<floatxx_t>();
  bonus_total = adb_ptr->get_normalized_value(bonus_total);

  floatxx_t sum_votes(0);
  for (auto &v : *it_dapp_votes) {
    sum_votes += v.second * v.second;
  }

  std::vector<dip_info_t> dip_infos;
  for (auto &v : *it_dapp_votes) {
    dip_info_t info;
    info.m_address = v.first;

    info.m_reward =
        v.second * v.second *
        participate_lambda(alpha, beta, *it_acc_to_contract_txs, *it_nr_infos) *
        bonus_total / sum_votes;
    dip_infos.push_back(info);
  }
  return std::make_unique<std::vector<dip_info_t>>(dip_infos);
}

std::string
dip_reward::dip_info_to_json(const std::vector<dip_info_t> &dip_infos) {
  boost::property_tree::ptree root;
  boost::property_tree::ptree arr;

  for (auto &info : dip_infos) {
    boost::property_tree::ptree p;
    auto dip_info_to_ptree = [&info, &p]() -> void {
      neb::util::bytes addr_bytes = neb::util::string_to_byte(info.m_address);
      floatxx_t f_reward = info.m_reward;
      nr::uintxx_t int_reward =
          softfloat_cast<floatxx_t::value_type, nr::uintxx_t>(
              f_reward.round_to_int());

      std::vector<std::pair<std::string, std::string>> kv_pair(
          {{"address", addr_bytes.to_base58()},
           {"reward", nr::internal::to_string(int_reward)}});
      for (auto &ele : kv_pair) {
        p.put(ele.first, ele.second);
      }
      return;
    };
    dip_info_to_ptree();
    arr.push_back(std::make_pair(std::string(), p));
  }
  root.add_child("dips", arr);

  std::stringstream ss;
  boost::property_tree::json_parser::write_json(ss, root, false);
  return ss.str();
}

std::unique_ptr<
    std::unordered_map<address_t, std::unordered_map<address_t, uint32_t>>>
dip_reward::account_call_contract_count(
    const std::vector<neb::fs::transaction_info_t> &txs) {

  std::unordered_map<address_t, std::unordered_map<address_t, uint32_t>> cnt;

  for (auto &tx : txs) {
    std::string acc_addr = tx.m_from;
    std::string contract_addr = tx.m_to;
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
  return std::make_unique<
      std::unordered_map<address_t, std::unordered_map<address_t, uint32_t>>>(
      cnt);
}

std::unique_ptr<
    std::unordered_map<address_t, std::unordered_map<address_t, floatxx_t>>>
dip_reward::account_to_contract_votes(
    const std::vector<neb::fs::transaction_info_t> &txs,
    const std::vector<neb::rt::nr::nr_info_t> &nr_infos) {

  std::unordered_map<address_t, std::unordered_map<address_t, floatxx_t>> ret;

  auto it_cnt = account_call_contract_count(txs);
  auto cnt = *it_cnt;

  for (auto &info : nr_infos) {
    std::string addr = info.m_address;
    floatxx_t score = info.m_nr_score * info.m_nr_score;

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
  return std::make_unique<
      std::unordered_map<address_t, std::unordered_map<address_t, floatxx_t>>>(
      ret);
}

std::unique_ptr<std::unordered_map<address_t, floatxx_t>>
dip_reward::dapp_votes(const std::unordered_map<
                       address_t, std::unordered_map<address_t, floatxx_t>>
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

  return std::make_unique<std::unordered_map<address_t, floatxx_t>>(ret);
}

floatxx_t dip_reward::participate_lambda(
    floatxx_t alpha, floatxx_t beta,
    const std::vector<neb::fs::transaction_info_t> &txs,
    const std::vector<neb::rt::nr::nr_info_t> &nr_infos) {

  std::unordered_set<std::string> addr_set;
  for (auto &tx : txs) {
    addr_set.insert(tx.m_from);
  }

  std::vector<floatxx_t> participate_nr;
  for (auto &info : nr_infos) {
    std::string addr = info.m_address;
    if (addr_set.find(addr) != addr_set.end()) {
      participate_nr.push_back(info.m_nr_score);
    }
  }

  floatxx_t gamma_p(0);
  for (auto &nr : participate_nr) {
    gamma_p += nr * nr;
  }
  floatxx_t variance(0);
  size_t participate_size = participate_nr.size();
  for (auto &nr : participate_nr) {
    floatxx_t tmp = nr * nr - gamma_p / participate_size;
    variance += tmp * tmp;
  }

  floatxx_t gamma_s(0);
  for (auto &info : nr_infos) {
    gamma_s += info.m_nr_score * info.m_nr_score;
  }

  return neb::math::min(
      gamma_p *
          neb::math::min(beta * gamma_p * gamma_p / variance, floatxx_t(1)) /
          (alpha * gamma_s),
      floatxx_t(1));
}
} // namespace dip
} // namespace rt
} // namespace neb
