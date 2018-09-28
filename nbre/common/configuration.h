// Copyright (C) 2018 go-nebulas authors
//
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

  void init_with_args(int argc, char *argv[]);

  const std::string &exec_name() const { return m_exec_name; } 
  const std::string &runtime_library_path() const { return m_runtime_library_path; }
protected:
  std::string m_exec_name;
  std::string m_runtime_library_path;
};
} // end namespace neb
