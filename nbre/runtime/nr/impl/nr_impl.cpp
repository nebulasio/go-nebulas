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

#include "runtime/nr/impl/nr_impl.h"
#include "common/common.h"
#include "common/configuration.h"
#include "common/util/conversion.h"
#include "fs/blockchain/nebulas_currency.h"
#include "runtime/nr/impl/nebulas_rank.h"

#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>
#include <chrono>
#include <thread>

namespace neb {
namespace rt {
namespace nr {

template <typename T> std::string to_string(const T &val) {
  std::stringstream ss;
  ss << val;
  return ss.str();
}

void convert_nr_info_to_ptree(const neb::rt::nr::nr_info_t &info,
                              boost::property_tree::ptree &p) {

  neb::util::bytes addr_bytes = neb::util::string_to_byte(info.m_address);

  uint32_t in_degree = info.m_in_degree;
  uint32_t out_degree = info.m_out_degree;
  uint32_t degrees = info.m_degrees;

  neb::floatxx_t in_val = info.m_in_val;
  neb::floatxx_t out_val = info.m_out_val;
  neb::floatxx_t in_outs = info.m_in_outs;

  neb::floatxx_t median = info.m_median;
  neb::floatxx_t weight = info.m_weight;
  neb::floatxx_t score = info.m_nr_score;

  std::vector<std::pair<std::string, std::string>> kv_pair(
      {{"address", addr_bytes.to_base58()},
       {"in_degree", to_string(in_degree)},
       {"out_degree", to_string(out_degree)},
       {"degrees", to_string(degrees)},
       {"in_val", to_string(in_val)},
       {"out_val", to_string(out_val)},
       {"in_outs", to_string(in_outs)},
       {"median", to_string(median)},
       {"weight", to_string(weight)},
       {"score", to_string(score)}});

  for (auto &ele : kv_pair) {
    p.put(ele.first, ele.second);
  }
}

std::string to_json(const std::vector<neb::rt::nr::nr_info_t> &rs) {
  boost::property_tree::ptree root;
  boost::property_tree::ptree arr;

  for (auto it = rs.begin(); it != rs.end(); it++) {
    const neb::rt::nr::nr_info_t &info = *it;
    boost::property_tree::ptree p;
    convert_nr_info_to_ptree(info, p);
    arr.push_back(std::make_pair(std::string(), p));
  }
  root.add_child("nrs", arr);

  std::stringstream ss;
  boost::property_tree::json_parser::write_json(ss, root, false);
  return ss.str();
}

std::string entry_point_nr_impl(uint64_t start_block, uint64_t end_block,
                                nr_float_t a, nr_float_t b, nr_float_t c,
                                nr_float_t d, int64_t mu, int64_t lambda) {

  std::string neb_db_path = neb::configuration::instance().neb_db_dir();
  neb::fs::blockchain bc(neb_db_path);
  neb::fs::blockchain_api ba(&bc);
  neb::rt::nr::transaction_db_ptr_t tdb_ptr =
      std::make_shared<neb::fs::transaction_db>(&ba);
  neb::rt::nr::account_db_ptr_t adb_ptr =
      std::make_shared<neb::fs::account_db>(&ba);

  LOG(INFO) << "start block: " << start_block << " , end block: " << end_block;
  neb::rt::nr::rank_params_t rp{a, b, c, d, mu, lambda};

  auto it_txs =
      tdb_ptr->read_transactions_from_db_with_duration(start_block, end_block);

  auto ret = neb::rt::nr::nebulas_rank::get_nr_score(
      tdb_ptr, adb_ptr, *it_txs, rp, start_block, end_block);
  return to_json(*ret);
}
} // namespace nr
} // namespace rt
} // namespace neb

