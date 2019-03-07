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
#include "fs/blockchain/blockchain_api_test.h"
#include "runtime/nr/impl/nebulas_rank.h"
#include "util/nebulas_currency.h"

namespace neb {
namespace rt {
namespace nr {

std::vector<std::shared_ptr<nr_info_t>>
entry_point_nr_impl(compatible_uint64_t start_block,
                    compatible_uint64_t end_block, version_t version,
                    compatible_int64_t a, compatible_int64_t b,
                    compatible_int64_t c, compatible_int64_t d,
                    nr_float_t theta, nr_float_t mu, nr_float_t lambda) {

  std::unique_ptr<neb::fs::blockchain_api_base> pba;
  if (neb::use_test_blockchain) {
    pba = std::unique_ptr<neb::fs::blockchain_api_base>(
        new neb::fs::blockchain_api_test());
  } else {
    pba = std::unique_ptr<neb::fs::blockchain_api_base>(
        new neb::fs::blockchain_api());
  }
  transaction_db_ptr_t tdb_ptr =
      std::make_unique<neb::fs::transaction_db>(pba.get());
  account_db_ptr_t adb_ptr = std::make_unique<neb::fs::account_db>(pba.get());

  LOG(INFO) << "start block: " << start_block << " , end block: " << end_block;
  neb::rt::nr::rank_params_t rp{a, b, c, d, theta, mu, lambda};
  std::vector<std::pair<std::string, uint64_t>> meta_info(
      {{"start_height", start_block},
       {"end_height", end_block},
       {"version", version}});

  return nebulas_rank::get_nr_score(tdb_ptr, adb_ptr, rp, start_block,
                                    end_block);
}
} // namespace nr
} // namespace rt
} // namespace neb

