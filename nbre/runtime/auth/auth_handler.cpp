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
#include "runtime/auth/auth_handler.h"
#include "fs/ir_manager/api/ir_list.h"
#include "runtime/auth/auth_table.h"

namespace neb {
namespace rt {

namespace auth {

auth_handler::auth_handler(fs::ir_list *ir_list) : m_ir_list(ir_list) {}

void auth_handler::handle_auth_npr(const nbre::NBREIR &compiled_ir) {}
} // namespace auth
} // namespace rt
} // namespace neb
