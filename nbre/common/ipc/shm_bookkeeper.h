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
#include "common/ipc/shm_base.h"
#include <boost/interprocess/allocators/allocator.hpp>
#include <boost/interprocess/containers/map.hpp>
#include <boost/interprocess/containers/string.hpp>
#include <boost/interprocess/managed_shared_memory.hpp>

namespace neb {
namespace ipc {

void clean_bookkeeper_env(const std::string &name);

namespace internal {
class shm_bookkeeper {
public:
  shm_bookkeeper(const std::string &name);

  std::unique_ptr<boost::interprocess::named_mutex>
  acquire_named_mutex(const std::string &name);
  void release_named_mutex(const std::string &name);

  std::unique_ptr<boost::interprocess::named_semaphore>
  acquire_named_semaphore(const std::string &name);
  void release_named_semaphore(const std::string &name);

  std::unique_ptr<boost::interprocess::named_condition>
  acquire_named_condition(const std::string &name);
  void release_named_condition(const std::string &name);

  void acquire(const std::string &name, const std::function<void()> &action);
  void release(const std::string &name, const std::function<void()> &action);

  void reset();

private:
  void acquire(const std::string &name, const std::function<void()> &action,
               uint8_t type);

  inline std::string mutex_name() { return m_name + ".mutex"; }
  inline std::string mem_name() { return m_name + ".mem"; }

protected:
  struct tag_counter_t {
  public:
    enum type_tag {
      boost_mutex = 0,
      boost_semaphore = 1,
      boost_condition = 2,
      other_unknown,
    };

    inline uint64_t data() { return m_data.m_data; }
    inline void set_data(uint64_t d) { m_data.m_data = d; }
    inline uint64_t counter() { return m_data.m_detail.m_counter; }
    inline type_tag type() {
      return static_cast<type_tag>(m_data.m_detail.m_type);
    }
    inline void set_type(type_tag tt) { m_data.m_detail.m_type = tt; }
    inline void set_counter(uint64_t c) { m_data.m_detail.m_counter = c; };

  protected:
    union tag_counter_data {
      uint64_t m_data;
      struct {
        uint8_t m_type : 2;
        uint64_t m_counter : 62;
      } m_detail;
    };

    tag_counter_data m_data;
  };

  typedef boost::interprocess::managed_shared_memory::segment_manager
      segment_manager_t;
  typedef boost::interprocess::allocator<char, segment_manager_t>
      char_allocator_t;
  typedef boost::interprocess::basic_string<char, std::char_traits<char>,
                                            char_allocator_t>
      char_string_t;

  typedef std::pair<const char_string_t, uint64_t> map_value_t;
  typedef std::pair<char_string_t, uint64_t> movable_map_value_t;
  typedef boost::interprocess::allocator<map_value_t, segment_manager_t>
      map_allocator_t;

  typedef boost::interprocess::map<char_string_t, uint64_t,
                                   std::less<char_string_t>, map_allocator_t>
      map_t;

  std::string m_name;

  std::unique_ptr<boost::interprocess::managed_shared_memory> m_segment;
  std::unique_ptr<map_allocator_t> m_allocator;
  std::unique_ptr<boost::interprocess::named_mutex> m_mutex;
  map_t *m_map;

}; // end class shm_bookkeeper
}
}
}
