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

#include "runtime/dip/dip_handler.h"
#include "common/configuration.h"
#include "core/neb_ipc/server/ipc_configuration.h"
#include "fs/ir_manager/api/ir_api.h"
#include "fs/proto/ir.pb.h"
#include "jit/jit_driver.h"
#include "runtime/dip/dip_reward.h"
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>
#include <ff/ff.h>

namespace neb {
namespace rt {
namespace dip {

void dip_handler::start(neb::block_height_t height,
                        neb::fs::rocksdb_storage *rs) {
  block_height_t dip_start_block =
      neb::configuration::instance().dip_start_block();
  block_height_t dip_block_interval =
      neb::configuration::instance().dip_block_interval();

  if (!dip_start_block || !dip_block_interval) {
    return;
  }

  if (height < dip_start_block + dip_block_interval) {
    LOG(INFO) << "wait to sync";
    return;
  }

  uint64_t interval_nums = (height - dip_start_block) / dip_block_interval;
  uint64_t hash_height = dip_start_block + dip_block_interval * interval_nums;

  if (height != hash_height) {
    return;
  }

  if (m_dip_reward.find(hash_height) != m_dip_reward.end()) {
    return;
  }

  auto dip_versions_ptr = neb::fs::ir_api::get_ir_versions("dip", rs);
  uint64_t dip_version = *dip_versions_ptr->begin();

  ff::para<> p;
  p([this, hash_height, dip_block_interval, dip_version]() {
    try {
      std::unique_lock<std::mutex> _l(m_sync_mutex);
      jit_driver &jd = jit_driver::instance();
      auto dip_reward = jd.run_ir<std::string>(
          "dip", hash_height, "_Z15entry_point_dipB5cxx11m", hash_height);
      LOG(INFO) << "dip reward returned";

      auto it_dip_infos = dip_reward::json_to_dip_info(dip_reward);
      dip_reward = dip_reward::dip_info_to_json(
          *it_dip_infos, {{"start_height", hash_height - dip_block_interval},
                          {"end_height", hash_height - 1},
                          {"version", dip_version}});
      LOG(INFO) << "dip reward meta info returned";

      m_dip_reward.insert(std::make_pair(hash_height, dip_reward));
    } catch (const std::exception &e) {
      LOG(INFO) << "jit driver execute dip failed " << e.what();
    }
  });
}

std::string dip_handler::get_dip_reward(neb::block_height_t height) {
  std::unique_lock<std::mutex> _l(m_sync_mutex);

  block_height_t dip_start_block =
      neb::configuration::instance().dip_start_block();
  block_height_t dip_block_interval =
      neb::configuration::instance().dip_block_interval();
  if (!dip_start_block || !dip_block_interval) {
    return std::string("{\"err\":\"dip params not init yet\"}");
  }

  uint64_t interval_nums = (height - dip_start_block) / dip_block_interval;
  height = dip_start_block + dip_block_interval * interval_nums;

  auto dip_reward = m_dip_reward.find(height);
  if (dip_reward == m_dip_reward.end()) {
    return std::string("{\"err\":\"dip this interval not found\"}");
  }
  return dip_reward->second;
}
} // namespace dip
} // namespace rt
} // namespace neb
