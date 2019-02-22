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
// clang-format off
define_ipc_param(uint64_t, p_holder)
define_ipc_param(uint64_t, p_height)
define_ipc_param(uint32_t, p_major)
define_ipc_param(uint32_t, p_minor)
define_ipc_param(uint32_t, p_patch)
define_ipc_param(std::string, p_nbre_root_dir)
define_ipc_param(std::string, p_nbre_exe_name)
define_ipc_param(std::string, p_neb_db_dir)
define_ipc_param(std::string, p_nbre_log_dir)
define_ipc_param(std::string, p_admin_pub_addr)
define_ipc_param(std::string, p_nbre_db_dir)
define_ipc_param(uint64_t, p_nbre_start_height)

define_ipc_pkg(nbre_version_req, p_holder, p_height)
define_ipc_pkg(nbre_version_ack, p_holder, p_major, p_minor, p_patch)
define_ipc_pkg(nbre_init_req, p_holder)
define_ipc_pkg(nbre_init_ack, p_nbre_root_dir, p_nbre_exe_name, p_neb_db_dir,
               p_nbre_log_dir, p_nbre_db_dir, p_admin_pub_addr,
               p_nbre_start_height)

// define_ipc_api(server_send_pkg, client_send_pkg)
define_ipc_api(nbre_version_req, nbre_version_ack)
define_ipc_api(nbre_init_ack, nbre_init_req)

    // clang-format on
