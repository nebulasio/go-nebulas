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

#pragma once
#include "fs/blockchain/transaction/transaction_db.h"

namespace neb {
namespace fs {

class transaction_db_v2 {
public:
  transaction_db_v2(transaction_db *tdb_ptr);

  std::unique_ptr<std::vector<transaction_info_t>>
  read_transactions_from_db_with_duration(block_height_t start_block,
                                          block_height_t end_block);

  static std::unique_ptr<std::vector<transaction_info_t>>
  read_transactions_with_succ(const std::vector<transaction_info_t> &txs);

  static std::unique_ptr<std::vector<transaction_info_t>>
  read_transactions_with_address_type(
      const std::vector<transaction_info_t> &txs, byte_t from_type,
      byte_t to_type);

protected:
  transaction_db *m_tdb_ptr;
};
} // namespace fs
} // namespace neb
