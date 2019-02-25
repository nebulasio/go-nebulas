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
#include "fs/fs_storage.h"
#include "fs/ir_manager/api/ir_api.h"
#include "fs/proto/ir.pb.h"
#include "jit/jit_driver.h"
#include "runtime/dip/dip_reward.h"
#include <boost/foreach.hpp>
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>
#include <ff/functionflow.h>

namespace neb {
namespace rt {
namespace dip {

dip_handler::dip_handler() : m_has_curr(false) {
  m_storage = neb::fs::fs_storage::instance().nbre_db_ptr();
}

void dip_handler::init_dip_params(block_height_t height) {

  if (!m_has_curr && m_incoming.empty()) {
    auto dip_versions_ptr = neb::fs::ir_api::get_ir_versions("dip", m_storage);
    if (dip_versions_ptr->empty()) {
      return;
    }
    LOG(INFO) << "dip versions not empty, size " << dip_versions_ptr->size();

    std::reverse(dip_versions_ptr->begin(), dip_versions_ptr->end());
    for (auto version : *dip_versions_ptr) {
      auto nbre_ir_ptr =
          neb::core::ir_warden::instance().get_ir_by_name_version("dip",
                                                                  version);
      block_height_t available_height = nbre_ir_ptr->height();
      m_incoming.push(std::make_pair(version, available_height));
    }
  }

  std::pair<version_t, block_height_t> tmp;

  while (!m_incoming.empty() && m_incoming.front().second <= height) {
    tmp = m_incoming.front();
    m_curr = tmp;
    m_has_curr = true;
    m_incoming.pop();
  }

  if (tmp.first && tmp.second) {
    try {
      jit_driver &jd = jit_driver::instance();
      LOG(INFO) << "to init dip params";
      jd.run_ir<std::string>("dip", std::numeric_limits<uint64_t>::max(),
                             neb::configuration::instance().dip_func_name(), 0);

      m_dip_history.push_back(dip_meta_t{
          neb::configuration::instance().dip_start_block(),
          neb::configuration::instance().dip_block_interval(), tmp.first});

      LOG(INFO) << "show dip history";
      for (auto &ele : m_dip_history) {
        LOG(INFO) << ele.m_start_block << ',' << ele.m_block_interval << ','
                  << ele.m_version;
      }
    } catch (const std::exception &e) {
      LOG(INFO) << "dip params init failed " << e.what();
    }
  }
}

void dip_handler::deploy(version_t version, block_height_t available_height) {
  m_incoming.push(std::make_pair(version, available_height));
}

void dip_handler::start(neb::block_height_t height,
                        const dip_meta_t *dip_meta) {
  init_dip_params(height);
  if (!m_has_curr) {
    LOG(INFO) << "dip params not init";
    return;
  }
  LOG(INFO) << "dip params init done";

  // get start block and block interval if default
  auto last_ele = m_dip_history.back();
  block_height_t dip_start_block = last_ele.m_start_block;
  block_height_t dip_block_interval = last_ele.m_block_interval;

  if (dip_meta) {
    LOG(INFO) << "dip meta not null";
    dip_start_block = dip_meta->m_start_block;
    dip_block_interval = dip_meta->m_block_interval;
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
  LOG(INFO) << "to start calculate dip reward for hash_height " << hash_height;

  if (m_dip_reward.find(hash_height) != m_dip_reward.end()) {
    LOG(INFO) << "dip reward already exists";
    return;
  }

  // get dip version if default
  std::string dip_name = "dip";
  auto dip_versions_ptr = neb::fs::ir_api::get_ir_versions(dip_name, m_storage);
  uint64_t dip_version = *dip_versions_ptr->begin();
  if (dip_meta) {
    dip_version = dip_meta->m_version;
  }

  ff::para<> p;
  p([this, dip_name, dip_version, hash_height]() {
    std::unique_lock<std::mutex> _l(m_sync_mutex);
    try {
      LOG(INFO) << "ff para run before lock";

      std::stringstream ss;
      ss << dip_name << dip_version;
      std::string name_version = ss.str();
      LOG(INFO) << "dip name version " << name_version;

      auto irs_ptr = neb::core::ir_warden::instance().get_ir_by_name_height(
          dip_name, hash_height);
      LOG(INFO) << "dip ir and depends size " << irs_ptr->size();

      jit_driver &jd = jit_driver::instance();
      LOG(INFO) << "jit driver run with " << name_version << ','
                << irs_ptr->size() << ','
                << neb::configuration::instance().dip_func_name() << ','
                << hash_height;
      auto dip_reward = jd.run<std::string>(
          name_version, *irs_ptr,
          neb::configuration::instance().dip_func_name(), hash_height);
      LOG(INFO) << "dip reward returned";

      write_dip_reward_to_storage(hash_height, dip_reward);
      LOG(INFO) << "write dip reward to storage";

    } catch (const std::exception &e) {
      LOG(INFO) << "jit driver execute dip failed " << e.what();
    }
  });
}

std::string
dip_handler::get_dip_reward_when_missing(neb::block_height_t height,
                                         const dip_meta_t &dip_meta) {

  LOG(INFO) << "call func get_dip_reward_when_missing";
  auto first_ele = m_dip_history.front();
  if (height < first_ele.m_start_block + first_ele.m_block_interval) {
    auto ret =
        boost::str(boost::format("{\"err\":\"available height is %1%\"}") %
                   (first_ele.m_start_block + first_ele.m_block_interval));
    LOG(INFO) << ret;
    return ret;
  }

  block_height_t dip_start_block = dip_meta.m_start_block;
  block_height_t dip_block_interval = dip_meta.m_block_interval;

  uint64_t interval_nums = (height - dip_start_block) / dip_block_interval;
  uint64_t hash_height = dip_start_block + dip_block_interval * interval_nums;

  LOG(INFO) << "start_block " << dip_start_block << ", block interval "
            << dip_block_interval << ", interval_nums " << interval_nums
            << ", hash_height " << hash_height;

  start(hash_height, &dip_meta);

  auto ret = std::string("{\"err\":\"dip reward missing, wait to restart\"}");
  LOG(INFO) << ret;
  return ret;
}

std::string dip_handler::get_dip_reward(neb::block_height_t height) {
  LOG(INFO) << "before lock";
  std::unique_lock<std::mutex> _l(m_sync_mutex);
  LOG(INFO) << "call func get_dip_reward";

  if (!m_has_curr) {
    auto ret = std::string("{\"err\":\"dip params not init yet\"}");
    LOG(INFO) << ret;
    return ret;
  }

  LOG(INFO) << "dip reward fist hash_height " << m_dip_reward.begin()->first;
  if (!m_dip_reward.empty() && height < m_dip_reward.begin()->first) {
    auto ret =
        boost::str(boost::format("{\"err\":\"available height is %1%\"}") %
                   (m_dip_reward.begin()->first));
    LOG(INFO) << ret;
    return ret;
  }

  LOG(INFO) << "dip history size " << m_dip_history.size();
  auto it_height = std::upper_bound(
      m_dip_history.begin(), m_dip_history.end(), dip_meta_t{height, 0, 0},
      [](const dip_meta_t &d1, const dip_meta_t &d2) {
        return d1.m_start_block < d2.m_start_block;
      });

  it_height--;
  block_height_t dip_start_block = it_height->m_start_block;
  block_height_t dip_block_interval = it_height->m_block_interval;
  LOG(INFO) << "find dip history start block " << dip_start_block
            << " , block interval " << dip_block_interval;

  uint64_t interval_nums = (height - dip_start_block) / dip_block_interval;
  uint64_t hash_height = dip_start_block + dip_block_interval * interval_nums;
  LOG(INFO) << "mapping height " << height << " to hash_height " << hash_height;

  auto ret = m_dip_reward.find(hash_height);
  if (ret == m_dip_reward.end()) {
    LOG(INFO) << "dip reward not exists";
    auto last_block = m_dip_history.back();
    if (hash_height - last_block.m_start_block >= last_block.m_block_interval) {
      auto ret = std::string("{\"err\":\"dip this interval not found\"}");
      LOG(INFO) << ret;
      return ret;
    }
    return get_dip_reward_when_missing(hash_height, *it_height);
  }
  LOG(INFO) << "dip reward exists";
  LOG(INFO) << ret->second;
  return ret->second;
}

void dip_handler::write_dip_reward_to_storage(neb::block_height_t hash_height,
                                              const std::string &dip_reward) {
  auto update_to_storage = [](const std::string &key,
                              const boost::property_tree::ptree &val_pt,
                              neb::fs::rocksdb_storage *rs) {
    std::stringstream ss;
    boost::property_tree::json_parser::write_json(ss, val_pt, false);
    rs->put(key, neb::util::string_to_byte(ss.str()));
  };

  LOG(INFO) << "call func write_dip_reward_from_storage";
  neb::util::bytes dip_rewards_bytes;
  std::string key = "dip_rewards";
  try {
    dip_rewards_bytes = m_storage->get(key);
  } catch (const std::exception &e) {
    LOG(INFO) << "dip reward empty " << e.what();

    boost::property_tree::ptree ele, arr, root;
    ele.put("", dip_reward);
    arr.push_back(std::make_pair("", ele));
    root.add_child(key, arr);
    update_to_storage(key, root, m_storage);
    m_dip_reward.insert(std::make_pair(hash_height, dip_reward));
    LOG(INFO) << "insert dip reward pair height " << hash_height
              << ", dip_reward " << dip_reward;
    return;
  }

  LOG(INFO) << "dip reward not empty";
  boost::property_tree::ptree root;
  std::stringstream ss(neb::util::byte_to_string(dip_rewards_bytes));
  boost::property_tree::json_parser::read_json(ss, root);

  boost::property_tree::ptree &arr = root.get_child(key);
  boost::property_tree::ptree ele;
  ele.put("", dip_reward);
  arr.push_back(std::make_pair("", ele));
  LOG(INFO) << "insert dip_reward";
  update_to_storage(key, root, m_storage);
  m_dip_reward.insert(std::make_pair(hash_height, dip_reward));
  LOG(INFO) << "insert dip reward pair height " << hash_height
            << ", dip_reward " << dip_reward;
}

void dip_handler::read_dip_reward_from_storage() {

  LOG(INFO) << "call func read_dip_reward_from_storage";
  std::string key = "dip_rewards";
  neb::util::bytes dip_rewards_bytes;
  try {
    dip_rewards_bytes = m_storage->get(key);
  } catch (const std::exception &e) {
    LOG(INFO) << "dip reward empty " << e.what();
    return;
  }

  LOG(INFO) << "dip reward not empty";
  boost::property_tree::ptree root;
  std::stringstream ss(neb::util::byte_to_string(dip_rewards_bytes));
  boost::property_tree::json_parser::read_json(ss, root);

  BOOST_FOREACH (boost::property_tree::ptree::value_type &v,
                 root.get_child(key)) {
    boost::property_tree::ptree pt = v.second;
    std::string dip_reward = pt.get<std::string>(std::string());

    boost::property_tree::ptree reward_pt;
    std::stringstream ss(dip_reward);
    boost::property_tree::json_parser::read_json(ss, reward_pt);
    block_height_t end_height = reward_pt.get<block_height_t>("end_height");
    m_dip_reward.insert(std::make_pair(end_height + 1, dip_reward));
    LOG(INFO) << "insert dip reward pair height " << end_height + 1
              << ", dip_reward " << dip_reward;
  }
}

} // namespace dip
} // namespace rt
} // namespace neb
