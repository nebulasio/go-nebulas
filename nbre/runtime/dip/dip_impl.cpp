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
#include "runtime/dip/dip_handler.h"
#include "runtime/dip/dip_reward.h"
#include "runtime/nr/impl/nebulas_rank.h"
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

  std::unique_ptr<neb::fs::blockchain_api_base> pba;
  if (neb::use_test_blockchain) {
    pba = std::unique_ptr<neb::fs::blockchain_api_base>(
        new neb::fs::blockchain_api_test());
  } else {
    pba = std::unique_ptr<neb::fs::blockchain_api_base>(
        new neb::fs::blockchain_api());
  }
  nr::transaction_db_ptr_t tdb_ptr =
      std::make_unique<neb::fs::transaction_db>(pba.get());
  nr::account_db_ptr_t adb_ptr =
      std::make_unique<neb::fs::account_db>(pba.get());

  std::vector<std::pair<std::string, std::string>> meta_info;
  meta_info.push_back(
      std::make_pair("start_height", std::to_string(start_block)));
  meta_info.push_back(std::make_pair("end_height", std::to_string(end_block)));
  meta_info.push_back(std::make_pair("version", std::to_string(version)));

  dip_ret_type ret;
  std::get<0>(ret) = 1;
  std::get<1>(ret) = meta_info_to_json(meta_info);

  auto &nr_result = std::get<2>(nr_ret);
  std::get<2>(ret) = dip_reward::get_dip_reward(
      start_block, end_block, height, nr_result, tdb_ptr, adb_ptr, alpha, beta);
  LOG(INFO) << "get dip reward resurned";

  std::get<3>(ret) = nr_ret;
  LOG(INFO) << "append nr_ret to dip_ret";

  return ret;
}

std::string dip_param_list(compatible_uint64_t dip_start_block,
                           compatible_uint64_t dip_block_interval,
                           const std::string &dip_reward_addr,
                           const std::string &dip_coinbase_addr, version_t v) {

  auto reward_addr_bytes = bytes::from_base58(dip_reward_addr);
  auto coinbase_addr_bytes = bytes::from_base58(dip_coinbase_addr);

  dip_params_t param;
  param.set<start_block>(dip_start_block);
  param.set<block_interval>(dip_block_interval);
  param.set<reward_addr>(address_to_string(reward_addr_bytes));
  param.set<coinbase_addr>(address_to_string(coinbase_addr_bytes));
  param.set<p_version>(v);

  LOG(INFO) << "init dip params: " << dip_start_block << ','
            << dip_block_interval << ',' << dip_reward_addr << ','
            << dip_coinbase_addr << ',' << v;
  LOG(INFO) << "begin to seri";
  auto ret = param.serialize_to_string();
  LOG(INFO) << "seri done";
  return ret;
}
} // namespace dip
} // namespace rt
} // namespace neb
