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

#include "core/neb_ipc/server/ipc_configuration.h"
#include "fs/blockchain/transaction/transaction_db.h"
#include <gtest/gtest.h>

TEST(test_fs, read_inter_transaction_from_db_with_duration) {
  std::string neb_db_path =
      neb::core::ipc_configuration::instance().neb_db_dir();
  neb::fs::blockchain b(neb_db_path);
  neb::fs::blockchain_api ba(&b);
  neb::fs::transaction_db tdb(&ba);

  auto txs = tdb.read_transactions_from_db_with_duration(204223, 204224);
  for (auto &tx : *txs) {
    LOG(INFO) << neb::util::string_to_byte(tx.m_from).to_base58();
    LOG(INFO) << neb::util::string_to_byte(tx.m_to).to_base58();
    LOG(INFO) << tx.m_tx_value;
    LOG(INFO) << tx.m_timestamp;
  }
}
