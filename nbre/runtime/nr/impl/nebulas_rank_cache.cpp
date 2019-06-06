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
#include "runtime/nr/impl/nebulas_rank_cache.h"
#include "fs/storage_holder_interface.h"
#include "util/db_mem_cache.h"
#include "util/lru_cache.h"
#include "util/one_time_calculator.h"
#include "util/singleton.h"

namespace neb {
namespace rt {
namespace nr {
class nr_db_mem_data
    : public util::db_mem_cache<std::string, std::shared_ptr<nr_result>> {
public:
  virtual bytes get_key_bytes(const std::string &k) {
    return string_to_byte(std::string("cache_nr_") + k);
  }
  virtual bytes serialize_data_to_bytes(const std::shared_ptr<nr_result> &v) {
    return string_to_byte(v->serialize_to_string());
  }
  virtual std::shared_ptr<nr_result>
  deserialize_data_from_bytes(const bytes &data) {
    std::shared_ptr<nr_result> v = std::make_shared<nr_result>();
    auto str_data = byte_to_string(data);
    v->deserialize_from_string(str_data);
    return v;
  }
};

class nebulas_rank_cache_instance
    : public util::singleton<nebulas_rank_cache_instance>,
      public util::one_time_calculator<std::string, nr_ret_type,
                                       nr_db_mem_data> {};

nr_ret_type nebulas_rank_cache::get_nr_score(const nr_function_t &func,
                                             block_height_t start_block,
                                             block_height_t end_block,
                                             uint64_t version) {

  nr_ret_type result;
  std::string key = param_to_key(start_block, end_block, version);
  bool status = nebulas_rank_cache_instance::instance()
                    .get_cached_or_cal_if_not_or_ignore(key, result, func);
  if (!status) {
    result->set<p_result_status>(core::result_status::is_running);
  }
  return result;
}

nr_ret_type nebulas_rank_cache::get_nr_score(const std::string &handle) {
  nr_ret_type result;
  bool status = nebulas_rank_cache_instance::instance().get_cached_or_ignore(
      handle, result);
  if (!status) {
    result->set<p_result_status>(core::result_status::no_cached);
  }
  return result;
}
} // namespace nr

} // namespace rt
} // namespace neb
