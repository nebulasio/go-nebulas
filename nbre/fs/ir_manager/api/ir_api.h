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
#include "fs/proto/ir.pb.h"
#include "fs/rocksdb_storage.h"

namespace neb {
namespace fs {

class ir_api {
public:
  std::vector<std::string> get_ir_list() {
    std::vector<std::string> t;
    return t;
  };

  void get_ir_list(std::vector<std::string> *t) { return; }
  static std::unique_ptr<std::vector<std::string>>
  get_ir_list(rocksdb_storage *rs);

  static std::unique_ptr<std::vector<version_t>>
  get_ir_versions(const std::string &name, rocksdb_storage *rs);

  typedef std::pair<bool, std::unique_ptr<nbre::NBREIR>> ir_ret_type;
  static ir_ret_type get_ir(const std::string &name, version_t version,
                            rocksdb_storage *rs);

  static bool ir_exist(const std::string &name, version_t version,
                       rocksdb_storage *rs);

  void get_ir_depends(const std::string &name, version_t version,
                      rocksdb_storage *rs,
                      std::vector<std::pair<std::string, version_t>> &irs);
};
}
} // namespace neb
