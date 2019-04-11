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
#include "common/util/conversion.h"
#include "core/neb_ipc/server/ipc_configuration.h"
#include "fs/blockchain/nebulas_currency.h"
#include "runtime/nr/impl/nebulas_rank.h"

namespace neb {
namespace rt {
namespace nr {

std::string entry_point_nr_impl(compatible_uint64_t start_block,
                                compatible_uint64_t end_block,
                                version_t version, compatible_int64_t a,
                                compatible_int64_t b, compatible_int64_t c,
                                compatible_int64_t d, nr_float_t theta,
                                nr_float_t mu, nr_float_t lambda) {

  std::string neb_db_path =
      neb::core::ipc_configuration::instance().neb_db_dir();
  neb::fs::blockchain bc(neb_db_path);
  neb::fs::blockchain_api ba(&bc);
  transaction_db_ptr_t tdb_ptr = std::make_unique<neb::fs::transaction_db>(&ba);
  account_db_ptr_t adb_ptr = std::make_unique<neb::fs::account_db>(&ba);

  LOG(INFO) << "start block: " << start_block << " , end block: " << end_block;
  neb::rt::nr::rank_params_t rp{a, b, c, d, theta, mu, lambda};
  std::vector<std::pair<std::string, uint64_t>> meta_info(
      {{"start_height", start_block},
       {"end_height", end_block},
       {"version", version}});

  auto ret =
      nebulas_rank::get_nr_score(tdb_ptr, adb_ptr, rp, start_block, end_block);
  return nebulas_rank::nr_info_to_json(*ret, meta_info);
}
} // namespace nr
} // namespace rt
} // namespace neb

