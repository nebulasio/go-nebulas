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
#include "common/common.h"
#include "common/configuration.h"
#include "common/util/byte.h"
#include "common/util/json_parser.h"
#include "common/util/version.h"
#include "core/ipc_configuration.h"
#include "fs/ir_manager/api/ir_api.h"
#include "fs/ir_manager/ir_manager_helper.h"
#include "fs/storage_holder.h"
#include "jit/cpp_ir.h"
#include "jit/jit_driver.h"
#include "runtime/dip/dip_handler.h"
#include "runtime/version.h"
#include <boost/format.hpp>
#include <ff/functionflow.h>

namespace neb {
namespace fs {

ir_manager::ir_manager() {
  m_storage = storage_holder::instance().nbre_db_ptr();
  m_blockchain = storage_holder::instance().neb_db_ptr();
}

ir_manager::~ir_manager() { m_storage->close_database(); }

std::unique_ptr<nbre::NBREIR> ir_manager::read_ir(const std::string &name,
                                                  uint64_t version) {
  std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
  std::stringstream ss;
  ss << name << version;

  neb::util::bytes nbre_bytes;
  try {
    nbre_bytes = m_storage->get(ss.str());
  } catch (const std::exception &e) {
    LOG(INFO) << "no such ir named " << name << " with version " << version
              << ' ' << e.what();
  }
  bool ret = nbre_ir->ParseFromArray(nbre_bytes.value(), nbre_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse nbre failed");
  }

  return nbre_ir;
}

std::unique_ptr<std::vector<nbre::NBREIR>>
ir_manager::read_irs(const std::string &name, block_height_t height,
                     bool depends) {
  std::vector<nbre::NBREIR> irs;

  neb::util::bytes bytes_versions;
  try {
    bytes_versions = m_storage->get(name);
  } catch (const std::exception &e) {
    LOG(INFO) << "get ir " << name << " failed " << e.what();
    return std::make_unique<std::vector<nbre::NBREIR>>(irs);
  }

  std::unordered_set<std::string> ir_set;
  auto versions_ptr = ir_api::get_ir_versions(name, m_storage);
  for (auto version : *versions_ptr) {
    read_ir_depends(name, version, height, depends, ir_set, irs);
    if (!irs.empty()) {
      break;
    }
  }

  return std::make_unique<std::vector<nbre::NBREIR>>(irs);
}

void ir_manager::read_ir_depends(const std::string &name, uint64_t version,
                                 block_height_t height, bool depends,
                                 std::unordered_set<std::string> &ir_set,
                                 std::vector<nbre::NBREIR> &irs) {

  if (name == neb::configuration::instance().rt_module_name() &&
      neb::rt::get_version() < neb::util::version(version)) {
    throw std::runtime_error("need to update nbre runtime version");
  }

  std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
  std::stringstream ss;
  ss << name << version;
  // ir already exists
  if (ir_set.find(ss.str()) != ir_set.end()) {
    return;
  }

  neb::util::bytes nbre_bytes;
  try {
    nbre_bytes = m_storage->get(ss.str());
  } catch (const std::exception &e) {
    LOG(INFO) << "get ir " << name << " failed " << e.what();
    return;
  }

  bool ret = nbre_ir->ParseFromArray(nbre_bytes.value(), nbre_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse nbre failed");
  }

  if (nbre_ir->height() <= height) {
    if (depends) {
      for (auto &dep : nbre_ir->depends()) {
        read_ir_depends(dep.name(), dep.version(), height, depends, ir_set,
                        irs);
      }
    }
    irs.push_back(*nbre_ir);
    ir_set.insert(ss.str());
  }
  return;
}

void ir_manager::parse_irs_till_latest() {

  std::chrono::system_clock::time_point start_time =
      std::chrono::system_clock::now();
  std::chrono::system_clock::time_point end_time;
  std::chrono::seconds time_spend;
  std::chrono::seconds time_interval = std::chrono::seconds(
      neb::configuration::instance().ir_warden_time_interval());

  do {
    parse_irs();
    end_time = std::chrono::system_clock::now();
    time_spend =
        std::chrono::duration_cast<std::chrono::seconds>(end_time - start_time);
    start_time = end_time;
  } while (time_spend > time_interval);
}

void ir_manager::parse_irs() {

  block_height_t start_height = ir_manager_helper::nbre_block_height(m_storage);
  block_height_t end_height = ir_manager_helper::lib_block_height(m_blockchain);

  ir_manager_helper::load_auth_table(m_storage, m_auth_table);

  std::string failed_flag =
      neb::configuration::instance().nbre_failed_flag_name();

  //! TODO: we may consider parallel here!
  for (block_height_t h = start_height + 1; h <= end_height; h++) {
    // LOG(INFO) << h;

    if (!ir_manager_helper::has_failed_flag(m_storage, failed_flag)) {
      ir_manager_helper::set_failed_flag(m_storage, failed_flag);
      parse_irs_by_height(h);
    }
    m_storage->put(
        std::string(neb::configuration::instance().nbre_max_height_name(),
                    std::allocator<char>()),
        neb::util::number_to_byte<neb::util::bytes>(h));
    ir_manager_helper::del_failed_flag(m_storage, failed_flag);

    neb::rt::dip::dip_handler::instance().start(h);
  }
}

void ir_manager::parse_irs_by_height(block_height_t height) {
  auto block = m_blockchain->load_block_with_height(height);

  for (auto &tx : block->transactions()) {
    auto &data = tx.data();
    const std::string &type = data.type();

    // ignore transaction other than transaction `protocol`
    std::string ir_tx_type =
        neb::configuration::instance().ir_tx_payload_type();
    if (type != ir_tx_type) {
      continue;
    }
    LOG(INFO) << height;

    boost::property_tree::ptree pt;
    neb::util::json_parser::read_json(data.payload(), pt);
    neb::util::bytes payload_bytes =
        neb::util::bytes::from_base64(pt.get<std::string>("Data"));


    std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
    bool ret =
        nbre_ir->ParseFromArray(payload_bytes.value(), payload_bytes.size());
    if (!ret) {
      throw std::runtime_error("parse transaction payload failed");
    }

    if (nbre_ir->ir_type() == ::neb::ir_type::cpp) {
      //! We need compile the code
      cpp::cpp_ir ci(nbre_ir->ir());
      neb::util::bytes ir = ci.llvm_ir_content();
      nbre_ir->set_ir(neb::util::byte_to_string(ir));

      auto bytes_long = nbre_ir->ByteSizeLong();
      payload_bytes = neb::util::bytes(bytes_long);
      nbre_ir->SerializeToArray((void *)payload_bytes.value(),
                                payload_bytes.size());
    }

    const std::string &from = tx.from();
    const std::string &name = nbre_ir->name();

    // deploy auth table
    if (neb::configuration::instance().auth_module_name() == name &&
        neb::core::ipc_configuration::instance().admin_pub_addr() == from) {
        ir_manager_helper::deploy_auth_table(m_storage, *nbre_ir.get(),
                                             m_auth_table, payload_bytes);
      continue;
    }

    uint64_t version = nbre_ir->version();
    auto it = m_auth_table.find(std::make_tuple(name, version, from));
    // ir not in auth table
    if (it == m_auth_table.end()) {
      LOG(INFO) << boost::str(
          boost::format("tuple <%1%, %2%, %3%> not in auth table") % name %
          version % from);
      ir_manager_helper::show_auth_table(m_auth_table);
      continue;
    }
    const uint64_t ht = nbre_ir->height();
    // ir in auth table but already invalid
    if (ht < std::get<0>(it->second) || ht >= std::get<1>(it->second)) {
      LOG(INFO) << "ir already becomes invalid";
      continue;
    }

    // update ir list and versions
    ir_manager_helper::update_ir_list(name, m_storage);
    ir_manager_helper::update_ir_versions(name, version, m_storage);

    // deploy ir
    ir_manager_helper::deploy_ir(name, version, payload_bytes, m_storage);

    deploy_if_dip(name, version, ht);
  }
}

void ir_manager::deploy_if_dip(const std::string &name, uint64_t version,
                               block_height_t available_height) {
  if (name != std::string("dip")) {
    return;
  }
  neb::rt::dip::dip_handler::instance().deploy(version, available_height);
}

} // namespace fs
} // namespace neb
