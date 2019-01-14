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
#include "common/common.h"
#include "common/configuration.h"
#include "core/neb_ipc/server/ipc_configuration.h"
#include "runtime/dip/dip_reward.h"
#include "runtime/nr/impl/nebulas_rank.h"

namespace neb {
namespace rt {
namespace dip {

std::string entry_point_dip_impl(uint64_t start_block, uint64_t end_block,
                                 version_t version, uint64_t height,
                                 const std::string &nr_result,
                                 dip_float_t alpha, dip_float_t beta) {

  std::string neb_db_path =
      neb::core::ipc_configuration::instance().neb_db_dir();
  neb::fs::blockchain bc(neb_db_path);
  neb::fs::blockchain_api ba(&bc);
  nr::transaction_db_ptr_t tdb_ptr =
      std::make_unique<neb::fs::transaction_db>(&ba);
  nr::account_db_ptr_t adb_ptr = std::make_unique<neb::fs::account_db>(&ba);

  auto ret = dip_reward::get_dip_reward(
      start_block, end_block, height, nr_result, tdb_ptr, adb_ptr, alpha, beta);
  std::vector<std::pair<std::string, uint64_t>> meta_info(
      {{"start_height", start_block},
       {"end_height", end_block},
       {"version", version}});

  return dip_reward::dip_info_to_json(*ret, meta_info);
}

void init_dip_params(uint64_t dip_start_block, uint64_t dip_block_interval,
                     const std::string &dip_reward_addr) {
  neb::configuration::instance().dip_start_block() = dip_start_block;
  neb::configuration::instance().dip_block_interval() = dip_block_interval;

  neb::util::bytes addr_bytes = neb::util::bytes::from_base58(dip_reward_addr);
  neb::configuration::instance().dip_reward_addr() =
      neb::util::byte_to_string(addr_bytes);

  LOG(INFO) << "init dip params, dip_start_block " << dip_start_block
            << ", dip_block_interval " << dip_block_interval
            << ", dip_reward_addr " << dip_reward_addr;
}

void init_dip_params(uint64_t dip_start_block, uint64_t dip_block_interval,
                     const std::string &dip_reward_addr,
                     const std::string &coinbase_addr) {
  neb::configuration::instance().dip_start_block() = dip_start_block;
  neb::configuration::instance().dip_block_interval() = dip_block_interval;

  neb::util::bytes reward_bytes =
      neb::util::bytes::from_base58(dip_reward_addr);
  neb::configuration::instance().dip_reward_addr() =
      neb::util::byte_to_string(reward_bytes);

  neb::util::bytes coinbase_bytes =
      neb::util::bytes::from_base58(coinbase_addr);
  neb::configuration::instance().coinbase_addr() =
      neb::util::byte_to_string(coinbase_bytes);

  LOG(INFO) << "init dip params, dip_start_block " << dip_start_block
            << ", dip_block_interval " << dip_block_interval
            << ", dip_reward_addr " << dip_reward_addr << ", coinbase_addr "
            << coinbase_addr;
}

} // namespace dip
} // namespace rt
} // namespace neb
