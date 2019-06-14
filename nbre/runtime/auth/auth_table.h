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
#include "common/version.h"
#include "fs/proto/ir.pb.h"
#include "runtime/auth/data_type.h"

namespace neb {
namespace rt {
namespace auth {
class auth_table {
public:
  bool is_ir_legitimate(const std::string &ir_name, const address_t &from_addr,
                        block_height_t h) const;

  inline const version &version() const { return m_version; }
  inline block_height_t available_height() const { return m_available_height; }

  static auth_table
  generate_auth_table_from_ir(const nbre::NBREIR &compiled_ir);

protected:
  static std::string auth_key(const std::string &ir_name,
                              const address_t &from_addr);

protected:
  typedef std::tuple<std::string, address_t, block_height_t, block_height_t>
      auth_val_t; // start_block, end_block
  std::map<std::string, auth_val_t> m_auth_table;
  class version m_version;
  block_height_t m_available_height;
};

} // namespace auth
} // namespace rt
} // namespace neb

