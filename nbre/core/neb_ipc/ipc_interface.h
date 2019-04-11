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

// interface get nbre version
int ipc_nbre_version(void *holder, uint64_t height);
typedef void (*nbre_version_callback_t)(enum ipc_status_code isc, void *holder,
                                        uint32_t major, uint32_t minor,
                                        uint32_t patch);
void set_recv_nbre_version_callback(nbre_version_callback_t func);

// interface get nbre ir list
int ipc_nbre_ir_list(void *holder);
typedef void (*nbre_ir_list_callback_t)(enum ipc_status_code isc, void *holder,
                                        const char *ir_name_list);
void set_recv_nbre_ir_list_callback(nbre_ir_list_callback_t func);

// interface get ir version list
int ipc_nbre_ir_versions(void *holder, const char *ir_name);
typedef void (*nbre_ir_versions_callback_t)(enum ipc_status_code isc,
                                            void *holder,
                                            const char *ir_versions);
void set_recv_nbre_ir_versions_callback(nbre_ir_versions_callback_t func);

// interface get nr handler
int ipc_nbre_nr_handler(void *holder, uint64_t start_block, uint64_t end_block,
                        uint64_t nr_version);
typedef void (*nbre_nr_handler_callback_t)(enum ipc_status_code isc,
                                           void *holder,
                                           const char *nr_handler);
void set_recv_nbre_nr_handler_callback(nbre_nr_handler_callback_t func);

// interface get nr result
int ipc_nbre_nr_result(void *holder, const char *nr_handler);
typedef void (*nbre_nr_result_callback_t)(enum ipc_status_code isc,
                                          void *holder, const char *nr_result);
void set_recv_nbre_nr_result_callback(nbre_nr_result_callback_t func);

// interface get dip reward
int ipc_nbre_dip_reward(void *holder, uint64_t height, uint64_t version);
typedef void (*nbre_dip_reward_callback_t)(enum ipc_status_code isc,
                                           void *holder,
                                           const char *dip_reward);
void set_recv_nbre_dip_reward_callback(nbre_dip_reward_callback_t func);

typedef struct {
  const char *m_nbre_root_dir;
  const char *m_nbre_exe_name;
  const char *m_neb_db_dir;
  const char *m_nbre_db_dir;
  const char *m_nbre_log_dir;
  const char *m_admin_pub_addr;
  uint64_t m_nbre_start_height;
} nbre_params_t;

// nbre ipc start and shutdown
int start_nbre_ipc(const nbre_params_t params);

void nbre_ipc_shutdown();

#ifdef __cplusplus
}
#endif

#endif
