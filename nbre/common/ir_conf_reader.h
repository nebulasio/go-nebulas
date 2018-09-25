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
#include "common/ir_ref.h"
#include "common/util/version.h"

namespace neb {

class ir_conf_reader {
public:
  ir_conf_reader(const std::string &conf_fp);

  inline const std::string &ir_fp() const { return m_ir_fp; }
  inline const ir_ref &self_ref() const { return m_self_ref; }
  inline const std::vector<ir_ref> depends() const { return m_depends; }
  inline block_height_t available_height() const { return m_available_height; }

protected:
  std::string m_ir_fp;
  ir_ref m_self_ref;
  std::vector<ir_ref> m_depends;
  block_height_t m_available_height;
};
}
