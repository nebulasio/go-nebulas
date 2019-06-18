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
#include "fs/proto/ir.pb.h"

namespace neb {
namespace fs {
namespace internal {

class ir_item_list_interface {
public:
  virtual bool ir_exist(version_t v) = 0;
  virtual void write_ir(const nbre::NBREIR &raw_ir,
                        const nbre::NBREIR &compiled_ir) = 0;

  virtual nbre::NBREIR get_raw_ir(version_t v) = 0;
  virtual nbre::NBREIR get_ir(version_t v) = 0;
  virtual bytes get_ir_brief_key_with_height(block_height_t height) = 0;
  virtual nbre::NBREIR find_ir_at_height(block_height_t height) = 0;

  virtual std::vector<std::string> get_ir_names() const = 0;
  virtual std::vector<version_t> get_ir_versions() const = 0;
};
} // namespace internal
} // namespace fs
} // namespace neb
