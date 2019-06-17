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
#include "common/version.h"
#include "compatible/version_check_base.h"

namespace neb {
namespace compatible {
namespace v {
class v_0_1_1_checker : public version_check_base {
public:
  v_0_1_1_checker();

  virtual bool is_ir_need_compile(const std::string &module_name,
                                  version_t version);

  virtual bool is_compatible_for(const std::string module_name, version_t v);

  virtual optional<rt::nr::nr_ret_type>
  get_nr_result(block_height_t start_block, block_height_t end_block,
                uint64_t version);

  virtual optional<rt::dip::dip_ret_type>
  get_dip_result(block_height_t start_block, block_height_t end_block,
                 uint64_t version);
};
} // namespace v
} // namespace compatible
} // namespace neb
