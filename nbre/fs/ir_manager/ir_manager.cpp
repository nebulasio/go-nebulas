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
#include "common/configuration.h"
#include "common/util/byte.h"
#include "common/util/version.h"
#include "core/neb_ipc/server/ipc_configuration.h"
#include "fs/ir_manager/ir_manager_helper.h"
#include "jit/jit_driver.h"
#include "runtime/dip/dip_handler.h"
#include "runtime/version.h"
#include <boost/format.hpp>

namespace neb {
namespace fs {

ir_manager::ir_manager(const std::string &path, const std::string &bc_path) {
  m_storage = std::make_unique<rocksdb_storage>();
  m_storage->open_database(path, storage_open_for_readwrite);

  m_blockchain =
      std::make_unique<blockchain>(bc_path, storage_open_for_readonly);
}

ir_manager::~ir_manager() {
  if (m_storage) {
    m_storage->close_database();
  }
}

std::unique_ptr<nbre::NBREIR> ir_manager::read_ir(const std::string &name,
                                                  uint64_t version) {
  std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
  std::stringstream ss;
  ss << name << version;

  neb::util::bytes nbre_bytes = m_storage->get(ss.str());
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

  std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
  neb::util::bytes bytes_versions;
  try {
    bytes_versions = m_storage->get(name);
  } catch (const std::exception &e) {
    LOG(INFO) << "get ir " << name << " failed " << e.what();
    return std::make_unique<std::vector<nbre::NBREIR>>(irs);
  }

  std::unordered_set<std::string> ir_set;
  size_t gap = sizeof(uint64_t) / sizeof(uint8_t);

  for (size_t i = 0; i < bytes_versions.size(); i += gap) {
    byte_t *bytes_version =
        bytes_versions.value() + (bytes_versions.size() - gap - i);

    if (bytes_version != nullptr) {
      uint64_t version =
          neb::util::byte_to_number<uint64_t>(bytes_version, gap);
      read_ir_depends(name, version, height, depends, ir_set, irs);
      if (!irs.empty()) {
        break;
      }
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

  block_height_t start_height =
      ir_manager_helper::nbre_block_height(m_storage.get());
  block_height_t end_height =
      ir_manager_helper::lib_block_height(m_blockchain.get());

  ir_manager_helper::load_auth_table(m_storage.get(), m_auth_table);
  neb::rt::dip::dip_handler::instance().start(start_height, end_height);

  std::string failed_flag =
      neb::configuration::instance().nbre_failed_flag_name();

  //! TODO: we may consider parallel here!
  for (block_height_t h = start_height + 1; h <= end_height; h++) {
    // LOG(INFO) << h;

    if (!ir_manager_helper::has_failed_flag(m_storage.get(), failed_flag)) {
      ir_manager_helper::set_failed_flag(m_storage.get(), failed_flag);
      parse_irs_by_height(h);
    }
    m_storage->put(
        std::string(neb::configuration::instance().nbre_max_height_name(),
                    std::allocator<char>()),
        neb::util::number_to_byte<neb::util::bytes>(h));
    ir_manager_helper::del_failed_flag(m_storage.get(), failed_flag);
  }
}

void ir_manager::deploy_auth_table(nbre::NBREIR &nbre_ir,
                                   const neb::util::bytes payload_bytes) {

  // TODO expect auth table exceed 128k bytes size
  LOG(INFO) << "before set auth table by jit, auth table size: "
            << m_auth_table.size();
  ir_manager_helper::run_auth_table(nbre_ir, m_auth_table);
  m_storage->put(neb::configuration::instance().nbre_auth_table_name(),
                 payload_bytes);
  LOG(INFO) << "updating auth table...";
  LOG(INFO) << "after set auth table by jit, auth table size: "
            << m_auth_table.size();
}

void ir_manager::show_auth_table() {

  LOG(INFO) << "\nshow auth table";
  for (auto &r : m_auth_table) {
    std::string key = boost::str(boost::format("key <%1%, %2%, %3%>, ") %
                                 std::get<0>(r.first) % std::get<1>(r.first) %
                                 std::get<2>(r.first));
    std::string val = boost::str(boost::format("val <%1%, %2%>") %
                                 std::get<0>(r.second) % std::get<1>(r.second));
    LOG(INFO) << key << val;
  }
}

void ir_manager::parse_irs_by_height(block_height_t height) {
  auto block = m_blockchain->load_block_with_height(height);

  for (auto &tx : block->transactions()) {
    auto &data = tx.data();
    const std::string &type = data.type();
    std::string from = tx.from();

    // ignore transaction other than transaction `protocol`
    std::string ir_tx_type =
        neb::configuration::instance().ir_tx_payload_type();
    if (type != ir_tx_type) {
      continue;
    }

    const std::string &payload = data.payload();
    neb::util::bytes payload_bytes = neb::util::string_to_byte(payload);
    std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
    bool ret =
        nbre_ir->ParseFromArray(payload_bytes.value(), payload_bytes.size());
    if (!ret) {
      throw std::runtime_error("parse transaction payload failed");
    }

    const std::string &name = nbre_ir->name();
    uint64_t version = nbre_ir->version();

    // deploy auth table
    if (neb::configuration::instance().auth_module_name() == name &&
        neb::core::ipc_configuration::instance().admin_pub_addr() == from) {
      deploy_auth_table(*nbre_ir.get(), payload_bytes);
      continue;
    }

    auto it = m_auth_table.find(std::make_tuple(name, version, from));
    // ir not in auth table
    if (it == m_auth_table.end()) {
      LOG(INFO) << boost::str(
          boost::format("tuple <%1%, %2%, %3%> not in auth table") % name %
          version % from);
      show_auth_table();
      continue;
    }
    const uint64_t height = nbre_ir->height();
    // ir in auth table but already invalid
    if (height < std::get<0>(it->second) || height >= std::get<1>(it->second)) {
      LOG(INFO) << "ir already becomes invalid";
      continue;
    }

    // deploy ir
    try {
      neb::util::bytes bytes_versions = m_storage->get(name);
      bytes_versions.append_bytes(
          neb::util::number_to_byte<neb::util::bytes>(version));
      m_storage->put(name, bytes_versions);
    } catch (const std::exception &e) {
      LOG(INFO) << "no such ir, start to deploy the first one";
      m_storage->put(name,
                     neb::util::number_to_byte<neb::util::bytes>(version));
    }

    std::stringstream ss;
    ss << name << version;
    m_storage->put(ss.str(), payload_bytes);
    LOG(INFO) << "deploy " << name << " version " << version
              << " successfully!";

    // update ir list
    std::string const_str_nbre_ir_list =
        neb::configuration::instance().nbre_ir_list_name();
    neb::util::bytes bytes_ir_list_json;
    try {
      bytes_ir_list_json = m_storage->get(const_str_nbre_ir_list);
    } catch (const std::exception &e) {
      LOG(INFO) << const_str_nbre_ir_list << " not in storage " << e.what();
    }
    ir_manager_helper::update_ir_list(
        neb::util::byte_to_string(bytes_ir_list_json), name, m_storage.get());
  }
}

} // namespace fs
} // namespace neb
