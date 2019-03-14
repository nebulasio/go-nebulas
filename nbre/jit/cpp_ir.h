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
#include "common/configuration.h"
#include "util/command.h"

namespace neb {
namespace cpp {
class cpp_ir {
public:
  typedef std::string name_version_t;
  typedef std::string cpp_content_t;
  typedef std::pair<name_version_t, cpp_content_t> cpp_t;
  cpp_ir(const cpp_t &cpp);

  neb::bytes llvm_ir_content();

protected:
  int make_ir_bitcode(const std::string &cpp_file,
                      const std::string &ir_bc_file);

  std::string generate_fp();

protected:
  name_version_t m_name_version;
  cpp_content_t m_cpp_content;
  std::string m_cpp_fp;
  std::string m_llvm_ir_fp;
  bool m_b_got_error;
  static std::atomic_int s_file_counter;
};
} // namespace cpp
} // namespace neb
