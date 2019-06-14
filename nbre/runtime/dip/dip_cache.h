/ Copyright(C) 2018 go - nebulas authors
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
#include "runtime/dip/data_type.h"

    namespace neb {
namespace rt {
namespace dip {
class dip_cache {
public:
  typedef std::function<dip_ret_type()> dip_function_t;
  virtual dip_ret_type get_dip_reward(const dip_function_t &func,
                                      block_height_t start_block,
                                      block_height_t end_block,
                                      uint64_t version,
                                      const dip_param_t &param);

  virtual dip_ret_type get_dip_reward(const std::string &handle);
};
} // namespace dip
} // namespace rt
} // namespace neb
