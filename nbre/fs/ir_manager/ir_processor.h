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
#include "common/common.h"
#include "core/net_ipc/nipc_pkg.h"

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
class ir_processor {
public:
  ir_processor(storage *s);

  virtual void parse_irs(
      util::wakeable_queue<std::shared_ptr<nbre_ir_transactions_req>> &q_txs);

protected:
  storage *m_storage;
  std::unique_ptr<ir_list> m_ir_list;
  std::unique_ptr<rt::auth::auth_handler> m_auth_handler;
  std::unique_ptr<util::persistent_flag> m_failed_flag;
  std::unique_ptr<util::persistent_type<block_height_t>> m_nbre_block_height;
};
} // namespace fs
} // namespace neb
