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
#include "fs/blockchain.h"
#include "fs/proto/ir.pb.h"
#include "fs/rocksdb_storage.h"

namespace neb {
namespace fs {

typedef std::tuple<module_t, version_t, address_t> auth_key_t;
typedef std::tuple<start_block_t, end_block_t> auth_val_t;

class ir_manager_helper {
public:
  static void set_failed_flag(rocksdb_storage *rs, const std::string &flag);
  static bool has_failed_flag(rocksdb_storage *rs, const std::string &flag);
  static void del_failed_flag(rocksdb_storage *rs, const std::string &flag);

  static block_height_t nbre_block_height(rocksdb_storage *rs);
  static block_height_t lib_block_height(blockchain *bc);

  static void run_auth_table(nbre::NBREIR &nbre_ir,
                             std::map<auth_key_t, auth_val_t> &auth_table);
  static void load_auth_table(rocksdb_storage *rs,
                              std::map<auth_key_t, auth_val_t> &auth_table);

  static void update_ir_list(const std::string &ir_list,
                             const std::string &ir_name, rocksdb_storage *rs);
};

} // namespace fs
} // namespace neb
