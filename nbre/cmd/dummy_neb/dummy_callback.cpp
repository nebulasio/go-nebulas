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
#include "cmd/dummy_neb/dummy_callback.h"

callback_handler::callback_handler() {}

void callback_handler::handle_version(void *holder, uint32_t major,
                                      uint32_t minor, uint32_t patch) {
  handle(m_version_handlers, holder, major, minor, patch);
}
void callback_handler::handle_ir_list(void *holder, const char *ir_name_list) {
  handle(m_ir_list_handlers, holder, ir_name_list);
}
void callback_handler::handle_ir_versions(void *holder,
                                          const char *ir_versions) {
  handle(m_ir_versions_handlers, holder, ir_versions);
}
void callback_handler::handle_nr(void *holder, const char *nr_handler_id) {
  handle(m_nr_handlers, holder, nr_handler_id);
}
void callback_handler::handle_nr_result(void *holder, const char *nr_result) {
  handle(m_nr_result_handlers, holder, nr_result);
}
void callback_handler::handle_nr_result_by_height(void *holder,
                                                  const char *nr_result) {
  handle(m_nr_result_by_height_handlers, holder, nr_result);
}
void callback_handler::handle_nr_sum(void *holder, const char *nr_sum) {
  handle(m_nr_sum_handlers, holder, nr_sum);
}
void callback_handler::handle_dip_reward(void *holder, const char *dip_reward) {
  handle(m_dip_reward_handlers, holder, dip_reward);
}

void nbre_version_callback(ipc_status_code isc, void *handler, uint32_t major,
                           uint32_t minor, uint32_t patch) {
  if (isc != ipc_status_succ) {
    LOG(ERROR) << "nbre_version_call_back got failed ";
    return;
  }
  callback_handler::instance().handle_version(handler, major, minor, patch);
}

void nbre_ir_list_callback(ipc_status_code isc, void *handler,
                           const char *ir_name_list) {
  if (isc != ipc_status_succ) {
    LOG(ERROR) << "nbre_ir_list_callback got failed ";
    return;
  }
  callback_handler::instance().handle_ir_list(handler, ir_name_list);
}
void nbre_ir_versions_callback(ipc_status_code isc, void *handler,
                               const char *ir_versions) {
  if (isc != ipc_status_succ) {
    LOG(ERROR) << "nbre_ir_versions_callback got failed ";
    return;
  }
  callback_handler::instance().handle_ir_versions(handler, ir_versions);
}

void nbre_nr_handle_callback(ipc_status_code isc, void *holder,
                             const char *nr_handle_id) {
  if (isc != ipc_status_succ) {
    LOG(ERROR) << "nbre_nr_handle_callback got failed ";
    return;
  }
  callback_handler::instance().handle_nr(holder, nr_handle_id);
}

void nbre_nr_result_callback(ipc_status_code isc, void *holder,
                             const char *nr_result) {
  if (isc != ipc_status_succ) {
    LOG(ERROR) << "nbre_nr_result_callback got failed ";
    return;
  }
  callback_handler::instance().handle_nr_result(holder, nr_result);
}

void nbre_nr_result_by_height_callback(ipc_status_code isc, void *holder,
                                       const char *nr_result) {
  if (isc != ipc_status_succ) {
    LOG(ERROR) << "nbre_nr_result_by_height_callback got failed ";
    return;
  }
  callback_handler::instance().handle_nr_result_by_height(holder, nr_result);
}

void nbre_nr_sum_callback(ipc_status_code isc, void *holder,
                          const char *nr_sum) {
  if (isc != ipc_status_succ) {
    LOG(ERROR) << "nbre_nr_sum_callback got failed ";
    return;
  }
  callback_handler::instance().handle_nr_sum(holder, nr_sum);
}

void nbre_dip_reward_callback(ipc_status_code isc, void *holder,
                              const char *dip_reward) {
  if (isc != ipc_status_succ) {
    LOG(ERROR) << "nbre_dip_reward_callback got failed ";
    return;
  }
  callback_handler::instance().handle_dip_reward(holder, dip_reward);
}

