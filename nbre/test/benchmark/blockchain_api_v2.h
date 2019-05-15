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
#include "fs/blockchain/blockchain_api.h"

namespace neb {
namespace fs {

class blockchain_api_v2 : public blockchain_api {
public:
  blockchain_api_v2();
  virtual ~blockchain_api_v2();

  virtual std::unique_ptr<std::vector<transaction_info_t>>
  get_block_transactions_api(block_height_t height);

protected:
  virtual std::unique_ptr<event_info_t>
  get_transaction_result_api(const neb::bytes &events_root,
                             const neb::bytes &tx_hash);
  std::unique_ptr<event_info_t> json_parse_event(const std::string &json);
};

} // namespace fs
} // namespace neb

