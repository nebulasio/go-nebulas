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
#include "common/common.h"
#include "fs/blockchain/account/account_db_interface.h"
#include "fs/blockchain/data_type.h"
#include "fs/blockchain/transaction/transaction_db_interface.h"
#include "runtime/nr/impl/data_type.h"

#define BLOCK_INTERVAL_MAGIC_NUM 128

namespace neb {
namespace rt {
namespace nr {
class nebulas_rank_algo {
public:
  virtual std::unique_ptr<std::vector<std::vector<neb::fs::transaction_info_t>>>
  split_transactions_by_block_interval(
      const std::vector<neb::fs::transaction_info_t> &txs,
      int32_t block_interval = BLOCK_INTERVAL_MAGIC_NUM);

  virtual void filter_empty_transactions_this_interval(
      std::vector<std::vector<neb::fs::transaction_info_t>> &txs);

  virtual std::unique_ptr<std::vector<transaction_graph_ptr_t>>
  build_transaction_graphs(
      const std::vector<std::vector<neb::fs::transaction_info_t>> &txs);

  virtual std::unique_ptr<std::unordered_set<address_t>>
  get_normal_accounts(const std::vector<neb::fs::transaction_info_t> &txs);

  virtual std::unique_ptr<std::unordered_map<address_t, floatxx_t>>
  get_account_balance_median(
      neb::block_height_t start_block,
      const std::unordered_set<address_t> &accounts,
      const std::vector<std::vector<neb::fs::transaction_info_t>> &txs,
      fs::account_db_interface *db_ptr);

  virtual std::unique_ptr<std::unordered_map<address_t, floatxx_t>>
  get_account_weight(
      const std::unordered_map<address_t, neb::rt::in_out_val_t> &in_out_vals,
      fs::account_db_interface *db_ptr);

  virtual std::unique_ptr<std::unordered_map<address_t, floatxx_t>>
  get_account_rank(
      const std::unordered_map<address_t, floatxx_t> &account_median,
      const std::unordered_map<address_t, floatxx_t> &account_weight,
      const rank_params_t &rp);

  virtual std::unique_ptr<std::vector<address_t>>
  sort_accounts(const std::unordered_set<address_t> &accounts);

protected:
  transaction_graph_ptr_t build_graph_from_transactions(
      const std::vector<neb::fs::transaction_info_t> &trans);

  block_height_t get_max_height_this_block_interval(
      const std::vector<neb::fs::transaction_info_t> &txs);

  floatxx_t f_account_weight(floatxx_t in_val, floatxx_t out_val);

  floatxx_t f_account_rank(int64_t a, int64_t b, int64_t c, int64_t d,
                           floatxx_t theta, floatxx_t mu, floatxx_t lambda,
                           floatxx_t S, floatxx_t R);
};
} // namespace nr
} // namespace rt
} // namespace neb
