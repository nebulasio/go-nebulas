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
#include "fs/blockchain.h"
#include "fs/proto/ir.pb.h"
#include <boost/property_tree/ptree.hpp>

namespace neb {
namespace fs {

typedef std::tuple<module_t, version_t, address_t> auth_key_t;
typedef std::tuple<start_block_t, end_block_t> auth_val_t;

class nbre_storage {
public:
  nbre_storage(const std::string &path, const std::string &bc_path);
  ~nbre_storage();
  nbre_storage(const nbre_storage &ns) = delete;
  nbre_storage &operator=(const nbre_storage &ns) = delete;

  std::vector<std::unique_ptr<nbre::NBREIR>>
  read_nbre_by_height(const std::string &name, block_height_t height,
                      bool depends_trace);

  std::unique_ptr<nbre::NBREIR>
  read_nbre_by_name_version(const std::string &name, uint64_t version);

  void write_nbre_until_sync();

private:
  void
  read_nbre_depends_recursive(const std::string &name, uint64_t version,
                              block_height_t height, bool depends_trace,
                              std::unordered_set<std::string> &pkg,
                              std::vector<std::unique_ptr<nbre::NBREIR>> &irs);

  block_height_t get_start_height();
  block_height_t get_end_height();

  void write_nbre();
  void write_nbre_by_height(block_height_t height);

  void set_auth_table();
  void set_auth_table_by_jit(std::unique_ptr<nbre::NBREIR> &nbre_ir);

  void update_ir_list(const std::string &nbre_ir_list_name,
                      const std::string &ir_name_list,
                      const std::string &ir_name);
  void update_ir_list_to_db(const std::string &nbre_ir_list_name,
                            const boost::property_tree::ptree &pt);

private:
  std::unique_ptr<rocksdb_storage> m_storage;
  std::unique_ptr<blockchain> m_blockchain;
  std::map<auth_key_t, auth_val_t> m_auth_table;
};
} // namespace fs
} // namespace neb
