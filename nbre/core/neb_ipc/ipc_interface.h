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
#ifndef NBRE_CORE_NEB_IPC_IPC_INTERFACE_H_
#define NBRE_CORE_NEB_IPC_IPC_INTERFACE_H_
#ifdef __cplusplus
#include <cstdint>
extern "C" {
#else
#include <stdint.h>
#endif

enum ipc_status_code {
  ipc_status_succ = 0,
  ipc_status_fail,
  ipc_status_timeout,
  ipc_status_exception,
  ipc_status_nbre_not_ready,
};

typedef void (*handle_recv_callback_func_t)(enum ipc_status_code isc, const char *);

void ipc_nbre_version(void *holder, uint64_t height);

typedef void (*nbre_version_callback_t)(enum ipc_status_code isc, void *holder,
                                        uint32_t major, uint32_t minor,
                                        uint32_t patch);

void set_recv_nbre_version_callback(nbre_version_callback_t func);

int start_nbre_ipc(const char *root_dir, const char *nbre_path,
                   const char *admin_pub_addr);

void nbre_ipc_shutdown();

#ifdef __cplusplus
}
#endif

#endif
