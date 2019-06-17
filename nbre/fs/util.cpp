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

#include "fs/util.h"
#include <boost/filesystem.hpp>
#include <chrono>
#include <ctime>

namespace neb {
namespace fs {

std::string cur_full_path() {
  return boost::filesystem::current_path().generic_string();
}

std::string cur_dir() {
  boost::filesystem::path cur_path = boost::filesystem::current_path();
  return cur_path.parent_path().generic_string();
}

std::string tmp_dir() {
  return boost::filesystem::temp_directory_path().generic_string();
}

std::string join_path(const std::string &parent, const std::string &fp) {
  boost::filesystem::path cur_path(parent);
  boost::filesystem::path np = cur_path / boost::filesystem::path(fp);
  return np.generic_string();
}

std::string parent_dir(const std::string &fp) {
  boost::filesystem::path cur_path(fp);
  return cur_path.parent_path().generic_string();
}

bool is_absolute_path(const std::string &fp) {
  boost::filesystem::path cur_path(fp);
  return cur_path.is_absolute();
}

bool exists(const std::string &p) {
  return boost::filesystem::exists(boost::filesystem::path(p));
}

std::string get_user_name() { return std::string("usr"); }

} // end namespace fs

std::string now() {
  auto nt = std::chrono::system_clock::now();
  std::time_t tt = std::chrono::system_clock::to_time_t(nt);
  auto ret = std::string(std::ctime(&tt));
  ret.erase(std::remove_if(ret.begin(), ret.end(),
                           [](unsigned char c) { return c == '\n'; }));
  return ret;
}

bytes concate_name_version(const std::string &name, version_t v) {
  return number_to_byte<bytes>(v) + name;
}

} // end namespace neb
