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
#include "common/ipc/shm_service_recv_handler.h"

namespace neb {
namespace ipc {
namespace internal {

shm_service_recv_handler::shm_service_recv_handler(
    boost::interprocess::managed_shared_memory *shmem, shm_service_op_queue *q)
    : m_shmem(shmem), m_op_queue(q) {}

void shm_service_recv_handler::handle_recv_op(
    const std::shared_ptr<shm_service_op_base> &op) {
  assert(op->op_id() == shm_service_op_base::op_recv_obj);

  shm_service_op_recv *recv_op = static_cast<shm_service_op_recv *>(op.get());

  void *data_pointer = recv_op->m_pointer;
  shm_type_id_t type_id = recv_op->m_type_id;
  if (data_pointer) {
    typename decltype(m_all_handlers)::const_iterator fr =
        m_all_handlers.find(type_id);
    if (fr != m_all_handlers.end()) {
      auto func = fr->second;
      m_handler_thread.schedule([func, data_pointer]() {
        try {
          func(data_pointer);
        } catch (const std::exception &e) {
          LOG(WARNING) << "got exception when call handler "
                       << std::string(typeid(e).name()) << " : " << e.what();
          // throw shm_handle_recv_failure(
          // std::string("shm_service_recv_handler, ") +
          // std::string(typeid(e).name()) + " : " + e.what());
        }
      });
    } else {
      LOG(WARNING) << "cannot find data handler";
    }
  }
}
} // namespace internal
} // namespace ipc
} // namespace neb
