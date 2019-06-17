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
#include "common/byte.h"
#include "common/common.h"
#include "compatible/compatible_checker.h"
#include "runtime/nr/impl/nebulas_rank_cache.h"
#include "util/lru_cache.h"

namespace neb {
namespace core {

class nr_handler {
public:
  nr_handler();

  void start(block_height_t start_block, block_height_t end_block,
             uint64_t nr_version);

  std::string get_nr_handle(block_height_t start_block,
                            block_height_t end_block, uint64_t nr_version);

  std::string get_nr_handle(block_height_t height);

  rt::nr::nr_ret_type get_nr_result(const std::string &nr_handle);

  bool get_nr_sum(floatxx_t &nr_sum, const std::string &handle);
  bool get_nr_addr_list(std::vector<address_t> &nr_addrs,
                        const std::string &handle);

  /*
  void run_if_default(block_height_t start_block, block_height_t end_block,
                      const std::string &nr_handle);
  void run_if_specify(block_height_t start_block, block_height_t end_block,
                      uint64_t nr_version, const std::string &nr_handle);
*/
protected:
};
} // namespace core
} // namespace neb
