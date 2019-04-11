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

#include "common/common.h"
#include "common/util/singleton.h"
#include "fs/rocksdb_storage.h"

namespace neb {
namespace rt {
namespace dip {

class dip_handler : public util::singleton<dip_handler> {
public:
  dip_handler();

  void start(neb::block_height_t height, neb::fs::rocksdb_storage *rs);
  void deploy(version_t version, block_height_t available_height);

  std::string get_dip_reward(neb::block_height_t height);

  void read_dip_reward_from_storage(neb::fs::rocksdb_storage *rs);
  void write_dip_reward_to_storage(const std::string &dip_reward,
                                   neb::fs::rocksdb_storage *rs);

  void init_dip_params(block_height_t height, neb::fs::rocksdb_storage *rs);

private:
  std::string get_dip_reward_when_missing(neb::block_height_t height);

private:
  mutable std::mutex m_sync_mutex;
  std::map<neb::block_height_t, std::string> m_dip_reward;

  bool m_has_curr;
  std::pair<version_t, block_height_t> m_curr;
  // suppose version and available height are in increasing order
  std::queue<std::pair<version_t, block_height_t>> m_incoming;
};
} // namespace dip
} // namespace rt
} // namespace neb
