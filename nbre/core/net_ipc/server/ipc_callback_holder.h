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
#include "core/net_ipc/ipc_interface.h"
#include "core/net_ipc/nipc_pkg.h"
#include "util/singleton.h"
#include <ff/network.h>

namespace neb {
namespace core {
class ipc_callback_holder : public neb::util::singleton<ipc_callback_holder> {
public:
  ipc_callback_holder() = default;
  ~ipc_callback_holder() = default;

  inline void add_callback(
      uint64_t callback_id,
      const std::function<void(enum ipc_status_code, ::ff::net::package *)>
          &func) {
    m_callbacks.insert(std::make_pair(callback_id, func));
  }
  inline std::function<void(enum ipc_status_code, ::ff::net::package *)>
  get_callback(uint64_t pkg_id) {
    if (m_callbacks.find(pkg_id) == m_callbacks.end()) {
      if (is_pkg_type_has_callback(pkg_id)) {
        LOG(WARNING) << "Cann't find callback for "
                     << pkg_type_id_to_name(pkg_id) << " " << pkg_id;
      }
      return [](enum ipc_status_code, ::ff::net::package *) {};
    }
    return m_callbacks[pkg_id];
  }

  template <typename PkgType>
  void call_callback(const std::shared_ptr<PkgType> &pkg) {
    uint64_t pkg_id = PkgType().type_id();
    if (m_callbacks.find(pkg_id) == m_callbacks.end()) {
      LOG(WARNING) << "cannot find pkg type";
      return;
    }
    m_callbacks[pkg_id](ipc_status_succ, pkg.get());
  }
protected:
  typedef std::function<void(enum ipc_status_code, ::ff::net::package *)>
      callback_func_t;
  std::unordered_map<uint64_t, callback_func_t> m_callbacks;
};

namespace internal {
template <typename T> struct issue_callback_with_error_helper {};

template <typename... ARGS>
struct issue_callback_with_error_helper<void (*)(ipc_status_code, ARGS...)> {
  using func_t = void (*)(ipc_status_code, ARGS...);
  static void call(const func_t &f, ipc_status_code isc) {
    LOG(INFO) << "issue callback with err";
    f(isc, ARGS()...);
    LOG(INFO) << "issue callback with err done";
  }
};
} // namespace internal

template <typename T>
void issue_callback_with_error(T &&func, ipc_status_code isc) {
  internal::issue_callback_with_error_helper<
      std::remove_const_t<std::remove_reference_t<T>>>::call(func, isc);
}

#define CHECK_NBRE_STATUS(a)                                                   \
  if (!m_ipc_server) {                                                         \
    return ipc_status_fail;                                                    \
  }                                                                            \
  if (m_got_exception_when_start_nbre) {                                       \
    m_ipc_server->schedule_task_in_service_thread([this]() {                   \
      issue_callback_with_error(a, ipc_status_exception);                      \
      return ipc_status_succ;                                                  \
    });                                                                        \
  } else if (!m_client_watcher->is_client_alive()) {                           \
    m_ipc_server->schedule_task_in_service_thread([this]() {                   \
      issue_callback_with_error(a, ipc_status_nbre_not_ready);                 \
      return ipc_status_succ;                                                  \
    });                                                                        \
  }
} // namespace core
} // namespace neb
