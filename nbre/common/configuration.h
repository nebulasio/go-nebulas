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
#include "common/address.h"
#include "common/common.h"
#include "util/singleton.h"

namespace neb {

struct configure_general_failure : public std::exception {
  inline configure_general_failure(const std::string &msg) : m_msg(msg) {}
  inline const char *what() const throw() { return m_msg.c_str(); }
protected:
  std::string m_msg;
};

class configuration : public util::singleton<configuration> {
public:
  configuration();
  configuration(const configuration &cf) = delete;
  configuration &operator=(const configuration &cf) = delete;
  configuration(configuration &&cf) = delete;
  ~configuration();

  inline int32_t ir_warden_time_interval() const { return 1; }

  // nbre storage ir config
  inline const char *ir_tx_payload_type() const { return "protocol"; }
  inline const char *nbre_max_height_name() const { return "nbre_max_height"; }
  inline const char *rt_module_name() const { return "runtime"; }

  inline const char *nbre_auth_table_name() const { return "nbre_auth_table"; }
  inline const char *auth_module_name() const { return "auth"; }
  inline const char *auth_func_name() const { return "entry_point_auth"; }

  inline const char *nbre_failed_flag_name() const {
    return "nbre_failed_flag";
  }

  inline const char *nr_func_name() const { return "entry_point_nr"; }
  inline const char *dip_func_name() const { return "entry_point_dip"; }

  // nbre api config
  inline const char *ir_list_name() const { return "ir_list"; }

  // dip conf
  inline const block_height_t &dip_start_block() const {
    return m_dip_start_block;
  }
  inline block_height_t &dip_start_block() { return m_dip_start_block; }
  inline const block_height_t &dip_block_interval() const {
    return m_dip_block_interval;
  }
  inline block_height_t &dip_block_interval() { return m_dip_block_interval; }

  // dip reward address
  const address_t &dip_reward_addr() const { return m_dip_reward_addr; }
  address_t &dip_reward_addr() { return m_dip_reward_addr; }

  // coinbase address
  const address_t &coinbase_addr() const { return m_coinbase_addr; }
  address_t &coinbase_addr() { return m_coinbase_addr; }

  // shared memory name identity
  inline const std::string &shm_name_identity() const {
    return m_shm_name_identity;
  }
  inline std::string &shm_name_identity() { return m_shm_name_identity; }

  // nbre root directory
  inline const std::string &nbre_root_dir() const { return m_nbre_root_dir; }
  inline std::string &nbre_root_dir() { return m_nbre_root_dir; }

  // nbre execute path
  inline const std::string &nbre_exe_name() const { return m_nbre_exe_name; }
  inline std::string &nbre_exe_name() { return m_nbre_exe_name; }

  // nebulas blockchain database directory
  inline const std::string &neb_db_dir() const { return m_neb_db_dir; }
  inline std::string &neb_db_dir() { return m_neb_db_dir; }

  // nbre database directory
  inline const std::string &nbre_db_dir() const { return m_nbre_db_dir; }
  inline std::string &nbre_db_dir() { return m_nbre_db_dir; }

  // nbre log directory
  inline const std::string &nbre_log_dir() const { return m_nbre_log_dir; }
  inline std::string &nbre_log_dir() { return m_nbre_log_dir; }

  // nbre storage auth table admin address
  inline const address_t &admin_pub_addr() const { return m_admin_pub_addr; }
  inline address_t &admin_pub_addr() { return m_admin_pub_addr; }

  // nbre start height
  inline const uint64_t &nbre_start_height() const {
    return m_nbre_start_height;
  }
  inline uint64_t &nbre_start_height() { return m_nbre_start_height; }

  // nbre net ipc listen
  inline const std::string &nipc_listen() const { return m_nipc_listen; }
  inline std::string &nipc_listen() { return m_nipc_listen; }

  // nbre net ipc port
  inline const uint16_t &nipc_port() const { return m_nipc_port; }
  inline uint16_t &nipc_port() { return m_nipc_port; }

  inline const std::string &get_exit_msg(uint32_t exit_code) const {
    assert(exit_code < m_exit_msg_list.size());
    return m_exit_msg_list[exit_code];
  };
  inline void set_exit_msg(const std::string &exit_msg) {
    m_exit_msg_list.push_back(exit_msg);
  }

protected:
  // dip conf
  block_height_t m_dip_start_block;
  block_height_t m_dip_block_interval;
  address_t m_dip_reward_addr;
  address_t m_coinbase_addr;

  // shm conf
  std::string m_shm_name_identity;

  // nbre init params conf
  std::string m_nbre_root_dir;
  std::string m_nbre_exe_name;
  std::string m_neb_db_dir;
  std::string m_nbre_db_dir;
  std::string m_nbre_log_dir;
  address_t m_admin_pub_addr;
  uint64_t m_nbre_start_height;
  std::string m_nipc_listen;
  uint16_t m_nipc_port;

  // exit code and msg list conf
  std::vector<std::string> m_exit_msg_list;
};

extern bool use_test_blockchain;
extern bool glog_log_to_stderr;
} // end namespace neb
