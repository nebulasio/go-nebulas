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
#include "common/math/softfloat.hpp"
#include "runtime/dip/dip_reward.h"
#include "runtime/stdrt.h"

namespace neb {
namespace rt {
namespace dip {

using dip_float_t = float32;
// using version_t = compatible_uint64_t;

dip_ret_type entry_point_dip_impl(compatible_uint64_t start_block,
                                  compatible_uint64_t end_block,
                                  version_t version, compatible_uint64_t height,
                                  const nr::nr_ret_type &nr_ret,
                                  dip_float_t alpha, dip_float_t beta);

dip_param_t make_dip_param(block_height_t start_block,
                           block_height_t block_interval,
                           const std::string &reward_addr,
                           const std::string &coinbase_addr, version_t v);

} // namespace dip
} // namespace rt
} // namespace neb
