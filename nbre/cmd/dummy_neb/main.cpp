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
#include "common/common.h"
#include "common/util/version.h"
#include "core/ipc_configuration.h"
#include "core/net_ipc/ipc_interface.h"
//#include "core/neb_ipc/server/ipc_server_endpoint.h"
#include "fs/util.h"

std::mutex local_mutex;
std::condition_variable local_cond_var;
bool to_quit = false;

void nbre_version_callback(ipc_status_code isc, void *handler, uint32_t major,
                           uint32_t minor, uint32_t patch) {
  LOG(INFO) << "got version: " << major << ", " << minor << ", " << patch;
  std::unique_lock<std::mutex> _l(local_mutex);
  to_quit = true;
  _l.unlock();
  // local_cond_var.notify_one();
}

void nbre_ir_list_callback(ipc_status_code isc, void *handler,
                           const char *ir_name_list) {
  LOG(INFO) << ir_name_list;
  std::unique_lock<std::mutex> _l(local_mutex);
  to_quit = true;
  _l.unlock();
}

void nbre_ir_versions_callback(ipc_status_code isc, void *handler,
                               const char *ir_versions) {
  LOG(INFO) << ir_versions;
  std::unique_lock<std::mutex> _l(local_mutex);
  to_quit = true;
  _l.unlock();
}

void nbre_nr_handler_callback(ipc_status_code isc, void *holder,
                              const char *nr_handler_id) {
  LOG(INFO) << nr_handler_id;
  std::unique_lock<std::mutex> _l(local_mutex);
  to_quit = true;
  _l.unlock();
}

void nbre_nr_result_callback(ipc_status_code isc, void *holder,
                             const char *nr_result) {
  LOG(INFO) << nr_result;
  std::unique_lock<std::mutex> _l(local_mutex);
  to_quit = true;
  _l.unlock();
}

void nbre_dip_reward_callback(ipc_status_code isc, void *holder,
                              const char *dip_reward) {
  LOG(INFO) << dip_reward;
  std::unique_lock<std::mutex> _l(local_mutex);
  to_quit = true;
  _l.unlock();
}

int main(int argc, char *argv[]) {
  FLAGS_logtostderr = true;

  //::google::InitGoogleLogging(argv[0]);

  const char *root_dir =
      neb::core::ipc_configuration::instance().nbre_root_dir().c_str();
  std::string nbre_path = neb::fs::join_path(root_dir, "bin/nbre");

  set_recv_nbre_version_callback(nbre_version_callback);
  set_recv_nbre_ir_list_callback(nbre_ir_list_callback);
  set_recv_nbre_ir_versions_callback(nbre_ir_versions_callback);
  set_recv_nbre_nr_handler_callback(nbre_nr_handler_callback);
  set_recv_nbre_nr_result_callback(nbre_nr_result_callback);
  set_recv_nbre_dip_reward_callback(nbre_dip_reward_callback);

  nbre_params_t params{
      root_dir,
      nbre_path.c_str(),
      neb::core::ipc_configuration::instance().neb_db_dir().c_str(),
      neb::core::ipc_configuration::instance().nbre_db_dir().c_str(),
      neb::core::ipc_configuration::instance().nbre_log_dir().c_str(),
      "auth address here!"};
  params.m_nipc_port = 6987;

  auto ret = start_nbre_ipc(params);
  if (ret != ipc_status_succ) {
    to_quit = false;
    nbre_ipc_shutdown();
    return -1;
  }

  uint64_t height = 100;

  ipc_nbre_version(&local_mutex, height);
  ipc_nbre_ir_list(&local_mutex);
  // ipc_nbre_ir_versions(&local_mutex, "dip");

  ipc_nbre_nr_handler(&local_mutex, 6600, 6650,
                      neb::util::version(0, 1, 0).data());
  while (true) {
    ipc_nbre_nr_result(&local_mutex,
                       "00000000000019c800000000000019fa0000000100000000");
    std::this_thread::sleep_for(std::chrono::seconds(1));
  }

  // while (true) {
  // ipc_nbre_dip_reward(&local_mutex, 60000);
  // std::this_thread::sleep_for(std::chrono::seconds(1));
  //}
  std::unique_lock<std::mutex> _l(local_mutex);
  if (to_quit) {
    return 0;
  }
  local_cond_var.wait(_l);

  return 0;
}
