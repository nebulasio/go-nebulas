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
#include "common/util/singleton.h"
#include "core/neb_ipc/ipc_interface.h"

namespace neb {
namespace core {
class ipc_callback_holder : public neb::util::singleton<ipc_callback_holder> {
public:
  ipc_callback_holder() = default;
  ~ipc_callback_holder() = default;

  nbre_version_callback_t m_nbre_version_callback;
  bool check_all_callbacks();
};
} // namespace core
} // namespace neb
