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
#include "fs/blockchain.h"

namespace neb {
namespace fs {

namespace util {
wei_t hex_val_cast(const std::string &hex_str);
} // namespace util

struct transaction_info_t {
  block_height_t m_height;
  int32_t m_status;
  std::string m_from;
  std::string m_to;
  std::string m_tx_type;
  wei_t m_tx_value;
  int64_t m_timestamp;
  wei_t m_gas_used;
  wei_t m_gas_price;
};

struct account_info_t {
  std::string m_address;
  wei_t m_balance;
};

struct event_info_t {
  int32_t m_status;
  wei_t m_gas_used;
};

class blockchain_api {
public:
  blockchain_api(blockchain *blockchain_ptr);

  std::unique_ptr<std::vector<transaction_info_t>>
  get_block_transactions_api(block_height_t height);

  std::unique_ptr<corepb::Account> get_account_api(const address_t &addr,
                                                   block_height_t height);
  std::unique_ptr<corepb::Transaction>
  get_transaction_api(const std::string &tx_hash, block_height_t height);

private:
  std::unique_ptr<event_info_t>
  get_transaction_result_api(const neb::util::bytes &events_root,
                             const neb::util::bytes &tx_hash);

  std::unique_ptr<event_info_t> json_parse_event(const std::string &json);

private:
  blockchain *m_blockchain;
};
} // namespace fs
} // namespace neb
