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
  LOG(INFO) << "shm_queue enter";
  if (!m_shmem) {
    LOG(INFO) << "shm_queue enter 1";
    throw shm_queue_failure("shmem can't be nullptr");
  }
  if (!m_capacity) {
    LOG(INFO) << "shm_queue enter 3";
    throw shm_queue_failure("capacity can't be 0");
  }
  m_allocator = new shmem_allocator_t(m_shmem->get_segment_manager());
  LOG(INFO) << "allocator is " << (void *)m_allocator;

  m_mutex = m_session->bookkeeper()->acquire_named_mutex(mutex_name());
  m_empty_cond =
      m_session->bookkeeper()->acquire_named_condition(empty_cond_name());
  m_full_cond =
      m_session->bookkeeper()->acquire_named_condition(full_cond_name());

  LOG(INFO) << "to allocate buffer";
  m_buffer =
      m_shmem->find_or_construct<shm_vector_t>(m_name.c_str())(*m_allocator);
  LOG(INFO) << "allocate buffer done";
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
  LOG(INFO) << "shm_queue done";
};

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
  m_buffer->push_back(e);
  if (m_buffer->size() == 1) {
    m_empty_cond->notify_one();
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
    LOG(INFO) << "return empty";
    return std::make_tuple<void *, shm_type_id_t, element_op_tag>(nullptr, 0,
                                                                  new_object);
  }
  vector_elem_t e;
  e = m_buffer->front();
  m_buffer->erase(m_buffer->begin());
  if (m_buffer->size() == m_capacity - 1) {
    m_full_cond->notify_one();
  }
  return std::make_tuple(m_shmem->get_address_from_handle(e.m_handle), e.m_type,
                         e.m_op_type);
}

std::tuple<void *, shm_type_id_t, shm_queue::element_op_tag>
shm_queue::try_pop_front() {
  boost::interprocess::scoped_lock<boost::interprocess::named_mutex> _l(
      *m_mutex);
  if (m_buffer->empty()) {
    LOG(INFO) << "return empty";
    return std::make_tuple<void *, shm_type_id_t, element_op_tag>(nullptr, 0,
                                                                  new_object);
  }
  vector_elem_t e;
  e = m_buffer->front();
  m_buffer->erase(m_buffer->begin());
  if (m_buffer->size() == m_capacity - 1) {
    m_full_cond->notify_one();
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
    LOG(INFO) << "wake up empty shm_queue";
    m_empty_cond->notify_all();
  }
}
size_t shm_queue::empty() const {
  boost::interprocess::scoped_lock<boost::interprocess::named_mutex> _l(
      *m_mutex);
  return m_buffer->empty();
}

shm_queue::~shm_queue() {
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
