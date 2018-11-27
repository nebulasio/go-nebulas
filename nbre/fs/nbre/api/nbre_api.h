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
#include "fs/rocksdb_storage.h"

namespace neb {
namespace fs {

class nbre_api {
public:
  nbre_api(const std::string &db_path,
           enum storage_open_flag open_flag = storage_open_default);
  ~nbre_api();
  nbre_api(const nbre_api &na) = delete;
  nbre_api &operator=(const nbre_api &na) = delete;

  std::shared_ptr<std::vector<std::string>> get_irs();

  std::shared_ptr<std::vector<version_t>>
  get_ir_versions(const std::string &ir_name);

private:
  std::unique_ptr<rocksdb_storage> m_storage;
};
}
} // namespace neb
