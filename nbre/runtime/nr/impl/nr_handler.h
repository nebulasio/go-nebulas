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
#include "common/util/byte.h"
#include "common/util/lru_cache.h"
#include "common/util/singleton.h"

namespace neb {
namespace rt {
namespace nr {

class nr_handler : public util::singleton<nr_handler> {
public:
  nr_handler();

  std::string get_nr_handler_id();

  void start(std::string nr_handler_id);

  void run_if_default(block_height_t start_block, block_height_t end_block);
  void run_if_specify(block_height_t start_block, block_height_t end_block,
                      uint64_t nr_version);

  std::string get_nr_result(const std::string &nr_handler_id);

private:
  mutable std::mutex m_sync_mutex;
  std::string m_nr_handler_id;
  lru_cache<std::string, std::string> m_nr_result;
};
} // namespace nr
} // namespace rt
} // namespace neb
