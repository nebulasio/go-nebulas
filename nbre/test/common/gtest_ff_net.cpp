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

#include "runtime/dip/data_type.h"
#include <ff/network.h>
#include <gtest/gtest.h>

TEST(test_ff_net, seri_and_deseri) {
  neb::rt::dip::dip_param_t param;
  param.set<p_start_block>(1);
  param.set<p_block_interval>(2);
  param.set<p_dip_reward_addr>(std::to_string(3));
  param.set<p_dip_coinbase_addr>(std::to_string(4));
  param.set<p_version>(0);
  auto ret = param.serialize_to_string();
  // std::cout << ret << std::endl;
}
