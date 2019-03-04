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
#include "common/address.h"
#include "common/common.h"
#include "fs/bc_storage_session.h"
#include "fs/proto/block.pb.h"
#include "fs/rocksdb_storage.h"

namespace neb {
namespace fs {

class blockchain {
public:
  static constexpr char const *Block_LIB = "blockchain_lib";
  static constexpr char const *Block_Tail = "blockchain_tail";

  static std::unique_ptr<corepb::Block> load_LIB_block();
  static std::unique_ptr<corepb::Block>
  load_block_with_height(block_height_t height);

  static void write_LIB_block(corepb::Block *block);

private:
  static std::unique_ptr<corepb::Block>
  load_block_with_tag_string(const std::string &tag);
}; // end class blockchain
} // end namespace fs
} // end namespace neb
