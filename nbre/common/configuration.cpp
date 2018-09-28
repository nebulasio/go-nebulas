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
#include <boost/foreach.hpp>
#include <string>
#include <set>
#include <exception>
#include <iostream>

namespace neb {
  namespace pt = boost::property_tree;
  configuration::configuration() { 
    try {
      pt::ptree ini_root;
      std::string ini_file_path = "jit_configuration.ini";

      pt::ini_parser::read_ini(ini_file_path, ini_root);
      m_exec_name = ini_root.get<std::string>("jit_config.exec_name"); 
      m_runtime_library_path = ini_root.get<std::string>("jit_config.runtime_library_path");
    }
    catch (const pt::ptree_error &e) {
      throw configure_general_failure(e.what());
    }
  }

  configuration::~configuration() = default;

  void configuration::init_with_args(int argc, char *argv[]) {

  }
}

