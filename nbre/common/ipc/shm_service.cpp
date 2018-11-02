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
#include "common/ipc/shm_service.h"

namespace neb {
namespace ipc {
namespace internal {

shm_service_base::shm_service_base(shm_role role, const std::string &shm_name,
                                   const std::string &shm_in_name,
                                   const std::string &shm_out_name,
                                   size_t mem_size, size_t shm_in_capacity,
                                   size_t shm_out_capacity)
    : m_role(role), m_mem_size(mem_size), m_shm_in_capacity(shm_in_capacity),
      m_shm_out_capacity(shm_out_capacity), m_shm_name(shm_name),
      m_shm_in_name(shm_in_name), m_shm_out_name(shm_out_name) {}

shm_service_base::~shm_service_base() {
  delete m_in_buffer;
  delete m_out_buffer;

  delete m_shmem;
  m_session->bookkeeper()->release(m_shm_name, [this]() {
    boost::interprocess::shared_memory_object::remove(m_shm_name.c_str());
  });
}

void shm_service_base::reset() {
  boost::interprocess::shared_memory_object::remove(m_shm_name.c_str());
  init_local_interprocess_var();
  m_in_buffer->reset();
  m_out_buffer->reset();
  m_session->reset();
}

void shm_service_base::run() { thread_func(); }

void shm_service_base::init_local_interprocess_var() {
  if (m_role == role_util) {
    m_session = std::unique_ptr<shm_session_base>(
        new shm_session_util(m_shm_name + ".session"));
  }
  if (m_role == role_server) {
    m_session = std::unique_ptr<shm_session_base>(
        new shm_session_server(m_shm_name + ".session"));
  }
  if (m_role == role_client) {
    m_session = std::unique_ptr<shm_session_base>(
        new shm_session_client(m_shm_name + ".session"));
  }

  m_session->bookkeeper()->acquire(m_shm_name, [this]() {
    m_shmem = new boost::interprocess::managed_shared_memory(
        boost::interprocess::open_or_create, m_shm_name.c_str(), m_mem_size);
  });
  m_char_allocator =
      std::make_unique<char_allocator_t>(m_shmem->get_segment_manager());
  m_default_allocator =
      std::make_unique<default_allocator_t>(m_shmem->get_segment_manager());

  m_in_buffer = new shm_queue(m_shm_in_name.c_str(), m_session.get(), m_shmem,
                              m_shm_in_capacity);
  m_out_buffer = new shm_queue(m_shm_out_name.c_str(), m_session.get(), m_shmem,
                               m_shm_out_capacity);
}

void shm_service_base::init_local_env() {
  init_local_interprocess_var();

  neb::core::command_queue::instance().listen_command<neb::core::exit_command>(
      this, [this](const std::shared_ptr<neb::core::exit_command> &) {
        m_exit_flag = true;
        m_op_queue->wake_up_if_empty();
      });
  m_op_queue =
      std::unique_ptr<shm_service_op_queue>(new shm_service_op_queue());
  m_constructer = std::unique_ptr<shm_service_construct_helper>(
      new shm_service_construct_helper(m_shmem, m_op_queue.get()));
  m_recv_handler = std::unique_ptr<shm_service_recv_handler>(
      new shm_service_recv_handler(m_shmem, m_op_queue.get()));
  m_queue_watcher = std::unique_ptr<shm_queue_watcher>(
      new shm_queue_watcher(m_in_buffer, m_op_queue.get()));
}

void shm_service_base::thread_func() {
  m_queue_watcher->start();
  m_session->start_session();
  try {
    while (!m_exit_flag) {
      auto ret = m_op_queue->pop_front();
      if (ret.first) {
        std::shared_ptr<shm_service_op_base> &op = ret.second;
        if (op->op_id() == shm_service_op_base::op_allocate_obj) {
          m_constructer->handle_construct_op(op);
        } else if (op->op_id() == shm_service_op_base::op_recv_obj) {
          m_recv_handler->handle_recv_op(op);
        } else if (op->op_id() == shm_service_op_base::op_push_back) {
          shm_service_op_push_back *push_op =
              static_cast<shm_service_op_push_back *>(op.get());
          m_out_buffer->push_back(push_op->m_type_id, push_op->m_pointer);
        } else if (op->op_id() == shm_service_op_base::op_destroy) {
          m_constructer->handle_destroy_op(op);
        } else if (op->op_id() == shm_service_op_base::op_general) {
          shm_service_op_general *og = (shm_service_op_general *)op.get();
          try {
            og->m_func();
          } catch (...) {
            LOG(ERROR) << "shm_service_base got exception when run general op";
          }
        }
      }
    }
  } catch (const std::exception &e) {
    LOG(ERROR) << "shm_service_base got: " << e.what();
  }
}
} // namespace internal
} // namespace ipc
} // namespace neb
