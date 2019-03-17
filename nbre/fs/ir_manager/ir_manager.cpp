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

#include "fs/ir_manager/ir_manager.h"
#include "common/byte.h"
#include "common/common.h"
#include "common/configuration.h"
#include "common/version.h"
#include "fs/bc_storage_session.h"
#include "fs/ir_manager/api/ir_api.h"
#include "fs/ir_manager/ir_manager_helper.h"
#include "fs/storage_holder.h"
#include "jit/cpp_ir.h"
#include "jit/jit_driver.h"
#include "runtime/dip/dip_handler.h"
#include "runtime/version.h"
#include "util/json_parser.h"
#include <boost/format.hpp>
#include <ff/functionflow.h>

namespace neb {
namespace fs {

ir_manager::ir_manager() {
  m_storage = storage_holder::instance().nbre_db_ptr();
}

ir_manager::~ir_manager() { m_storage->close_database(); }

std::unique_ptr<nbre::NBREIR> ir_manager::read_ir(const std::string &name,
                                                  version_t version) {
  return ir_api::get_ir(name, version, m_storage).second;
}

std::unique_ptr<std::vector<nbre::NBREIR>>
ir_manager::read_irs(const std::string &name, block_height_t height,
                     bool depends) {
  auto irs = std::make_unique<std::vector<nbre::NBREIR>>();

  neb::bytes bytes_versions;
  try {
    bytes_versions = m_storage->get(name);
  } catch (const std::exception &e) {
    LOG(INFO) << "get ir " << name << " failed " << e.what();
    return irs;
  }

  auto versions_ptr = ir_api::get_ir_versions(name, m_storage);
  for (auto version : *versions_ptr) {
    read_ir_depends(name, version, height, depends, *irs);
    if (!irs->empty()) {
      break;
    }
  }

  return irs;
}

void ir_manager::read_ir_depends(const std::string &name, version_t version,
                                 block_height_t height, bool depends,
                                 std::vector<nbre::NBREIR> &irs) {

  if (name == neb::configuration::instance().rt_module_name() &&
      neb::rt::get_version() < neb::version(version)) {
    throw std::runtime_error("need to update nbre runtime version");
  }

  std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
  std::unordered_set<std::string> ir_set;
  std::queue<std::pair<std::string, version_t>> q;
  q.push(std::make_pair(name, version));

  while (!q.empty()) {
    auto &ele = q.front();
    q.pop();
    std::stringstream ss;
    ss << ele.first << ele.second;

    if (ir_set.find(ss.str()) == ir_set.end()) {
      ir_set.insert(ss.str());
      neb::bytes nbre_bytes;
      try {
        nbre_bytes = m_storage->get(ss.str());
      } catch (const std::exception &e) {
        LOG(INFO) << "get ir " << ele.first << " failed " << e.what();
        return;
      }

      bool ret = nbre_ir->ParseFromArray(nbre_bytes.value(), nbre_bytes.size());
      if (!ret) {
        throw std::runtime_error("parse nbre failed");
      }

      if (nbre_ir->height() <= height) {
        if (depends) {
          for (auto &deps : nbre_ir->depends()) {
            q.push(std::make_pair(deps.name(), deps.version()));
          }
        }
        irs.push_back(*nbre_ir);
      }
    }
  }
}

void ir_manager::parse_irs(
    util::wakeable_queue<std::shared_ptr<nbre_ir_transactions_req>> &q_txs) {

  ir_manager_helper::load_auth_table(m_storage, m_auth_table);
  block_height_t last_height = ir_manager_helper::nbre_block_height(m_storage);

  while (!q_txs.empty()) {
    auto ele = q_txs.try_pop_front();
    if (!ele.first) {
      break;
    }

    auto h = ele.second->get<p_height>();
    if (h < last_height + 1) {
      continue;
    }
    if (h > last_height + 1) {
      parse_when_missing_block(last_height + 1, h);
    }

    const auto &txs = ele.second->get<p_ir_transactions>();
    parse_next_block(h, txs);
    last_height = h;
  }
}

void ir_manager::parse_when_missing_block(block_height_t start_height,
                                          block_height_t end_height) {
  std::string ir_tx_type = neb::configuration::instance().ir_tx_payload_type();

  for (block_height_t h = start_height; h < end_height; h++) {
    auto block = blockchain::load_block_with_height(h);
    std::vector<corepb::Transaction> txs;

    for (auto &tx : block->transactions()) {
      auto &data = tx.data();
      const std::string &type = data.type();
      if (type == ir_tx_type) {
        txs.push_back(tx);
      }
    }
    parse_with_height(h, txs);
  }
}

void ir_manager::parse_next_block(block_height_t height,
                                  const std::vector<std::string> &txs_seri) {
  std::vector<corepb::Transaction> txs;
  if (txs_seri.empty()) {
    parse_with_height(height, txs);
    return;
  }

  for (auto &tx_seri : txs_seri) {
    auto tx_bytes = string_to_byte(tx_seri);
    std::unique_ptr<corepb::Transaction> tx =
        std::make_unique<corepb::Transaction>();
    bool ret = tx->ParseFromArray(tx_bytes.value(), tx_bytes.size());
    if (!ret) {
      throw std::runtime_error("parse transaction failed");
    }
    txs.push_back(*tx);
  }
  parse_with_height(height, txs);
}

void ir_manager::parse_with_height(
    block_height_t height, const std::vector<corepb::Transaction> &txs) {

  std::string failed_flag =
      neb::configuration::instance().nbre_failed_flag_name();

  if (!ir_manager_helper::has_failed_flag(m_storage, failed_flag)) {
    ir_manager_helper::set_failed_flag(m_storage, failed_flag);
    parse_irs_by_height(height, txs);
  }
  m_storage->put(
      std::string(neb::configuration::instance().nbre_max_height_name(),
                  std::allocator<char>()),
      neb::number_to_byte<neb::bytes>(height));
  ir_manager_helper::del_failed_flag(m_storage, failed_flag);

  neb::rt::dip::dip_handler::instance().start(height);
}

void ir_manager::parse_irs_by_height(
    block_height_t height, const std::vector<corepb::Transaction> &txs) {

  for (auto &tx : txs) {
    auto &data = tx.data();
    const std::string &type = data.type();

    // ignore transaction other than transaction `protocol`
    std::string ir_tx_type =
        neb::configuration::instance().ir_tx_payload_type();
    if (type != ir_tx_type) {
      LOG(INFO) << "ignore ir with type: " << type;
      continue;
    }

    boost::property_tree::ptree pt;
    neb::util::json_parser::read_json(data.payload(), pt);
    neb::bytes payload_bytes =
        neb::bytes::from_base64(pt.get<std::string>("Data"));

    std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
    bool ret =
        nbre_ir->ParseFromArray(payload_bytes.value(), payload_bytes.size());
    if (!ret) {
      LOG(ERROR) << "parse transaction payload failed " << type;
      throw std::runtime_error("parse transaction payload failed");
    }

    const std::string &name = nbre_ir->name();
    version_t version = nbre_ir->version();
    if (ir_api::ir_exist(name, version, m_storage)) {
      LOG(INFO) << "ignore ir name: " << name << " with version: " << version;
      continue;
    }

    address_t from = to_address(tx.from());
    // deploy auth table
    if (neb::configuration::instance().auth_module_name() == name &&
        neb::configuration::instance().admin_pub_addr() == from) {
      ir_manager_helper::compile_payload_code(nbre_ir.get(), payload_bytes);
      ir_manager_helper::deploy_auth_table(m_storage, *nbre_ir.get(),
                                           m_auth_table, payload_bytes);
      continue;
    }

    auto it = m_auth_table.find(std::make_tuple(name, from));
    // ir not in auth table
    if (it == m_auth_table.end()) {
      LOG(INFO) << boost::str(
          boost::format("tuple <%1%, %2%> not in auth table") % name %
          std::to_string(from));
      ir_manager_helper::show_auth_table(m_auth_table);
      continue;
    }
    block_height_t ht = nbre_ir->height();
    // ir in auth table but already invalid
    if (ht < std::get<0>(it->second) || ht >= std::get<1>(it->second)) {
      LOG(INFO) << "ir already becomes invalid";
      continue;
    }
    ir_manager_helper::compile_payload_code(nbre_ir.get(), payload_bytes);

    // update ir list and versions
    ir_manager_helper::update_ir_list(name, m_storage);
    ir_manager_helper::update_ir_versions(name, version, m_storage);

    // deploy ir
    ir_manager_helper::deploy_ir(name, version, payload_bytes, m_storage);

    deploy_if_dip(name, version, ht);
  }
}

void ir_manager::deploy_if_dip(const std::string &name, version_t version,
                               block_height_t available_height) {
  if (name != std::string("dip")) {
    return;
  }
  neb::rt::dip::dip_handler::instance().deploy(version, available_height);
}

} // namespace fs
} // namespace neb
