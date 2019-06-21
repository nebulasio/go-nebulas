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

#include "runtime/dip/dip_impl.h"
#include "common/address.h"
#include "common/common.h"
#include "common/configuration.h"
#include "fs/blockchain/blockchain_api_test.h"
#include "runtime/dip/dip_reward.h"
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>

namespace neb {
namespace rt {
namespace dip {

dip_ret_type entry_point_dip_impl(compatible_uint64_t start_block,
                                  compatible_uint64_t end_block,
                                  version_t version, compatible_uint64_t height,
                                  const nr::nr_ret_type &nr_ret,
                                  dip_float_t alpha, dip_float_t beta) {

  /*
  std::unique_ptr<neb::fs::blockchain_api_base> pba;
  if (neb::use_test_blockchain) {
    pba = std::unique_ptr<neb::fs::blockchain_api_base>(
        new neb::fs::blockchain_api_test());
  } else {
    pba = std::unique_ptr<neb::fs::blockchain_api_base>(
        new neb::fs::blockchain_api());
  }
  auto tdb_ptr = std::make_unique<neb::fs::transaction_db>(pba.get());
  auto adb_ptr = std::make_unique<neb::fs::account_db>(pba.get());

  dip_ret_type ret;
  dip_result result;
  auto &nr_result = std::get<1>(nr_ret);
  auto dr = dip_reward::get_dip_reward(start_block, end_block, height,
                                       nr_result->get<p_nr_items>(), tdb_ptr,
                                       adb_ptr, alpha, beta);
  std::get<0>(ret) = 1;
  result.set<p_start_block, p_end_block, p_version>(start_block, end_block,
                                                    version);
  result.set<p_dip_items>(dr);
  LOG(INFO) << "get dip reward resurned";

  return ret;*/
  return dip_ret_type();
}

dip_param_t make_dip_param(block_height_t start_block,
                           block_height_t block_interval,
                           const std::string &reward_addr,
                           const std::string &coinbase_addr, version_t v) {

  auto reward_addr_bytes = bytes::from_base58(reward_addr);
  auto coinbase_addr_bytes = bytes::from_base58(coinbase_addr);
  dip_param_t param;
  param.set<p_start_block, p_block_interval, p_version>(start_block,
                                                        block_interval, v);
  param.set<p_dip_reward_addr, p_dip_coinbase_addr>(
      std::to_string(reward_addr_bytes), std::to_string(coinbase_addr_bytes));
  return param;
}
} // namespace dip
} // namespace rt
} // namespace neb
