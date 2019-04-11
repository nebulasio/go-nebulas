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
#pragma once
#include "common/common.h"
#include <boost/interprocess/allocators/allocator.hpp>
#include <boost/interprocess/containers/string.hpp>
#include <boost/interprocess/containers/vector.hpp>
#include <boost/interprocess/managed_shared_memory.hpp>
#include <boost/interprocess/sync/named_condition.hpp>
#include <boost/interprocess/sync/named_mutex.hpp>
#include <boost/interprocess/sync/named_semaphore.hpp>
#include <boost/interprocess/sync/scoped_lock.hpp>
#include <boost/process/child.hpp>

namespace neb {
namespace ipc {

enum shm_role { role_util, role_server, role_client };
struct shm_server {
  inline static std::string role_name(const std::string &name) {
    return name + std::string(".server");
  }
};
struct shm_client {
  inline static std::string role_name(const std::string &name) {
    return name + std::string(".client");
  }
};
typedef uint32_t shm_type_id_t;
namespace internal {
template <typename T> struct shm_other_side_role {};
template <> struct shm_other_side_role<shm_server> { typedef shm_client type; };
template <> struct shm_other_side_role<shm_client> { typedef shm_server type; };
}

template <typename A>
auto check_exists(const std::string &v) -> typename std::enable_if<
    std::is_same<A, boost::interprocess::named_mutex>::value ||
        std::is_same<A, boost::interprocess::named_semaphore>::value ||
        std::is_same<A, boost::interprocess::named_condition>::value,
    bool>::type {
  try {
    A _l(boost::interprocess::open_only, v.c_str());
  } catch (...) {
    return false;
  }
  return true;
}

struct shm_service_failure : public std::exception {
  inline shm_service_failure(const std::string &msg) : m_msg(msg) {}
  inline const char *what() const throw() { return m_msg.c_str(); }

protected:
  const std::string m_msg;
};

struct shm_init_failure : public std::exception {
public:
  inline shm_init_failure(const std::string &msg) : m_msg(msg) {}
  inline const char *what() const throw() { return m_msg.c_str(); }

protected:
  const std::string m_msg;
};

struct shm_handle_recv_failure : public std::exception {
public:
  inline shm_handle_recv_failure(const std::string &msg) : m_msg(msg) {}
  inline const char *what() const throw() { return m_msg.c_str(); }

protected:
  const std::string m_msg;
};

typedef boost::interprocess::managed_shared_memory::segment_manager
    segment_manager_t;
typedef boost::interprocess::allocator<void, segment_manager_t>
    default_allocator_t;

namespace internal {
template <typename T> struct boost_ipc_vector_helper {
  typedef boost::interprocess::allocator<T, segment_manager_t> elem_allocator_t;
  typedef boost::interprocess::vector<T, elem_allocator_t> vector_t;
};
} // namespace internal
typedef boost::interprocess::allocator<char, segment_manager_t>
    char_allocator_t;
typedef boost::interprocess::basic_string<char, std::char_traits<char>,
                                          char_allocator_t>
    char_string_t;

template <typename T>
using vector = typename internal::boost_ipc_vector_helper<T>::vector_t;

}
}
