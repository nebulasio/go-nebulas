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
#include "fs/proto/ir.pb.h"

namespace neb {
namespace internal {
class jit_driver_impl;
}
namespace core {
class driver;
}
class jit_driver {
public:
  jit_driver();
  ~jit_driver();
  void run(core::driver *d,
           const std::vector<std::shared_ptr<nbre::NBREIR>> &irs,
           const std::string &func_name, void *param);

  void auth_run(const nbre::NBREIR &ir, const std::string &func_name,
                auth_table_t &auth_table);

protected:
  std::unique_ptr<internal::jit_driver_impl> m_impl;
}; // end class jit_driver;
} // end namespace neb
