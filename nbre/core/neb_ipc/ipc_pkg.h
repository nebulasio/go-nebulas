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
#include "core/neb_ipc/internal/ipc_pkg_internal.h"
#include "core/neb_ipc/ipc_common.h"

namespace neb {
namespace core {
enum {
  ipc_pkg_nbre_version_req,
  ipc_pkg_nbre_version_ack,
  ipc_pkg_nbre_init_req,
  ipc_pkg_nbre_init_ack,
};
namespace ipc_pkg {
using namespace internal;

using height = ipc_elem_base<0, uint64_t>;
using nbre_version_req = define_ipc_pkg<ipc_pkg_nbre_version_req, height>;

using major = ipc_elem_base<1, uint32_t>;
using minor = ipc_elem_base<2, uint32_t>;
using patch = ipc_elem_base<3, uint32_t>;
using nbre_version_ack =
    define_ipc_pkg<ipc_pkg_nbre_version_ack, major, minor, patch>;

using nbre_init_req = define_ipc_pkg<ipc_pkg_nbre_init_req>;
using nbre_root_dir = ipc_elem_base<3, neb::ipc::char_string_t>;
using admin_pub_addr = ipc_elem_base<4, neb::ipc::char_string_t>;
using nbre_init_ack =
    define_ipc_pkg<ipc_pkg_nbre_init_ack, nbre_root_dir, admin_pub_addr>;

} // namespace ipc_pkg

add_pkg_module_info(ipc_pkg::nbre_version_req, "foo",
                    "_Z15entry_point_fooPN3neb4core6driverEPv")



} // namespace core
} // namespace neb
