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

#include "runtime/dip/dip_impl.h"
#include "common/common.h"
#include "common/configuration.h"
#include "runtime/nr/impl/nebulas_rank.h"

namespace neb {
namespace rt {
namespace dip {

std::string entry_point_dip_impl(uint64_t height) {

  neb::block_height_t start_block = 1111780;
  neb::block_height_t end_block = 1117539;

  std::string neb_db_path = neb::configuration::instance().neb_db_dir();
  neb::fs::blockchain bc(neb_db_path);
  neb::fs::blockchain_api ba(&bc);
  neb::rt::nr::transaction_db_ptr_t tdb_ptr =
      std::make_shared<neb::fs::transaction_db>(&ba);

  auto it_txs =
      tdb_ptr->read_transactions_from_db_with_duration(start_block, end_block);
  auto it_account_call_contract_txs =
      tdb_ptr->read_transactions_with_address_type(*it_txs, 0x57, 0x58);
  return std::string();
}

} // namespace dip
} // namespace rt
} // namespace neb
