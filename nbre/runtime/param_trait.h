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
#include "jit/jit_driver.h"
#include "runtime/nr/impl/data_type.h"

namespace neb {
namespace rt {
class param_trait {
public:
  template <class T>
  static T get_param(const nbre::NBREIR &ir, const std::string &func_name) {
    std::vector<nbre::NBREIR> irs;
    irs.push_back(ir);

    std::string key = ir.name() + std::to_string(ir.version()) + "get_param";
    return jit_driver::instance().run<T>(key, irs, func_name);
  }
};
} // namespace rt
} // namespace neb
