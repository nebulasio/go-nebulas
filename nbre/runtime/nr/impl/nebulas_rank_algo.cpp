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
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the // GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//
#include "runtime/nr/impl/nebulas_rank_algo.h"
#include "common/int128_conversion.h"
#include "common/math.h"
#include "common/nebulas_currency.h"

namespace neb {
namespace rt {
namespace nr {

std::unique_ptr<std::vector<std::vector<neb::fs::transaction_info_t>>>
nebulas_rank_algo::split_transactions_by_block_interval(
    const std::vector<neb::fs::transaction_info_t> &txs,
    int32_t block_interval) {

  auto ret =
      std::make_unique<std::vector<std::vector<neb::fs::transaction_info_t>>>();

  if (block_interval < 1 || txs.empty()) {
    return ret;
  }

  auto it = txs.begin();
  block_height_t block_first = it->m_height;
  it = txs.end();
  it--;
  block_height_t block_last = it->m_height;

  std::vector<neb::fs::transaction_info_t> v;
  it = txs.begin();
  block_height_t b = block_first;
  while (b <= block_last) {
    block_height_t h = it->m_height;
    if (h < b + block_interval) {
      v.push_back(*it++);
    } else {
      ret->push_back(v);
      v.clear();
      b += block_interval;
    }
    if (it == txs.end()) {
      ret->push_back(v);
      break;
    }
  }
  return ret;
}

void nebulas_rank_algo::filter_empty_transactions_this_interval(
    std::vector<std::vector<neb::fs::transaction_info_t>> &txs) {
  for (auto it = txs.begin(); it != txs.end();) {
    if (it->empty()) {
      it = txs.erase(it);
    } else {
      it++;
    }
  }
}

transaction_graph_ptr_t nebulas_rank_algo::build_graph_from_transactions(
    const std::vector<neb::fs::transaction_info_t> &trans) {
  auto ret = std::make_unique<neb::rt::transaction_graph>();

  for (auto ite = trans.begin(); ite != trans.end(); ite++) {
    address_t from = ite->m_from;
    address_t to = ite->m_to;
    wei_t value = ite->m_tx_value;
    int64_t timestamp = ite->m_timestamp;
    ret->add_edge(from, to, value, timestamp);
  }
  return ret;
}

std::unique_ptr<std::vector<transaction_graph_ptr_t>>
nebulas_rank_algo::build_transaction_graphs(
    const std::vector<std::vector<neb::fs::transaction_info_t>> &txs) {

  std::unique_ptr<std::vector<transaction_graph_ptr_t>> tgs =
      std::make_unique<std::vector<transaction_graph_ptr_t>>();

  for (auto it = txs.begin(); it != txs.end(); it++) {
    auto p = build_graph_from_transactions(*it);
    tgs->push_back(std::move(p));
  }
  return tgs;
}

block_height_t nebulas_rank_algo::get_max_height_this_block_interval(
    const std::vector<neb::fs::transaction_info_t> &txs) {
  if (txs.empty()) {
    return 0;
  }
  // suppose transactions in height increasing order
  return txs.back().m_height;
}

std::unique_ptr<std::unordered_set<address_t>>
nebulas_rank_algo::get_normal_accounts(
    const std::vector<neb::fs::transaction_info_t> &txs) {

  auto ret = std::make_unique<std::unordered_set<address_t>>();

  for (auto it = txs.begin(); it != txs.end(); it++) {
    auto from = it->m_from;
    ret->insert(from);

    auto to = it->m_to;
    ret->insert(to);
  }
  return ret;
}

std::unique_ptr<std::unordered_map<address_t, floatxx_t>>
nebulas_rank_algo::get_account_balance_median(
    neb::block_height_t start_block,
    const std::unordered_set<address_t> &accounts,
    const std::vector<std::vector<neb::fs::transaction_info_t>> &txs,
    fs::account_db_interface *db_ptr) {

  auto ret = std::make_unique<std::unordered_map<address_t, floatxx_t>>();
  std::unordered_map<address_t, std::vector<wei_t>> addr_balance_v;

  block_height_t max_height = start_block;
  for (auto it = txs.begin(); it != txs.end(); it++) {
    block_height_t height = get_max_height_this_block_interval(*it);
    height = std::max(height, max_height);
    for (auto ite = accounts.begin(); ite != accounts.end(); ite++) {
      address_t addr = *ite;
      wei_t balance = db_ptr->get_account_balance_internal(addr, height);
      addr_balance_v[addr].push_back(balance);
    }
    max_height = std::max(max_height, height);
  }

  floatxx_t zero = softfloat_cast<uint32_t, typename floatxx_t::value_type>(0);
  for (auto it = addr_balance_v.begin(); it != addr_balance_v.end(); it++) {
    std::vector<wei_t> &v = it->second;
    sort(v.begin(), v.end(),
         [](const wei_t &w1, const wei_t &w2) { return w1 < w2; });
    size_t v_len = v.size();
    floatxx_t median = to_float<floatxx_t>(v[v_len >> 1]);
    if ((v_len & 0x1) == 0) {
      auto tmp = to_float<floatxx_t>(v[(v_len >> 1) - 1]);
      median = (median + tmp) / 2;
    }

    floatxx_t normalized_median = wei_to_nas(median);
    ret->insert(std::make_pair(it->first, math::max(zero, normalized_median)));
  }

  return ret;
}

floatxx_t nebulas_rank_algo::f_account_weight(floatxx_t in_val,
                                              floatxx_t out_val) {
  floatxx_t pi = math::constants<floatxx_t>::pi();
  floatxx_t zero = softfloat_cast<uint32_t, typename floatxx_t::value_type>(0);
  floatxx_t two = softfloat_cast<uint32_t, typename floatxx_t::value_type>(2);
  floatxx_t four = softfloat_cast<uint32_t, typename floatxx_t::value_type>(4);

  floatxx_t atan_val = pi / two;
  if (in_val > zero) {
    atan_val = math::arctan(out_val / in_val);
  }
  auto tmp = math::sin(pi / four - atan_val);
  return (in_val + out_val) * math::exp((zero - two) * tmp * tmp);
}

std::unique_ptr<std::unordered_map<address_t, floatxx_t>>
nebulas_rank_algo::get_account_weight(
    const std::unordered_map<address_t, neb::rt::in_out_val_t> &in_out_vals,
    fs::account_db_interface *db_ptr) {

  auto ret = std::make_unique<std::unordered_map<address_t, floatxx_t>>();

  for (auto it = in_out_vals.begin(); it != in_out_vals.end(); it++) {
    wei_t in_val = it->second.m_in_val;
    wei_t out_val = it->second.m_out_val;

    floatxx_t f_in_val = to_float<floatxx_t>(in_val);
    floatxx_t f_out_val = to_float<floatxx_t>(out_val);

    floatxx_t normalized_in_val = wei_to_nas(f_in_val);
    floatxx_t normalized_out_val = wei_to_nas(f_out_val);

    auto tmp = f_account_weight(normalized_in_val, normalized_out_val);
    ret->insert(std::make_pair(it->first, tmp));
  }
  return ret;
}

floatxx_t nebulas_rank_algo::f_account_rank(int64_t a, int64_t b, int64_t c,
                                            int64_t d, floatxx_t theta,
                                            floatxx_t mu, floatxx_t lambda,
                                            floatxx_t S, floatxx_t R) {
  floatxx_t zero = softfloat_cast<uint32_t, typename floatxx_t::value_type>(0);
  floatxx_t one = softfloat_cast<uint32_t, typename floatxx_t::value_type>(1);

  auto gamma = math::pow(theta * R / (R + mu), lambda);
  auto ret = zero;
  if (S > zero) {
    ret = (S / (one + math::pow(a / S, one / b))) * gamma;
  }
  return ret;
}

std::unique_ptr<std::unordered_map<address_t, floatxx_t>>
nebulas_rank_algo::get_account_rank(
    const std::unordered_map<address_t, floatxx_t> &account_median,
    const std::unordered_map<address_t, floatxx_t> &account_weight,
    const rank_params_t &rp) {

  auto ret = std::make_unique<std::unordered_map<address_t, floatxx_t>>();

  for (auto it_m = account_median.begin(); it_m != account_median.end();
       it_m++) {
    auto it_w = account_weight.find(it_m->first);
    if (it_w != account_weight.end()) {
      floatxx_t rank_val =
          f_account_rank(rp.m_a, rp.m_b, rp.m_c, rp.m_d, rp.m_theta, rp.m_mu,
                         rp.m_lambda, it_m->second, it_w->second);
      ret->insert(std::make_pair(it_m->first, rank_val));
    }
  }

  return ret;
}

std::unique_ptr<std::vector<address_t>> nebulas_rank_algo::sort_accounts(
    const std::unordered_set<address_t> &accounts) {
  auto ret = std::make_unique<std::vector<address_t>>();
  for (auto &acc : accounts) {
    ret->push_back(acc);
  }
  sort(ret->begin(), ret->end());
  return ret;
}

} // namespace nr
} // namespace rt
} // namespace neb
