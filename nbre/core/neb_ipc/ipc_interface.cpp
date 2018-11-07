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

int start_nbre_ipc(const char *root_dir, const char *nbre_path,
                   const char *admin_pub_addr) {
  try {
    _ipc =
        std::make_shared<neb::core::ipc_server_endpoint>(root_dir, nbre_path);
    _ipc->init_params(admin_pub_addr);

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

int ipc_nbre_ir_list(void *holder) {
  return _ipc->send_nbre_ir_list_req(holder);
}
