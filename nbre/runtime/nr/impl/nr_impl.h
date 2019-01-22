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
namespace nr {

typedef typename infer_type_helper<uint64_t>::type uintyy_t;
typedef typename infer_type_helper<int64_t>::type intyy_t;

using nr_float_t = float32;
using version_t = uintyy_t;
std::string entry_point_nr_impl(uintyy_t start_block, uintyy_t end_block,
                                version_t version, intyy_t a, intyy_t b,
                                intyy_t c, intyy_t d, nr_float_t theta,
                                nr_float_t mu, nr_float_t lambda);
} // namespace nr
} // namespace rt
} // namespace neb
