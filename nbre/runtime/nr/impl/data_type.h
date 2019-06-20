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
#include "core/net_ipc/nipc_pkg.h"

namespace neb {
namespace fs {
class transaction_db_interface;
class account_db_interface;
} // namespace fs
namespace rt {
namespace nr {

typedef ::ff::net::ntpackage<1, p_start_block, p_block_interval, p_version>
    nr_param_t;

struct rank_params_t {
  int64_t m_a;
  int64_t m_b;
  int64_t m_c;
  int64_t m_d;
  floatxx_t m_theta;
  floatxx_t m_mu;
  floatxx_t m_lambda;
};

using nr_ret_type = std::shared_ptr<nr_result>;
} // namespace nr
} // namespace rt
} // namespace neb
