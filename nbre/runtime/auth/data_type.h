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
namespace rt {
namespace auth {
typedef std::string name_t;
// typedef std::string address_t;
//! Note: it's dangerous to change this type, we may have compatible issues
//! here. You may check ir/auth_table/auth.cpp
typedef std::tuple<name_t, std::string, block_height_t, block_height_t>
    auth_items_t;
} // namespace auth
} // namespace rt
} // namespace neb
