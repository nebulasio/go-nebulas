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
#include "common/byte.h"
#include "common/common.h"

namespace neb {
namespace fs {
enum storage_open_flag {
  storage_open_for_readwrite,
  storage_open_for_readonly,
  storage_open_default = storage_open_for_readonly,
};

struct storage_general_failure : public std::exception {
  inline storage_general_failure(const std::string &msg) : m_msg(msg) {}
  inline const char *what() const throw() { return m_msg.c_str(); }
protected:
  std::string m_msg;
};
struct storage_exception_no_such_key : public std::exception {
  inline const char *what() const throw() { return "no such key in storage"; }
};
struct storage_exception_no_init : public std::exception {
  inline const char *what() const throw() { return "storage no initialized"; }
};

class storage {
public:
  template <typename T, typename KT>
  auto get(const KT &key) -> typename std::enable_if<
      std::is_arithmetic<T>::value && std::is_arithmetic<KT>::value, T>::type {
    return byte_to_number<T>(get_bytes(number_to_byte<bytes>(key)));
  }

  template <typename T, typename KT>
  auto put(const KT &key, const T &val) ->
      typename std::enable_if<std::is_arithmetic<T>::value &&
                                  std::is_arithmetic<KT>::value,
                              void>::type {
    put_bytes(number_to_byte<bytes>(key), number_to_byte<bytes>(val));
  }
  template <typename KT>
  auto del(const KT &key) ->
      typename std::enable_if<std::is_arithmetic<KT>::value, void>::type {
    del_by_bytes(number_to_byte<bytes>(key));
  }

  bytes get(const std::string &key) { return get_bytes(string_to_byte(key)); }
  void put(const std::string &key, const bytes &value) {
    return put_bytes(string_to_byte(key), value);
  }
  void del(const std::string &key) { del_by_bytes(string_to_byte(key)); }

  virtual bytes get_bytes(const bytes &key) = 0;
  virtual void put_bytes(const bytes &key, const bytes &val) = 0;
  virtual void del_by_bytes(const bytes &key) = 0;

  virtual void enable_batch() = 0;
  virtual void disable_batch() = 0;
  virtual void flush() = 0;
};
}
}
