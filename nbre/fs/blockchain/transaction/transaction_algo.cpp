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
#include "fs/blockchain/transaction/transaction_algo.h"
namespace neb {
namespace fs {
namespace algo {

std::vector<transaction_info_t>
read_transactions_with_address_type(const std::vector<transaction_info_t> &txs,
                                    byte_t from_type, byte_t to_type) {

  auto ret = std::vector<transaction_info_t>();
  for (auto &tx : txs) {
    neb::bytes from_bytes = tx.m_from;
    neb::bytes to_bytes = tx.m_to;

    if (from_bytes[1] == from_type && to_bytes[1] == to_type) {
      ret.push_back(tx);
    }
  }
  return ret;
}
std::vector<transaction_info_t>
read_transactions_with_succ(const std::vector<transaction_info_t> &txs) {
  auto ptr = std::vector<transaction_info_t>();
  for (auto &tx : txs) {
    if (tx.m_status) {
      ptr.push_back(tx);
    }
  }
  return ptr;
}
} // namespace algo
} // namespace fs
} // namespace neb
