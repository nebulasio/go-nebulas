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
  const std::string &dip_reward_addr() const { return m_dip_reward_addr; }
  std::string &dip_reward_addr() { return m_dip_reward_addr; }

protected:
  block_height_t m_dip_start_block;
  block_height_t m_dip_block_interval;
  std::string m_dip_reward_addr;
};
} // end namespace neb
