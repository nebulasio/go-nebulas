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
#include "runtime/version_impl.h"
#include "core/driver.h"
#include "core/neb_ipc/ipc_common.h"
#include "core/neb_ipc/ipc_pkg.h"
#include "runtime/version.h"

int entry_point_foo_impl(neb::core::driver *d) {
  auto ipc_conn = d->ipc_conn();
  neb::core::nbre_version_ack *ack =
      ipc_conn->construct<neb::core::nbre_version_ack>();
  neb::util::version v = neb::rt::get_version();
  ack->m_major = v.major_version();
  ack->m_minor = v.minor_version();
  ack->m_patch = v.patch_version();
  ipc_conn->push_back(ack);
  return 0;
}
