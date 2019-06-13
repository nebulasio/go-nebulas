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
#include "runtime/nr/impl/nebulas_rank_calculator.h"
#include "common/int128_conversion.h"
#include "fs/blockchain/account/account_db_interface.h"
#include "fs/blockchain/transaction/transaction_algo.h"
#include "fs/blockchain/transaction/transaction_db_interface.h"
#include "runtime/nr/impl/nebulas_rank_algo.h"

namespace neb {
namespace rt {
namespace nr {

nebulas_rank_calculator::nebulas_rank_calculator(
    nebulas_rank_algo *algo, fs::transaction_db_interface *tdb_ptr,
    fs::account_db_interface *adb_ptr)
    : m_algo(algo), m_tdb_ptr(tdb_ptr), m_adb_ptr(adb_ptr) {}

std::vector<nr_item>
nebulas_rank_calculator::get_nr_score(const rank_params_t &rp,
                                      neb::block_height_t start_block,
                                      neb::block_height_t end_block) {

  auto start_time = std::chrono::high_resolution_clock::now();
  LOG(INFO) << "start to read neb db transaction";
  // transactions in total and account inter transactions
  auto txs_ptr = m_tdb_ptr->read_transactions_from_db_with_duration(start_block,
                                                                    end_block);
  LOG(INFO) << "raw tx size: " << txs_ptr->size();
  auto inter_txs_ptr = fs::algo::read_transactions_with_address_type(
      *txs_ptr, NAS_ADDRESS_ACCOUNT_MAGIC_NUM, NAS_ADDRESS_ACCOUNT_MAGIC_NUM);
  LOG(INFO) << "account to account: " << inter_txs_ptr->size();
  auto succ_inter_txs_ptr =
      fs::algo::read_transactions_with_succ(*inter_txs_ptr);
  LOG(INFO) << "succ account to account: " << succ_inter_txs_ptr->size();

  // graph operation
  auto txs_v_ptr =
      m_algo->split_transactions_by_block_interval(*succ_inter_txs_ptr);
  LOG(INFO) << "split by block interval: " << txs_v_ptr->size();

  m_algo->filter_empty_transactions_this_interval(*txs_v_ptr);
  auto tgs_ptr = m_algo->build_transaction_graphs(*txs_v_ptr);
  if (tgs_ptr->empty()) {
    return std::vector<nr_item>();
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
  auto accounts_ptr = m_algo->get_normal_accounts(*inter_txs_ptr);
  LOG(INFO) << "account size: " << accounts_ptr->size();

  std::unordered_map<neb::address_t, neb::wei_t> addr_balance;
  for (auto &acc : *accounts_ptr) {
    auto balance = m_adb_ptr->get_balance(acc, start_block);
    addr_balance.insert(std::make_pair(acc, balance));
  }
  LOG(INFO) << "done with get balance";
  m_adb_ptr->update_height_address_val_internal(start_block, *txs_ptr,
                                                addr_balance);
  LOG(INFO) << "done with set height address";

  auto account_median_ptr = m_algo->get_account_balance_median(
      start_block, *accounts_ptr, *txs_v_ptr, m_adb_ptr);
  LOG(INFO) << "done with get account balance median";
  auto account_weight_ptr = m_algo->get_account_weight(in_out_vals, m_adb_ptr);
  LOG(INFO) << "done with get account weight";

  auto account_median = *account_median_ptr;
  auto account_weight = *account_weight_ptr;
  auto account_rank_ptr =
      m_algo->get_account_rank(account_median, account_weight, rp);
  auto account_rank = *account_rank_ptr;
  LOG(INFO) << "account rank size: " << account_rank.size();

  auto sorted_accounts_ptr = m_algo->sort_accounts(*accounts_ptr);
  std::vector<nr_item> infos;
  for (auto it = sorted_accounts_ptr->begin(); it != sorted_accounts_ptr->end();
       it++) {
    address_t addr = *it;
    if (in_out_vals.find(addr) != in_out_vals.end() &&
        account_median.find(addr) != account_median.end() &&
        account_weight.find(addr) != account_weight.end() &&
        account_rank.find(addr) != account_rank.end()) {
      auto in_outs = in_out_vals[addr].m_in_val + in_out_vals[addr].m_out_val;
      auto f_in_outs = to_float<floatxx_t>(in_outs);

      nr_item info;
      info.set<p_nr_item_addr, p_nr_item_score, p_nr_item_weight,
               p_nr_item_in_outs, p_nr_item_median>(
          std::to_string(addr), account_rank[addr], account_weight[addr],
          f_in_outs, account_median[addr]);
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

} // namespace nr
} // namespace rt
} // namespace neb
