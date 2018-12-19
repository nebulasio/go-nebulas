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
  inline const char *auth_func_mangling_name() const {
    return "_Z16entry_point_authB5cxx11v";
  }

  inline const char *nbre_failed_flag_name() const {
    return "nbre_failed_flag";
  }

  // nbre api config
  inline const char *nbre_ir_list_name() const { return "nbre_ir_list"; }
  inline const char *ir_list_name() const { return "ir_list"; }

protected:
  std::string m_nbre_root_dir;
  std::string m_nbre_exe_name;
  std::string m_neb_db_dir;
  std::string m_nbre_db_dir;
  std::string m_nbre_log_dir;
  std::string m_admin_pub_addr;
};
} // end namespace neb
