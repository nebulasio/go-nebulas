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
#include "fs/blockchain/data_type.h"
#include "fs/proto/block.pb.h"

namespace neb {
namespace fs {
class blockchain;
class blockchain_api_base {
public:
  blockchain_api_base(blockchain *bc);
  virtual ~blockchain_api_base();
  virtual std::vector<transaction_info_t>
  get_block_transactions_api(block_height_t height) = 0;

  virtual std::unique_ptr<corepb::Account>
  get_account_api(const address_t &addr, block_height_t height = 0) = 0;
  virtual std::unique_ptr<corepb::Transaction>
  get_transaction_api(const bytes &tx_hash) = 0;

protected:
  blockchain *m_blockchain;
};

class blockchain_api : public blockchain_api_base {
public:
  blockchain_api(blockchain *bc);
  virtual ~blockchain_api();

  virtual std::vector<transaction_info_t>
  get_block_transactions_api(block_height_t height);

  virtual std::unique_ptr<corepb::Account>
  get_account_api(const address_t &addr, block_height_t height = 0);
  virtual std::unique_ptr<corepb::Transaction>
  get_transaction_api(const bytes &tx_hash);

  virtual void get_transfer_event(const neb::bytes &events_root,
                                  const neb::bytes &tx_hash,
                                  std::vector<transaction_info_t> &events,
                                  transaction_info_t &info);

protected:
  virtual void json_parse_event(const std::string &json,
                                std::vector<transaction_info_t> &events,
                                transaction_info_t &info);

};
} // namespace fs
} // namespace neb
