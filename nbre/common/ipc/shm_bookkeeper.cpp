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
#include "common/ipc/shm_bookkeeper.h"
#include "common/common.h"
#include "common/ipc/shm_base.h"

namespace neb {
namespace ipc {
void clean_bookkeeper_env(const std::string &name) {
  boost::interprocess::shared_memory_object::remove(name.c_str());
  boost::interprocess::named_mutex::remove((name + ".mutex").c_str());
}

namespace internal {

size_t bookkeeper_mem_size = 64 * 1024;

shm_bookkeeper::shm_bookkeeper(const std::string &name) : m_name(name) {
  try {
    m_segment = std::unique_ptr<boost::interprocess::managed_shared_memory>(
        new boost::interprocess::managed_shared_memory(
            boost::interprocess::open_or_create, m_name.c_str(),
            bookkeeper_mem_size));

    m_allocator = std::unique_ptr<map_allocator_t>(
        new map_allocator_t(m_segment->get_segment_manager()));

    m_mutex = std::unique_ptr<boost::interprocess::named_mutex>(
        new boost::interprocess::named_mutex(
            boost::interprocess::open_or_create, mutex_name().c_str()));

    m_map = m_segment->find_or_construct<map_t>(mem_name().c_str())(
        std::less<char_string_t>(), *m_allocator);
  } catch (const std::exception &e) {
    m_mutex.reset();
    m_segment.reset();
    throw shm_init_failure(std::string("shm_bookkeeper, ") +
                           std::string(typeid(e).name()) + " : " + e.what());
  }
}
void shm_bookkeeper::reset() {
  if (!m_mutex)
    return;
  //! We may fail to acquire m_mutex, thus, we just don't.
  m_mutex->try_lock();

  for (auto it = m_map->begin(); it != m_map->end(); ++it) {
    auto first = it->first;
    auto second = it->second;
    tag_counter_t tt;
    tt.set_data(second);
    if (tt.type() == tag_counter_t::boost_mutex) {
      boost::interprocess::named_mutex::remove(first.c_str());
    } else if (tt.type() == tag_counter_t::boost_semaphore) {
      boost::interprocess::named_semaphore::remove(first.c_str());
    } else if (tt.type() == tag_counter_t::boost_condition) {
      boost::interprocess::named_condition::remove(first.c_str());
    }
  }

  m_mutex->unlock();

  boost::interprocess::named_mutex::remove(mutex_name().c_str());
  boost::interprocess::shared_memory_object::remove(m_name.c_str());

  // LOG(INFO) << "realloc local vars";
  // m_segment = std::unique_ptr<boost::interprocess::managed_shared_memory>(
  // new boost::interprocess::managed_shared_memory(
  // boost::interprocess::open_or_create, m_name.c_str(),
  // bookkeeper_mem_size));

  // m_allocator = std::unique_ptr<map_allocator_t>(
  // new map_allocator_t(m_segment->get_segment_manager()));

  // m_mutex = std::unique_ptr<boost::interprocess::named_mutex>(
  // new boost::interprocess::named_mutex(boost::interprocess::open_or_create,
  // mutex_name().c_str()));

  // m_map = m_segment->find_or_construct<map_t>(mem_name().c_str())(
  // std::less<char_string_t>(), *m_allocator);
}

void shm_bookkeeper::acquire(const std::string &name,
                             const std::function<void()> &action) {
  acquire(name, action, tag_counter_t::other_unknown);
}

void shm_bookkeeper::acquire(const std::string &name,
                             const std::function<void()> &action,
                             uint8_t type) {
  boost::interprocess::scoped_lock<boost::interprocess::named_mutex> _l(
      *m_mutex);
  action();
  char_string_t cs(name.c_str(), *m_allocator);
  if (m_map->find(cs) == m_map->end()) {
    tag_counter_t tt;
    tt.set_type(static_cast<tag_counter_t::type_tag>(type));
    tt.set_counter(1);
    m_map->insert(std::pair<const char_string_t, uint64_t>(cs, tt.data()));
  } else {
    tag_counter_t tt;
    tt.set_data(m_map->operator[](cs));
    tt.set_counter(tt.counter() + 1);
    m_map->operator[](cs) = tt.data();
  }
}

void shm_bookkeeper::release(const std::string &name,
                             const std::function<void()> &action) {
  boost::interprocess::scoped_lock<boost::interprocess::named_mutex> _l(
      *m_mutex);

  char_string_t cs(name.c_str(), *m_allocator);
  if (m_map->find(cs) == m_map->end()) {
    return;
  }
  tag_counter_t tt;
  tt.set_data(m_map->operator[](cs));
  tt.set_counter(tt.counter() - 1);
  m_map->operator[](cs) = tt.data();
  if (tt.counter() == 0) {
    m_map->erase(cs);
    action();
  }
}
std::unique_ptr<boost::interprocess::named_mutex>
shm_bookkeeper::acquire_named_mutex(const std::string &name) {
  std::unique_ptr<boost::interprocess::named_mutex> ret;
  acquire(name,
          [&ret, name]() {
            ret = std::unique_ptr<boost::interprocess::named_mutex>(
                new boost::interprocess::named_mutex(
                    boost::interprocess::open_or_create, name.c_str()));
          },
          tag_counter_t::boost_mutex);
  return ret;
}

void shm_bookkeeper::release_named_mutex(const std::string &name) {
  release(name,
          [name]() { boost::interprocess::named_mutex::remove(name.c_str()); });
}

std::unique_ptr<boost::interprocess::named_semaphore>
shm_bookkeeper::acquire_named_semaphore(const std::string &name) {
  std::unique_ptr<boost::interprocess::named_semaphore> ret;
  acquire(name,
          [&ret, name]() {
            ret = std::unique_ptr<boost::interprocess::named_semaphore>(
                new boost::interprocess::named_semaphore(
                    boost::interprocess::open_or_create, name.c_str(), 0));
          },
          tag_counter_t::boost_semaphore);
  return ret;
}

void shm_bookkeeper::release_named_semaphore(const std::string &name) {
  release(name, [name]() {
    boost::interprocess::named_semaphore::remove(name.c_str());
  });
}

std::unique_ptr<boost::interprocess::named_condition>
shm_bookkeeper::acquire_named_condition(const std::string &name) {
  std::unique_ptr<boost::interprocess::named_condition> ret;
  acquire(name,
          [&ret, name]() {
            ret = std::unique_ptr<boost::interprocess::named_condition>(
                new boost::interprocess::named_condition(
                    boost::interprocess::open_or_create, name.c_str()));
          },
          tag_counter_t::boost_condition);
  return ret;
}
void shm_bookkeeper::release_named_condition(const std::string &name) {
  release(name, [name]() {
    boost::interprocess::named_condition::remove(name.c_str());
  });
}
}
}
} // namespace neb
