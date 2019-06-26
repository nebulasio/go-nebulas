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
#include "runtime/dip/dip_calculator.h"
#include "common/int128_conversion.h"
#include "fs/blockchain/account/account_db_interface.h"
#include "fs/blockchain/transaction/transaction_algo.h"
#include "fs/blockchain/transaction/transaction_db_interface.h"
#include "runtime/dip/dip_algo.h"

namespace neb {
namespace rt {
namespace dip {
dip_calculator::dip_calculator(dip_algo *algo,
                               fs::transaction_db_interface *tdb_ptr,
                               fs::account_db_interface *adb_ptr)
    : m_algo(algo), m_tdb_ptr(tdb_ptr), m_adb_ptr(adb_ptr) {}

std::vector<dip_item> dip_calculator::get_dip_reward(
    block_height_t start_block, block_height_t end_block,
    const nr::nr_ret_type &nr_result, floatxx_t alpha, floatxx_t beta,
    const dip_param_t &dip_param) {

  auto txs = m_tdb_ptr->read_transactions_from_db_with_duration(start_block,
                                                                end_block);
  LOG(INFO) << "transaction size " << txs.size();

  auto it_nr_infos = nr_result->get<p_nr_items>();
  auto acc_to_contract_txs = fs::algo::read_transactions_with_address_type(
      txs, NAS_ADDRESS_ACCOUNT_MAGIC_NUM, NAS_ADDRESS_CONTRACT_MAGIC_NUM);
  m_algo->ignore_account_transfer_contract(acc_to_contract_txs, "binary");
  LOG(INFO) << "account to contract size " << acc_to_contract_txs.size();
  // dapp total votes
  auto acc_to_contract_votes =
      m_algo->account_to_contract_votes(acc_to_contract_txs, it_nr_infos);
  LOG(INFO) << "account to contract votes " << acc_to_contract_votes.size();
  auto dapp_votes = m_algo->dapp_votes(acc_to_contract_votes);
  LOG(INFO) << "dapp votes size " << dapp_votes.size();

  // bonus pool in total
  address_t dip_reward_addr = to_address(dip_param.get<p_dip_reward_addr>());
  address_t dip_coinbase_addr =
      to_address(dip_param.get<p_dip_coinbase_addr>());

  wei_t balance = m_adb_ptr->get_balance(dip_reward_addr, end_block);
  floatxx_t bonus_total = to_float<floatxx_t>(balance);
  LOG(INFO) << "bonus total " << bonus_total;
  // bonus_total = adb_ptr->get_normalized_value(bonus_total);

  floatxx_t sum_votes(0);
  for (auto &v : dapp_votes) {
    sum_votes += v.second * v.second;
  }
  LOG(INFO) << "sum votes " << sum_votes;

  floatxx_t reward_sum(0);
  std::vector<dip_item> dip_infos;
  for (auto &v : dapp_votes) {
    dip_item di;

    floatxx_t reward_in_wei =
        v.second * v.second *
        m_algo->participate_lambda(alpha, beta, acc_to_contract_txs,
                                   it_nr_infos) *
        bonus_total / sum_votes;
    reward_sum += reward_in_wei;

    di.set<p_dip_contract, p_dip_deployer, p_dip_reward_value>(
        std::to_string(v.first),
        std::to_string(m_adb_ptr->get_contract_deployer(v.first)),
        reward_in_wei);
    dip_infos.push_back(di);
  }
  LOG(INFO) << "reward sum " << reward_sum << ", bonus total " << bonus_total;
  // assert(reward_sum <= bonus_total);
  m_algo->back_to_coinbase(dip_infos, bonus_total - reward_sum,
                           dip_coinbase_addr);
  LOG(INFO) << "back to coinbase";
  return dip_infos;
}
} // namespace dip
} // namespace rt
} // namespace neb
