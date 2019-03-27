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

namespace neb {
namespace core {

class ipc_configuration : public util::singleton<ipc_configuration> {
public:
  ipc_configuration();
  ipc_configuration(const ipc_configuration &cf) = delete;
  ipc_configuration &operator=(const ipc_configuration &cf) = delete;
  ipc_configuration(ipc_configuration &&cf) = delete;
  ~ipc_configuration();

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
  const std::string &admin_pub_addr() const { return m_admin_pub_addr; }
  std::string &admin_pub_addr() { return m_admin_pub_addr; }

  // nbre start height
  inline const uint64_t &nbre_start_height() const {
    return m_nbre_start_height;
  }
  inline uint64_t &nbre_start_height() { return m_nbre_start_height; }

protected:
  std::string m_nbre_root_dir;
  std::string m_nbre_exe_name;
  std::string m_neb_db_dir;
  std::string m_nbre_db_dir;
  std::string m_nbre_log_dir;
  std::string m_admin_pub_addr;
  uint64_t m_nbre_start_height;
};
} // namespace core
} // namespace neb
