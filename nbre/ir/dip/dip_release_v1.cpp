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

std::string entry_point_nr(neb::compatible_uint64_t start_block,
                           neb::compatible_uint64_t end_block);

std::string entry_point_dip(neb::compatible_uint64_t height) {
  uint64_t block_nums_of_a_day = 10;
  uint64_t days = 2;
  neb::compatible_uint64_t dip_start_block = 140;
  neb::compatible_uint64_t dip_block_interval = days * block_nums_of_a_day;
  std::string dip_reward_addr =
      std::string("n1c6y4ctkMeZk624QWBTXuywmNpCWmJZiBq");
  std::string coinbase_addr =
      std::string("n1HrPpwwH5gTA2d7QCkVjMw14YbN1NNNXHc");

  if (!height) {
    neb::rt::dip::init_dip_params(dip_start_block, dip_block_interval,
                                  dip_reward_addr, coinbase_addr);
    return std::string("{\"err\":\"init dip params\"}");
  }

  if (height < dip_start_block + dip_block_interval) {
    return std::string("{\"err\":\"invalid height\"}");
  }

  auto to_version_t = [](uint32_t major_version, uint16_t minor_version,
                         uint16_t patch_version) -> neb::rt::dip::version_t {
    return (0ULL + major_version) + ((0ULL + minor_version) << 32) +
           ((0ULL + patch_version) << 48);
  };

  uint64_t interval_nums = (height - dip_start_block) / dip_block_interval;
  neb::compatible_uint64_t start_block =
      dip_start_block + dip_block_interval * interval_nums;
  neb::compatible_uint64_t end_block = start_block - 1;
  start_block -= dip_block_interval;

  std::string nr_result = entry_point_nr(start_block, end_block);

  neb::rt::dip::dip_float_t alpha = 1e-32;
  neb::rt::dip::dip_float_t beta = 1;
  return neb::rt::dip::entry_point_dip_impl(start_block, end_block,
                                            to_version_t(0, 0, 1), height,
                                            nr_result, alpha, beta);
}
