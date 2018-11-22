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

#include "ir/nr/algo/nebulas_rank.h"

namespace neb {
namespace nr {

nebulas_rank::nebulas_rank() {}
nebulas_rank::nebulas_rank(const transaction_db_ptr_t tdb_ptr,
                           const account_db_ptr_t adb_ptr,
                           const rank_params_t &rp, block_height_t start_block,
                           block_height_t end_block)
    : m_tdb_ptr(tdb_ptr), m_adb_ptr(adb_ptr), m_rp(rp),
      m_start_block(start_block), m_end_block(end_block) {}

std::shared_ptr<std::vector<std::vector<transaction_info_t>>>
nebulas_rank::split_transactions_by_x_block_interval(
    const std::vector<transaction_info_t> &txs, int32_t block_interval) {

  std::vector<std::vector<transaction_info_t>> ret;

  if (block_interval < 1 || txs.empty()) {
    return std::make_shared<std::vector<std::vector<transaction_info_t>>>(ret);
  }

  auto it = txs.begin();
  block_height_t block_first = it->m_height;
  it = txs.end();
  it--;
  block_height_t block_last = it->m_height;

  std::vector<transaction_info_t> v;
  it = txs.begin();
  block_height_t b = block_first;
  while (b <= block_last) {
    block_height_t h = it->m_height;
    if (h < b + block_interval) {
      v.push_back(*it++);
    } else {
      ret.push_back(v);
      v.clear();
      b += block_interval;
    }
    if (it == txs.end()) {
      ret.push_back(v);
      break;
    }
  }
  return std::make_shared<std::vector<std::vector<transaction_info_t>>>(ret);
}

void nebulas_rank::filter_empty_transactions_this_interval(
    std::vector<std::vector<transaction_info_t>> &txs) {
  for (auto it = txs.begin(); it != txs.end();) {
    if (it->empty()) {
      it = txs.erase(it);
    } else {
      it++;
    }
  }
}

template <class TransInfo>
transaction_graph_ptr_t nebulas_rank::build_graph_from_transactions(
    const std::vector<TransInfo> &trans) {
  transaction_graph_ptr_t ret = std::make_shared<neb::rt::transaction_graph>();

  for (auto ite = trans.begin(); ite != trans.end(); ite++) {
    std::string from = ite->m_from;
    std::string to = ite->m_to;
    std::string tx_value = ite->m_tx_value;
    double value = std::stod(tx_value);
    double timestamp = std::stod(ite->m_timestamp);
    ret->add_edge(from, to, value, timestamp);
  }
  return ret;
}

std::vector<transaction_graph_ptr_t> nebulas_rank::build_transaction_graphs(
    const std::vector<std::vector<transaction_info_t>> &txs) {
  std::vector<transaction_graph_ptr_t> tgs;

  for (auto it = txs.begin(); it != txs.end(); it++) {
    auto p = build_graph_from_transactions(*it);
    tgs.push_back(p);
  }
  return tgs;
}

block_height_t nebulas_rank::get_max_height_this_block_interval(
    const std::vector<transaction_info_t> &txs) {
  if (txs.size() > 0) {
    return txs[txs.size() - 1].m_height;
  }
  return 0;
}

std::shared_ptr<std::unordered_set<std::string>>
nebulas_rank::get_normal_accounts(const std::vector<transaction_info_t> &txs) {

  std::unordered_set<std::string> ret;

  for (auto it = txs.begin(); it != txs.end(); it++) {
    std::string from = it->m_from;
    ret.insert(from);

    std::string to = it->m_to;
    ret.insert(to);
  }
  return std::make_shared<std::unordered_set<std::string>>(ret);
}

std::shared_ptr<std::unordered_map<std::string, double>>
nebulas_rank::get_account_balance_median(
    const std::unordered_set<std::string> &accounts,
    const std::vector<std::vector<transaction_info_t>> &txs,
    const account_db_ptr_t db_ptr,
    std::unordered_map<account_address_t, account_balance_t> &addr_balance) {

  std::unordered_map<std::string, double> ret;
  std::unordered_map<std::string, std::vector<double>> addr_balance_v;

  db_ptr->set_height_address_val(m_start_block, m_end_block, addr_balance);

  for (auto it = txs.begin(); it != txs.end(); it++) {
    block_height_t max_height_this_interval =
        get_max_height_this_block_interval(*it);
    for (auto ite = accounts.begin(); ite != accounts.end(); ite++) {
      std::string addr = *ite;
      double balance = boost::lexical_cast<double>(
          db_ptr->get_account_balance(max_height_this_interval, addr));
      addr_balance_v[addr].push_back(balance);
    }
  }

  for (auto it = addr_balance_v.begin(); it != addr_balance_v.end(); it++) {
    std::vector<double> v = it->second;
    sort(v.begin(), v.end());
    int v_len = v.size();
    double median = v[v_len / 2];
    if ((v_len & 1) == 0) {
      median = (median + v[v_len / 2 - 1]) / 2;
    }

    double normalized_median = db_ptr->get_normalized_value(median);
    ret.insert(std::make_pair(it->first, fmax(0.0, normalized_median)));
  }

  return std::make_shared<std::unordered_map<std::string, double>>(ret);
}

double nebulas_rank::f_account_weight(double in_val, double out_val) {
  double pi = acos(-1.0);
  double atan_val = (in_val == 0 ? pi / 2 : atan(out_val / in_val));
  return (in_val + out_val) * exp((-2) * pow(sin(pi / 4.0 - atan_val), 2.0));
}

std::shared_ptr<std::unordered_map<std::string, double>>
nebulas_rank::get_account_weight(
    const std::unordered_map<std::string, neb::rt::in_out_val_t> &in_out_vals,
    const account_db_ptr_t db_ptr) {

  std::unordered_map<std::string, double> ret;

  for (auto it = in_out_vals.begin(); it != in_out_vals.end(); it++) {
    double in_val = it->second.m_in_val;
    double out_val = it->second.m_out_val;

    double normalized_in_val = db_ptr->get_normalized_value(in_val);
    double normalized_out_val = db_ptr->get_normalized_value(out_val);
    ret.insert(std::make_pair(
        it->first, f_account_weight(normalized_in_val, normalized_out_val)));
  }
  return std::make_shared<std::unordered_map<std::string, double>>(ret);
}

double nebulas_rank::f_account_rank(double a, double b, double c, double d,
                                    double mu, double lambda, double S,
                                    double R) {
  return pow(S * a / (S + b), mu) * pow(R * c / (R + d), lambda);
}

std::shared_ptr<std::unordered_map<std::string, double>>
nebulas_rank::get_account_rank(
    const std::unordered_map<std::string, double> &account_median,
    const std::unordered_map<std::string, double> &account_weight,
    const rank_params_t &rp) {

  std::unordered_map<std::string, double> ret;

  for (auto it_m = account_median.begin(); it_m != account_median.end();
       it_m++) {
    auto it_w = account_weight.find(it_m->first);
    if (it_w != account_weight.end()) {
      double rank_val = f_account_rank(rp.m_a, rp.m_b, rp.m_c, rp.m_d, rp.m_mu,
                                       rp.m_lambda, it_m->second, it_w->second);
      ret.insert(std::make_pair(it_m->first, rank_val));
    }
  }
  return std::make_shared<std::unordered_map<std::string, double>>(ret);
}

std::shared_ptr<std::unordered_map<std::string, double>>
nebulas_rank::get_account_score_service() {

  auto it_ret = m_tdb_ptr->read_inter_transaction_from_db_with_duration(
      m_start_block, m_end_block);
  auto ret = *it_ret;

  auto it_txs = split_transactions_by_x_block_interval(ret);
  auto txs = *it_txs;
  filter_empty_transactions_this_interval(txs);

  std::vector<transaction_graph_ptr_t> tgs = build_transaction_graphs(txs);
  if (tgs.empty()) {
    return std::make_shared<std::unordered_map<std::string, double>>();
  }

  for (auto it = tgs.begin(); it != tgs.end(); it++) {
    transaction_graph_ptr_t ptr = *it;
    neb::rt::graph_algo::remove_cycles_based_on_time_sequence(
        ptr->internal_graph());
    neb::rt::graph_algo::merge_edges_with_same_from_and_same_to(
        ptr->internal_graph());
  }

  transaction_graph_ptr_t tg = neb::rt::graph_algo::merge_graphs(tgs);
  neb::rt::graph_algo::merge_topk_edges_with_same_from_and_same_to(
      tg->internal_graph());

  auto it_accounts = get_normal_accounts(ret);
  auto accounts = *it_accounts;
  std::unordered_map<account_address_t, account_balance_t> addr_balance;
  auto it_median =
      get_account_balance_median(accounts, txs, m_adb_ptr, addr_balance);
  auto median = *it_median;

  auto it_in_out_vals =
      neb::rt::graph_algo::get_in_out_vals(tg->internal_graph());
  auto in_out_vals = *it_in_out_vals;
  auto it_account_weight = get_account_weight(in_out_vals, m_adb_ptr);
  auto account_weight = *it_account_weight;

  return get_account_rank(median, account_weight, m_rp);
}
} // namespace nr
} // namespace neb
