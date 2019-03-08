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
#include "core/net_ipc/nipc_pkg.h"

namespace neb {
namespace core {

std::string pkg_type_id_to_name(uint64_t type) {
  std::string ret;
  switch (type) {
  case heart_beat_pkg:
    ret = "heart_beat_pkg";
    break;
  case nipc_last_pkg_id:
    ret = "nipc_last_pkg_id";
    break;

#define define_ipc_param(type, name)

#define define_ipc_pkg(type, ...)                                              \
  case JOIN(type, _pkg):                                                       \
    ret = #type;                                                               \
    break;

#define define_ipc_api(req, ack)

#include "core/net_ipc/ipc_interface_impl.h"

#undef define_ipc_api
#undef define_ipc_pkg
#undef define_ipc_param
  default:
    ret = "not_def_in_nipc";
  }

  return ret;
}

bool is_pkg_type_has_callback(uint64_t type) {

#define define_ipc_param(type, name)
#define define_ipc_pkg(type, ...)
#define define_ipc_api(req, ack)                                               \
  if (type == JOIN(req, _pkg))                                                 \
    return true;

#include "core/net_ipc/ipc_interface_impl.h"

#undef define_ipc_api
#undef define_ipc_pkg
#undef define_ipc_param
  return false;
}
} // namespace core
} // namespace neb
