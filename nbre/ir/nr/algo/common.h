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

#pragma once

#include <boost/multiprecision/cpp_int.hpp>
#include <memory>
#include <unordered_map>
#include <unordered_set>
#include <vector>

namespace neb {
namespace nr {

using block_height_t = uint64_t;
using account_address_t = std::string;

using int128_t = boost::multiprecision::int128_t;
using account_balance_t = int128_t;

struct transaction_info_t {
  block_height_t m_height;
  std::string m_from;
  std::string m_to;
  std::string m_tx_value;
  std::string m_timestamp;
};

class transaction_db {
public:
  virtual std::shared_ptr<std::vector<transaction_info_t>>
  read_inter_transaction_from_db_with_duration(block_height_t start_block,
                                               block_height_t end_block) = 0;
};

class account_db {
public:
  virtual void set_height_address_val(
      block_height_t start_block, block_height_t end_block,
      std::unordered_map<account_address_t, account_balance_t>
          &addr_balance) = 0;
  virtual double get_account_balance(block_height_t height,
                                     account_address_t addr) = 0;
  virtual double get_normalized_value(double median) = 0;
};
}
} // namespace neb
