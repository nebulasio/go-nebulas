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
#include "common/address.h"
#include "common/common.h"
#include <ff/network.h>

enum ipc_pkg_type {
  heart_beat_pkg,
#define define_ipc_param(type, name)

#define define_ipc_pkg(type, ...) JOIN(type, _pkg),

#define define_ipc_api(req, ack)

#include "core/net_ipc/ipc_interface_impl.h"

#undef define_ipc_api
#undef define_ipc_pkg
#undef define_ipc_param
  nipc_last_pkg_id,
};

typedef ::ff::net::ntpackage<heart_beat_pkg> heart_beat_t;

#define define_ipc_param(type, name) define_nt(name, type);

#define define_ipc_pkg(type, ...)                                              \
  typedef ::ff::net::ntpackage<JOIN(type, _pkg), __VA_ARGS__> type;

#define define_ipc_api(req, ack)

#include "core/net_ipc/ipc_interface_impl.h"

#undef define_ipc_api
#undef define_ipc_pkg
#undef define_ipc_param

namespace neb {
namespace core {
template <typename PkgType1, typename PkgType2>
std::shared_ptr<PkgType1> new_ack_pkg(PkgType2 req) {
  auto ack = std::make_shared<PkgType1>();
  ack->template set<p_holder>(req->template get<p_holder>());
  return ack;
}

std::string pkg_type_id_to_name(uint64_t type);
bool is_pkg_type_has_callback(uint64_t type);

std::string convert_nr_result_to_json(const nr_result &nr);

template <typename T> struct get_pkg_ack_type { typedef T type; };
#define define_ipc_param(type, name)
#define define_ipc_pkg(type, ...)

#define define_ipc_api(req, ack)                                               \
  template <> struct get_pkg_ack_type<req> { typedef ack type; };

#include "core/net_ipc/ipc_interface_impl.h"

#undef define_ipc_api
#undef define_ipc_pkg
#undef define_ipc_param

struct result_status {
  const static uint32_t succ = 0;
  const static uint32_t is_running = 1;
  const static uint32_t no_cached = 2;
  const static uint32_t unknown = 255;
};

std::string result_status_to_string(uint32_t status);

template <class T> bool is_succ(const T &pkg) {
  return pkg.template get<p_result_status>() == result_status::succ;
}
template <class T> bool is_succ(const std::shared_ptr<T> &pkg) {
  return pkg->template get<p_result_status>() == result_status::succ;
}

} // namespace core
} // namespace neb
