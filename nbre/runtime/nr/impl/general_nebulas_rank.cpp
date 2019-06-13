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
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the // GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//
#include "runtime/nr/impl/general_nebulas_rank.h"
#include "runtime/nr/impl/nebulas_rank_cache.h"
#include "runtime/nr/impl/nebulas_rank_calculator.h"
#include "runtime/util.h"

namespace neb {
namespace rt {
namespace nr {

general_nebulas_rank::general_nebulas_rank(nebulas_rank_calculator *calculator,
                                           nebulas_rank_cache *cache)
    : m_calculator(calculator), m_cache(cache) {
  if (!m_calculator) {
    throw std::runtime_error("no nr calculator available");
  }
  if (!m_cache) {
    throw std::runtime_error("no nr cache available");
  }
}

nr_ret_type general_nebulas_rank::get_nr_score(const rank_params_t &rp,
                                               block_height_t start_block,
                                               block_height_t end_block,
                                               uint64_t version) {

  auto functor = [this, rp, start_block, end_block, version]() {
    std::vector<nr_item> nrs =
        m_calculator->get_nr_score(rp, start_block, end_block);
    std::shared_ptr<nr_result> result = std::make_shared<nr_result>();
    result->set<p_start_block, p_end_block, p_nr_version>(start_block,
                                                          end_block, version);
    result->set<p_nr_items>(nrs);
    result->set<p_result_status>(core::result_status::succ);
    return result;
  };

  return m_cache->get_nr_score(functor, start_block, end_block, version);
}
} // namespace nr
} // namespace rt
} // namespace neb
