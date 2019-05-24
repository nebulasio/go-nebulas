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
namespace jit {
class jit_mangled_entry_point {
public:
  jit_mangled_entry_point();

  virtual ~jit_mangled_entry_point();

  virtual std::string get_mangled_entry_name(const std::string &entry_name);

protected:
  void init_prog_slice();
  void gen_mangle_name_for_entry(const std::string &entry_name);
  std::string get_storage_key(const std::string &entry_name);

protected:
  std::mutex m_mutex;
  std::unordered_map<std::string, std::string> m_prog_slice;
  std::unordered_map<std::string, std::string> m_mangled_entry_names;
};
} // namespace jit
} // namespace neb
