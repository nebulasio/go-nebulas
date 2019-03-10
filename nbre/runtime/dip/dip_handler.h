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
#include "common/common.h"
#include "common/util/singleton.h"
#include "fs/rocksdb_storage.h"
#include "util/thread_safe_map.h"
#include "util/thread_safe_vector.h"
#include <ff/network.h>

define_nt(start_block, neb::block_height_t);
define_nt(block_interval, neb::block_height_t);
define_nt(reward_addr, std::string);
define_nt(coinbase_addr, std::string);
define_nt(version, uint64_t);

namespace neb {
namespace rt {
namespace dip {

typedef ::ff::net::ntpackage<1, start_block, block_interval, reward_addr,
                             coinbase_addr, version>
    dip_params_t;

class dip_handler : public util::singleton<dip_handler> {
public:
  dip_handler();

  void deploy(version_t version, block_height_t available_height);
  void start(block_height_t height, const dip_params_t *dip_params = nullptr);

  std::shared_ptr<dip_params_t> get_dip_params(block_height_t height);
  std::string get_dip_reward(neb::block_height_t height);

private:
  std::string get_dip_reward_when_missing(block_height_t height,
                                          const dip_params_t &dip_params);

  void init_dip_params(block_height_t height);
  void load_dip_rewards();
  void write_to_storage(block_height_t height, const std::string &dip_reward);

private:
  neb::fs::rocksdb_storage *m_storage;
  thread_safe_map<block_height_t, std::string> m_dip_reward;
  // dip params info list
  thread_safe_vector<std::shared_ptr<dip_params_t>> m_dip_params_list;

  bool m_has_curr;
  // suppose version and available height are in increasing order
  std::queue<std::pair<version_t, block_height_t>> m_incoming;
};
} // namespace dip
} // namespace rt
} // namespace neb
