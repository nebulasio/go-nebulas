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
#include "compatible/compatible_checker.h"
#include "compatible/compatible_check_base.h"
#include "util/singleton.h"

namespace neb {
namespace compatible {

compatible_checker::~compatible_checker() {}

void compatible_checker::init() {}
bool compatible_checker::is_ir_need_compile(const std::string &name,
                                            version_t version) {
  return false;
}

optional<rt::nr::nr_ret_type>
compatible_checker::get_nr_result(const std::string &handle) {}

optional<floatxx_t> compatible_checker::get_nr_sum(const std::string &handle) {}

optional<std::vector<address_t>>
compatible_checker::get_nr_addr_list(const std::string &handle) {}

optional<rt::dip::dip_ret_type>
compatible_checker::get_dip_result(const std::string &handle) {}

optional<std::string> compatible_checker::get_nr_handle(block_height_t height) {

}
optional<std::string>
compatible_checker::get_dip_handle(block_height_t height) {}

} // namespace compatible
} // namespace neb
