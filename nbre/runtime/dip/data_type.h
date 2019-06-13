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
#include "runtime/nr/impl/data_type.h"


namespace neb {
namespace rt {
namespace dip {

typedef ::ff::net::ntpackage<1, p_start_block, p_block_interval,
                             p_dip_reward_addr, p_dip_coinbase_addr, p_version>
    dip_param_t;

/*
struct dip_info_t {
address_t m_deployer;
address_t m_contract;
std::string m_reward;
};
*/

using dip_ret_type = std::shared_ptr<dip_result>;
} // namespace dip
} // namespace rt
} // namespace neb
