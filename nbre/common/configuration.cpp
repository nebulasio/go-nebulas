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
#include "common/configuration.h"

#include <boost/property_tree/ptree.hpp>
#include <boost/property_tree/ini_parser.hpp>
#include <boost/format.hpp>
#include <boost/program_options.hpp>
#include <string>
#include <set>
#include <exception>
#include <iostream>

namespace neb {
namespace pt = boost::property_tree;
namespace po = boost::program_options;

#define KTS(v) #v
#define STR(v) KTS(v)
configuration::configuration() {
#ifdef NDEBUG
  // supervisor start failed with getenv
#else
  m_nbre_root_dir = std::getenv("NBRE_ROOT_DIR");
  m_nbre_exe_name = std::getenv("NBRE_EXE_NAME");
  m_neb_db_dir = std::getenv("NEB_DB_DIR");
  m_nbre_db_dir = std::getenv("NBRE_DB_DIR");
  m_nbre_log_dir = std::getenv("NBRE_LOG_DIR");
#endif
}
#undef STR
#undef KTS

configuration::~configuration() = default;

bool use_test_blockchain = false;
bool glog_log_to_stderr = true;
} // namespace neb

