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
#include "common/int128_conversion.h"
#include "common/nebulas_currency.h"
#include "core/execution_context.h"
#include "fs/blockchain/account/account_db.h"
#include "fs/blockchain/blockchain_api_test.h"
#include "fs/blockchain/transaction/transaction_db.h"
#include "runtime/nr/graph/graph_algo.h"
#include "runtime/nr/impl/general_nebulas_rank.h"
#include "runtime/nr/impl/nebulas_rank_algo.h"
#include "runtime/nr/impl/nebulas_rank_cache.h"
#include "runtime/nr/impl/nebulas_rank_calculator.h"

namespace neb {
namespace rt {
namespace nr {

nr_ret_type entry_point_nr_impl(compatible_uint64_t start_block,
                                compatible_uint64_t end_block,
                                version_t version, compatible_int64_t a,
                                compatible_int64_t b, compatible_int64_t c,
                                compatible_int64_t d, nr_float_t theta,
                                nr_float_t mu, nr_float_t lambda) {

  std::unique_ptr<neb::fs::blockchain_api_base> pba;
  if (neb::use_test_blockchain) {
    pba = std::unique_ptr<neb::fs::blockchain_api_base>(
        new neb::fs::blockchain_api_test());
  } else {
    pba = std::unique_ptr<neb::fs::blockchain_api_base>(
        new neb::fs::blockchain_api(core::context->blockchain()));
  }
  auto tdb_ptr = std::make_unique<neb::fs::transaction_db>(pba.get());
  auto adb_ptr = std::make_unique<neb::fs::account_db>(pba.get());

  auto ga_ptr = std::make_unique<graph_algo>();
  auto nra_ptr = std::make_unique<nebulas_rank_algo>();
  auto nrc_ptr = std::make_unique<nebulas_rank_calculator>(
      ga_ptr.get(), nra_ptr.get(), tdb_ptr.get(), adb_ptr.get());

  auto gnr_ptr =
      std::make_unique<general_nebulas_rank>(nrc_ptr.get(), nr_cache.get());
  rank_params_t rp;
  rp.m_a = a;
  rp.m_b = b;
  rp.m_c = c;
  rp.m_d = d;
  rp.m_theta = theta;
  rp.m_mu = mu;
  rp.m_lambda = lambda;
  return gnr_ptr->get_nr_score(rp, start_block, end_block, version);
}

} // namespace nr
} // namespace rt
} // namespace neb

