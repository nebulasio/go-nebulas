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

void callback_handler::handle_nr(void *holder, const char *nr_handler_id) {
  handle(m_nr_handlers, holder, nr_handler_id);
}
void callback_handler::handle_nr_result(void *holder, const char *nr_result) {
  handle(m_nr_result_handlers, holder, nr_result);
}

void nbre_version_callback(ipc_status_code isc, void *handler, uint32_t major,
                           uint32_t minor, uint32_t patch) {}

void nbre_ir_list_callback(ipc_status_code isc, void *handler,
                           const char *ir_name_list) {}
void nbre_ir_versions_callback(ipc_status_code isc, void *handler,
                               const char *ir_versions) {}

void nbre_nr_handle_callback(ipc_status_code isc, void *holder,
                             const char *nr_handle_id) {
  LOG(INFO) << "nbre_nr_handle_callback: got handle id: " << nr_handle_id;
  callback_handler::instance().handle_nr(holder, nr_handle_id);
}

void nbre_nr_result_callback(ipc_status_code isc, void *holder,
                             const char *nr_result) {
  callback_handler::instance().handle_nr_result(holder, nr_result);
}

void nbre_dip_reward_callback(ipc_status_code isc, void *holder,
                              const char *dip_reward) {}

