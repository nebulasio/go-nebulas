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

#include "fs/blockchain.h"
#include "common/util/byte.h"

namespace neb {
namespace fs {
blockchain::blockchain(const std::string &path) {
  m_storage = std::unique_ptr<rocksdb_storage>(new rocksdb_storage());
  m_storage->open_database(path, storage_open_for_readonly);
}

std::shared_ptr<corepb::Block> blockchain::load_tail_block() {
  return load_block_with_tag_string(Block_Tail);
}

std::shared_ptr<corepb::Block> blockchain::load_LIB_block() {
  return load_block_with_tag_string(Block_LIB);
}

std::shared_ptr<corepb::Block>
blockchain::load_block_with_height(block_height_t height) {
  std::shared_ptr<corepb::Block> block = std::make_shared<corepb::Block>();
  if (!m_storage) {
    return block;
  }

  neb::util::bytes height_hash =
      m_storage->get_bytes(neb::util::number_to_byte<neb::util::bytes>(height));
  neb::util::bytes block_bytes = m_storage->get_bytes(height_hash);
  block->ParseFromArray(block_bytes.value(), block_bytes.size());
  return block;
}

std::shared_ptr<corepb::Block>
blockchain::load_block_with_tag_string(const std::string &tag) {
  std::shared_ptr<corepb::Block> block = std::make_shared<corepb::Block>();
  if (!m_storage)
    return block;
  neb::util::bytes tail_hash =
      m_storage->get_bytes(neb::util::string_to_byte(tag));

  neb::util::bytes block_bytes = m_storage->get_bytes(tail_hash);

  block->ParseFromArray(block_bytes.value(), block_bytes.size());
  return block;
}
}
}
