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
#include "core/ir_warden.h"
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

dip_handler::dip_handler() : m_has_curr(false) {}

void dip_handler::check_dip_params(block_height_t height,
                                   neb::fs::rocksdb_storage *rs) {

  if (!m_has_curr && m_incoming.empty()) {
    auto dip_versions_ptr = neb::fs::ir_api::get_ir_versions("dip", rs);
    if (dip_versions_ptr->empty()) {
      return;
    }

    std::reverse(dip_versions_ptr->begin(), dip_versions_ptr->end());
    for (auto version : *dip_versions_ptr) {
      auto nbre_ir_ptr =
          neb::core::ir_warden::instance().get_ir_by_name_version("dip",
                                                                  version);
      block_height_t available_height = nbre_ir_ptr->height();
      m_incoming.push(std::make_pair(version, available_height));
    }
  }

  if (!m_incoming.empty()) {
    auto first_ele = m_incoming.front();
    block_height_t available_height = first_ele.second;
    if (available_height <= height) {
      m_curr = first_ele;
      m_has_curr = true;
      m_incoming.pop();

      try {
        jit_driver &jd = jit_driver::instance();
        jd.run_ir<std::string>("dip", std::numeric_limits<uint64_t>::max(),
                               "_Z15entry_point_dipB5cxx11m", 0);
      } catch (const std::exception &e) {
        LOG(INFO) << "dip params init failed " << e.what();
      }
    }
  }
}

void dip_handler::deploy(version_t version, block_height_t available_height) {
  m_incoming.push(std::make_pair(version, available_height));
}

void dip_handler::start(neb::block_height_t height,
                        neb::fs::rocksdb_storage *rs) {
  check_dip_params(height, rs);
  if (!m_has_curr) {
    LOG(INFO) << "dip params not init";
    return;
  }

  block_height_t dip_start_block =
      neb::configuration::instance().dip_start_block();
  block_height_t dip_block_interval =
      neb::configuration::instance().dip_block_interval();

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

      m_dip_reward.insert(std::make_pair(hash_height, dip_reward));
    } catch (const std::exception &e) {
      LOG(INFO) << "jit driver execute dip failed " << e.what();
    }
  });
}

std::string dip_handler::get_dip_reward(neb::block_height_t height) {
  std::unique_lock<std::mutex> _l(m_sync_mutex);

  if (!m_has_curr) {
    return std::string("{\"err\":\"dip params not init yet\"}");
  }

  block_height_t dip_start_block =
      neb::configuration::instance().dip_start_block();
  block_height_t dip_block_interval =
      neb::configuration::instance().dip_block_interval();

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
