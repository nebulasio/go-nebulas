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

#include "runtime/dip/dip_impl.h"

std::string entry_point_nr(uint64_t start_block, uint64_t end_block);

std::string entry_point_dip(uint64_t height) {
  uint64_t block_nums_of_a_day = 100;
  uint64_t days = 3;
  uint64_t dip_start_block = 1;
  uint64_t dip_block_interval = days * block_nums_of_a_day;
  std::string dip_reward_addr =
      std::string("n1c6y4ctkMeZk624QWBTXuywmNpCWmJZiBq");

  if (!height) {
    neb::rt::dip::init_dip_params(dip_start_block, dip_block_interval,
                                  dip_reward_addr);
    return std::string("{\"res\":\"init dip params\"}");
  }

  if (height < dip_start_block + dip_block_interval) {
    return std::string("{\"err\":\"invalid height\"}");
  }
  uint64_t interval_nums = (height - dip_start_block) / dip_block_interval;
  uint64_t start_block = dip_start_block + dip_block_interval * interval_nums;
  uint64_t end_block = start_block - 1;
  start_block -= dip_block_interval;

  std::string nr_result = entry_point_nr(start_block, end_block);

  neb::rt::dip::dip_float_t alpha = 1;
  neb::rt::dip::dip_float_t beta = 1;
  return neb::rt::dip::entry_point_dip_impl(start_block, end_block, height,
                                            nr_result, alpha, beta);
}
