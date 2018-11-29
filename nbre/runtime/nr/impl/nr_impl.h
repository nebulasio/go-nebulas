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

#include "runtime/stdrt.h"
#include <string>
#include <vector>

struct nr_info_t {
  std::string m_address;
  std::string m_date;
  uint32_t m_in_degree;
  uint32_t m_out_degree;
  uint32_t m_degrees;
  long double m_in_val;
  long double m_out_val;
  long double m_in_outs;
  double m_median;
  double m_weight;
  double m_nr_score;
};

std::vector<nr_info_t> entry_point_nr_impl(neb::core::driver *d, void *param);
