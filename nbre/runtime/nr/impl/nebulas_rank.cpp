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

#include "runtime/nr/impl/nebulas_rank.h"
#include "common/int128_conversion.h"
#include "common/log.h"
#include "common/version.h"
#include <boost/algorithm/string/replace.hpp>
#include <boost/foreach.hpp>
#include <boost/property_tree/json_parser.hpp>
#include <chrono>
#include <thread>

namespace neb {
namespace rt {

namespace nr {

std::unique_ptr<std::vector<std::vector<neb::fs::transaction_info_t>>>
nebulas_rank::split_transactions_by_block_interval(
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

void nebulas_rank::filter_empty_transactions_this_interval(
    std::vector<std::vector<neb::fs::transaction_info_t>> &txs) {
  for (auto it = txs.begin(); it != txs.end();) {
    if (it->empty()) {
      it = txs.erase(it);
    } else {
      it++;
    }
  }
}

transaction_graph_ptr_t nebulas_rank::build_graph_from_transactions(
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
nebulas_rank::build_transaction_graphs(
    const std::vector<std::vector<neb::fs::transaction_info_t>> &txs) {

  std::unique_ptr<std::vector<transaction_graph_ptr_t>> tgs =
      std::make_unique<std::vector<transaction_graph_ptr_t>>();

  for (auto it = txs.begin(); it != txs.end(); it++) {
    auto p = build_graph_from_transactions(*it);
    tgs->push_back(std::move(p));
  }
  return tgs;
}

block_height_t nebulas_rank::get_max_height_this_block_interval(
    const std::vector<neb::fs::transaction_info_t> &txs) {
  if (txs.empty()) {
    return 0;
  }
  // suppose transactions in height increasing order
  return txs.back().m_height;
}

std::unique_ptr<std::unordered_set<address_t>>
nebulas_rank::get_normal_accounts(
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
nebulas_rank::get_account_balance_median(
    const std::unordered_set<address_t> &accounts,
    const std::vector<std::vector<neb::fs::transaction_info_t>> &txs,
    const account_db_ptr_t &db_ptr) {

  auto ret = std::make_unique<std::unordered_map<address_t, floatxx_t>>();
  std::unordered_map<address_t, std::vector<wei_t>> addr_balance_v;

  for (auto it = txs.begin(); it != txs.end(); it++) {
    block_height_t max_height = get_max_height_this_block_interval(*it);
    for (auto ite = accounts.begin(); ite != accounts.end(); ite++) {
      address_t addr = *ite;
      wei_t balance = db_ptr->get_account_balance_internal(addr, max_height);
      addr_balance_v[addr].push_back(balance);
    }
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

    floatxx_t normalized_median = db_ptr->get_normalized_value(median);
    ret->insert(std::make_pair(it->first, math::max(zero, normalized_median)));
  }

  return ret;
}

floatxx_t nebulas_rank::f_account_weight(floatxx_t in_val, floatxx_t out_val) {
  floatxx_t pi = math::constants<floatxx_t>::pi();
  floatxx_t atan_val = pi / 2.0;
  if (in_val > 0) {
    atan_val = math::arctan(out_val / in_val);
  }
  auto tmp = math::sin(pi / 4.0 - atan_val);
  return (in_val + out_val) * math::exp((-2.0) * tmp * tmp);
}

std::unique_ptr<std::unordered_map<address_t, floatxx_t>>
nebulas_rank::get_account_weight(
    const std::unordered_map<address_t, neb::rt::in_out_val_t> &in_out_vals,
    const account_db_ptr_t &db_ptr) {

  auto ret = std::make_unique<std::unordered_map<address_t, floatxx_t>>();

  for (auto it = in_out_vals.begin(); it != in_out_vals.end(); it++) {
    wei_t in_val = it->second.m_in_val;
    wei_t out_val = it->second.m_out_val;

    floatxx_t f_in_val = to_float<floatxx_t>(in_val);
    floatxx_t f_out_val = to_float<floatxx_t>(out_val);

    floatxx_t normalized_in_val = db_ptr->get_normalized_value(f_in_val);
    floatxx_t normalized_out_val = db_ptr->get_normalized_value(f_out_val);

    auto tmp = f_account_weight(normalized_in_val, normalized_out_val);
    ret->insert(std::make_pair(it->first, tmp));
  }
  return ret;
}

floatxx_t nebulas_rank::f_account_rank(int64_t a, int64_t b, int64_t c,
                                       int64_t d, floatxx_t theta, floatxx_t mu,
                                       floatxx_t lambda, floatxx_t S,
                                       floatxx_t R) {
  floatxx_t zero = softfloat_cast<uint32_t, typename floatxx_t::value_type>(0);
  floatxx_t one = softfloat_cast<uint32_t, typename floatxx_t::value_type>(1);
  auto gamma = math::pow(theta * R / (R + mu), lambda);
  auto ret = zero;
  if (S > 0) {
    ret = (S / (one + math::pow(a / S, one / b))) * gamma;
  }
  return ret;
}

std::unique_ptr<std::unordered_map<address_t, floatxx_t>>
nebulas_rank::get_account_rank(
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

std::vector<std::shared_ptr<nr_info_t>> nebulas_rank::get_nr_score(
    const transaction_db_ptr_t &tdb_ptr, const account_db_ptr_t &adb_ptr,
    const rank_params_t &rp, neb::block_height_t start_block,
    neb::block_height_t end_block) {

  auto start_time = std::chrono::high_resolution_clock::now();
  // transactions in total and account inter transactions
  auto txs_ptr =
      tdb_ptr->read_transactions_from_db_with_duration(start_block, end_block);
  LOG(INFO) << "raw tx size: " << txs_ptr->size();
  auto inter_txs_ptr = fs::transaction_db::read_transactions_with_address_type(
      *txs_ptr, NAS_ADDRESS_ACCOUNT_MAGIC_NUM, NAS_ADDRESS_ACCOUNT_MAGIC_NUM);
  LOG(INFO) << "account to account: " << inter_txs_ptr->size();

  // graph operation
  auto txs_v_ptr = split_transactions_by_block_interval(*inter_txs_ptr);
  LOG(INFO) << "split by block interval: " << txs_v_ptr->size();

  filter_empty_transactions_this_interval(*txs_v_ptr);
  auto tgs_ptr = build_transaction_graphs(*txs_v_ptr);
  if (tgs_ptr->empty()) {
    return std::vector<std::shared_ptr<nr_info_t>>();
  }
  LOG(INFO) << "we have " << tgs_ptr->size() << " subgraphs.";
  for (auto it = tgs_ptr->begin(); it != tgs_ptr->end(); it++) {
    transaction_graph *ptr = it->get();
    graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
        ptr->internal_graph());
    graph_algo::merge_edges_with_same_from_and_same_to(ptr->internal_graph());
  }
  LOG(INFO) << "done with remove cycle.";

  transaction_graph *tg = neb::rt::graph_algo::merge_graphs(*tgs_ptr);
  graph_algo::merge_topk_edges_with_same_from_and_same_to(tg->internal_graph());
  LOG(INFO) << "done with merge graphs.";

  // in_out amount
  auto in_out_vals_p = graph_algo::get_in_out_vals(tg->internal_graph());
  auto in_out_vals = *in_out_vals_p;
  LOG(INFO) << "done with get in_out_vals";

  // median, weight, rank
  auto accounts_ptr = get_normal_accounts(*inter_txs_ptr);
  LOG(INFO) << "account size: " << accounts_ptr->size();

  std::unordered_map<neb::address_t, neb::wei_t> addr_balance;
  for (auto &acc : *accounts_ptr) {
    auto balance = adb_ptr->get_balance(acc, start_block);
    addr_balance.insert(std::make_pair(acc, balance));
  }
  LOG(INFO) << "done with get balance";
  adb_ptr->set_height_address_val_internal(*txs_ptr, addr_balance);
  LOG(INFO) << "done with set height address";

  auto account_median_ptr =
      get_account_balance_median(*accounts_ptr, *txs_v_ptr, adb_ptr);
  LOG(INFO) << "done with get account balance median";
  auto account_weight_ptr = get_account_weight(in_out_vals, adb_ptr);
  LOG(INFO) << "done with get account weight";

  auto account_median = *account_median_ptr;
  auto account_weight = *account_weight_ptr;
  auto account_rank_ptr = get_account_rank(account_median, account_weight, rp);
  auto account_rank = *account_rank_ptr;
  LOG(INFO) << "account rank size: " << account_rank.size();

  std::vector<std::shared_ptr<nr_info_t>> infos;
  for (auto it = accounts_ptr->begin(); it != accounts_ptr->end(); it++) {
    address_t addr = *it;
    if (in_out_vals.find(addr) != in_out_vals.end() &&
        account_median.find(addr) != account_median.end() &&
        account_weight.find(addr) != account_weight.end() &&
        account_rank.find(addr) != account_rank.end()) {
      auto in_outs = in_out_vals[addr].m_in_val + in_out_vals[addr].m_out_val;
      auto f_in_outs = to_float<floatxx_t>(in_outs);

      auto info = std::shared_ptr<nr_info_t>(
          new nr_info_t({addr, f_in_outs, account_median[addr],
                         account_weight[addr], account_rank[addr]}));
      infos.push_back(info);
    }
  }

  auto end_time = std::chrono::high_resolution_clock::now();
  LOG(INFO) << "time spend: "
            << std::chrono::duration_cast<std::chrono::seconds>(end_time -
                                                                start_time)
                   .count()
            << " seconds";
  return infos;
}

void nebulas_rank::convert_nr_info_to_ptree(const nr_info_t &info,
                                            boost::property_tree::ptree &p) {

  neb::bytes addr_bytes = info.m_address;
  floatxx_t f_in_outs = info.m_in_outs;
  floatxx_t f_median = info.m_median;
  floatxx_t f_weight = info.m_weight;
  floatxx_t f_nr_score = info.m_nr_score;

  p.put(std::string("address"), addr_bytes.to_base58());
  p.put(std::string("in_outs"), neb::math::to_string(f_in_outs));
  p.put(std::string("median"), neb::math::to_string(f_median));
  p.put(std::string("weight"), neb::math::to_string(f_weight));
  p.put(std::string("score"), neb::math::to_string(f_nr_score));
}

void nebulas_rank::full_fill_meta_info(
    const std::vector<std::pair<std::string, std::string>> &meta,
    boost::property_tree::ptree &root) {

  assert(meta.size() == 3);

  for (auto &ele : meta) {
    root.put(ele.first, ele.second);
  }
}

str_uptr_t nebulas_rank::nr_info_to_json(const nr_ret_type &nr_ret) {

  boost::property_tree::ptree root;
  boost::property_tree::ptree arr;

  const auto &meta_info_json = std::get<1>(nr_ret);
  const auto &meta_info = neb::rt::json_to_meta_info(meta_info_json);
  if (!meta_info.empty()) {
    full_fill_meta_info(meta_info, root);
  }

  auto &nr_infos = std::get<2>(nr_ret);
  if (nr_infos.empty()) {
    boost::property_tree::ptree p;
    arr.push_back(std::make_pair(std::string(), p));
  }

  for (auto &it : nr_infos) {
    const neb::rt::nr::nr_info_t &info = *it;
    boost::property_tree::ptree p;
    convert_nr_info_to_ptree(info, p);
    arr.push_back(std::make_pair(std::string(), p));
  }
  root.add_child("nrs", arr);

  std::stringstream ss;
  boost::property_tree::json_parser::write_json(ss, root, false);
  auto tmp_ptr = std::make_unique<std::string>(ss.str());
  boost::replace_all(*tmp_ptr, "[\"\"]", "[]");
  return tmp_ptr;
}

nr_ret_type nebulas_rank::json_to_nr_info(const std::string &nr_result) {

  nr_ret_type nr_ret;
  std::get<0>(nr_ret) = 0;

  boost::property_tree::ptree pt;
  std::stringstream ss(nr_result);
  boost::property_tree::json_parser::read_json(ss, pt);

  auto start_block = pt.get<std::string>("start_height");
  auto end_block = pt.get<std::string>("end_height");
  auto version = pt.get<std::string>("version");

  std::vector<std::pair<std::string, std::string>> meta_info;
  meta_info.push_back(std::make_pair("start_height", start_block));
  meta_info.push_back(std::make_pair("end_height", end_block));
  meta_info.push_back(std::make_pair("version", version));
  std::get<1>(nr_ret) = meta_info_to_json(meta_info);

  boost::property_tree::ptree nrs = pt.get_child("nrs");
  auto &infos = std::get<2>(nr_ret);

  BOOST_FOREACH (boost::property_tree::ptree::value_type &v, nrs) {
    boost::property_tree::ptree nr = v.second;
    auto info_ptr = std::make_shared<nr_info_t>();
    nr_info_t &info = *info_ptr;
    const auto &address = nr.get<base58_address_t>("address");
    info.m_address = neb::bytes::from_base58(address);

    const auto &in_outs = nr.get<std::string>("in_outs");
    const auto &median = nr.get<std::string>("median");
    const auto &weight = nr.get<std::string>("weight");
    const auto &score = nr.get<std::string>("score");
    info.m_in_outs = neb::math::from_string<floatxx_t>(in_outs);
    info.m_median = neb::math::from_string<floatxx_t>(median);
    info.m_weight = neb::math::from_string<floatxx_t>(weight);
    info.m_nr_score = neb::math::from_string<floatxx_t>(score);
    infos.push_back(info_ptr);
  }
  std::get<0>(nr_ret) = 1;

  return nr_ret;
}

str_uptr_t nebulas_rank::get_nr_sum_str(const nr_ret_type &nr_ret) {

  boost::property_tree::ptree root;

  const auto &meta_info_json = std::get<1>(nr_ret);
  const auto &meta_info = neb::rt::json_to_meta_info(meta_info_json);
  if (!meta_info.empty()) {
    full_fill_meta_info(meta_info, root);
  }

  auto &nr_infos = std::get<2>(nr_ret);
  floatxx_t zero = softfloat_cast<uint32_t, typename floatxx_t::value_type>(0);
  floatxx_t sum_in_outs = zero;
  floatxx_t sum_score = zero;

  for (auto &info : nr_infos) {
    sum_in_outs += info->m_in_outs;
    sum_score += info->m_nr_score;
  }

  boost::property_tree::ptree pt;
  auto str_in_outs = neb::math::to_string(from_float(sum_in_outs));
  auto str_score = neb::math::to_string(from_float(sum_score));

  pt.put(std::string("in_outs"), str_in_outs);
  pt.put(std::string("score"), str_score);
  root.add_child("sum", pt);

  std::stringstream ss;
  boost::property_tree::json_parser::write_json(ss, root, false);
  auto tmp_ptr = std::make_unique<std::string>(ss.str());
  return tmp_ptr;
}

} // namespace nr
} // namespace rt
} // namespace neb
