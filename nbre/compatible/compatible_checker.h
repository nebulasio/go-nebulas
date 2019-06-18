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
#include "compatible/compatible_check_base.h"
#include "compatible/version_check_interface.h"
#include "runtime/dip/data_type.h"
#include "runtime/nr/impl/data_type.h"

namespace neb {
namespace compatible {
class compatible_checker : public compatible_check_base {
public:
  virtual ~compatible_checker();

  virtual void init();

  virtual bool is_ir_need_compile(const std::string &name, version_t version);

  virtual optional<rt::nr::nr_ret_type>
  get_nr_result(const std::string &handle);

  virtual optional<floatxx_t> get_nr_sum(const std::string &handle);

  virtual optional<std::vector<address_t>>
  get_nr_addr_list(const std::string &handle);

  virtual optional<rt::dip::dip_ret_type>
  get_dip_result(const std::string &handle);

  virtual optional<std::string> get_nr_handle(block_height_t height);

  virtual optional<std::string> get_dip_handle(block_height_t height);

protected:
  std::vector<std::unique_ptr<version_check_interface>> m_checkers;
};
} // namespace compatible
} // namespace neb
