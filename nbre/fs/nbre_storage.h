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

namespace neb {
namespace fs {

class nbre_storage {
public:
  nbre_storage(const std::string &path, const std::string &bc_path);
  nbre_storage(const nbre_storage &ns) = delete;
  nbre_storage &operator=(const nbre_storage &ns) = delete;

  std::shared_ptr<nbre::NBREIR>
  read_nbre_by_name_version(const std::string &name, uint64_t version);
  void write_nbre_by_height(block_height_t height);

private:
  std::unique_ptr<rocksdb_storage> m_storage;
  std::unique_ptr<blockchain> m_blockchain;

  static constexpr char const *m_payload_type = "protocol";
};
} // namespace fs
} // namespace neb
