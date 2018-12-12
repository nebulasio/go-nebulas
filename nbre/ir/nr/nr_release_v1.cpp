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

#include "runtime/nr/impl/nr_impl.h"

std::string entry_point_nr(uint64_t start_block, uint64_t end_block) {
  neb::rt::nr::nr_float_t a = 2000.0;
  neb::rt::nr::nr_float_t b = 200000.0;
  neb::rt::nr::nr_float_t c = 100.0;
  neb::rt::nr::nr_float_t d = 1000.0;
  int64_t mu = 1;
  int64_t lambda = 3;
  return neb::rt::nr::entry_point_nr_impl(start_block, end_block, a, b, c, d,
                                          mu, lambda);
}

