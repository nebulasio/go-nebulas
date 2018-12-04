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

#pragma once

#include "fs/blockchain/account/account_db.h"
#include "fs/blockchain/transaction/transaction_db.h"
#include "runtime/nr/graph/algo.h"

namespace neb {
namespace rt {

namespace nr {

struct rank_params_t {
  float64 m_a;
  float64 m_b;
  float64 m_c;
  float64 m_d;
  float64 m_mu;
  float64 m_lambda;
};

using transaction_db_ptr_t = std::shared_ptr<neb::fs::transaction_db>;
using account_db_ptr_t = std::shared_ptr<neb::fs::account_db>;
using transaction_graph_ptr_t = std::shared_ptr<neb::rt::transaction_graph>;

class nebulas_rank {
public:
  static auto split_transactions_by_block_interval(
      const std::vector<neb::fs::transaction_info_t> &txs,
      int32_t block_interval = 128)
      -> std::shared_ptr<std::vector<std::vector<neb::fs::transaction_info_t>>>;

  static void filter_empty_transactions_this_interval(
      std::vector<std::vector<neb::fs::transaction_info_t>> &txs);

  static auto build_transaction_graphs(
      const std::vector<std::vector<neb::fs::transaction_info_t>> &txs)
      -> std::vector<transaction_graph_ptr_t>;

  static auto
  get_normal_accounts(const std::vector<neb::fs::transaction_info_t> &txs)
      -> std::shared_ptr<std::unordered_set<std::string>>;

  static auto get_account_balance_median(
      const std::unordered_set<std::string> &accounts,
      const std::vector<std::vector<neb::fs::transaction_info_t>> &txs,
      const account_db_ptr_t db_ptr,
      std::unordered_map<address_t, wei_t> &addr_balance)
      -> std::shared_ptr<std::unordered_map<std::string, float64>>;

  static auto get_account_weight(
      const std::unordered_map<std::string, neb::rt::in_out_val_t> &in_out_vals,
      const account_db_ptr_t db_ptr)
      -> std::shared_ptr<std::unordered_map<std::string, float64>>;

  static auto get_account_rank(
      const std::unordered_map<std::string, float64> &account_median,
      const std::unordered_map<std::string, float64> &account_weight,
      const rank_params_t &rp)
      -> std::shared_ptr<std::unordered_map<std::string, float64>>;

private:
  static transaction_graph_ptr_t build_graph_from_transactions(
      const std::vector<neb::fs::transaction_info_t> &trans);

  static block_height_t get_max_height_this_block_interval(
      const std::vector<neb::fs::transaction_info_t> &txs);

  static float64 max(const float64 &x, const float64 &y) {
    return x > y ? x : y;
  }

  static float64 f_account_weight(float64 in_val, float64 out_val);

  static float64 f_account_rank(float64 a, float64 b, float64 c, float64 d,
                                float64 mu, float64 lambda, float64 S,
                                float64 R);
}; // class nebulas_rank
} // namespace nr
} // namespace rt
} // namespace neb

