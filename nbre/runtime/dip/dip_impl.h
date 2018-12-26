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

#include "common/math/softfloat.hpp"
#include <string>

namespace neb {
namespace rt {
namespace dip {

using dip_float_t = float32;
using version_t = uint64_t;
std::string entry_point_dip_impl(uint64_t start_block, uint64_t end_block,
                                 uint64_t height, const std::string &nr_result,
                                 dip_float_t alpha, dip_float_t beta,
                                 version_t version);

void init_dip_params(uint64_t dip_start_block, uint64_t dip_block_interval,
                     const std::string &dip_reward_addr);
} // namespace dip
} // namespace rt
} // namespace neb
