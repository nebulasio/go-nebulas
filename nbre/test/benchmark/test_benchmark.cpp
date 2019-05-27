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

#include "common/version.h"
#include "fs/blockchain/account/account_db_v2.h"
#include "fs/blockchain/blockchain_api_v2.h"
#include "fs/blockchain/transaction/transaction_db_v2.h"
#include "runtime/dip/dip_impl.h"
#include "runtime/dip/dip_reward_v2.h"
#include "runtime/nr/impl/nebulas_rank_v2.h"
#include "runtime/nr/impl/nr_impl.h"
#include <boost/program_options.hpp>

namespace po = boost::program_options;

neb::rt::nr::nr_ret_type entry_point_nr_impl(
    uint64_t start_block, uint64_t end_block, neb::rt::nr::version_t version,
    int64_t a, int64_t b, int64_t c, int64_t d, neb::rt::nr::nr_float_t theta,
    neb::rt::nr::nr_float_t mu, neb::rt::nr::nr_float_t lambda) {

  std::unique_ptr<neb::fs::blockchain_api_base> pba =
      std::unique_ptr<neb::fs::blockchain_api_base>(
          new neb::fs::blockchain_api_v2());
  neb::rt::nr::transaction_db_ptr_t tdb_ptr =
      std::make_unique<neb::fs::transaction_db>(pba.get());
  neb::rt::nr::account_db_ptr_t adb_ptr =
      std::make_unique<neb::fs::account_db>(pba.get());

  auto tdb_ptr_v2 = std::make_unique<neb::fs::transaction_db_v2>(tdb_ptr.get());
  auto adb_ptr_v2 = std::make_unique<neb::fs::account_db_v2>(adb_ptr.get());

  LOG(INFO) << "start block: " << start_block << " , end block: " << end_block;
  neb::rt::nr::rank_params_t rp{a, b, c, d, theta, mu, lambda};

  std::vector<std::pair<std::string, std::string>> meta_info;
  meta_info.push_back(
      std::make_pair("start_height", std::to_string(start_block)));
  meta_info.push_back(std::make_pair("end_height", std::to_string(end_block)));
  meta_info.push_back(std::make_pair("version", std::to_string(version)));

  neb::rt::nr::nr_ret_type ret;
  std::get<0>(ret) = 1;
  std::get<1>(ret) = neb::rt::meta_info_to_json(meta_info);

  // neb::rt::nr::nebulas_rank::get_nr_score(tdb_ptr, adb_ptr, rp, start_block,
  // end_block);
  std::get<2>(ret) = neb::rt::nr::nebulas_rank_v2::get_nr_score(
      tdb_ptr_v2, adb_ptr_v2, rp, start_block, end_block);

  std::unordered_set<neb::address_t> diff_set;
  for (auto h = start_block; h < end_block; h++) {
    for (auto &ele : std::get<2>(ret)) {
      auto addr = ele->m_address;
      auto b1 = adb_ptr_v2->get_balance(addr, h);
      auto b2 = adb_ptr_v2->get_account_balance_internal(addr, h);
      std::cout << h << ',' << addr.to_base58() << ',' << b1 << ',' << b2
                << std::endl;
      if (b1 != b2) {
        diff_set.insert(addr);
      }
    }
  }
  LOG(INFO) << "diff set size " << diff_set.size();
  return ret;
}

neb::rt::dip::dip_ret_type entry_point_dip_impl(
    uint64_t start_block, uint64_t end_block, neb::rt::dip::version_t version,
    uint64_t height, const neb::rt::nr::nr_ret_type &nr_ret,
    neb::rt::dip::dip_float_t alpha, neb::rt::dip::dip_float_t beta) {

  std::unique_ptr<neb::fs::blockchain_api_base> pba =
      std::unique_ptr<neb::fs::blockchain_api_base>(
          new neb::fs::blockchain_api_v2());
  neb::rt::nr::transaction_db_ptr_t tdb_ptr =
      std::make_unique<neb::fs::transaction_db>(pba.get());
  neb::rt::nr::account_db_ptr_t adb_ptr =
      std::make_unique<neb::fs::account_db>(pba.get());

  auto tdb_ptr_v2 = std::make_unique<neb::fs::transaction_db_v2>(tdb_ptr.get());
  auto adb_ptr_v2 = std::make_unique<neb::fs::account_db_v2>(adb_ptr.get());

  std::vector<std::pair<std::string, std::string>> meta_info;
  meta_info.push_back(
      std::make_pair("start_height", std::to_string(start_block)));
  meta_info.push_back(std::make_pair("end_height", std::to_string(end_block)));
  meta_info.push_back(std::make_pair("version", std::to_string(version)));

  neb::rt::dip::dip_ret_type ret;
  std::get<0>(ret) = 1;
  std::get<1>(ret) = neb::rt::meta_info_to_json(meta_info);

  auto &nr_result = std::get<2>(nr_ret);
  std::get<2>(ret) = neb::rt::dip::dip_reward_v2::get_dip_reward(
      start_block, end_block, height, nr_result, tdb_ptr_v2, adb_ptr_v2, alpha,
      beta);
  LOG(INFO) << "get dip reward resurned";

  std::get<3>(ret) = nr_ret;
  LOG(INFO) << "append nr_ret to dip_ret";

  return ret;
}

int main(int argc, char *argv[]) {

  po::options_description desc("Nr");
  desc.add_options()("help", "show help message")(
      "start_block", po::value<uint64_t>(), "Start block height")(
      "end_block", po::value<uint64_t>(), "End block height")(
      "address", po::value<std::string>(),
      "base58 address")("db_path", po::value<std::string>(), "neb db path");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);

  if (vm.count("help")) {
    std::cout << desc << "\n";
    return 1;
  }

  if (!vm.count("start_block")) {
    std::cout << "You must specify \"start_block\"!" << std::endl;
    return 1;
  }
  if (!vm.count("end_block")) {
    std::cout << "You must specify \"end_block\"!" << std::endl;
    return 1;
  }
  if (!vm.count("db_path")) {
    std::cout << "You must specify \"db_path\"!" << std::endl;
    return 1;
  }

  neb::compatible_int64_t a = 100;
  neb::compatible_int64_t b = 2;
  neb::compatible_int64_t c = 6;
  neb::compatible_int64_t d = -9;
  neb::rt::nr::nr_float_t theta = 1;
  neb::rt::nr::nr_float_t mu = 1;
  neb::rt::nr::nr_float_t lambda = 2;

  neb::rt::dip::dip_float_t alpha = 8e-3;
  neb::rt::dip::dip_float_t beta = 1;

  uint64_t start_block = vm["start_block"].as<uint64_t>();
  uint64_t end_block = vm["end_block"].as<uint64_t>();
  std::string neb_path = vm["db_path"].as<std::string>();

  neb::fs::bc_storage_session::instance().init(neb_path,
                                               neb::fs::storage_open_default);

  auto nr_ret = entry_point_nr_impl(start_block, end_block, 0, a, b, c, d,
                                    theta, mu, lambda);
  auto dip_ret =
      entry_point_dip_impl(start_block, end_block, 0, 0, nr_ret, alpha, beta);

  neb::fs::bc_storage_session::instance().release();
  return 0;
}
