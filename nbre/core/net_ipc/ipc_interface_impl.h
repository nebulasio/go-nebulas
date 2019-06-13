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
//define_ipc_api(nbre_init_ack, nbre_init_req)
define_ipc_api(nbre_version_req, nbre_version_ack)

define_ipc_param(std::string, p_ir_name_list)
define_ipc_pkg(nbre_ir_list_req, p_holder)
define_ipc_pkg(nbre_ir_list_ack, p_holder, p_ir_name_list)
define_ipc_api(nbre_ir_list_req, nbre_ir_list_ack)

define_ipc_param(std::string, p_ir_name)
define_ipc_param(std::string, p_ir_versions)
define_ipc_pkg(nbre_ir_versions_req, p_holder, p_ir_name)
define_ipc_pkg(nbre_ir_versions_ack, p_holder, p_ir_versions)
define_ipc_api(nbre_ir_versions_req, nbre_ir_versions_ack)

define_ipc_param(uint64_t, p_start_block)
define_ipc_param(uint64_t, p_end_block)
define_ipc_param(uint64_t, p_nr_version)
define_ipc_param(std::string, p_nr_handle)
define_ipc_pkg(nbre_nr_handle_req, p_holder, p_start_block, p_end_block, p_nr_version)
define_ipc_pkg(nbre_nr_handle_ack, p_holder, p_nr_handle)
define_ipc_api(nbre_nr_handle_req, nbre_nr_handle_ack)

define_ipc_param(std::string, p_nr_result)
define_ipc_param(std::string, p_nr_item_addr)
define_ipc_param(neb::floatxx_t, p_nr_item_in_outs)
define_ipc_param(neb::floatxx_t, p_nr_item_median)
define_ipc_param(neb::floatxx_t, p_nr_item_weight)
define_ipc_param(neb::floatxx_t, p_nr_item_score)
define_ipc_param(uint32_t, p_result_status)
define_ipc_pkg(nr_item, p_nr_item_addr, p_nr_item_in_outs, p_nr_item_weight, p_nr_item_median, p_nr_item_score)
define_ipc_param(std::vector<nr_item>, p_nr_items)
define_ipc_pkg(nr_result, p_start_block, p_end_block, p_nr_version, p_result_status, p_nr_items)
define_ipc_pkg(nbre_nr_result_by_handle_req, p_holder, p_nr_handle)
define_ipc_pkg(nbre_nr_result_by_handle_ack, p_holder, p_nr_result)
define_ipc_api(nbre_nr_result_by_handle_req, nbre_nr_result_by_handle_ack)

define_ipc_pkg(nbre_nr_result_by_height_req, p_holder, p_height)
define_ipc_pkg(nbre_nr_result_by_height_ack, p_holder, p_nr_result)
define_ipc_api(nbre_nr_result_by_height_req, nbre_nr_result_by_height_ack)

define_ipc_param(std::string, p_nr_sum)
define_ipc_pkg(nbre_nr_sum_req, p_holder, p_height)
define_ipc_pkg(nbre_nr_sum_ack, p_holder, p_nr_sum)
define_ipc_api(nbre_nr_sum_req, nbre_nr_sum_ack)

define_ipc_param(uint64_t, p_version)
define_ipc_param(uint64_t, p_block_interval)
define_ipc_param(std::string, p_dip_reward_addr)
define_ipc_param(std::string, p_dip_coinbase_addr)
define_ipc_param(std::string, p_dip_deployer)
define_ipc_param(std::string, p_dip_contract)
define_ipc_param(neb::floatxx_t, p_dip_reward_value)
define_ipc_pkg(dip_item, p_dip_deployer, p_dip_contract, p_dip_reward_value)
define_ipc_param(std::vector<dip_item>, p_dip_items)
define_ipc_pkg(dip_result, p_start_block, p_end_block, p_version, p_result_status, p_dip_items)
define_ipc_param(std::string, p_dip_reward)
define_ipc_pkg(nbre_dip_reward_req, p_holder, p_height, p_version)
define_ipc_pkg(nbre_dip_reward_ack, p_holder, p_dip_reward)
define_ipc_api(nbre_dip_reward_req, nbre_dip_reward_ack)

define_ipc_param(std::vector<std::string>, p_ir_transactions)
define_ipc_pkg(nbre_ir_transactions_req, p_holder, p_height, p_ir_transactions)
define_ipc_param(std::string, p_raw_key)
define_ipc_param(std::string, p_ir_key)
define_ipc_pkg(nr_param_storage_t, p_start_block, p_block_interval, p_version, p_raw_key, p_ir_key)
define_ipc_pkg(dip_param_storage_t, p_start_block, p_block_interval, p_dip_reward_addr, p_dip_coinbase_addr, p_version, p_raw_key, p_ir_key)
define_ipc_pkg(auth_param_storage_t, p_start_block, p_version, p_raw_key, p_ir_key)
    // clang-format on
