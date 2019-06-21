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
#include "common/address.h"
#include "fs/blockchain/data_type.h"
#include "runtime/dip/data_type.h"

namespace neb {
namespace rt {
namespace dip {
class dip_algo {
public:
  virtual std::unordered_map<address_t, std::unordered_map<address_t, uint32_t>>
  account_call_contract_count(
      const std::vector<neb::fs::transaction_info_t> &txs);

  virtual std::unordered_map<address_t,
                             std::unordered_map<address_t, floatxx_t>>
  account_to_contract_votes(const std::vector<neb::fs::transaction_info_t> &txs,
                            const std::vector<nr_item> &nr_infos);

  virtual std::unordered_map<address_t, floatxx_t>
  dapp_votes(const std::unordered_map<address_t,
                                      std::unordered_map<address_t, floatxx_t>>
                 &acc_contract_votes);

  virtual floatxx_t
  participate_lambda(floatxx_t alpha, floatxx_t beta,
                     const std::vector<neb::fs::transaction_info_t> &txs,
                     const std::vector<nr_item> &nr_infos);

  virtual void back_to_coinbase(std::vector<dip_item> &dip_infos,
                                floatxx_t reward_left,
                                const address_t &coinbase_addr);

  virtual void ignore_account_transfer_contract(
      std::vector<neb::fs::transaction_info_t> &txs,
      const std::string &tx_type);
};
} // namespace dip
} // namespace rt
} // namespace neb
