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
#include "util/db_mem_cache.h"
#include "util/one_time_calculator.h"

namespace neb {
namespace fs {
class storage;
}
namespace rt {
namespace nr {
namespace internal {

class nr_db_mem_data
    : public util::db_mem_cache<std::string, std::shared_ptr<nr_result>> {
public:
  nr_db_mem_data(fs::storage *db);

  virtual bytes get_key_bytes(const std::string &k);
  virtual bytes serialize_data_to_bytes(const std::shared_ptr<nr_result> &v);
  virtual std::shared_ptr<nr_result>
  deserialize_data_from_bytes(const bytes &data);
};
} // namespace internal

class nebulas_rank_cache {
public:
  typedef std::function<std::shared_ptr<nr_result>()> nr_function_t;

  nebulas_rank_cache(fs::storage *s);

  virtual nr_ret_type get_nr_score(const nr_function_t &func,
                                   block_height_t start_block,
                                   block_height_t end_block, uint64_t version);

  virtual nr_ret_type get_nr_score(const std::string &handle);

protected:
  fs::storage *m_storage;
  typedef util::one_time_calculator<std::string, nr_ret_type,
                                    internal::nr_db_mem_data>
      calculator_t;
  std::unique_ptr<calculator_t> m_calculator;
};
extern std::shared_ptr<nebulas_rank_cache> nr_cache;
} // namespace nr
} // namespace rt
} // namespace neb
