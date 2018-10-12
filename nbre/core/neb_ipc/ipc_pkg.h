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
#include "core/neb_ipc/ipc_common.h"

//! According to boost.interprocess document, we should never use reference,
//! pointer to store data, and we should never use virtual functions here. We
//! must make sure each class here is POD, thus we can pass it with interprocess
//! communication.
//
//! Also, our IPC framework needs *pkg_identifier* in each class.
namespace neb {
namespace core {
using ipc_pkg_type_id_t = neb::ipc::shm_type_id_t;

enum {
  ipc_pkg_nbre_version_req,
  ipc_pkg_nbre_version_ack,
};
namespace internal {
template <ipc_pkg_type_id_t type> struct empty_req_pkg_t {
  const static ipc_pkg_type_id_t pkg_identifier = type;
};
template <ipc_pkg_type_id_t type>
const ipc_pkg_type_id_t empty_req_pkg_t<type>::pkg_identifier;
} // namespace internal

typedef internal::empty_req_pkg_t<ipc_pkg_nbre_version_req> nbre_version_req;
struct nbre_version_ack {
  const static ipc_pkg_type_id_t pkg_identifier = ipc_pkg_nbre_version_ack;
  uint32_t m_major;
  uint32_t m_minor;
  uint32_t m_patch;
};
}
} // namespace neb
