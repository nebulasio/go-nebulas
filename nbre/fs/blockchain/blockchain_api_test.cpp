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
#include "fs/blockchain/blockchain_api_test.h"

namespace neb {
namespace fs {

blockchain_api_test::~blockchain_api_test() {}

std::unique_ptr<std::vector<transaction_info_t>>
blockchain_api_test::get_block_transactions_api(block_height_t height) {
  auto ret = std::make_unique<std::vector<transaction_info_t>>();

  return ret;
}

std::unique_ptr<corepb::Account>
blockchain_api_test::get_account_api(const address_t &addr,
                                     block_height_t height) {
  auto ret = std::make_unique<corepb::Account>();
  return ret;
}

std::unique_ptr<corepb::Transaction>
blockchain_api_test::get_transaction_api(const std::string &tx_hash,
                                         block_height_t height) {
  auto ret = std::make_unique<corepb::Transaction>();

  return ret;
}
} // namespace fs
} // namespace neb
