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

#include "fs/blockchain/transaction/transaction_db.h"
#include "fs/blockchain/blockchain_api.h"

namespace neb {
namespace fs {

transaction_db::transaction_db(blockchain_api_base *blockchain_ptr)
    : m_blockchain(blockchain_ptr) {}

std::vector<transaction_info_t>
transaction_db::read_transactions_from_db_with_duration(
    block_height_t start_block, block_height_t end_block) {

  std::vector<transaction_info_t> ret;

  for (block_height_t h = start_block; h < end_block; h++) {
    auto tmp = m_blockchain->get_block_transactions_api(h);
    ret.insert(ret.end(), tmp.begin(), tmp.end());
  }
  return ret;
}


} // namespace fs
} // namespace neb
