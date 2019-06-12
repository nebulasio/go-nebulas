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
#include "fs/proto/ir.pb.h"
#include "runtime/auth/data_type.h"

namespace neb {
namespace rt {
namespace auth {
class auth_table {
public:
  typedef std::tuple<std::string, address_t> auth_key_t; // ir_name, address
  typedef std::tuple<block_height_t, block_height_t>
      auth_val_t; // start_block, end_block

  virtual void update_auth_table(const auth::auth_items_t &auth_items);

  virtual bool is_ir_legitimate(const std::string ir_name,
                                const address_t &from_addr);

protected:
  typedef std::map<auth_key_t, auth_val_t> auth_table_t;
  auth_table_t m_auth_table;
};

} // namespace auth
} // namespace rt
} // namespace neb

