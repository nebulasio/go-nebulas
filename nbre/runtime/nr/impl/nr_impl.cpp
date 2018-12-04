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

#include "runtime/nr/impl/nr_impl.h"
#include "common/common.h"
#include "common/configuration.h"
#include "common/util/int_conversion.h"
#include "fs/blockchain/nebulas_currency.h"
#include "runtime/nr/impl/nebulas_rank.h"

std::vector<nr_info_t> entry_point_nr_impl(neb::core::driver *d, void *param) {
  neb::block_height_t start_block = 1111780;
  neb::block_height_t end_block = 1117539;

  std::string neb_db_path = neb::configuration::instance().neb_db_dir();
  neb::fs::blockchain bc(neb_db_path);
  neb::fs::blockchain_api ba(&bc);
  neb::rt::nr::transaction_db_ptr_t tdb_ptr =
      std::make_shared<neb::fs::transaction_db>(&ba);
  neb::rt::nr::account_db_ptr_t adb_ptr =
      std::make_shared<neb::fs::account_db>(&ba);

  LOG(INFO) << "start block: " << start_block << " , end block: " << end_block;
  neb::rt::nr::rank_params_t rp{2000.0, 200000.0, 100.0, 1000.0, 0.75, 3.14};

  auto it_txs =
      tdb_ptr->read_transactions_from_db_with_duration(start_block, end_block);
  // account inter transactions
  auto it_account_inter_txs = tdb_ptr->read_account_inter_transactions(*it_txs);
  auto account_inter_txs = *it_account_inter_txs;
  LOG(INFO) << "account to account: " << account_inter_txs.size();

  // graph
  auto it_txs_v =
      neb::rt::nr::nebulas_rank::split_transactions_by_block_interval(
          account_inter_txs);
  auto txs_v = *it_txs_v;
  LOG(INFO) << "split by block interval: " << txs_v.size();

  neb::rt::nr::transaction_graph_ptr_t tg =
      std::make_shared<neb::rt::transaction_graph>();

  neb::rt::nr::nebulas_rank::filter_empty_transactions_this_interval(txs_v);
  std::vector<neb::rt::nr::transaction_graph_ptr_t> tgs =
      neb::rt::nr::nebulas_rank::build_transaction_graphs(txs_v);
  if (tgs.empty()) {
    return std::vector<nr_info_t>();
  }
  LOG(INFO) << "we have " << tgs.size() << " subgraphs.";
  for (auto it = tgs.begin(); it != tgs.end(); it++) {
    neb::rt::nr::transaction_graph_ptr_t ptr = *it;
    neb::rt::graph_algo::remove_cycles_based_on_time_sequence(
        ptr->internal_graph());
    neb::rt::graph_algo::merge_edges_with_same_from_and_same_to(
        ptr->internal_graph());
  }
  LOG(INFO) << "done with remove cycle.";

  tg = neb::rt::graph_algo::merge_graphs(tgs);
  neb::rt::graph_algo::merge_topk_edges_with_same_from_and_same_to(
      tg->internal_graph());
  LOG(INFO) << "done with merge graphs.";

  // median
  auto it_accounts =
      neb::rt::nr::nebulas_rank::get_normal_accounts(account_inter_txs);
  auto accounts = *it_accounts;
  LOG(INFO) << "account size: " << accounts.size();

  std::unordered_map<neb::address_t, neb::wei_t> addr_balance;
  for (auto &acc : accounts) {
    auto balance = adb_ptr->get_balance(acc, start_block);
    addr_balance.insert(std::make_pair(acc, balance));
  }
  adb_ptr->set_height_address_val_internal(*it_txs, addr_balance);

  auto it_account_median =
      neb::rt::nr::nebulas_rank::get_account_balance_median(
          accounts, txs_v, adb_ptr, addr_balance);
  auto account_median = *it_account_median;

  // degree and in_out amount
  auto it_in_out_degrees =
      neb::rt::graph_algo::get_in_out_degrees(tg->internal_graph());
  auto in_out_degrees = *it_in_out_degrees;
  auto it_degrees = neb::rt::graph_algo::get_degree_sum(tg->internal_graph());
  auto degrees = *it_degrees;
  auto it_in_out_vals =
      neb::rt::graph_algo::get_in_out_vals(tg->internal_graph());
  auto in_out_vals = *it_in_out_vals;
  auto it_stakes = neb::rt::graph_algo::get_stakes(tg->internal_graph());
  auto stakes = *it_stakes;

  // weight and rank
  auto it_account_weight =
      neb::rt::nr::nebulas_rank::get_account_weight(in_out_vals, adb_ptr);
  auto account_weight = *it_account_weight;
  auto it_account_rank = neb::rt::nr::nebulas_rank::get_account_rank(
      account_median, account_weight, rp);
  auto account_rank = *it_account_rank;
  LOG(INFO) << "account rank size: " << account_rank.size();

  std::vector<nr_info_t> infos;
  for (auto it = accounts.begin(); it != accounts.end(); it++) {
    std::string addr = *it;
    if (account_median.find(addr) == account_median.end() ||
        account_rank.find(addr) == account_rank.end() ||
        in_out_degrees.find(addr) == in_out_degrees.end() ||
        in_out_vals.find(addr) == in_out_vals.end() ||
        stakes.find(addr) == stakes.end()) {
      continue;
    }

    float64 nas_in_val = adb_ptr->get_normalized_value(
        neb::detail_int128(in_out_vals.find(addr)->second.m_in_val)
            .to_float64());
    float64 nas_out_val = adb_ptr->get_normalized_value(
        neb::detail_int128(in_out_vals.find(addr)->second.m_out_val)
            .to_float64());
    float64 nas_stake = adb_ptr->get_normalized_value(
        neb::detail_int128(stakes.find(addr)->second).to_float64());

    nr_info_t info{addr,
                   in_out_degrees[addr].m_in_degree,
                   in_out_degrees[addr].m_out_degree,
                   degrees[addr],
                   nas_in_val,
                   nas_out_val,
                   nas_stake,
                   account_median[addr],
                   account_weight[addr],
                   account_rank[addr]};
    infos.push_back(info);

    if (account_rank[addr] > 0) {
      neb::util::bytes addr_bytes = neb::util::string_to_byte(addr);
      LOG(INFO) << addr_bytes.to_base58() << ',' << degrees[addr] << ','
                << nas_stake << ',' << account_median[addr] << ','
                << account_weight[addr] << ',' << account_rank[addr];
    }
  }

  return infos;
}
