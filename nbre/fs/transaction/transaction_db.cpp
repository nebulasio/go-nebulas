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

#include "fs/transaction/transaction_db.h"

namespace neb {
namespace fs {

transaction_db::transaction_db(const std::string &path) : blockchain(path) {}

std::shared_ptr<std::vector<transaction_info_t>>
transaction_db::read_inter_transaction_from_db_with_duration(
    block_height_t start_block, block_height_t end_block) {

  std::vector<transaction_info_t> txs;

  for (block_height_t h = start_block; h < end_block; h++) {
    auto block = this->load_block_with_height(h);

    // TODO type revise
    std::string timestamp = std::to_string(block->header().timestamp());

    for (auto &tx : block->transactions()) {
      transaction_info_t info;
      neb::util::bytes from_bytes = neb::util::string_to_byte(tx.from());
      if (from_bytes[1] == 0x58) {
        continue;
      }
      neb::util::bytes to_bytes = neb::util::string_to_byte(tx.to());
      if (to_bytes[1] == 0x58) {
        continue;
      }

      info.m_from = tx.from();
      info.m_to = tx.to();
      info.m_tx_value = tx.value();
      info.m_timestamp = timestamp;

      txs.push_back(info);
    }
  }

  return std::make_shared<std::vector<transaction_info_t>>(txs);
}
} // namespace fs
} // namespace neb
