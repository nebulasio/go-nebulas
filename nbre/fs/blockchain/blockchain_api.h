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
#include "common/address.h"
#include "fs/blockchain.h"

namespace neb {
namespace fs {

struct transaction_info_t {
  block_height_t m_height;
  int32_t m_status; // 0: fail, 1: succ, 2: special
  address_t m_from;
  address_t m_to;
  std::string m_tx_type; // "binary", "call", "deploy", "protocol"
  wei_t m_tx_value;
  int64_t m_timestamp; // no use
  wei_t m_gas_used;
  wei_t m_gas_price;
};

struct account_info_t {
  address_t m_address;
  wei_t m_balance;
};

struct event_info_t {
  int32_t m_status;
  wei_t m_gas_used;
};

class blockchain_api_base {
public:
  virtual ~blockchain_api_base();
  virtual std::unique_ptr<std::vector<transaction_info_t>>
  get_block_transactions_api(block_height_t height) = 0;

  virtual std::unique_ptr<corepb::Account>
  get_account_api(const address_t &addr, block_height_t height) = 0;
  virtual std::unique_ptr<corepb::Transaction>
  get_transaction_api(const std::string &tx_hash, block_height_t height) = 0;
};

class blockchain_api : public blockchain_api_base {
public:
  blockchain_api();
  virtual ~blockchain_api();

  virtual std::unique_ptr<std::vector<transaction_info_t>>
  get_block_transactions_api(block_height_t height);

  virtual std::unique_ptr<corepb::Account>
  get_account_api(const address_t &addr, block_height_t height);
  virtual std::unique_ptr<corepb::Transaction>
  get_transaction_api(const std::string &tx_hash, block_height_t height);

private:
  std::unique_ptr<event_info_t>
  get_transaction_result_api(const neb::bytes &events_root,
                             const neb::bytes &tx_hash);
  std::unique_ptr<event_info_t> json_parse_event(const std::string &json);
};

} // namespace fs
} // namespace neb
