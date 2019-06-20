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

#pragma once
#include "runtime/nr/impl/data_type.h"

namespace neb {
namespace fs {
class transaction_db_interface;
class account_db_interface;
} // namespace fs

namespace rt {
class graph_algo;
namespace nr {
class nebulas_rank_algo;
class nebulas_rank_calculator {
public:
  nebulas_rank_calculator(graph_algo *galgo, nebulas_rank_algo *nralgo,
                          fs::transaction_db_interface *tdb_ptr,
                          fs::account_db_interface *adb_ptr);

  virtual std::vector<nr_item> get_nr_score(const rank_params_t &rp,
                                            neb::block_height_t start_block,
                                            neb::block_height_t end_block);

protected:
  graph_algo *m_graph_algo;
  nebulas_rank_algo *m_nr_algo;
  fs::transaction_db_interface *m_tdb_ptr;
  fs::account_db_interface *m_adb_ptr;
};
} // namespace nr
} // namespace rt
} // namespace neb
