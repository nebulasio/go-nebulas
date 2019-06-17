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
#include "common/byte.h"
#include "common/common.h"

namespace neb {
namespace fs {

std::string cur_full_path();
std::string cur_dir();
std::string tmp_dir();
std::string join_path(const std::string &parent, const std::string &fp);
std::string parent_dir(const std::string &fp);
bool is_absolute_path(const std::string &fp);
bool exists(const std::string &p);
std::string get_user_name();

} // end namespace fs

std::string now();

bytes concate_name_version(const std::string &name, version_t v);

} // end namespace neb

