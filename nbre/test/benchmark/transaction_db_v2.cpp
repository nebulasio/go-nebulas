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
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the // GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//

#include "test/benchmark/transaction_db_v2.h"

namespace neb {
namespace fs {

transaction_db_v2::transaction_db_v2(transaction_db *tdb_ptr)
    : m_tdb_ptr(tdb_ptr) {}

std::unique_ptr<std::vector<transaction_info_t>>
transaction_db_v2::read_transactions_from_db_with_duration(
    block_height_t start_block, block_height_t end_block) {
  return m_tdb_ptr->read_transactions_from_db_with_duration(start_block,
                                                            end_block);
}

std::unique_ptr<std::vector<transaction_info_t>>
transaction_db_v2::read_transactions_with_succ(
    const std::vector<transaction_info_t> &txs) {
  auto ptr = std::make_unique<std::vector<transaction_info_t>>();
  for (auto &tx : txs) {
    if (tx.m_status) {
      ptr->push_back(tx);
    }
  }
  return ptr;
}

} // namespace fs
} // namespace neb
