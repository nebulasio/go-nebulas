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
#include "common/common.h"
#include "core/neb_ipc/internal/ipc_pkg_internal.h"
#include "core/neb_ipc/ipc_common.h"

namespace neb {
namespace core {
enum {
  ipc_pkg_nbre_version_req,
  ipc_pkg_nbre_version_ack,
  ipc_pkg_nbre_init_req,
  ipc_pkg_nbre_init_ack,
  ipc_pkg_nbre_ir_list_req,
  ipc_pkg_nbre_ir_list_ack,
  ipc_pkg_nbre_versions_req,
  ipc_pkg_nbre_versions_ack,
  ipc_pkg_nbre_nr_handler_req,
  ipc_pkg_nbre_nr_handler_ack,
  ipc_pkg_nbre_nr_result_req,
  ipc_pkg_nbre_nr_result_ack,
  ipc_pkg_nbre_dip_reward_req,
  ipc_pkg_nbre_dip_reward_ack,
};
namespace ipc_pkg {
using namespace internal;

using height = ipc_elem_base<0, uint64_t>;
using nbre_version_req = define_ipc_pkg<ipc_pkg_nbre_version_req, height>;
using major = ipc_elem_base<1, uint32_t>;
using minor = ipc_elem_base<2, uint32_t>;
using patch = ipc_elem_base<3, uint32_t>;
using nbre_version_ack =
    define_ipc_pkg<ipc_pkg_nbre_version_ack, major, minor, patch>;

using nbre_init_req = define_ipc_pkg<ipc_pkg_nbre_init_req>;
using nbre_root_dir = ipc_elem_base<104, neb::ipc::char_string_t>;
using nbre_exe_name = ipc_elem_base<105, neb::ipc::char_string_t>;
using neb_db_dir = ipc_elem_base<106, neb::ipc::char_string_t>;
using nbre_db_dir = ipc_elem_base<107, neb::ipc::char_string_t>;
using nbre_log_dir = ipc_elem_base<108, neb::ipc::char_string_t>;
using admin_pub_addr = ipc_elem_base<109, neb::ipc::char_string_t>;
using nbre_start_height = ipc_elem_base<110, uint64_t>;
using nbre_init_ack =
    define_ipc_pkg<ipc_pkg_nbre_init_ack, nbre_root_dir, nbre_exe_name,
                   neb_db_dir, nbre_db_dir, nbre_log_dir, admin_pub_addr,
                   nbre_start_height>;

using nbre_ir_list_req = define_ipc_pkg<ipc_pkg_nbre_ir_list_req>;
using ir_name_list =
    ipc_elem_base<10, neb::ipc::vector<neb::ipc::char_string_t>>;
using nbre_ir_list_ack = define_ipc_pkg<ipc_pkg_nbre_ir_list_ack, ir_name_list>;

using ir_name = ipc_elem_base<11, neb::ipc::char_string_t>;
using nbre_ir_versions_req = define_ipc_pkg<ipc_pkg_nbre_versions_req, ir_name>;
using ir_versions = ipc_elem_base<12, neb::ipc::vector<uint64_t>>;
using nbre_ir_versions_ack =
    define_ipc_pkg<ipc_pkg_nbre_versions_ack, ir_name, ir_versions>;

using start_block = ipc_elem_base<13, uint64_t>;
using end_block = ipc_elem_base<14, uint64_t>;
using nr_version = ipc_elem_base<15, uint64_t>;
using nbre_nr_handler_req = define_ipc_pkg<ipc_pkg_nbre_nr_handler_req,
                                           start_block, end_block, nr_version>;
using nr_handler_id = ipc_elem_base<16, neb::ipc::char_string_t>;
using nbre_nr_handler_ack =
    define_ipc_pkg<ipc_pkg_nbre_nr_handler_ack, nr_handler_id>;

using nbre_nr_result_req =
    define_ipc_pkg<ipc_pkg_nbre_nr_result_req, nr_handler_id>;
using nr_result = ipc_elem_base<17, neb::ipc::char_string_t>;
using nbre_nr_result_ack =
    define_ipc_pkg<ipc_pkg_nbre_nr_result_ack, nr_result>;

using nbre_dip_reward_req = define_ipc_pkg<ipc_pkg_nbre_dip_reward_req, height>;
using dip_reward = ipc_elem_base<18, neb::ipc::char_string_t>;
using nbre_dip_reward_ack =
    define_ipc_pkg<ipc_pkg_nbre_dip_reward_ack, dip_reward>;

} // namespace ipc_pkg
} // namespace core
} // namespace neb
