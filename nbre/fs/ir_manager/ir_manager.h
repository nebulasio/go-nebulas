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
#include "common/common.h"
#include "core/net_ipc/nipc_pkg.h"
#include "fs/ir_manager/ir_manager_helper.h"
#include "util/wakeable_queue.h"

namespace neb {
namespace fs {

class ir_manager {
public:
  ir_manager();
  ~ir_manager();
  ir_manager(const ir_manager &im) = delete;
  ir_manager &operator=(const ir_manager &im) = delete;

  std::unique_ptr<nbre::NBREIR> read_ir(const std::string &name,
                                        version_t version);
  std::unique_ptr<std::vector<nbre::NBREIR>>
  read_irs(const std::string &name, block_height_t height, bool depends);

  void parse_irs(
      util::wakeable_queue<std::shared_ptr<nbre_ir_transactions_req>> &q_txs);

private:
  void read_ir_depends(const std::string &name, version_t version,
                       block_height_t height, bool depends,
                       std::vector<nbre::NBREIR> &irs);

  void parse_next_block(block_height_t height,
                        const std::vector<std::string> &txs_seri);
  void parse_when_missing_block(block_height_t start_block,
                                block_height_t end_height);

  void parse_irs_by_height(block_height_t height,
                           const std::vector<corepb::Transaction> &txs);
  void parse_with_height(block_height_t height,
                         const std::vector<corepb::Transaction> &txs);

  void deploy_if_dip(const std::string &name, version_t version,
                     block_height_t available_height);

private:
  rocksdb_storage *m_storage;
  std::map<auth_key_t, auth_val_t> m_auth_table;
};
} // namespace fs
} // namespace neb
