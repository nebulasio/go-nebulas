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
#include "fs/proto/ir.pb.h"
#include "jit/jit_driver.h"
#include "runtime/dip/dip_reward.h"
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>
#include <ff/ff.h>

namespace neb {
namespace rt {
namespace dip {

void dip_handler::start(neb::block_height_t nbre_max_height,
                        neb::block_height_t lib_height) {
  std::unique_lock<std::mutex> _l(m_sync_mutex);
  block_height_t dip_start_block =
      neb::configuration::instance().dip_start_block();
  block_height_t dip_block_interval =
      neb::configuration::instance().dip_block_interval();

  if (!dip_start_block || !dip_block_interval) {
    return;
  }

  if (nbre_max_height < dip_start_block + dip_block_interval) {
    return;
  }

  assert(nbre_max_height <= lib_height);
  if (nbre_max_height + dip_block_interval < lib_height) {
    return;
  }

  uint64_t interval_nums =
      (nbre_max_height - dip_start_block) / dip_block_interval;
  uint64_t height = dip_start_block + dip_block_interval * interval_nums;

  if (m_dip_reward.find(height) != m_dip_reward.end()) {
    return;
  }

  ff::para<> p;
  p([this, height, dip_block_interval]() {
    try {
      jit_driver &jd = jit_driver::instance();
      auto dip_reward = jd.run_ir<std::string>(
          "dip", height, "_Z15entry_point_dipB5cxx11m", height);

      auto it_dip_infos = dip_reward::json_to_dip_info(dip_reward);
      dip_reward = dip_reward::dip_info_to_json(
          *it_dip_infos, {{"start_height", height - dip_block_interval},
                          {"end_height", height - 1}});

      m_dip_reward.insert(std::make_pair(height, dip_reward));
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
  uint64_t interval_nums = (height - dip_start_block) / dip_block_interval;
  height = dip_start_block + dip_block_interval * interval_nums;

  auto dip_reward = m_dip_reward.find(height);
  if (dip_reward == m_dip_reward.end()) {
    return std::string("{\"err\":\"not complete yet\"}");
  }
  return dip_reward->second;
}
} // namespace dip
} // namespace rt
} // namespace neb
