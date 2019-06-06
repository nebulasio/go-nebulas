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
#include "runtime/dip/data_type.h"
#include "runtime/nr/impl/data_type.h"

namespace neb {
namespace compatible {
class compatible_check_base {
public:
  virtual bool is_ir_need_compile(const std::string &name,
                                  uint64_t version) = 0;
  virtual bool get_nr_result(rt::nr::nr_ret_type &nr,
                             block_height_t start_block,
                             block_height_t end_block, uint64_t version) = 0;
  virtual bool get_dip_result(rt::dip::dip_ret_type &dip,
                              block_height_t start_block,
                              block_height_t end_block, uint64_t version) = 0;
};
} // namespace compatible
} // namespace neb
