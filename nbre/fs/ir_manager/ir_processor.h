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
#include "common/byte.h"
#include "common/common.h"
#include "core/net_ipc/nipc_pkg.h"
#include "fs/proto/block.pb.h"
#include "fs/proto/ir.pb.h"

namespace neb {
namespace util {
template <class T> class wakeable_queue;
template <class T> class persistent_type;
class persistent_flag;
}
namespace rt {
namespace auth {
class auth_handler;
}
} // namespace rt
namespace fs {
class ir_list;
class storage;
class blockchain;
class ir_processor {
public:
  ir_processor(storage *s, blockchain *bc);

  virtual optional<nbre::NBREIR> get_ir_with_version(const std::string &name,
                                                     version_t v);

  virtual optional<bytes> get_ir_brief_key_with_height(const std::string &name,
                                                       block_height_t h);

  virtual optional<nbre::NBREIR> get_ir_with_height(const std::string &name,
                                                    block_height_t h);

  //! include param itself
  std::vector<nbre::NBREIR> get_ir_depends(const nbre::NBREIR &ir);

  std::vector<std::string> get_ir_names() const;

  std::vector<version_t> get_ir_versions(const std::string &name) const;

  virtual void parse_irs(
      util::wakeable_queue<std::shared_ptr<nbre_ir_transactions_req>> &q_txs);

  inline storage *storage() { return m_storage; }

  inline blockchain *blockchain() {
    return m_blockchain;
  }

protected:
  virtual void parse_missed_blocks_between(block_height_t start_block,
                                           block_height_t end_block);

  virtual void
  parse_block_with_height_and_txs(block_height_t height,
                                  const std::vector<std::string> &str_txs);
  virtual void
  parse_block_with_height_and_txs(block_height_t height,
                                  const std::vector<corepb::Transaction> &txs);

  virtual void parse_ir_txs(block_height_t height,
                            const std::vector<corepb::Transaction> &txs);

  virtual nbre::NBREIR compile_payload_code(const nbre::NBREIR &raw_ir);

protected:
  class storage *m_storage;
  class blockchain *m_blockchain;
  std::unique_ptr<ir_list> m_ir_list;
  std::unique_ptr<rt::auth::auth_handler> m_auth_handler;
  std::unique_ptr<util::persistent_flag> m_failed_flag;
  std::unique_ptr<util::persistent_type<block_height_t>> m_nbre_block_height;
};
} // namespace fs
} // namespace neb
