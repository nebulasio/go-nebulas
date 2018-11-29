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
  double m_a;
  double m_b;
  double m_c;
  double m_d;
  double m_mu;
  double m_lambda;
};

using transaction_db_ptr_t = std::shared_ptr<neb::fs::transaction_db>;
using account_db_ptr_t = std::shared_ptr<neb::fs::account_db>;
using transaction_graph_ptr_t = std::shared_ptr<neb::rt::transaction_graph>;

class nebulas_rank {
public:
  nebulas_rank();
  nebulas_rank(const transaction_db_ptr_t tdb_ptr,
               const account_db_ptr_t adb_ptr, const rank_params_t &rp,
               block_height_t start_block, block_height_t end_block);

public:
  auto split_transactions_by_block_interval(
      const std::vector<neb::fs::transaction_info_t> &txs,
      int32_t block_interval = 128)
      -> std::shared_ptr<std::vector<std::vector<neb::fs::transaction_info_t>>>;

  void filter_empty_transactions_this_interval(
      std::vector<std::vector<neb::fs::transaction_info_t>> &txs);

  auto build_transaction_graphs(
      const std::vector<std::vector<neb::fs::transaction_info_t>> &txs)
      -> std::vector<transaction_graph_ptr_t>;

  auto get_normal_accounts(const std::vector<neb::fs::transaction_info_t> &txs)
      -> std::shared_ptr<std::unordered_set<std::string>>;

  auto get_account_balance(const std::unordered_set<address_t> &accounts,
                           const account_db_ptr_t db_ptr)
      -> std::shared_ptr<std::unordered_map<address_t, wei_t>>;

  auto get_account_balance_median(
      const std::unordered_set<std::string> &accounts,
      const std::vector<std::vector<neb::fs::transaction_info_t>> &txs,
      const account_db_ptr_t db_ptr,
      std::unordered_map<address_t, wei_t> &addr_balance)
      -> std::shared_ptr<std::unordered_map<std::string, double>>;

  auto get_account_weight(
      const std::unordered_map<std::string, neb::rt::in_out_val_t> &in_out_vals,
      const account_db_ptr_t db_ptr)
      -> std::shared_ptr<std::unordered_map<std::string, double>>;

  auto get_account_rank(
      const std::unordered_map<std::string, double> &account_median,
      const std::unordered_map<std::string, double> &account_weight,
      const rank_params_t &rp)
      -> std::shared_ptr<std::unordered_map<std::string, double>>;

private:
  template <class TransInfo>
  transaction_graph_ptr_t
  build_graph_from_transactions(const std::vector<TransInfo> &trans);

  block_height_t get_max_height_this_block_interval(
      const std::vector<neb::fs::transaction_info_t> &txs);

  double f_account_weight(double in_val, double out_val);

  double f_account_rank(double a, double b, double c, double d, double mu,
                        double lambda, double S, double R);

private:
  transaction_db_ptr_t m_tdb_ptr;
  account_db_ptr_t m_adb_ptr;
  rank_params_t m_rp;
  block_height_t m_start_block;
  block_height_t m_end_block;
}; // class nebulas_rank
} // namespace nr
} // namespace rt
} // namespace neb

