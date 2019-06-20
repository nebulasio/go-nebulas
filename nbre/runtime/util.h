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

// using str_sptr_t = std::shared_ptr<std::string>;
// using str_uptr_t = std::unique_ptr<std::string>;

/*
std::string meta_info_to_json(
    const std::vector<std::pair<std::string, std::string>> &meta_info);

std::vector<std::pair<std::string, std::string>>
json_to_meta_info(const std::string &json);
*/
std::string param_to_key(block_height_t start_block, block_height_t end_block,
                         uint64_t version);
} // namespace rt
} // namespace neb
