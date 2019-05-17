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

#pragma once
#include "runtime/nr/impl/nebulas_rank.h"
#include "test/benchmark/account_db_v2.h"
#include "test/benchmark/transaction_db_v2.h"

#define BLOCK_INTERVAL_MAGIC_NUM 128

namespace neb {
namespace rt {
namespace nr {

using transaction_db_v2_ptr_t = std::unique_ptr<neb::fs::transaction_db_v2>;
using account_db_v2_ptr_t = std::unique_ptr<neb::fs::account_db_v2>;

class nebulas_rank_v2 {
public:
  static std::vector<std::shared_ptr<nr_info_t>>
  get_nr_score(const transaction_db_v2_ptr_t &tdb_ptr,
               const account_db_v2_ptr_t &adb_ptr, const rank_params_t &rp,
               neb::block_height_t start_block, neb::block_height_t end_block);

private:
  static auto split_transactions_by_block_interval(
      const std::vector<neb::fs::transaction_info_t> &txs,
      int32_t block_interval = BLOCK_INTERVAL_MAGIC_NUM)
      -> std::unique_ptr<std::vector<std::vector<neb::fs::transaction_info_t>>>;

  static void filter_empty_transactions_this_interval(
      std::vector<std::vector<neb::fs::transaction_info_t>> &txs);

  static auto build_transaction_graphs(
      const std::vector<std::vector<neb::fs::transaction_info_t>> &txs)
      -> std::unique_ptr<std::vector<transaction_graph_ptr_t>>;

  static auto
  get_normal_accounts(const std::vector<neb::fs::transaction_info_t> &txs)
      -> std::unique_ptr<std::unordered_set<address_t>>;

  static auto get_account_balance_median(
      neb::block_height_t start_block,
      const std::unordered_set<address_t> &accounts,
      const std::vector<std::vector<neb::fs::transaction_info_t>> &txs,
      const account_db_v2_ptr_t &db_ptr)
      -> std::unique_ptr<std::unordered_map<address_t, floatxx_t>>;

  static auto get_account_weight(
      const std::unordered_map<address_t, neb::rt::in_out_val_t> &in_out_vals,
      const account_db_v2_ptr_t &db_ptr)
      -> std::unique_ptr<std::unordered_map<address_t, floatxx_t>>;

  static auto get_account_rank(
      const std::unordered_map<address_t, floatxx_t> &account_median,
      const std::unordered_map<address_t, floatxx_t> &account_weight,
      const rank_params_t &rp)
      -> std::unique_ptr<std::unordered_map<address_t, floatxx_t>>;

  static transaction_graph_ptr_t build_graph_from_transactions(
      const std::vector<neb::fs::transaction_info_t> &trans);

  static block_height_t get_max_height_this_block_interval(
      const std::vector<neb::fs::transaction_info_t> &txs);

  static floatxx_t f_account_weight(floatxx_t in_val, floatxx_t out_val);

  static floatxx_t f_account_rank(int64_t a, int64_t b, int64_t c, int64_t d,
                                  floatxx_t theta, floatxx_t mu,
                                  floatxx_t lambda, floatxx_t S, floatxx_t R);

}; // class nebulas_rank

} // namespace nr
} // namespace rt
} // namespace neb
