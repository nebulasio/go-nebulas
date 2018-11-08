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

blockchain::blockchain(const std::string &path,
                       enum storage_open_flag open_flag)
    : m_path(path), m_open_flag(open_flag) {
  m_storage = std::make_unique<rocksdb_storage>();
  m_storage->open_database(path, open_flag);
}

blockchain::~blockchain() {
  if (m_storage) {
    m_storage->close_database();
  }
}

std::shared_ptr<corepb::Block> blockchain::load_tail_block() {
  return load_block_with_tag_string(
      std::string(Block_Tail, std::allocator<char>()));
}

std::shared_ptr<corepb::Block> blockchain::load_LIB_block() {
  return load_block_with_tag_string(
      std::string(Block_LIB, std::allocator<char>()));
}

std::shared_ptr<corepb::Block>
blockchain::load_block_with_height(block_height_t height) {
  std::shared_ptr<corepb::Block> block = std::make_shared<corepb::Block>();
  neb::util::bytes height_hash =
      m_storage->get_bytes(neb::util::number_to_byte<neb::util::bytes>(height));
  neb::util::bytes block_bytes = m_storage->get_bytes(height_hash);

  bool ret = block->ParseFromArray(block_bytes.value(), block_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse block failed");
  }
  return block;
}

std::shared_ptr<corepb::Block>
blockchain::load_block_with_tag_string(const std::string &tag) {
  if (m_storage) {
    m_storage->close_database();
  }
  m_storage->open_database(m_path, m_open_flag);

  std::shared_ptr<corepb::Block> block = std::make_shared<corepb::Block>();
  neb::util::bytes tail_hash =
      m_storage->get_bytes(neb::util::string_to_byte(tag));

  neb::util::bytes block_bytes = m_storage->get_bytes(tail_hash);

  bool ret = block->ParseFromArray(block_bytes.value(), block_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse block failed");
  }
  return block;
}
} // namespace fs
} // namespace neb
