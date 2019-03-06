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
#include "cmd/dummy_neb/dummies/dummy_base.h"
#include "common/configuration.h"
#include "fs/util.h"
#include "util/command.h"

dummy_base::dummy_base(const std::string &name)
    : m_name(name), m_current_height(0) {}

dummy_base::~dummy_base() {}
const std::string &dummy_base::db_path() const {
  if (m_db_path.empty()) {
    auto dir = neb::configuration::instance().neb_db_dir();
    m_db_path = neb::fs::join_path(dir, std::string("dummy_") + m_name +
                                            std::string("_.db"));
    LOG(INFO) << "neb deb path: " << m_db_path;
  }
  return m_db_path;
}

void dummy_base::init_from_db() {
  bc_storage_session::instance().init(db_path(),
                                      neb::fs::storage_open_for_readwrite);
}

void dummy_base::clean_db() {
  std::stringstream ss;
  ss << "rm -rf " << db_path();
  neb::command_executor::execute_command(ss.str());
}
