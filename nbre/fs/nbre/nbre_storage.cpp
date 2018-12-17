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

#include "fs/nbre/nbre_storage.h"
#include "common/configuration.h"
#include "common/util/byte.h"
#include "common/util/version.h"
#include "fs/nbre/flag_storage.h"
#include "jit/jit_driver.h"
#include "runtime/dip/dip_handler.h"
#include "runtime/version.h"
#include <boost/foreach.hpp>
#include <boost/property_tree/json_parser.hpp>

namespace neb {
namespace fs {

nbre_storage::nbre_storage(const std::string &path,
                           const std::string &bc_path) {
  m_storage = std::make_unique<rocksdb_storage>();
  m_storage->open_database(path, storage_open_for_readwrite);

  m_blockchain =
      std::make_unique<blockchain>(bc_path, storage_open_for_readonly);
}

nbre_storage::~nbre_storage() {
  if (m_storage) {
    m_storage->close_database();
  }
}

std::vector<std::unique_ptr<nbre::NBREIR>>
nbre_storage::read_nbre_by_height(const std::string &name,
                                  block_height_t height, bool depends_trace) {

  std::vector<std::unique_ptr<nbre::NBREIR>> ret;
  std::unordered_set<std::string> pkgs;

  std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
  neb::util::bytes bytes_versions = m_storage->get(name);

  size_t gap = sizeof(uint64_t) / sizeof(uint8_t);
  for (size_t i = 0; i < bytes_versions.size(); i += gap) {
    byte_t *bytes_version =
        bytes_versions.value() + (bytes_versions.size() - gap - i);

    if (bytes_version != nullptr) {
      uint64_t version =
          neb::util::byte_to_number<uint64_t>(bytes_version, gap);
      read_nbre_depends_recursive(name, version, height, depends_trace, pkgs,
                                  ret);
      if (!ret.empty()) {
        break;
      }
    }
  }
  return ret;
}

void nbre_storage::read_nbre_depends_recursive(
    const std::string &name, uint64_t version, block_height_t height,
    bool depends_trace, std::unordered_set<std::string> &pkg,
    std::vector<std::unique_ptr<nbre::NBREIR>> &irs) {

  if (name == neb::configuration::instance().rt_module_name() &&
      neb::rt::get_version() < neb::util::version(version)) {
    throw std::runtime_error("nbre runtime pkg version is too old");
  }

  std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
  std::string name_version = name + std::to_string(version);
  if (pkg.find(name_version) != pkg.end()) {
    return;
  }

  neb::util::bytes nbre_bytes = m_storage->get(name_version);
  bool ret = nbre_ir->ParseFromArray(nbre_bytes.value(), nbre_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse nbre failed");
  }

  if (nbre_ir->height() <= height) {
    if (depends_trace) {
      for (auto &dep : nbre_ir->depends()) {
        read_nbre_depends_recursive(dep.name(), dep.version(), height,
                                    depends_trace, pkg, irs);
      }
    }
    irs.push_back(std::move(nbre_ir));
    pkg.insert(name_version);
  }
}

std::unique_ptr<nbre::NBREIR>
nbre_storage::read_nbre_by_name_version(const std::string &name,
                                        uint64_t version) {
  std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
  std::string name_version = name + std::to_string(version);

  neb::util::bytes nbre_bytes = m_storage->get(name_version);
  bool ret = nbre_ir->ParseFromArray(nbre_bytes.value(), nbre_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse nbre failed");
  }

  return nbre_ir;
}

void nbre_storage::write_nbre_until_sync() {

  std::chrono::system_clock::time_point start_time;
  std::chrono::system_clock::time_point end_time;
  std::chrono::seconds time_spend;

  do {
    start_time = std::chrono::system_clock::now();
    write_nbre();
    end_time = std::chrono::system_clock::now();
    time_spend =
        std::chrono::duration_cast<std::chrono::seconds>(end_time - start_time);
  } while (time_spend >
           std::chrono::seconds(
               neb::configuration::instance().ir_warden_time_interval()));
}

block_height_t nbre_storage::get_start_height() {

  block_height_t start_height = 0;
  try {
    start_height = neb::util::byte_to_number<block_height_t>(m_storage->get(
        std::string(neb::configuration::instance().nbre_max_height_name(),
                    std::allocator<char>())));
  } catch (std::exception &e) {
    m_storage->put(
        std::string(neb::configuration::instance().nbre_max_height_name(),
                    std::allocator<char>()),
        neb::util::number_to_byte<neb::util::bytes>(start_height));
  }
  return start_height;
}

block_height_t nbre_storage::get_end_height() {

  std::unique_ptr<corepb::Block> end_block = m_blockchain->load_LIB_block();
  block_height_t end_height = end_block->height();
  return end_height;
}

void nbre_storage::write_nbre() {

  block_height_t start_height = get_start_height();
  block_height_t end_height = get_end_height();
  LOG(INFO) << "start height " << start_height << ',' << "end height "
            << end_height;

  set_auth_table();
  neb::rt::dip::dip_handler::instance().start(start_height, end_height);

  flag_storage fs(m_storage.get());
  std::string failed_flag =
      neb::configuration::instance().nbre_failed_flag_name();

  //! TODO: we may consider parallel here!
  for (block_height_t h = start_height + 1; h <= end_height; h++) {
    LOG(INFO) << h;

    if (!fs.has_flag(failed_flag)) {
      fs.set_flag(failed_flag);
      write_nbre_by_height(h);
    }
    m_storage->put(
        std::string(neb::configuration::instance().nbre_max_height_name(),
                    std::allocator<char>()),
        neb::util::number_to_byte<neb::util::bytes>(h));
    fs.del_flag(failed_flag);
  }
}

void nbre_storage::write_nbre_by_height(block_height_t height) {

  auto block = m_blockchain->load_block_with_height(height);

  for (auto &tx : block->transactions()) {
    auto &data = tx.data();
    const std::string &type = data.type();
    std::string from = tx.from();

    if (type == neb::configuration::instance().ir_tx_payload_type()) {
      const std::string &payload = data.payload();
      neb::util::bytes payload_bytes = neb::util::string_to_byte(payload);

      std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
      bool ret =
          nbre_ir->ParseFromArray(payload_bytes.value(), payload_bytes.size());
      if (!ret) {
        throw std::runtime_error("parse transaction payload failed");
      }

      const std::string &name = nbre_ir->name();
      const uint64_t version = nbre_ir->version();
      std::string from_base58 = neb::util::string_to_byte(from).to_base58();
      std::string admin_base58 =
          neb::util::string_to_byte(
              neb::configuration::instance().admin_pub_addr())
              .to_base58();
      LOG(INFO) << "from address: " << from_base58
                << ", admin address: " << admin_base58;
      LOG(INFO) << "module name: " << name << ", auth module name: "
                << neb::configuration::instance().auth_module_name();

      if (neb::configuration::instance().auth_module_name() == name &&
          neb::configuration::instance().admin_pub_addr() == from) {
        // TODO expect auth table exceed 128k bytes size

        LOG(INFO) << "before set auth table by jit, auth table size: "
                  << m_auth_table.size();
        set_auth_table_by_jit(nbre_ir);
        m_storage->put(neb::configuration::instance().nbre_auth_table_name(),
                       payload_bytes);
        LOG(INFO) << "updating auth table...";
        LOG(INFO) << "after set auth table by jit, auth table size: "
                  << m_auth_table.size();
        continue;
      }

      auto it = m_auth_table.find(std::make_tuple(name, version, from));
      if (it == m_auth_table.end()) {
        LOG(INFO) << boost::str(
            boost::format("tuple <%1%, %2%, %3%> not in auth table") % name %
            version % from);
        LOG(INFO) << "\nshow auth table";
        for (auto &r : m_auth_table) {
          std::string key = boost::str(
              boost::format("key <%1%, %2%, %3%>, ") % std::get<0>(r.first) %
              std::get<1>(r.first) % std::get<2>(r.first));
          std::string val =
              boost::str(boost::format("val <%1%, %2%>") %
                         std::get<0>(r.second) % std::get<1>(r.second));
          LOG(INFO) << key << val;
        }
        continue;
      }
      const uint64_t height = nbre_ir->height();
      if (height < std::get<0>(it->second) ||
          height >= std::get<1>(it->second)) {
        LOG(INFO) << "address has no access to deploy code";
        continue;
      }

      try {
        neb::util::bytes bytes_versions = m_storage->get(name);
        bytes_versions.append_bytes(
            neb::util::number_to_byte<neb::util::bytes>(version));
        m_storage->put(name, bytes_versions);
      } catch (const std::exception &e) {
        m_storage->put(name,
                       neb::util::number_to_byte<neb::util::bytes>(version));
      }

      m_storage->put(name + std::to_string(version), payload_bytes);
      LOG(INFO) << "deploy " << name << " version " << version
                << " successfully!";

      std::string nbre_ir_list_name =
          neb::configuration::instance().nbre_ir_list_name();
      neb::util::bytes bytes_ir_name_list;
      try {
        bytes_ir_name_list = m_storage->get(nbre_ir_list_name);
      } catch (const std::exception &e) {
        LOG(INFO) << e.what();
      }
      update_ir_list(nbre_ir_list_name,
                     neb::util::byte_to_string(bytes_ir_name_list), name);

    }
  }
}

void nbre_storage::update_ir_list_to_db(const std::string &nbre_ir_list_name,
                                        const boost::property_tree::ptree &pt) {
  std::stringstream ss;
  boost::property_tree::json_parser::write_json(ss, pt);
  m_storage->put(nbre_ir_list_name, neb::util::string_to_byte(ss.str()));
}

void nbre_storage::update_ir_list(const std::string &nbre_ir_list_name,
                                  const std::string &ir_name_list,
                                  const std::string &ir_name) {

  std::string ir_list_name = neb::configuration::instance().ir_list_name();

  if (ir_name_list.empty()) {
    boost::property_tree::ptree ele, arr, root;
    ele.put("", ir_name);
    arr.push_back(std::make_pair("", ele));
    root.add_child(ir_list_name, root);
    update_ir_list_to_db(nbre_ir_list_name, root);
    return;
  }

  boost::property_tree::ptree root;
  std::stringstream ss(ir_name_list);
  boost::property_tree::json_parser::read_json(ss, root);

  BOOST_FOREACH (boost::property_tree::ptree::value_type &v,
                 root.get_child(ir_list_name)) {
    boost::property_tree::ptree pt = v.second;
    if (ir_name == pt.get<std::string>(std::string())) {
      return;
    }
  }

  boost::property_tree::ptree &arr = root.get_child(ir_list_name);
  boost::property_tree::ptree ele;
  ele.put("", ir_name);
  arr.push_back(std::make_pair("", ele));

  update_ir_list_to_db(nbre_ir_list_name, root);
}

void nbre_storage::set_auth_table() {

  if (!m_auth_table.empty()) {
    return;
  }

  std::unique_ptr<nbre::NBREIR> nbre_ir = std::make_unique<nbre::NBREIR>();
  try {
    auto payload_bytes =
        m_storage->get(neb::configuration::instance().nbre_auth_table_name());

    bool ret =
        nbre_ir->ParseFromArray(payload_bytes.value(), payload_bytes.size());
    if (!ret) {
      throw std::runtime_error("parse payload auth table failed");
    }
  } catch (const std::exception &e) {
    LOG(INFO) << e.what();
    return;
  }

  set_auth_table_by_jit(nbre_ir);
}

void nbre_storage::set_auth_table_by_jit(
    std::unique_ptr<nbre::NBREIR> &nbre_ir) {

  auth_table_t auth_table_raw;

  try {
    jit_driver &jd = jit_driver::instance();
    std::stringstream ss;
    ss << nbre_ir->name() << nbre_ir->version();
    LOG(INFO) << "set auth table by jit " << ss.str();

    std::vector<std::unique_ptr<nbre::NBREIR>> irs;
    irs.push_back(std::move(nbre_ir));

    auth_table_raw = jd.run<auth_table_t>(
        ss.str(), irs,
        neb::configuration::instance().auth_func_mangling_name());
    LOG(INFO) << "jit driver return size: " << auth_table_raw.size();

  } catch (const std::exception &e) {
    LOG(INFO) << e.what();
  }

  m_auth_table.clear();
  for (auto &r : auth_table_raw) {
    auth_key_t k =
        std::make_tuple(std::get<0>(r), std::get<1>(r), std::get<2>(r));
    auth_val_t v = std::make_tuple(std::get<3>(r), std::get<4>(r));
    m_auth_table.insert(std::make_pair(k, v));
  }
}

} // namespace fs
} // namespace neb
