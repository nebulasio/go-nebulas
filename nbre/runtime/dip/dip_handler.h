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
#include "common/util/singleton.h"

namespace neb {
namespace rt {
namespace dip {

class dip_handler : public util::singleton<dip_handler> {
public:
  static const uint64_t block_nums_of_a_day = 100;
  static const uint64_t days = 3;
  static const uint64_t dip_start_block = 1;
  static const uint64_t dip_block_interval = days * block_nums_of_a_day;

  void start(neb::block_height_t nbre_max_height,
             neb::block_height_t lib_height);

  std::string get_dip_reward(neb::block_height_t height);

private:
  mutable std::mutex m_sync_mutex;
  std::unordered_map<neb::block_height_t, std::string> m_dip_reward;
};
} // namespace dip
} // namespace rt
} // namespace neb
