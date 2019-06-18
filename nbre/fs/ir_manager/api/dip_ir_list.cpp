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
#include "fs/ir_manager/api/dip_ir_list.h"
#include "common/configuration.h"
#include "runtime/dip/data_type.h"
#include "runtime/param_trait.h"

namespace neb {
namespace fs {
dip_ir_list::dip_ir_list(storage *storage)
    : internal::ir_item_list_base<dip_params_storage_t>(storage, "dip") {}

dip_ir_list::item_type
dip_ir_list::get_ir_param(const nbre::NBREIR &compiled_ir) {
  rt::dip::dip_param_t param = rt::param_trait::get_param<rt::dip::dip_param_t>(
      compiled_ir, configuration::instance().dip_param_func_name());

  dip_param_storage_t ret;

  ret.set<p_start_block, p_version>(compiled_ir.height(),
                                    compiled_ir.version());
  return ret;
}

} // namespace fs
} // namespace neb
