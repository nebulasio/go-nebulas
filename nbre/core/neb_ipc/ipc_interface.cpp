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
#include "core/neb_ipc/ipc_interface.h"
#include "core/neb_ipc/server/ipc_callback_holder.h"
#include "core/neb_ipc/server/ipc_server_endpoint.h"

std::shared_ptr<neb::core::ipc_server_endpoint> _ipc;

int start_nbre_ipc(const nbre_params_t params) {
  try {
    FLAGS_log_dir = params.m_nbre_log_dir;
    google::InitGoogleLogging("nbre-server");

    _ipc = std::make_shared<neb::core::ipc_server_endpoint>();
    LOG(INFO) << "ipc server construct";
    _ipc->init_params(params);

    if (_ipc->start()) {
      LOG(INFO) << "start nbre succ";
      return ipc_status_succ;
    } else {
      LOG(ERROR) << "start nbre failed";
      return ipc_status_fail;
    }
  } catch (const std::exception &e) {
    LOG(ERROR) << "start nbre got exception " << typeid(e).name() << ":"
               << e.what();
    return ipc_status_exception;
  } catch (...) {
    LOG(ERROR) << "start nbre got unknown exception ";
    return ipc_status_exception;
  }
}

void nbre_ipc_shutdown() {
  _ipc->shutdown();
  _ipc.reset();
}

void set_recv_nbre_version_callback(nbre_version_callback_t func) {
  neb::core::ipc_callback_holder::instance().m_nbre_version_callback = func;
}

int ipc_nbre_version(void *holder, uint64_t height) {
  return _ipc->send_nbre_version_req(holder, height);
}

void set_recv_nbre_ir_list_callback(nbre_ir_list_callback_t func) {
  neb::core::ipc_callback_holder::instance().m_nbre_ir_list_callback = func;
}

int ipc_nbre_ir_list(void *holder) {
  return _ipc->send_nbre_ir_list_req(holder);
}

void set_recv_nbre_ir_versions_callback(nbre_ir_versions_callback_t func) {
  neb::core::ipc_callback_holder::instance().m_nbre_ir_versions_callback = func;
}

int ipc_nbre_ir_versions(void *holder, const char *ir_name) {
  return _ipc->send_nbre_ir_versions_req(holder, ir_name);
}

void set_recv_nbre_nr_handler_callback(nbre_nr_handler_callback_t func) {
  neb::core::ipc_callback_holder::instance().m_nbre_nr_handler_callback = func;
}

int ipc_nbre_nr_handler(void *holder, uint64_t start_block, uint64_t end_block,
                        uint64_t nr_version) {
  return _ipc->send_nbre_nr_handler_req(holder, start_block, end_block,
                                        nr_version);
}

void set_recv_nbre_nr_result_callback(nbre_nr_result_callback_t func) {
  neb::core::ipc_callback_holder::instance().m_nbre_nr_result_callback = func;
}

int ipc_nbre_nr_result(void *holder, const char *nr_handler) {
  return _ipc->send_nbre_nr_result_req(holder, nr_handler);
}

void set_recv_nbre_dip_reward_callback(nbre_dip_reward_callback_t func) {
  neb::core::ipc_callback_holder::instance().m_nbre_dip_reward_callback = func;
}

int ipc_nbre_dip_reward(void *holder, uint64_t height, uint64_t version) {
  return _ipc->send_nbre_dip_reward_req(holder, height);
}
