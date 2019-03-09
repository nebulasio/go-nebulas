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

#include "common/address.h"
#include "runtime/dip/dip_impl.h"
#include "runtime/dip/dip_reward.h"
#include <vector>

void height_interval(uint64_t height) {

  uint64_t block_nums_of_a_day = 97;
  uint64_t days = 3;
  uint64_t dip_start_block = 232;
  uint64_t dip_block_interval = days * block_nums_of_a_day;
  assert(height >= dip_start_block + dip_block_interval);

  uint64_t interval_nums = (height - dip_start_block) / dip_block_interval;
  uint64_t start_block = dip_start_block + dip_block_interval * interval_nums;
  uint64_t end_block = start_block - 1;
  start_block -= dip_block_interval;

  std::cout << "block interval [" << start_block << ',' << end_block << ']'
            << std::endl;
}

void nr(uint64_t start_block, uint64_t end_block) {

  if (start_block > end_block) {
    std::cout << std::string(
                     "{\"err\":\"start height must less than end height\"}")
              << std::endl;
    return;
  }
  uint64_t block_nums_of_a_day = 24 * 3600 / 15;
  uint64_t days = 7;
  uint64_t max_nr_block_interval = days * block_nums_of_a_day;
  if (start_block + max_nr_block_interval < end_block) {
    std::cout << std::string("{\"err\":\"nr block interval out of range\"}")
              << std::endl;
    return;
  }
  std::cout << "run nr successfully" << std::endl;
}

void dip(uint64_t height) {

  uint64_t block_nums_of_a_day = 24 * 3600 / 15;
  uint64_t days = 7;
  uint64_t dip_start_block = 1540000;
  uint64_t dip_block_interval = days * block_nums_of_a_day;
  neb::base58_address_t dip_reward_addr =
      std::string("n1YubAA3VVi2HEDw3VSaJ2ZcjzYKXL6SuQw");

  if (!height) {
    auto ret = neb::rt::dip::dip_param_list(dip_start_block, dip_block_interval,
                                            dip_reward_addr, std::string(), 0);
    std::cout << ret << std::endl;
    return;
  }

  if (height < dip_start_block + dip_block_interval) {
    std::cout << std::string("{\"err\":\"invalid height\"}") << std::endl;
    return;
  }

  uint64_t interval_nums = (height - dip_start_block) / dip_block_interval;
  uint64_t start_block = dip_start_block + dip_block_interval * interval_nums;
  uint64_t end_block = start_block - 1;
  start_block -= dip_block_interval;

  std::cout << "start height: " << start_block << ", end height: " << end_block
            << std::endl;
  nr(start_block, end_block);
}

int main(int argc, char *argv[]) {

  using dip_info_t = neb::rt::dip::dip_info_t;
  using floatxx_t = neb::floatxx_t;
  std::vector<std::shared_ptr<dip_info_t>> v;
  v.push_back(std::shared_ptr<dip_info_t>(new dip_info_t{
      neb::to_address("addr1"), neb::to_address("addr3"), "1.1"}));
  v.push_back(std::shared_ptr<dip_info_t>(new dip_info_t{
      neb::to_address("addr2"), neb::to_address("addr4"), "2.1"}));

  std::cout << neb::rt::dip::dip_reward::dip_info_to_json(v) << std::endl;

  uint64_t height;
  uint64_t start_block, end_block;

  while (std::cin >> height) {
    // nr(start_block, end_block);
    dip(height);
  }
  return 0;
}
