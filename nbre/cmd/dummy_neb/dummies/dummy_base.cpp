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
#include "fs/blockchain.h"
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
  try {
    auto block = neb::fs::blockchain::load_LIB_block();
    m_current_height = block->height();
    m_current_height++;
  } catch (const std::exception &e) {
    LOG(INFO) << "init from empty db";
  }
  LOG(INFO) << "init block lib height is " << m_current_height;
}

void dummy_base::clean_db() {
  LOG(INFO) << "clean db";
  std::stringstream ss;
  ss << "rm -rf " << db_path();
  neb::util::command_executor::execute_command(ss.str());
  ss.clear();
  ss << "rm -rf " << neb::configuration::instance().nbre_db_dir();
  neb::util::command_executor::execute_command(ss.str());
}

void dummy_base::random_increase_version(neb::version &v) {
  int k = std::rand() % 3;
  switch (k) {
  case 0:
    v.major_version() += 1;
    break;
  case 1:
    v.minor_version() += 1;
    break;
  case 2:
    v.patch_version() += 1;
    break;
  }
}
