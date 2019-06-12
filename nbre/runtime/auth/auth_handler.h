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
#include "fs/proto/ir.pb.h"
#include "runtime/auth/data_type.h"

namespace neb {
namespace fs {
class ir_list;
}
namespace rt {
namespace auth {
class auth_table;
class auth_handler {
public:
  auth_handler(fs::ir_list *ir_list);

  virtual void handle_auth_nir(const nbre::NBREIR &ir);

  virtual bool is_ir_legitimate(const std::string ir_name,
                                const address_t &from_addr);

protected:
  fs::ir_list *m_ir_list;
  std::unique_ptr<auth_table> m_auth_table;
};
} // namespace auth
} // namespace rt
} // namespace neb
