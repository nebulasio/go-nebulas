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
#include "cmd/dummy_neb/dummy_common.h"
#include "core/net_ipc/nipc_pkg.h"
#include "util/controller.h"
#include <ff/network.h>

enum {
  cli_brief_req_pkg = ctl_last_pkg_id,
  cli_brief_ack_pkg,
  cli_submit_ir_pkg,
  cli_submit_ack_pkg,
};

define_nt(p_account_num, uint64_t);
define_nt(p_nr_ir_status, std::vector<std::string>);
define_nt(p_dip_ir_status, std::vector<std::string>);
define_nt(p_auth_ir_status, std::vector<std::string>);
define_nt(p_checker_status, std::string);

typedef ff::net::ntpackage<cli_brief_req_pkg> cli_brief_req_t;
typedef ff::net::ntpackage<cli_brief_ack_pkg, p_height, p_account_num,
                           p_checker_status>
    cli_brief_ack_t;

define_nt(p_type, std::string);
define_nt(p_payload, std::string);
define_nt(p_result, std::string);
typedef ff::net::ntpackage<cli_submit_ir_pkg, p_type, p_payload>
    cli_submit_ir_t;

typedef ff::net::ntpackage<cli_submit_ack_pkg, p_result> cli_submit_ack_t;

