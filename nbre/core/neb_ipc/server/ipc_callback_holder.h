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
  nbre_ir_list_callback_t m_nbre_ir_list_callback;
  nbre_ir_versions_callback_t m_nbre_ir_versions_callback;
  nbre_nr_handler_callback_t m_nbre_nr_handler_callback;
  nbre_nr_result_callback_t m_nbre_nr_result_callback;
  nbre_dip_reward_callback_t m_nbre_dip_reward_callback;

  bool check_all_callbacks();
};

namespace internal {
template <typename T> struct issue_callback_with_error_helper {};

template <typename... ARGS>
struct issue_callback_with_error_helper<void (*)(ipc_status_code, ARGS...)> {
  using func_t = void (*)(ipc_status_code, ARGS...);
  static void call(const func_t &f, ipc_status_code isc) { f(isc, ARGS()...); }
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
