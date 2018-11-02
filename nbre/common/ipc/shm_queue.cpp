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
#include "common/ipc/shm_queue.h"

namespace neb {
namespace ipc {
namespace internal {

shm_queue::shm_queue(const std::string &name, shm_session_base *session,
                     boost::interprocess::managed_shared_memory *shmem,
                     size_t capacity)
    : m_name(name), m_shmem(shmem), m_capacity(capacity), m_session(session) {
  try {
    if (!m_shmem) {
      throw shm_queue_failure("shmem can't be nullptr");
    }
    if (!m_capacity) {
      throw shm_queue_failure("capacity can't be 0");
    }
    m_allocator = new shmem_allocator_t(m_shmem->get_segment_manager());

    m_mutex = m_session->bookkeeper()->acquire_named_mutex(mutex_name());
    m_empty_cond =
        m_session->bookkeeper()->acquire_named_condition(empty_cond_name());
    m_full_cond =
        m_session->bookkeeper()->acquire_named_condition(full_cond_name());

    m_buffer =
        m_shmem->find_or_construct<shm_vector_t>(m_name.c_str())(*m_allocator);
    if (!m_mutex) {
      throw shm_queue_failure("alloc mutex fail");
    }
    if (!m_empty_cond) {
      throw shm_queue_failure("alloc empty cond fail");
    }
    if (!m_full_cond) {
      throw shm_queue_failure("alloc full cond fail");
    }
    if (!m_buffer) {
      throw shm_queue_failure("alloc vector fail");
    }
  } catch (const std::exception &e) {
    throw shm_init_failure(std::string("shm_queue, ") +
                           std::string(typeid(e).name()) + " : " + e.what());
  }
};

void shm_queue::reset() {
  m_session->bookkeeper()->reset();
  // boost::interprocess::named_mutex::remove(mutex_name().c_str());
  // boost::interprocess::named_condition::remove(empty_cond_name().c_str());
  // boost::interprocess::named_condition::remove(full_cond_name().c_str());
  // m_session->reset();
}
void shm_queue::push_back(shm_type_id_t type_id, void *ptr) {
  boost::interprocess::scoped_lock<boost::interprocess::named_mutex> _l(
      *m_mutex);
  if (m_buffer->size() == m_capacity) {
    m_full_cond->wait(_l);
  }
  if (m_buffer->size() >= m_capacity)
    return;
  boost::interprocess::managed_shared_memory::handle_t h =
      m_shmem->get_handle_from_address(ptr);
  vector_elem_t e;
  e.m_handle = h;
  e.m_type = type_id;
  e.m_op_type = new_object;
  try {
    m_buffer->push_back(e);
  } catch (const std::exception &e) {
    LOG(ERROR) << "got exception " << e.what();
  }
  if (m_buffer->size() == 1) {
    m_empty_cond->notify_all();
  }
}

std::tuple<void *, shm_type_id_t, shm_queue::element_op_tag>
shm_queue::pop_front() {
  boost::interprocess::scoped_lock<boost::interprocess::named_mutex> _l(
      *m_mutex);
  if (m_buffer->empty()) {
    m_empty_cond->wait(_l);
  }
  if (m_buffer->empty()) {
    return std::make_tuple<void *, shm_type_id_t, element_op_tag>(nullptr, 0,
                                                                  new_object);
  }
  vector_elem_t e;
  e = m_buffer->front();
  m_buffer->erase(m_buffer->begin());
  if (m_buffer->size() == m_capacity - 1) {
    m_full_cond->notify_all();
  }
  return std::make_tuple(m_shmem->get_address_from_handle(e.m_handle), e.m_type,
                         e.m_op_type);
}

std::tuple<void *, shm_type_id_t, shm_queue::element_op_tag>
shm_queue::try_pop_front() {
  boost::interprocess::scoped_lock<boost::interprocess::named_mutex> _l(
      *m_mutex);
  if (m_buffer->empty()) {
    return std::make_tuple<void *, shm_type_id_t, element_op_tag>(nullptr, 0,
                                                                  new_object);
  }
  vector_elem_t e;
  e = m_buffer->front();
  m_buffer->erase(m_buffer->begin());
  if (m_buffer->size() == m_capacity - 1) {
    m_full_cond->notify_all();
  }
  return std::make_tuple(m_shmem->get_address_from_handle(e.m_handle), e.m_type,
                         e.m_op_type);
}

size_t shm_queue::size() const {
  boost::interprocess::scoped_lock<boost::interprocess::named_mutex> _l(
      *m_mutex);
  return m_buffer->size();
}

void shm_queue::wake_up_if_empty() {
  boost::interprocess::scoped_lock<boost::interprocess::named_mutex> _l(
      *m_mutex);

  if (m_buffer->empty()) {
    m_empty_cond->notify_all();
  }
}
size_t shm_queue::empty() const {
  boost::interprocess::scoped_lock<boost::interprocess::named_mutex> _l(
      *m_mutex);
  return m_buffer->empty();
}

shm_queue::~shm_queue() {
  LOG(INFO) << "m_shmem: " << (void *)m_shmem;
  if (m_shmem && m_buffer) {
    m_shmem->destroy_ptr(m_buffer);
  }
  if (m_allocator) {
    delete m_allocator;
  }
  m_session->bookkeeper()->release_named_mutex(mutex_name());
  m_session->bookkeeper()->release_named_condition(empty_cond_name());
  m_session->bookkeeper()->release_named_condition(full_cond_name());
}
}
}
}
