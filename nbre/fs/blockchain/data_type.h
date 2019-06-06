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
} // namespace fs
} // namespace neb
