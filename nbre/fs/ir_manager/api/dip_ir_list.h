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
#include "fs/ir_manager/api/ir_item_list_base.h"
#include "fs/nt_item_simple_storage.h"

namespace neb {
namespace fs {
class dip_ir_list : public internal::ir_item_list_base<dip_params_storage_t> {
public:
  typedef dip_params_storage_t::item_type item_type;
  dip_ir_list(storage *storage);
  virtual item_type get_ir_param(const nbre::NBREIR &compiled_ir);
};
} // namespace fs
} // namespace neb
